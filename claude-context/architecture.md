# Notiflex 현재 아키텍처

> 스냅샷 기준: 2026-08-08, GKE 클러스터와 Git `main` 실측. 이 문서는 현재 구조를 빠르게 파악하기 위한 요약이며, 값의 선택 이유와 변경 이력은 [`docs/architecture-decisions.md`](../docs/architecture-decisions.md)를 따른다.

## 3층 지식 구조

- `CLAUDE.md`와 `AGENTS.md`: 프로젝트 메타데이터, 작업 방식, 안전 규칙을 제공한다.
- `claude-context/`: 현재 배포된 아키텍처, 연결 관계, 핵심 설정을 요약한다. 아키텍처 변경 시 이 문서를 함께 갱신한다.
- `docs/architecture-decisions.md`: 기술을 선택하거나 교체한 이유와 과거 결정을 보존한다.

## 클러스터 토폴로지

| 항목 | 현재 설정 |
|---|---|
| GCP 프로젝트 | `project-10edc337-9677-4dfc-91a` |
| 클러스터 | GKE Standard `notiflex-cluster`, `asia-northeast3-a` |
| Kubernetes | `1.35.6-gke.1250000` |
| 노드풀 | `default-pool` 2대, `api-pool`·`worker-pool`·`ops-pool` 각 1대; 모두 Spot VM |
| 컨테이너 런타임 | Container-Optimized OS, containerd `2.1.7` |
| Workload Identity | `project-10edc337-9677-4dfc-91a.svc.id.goog` |
| GKE 기능 | Gateway API, Secret Manager add-on, Workload Identity/GKE metadata server |
| 외부 진입점 | Regional External Managed Gateway, `35.216.70.162:80` |

`default-pool`은 `e2-medium` 2대로 Valkey·관측성·컨트롤러를 실행한다. `api-pool`은 `e2-medium` 1대로 API를 전담하고, `worker-pool`은 `e2-standard-2` 1대로 향후 Kafka를, `ops-pool`은 `e2-small` 1대로 향후 운영 작업을 받을 준비가 되어 있다. 새 풀은 모두 `pd-standard` 50GiB, Spot, `GKE_METADATA`를 사용한다. 현재 worker·ops 풀에는 시스템 DaemonSet만 실행되며 taint는 적용하지 않았다.

## 컴포넌트와 연결 관계

```text
사용자
  │ HTTP :80
  ▼
GKE Regional External Gateway (notiflex-gateway, 35.216.70.162)
  │ HTTPRoute notiflex-route: PathPrefix /
  ▼
stable Service notiflex-api :80
  │ targetPort http/:8080, Rollout이 stable ReplicaSet selector 관리
  ▼
Argo Rollout notiflex-api (replicas 1, api-pool nodeSelector, image sha-059f3ab, API v0.3.2)
  │
  ├─ DNS/TCP ──▶ valkey-primary.notiflex.svc.cluster.local:6379
  │               └─ Valkey StatefulSet 1 Pod, standalone, 공유 키 notiflex:id
  │
  └─ CSI mount ─▶ /mnt/secrets/valkey-password
                  └─ SecretProviderClass notiflex-secrets
                     └─ GCP Secret Manager valkey-password/versions/latest

Canary 시 별도 ReplicaSet
  └─ canary Service notiflex-api-preview :80
     (현재 HTTPRoute가 직접 참조하지 않음)
```

### 애플리케이션

- Go 표준 `net/http` 서버이며 컨테이너 포트는 `8080`이다.
- `GET /health`는 상태, `GET /version`은 앱·Go·Pod 정보, `GET /id`는 Valkey `INCR notiflex:id` 결과를 반환한다.
- `VALKEY_ADDR`는 내부 DNS로 고정하고, credential은 `VALKEY_PASSWORD_FILE`의 읽기 전용 파일에서 읽는다.
- readiness는 `/health`를 최초 2초 뒤 5초마다, liveness는 최초 5초 뒤 10초마다 검사한다.
- 리소스는 request `25m/32Mi`, limit `200m/64Mi`이다. non-root, read-only root filesystem, capability 전체 제거, RuntimeDefault seccomp를 사용한다.
- `cloud.google.com/gke-nodepool: api-pool` nodeSelector로 API Pod를 API 전용 풀에만 배치한다. taint가 없으므로 다른 일반 Pod가 api-pool에 배치되는 것까지 차단하지는 않는다.

### 트래픽과 Canary

- `Gateway/notiflex-gateway`는 `gke-l7-regional-external-managed` 클래스를 사용하며 같은 namespace의 HTTPRoute만 허용한다.
- `HTTPRoute/notiflex-route`는 모든 경로(`/`)를 stable Service `notiflex-api:80`으로 보낸다.
- `HealthCheckPolicy/notiflex-healthcheck`는 stable Service의 `/health:8080`을 15초 주기로 검사한다. timeout 5초, healthy threshold 1, unhealthy threshold 2이다.
- Rollout은 stable Service `notiflex-api`와 canary Service `notiflex-api-preview`의 Pod hash selector를 관리한다.
- Canary 단계는 `20% → 30초 대기 → 50% → 30초 대기 → 80% → 30초 대기 → 100%`이다.
- 현재 replica가 1이고 HTTPRoute가 stable Service 하나만 참조하므로 setWeight는 정밀한 외부 요청 비율 분산이 아니다. 실제 가중 트래픽에는 replica 확장과 Gateway API traffic routing 연동이 필요하다.

### 공유 상태와 credential

- Valkey는 Bitnami chart `6.2.6`, 앱 `9.1.1`, standalone StatefulSet 1개로 동작한다. request `10m/64Mi`, limit `200m/128Mi`이다.
- API의 Kubernetes ServiceAccount `notiflex-api`는 GCP Service Account `notiflex-sa@project-10edc337-9677-4dfc-91a.iam.gserviceaccount.com`에 연결된다.
- Pod는 장기 GCP 서비스 계정 키를 저장하지 않는다. GKE metadata server가 Pod identity를 발급하고, GSA에 부여된 최소 IAM 권한으로 Secret Manager 값을 읽는다.
- GKE managed CSI driver `secrets-store-gke.csi.k8s.io`와 provider `gke`가 Secret Manager의 `valkey-password` 최신 버전을 읽기 전용 파일로 마운트한다.
- Git 저장소, Rollout 환경변수, 컨테이너 이미지에는 실제 비밀번호가 없다. Kubernetes Secret은 Valkey chart가 소비하지만 API credential 전달 경로는 CSI 파일이다.

## 배포 파이프라인

```text
개발자 git push (main, app/** 변경)
  → GitHub Actions CI
     → GitHub OIDC + Workload Identity Federation으로 GCP 인증
     → Docker build
     → Artifact Registry notiflex/api:sha-<7자리 커밋>
     → k8s/smb/rollout.yaml 이미지 태그 자동 커밋
  → Argo CD Application notiflex-smb
     → main의 k8s/smb 감시, auto-sync + prune + selfHeal
  → Argo Rollouts controller
     → Canary 단계 실행 및 stable/canary Service selector 전환
```

- GitHub Actions 권한은 `id-token: write`, `contents: write`이다. 장기 GCP 키 대신 OIDC를 사용하며 repository secrets에는 provider·service account·project 식별자만 둔다.
- Artifact Registry 경로는 `asia-northeast3-docker.pkg.dev/project-10edc337-9677-4dfc-91a/notiflex/api:<TAG>`이다.
- Argo CD Application은 `main/k8s/smb`를 `notiflex` namespace로 배포하고 `main`의 최신 커밋을 추적한다. 2026-08-08 실측 상태는 Synced/Healthy이다.
- 현재 Rollout은 step 6/6, Healthy, 1/1 Ready이며 stable ReplicaSet hash는 `7f56884766`이다.

## 관측 가능성

| 컴포넌트 | 역할과 현재 설정 |
|---|---|
| Prometheus | kube-prometheus-stack chart `88.1.3`; 메트릭 수집, retention 2일, `256Mi` request/`512Mi` limit |
| Grafana | 메트릭·로그 조회; dashboard sidecar가 모든 namespace의 dashboard ConfigMap 검색 |
| Alertmanager | PrometheusRule 알림 라우팅; `PodRestartTooMany`는 `notiflex` Pod가 5분간 2회 초과 재시작 후 1분 지속 시 warning |
| Loki | chart `7.2.0`, 앱 `3.6.11`; SingleBinary 1 replica, filesystem PVC 5Gi, 인증 비활성 |
| Fluent Bit | chart `0.57.9`, 앱 `5.0.9`; 노드별 DaemonSet이 컨테이너 로그를 읽어 `loki-gateway.monitoring.svc:80`으로 전송 |
| GKE managed collectors | `gmp-system` collector 5/5와 GKE metrics agent가 노드 메트릭을 수집 |
| Tempo | 아직 설치하지 않았으며 ch8에서 도입 예정 |

Prometheus, Grafana, Alertmanager, Loki, Fluent Bit Pod는 현재 모두 Ready이다. Grafana와 Prometheus는 ClusterIP이므로 기본적으로 클러스터 내부 접근 또는 port-forward를 사용한다.

## 주요 namespace

| Namespace | 주요 컴포넌트 |
|---|---|
| `notiflex` | API Rollout/Pod, stable·preview Service, Gateway, HTTPRoute, HealthCheckPolicy, Valkey StatefulSet, ServiceAccount, SecretProviderClass |
| `argocd` | Application controller, API server, repo server, ApplicationSet, Dex, Redis; `notiflex-smb` Synced/Healthy |
| `argo-rollouts` | Argo Rollouts controller `v1.9.1` 1/1 Ready |
| `monitoring` | Prometheus, Grafana, Alertmanager, Loki, Fluent Bit, kube-state-metrics, node-exporter, PrometheusRule |
| `kube-system` | GKE DNS/네트워크/메트릭 구성과 Secret Manager CSI driver/provider·metadata server 각 5/5 Ready |
| `gmp-system` | Google Managed Prometheus operator·collector |

현재 API Pod는 `api-pool`, Valkey Pod는 `default-pool`에서 Running이며 재시작은 0회다. 이 문서의 IP, revision, Pod hash, 버전과 Ready 수는 가변 정보이므로 아키텍처 변경이나 환경 재구축 후 다시 실측해야 한다.
