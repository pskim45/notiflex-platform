# Kafka Topic 삭제

Kafka Topic 삭제는 메시지와 운영 이력을 되돌리기 어려운 파괴적 작업이다. 조회만으로 대상을 확인한 뒤, 사용자에게 정확한 Topic과 데이터 손실 범위를 보여주고 명시적인 승인을 받아야 한다.

## 사전 확인

1. 클러스터·namespace·Topic 이름을 정확히 확인한다.
2. Topic의 partition, replica, 상태와 미처리 메시지(lag)를 확인한다.
3. 이 Topic을 사용하는 Producer와 Consumer를 파악하고, 재생성 시 복구할 설정을 기록한다.
4. 보존해야 할 메시지가 있다면 삭제 전에 별도 저장소로 백업한다.
5. 사용자에게 `kafka/notiflex-kafka/<topic-name>` 삭제, 메시지 영구 손실, 영향받는 Producer·Consumer를 명시하고 승인을 받는다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops get kafkatopic -n kafka
kubectl --context gke-sysnet4admin_book_gitaiops get kafkatopic <topic-name> -n kafka -o yaml
kubectl --context gke-sysnet4admin_book_gitaiops get pods -n kafka
```

## 실행

1. Producer를 중지하거나 해당 Topic으로의 신규 메시지 유입을 차단한다.
2. Consumer lag가 0이 될 때까지 기다리고 마지막 처리 결과를 확인한다.
3. GitOps 소유 리소스이면 `k8s/kafka/`의 `KafkaTopic` 선언을 제거해 검토·커밋·push하고 Argo CD 동기화를 기다린다. `kubectl delete`로 먼저 삭제하면 self-heal로 다시 생성될 수 있다.
4. GitOps 소유가 아닌 Topic만 승인된 정확한 이름으로 삭제한다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops delete kafkatopic <topic-name> -n kafka
```

## 사후 검증

1. `KafkaTopic`과 실제 Kafka Topic이 사라졌는지 확인한다.
2. 관련 Argo CD Application이 `Synced/Healthy`인지 확인한다.
3. Producer·Consumer 로그에서 예상하지 못한 재시도나 오류가 없는지 확인한다.
4. 실행자, 승인자, 대상, 시각, 백업 위치와 검증 결과를 운영 기록에 남긴다.

## 중단 조건

- Topic 이름, 소유 Application, Producer·Consumer 중 하나라도 식별되지 않으면 삭제하지 않는다.
- lag가 남아 있거나 백업 필요 여부가 결정되지 않았으면 삭제하지 않는다.
- 사용자의 명시적인 삭제 승인이 없으면 조회 결과만 보고하고 중단한다.
