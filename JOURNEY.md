# Notiflex 여정 기록

이 파일은 프로젝트에서 실제로 진행하고 검증한 내용을 기록한다. 각 단계가 완료되면 실제 결과를 기준으로 갱신한다.

## 진행 현황

| 챕터 | 서브챕터 | 상태 | 완료일 | 비고 |
|------|---------|------|--------|------|
| ch2 | 2.2 설치 확인 | ✅ | 2026-08-03 | Git, gcloud, kubectl, Codex 실행 환경 확인 |
| ch2 | 2.3 gcloud 설정 | ✅ | 2026-08-03 | 프로젝트·서울 리전·존 및 활성 계정 확인 |
| ch2 | 2.4 GitHub 저장소 | ✅ | 2026-08-03 | public 원격 저장소 `pskim45/notiflex-platform` 연결 |
| ch2 | 2.5 GKE 클러스터 | ✅ | 2026-08-03 | GKE Standard, Spot `e2-medium` 2노드, Gateway API 활성화 |
| ch2 | 2.6 빌드/배포 | ✅ | 2026-08-03 | Cloud Build 테스트 성공, `v0.1.0` 배포 및 API 검증 |
| ch2 | 2.7 첫 커밋 | ✅ | 2026-08-03 | 코드·매니페스트·문서 최초 커밋 및 GitHub 푸시 |
| ch2 | update-docs 스킬 | ✅ | 2026-08-03 | 저장소 문서 동적 탐색·동기화·검증·커밋 워크플로 추가 |
| ch3 | 3.2 GitOps 도구 | ✅ | 2026-08-04 | Argo CD 3.3.6 설치, private GitHub 저장소 연결, `notiflex-smb` Synced/Healthy 확인 |
| ch3 | 3.3 기능 추가 | ✅ | 2026-08-04 | `/version`에 앱·Go·Pod 정보 추가, `v0.1.2` 롤링 업데이트 및 응답 검증 |
| ch3 | 3.4 CI | ✅ | 2026-08-04 | GitHub Actions와 Workload Identity Federation으로 테스트·SHA 이미지 빌드·푸시 자동화 |
| ch3 | 3.5 CI-CD 연결 | ✅ | 2026-08-04 | 코드-only push로 `v0.1.3` 테스트·SHA 이미지 게시·매니페스트 커밋·Argo CD 자동 배포 검증 |
| ch4 | 4.2 메트릭 모니터링 | ✅ | 2026-08-04 | kube-prometheus-stack 설치, active target 16개 Up 및 Notiflex 대시보드 로딩 검증 |
| ch4 | 4.3 로그 수집 | ✅ | 2026-08-04 | Loki SingleBinary와 노드별 Fluent Bit 설치, Grafana 데이터소스 및 실시간 로그 조회 검증 |
| ch4 | 4.4 알림 | 🚧 | | `PodRestartTooMany` 로드 및 health 확인 완료, 외부 receiver 연결·Firing 검증 대기 |
| ch5 | 5.2 트래픽 관리 | ✅ | 2026-08-04 | GKE 리전 외부 Gateway·HTTPRoute·HealthCheckPolicy 구성, 외부 `/health`·`/id`·`/version` HTTP 200 검증 |
| ch5 | 5.3 무중단 배포 | ✅ | 2026-08-04 | Argo Rollouts Blue/Green 전환 및 `v0.2.2` 재배포에서 preview·30초 자동 승격·active 전환·구 버전 축소 검증 |
| ch5 | 5.4 ADR | ✅ | 2026-08-04 | ch3~ch5의 아키텍처 결정 7건을 시간 순서로 정식 기록 |
| 운영 | 비용 중단 | ✅ | 2026-08-04 | 복구 런북을 GitHub에 보존한 뒤 GKE·Gateway/LB·디스크·Artifact Registry·Cloud Build 버킷 삭제 |
| 운영 | 환경 복구 | ✅ | 2026-08-08 | 복구 런북으로 GKE·Artifact Registry·애플리케이션·GitOps·Gateway·관측성 스택 재구축 및 외부 API 검증 |
| ch6 | 6.1 캐시 | ✅ | 2026-08-08 | Valkey standalone 설치, API `v0.3.0`의 `INCR notiflex:id` 전환 및 외부 ID 1~6·저장값 일치 검증 |
| ch6 | 6.2 시크릿 관리 | ✅ | 2026-08-08 | GCP Secret Manager·Workload Identity·GKE managed CSI로 Valkey credential 파일 마운트 및 외부 API 검증 |
| ch6 | 6.3 Canary 전환 | ✅ | 2026-08-08 | Argo Rollouts Canary 전환, `v0.3.2`에서 20%·50%·80% 각 30초 pause와 100% 승격 검증 |
| ch6 | 6.4 아키텍처 컨텍스트 | ✅ | 2026-08-08 | `claude-context/architecture.md`에 현재 컴포넌트·연결 관계·핵심 설정·namespace·배포 및 관측성 경로를 클러스터 실측 기준으로 기록 |
| ch7 | 7.2 멀티 노드풀 | ✅ | 2026-08-08 | `api-pool`·`worker-pool`·`ops-pool` 생성, API nodeSelector 적용과 Canary 재배포 후 전용 노드 배치·CSI credential·외부 API 검증 |
| ch7 | 7.3 App of Apps | ✅ | 2026-08-08 | `root-app` 아래 bootstrap·API·Valkey·Prometheus·Loki·Fluent Bit·모니터링 설정 7개 하위 앱과 wave 0→1→2 순서를 구성하고 전체 Synced/Healthy 검증 |
| ch7 | 7.4 멀티테넌시 | ✅ | 2026-08-08 | `enterprise` 전용 namespace·Canary Rollout·RBAC·ResourceQuota·CSI identity를 추가하고 App of Apps 동기화와 API/공유 Valkey 동작 검증 |
| ch7 | 권한 분리 체험 | ✅ | 2026-08-08 | 로컬 `settings.local.json`으로 `kubectl delete/apply` 차단과 노드풀 삭제 승인 요청을 확인하고, 삭제 거부 후 설정·백업·Git 변경 없이 완전히 되돌림 |
| ch8 | 8.1 메시징 | ✅ | 2026-08-08 | Strimzi `0.51.0`·Kafka `4.1.0` KRaft 단일 브로커와 `notifications` Topic을 GitOps로 설치하고 API Producer/Consumer·외부 `/id` 이벤트 수신 검증 |
| ch8 | 8.2 트레이싱 | ✅ | 2026-08-08 | Tempo SingleBinary와 Grafana 데이터소스, OTel SDK를 GitOps로 설치하고 실제 Trace에서 HTTP→Valkey→Kafka produce→consume Span 연결 검증 |
| ch8 | 8.3 CronJob | ✅ | 2026-08-08 | `notiflex-healthcheck`가 5분마다 내부 API `/health`를 점검하도록 선언하고 ops-pool 배치·성공 Job 로그 검증 |
| ch8 | 위험 작업 가드레일 | ✅ | 2026-08-08 | Kafka Topic 삭제·CronJob 수동 실행·테넌트 Namespace 삭제의 사전 확인, 승인, 실행, 검증 및 중단 조건을 절차서로 기록 |
| ch9 | 9.1 저장소 분석 | ✅ | 2026-08-08 | 최종 56개 파일·약 3,160줄·77개 커밋을 분석하고 Argo CD 12개 앱 Synced/Healthy 및 Git↔클러스터 일치 확인 |
| ch9 | 9.2 회고 | ✅ | 2026-08-08 | managed GKE·Argo 생태계·Grafana 통합 관측·GitOps 호환이라는 반복 선택 기준과 기술 부채를 종합 |
| ch9 | 9.3 온보딩 문서 | ✅ | 2026-08-08 | 실측 노드풀·namespace별 Pod, 저장소 구조, Argo CD·Grafana·API 접근, 배포 흐름과 운영 Q&A를 `ONBOARDING.md`에 기록 |
| ch9 | 9.4 GitAIOps 분석 | ✅ | 2026-08-08 | Git의 선언·기억, AI의 해석·변경, Ops의 적용·검증이 JOURNEY·ADR·가드레일로 되먹임되는 루프 분석 |
| ch9 | 9.5 마무리 | ✅ | 2026-08-08 | 알림·헬스체크, 진짜 비동기 처리, 보안, 고가용성·스케일링, 데이터 격리와 비용 최적화 로드맵 제안 |
| 운영 | 최종 비용 중단 | ✅ | 2026-08-08 | GKE·VM·디스크·Gateway/LB·IP·Artifact Registry·Secret·notiflex-sa·proxy subnet 삭제 후 과금 가능 자원 0건 전수 검증, Git·무료 IAM/WIF만 보존 |

## 도구 선택 기록

| 영역 | 선택 | 검토한 대안 | 선택 이유 |
|------|------|-------------|----------|
| 컨테이너 런타임 | GKE Standard (Zonal) | GKE Autopilot | 노드 구성과 단계별 플랫폼 확장을 직접 실습하기 위해 선택 |
| 이미지 저장소 | Artifact Registry | Docker Hub | GCP IAM 및 Cloud Build와의 네이티브 통합 |
| 이미지 빌드 | GitHub Actions Docker build | Cloud Build, 로컬 Docker 빌드 | 코드 push 시 테스트·SHA 이미지 빌드·Artifact Registry 게시를 자동화 |
| CI | GitHub Actions + Workload Identity Federation | Cloud Build Trigger, Jenkins, 서비스 계정 키 | GitHub push와 직접 연동하고 장기 키 없이 최소 권한으로 Artifact Registry에 게시 |
| 메트릭 모니터링 | Prometheus + Grafana | Google Cloud Monitoring, Datadog | Kubernetes 표준 메트릭과 이후 Loki·Tempo를 Grafana에 통합하기 위해 선택 |
| 로그 수집 | Loki + Fluent Bit | Elasticsearch, Grafana Alloy | 경량 구성으로 Kubernetes stdout/stderr 로그를 수집하고 기존 Grafana Explore에 통합하기 위해 선택 |
| 알림 규칙 | PrometheusRule + Alertmanager | Grafana Alerting | kube-prometheus-stack의 기존 평가·라우팅 경로를 재사용하고 규칙을 YAML로 버전 관리하기 위해 선택 |
| 외부 트래픽 관리 | GKE Gateway API | Ingress NGINX, Istio | GKE 네이티브 리전 외부 로드밸런서와 Kubernetes 표준 API를 별도 Controller 없이 사용하기 위해 선택 |
| 무중단 배포 | Argo Rollouts Canary | Blue/Green, Flagger, Kubernetes Rolling Update | 동일 Argo GitOps 구성에서 단계적 weight와 관찰 시간을 선언하고 신규 버전 노출 위험을 줄이기 위해 전환 |
| 캐시·상태 공유 | Valkey | Redis, Memcached, DragonflyDB | Redis 호환 `INCR`로 Pod 간 ID를 원자적으로 공유하고 BSD 라이선스를 유지 |
| 시크릿 관리 | GKE Secret Manager CSI + Workload Identity | Kubernetes Secret 직접 주입, Sealed Secrets, External Secrets Operator | 장기 SA 키 없이 Secret Manager 값을 읽기 전용 파일로 전달하고 IAM 최소 권한 적용 |
| 문서 동기화 | 저장소 범위 Codex 스킬 | 전역 개인 스킬, 고정 문서 목록 | 팀과 공유하고 이후 장의 신규 문서도 수정 없이 동적으로 처리 |
| 노드 스케줄링 | GKE 역할별 노드풀 + nodeSelector | 단일 노드풀, taint/toleration, nodeAffinity | GKE 자동 노드풀 라벨로 API 배치를 단순하고 명시적으로 분리하고 이후 worker·ops 워크로드 확장 기반 마련 |
| 다중 앱 관리 | Argo CD App of Apps + sync wave | 개별 수동 Application, ApplicationSet | 단일 클러스터의 7개 하위 앱을 Git으로 묶고 namespace→backend→수집기·설정·API 의존 순서를 명시 |
| 멀티테넌시 (ch7.4) | Namespace 분리 + 고객별 Rollout·RBAC·ResourceQuota | 단일 namespace 라벨 격리, vCluster, 고객별 클러스터 | 학습 환경의 비용을 유지하면서 고객별 배포·권한·자원 한도를 분리한다. Valkey와 credential은 현재 공유하므로 데이터 격리는 아님 |
| 이벤트 메시징 | Kafka + Strimzi Operator | RabbitMQ, NATS, Valkey Streams | KRaft 영속 메시징과 Consumer Group을 사용하고 Kafka CRD를 기존 App of Apps GitOps 흐름으로 관리 |
| 분산 트레이싱 | Grafana Tempo + OpenTelemetry | Jaeger, Zipkin | 기존 Grafana에 메트릭·로그·Trace를 통합하고 OTLP 표준으로 HTTP·Valkey·Kafka 구간을 연결 |
| 배치 자동화 (ch8.3) | K8s CronJob | 외부 cron + Kubernetes 외부 트리거, Argo Workflows | Kubernetes 네이티브 스케줄링, ops-pool 배치, Argo CD 매니페스트 관리로 단순 헬스체크를 자동화 |

## 현재 검증 버전

> 2026-08-08 복구 런북으로 GCP 워크로드를 재구축하고 아래 버전을 다시 검증했다.

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | 1.25.12 | 실행 중 API `/version` 응답으로 확인 |
| Notiflex 이미지 | `sha-eee42f5` | CI run `31255519147`에서 테스트·빌드·게시, SMB·Enterprise API `v0.5.0` Canary 완료와 Tempo Trace 검증 |
| GKE | `1.35.6-gke.1250000` | 최초 클러스터 생성 |
| ArgoCD | `v3.3.6` | 최초 설치 및 `notiflex-smb` Application 연결 |
| kube-prometheus-stack | chart `88.1.3` | Prometheus `3.13.2`, Grafana `13.1.1`, Alertmanager `0.33.1` |
| Loki | chart `7.2.0` | Loki `3.6.11`, SingleBinary, filesystem 5Gi |
| Fluent Bit | chart `0.57.9` | Fluent Bit `5.0.9`, 노드별 DaemonSet |
| Argo Rollouts | `v1.9.1` | Canary 20%→50%→80%→100%, 단계별 30초 pause 검증 |
| Valkey | chart `6.2.6`, app `9.1.1`, client `valkey-go v1.0.73` | standalone 1 Pod, `10m/64Mi` request, 인증 PING 및 `INCR` 검증 |
| GKE Secret Manager CSI | GKE managed `secrets-store-gke.csi.k8s.io`, provider `gke` | Workload Identity pool과 모든 노드풀 `GKE_METADATA`, driver/provider 각 5/5 Ready |
| Kafka | Strimzi chart/operator `0.51.0`, Kafka `4.1.0`, IBM/sarama `v1.47.0` | KRaft 단일 dual-role broker, metadata `4.1-IV1`, `notifications` 3 partitions |
| Tempo·OTel | Tempo chart `1.24.4`, app `2.9.0`; OTel SDK `1.43.0`, otelhttp `0.68.0`, gRPC `1.80.0` | OTLP gRPC 4317, 24h retention, TraceID와 HTTP·Valkey·Kafka Span 조회 검증 |

## 현재 리소스 스냅샷

> 아래는 2026-08-08 삭제 직전 검증한 마지막 실행 상태이자 복구 목표다. 현재 리소스는 삭제되어 존재하지 않으며 재구축 절차는 `docs/shutdown-recovery.md`를 따른다.

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|-----------|---------|---------------|
| `default-pool` | `e2-medium` Spot VM | 2 | 관측성, Argo CD 및 컨트롤러 |
| `api-pool` | `e2-medium` Spot VM | 1 | SMB·Enterprise `notiflex-api` 각 1 replica |
| `worker-pool` | `e2-standard-2` Spot VM | 1 | Valkey, Strimzi operator, Kafka broker·Topic Operator, Tempo SingleBinary |
| `ops-pool` | `e2-small` Spot VM | 1 | `notiflex-healthcheck` CronJob과 시스템 DaemonSet |

| Kubernetes 리소스 | 네임스페이스 | 상태 |
|---------------------|---------------|------|
| Rollout `notiflex-api` | `notiflex` | Canary, `sha-eee42f5`, step 6/6, Healthy, API `v0.5.0`, OTLP·Kafka·CSI 설정 사용 |
| Service `notiflex-api` | `notiflex` | ClusterIP, 80 → 8080 |
| Service `notiflex-api-preview` | `notiflex` | Canary ReplicaSet용 ClusterIP, 80 → 8080 |
| Application `notiflex-smb` | `argocd` | Synced, Healthy, auto-sync/prune/selfHeal 활성화 |
| Application `root-app` | `argocd` | `argocd/apps/` 감시, 하위 Application 11개 자동 등록, 전체 12개 Application Synced/Healthy |
| Application `bootstrap` | `argocd` | wave 0에서 `notiflex`·`enterprise`·`monitoring`·`kafka` namespace 관리, Synced/Healthy |
| Application `strimzi`·`kafka` | `argocd` | wave 1 operator → wave 2 cluster/topic 순서, 모두 Synced/Healthy |
| Application `tempo` | `argocd` | chart `1.24.4`, worker-pool 배치, Synced/Healthy |
| Application `valkey`·`kube-prometheus`·`loki`·`fluent-bit` | `argocd` | 외부 Helm chart + Git values 다중 source, 모두 Synced/Healthy |
| Application `monitoring-config` | `argocd` | `monitoring/`과 `k8s/monitoring/` 일반 YAML 관리, Synced/Healthy |
| Prometheus·Grafana·Alertmanager | `monitoring` | 모든 Pod Running, active scrape target 28/28 Up |
| ConfigMap `notiflex-dashboard` | `monitoring` | Grafana sidecar 로딩 완료, CPU·메모리·재시작 패널 구성 |
| StatefulSet `loki`·Deployment `loki-gateway` | `monitoring` | Running, PVC 5Gi Bound, LogQL 조회 성공 |
| DaemonSet `fluent-bit` | `monitoring` | 5/5 Ready, 모든 노드의 로그를 Loki로 push |
| ConfigMap `loki-datasource` | `monitoring` | Grafana sidecar 로딩 및 datasource reload 200 확인 |
| PrometheusRule `pod-restart-alert` | `monitoring` | Operator 검증 완료, `PodRestartTooMany` health `ok`·현재 `inactive` |
| Gateway `notiflex-gateway` | `notiflex` | `35.216.50.229`, `Programmed=True`, `GatewayHealthy=True` |
| HTTPRoute `notiflex-route` | `notiflex` | `/` → `notiflex-api:80`, Accepted·ResolvedRefs·Reconciled=True |
| HealthCheckPolicy `notiflex-healthcheck` | `notiflex` | `/health:8080`, Gateway backend Healthy |
| CronJob `notiflex-healthcheck` | `notiflex` | `*/5 * * * *`, ops-pool, 실패 Job 정리 후 예약 Job `Completed`·HTTP 200 검증 |
| Deployment `argo-rollouts` | `argo-rollouts` | controller `v1.9.1`, 1/1 Ready |
| StatefulSet `valkey-primary` | `notiflex` | standalone, 1/1 Ready, Secret `valkey` 참조, `notiflex:id` 공유 |
| ServiceAccount `notiflex-api` | `notiflex` | GSA `notiflex-sa`와 Workload Identity 연결, `valkey-password` accessor만 부여 |
| SecretProviderClass `notiflex-secrets` | `notiflex` | Secret Manager `valkey-password` version 1을 읽기 전용 파일로 mount, mounted=true |
| Rollout `notiflex-api` | `enterprise` | 고객 전용 Canary, `sha-eee42f5`, API `v0.5.0`, 1/1 Ready, OTLP 설정, 내부 ClusterIP만 사용 |
| RBAC·ResourceQuota | `enterprise` | `tenant-reader`는 조회만 허용, Pod 10·Service 5 및 CPU/메모리 상한 적용 |
| SecretProviderClass `notiflex-secrets` | `enterprise` | 동일 GSA와 Secret Manager credential을 CSI 파일로 mount, mounted=true |
| Kafka `notiflex-kafka` | `kafka` | KRaft dual-role 1 broker, Kafka 4.1.0, PVC 5Gi, `worker-pool`, Ready=True |
| KafkaTopic `notifications` | `kafka` | 3 partitions, replication factor 1, Ready=True |
| StatefulSet `tempo` | `monitoring` | app `2.9.0`, 1/1 Ready, worker-pool, OTLP gRPC 4317·HTTP 4318, ephemeral 24h 설정 |
| ConfigMap `tempo-datasource` | `monitoring` | Grafana sidecar 파일 생성과 datasource reload HTTP 200 확인 |

## TODO

- [ ] `/live`·`/ready`·의존성 진단을 분리하고 CronJob·API 5xx·Kafka consumer lag를 Alertmanager 외부 receiver와 연결한다.
- [ ] 실제 알림 worker에 retry·DLQ·idempotency를 구현하고 요청 경로의 동기 Kafka publish를 outbox 또는 동등한 비동기 수락 구조로 분리한다.
- [ ] NetworkPolicy, Gateway TLS·인증·rate limit과 정책 엔진을 적용해 tenant 및 플랫폼 통신 경계를 강제한다.
- [ ] API 2 replicas, PDB, HPA와 실제 Gateway weighted routing을 구성하고 메트릭 기반 Canary 자동 승격·중단을 검증한다.
- [ ] 고객별 Valkey·credential·Kafka 경계를 설계하고 Kafka·Valkey 고가용성 및 데이터 백업·복구 목표(RPO/RTO)를 정의한다.
- [ ] VPA recommendation과 장기 사용량을 기준으로 default-pool·ops-pool requests와 비용을 right-sizing한다.
- [ ] `PodRestartTooMany`의 현재 임계값(5분 내 3회 이상)은 초기값이다. 운영 데이터를 충분히 수집한 뒤 정상 재시작 빈도와 오탐 여부를 검토하여 시간 범위, 횟수, `for` 지속 시간을 조정한다.
- [ ] Slack·이메일 등 Alertmanager 외부 receiver를 연결하고, 서비스에 영향을 주지 않는 테스트 워크로드로 `PodRestartTooMany`의 Firing과 실제 메시지 수신을 검증한다.

## 트러블슈팅 이력

| 챕터 | 문제 | 해결 |
|------|------|------|
| ch2.5 | Kubernetes Engine API가 비활성화되어 클러스터 조회 실패 | `container.googleapis.com` 활성화 후 GKE 클러스터 생성 |
| ch2.5 | 생성 직후 GatewayClass가 보이지 않음 | 컨트롤러 반영을 기다린 뒤 4개 GatewayClass의 `ACCEPTED=True` 확인 |
| ch2.6 | Cloud Build API가 비활성화됨 | `cloudbuild.googleapis.com` 활성화 |
| ch2.6 | 기본 Compute 서비스 계정이 Cloud Build 소스 버킷을 읽지 못함 | 해당 서비스 계정에 `roles/cloudbuild.builds.builder` 부여 |
| ch2.6 | 같은 `v0.1.0` 태그 재빌드 후 실행 중 Pod와 Registry digest 불일치 | `imagePullPolicy: Always` 적용 후 최신 digest로 rollout 및 API 재검증 |
| ch2.6 | Spot VM 노드가 모두 사라져 API Pod 2개가 `Pending` 상태로 유지됨 | `default-pool`을 2대로 resize한 뒤 노드 `Ready`, Deployment 2/2 및 Service 엔드포인트 복구 확인 |
| ch3.4 | 조직 정책이 서비스 계정 키 생성을 차단 | Workload Identity Federation으로 전환해 GitHub OIDC 단기 인증 구성 |
| ch3.4 | 최초 CI 실행에서 IAM 바인딩 전파 지연으로 이미지 push 403 | 동일 Job 재실행 후 테스트·빌드 및 `sha-962fe0e` 이미지 push 성공 |
| ch3.5 | `set -e` 환경에서 변경을 확인하는 `git diff --quiet &&` 구문이 exit 1로 Step 중단 | 명시적인 `if git diff --quiet; then` 조건문으로 수정 후 전체 CI-CD 성공 |
| ch4.2 | node-exporter 10m CPU request가 99% 예약된 노드에 스케줄되지 않음 | request를 5m으로 축소하고 Helm values에 반영해 DaemonSet 2/2 복구 |
| ch4.2 | GKE CoreDNS가 metrics 포트 9153을 노출하지 않아 target 2개 Down | chart의 `coreDns.enabled=false`로 불필요한 scrape target 제거, active target 16/16 Up 확인 |
| ch4.3 | 노드 CPU 요청량이 약 938m/940m라 Fluent Bit Pod가 Pending | 실제 사용량과 limit를 확인하고 CPU request를 1m으로 낮춰 DaemonSet 2/2 배치 |
| ch4.3 | 기존 컨테이너 로그를 처음부터 읽자 Loki가 오래된 timestamp를 `entry too far behind`로 거부 | 과거 재수집 옵션을 제거하고 설치 이후 새 로그를 수집하는 안정적인 tail 구성으로 복원 |
| ch5.2 | 리전 외부 Gateway가 active proxy-only subnet 부재로 Programmed=False | `default` VPC의 서울 리전에 `proxy-only-subnet`(`172.16.0.0/23`)을 생성해 외부 IP와 로드밸런서 프로비저닝 완료 |
| ch5.2 | Gateway 생성 직후 백엔드가 `no healthy upstream`으로 HTTP 503 반환 | HealthCheckPolicy 적용 상태와 GCP backend health를 확인하고 endpoint 2개가 Healthy로 전환된 뒤 HTTP 200 재검증 |
| ch5.3 | 로컬 환경에 Go 도구가 없어 `gofmt`와 `go test ./...`를 직접 실행하지 못함 | GitHub Actions run `30894204759`에서 테스트·이미지 빌드·매니페스트 갱신 전체 성공 확인 |
| ch5.3 | 최초 Deployment→Rollout 리소스 종류 전환 중 외부 헬스 확인 1회가 5초 내 응답하지 않음 | 초기 전환 완료 후 Blue→Green 버전 배포 구간을 별도로 관찰해 Blue 유지, Green 준비, 자동 승격 및 `v0.2.0` HTTP 200 확인 |
| 운영 복구 | 재구축 시 두 API replica가 같은 노드에 배치되어 CPU 예약량이 940m/940m에 도달하고 Fluent Bit 한 Pod가 Pending | Fluent Bit CPU request를 `1m`에서 `0`으로 낮추고 2노드 DaemonSet 2/2 Ready 및 Loki 수집 시작 확인 |
| ch6.1 | 2노드에서 Blue/Green active·preview와 Valkey가 동시에 뜰 때 CPU 부족 위험 | GitOps 매니페스트의 replicas를 2→1로 낮추고 Valkey `resourcesPreset=none`, request `50m/64Mi` 적용 |
| ch6.2 | Workload Identity metadata server와 managed CSI가 노드당 220m를 추가해 Valkey와 관측성 Pod가 Pending | 관측성 CPU request를 0으로 축소하고 API 25m·Valkey 10m 및 API/Valkey anti-affinity를 적용해 전체 Pod Running 복구 |
| ch6.2 | 실패한 Valkey Helm RollingUpdate가 기존 Pending Pod의 50m revision을 반복 생성 | StatefulSet 템플릿 10m 확인 후 승인받아 Pending Pod를 재생성하고 Helm revision 4 `deployed`로 정상화 |
| ch6.3 | replica 1과 stable Service 기반 HTTPRoute에서는 setWeight 20/50/80이 실제 Gateway 요청 비율을 정밀 분할하지 못함 | Canary 단계·pause·stable/canary Service 전환을 검증하고, 실제 가중 트래픽은 ch7 replica 확장과 Gateway traffic routing 연동 과제로 기록 |
| ch7.3 | namespace 리소스를 `notiflex-smb`에서 wave 0 `bootstrap`으로 한 번에 이동하자 기존 앱 prune가 `notiflex` namespace를 삭제 | bootstrap이 namespace를 재생성하고 Secret Manager의 기존 credential로 Valkey Secret 복구, Valkey·API·Gateway 재동기화. 기존 Valkey PVC와 `notiflex:id` 데이터는 손실되어 새 PVC에서 ID 1부터 재시작했으며, 이후 namespace는 bootstrap 단독 소유와 `Prune=false`로 보호 |
| ch7.4 | Enterprise API 최초 기동 시 Workload Identity IAM 전파 동안 약 2분 Pending | IAM 전파 후 자동 Running 전환과 CSI mount를 확인. Enterprise `/id`가 2, 이어 SMB `/id`가 3을 반환해 두 tenant가 같은 Valkey 키를 공유함을 명시 |
| ch8.1 | 로컬 Go 미설치와 첫 임시 Go ZIP 압축 해제 중 timeout으로 표준 라이브러리 일부 누락 | 공식 Go 1.25.10 ZIP의 SHA-256을 검증하고 새 임시 디렉터리에 완전히 해제한 뒤 `gofmt`·`go mod tidy`·`go test ./...` 성공 |
| ch8.2 | Tempo를 가이드 기본값인 `ops-pool`에 배치하자 e2-small 노드 메모리 실사용 95%·request 88% 도달 | 추가 비용 없이 여유 있는 `worker-pool`로 옮겨 메모리 압박 위험을 낮추고 Tempo Ready와 Trace 수집을 재검증 |
| ch8.3 | `curlimages/curl`이 사용자 이름 `curl_user`를 선언해 kubelet이 `runAsNonRoot`를 숫자 UID로 검증하지 못하고 Job 생성 실패 | 이미지의 비루트 사용자는 유지하고 해당 검사만 제거했다. 승인 후 실패 Job 2개를 삭제해 `Forbid` 차단을 해소하고 다음 예약 Job의 HTTP 200 완료를 확인 |
