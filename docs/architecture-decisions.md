# Architecture Decision Records

이 문서는 Notiflex 플랫폼에 실제 적용한 아키텍처 결정을 시간 순서대로 기록한다.

## ADR-001: 배포 자동화는 Argo CD (3장)

**시점**: 2026-08 / **결정**: Git을 배포 상태의 기준으로 삼고 Argo CD로 Kubernetes 매니페스트를 자동 동기화한다. Flux, Jenkins X, Spinnaker는 사용하지 않는다.

**이유**:

- Web UI에서 동기화와 배포 상태를 실시간으로 확인할 수 있어 학습 과정의 시각적 피드백이 좋다.
- Application CRD로 Git 경로와 대상 네임스페이스의 관계를 선언적으로 관리할 수 있다.
- `selfHeal`과 `prune`으로 클러스터의 수동 변경과 Git 상태의 차이를 자동으로 해소할 수 있다.
- GKE Standard와 호환되고, 현재 e2-medium 노드 2개에서 감당할 수 있는 리소스 규모다.

## ADR-002: CI는 GitHub Actions와 Workload Identity Federation (3장)

**시점**: 2026-08 / **결정**: GitHub Actions로 테스트·컨테이너 빌드·Artifact Registry 게시·배포 매니페스트 갱신을 자동화하고 GCP 인증에는 Workload Identity Federation을 사용한다. Cloud Build Trigger, Jenkins, 장기 서비스 계정 키는 사용하지 않는다.

**이유**:

- 코드가 있는 GitHub에서 push 이벤트와 CI를 함께 관리하므로 별도 CI 서버나 트리거 시스템이 필요 없다.
- `.github/workflows/ci.yaml` 한 파일로 파이프라인을 선언하고 변경 이력을 Git에 남길 수 있다.
- SHA 기반 이미지 태그와 매니페스트 자동 커밋으로 이미지와 배포 버전을 추적할 수 있다.
- Workload Identity Federation의 단기 자격 증명을 사용해 장기 서비스 계정 키의 보관·회전 위험을 피한다.

## ADR-003: 메트릭은 Prometheus와 Grafana (4장)

**시점**: 2026-08 / **결정**: kube-prometheus-stack의 Prometheus로 Kubernetes 메트릭을 수집하고 Grafana로 시각화한다. Datadog과 Google Cloud Monitoring을 주 모니터링 플랫폼으로 사용하지 않는다.

**이유**:

- Prometheus는 Kubernetes 모니터링의 사실상 표준이며 CRD와 PromQL 생태계가 성숙했다.
- 자체 호스팅 오픈소스 구성으로 학습 환경에서 SaaS 구독 비용이 들지 않는다.
- Helm chart로 Prometheus, Grafana, Alertmanager, node-exporter, kube-state-metrics를 검증된 조합으로 설치할 수 있다.
- Grafana를 이후 Loki 로그와 Tempo 트레이스의 공통 UI로 확장해 관측 도구 파편화를 줄일 수 있다.

## ADR-004: 중앙 로그는 Loki와 Fluent Bit (4장)

**시점**: 2026-08 / **결정**: Fluent Bit DaemonSet으로 컨테이너 로그를 수집하고 Loki SingleBinary에 저장해 Grafana에서 조회한다. ELK Stack과 Google Cloud Logging을 주 로그 플랫폼으로 사용하지 않는다.

**이유**:

- Loki와 Fluent Bit은 이미 여러 플랫폼 구성 요소가 실행 중인 e2-medium 환경에서도 운영 가능한 경량 조합이다.
- 기존 Grafana에 Loki 데이터소스만 추가해 메트릭과 로그를 같은 UI에서 조회할 수 있다.
- 라벨 기반 인덱싱으로 전체 텍스트를 인덱싱하는 방식보다 저장·메모리 비용이 낮다.
- Fluent Bit을 DaemonSet으로 배포해 각 노드의 컨테이너 stdout/stderr를 일관되게 수집할 수 있다.

## ADR-005: 알림 규칙은 PrometheusRule과 Alertmanager (4장)

**시점**: 2026-08 / **결정**: PrometheusRule CRD로 알림 조건을 선언하고 Alertmanager로 라우팅한다. Grafana UI에만 존재하는 알림 규칙은 사용하지 않는다.

**이유**:

- 알림 규칙을 YAML로 Git에 보관해 리뷰, 변경 추적, 재현이 가능한 GitOps 흐름을 유지한다.
- kube-prometheus-stack에 Alertmanager와 PrometheusRule CRD가 이미 포함되어 추가 플랫폼 설치가 필요 없다.
- Kubernetes와 Prometheus 운영에서 널리 쓰이는 표준적인 구성이다.
- Alertmanager의 그룹화, 억제, 라우팅 트리로 이후 Slack·이메일·온콜 채널을 확장할 수 있다.

## ADR-006: 외부 진입점은 Gateway API (5장)

**시점**: 2026-08 / **결정**: GKE의 리전 외부 Gateway API와 HTTPRoute로 Notiflex API를 노출한다. Ingress NGINX와 Istio는 도입하지 않는다.

**이유**:

- Gateway API는 Ingress의 한계를 보완하는 Kubernetes 공식 차세대 트래픽 관리 표준이다.
- GKE의 `gke-l7-regional-external-managed` GatewayClass를 사용하므로 별도 Ingress Controller를 운영하지 않아도 된다.
- Gateway와 HTTPRoute가 인프라 진입점과 애플리케이션 라우팅 책임을 분리한다.
- HTTPRoute의 `backendRefs` 구조를 이용해 이후 점진적 배포와 트래픽 분배로 확장할 수 있다.

## ADR-007: 무중단 배포는 Argo Rollouts Blue/Green (5장, ADR-010으로 대체)

**시점**: 2026-08 / **결정**: Kubernetes Deployment의 Rolling Update 대신 Argo Rollouts의 Blue/Green 전략을 사용한다. 현재 단계에서는 Flagger와 Canary 전략을 도입하지 않는다.

**이유**:

- 기존 Argo CD와 같은 Argo 생태계를 사용해 GitOps 배포 상태와 Rollout 진행을 함께 관찰할 수 있다.
- Rollout CRD의 YAML로 active·preview Service, 자동 승격, 이전 버전 축소 지연을 선언할 수 있다.
- 새 Green ReplicaSet이 준비되는 동안 기존 Blue가 모든 외부 트래픽을 처리해 배포와 트래픽 전환을 분리한다.
- 현재 2 replica 규모에서는 일시적인 리소스 2배 사용을 감당할 수 있고, Canary는 신뢰할 수 있는 메트릭 기반 판정 기준을 마련한 뒤 도입하는 편이 안전하다.

## ADR-008: Pod 간 상태 공유는 Valkey (6장)

**시점**: 2026-08 / **결정**: Pod별 인메모리 ID 카운터를 Valkey standalone의 원자적 `INCR`로 대체한다. Redis, Memcached, DragonflyDB는 사용하지 않는다.

**이유**:

- 여러 API Pod가 동일한 카운터를 사용해 중복 ID를 방지한다.
- Redis 프로토콜과 클라이언트 생태계를 사용하면서 BSD 라이선스를 유지한다.
- 현재 학습 환경에서는 1 Pod와 PVC를 사용하는 standalone 구성이 충분하다.
- `resourcesPreset=none`과 명시적 request/limit으로 2노드 e2-medium 예산에 맞출 수 있다.

## ADR-009: 애플리케이션 credential은 Secret Manager CSI와 Workload Identity (6장)

**시점**: 2026-08 / **결정**: Notiflex가 사용하는 Valkey credential을 GCP Secret Manager에 저장하고 GKE managed Secrets Store CSI로 읽기 전용 파일에 마운트한다. 장기 서비스 계정 키와 애플리케이션 환경변수 credential은 사용하지 않는다.

**이유**:

- Workload Identity의 단기 토큰으로 GCP에 인증해 JSON 서비스 계정 키를 만들지 않는다.
- `notiflex-api` KSA와 `notiflex-sa` GSA에 `valkey-password` accessor 최소 권한만 부여한다.
- credential 값은 Git과 Pod 환경변수에 없고 Secret Manager가 버전과 IAM 감사 경계를 제공한다.
- GKE managed driver와 provider를 사용해 별도 외부 Secret Operator를 운영하지 않는다.

## ADR-010: 배포 전략은 Argo Rollouts Canary (6장)

**시점**: 2026-08 / **결정**: ADR-007의 Blue/Green을 Argo Rollouts Canary로 대체하고 20%→50%→80%→100% 단계마다 30초 pause를 둔다. 별도 Flagger나 서비스 메시는 도입하지 않는다.

**이유**:

- 기존 Argo CD와 Rollout CRD를 유지하면서 새 버전 노출을 단계적으로 관찰할 수 있다.
- `v0.3.2` 배포에서 step 1·3·5의 pause와 step 6 승격을 실제 확인했다.
- Blue/Green의 0%→100% 일괄 전환보다 문제를 발견하고 중단할 관찰 구간이 명시적이다.
- 현재 replica 1과 stable Service 기반 HTTPRoute에서는 정밀한 Gateway 트래픽 비율이 아니므로, 실제 가중 라우팅은 노드풀·replica 확장 후 별도로 도입한다.

## ADR-011: 워크로드 배치는 역할별 GKE 노드풀과 nodeSelector (7장)

**시점**: 2026-08 / **결정**: `default-pool` 외에 `api-pool`·`worker-pool`·`ops-pool`을 만들고 GKE 자동 노드풀 라벨을 사용하는 nodeSelector로 API를 분리한다. 현재는 taint/toleration과 복잡한 nodeAffinity를 도입하지 않는다.

**이유**:

- API를 `api-pool`에만 배치해 Valkey·관측성·향후 Kafka의 리소스 경합과 분리한다.
- `cloud.google.com/gke-nodepool` 자동 라벨은 별도 커스텀 라벨 없이 선언이 단순하다.
- 모든 새 풀에 `GKE_METADATA`를 설정해 Workload Identity와 Secret Manager CSI 접근을 유지한다.
- worker·ops 풀을 미리 분리해 이후 Kafka와 CronJob의 배치 위치를 명확히 한다.

## ADR-012: 여러 앱은 Argo CD App of Apps로 관리 (7장)

**시점**: 2026-08 / **결정**: `root-app`이 `argocd/apps/`를 감시하고 API, Valkey, Prometheus, Loki, Fluent Bit, 모니터링 설정 Application을 등록한다. 단일 클러스터에서는 ApplicationSet 대신 App of Apps를 사용한다.

**이유**:

- 일반 Kubernetes YAML과 외부 Helm chart를 모두 Git push와 Argo CD 자동 동기화라는 동일한 운영 흐름으로 관리한다.
- 기존 chart 버전과 `helm-values/`를 다중 source Application에 명시해 수동 `helm upgrade` 의존성을 제거한다.
- 앱 6개 규모에서는 Application별 순수 YAML이 템플릿 generator보다 읽고 문제를 추적하기 쉽다.
- 기존 Valkey·Grafana credential Secret과 PVC를 유지하면서 Helm CLI 설치 리소스를 안전하게 인수할 수 있다.
