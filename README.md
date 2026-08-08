# Notiflex Platform

Notiflex는 B2B 환경에서 여러 채널의 알림을 안정적으로 전달하기 위한 알림 SaaS 실습 프로젝트입니다. 이 저장소는 Go 애플리케이션을 컨테이너로 빌드하고 GKE에 배포한 뒤, GitOps와 관측 가능성, 점진적 배포까지 확장하는 과정을 담습니다.

## 현재 상태

Notiflex API `v0.5.0`가 GKE의 전용 `api-pool`에서 실행 중입니다. SMB와 Enterprise는 각각 `notiflex`·`enterprise` namespace의 독립 Canary Rollout으로 배포되며, RBAC와 ResourceQuota도 분리됩니다. 두 환경은 현재 Valkey와 Secret Manager credential을 공유하므로 배포·권한·자원은 분리되지만 데이터는 분리되지 않습니다. SMB API는 Kafka로 알림 이벤트를 비동기 전달하고 두 API는 Tempo로 Trace를 보냅니다. Argo CD App of Apps가 전체 구성을 Git에서 관리하며 root 포함 12개 Application은 `Synced/Healthy`입니다.

## 기술 스택

- 애플리케이션: Go 표준 라이브러리
- 컨테이너: 멀티 스테이지 빌드, `scratch` 런타임 이미지
- 클라우드: Google Cloud Platform
- 런타임: GKE Standard(영역 클러스터, Spot VM)
- 이미지 저장소: Artifact Registry
- GitOps: Argo CD
- 비동기 메시징: Strimzi, Kafka
- 관측 가능성: Prometheus, Grafana, Loki, Fluent Bit, Tempo, OpenTelemetry
- 상태 공유: Valkey standalone
- 시크릿: GCP Secret Manager, Workload Identity, GKE managed Secrets Store CSI
- 배포 전략: Rolling Update → Blue/Green → Argo Rollouts Canary

## 디렉터리 구조

```text
notiflex-platform/
├── app/                  # Go 애플리케이션과 Dockerfile
├── claude-context/
│   └── architecture.md   # 현재 컴포넌트·연결·설정 아키텍처 스냅샷
├── argocd/
│   ├── root-app.yaml     # 모든 하위 Application의 GitOps 진입점
│   └── apps/             # API와 Helm chart별 Application 선언
├── k8s/
│   ├── kafka/            # Strimzi Kafka cluster와 Topic
│   ├── monitoring/       # PrometheusRule 등 관측성 매니페스트
│   ├── enterprise/       # Enterprise tenant API·identity·RBAC·quota
│   └── smb/              # 애플리케이션 Kubernetes 매니페스트
├── helm-values/          # 관측성 Helm chart 경량 설정
├── monitoring/           # Grafana 대시보드와 데이터소스
├── docs/
│   ├── architecture-decisions.md # 시간순 아키텍처 결정 기록
│   └── shutdown-recovery.md      # 비용 중단 후 환경 복구 런북
├── command-guardrails/   # 위험 운영 작업의 확인·승인·검증 절차
├── .github/
│   └── workflows/        # GitHub Actions 워크플로
├── AGENTS.md             # Codex 등 AI 에이전트용 프로젝트 지침
└── README.md
```

## GCP 환경

| 항목 | 값 |
|---|---|
| 프로젝트 ID | `project-10edc337-9677-4dfc-91a` |
| 리전 | `asia-northeast3` (서울) |
| 존 | `asia-northeast3-a` |
| GKE 클러스터 | `notiflex-cluster` |
| kubectl 컨텍스트 | `gke-sysnet4admin_book_gitaiops` |
| Artifact Registry | `asia-northeast3-docker.pkg.dev/project-10edc337-9677-4dfc-91a/notiflex` |
| Kubernetes 네임스페이스 | `notiflex` |

## 개발 준비

필요한 도구는 Git, Go, Google Cloud CLI, Docker, `kubectl`입니다. 먼저 현재 로컬 설정이 위 표와 일치하는지 확인합니다.

```bash
gcloud config list
gcloud auth configure-docker asia-northeast3-docker.pkg.dev
kubectl config get-contexts
```

GKE에 접근할 때는 다른 클러스터를 실수로 변경하지 않도록 모든 명령에 컨텍스트를 명시합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get nodes
kubectl --context gke-sysnet4admin_book_gitaiops get pods -n notiflex
```

## 로컬 실행과 배포

애플리케이션을 로컬에서 검증하고 실행합니다.

```bash
cd app
go test ./...
VALKEY_ADDR=localhost:6379 \
VALKEY_PASSWORD='<LOCAL_VALKEY_PASSWORD>' \
go run .
```

클러스터에서는 비밀번호 환경변수 대신 GKE CSI가 마운트한 `VALKEY_PASSWORD_FILE=/mnt/secrets/valkey-password`를 사용합니다.

필요하면 다음 명령으로 Cloud Build를 수동 실행해 이미지를 Artifact Registry에 게시할 수 있습니다.

```bash
gcloud builds submit app/ \
  --project=project-10edc337-9677-4dfc-91a \
  --tag=asia-northeast3-docker.pkg.dev/project-10edc337-9677-4dfc-91a/notiflex/api:<VERSION>
```

`main` 브랜치에서 `app/**`가 변경되면 `.github/workflows/ci.yaml`이 자동으로 테스트·빌드·푸시를 수행합니다. GCP 인증은 장기 서비스 계정 키 대신 Workload Identity Federation을 사용하고, 이미지는 `sha-<커밋 앞 7자리>` 태그로 게시합니다. 빌드 성공 후 워크플로가 `k8s/smb/rollout.yaml`의 이미지 태그를 커밋하면 Argo CD가 변경을 감지해 자동 배포합니다. CI는 클러스터에 직접 접근하지 않습니다.

최초 구성에서는 `root-app` 하나만 적용합니다. 이후 root가 `argocd/apps/`의 하위 Application을 등록하고, 각 앱이 일반 YAML 또는 외부 Helm chart와 Git의 values를 자동 동기화합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops apply -f argocd/root-app.yaml
kubectl --context gke-sysnet4admin_book_gitaiops get application -n argocd
```

root sync 순서는 Application annotation으로 고정한다.

```text
wave 0: bootstrap — notiflex·enterprise·monitoring·kafka namespace
wave 1: valkey, kube-prometheus, loki, strimzi — 데이터·관측성 backend와 Kafka operator
wave 2: kafka, tempo, fluent-bit, monitoring-config, notiflex-smb, notiflex-enterprise — 메시징·트레이싱·수집기·설정·tenant API
```

## 멀티테넌시

`enterprise`는 별도 namespace, Canary Rollout, stable·preview ClusterIP Service, ServiceAccount/CSI identity, 읽기 전용 `tenant-reader` RBAC와 ResourceQuota를 갖습니다. API는 SMB와 같이 `api-pool`에 배치되지만 Enterprise Service는 외부 Gateway에 연결하지 않아 클러스터 내부에서만 접근할 수 있습니다.

두 API는 `valkey-primary.notiflex.svc.cluster.local:6379`와 동일한 Secret Manager `valkey-password`를 사용합니다. 따라서 현재 구성은 실행 환경 격리이며 고객 데이터 격리가 아닙니다. 실제 검증에서도 Enterprise `/id` 다음 SMB `/id` 값이 연속 증가했습니다. 강한 고객 데이터 격리가 필요하면 고객별 Valkey와 credential을 분리해야 합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get application notiflex-enterprise -n argocd
kubectl --context gke-sysnet4admin_book_gitaiops get rollout,pod,resourcequota -n enterprise
kubectl --context gke-sysnet4admin_book_gitaiops auth can-i get pods --as=system:serviceaccount:enterprise:tenant-reader -n enterprise
kubectl --context gke-sysnet4admin_book_gitaiops auth can-i delete pods --as=system:serviceaccount:enterprise:tenant-reader -n enterprise
kubectl --context gke-sysnet4admin_book_gitaiops port-forward service/notiflex-api -n enterprise 8081:80
```

GitOps 동기화 상태는 다음과 같이 확인합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get application -n argocd
```

## 중앙 로그 수집

Loki chart `7.2.0`(Loki `3.6.11`)과 Fluent Bit chart `0.57.9`(Fluent Bit `5.0.9`)는 각각 Argo CD Application으로 `monitoring` 네임스페이스에 설치됩니다. Loki는 5Gi 영구 볼륨을 사용하는 SingleBinary 구성이고, Fluent Bit은 각 노드에서 새 컨테이너 stdout/stderr 로그를 수집합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get application \
  loki fluent-bit monitoring-config -n argocd
```

일상 변경에서는 `helm upgrade`를 직접 실행하지 않는다. `argocd/apps/*.yaml`의 chart 버전이나 `helm-values/*.yaml`을 수정해 push하면 Argo CD가 적용한다.

Grafana의 **Explore**에서 데이터소스로 `Loki`를 선택한 뒤 `{namespace="monitoring"}` 또는 `{namespace="notiflex"}` LogQL 쿼리를 실행합니다. Fluent Bit 설치 이후 새로 출력된 로그부터 조회됩니다.

## 외부 API 접근

GKE Gateway API가 리전 외부 HTTP 로드밸런서를 구성합니다. `notiflex-gateway`의 `HTTPRoute`는 모든 경로를 `notiflex-api` Service로 전달하며, `HealthCheckPolicy`는 Pod의 `/health:8080`을 확인합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get gateway,httproute -n notiflex
curl http://35.216.50.229/health
curl http://35.216.50.229/id
curl http://35.216.50.229/version
```

현재 외부 IP는 `35.216.50.229`이며 세 엔드포인트 모두 HTTP 200 응답을 확인했습니다. 리전 외부 Gateway에 필요한 `proxy-only-subnet`은 `default` VPC의 `asia-northeast3` 리전에 `172.16.0.0/23` 대역으로 구성되어 있습니다. 현재 리스너는 HTTP이므로 민감한 운영 트래픽을 받기 전에는 도메인과 TLS 인증서를 추가해야 합니다.

## Canary 배포

Argo Rollouts `v1.9.1`이 `argo-rollouts` 네임스페이스에서 실행됩니다. `notiflex-api` Rollout은 새 ReplicaSet에 20%, 50%, 80% weight를 순서대로 설정하고 각 단계에서 30초간 관찰한 뒤 100%로 승격합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops create namespace argo-rollouts
kubectl --context gke-sysnet4admin_book_gitaiops apply --server-side \
  -n argo-rollouts \
  -f https://github.com/argoproj/argo-rollouts/releases/download/v1.9.1/install.yaml

kubectl --context gke-sysnet4admin_book_gitaiops get rollout notiflex-api -n notiflex
kubectl --context gke-sysnet4admin_book_gitaiops get rs,svc -n notiflex
curl http://35.216.50.229/version
```

현재 `v0.5.0`/`sha-eee42f5`가 step 6(100%)에서 Healthy입니다. 현재 replica가 1이고 HTTPRoute가 stable Service만 참조하므로 이 weight는 Rollout 진행 단계이며 Gateway 수준의 정밀한 요청 비율은 아닙니다. 실제 20/50/80 트래픽 분할에는 replica 확장과 Gateway API traffic routing 연동이 필요합니다.

## 비동기 메시징과 분산 트레이싱

Strimzi `0.51.0`이 Kafka `4.1.0` KRaft 단일 브로커와 `notifications` Topic(3 partitions)을 관리합니다. SMB API는 `/id` 처리 후 이벤트를 Kafka에 기록하고 Consumer Group이 비동기로 처리합니다. Kafka와 Tempo는 `worker-pool`에 배치됩니다.

API는 OpenTelemetry OTLP gRPC로 Tempo `2.9.0`에 Trace를 전송합니다. Grafana의 Tempo 데이터소스에서 `notiflex-api` HTTP 요청과 `valkey.incr`, `kafka.produce`, `kafka.consume` Span의 연결을 조회할 수 있습니다.

`notiflex-healthcheck` CronJob은 `ops-pool`에서 5분마다 내부 SMB API `/health`를 호출하도록 구성했습니다. 수정된 템플릿의 수동 Job은 HTTP 200을 확인했지만, 수정 전 생성된 실패 Job이 `Forbid` 정책으로 후속 예약 실행을 막고 있어 해당 Job을 승인 후 정리해야 합니다. 수동 실행과 정리는 예약 실행과의 중복 및 부작용을 확인한 뒤 [CronJob 수동 실행 절차](command-guardrails/cronjob-manual-run.md)를 따릅니다.

## 메트릭 모니터링

Prometheus, Grafana, Alertmanager, kube-state-metrics, node-exporter는 Argo CD의 `kube-prometheus` Application이 chart `88.1.3`으로 `monitoring` 네임스페이스에 설치합니다. 재현 가능한 설정은 `helm-values/kube-prometheus.yaml`에 있습니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get application \
  kube-prometheus monitoring-config -n argocd
```

Grafana 접속용 포트 포워딩:

```bash
kubectl --context gke-sysnet4admin_book_gitaiops \
  port-forward service/kube-prometheus-grafana -n monitoring 3000:80
```

브라우저에서 `http://localhost:3000`에 접속합니다. 사용자 이름은 `admin`이며 비밀번호는 `kube-prometheus-grafana` Secret에서 조회합니다. Notiflex 대시보드의 CPU·메모리·재시작 패널은 현재 동작하고, HTTP 요청 패널은 API가 `/metrics`에서 `http_requests_total`을 노출한 뒤 데이터가 표시됩니다.

로컬 포트 포워딩 후 API를 확인할 수 있습니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops port-forward service/notiflex-api -n notiflex 8080:80
curl http://localhost:8080/health
curl http://localhost:8080/version
curl http://localhost:8080/id
```

## 알림

Alertmanager는 `kube-prometheus-stack`에 포함되어 실행됩니다. `PodRestartTooMany` 규칙은 `notiflex` 네임스페이스에서 같은 컨테이너가 5분 동안 3회 이상 재시작하고 그 상태가 1분간 지속되면 경고를 발생시킵니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops apply \
  -f k8s/monitoring/pod-restart-alert.yaml

kubectl --context gke-sysnet4admin_book_gitaiops get \
  prometheusrule pod-restart-alert -n monitoring
```

규칙은 Prometheus에서 `health: ok`로 로드된 것을 확인했습니다. 현재 외부 receiver가 없으므로 Firing 상태가 되어도 Slack이나 이메일로 전달되지는 않습니다. 수신 채널을 연결한 뒤 서비스에 영향을 주지 않는 테스트 워크로드로 Firing과 실제 메시지 수신을 추가 검증해야 합니다. 임계값은 초기값이며 운영 데이터와 오탐 빈도를 보고 조정합니다.

## AI 에이전트 사용

Codex를 포함한 AI 에이전트는 작업 전에 [AGENTS.md](AGENTS.md)를 읽어야 합니다. 이 파일에는 프로젝트 컨텍스트, GCP 대상, 검증 및 안전 규칙이 정의되어 있습니다.

각 장을 마치면 저장소 전용 `$update-docs` 스킬로 그 시점의 코드·인프라와 모든 문서를 동기화하고 검증된 변경을 커밋합니다.

ch3 이후 주요 기술 선택의 근거와 검토한 대안은 [Architecture Decision Records](docs/architecture-decisions.md)에 시간 순서대로 누적합니다.

현재 클러스터의 컴포넌트, 연결 관계, 배포·credential·관측성 설정은 [현재 아키텍처](claude-context/architecture.md)에서 한눈에 확인할 수 있습니다.

비용 중단을 위해 실습 인프라를 삭제한 뒤 재구축할 때는 [GCP 실습 환경 종료 및 복구](docs/shutdown-recovery.md)를 따릅니다.

Kafka Topic 삭제, CronJob 수동 실행, 테넌트 Namespace 삭제처럼 영향이 큰 작업은 [운영 가드레일](command-guardrails/)의 사전 확인·승인·사후 검증 절차를 따릅니다.

```text
$update-docs
```

Codex에서는 `$update-docs`로 명시 호출하거나 `/skills`에서 선택합니다. 스킬은 [.agents/skills/update-docs/SKILL.md](.agents/skills/update-docs/SKILL.md)에 있으며, 실행할 때마다 새로 추가되거나 이름이 바뀐 문서까지 동적으로 탐색합니다. 푸시는 자동으로 수행하지 않으므로 필요하면 별도로 요청합니다.
