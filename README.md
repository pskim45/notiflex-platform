# Notiflex Platform

Notiflex는 B2B 환경에서 여러 채널의 알림을 안정적으로 전달하기 위한 알림 SaaS 실습 프로젝트입니다. 이 저장소는 Go 애플리케이션을 컨테이너로 빌드하고 GKE에 배포한 뒤, GitOps와 관측 가능성, 점진적 배포까지 확장하는 과정을 담습니다.

## 현재 상태

Notiflex API `v0.1.2`가 GKE에 배포되어 있습니다. `/health` 상태 확인, 애플리케이션·Go·Pod 정보를 반환하는 `/version`, Pod별 순차 ID를 발급하는 `/id` API를 제공하며, Kubernetes Deployment는 replica 2개로 실행됩니다. GitHub Actions가 `app/` 변경을 테스트하고 SHA 태그 이미지를 Artifact Registry에 게시하며, Argo CD `v3.3.6`이 `k8s/smb`를 자동 동기화합니다. 관측 가능성 구성은 이후 단계에서 추가합니다.

## 기술 스택

- 애플리케이션: Go 표준 라이브러리
- 컨테이너: 멀티 스테이지 빌드, `scratch` 런타임 이미지
- 클라우드: Google Cloud Platform
- 런타임: GKE Standard(영역 클러스터, Spot VM)
- 이미지 저장소: Artifact Registry
- GitOps: Argo CD
- 관측 가능성: Prometheus, Grafana, Loki, Fluent Bit, Tempo
- 배포 전략: Rolling Update → Blue/Green → Canary

## 디렉터리 구조

```text
notiflex-platform/
├── app/                  # Go 애플리케이션과 Dockerfile
├── argocd/
│   └── notiflex-smb.yaml # Argo CD Application 선언
├── k8s/
│   └── smb/              # Kubernetes 매니페스트
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

Cloud Build로 이미지를 빌드하고 Artifact Registry에 게시합니다.

```bash
gcloud builds submit app/ \
  --project=project-10edc337-9677-4dfc-91a \
  --tag=asia-northeast3-docker.pkg.dev/project-10edc337-9677-4dfc-91a/notiflex/api:v0.1.2
```

`main` 브랜치에서 `app/**`가 변경되면 `.github/workflows/ci.yaml`이 자동으로 테스트·빌드·푸시를 수행합니다. GCP 인증은 장기 서비스 계정 키 대신 Workload Identity Federation을 사용하고, 이미지는 `sha-<커밋 앞 7자리>` 태그로 게시합니다.

애플리케이션 매니페스트는 Argo CD가 `main` 브랜치의 `k8s/smb` 디렉터리에서 자동 동기화합니다. 초기 구성이나 수동 검증이 필요할 때는 다음 명령을 사용할 수 있습니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops apply -f k8s/smb/namespace.yaml
kubectl --context gke-sysnet4admin_book_gitaiops apply -f k8s/smb/
kubectl --context gke-sysnet4admin_book_gitaiops rollout status deployment/notiflex-api -n notiflex
```

GitOps 동기화 상태는 다음과 같이 확인합니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops get application notiflex-smb -n argocd
```

로컬 포트 포워딩 후 API를 확인할 수 있습니다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops port-forward service/notiflex-api -n notiflex 8080:80
curl http://localhost:8080/health
curl http://localhost:8080/version
curl http://localhost:8080/id
```

## AI 에이전트 사용

Codex를 포함한 AI 에이전트는 작업 전에 [AGENTS.md](AGENTS.md)를 읽어야 합니다. 이 파일에는 프로젝트 컨텍스트, GCP 대상, 검증 및 안전 규칙이 정의되어 있습니다.

각 장을 마치면 저장소 전용 `$update-docs` 스킬로 그 시점의 코드·인프라와 모든 문서를 동기화하고 검증된 변경을 커밋합니다.

```text
$update-docs
```

Codex에서는 `$update-docs`로 명시 호출하거나 `/skills`에서 선택합니다. 스킬은 [.agents/skills/update-docs/SKILL.md](.agents/skills/update-docs/SKILL.md)에 있으며, 실행할 때마다 새로 추가되거나 이름이 바뀐 문서까지 동적으로 탐색합니다. 푸시는 자동으로 수행하지 않으므로 필요하면 별도로 요청합니다.
