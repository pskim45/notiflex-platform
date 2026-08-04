# Notiflex Platform

Notiflex는 B2B 환경에서 여러 채널의 알림을 안정적으로 전달하기 위한 알림 SaaS 실습 프로젝트입니다. 이 저장소는 Go 애플리케이션을 컨테이너로 빌드하고 GKE에 배포한 뒤, GitOps와 관측 가능성, 점진적 배포까지 확장하는 과정을 담습니다.

## 현재 상태

Notiflex API `v0.2.2`가 GKE에 배포되어 있습니다. `/health` 상태 확인, 애플리케이션·Go·Pod 정보를 반환하는 `/version`, Pod별 순차 ID를 발급하는 `/id` API를 제공하며, Argo Rollouts Rollout은 replica 2개를 Blue/Green 전략으로 운영합니다. GitHub Actions가 `app/` 변경을 테스트하고 SHA 태그 이미지를 Artifact Registry에 게시한 뒤 Rollout 매니페스트를 자동 커밋하며, Argo CD `v3.3.6`이 이를 감지해 배포합니다. Prometheus와 Grafana가 메트릭을 제공하고, Loki와 Fluent Bit이 컨테이너 로그를 수집합니다. Alertmanager와 Pod 재시작 규칙은 구성됐으며 외부 알림 수신 채널은 아직 연결하지 않았습니다.

## 기술 스택

- 애플리케이션: Go 표준 라이브러리
- 컨테이너: 멀티 스테이지 빌드, `scratch` 런타임 이미지
- 클라우드: Google Cloud Platform
- 런타임: GKE Standard(영역 클러스터, Spot VM)
- 이미지 저장소: Artifact Registry
- GitOps: Argo CD
- 관측 가능성: Prometheus, Grafana, Loki, Fluent Bit (Tempo는 ch8 예정)
- 배포 전략: Rolling Update → Blue/Green → Canary

## 디렉터리 구조

```text
notiflex-platform/
├── app/                  # Go 애플리케이션과 Dockerfile
├── argocd/
│   └── notiflex-smb.yaml # Argo CD Application 선언
├── k8s/
│   ├── monitoring/       # PrometheusRule 등 관측성 매니페스트
│   └── smb/              # 애플리케이션 Kubernetes 매니페스트
├── helm-values/          # 관측성 Helm chart 경량 설정
├── monitoring/           # Grafana 대시보드와 데이터소스
├── docs/
│   └── architecture-decisions.md # 시간순 아키텍처 결정 기록
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
go run .
```

필요하면 다음 명령으로 Cloud Build를 수동 실행해 이미지를 Artifact Registry에 게시할 수 있습니다.

```bash
gcloud builds submit app/ \
  --project=project-10edc337-9677-4dfc-91a \
  --tag=asia-northeast3-docker.pkg.dev/project-10edc337-9677-4dfc-91a/notiflex/api:<VERSION>
```

`main` 브랜치에서 `app/**`가 변경되면 `.github/workflows/ci.yaml`이 자동으로 테스트·빌드·푸시를 수행합니다. GCP 인증은 장기 서비스 계정 키 대신 Workload Identity Federation을 사용하고, 이미지는 `sha-<커밋 앞 7자리>` 태그로 게시합니다. 빌드 성공 후 워크플로가 `k8s/smb/rollout.yaml`의 이미지 태그를 커밋하면 Argo CD가 변경을 감지해 자동 배포합니다. CI는 클러스터에 직접 접근하지 않습니다.

애플리케이션 매니페스트는 Argo CD가 `main` 브랜치의 `k8s/smb` 디렉터리에서 자동 동기화합니다. 초기 구성이나 수동 검증이 필요할 때는 다음 명령을 사용할 수 있습니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops apply -f k8s/smb/namespace.yaml
kubectl --context gke-sysnet4admin_book_gitaiops apply -f k8s/smb/
kubectl --context gke-sysnet4admin_book_gitaiops get rollout notiflex-api -n notiflex
```

GitOps 동기화 상태는 다음과 같이 확인합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get application notiflex-smb -n argocd
```

## 중앙 로그 수집

Loki chart `7.2.0`(Loki `3.6.11`)과 Fluent Bit chart `0.57.9`(Fluent Bit `5.0.9`)를 `monitoring` 네임스페이스에 설치합니다. Loki는 5Gi 영구 볼륨을 사용하는 SingleBinary 구성이고, Fluent Bit은 각 노드에서 새 컨테이너 stdout/stderr 로그를 수집합니다.

```bash
helm upgrade --install loki grafana/loki \
  --version 7.2.0 \
  --namespace monitoring \
  --values helm-values/loki.yaml

helm upgrade --install fluent-bit fluent/fluent-bit \
  --version 0.57.9 \
  --namespace monitoring \
  --values helm-values/fluent-bit.yaml

kubectl --context gke-sysnet4admin_book_gitaiops apply \
  -f monitoring/loki-datasource.yaml
```

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

## Blue/Green 배포

Argo Rollouts `v1.9.1`이 `argo-rollouts` 네임스페이스에서 실행됩니다. `notiflex-api` Rollout은 새 Green ReplicaSet을 `notiflex-api-preview` Service에 연결하고 30초 동안 준비 상태를 유지한 뒤, `notiflex-api` active Service를 자동으로 전환합니다. 이전 Blue ReplicaSet은 전환 30초 후 scale down됩니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops create namespace argo-rollouts
kubectl --context gke-sysnet4admin_book_gitaiops apply --server-side \
  -n argo-rollouts \
  -f https://github.com/argoproj/argo-rollouts/releases/download/v1.9.1/install.yaml

kubectl --context gke-sysnet4admin_book_gitaiops get rollout notiflex-api -n notiflex
kubectl --context gke-sysnet4admin_book_gitaiops get rs,svc -n notiflex
curl http://35.216.50.229/version
```

최근 재배포에서는 `v0.2.1` Blue가 외부 트래픽을 받는 동안 `v0.2.2` Green이 별도 ReplicaSet으로 준비되고, 자동 승격 후 active Service와 외부 응답이 `v0.2.2`로 전환되는 것을 확인했습니다.

## 메트릭 모니터링

Prometheus, Grafana, Alertmanager, kube-state-metrics, node-exporter는 `kube-prometheus-stack` chart `88.1.3`으로 `monitoring` 네임스페이스에 설치됩니다. 재현 가능한 설정은 `helm-values/kube-prometheus.yaml`에 있습니다.

```bash
helm upgrade --install kube-prometheus prometheus-community/kube-prometheus-stack \
  --version 88.1.3 \
  --namespace monitoring --create-namespace \
  --values helm-values/kube-prometheus.yaml

kubectl --context gke-sysnet4admin_book_gitaiops apply \
  -f monitoring/notiflex-dashboard.yaml
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

비용 중단을 위해 실습 인프라를 삭제한 뒤 재구축할 때는 [GCP 실습 환경 종료 및 복구](docs/shutdown-recovery.md)를 따릅니다.

```text
$update-docs
```

Codex에서는 `$update-docs`로 명시 호출하거나 `/skills`에서 선택합니다. 스킬은 [.agents/skills/update-docs/SKILL.md](.agents/skills/update-docs/SKILL.md)에 있으며, 실행할 때마다 새로 추가되거나 이름이 바뀐 문서까지 동적으로 탐색합니다. 푸시는 자동으로 수행하지 않으므로 필요하면 별도로 요청합니다.
