package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/valkey-io/valkey-go"
)

const version = "v0.4.0"

const idKey = "notiflex:id"

type idStore interface {
	NextID(context.Context) (uint64, error)
}

type valkeyIDStore struct {
	client valkey.Client
}

type api struct {
	podName string
	ids     idStore
	events  eventPublisher
}

type eventPublisher interface {
	Publish(context.Context, notificationEvent) error
}

type notificationEvent struct {
	ID          string    `json:"id"`
	GeneratedBy string    `json:"generated_by"`
	Timestamp   time.Time `json:"timestamp"`
}

type kafkaEvents struct {
	producer sarama.SyncProducer
	group    sarama.ConsumerGroup
}

type healthResponse struct {
	Status string `json:"status"`
}

type versionResponse struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Hostname  string `json:"hostname"`
}

type idResponse struct {
	ID          string `json:"id"`
	GeneratedBy string `json:"generated_by"`
}

func main() {
	podName, err := os.Hostname()
	if err != nil {
		podName = "unknown"
	}

	password, err := valkeyPassword()
	if err != nil {
		log.Fatal(err)
	}

	ids, err := newValkeyIDStore(os.Getenv("VALKEY_ADDR"), password)
	if err != nil {
		log.Fatal(err)
	}
	defer ids.Close()

	events, err := newKafkaEvents(os.Getenv("KAFKA_BROKER"))
	if err != nil {
		log.Fatal(err)
	}
	if events != nil {
		defer events.Close()
		go events.Consume(context.Background())
	}

	server := &http.Server{
		Addr:              ":8080",
		Handler:           newHandler(&api{podName: podName, ids: ids, events: events}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("notiflex API listening on %s (pod: %s)", server.Addr, podName)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func newKafkaEvents(broker string) (*kafkaEvents, error) {
	if broker == "" {
		return nil, nil
	}

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_1_0_0
	cfg.Producer.Return.Successes = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	producer, err := sarama.NewSyncProducer([]string{broker}, cfg)
	if err != nil {
		return nil, fmt.Errorf("Kafka producer 연결 실패: %w", err)
	}
	group, err := sarama.NewConsumerGroup([]string{broker}, "notiflex-api", cfg)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("Kafka consumer 연결 실패: %w", err)
	}
	return &kafkaEvents{producer: producer, group: group}, nil
}

func (k *kafkaEvents) Publish(ctx context.Context, event notificationEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	message := &sarama.ProducerMessage{Topic: "notifications", Key: sarama.StringEncoder(event.ID), Value: sarama.ByteEncoder(payload)}
	_, _, err = k.producer.SendMessage(message)
	return err
}

func (k *kafkaEvents) Consume(ctx context.Context) {
	handler := kafkaConsumerHandler{}
	for ctx.Err() == nil {
		if err := k.group.Consume(ctx, []string{"notifications"}, handler); err != nil {
			log.Printf("Kafka consumer 오류: %v", err)
			time.Sleep(time.Second)
		}
	}
}

func (k *kafkaEvents) Close() {
	if err := k.group.Close(); err != nil {
		log.Printf("Kafka consumer 종료: %v", err)
	}
	if err := k.producer.Close(); err != nil {
		log.Printf("Kafka producer 종료: %v", err)
	}
}

type kafkaConsumerHandler struct{}

func (kafkaConsumerHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (kafkaConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (kafkaConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Kafka consumer: received message on %s: %s", message.Topic, message.Value)
		session.MarkMessage(message, "")
	}
	return nil
}

func valkeyPassword() (string, error) {
	if path := os.Getenv("VALKEY_PASSWORD_FILE"); path != "" {
		password, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("Valkey 비밀번호 파일 읽기 실패: %w", err)
		}
		if len(password) == 0 {
			return "", fmt.Errorf("Valkey 비밀번호 파일이 비어 있음")
		}
		return string(password), nil
	}

	password := os.Getenv("VALKEY_PASSWORD")
	if password == "" {
		return "", fmt.Errorf("VALKEY_PASSWORD_FILE 또는 VALKEY_PASSWORD가 필요함")
	}
	return password, nil
}

func newValkeyIDStore(addr, password string) (*valkeyIDStore, error) {
	if addr == "" {
		return nil, fmt.Errorf("VALKEY_ADDR is required")
	}

	var lastErr error
	for attempt := 1; attempt <= 10; attempt++ {
		client, err := valkey.NewClient(valkey.ClientOption{
			InitAddress: []string{addr},
			Password:    password,
		})
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err = client.Do(ctx, client.B().Ping().Build()).Error()
			cancel()
			if err == nil {
				return &valkeyIDStore{client: client}, nil
			}
			client.Close()
		}

		lastErr = err
		log.Printf("Valkey 연결 재시도 %d/10: %v", attempt, err)
		if attempt < 10 {
			time.Sleep(3 * time.Second)
		}
	}

	return nil, fmt.Errorf("Valkey 연결 실패: %w", lastErr)
}

func (s *valkeyIDStore) NextID(ctx context.Context) (uint64, error) {
	id, err := s.client.Do(ctx, s.client.B().Incr().Key(idKey).Build()).AsInt64()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (s *valkeyIDStore) Close() {
	s.client.Close()
}

func newHandler(service *api) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", service.health)
	mux.HandleFunc("GET /version", service.version)
	mux.HandleFunc("GET /id", service.nextID)
	return mux
}

func (a *api) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, versionResponse{
		Version:   version,
		GoVersion: runtime.Version(),
		Hostname:  a.podName,
	})
}

func (a *api) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func (a *api) nextID(w http.ResponseWriter, r *http.Request) {
	id, err := a.ids.NextID(r.Context())
	if err != nil {
		log.Printf("generate ID: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "Valkey unavailable"})
		return
	}
	idString := strconv.FormatUint(id, 10)
	if a.events != nil {
		if err := a.events.Publish(r.Context(), notificationEvent{ID: idString, GeneratedBy: a.podName, Timestamp: time.Now().UTC()}); err != nil {
			log.Printf("Kafka publish: %v", err)
			writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "Kafka unavailable"})
			return
		}
	}
	writeJSON(w, http.StatusOK, idResponse{
		ID:          idString,
		GeneratedBy: a.podName,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
