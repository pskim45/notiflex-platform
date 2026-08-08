# Notiflex Platform 온보딩

이 문서는 새 팀원이 Notiflex의 구조와 운영 방식을 이해하고, 안전하게 조회·개발·배포를 시작하기 위한 안내서다. 2026-08-08 비용 중단으로 현재 클러스터는 삭제된 상태다. 클러스터 관련 표는 삭제 직전 스냅샷이자 복구 목표이며, 먼저 [`docs/shutdown-recovery.md`](docs/shutdown-recovery.md)로 환경을 복구한 뒤 조회 명령을 실행한다.

## 1. 먼저 읽을 문서

다음 순서로 읽으면 현재 상태와 변경 이유를 빠르게 이해할 수 있다.

1. [`AGENTS.md`](AGENTS.md): 고정 GCP 대상, 작업 및 안전 규칙
2. [`README.md`](README.md): 개발·배포·접근 방법
3. [`claude-context/architecture.md`](claude-context/architecture.md): 현재 컴포넌트와 연결 관계
4. [`JOURNEY.md`](JOURNEY.md): 실제 진행 상태, 버전, 장애 해결 이력
5. [`docs/architecture-decisions.md`](docs/architecture-decisions.md): 기술 선택 이유와 대안
6. [`command-guardrails/`](command-guardrails/): 위험 작업의 확인·승인·검증 절차

모든 `kubectl` 명령에는 `--context gke-sysnet4admin_book_gitaiops`를 명시한다. Secret 값, 토큰, kubeconfig와 서비스 계정 키는 Git에 기록하거나 작업 로그에 남기지 않는다. 삭제·rollback·Canary abort 같은 작업은 정확한 대상과 영향을 확인하고 사용자 승인을 받은 뒤 실행한다.

## 2. 개발 환경 준비

필요한 도구는 Git, Go, Docker, Google Cloud CLI, `kubectl`, `gh`다. Argo Rollouts CLI 플러그인은 Canary 상세 조회와 제어에 권장한다.

```powershell
git --version
go version
docker version
gcloud version
kubectl --context gke-sysnet4admin_book_gitaiops version --client
gh --version
gcloud config list
kubectl --context gke-sysnet4admin_book_gitaiops config get-contexts
```

고정 환경은 다음과 같다.

| 항목 | 값 |
|---|---|
| GCP 프로젝트 | `project-10edc337-9677-4dfc-91a` |
| 리전 / 존 | `asia-northeast3` / `asia-northeast3-a` |
| GKE 클러스터 | `notiflex-cluster` |
| kubectl context | `gke-sysnet4admin_book_gitaiops` |
| Artifact Registry | `asia-northeast3-docker.pkg.dev/project-10edc337-9677-4dfc-91a/notiflex` |
| 기본 브랜치 | `main` |

현재 로컬 설정이 다르면 임의로 바꾸지 말고 차이를 먼저 확인한다.

## 3. 저장소 구조

```text
notiflex-platform/
├── app/                    # Go API, 테스트, Dockerfile
├── k8s/
│   ├── bootstrap/          # namespace bootstrap
│   ├── smb/                # SMB API, Gateway, CronJob, CSI
│   ├── enterprise/         # Enterprise API, RBAC, quota, CSI
│   ├── kafka/              # KafkaNodePool, Kafka, KafkaTopic
│   └── monitoring/         # PrometheusRule
├── argocd/
│   ├── root-app.yaml       # App of Apps 진입점
│   └── apps/               # 하위 Application 11개
├── helm-values/            # Helm 앱의 Git 관리 values
├── monitoring/             # Grafana dashboard와 Loki·Tempo datasource
├── claude-context/         # 현재 아키텍처 스냅샷
├── command-guardrails/     # 위험 운영 작업 절차
├── docs/                   # ADR와 종료·복구 런북
├── .github/workflows/      # GitHub Actions CI/CD
├── .agents/skills/         # 저장소 전용 update-docs 워크플로
├── AGENTS.md               # AI 에이전트 및 작업 안전 규칙
└── JOURNEY.md              # 진행·버전·리소스·트러블슈팅 기록
```

`.claude/`는 현재 없다. ch7에서 로컬 권한 규칙을 실험한 뒤 되돌렸으며, 개인별 `.claude/settings.local.json`을 팀 공통 정책으로 간주하지 않는다. 공통 규칙은 `AGENTS.md`와 `command-guardrails/`에 둔다.

## 4. 마지막 검증 클러스터와 복구 목표

### 노드풀

| 노드풀 | 머신 타입 | 노드 | 용도 |
|---|---|---:|---|
| `default-pool` | `e2-medium` | 2 | Argo CD, Prometheus, Loki와 컨트롤러 |
| `api-pool` | `e2-medium` | 1 | SMB·Enterprise API |
| `worker-pool` | `e2-standard-2` | 1 | Valkey, Kafka, Tempo |
| `ops-pool` | `e2-small` | 1 | 헬스체크 CronJob |

모든 노드는 Spot VM이며 현재 `Ready`다. 다음 명령으로 다시 확인한다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops get nodes -L cloud.google.com/gke-nodepool
kubectl --context gke-sysnet4admin_book_gitaiops top nodes
```

### namespace별 Pod

| Namespace | Pod 수 | 역할 |
|---|---:|---|
| `argocd` | 7 | GitOps controller, server, repository, Redis, Dex |
| `argo-rollouts` | 1 | Canary controller |
| `notiflex` | 5 | SMB API, Valkey, 완료된 healthcheck Job |
| `enterprise` | 1 | Enterprise API |
| `kafka` | 3 | Strimzi, Kafka broker/controller, entity operator |
| `monitoring` | 18 | Prometheus, Grafana, Alertmanager, Loki, Fluent Bit, Tempo |
| `kube-system` | 57 | GKE 네트워크, DNS, CSI, metrics, logging DaemonSet |
| `gmp-system` | 6 | GKE managed Prometheus collector/operator |
| `gke-managed-cim` | 1 | managed kube-state-metrics |

Pod 수에는 완료된 Job이 포함되며 controller 재배포에 따라 달라질 수 있다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops get pods -A
kubectl --context gke-sysnet4admin_book_gitaiops get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded
```

### 핵심 상태 확인

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops get applications -n argocd
kubectl --context gke-sysnet4admin_book_gitaiops get rollouts -A
kubectl --context gke-sysnet4admin_book_gitaiops get kafka,kafkatopic -n kafka
kubectl --context gke-sysnet4admin_book_gitaiops get cronjob,jobs -n notiflex -l app.kubernetes.io/name=notiflex-healthcheck
kubectl --context gke-sysnet4admin_book_gitaiops get pvc -A
```

현재 Argo CD Application 12개는 모두 `Synced/Healthy`다. API는 tenant별 1 replica이며 Kafka·Valkey도 단일 인스턴스이므로 프로덕션 고가용성 구성은 아니다.

## 5. 서비스 접근

### Argo CD UI

터미널 1에서 포트 포워딩한다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops port-forward service/argocd-server -n argocd 8080:443
```

브라우저에서 `https://localhost:8080`에 접속한다. 사용자 이름은 `admin`이다. 초기 Secret이 남아 있는 현재 환경에서는 다음 명령으로 로컬 변수에 비밀번호를 디코딩한다. 값을 채팅, 이슈, 문서 또는 Git에 복사하지 않는다.

```powershell
$encodedPassword = kubectl --context gke-sysnet4admin_book_gitaiops get secret argocd-initial-admin-secret -n argocd -o jsonpath='{.data.password}'
[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($encodedPassword))
Remove-Variable encodedPassword
```

초기 Secret이 없다면 현재 관리자에게 접근 권한 또는 비밀번호 재설정 절차를 요청한다.

### Grafana

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops port-forward service/kube-prometheus-grafana -n monitoring 3000:80
```

`http://localhost:3000`에 접속한다. 사용자 이름은 `admin`이며 비밀번호는 `kube-prometheus-grafana` Secret에서 로컬로 조회한다.

```powershell
$encodedPassword = kubectl --context gke-sysnet4admin_book_gitaiops get secret kube-prometheus-grafana -n monitoring -o jsonpath='{.data.admin-password}'
[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($encodedPassword))
Remove-Variable encodedPassword
```

Grafana Explore에서 데이터소스를 목적에 맞게 선택한다.

| 데이터소스 | 용도 | 시작 예시 |
|---|---|---|
| Prometheus | CPU·메모리·재시작·Kafka metrics | `up` |
| Loki | Kubernetes stdout/stderr 로그 | `{namespace="notiflex"}` |
| Tempo | HTTP→Valkey→Kafka Trace | TraceID 조회 또는 service `notiflex-api` 검색 |

### API

현재 Gateway IP는 `35.216.50.229`이며 HTTP만 사용한다.

```powershell
$gatewayIp = kubectl --context gke-sysnet4admin_book_gitaiops get gateway notiflex-gateway -n notiflex -o jsonpath='{.status.addresses[0].value}'
Invoke-RestMethod "http://$gatewayIp/health"
Invoke-RestMethod "http://$gatewayIp/version"
Invoke-RestMethod "http://$gatewayIp/id"
```

Enterprise API는 외부 Gateway에 연결되지 않는다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops port-forward service/notiflex-api -n enterprise 8081:80
Invoke-RestMethod "http://localhost:8081/health"
```

## 6. 개발과 배포 플로우

로컬에서는 먼저 테스트한다.

```powershell
Set-Location app
go test ./...
Set-Location ..
```

애플리케이션 변경의 배포 흐름은 다음과 같다.

```text
main에 app/** push
  → GitHub Actions
     → Docker build 중 go test ./...
     → Artifact Registry에 api:sha-<7자리> push
     → SMB·Enterprise Rollout 이미지 태그 갱신
     → github-actions[bot] 배포 commit
  → Argo CD auto-sync
  → Argo Rollouts Canary 20% → 50% → 80% → 100%
```

일반 YAML은 `k8s/`, Helm 앱은 `argocd/apps/`의 chart version과 `helm-values/`를 수정한다. 일상 변경에서 `kubectl apply`나 `helm upgrade`로 Git을 우회하지 않는다.

변경 전후 최소 확인 항목:

```powershell
git status --short --branch
git diff --check
kubectl --context gke-sysnet4admin_book_gitaiops get applications -n argocd
kubectl --context gke-sysnet4admin_book_gitaiops get rollouts -A
```

## 7. 자주 묻는 질문

### Q1. Canary를 중단하려면?

먼저 Rollout 상태, 현재 step과 stable revision을 조회한다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops get rollout notiflex-api -n notiflex -o yaml
```

abort는 배포에 영향을 주는 작업이다. 정확한 namespace와 영향 범위를 확인하고 승인받은 뒤 Argo Rollouts 플러그인으로 실행한다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops argo rollouts abort notiflex-api -n notiflex
kubectl --context gke-sysnet4admin_book_gitaiops argo rollouts status notiflex-api -n notiflex
```

Git의 이미지 선언을 수정하지 않으면 Argo CD가 다시 배포할 수 있으므로 원인 수정과 Git 변경을 함께 처리한다.

### Q2. API 로그는 어떻게 검색하나?

빠른 확인은 Pod 로그를 사용한다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops logs -n notiflex -l app.kubernetes.io/name=notiflex-api --tail=100
```

장기·다중 Pod 검색은 Grafana Loki에서 다음 LogQL을 사용한다.

```logql
{namespace="notiflex"} |= "Kafka"
```

### Q3. Trace는 어떻게 추적하나?

API 응답을 만든 뒤 Grafana Explore에서 Tempo를 선택한다. 로그에 TraceID가 있으면 직접 조회하고, 없으면 service name `notiflex-api`와 시간 범위로 검색한다. 정상 Trace에는 `notiflex-api`, `valkey.incr`, `kafka.produce`, `kafka.consume` Span이 연결된다.

### Q4. Kafka Topic을 추가하려면?

`k8s/kafka/kafka-cluster.yaml`에 `KafkaTopic`을 선언하거나 topic별 파일로 분리하고 `strimzi.io/cluster: notiflex-kafka` 라벨을 지정한다. server dry-run과 diff를 검토한 뒤 push하고 `kafka` Application의 동기화를 확인한다. Topic 삭제는 [`command-guardrails/kafka-topic-delete.md`](command-guardrails/kafka-topic-delete.md)를 반드시 따른다.

### Q5. 새 tenant를 추가하려면?

1. `k8s/<tenant>/`에 Rollout, stable·preview Service, ServiceAccount, CSI, RBAC, quota를 선언한다.
2. `k8s/bootstrap/namespaces.yaml`에 tenant namespace를 추가한다.
3. `argocd/apps/notiflex-<tenant>.yaml`을 wave 2 Application으로 추가한다.
4. KSA↔GSA Workload Identity IAM binding과 Secret Manager 접근을 최소 권한으로 부여한다.
5. ClusterIP, CSI mount, RBAC deny/allow, quota와 공유 데이터 범위를 검증한다.

현재 SMB와 Enterprise는 Valkey와 credential을 공유하므로 새 tenant도 별도 데이터 경계가 필요하면 고객별 상태와 credential을 설계해야 한다. tenant 삭제는 [`command-guardrails/tenant-namespace-delete.md`](command-guardrails/tenant-namespace-delete.md)를 따른다.

### Q6. Alertmanager 알림은 어디서 확인하나?

현재 `PodRestartTooMany` 규칙은 로드됐지만 Slack·이메일 receiver는 연결되지 않았다. UI는 다음과 같이 확인한다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops port-forward service/kube-prometheus-kube-prome-alertmanager -n monitoring 9093:9093
```

브라우저에서 `http://localhost:9093`에 접속한다. 실제 외부 메시지 수신은 receiver 구성 후 별도로 검증해야 한다.

### Q7. Helm으로 설치된 앱은 어떻게 변경하나?

`helm upgrade`를 직접 실행하지 않는다. `argocd/apps/<app>.yaml`의 고정 chart version 또는 `helm-values/<app>.yaml`을 수정하고 push한다. Argo CD가 외부 chart와 Git values를 결합해 적용한다.

### Q8. 수동 CronJob을 실행하려면?

예약 Job과 충돌, 외부 호출과 데이터 변경 여부를 먼저 확인하고 [`command-guardrails/cronjob-manual-run.md`](command-guardrails/cronjob-manual-run.md)를 따른다. 수동 Job은 CronJob history 정리 대상이 아니므로 삭제도 별도 승인을 받는다.

## 8. 첫 변경 체크리스트

- [ ] `AGENTS.md`와 현재 아키텍처를 읽었다.
- [ ] 고정 GCP 프로젝트·zone·kubectl context가 일치한다.
- [ ] `git status`에서 기존 사용자 변경을 확인했다.
- [ ] 변경이 일반 YAML인지 Helm values인지 소유 경로를 확인했다.
- [ ] Go 변경이면 `go test ./...`를 통과했다.
- [ ] 매니페스트 diff와 고정 이미지·chart 버전을 확인했다.
- [ ] Secret 값이나 장기 credential을 추가하지 않았다.
- [ ] push 후 CI, Argo CD, Rollout과 실제 API 결과를 확인했다.
- [ ] 아키텍처·운영 결정이 바뀌었다면 JOURNEY·ADR·컨텍스트를 갱신했다.

작업 중 인프라를 중단하거나 재구축해야 한다면 [`docs/shutdown-recovery.md`](docs/shutdown-recovery.md)를 먼저 읽는다.
