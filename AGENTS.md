# AGENTS.md

이 파일은 이 저장소에서 작업하는 Codex 및 호환 AI 에이전트의 기본 컨텍스트와 행동 규칙이다. 저장소 전체에 적용되며, 하위 디렉터리에 별도의 `AGENTS.md`가 있으면 더 구체적인 하위 지침을 함께 따른다.

## 프로젝트 컨텍스트

- 프로젝트명: Notiflex
- 목적: B2B 고객에게 여러 채널의 알림을 전달하는 SaaS 플랫폼을 구축하고 운영한다.
- 현재 단계: 프로젝트 골격만 구성된 초기 상태다. 존재하지 않는 구현이나 배포 상태를 추정하지 않는다.
- 애플리케이션: 외부 웹 프레임워크 없이 Go 표준 라이브러리를 사용한다.
- 컨테이너: 정적 Go 바이너리를 빌드하고 최종 이미지는 `scratch`를 사용한다.
- 인프라: GKE Standard 영역 클러스터와 Spot VM을 사용한다.
- GitOps: Argo CD를 기준으로 한다.
- 관측 가능성: Prometheus, Grafana, Loki, Fluent Bit, Tempo를 사용한다.
- 배포 전략은 Rolling Update, Blue/Green, Canary 순으로 발전시킨다.

## 저장소 구조

- `app/`: Go 소스, 테스트, 모듈 파일, Dockerfile
- `k8s/smb/`: Notiflex Kubernetes 매니페스트
- `.github/workflows/`: CI/CD 워크플로
- `JOURNEY.md`: 실습 진행 상태, 선택 이유, 버전, 리소스 및 트러블슈팅 기록. 파일이 존재하면 작업 전 읽고, 단계 완료 후 실제 결과만 반영한다.

## 고정 GCP 설정

| 항목 | 값 |
|---|---|
| GCP 프로젝트 ID | `project-10edc337-9677-4dfc-91a` |
| 기본 리전 | `asia-northeast3` |
| 기본 존 | `asia-northeast3-a` |
| GKE 클러스터 | `notiflex-cluster` |
| kubectl 컨텍스트 | `gke-sysnet4admin_book_gitaiops` |
| Artifact Registry 저장소 | `notiflex` |
| 이미지 경로 접두사 | `asia-northeast3-docker.pkg.dev/project-10edc337-9677-4dfc-91a/notiflex` |
| 기본 애플리케이션 이미지 | `asia-northeast3-docker.pkg.dev/project-10edc337-9677-4dfc-91a/notiflex/api:<VERSION>` |
| 애플리케이션 네임스페이스 | `notiflex` |

설정값을 바꾸라는 명시적 요청이 없다면 위 값을 사용한다. 실행 전 `gcloud config list`로 활성 프로젝트·리전·존을 확인하며, 값이 다르면 변경하지 말고 사용자에게 차이를 알린다.

## 행동 규칙

1. 변경 전에 관련 파일, Git 상태, 현재 로컬 또는 클러스터 상태를 먼저 확인한다.
2. 사용자가 요청한 범위만 변경한다. 기존 사용자 변경사항과 무관한 파일은 되돌리거나 정리하지 않는다.
3. 구현 후 영향 범위에 맞는 검증을 수행한다. Go 변경은 최소한 `gofmt`와 `go test ./...`를 실행하고, 매니페스트 변경은 가능한 경우 클라이언트 측 dry-run 또는 동등한 정적 검증을 한다.
4. 코드, 매니페스트, 문서의 이미지 태그와 설정값을 일관되게 유지한다. `latest` 대신 명시적인 버전 태그를 선호한다.
5. 비밀값, 토큰, 서비스 계정 키, kubeconfig 내용을 저장소에 기록하거나 출력하지 않는다. Secret에는 평문 자격 증명을 커밋하지 않는다.
6. 외부 시스템을 변경하는 명령은 대상과 현재 상태를 확인한 뒤 실행한다. 생성·수정·배포처럼 명백히 요청 범위인 작업은 진행하되 결과를 검증한다.
7. 삭제, 강제 푸시, 롤백, 클러스터 재생성, 대규모 리소스 변경처럼 파괴적이거나 복구 비용이 큰 작업은 정확한 대상을 제시하고 사용자 확인을 받은 뒤 실행한다.
8. 실제로 확인하지 않은 배포 성공, 리소스 상태, 테스트 통과를 주장하지 않는다. 검증할 수 없으면 이유와 미검증 항목을 명시한다.
9. 새 구성 요소나 운영 결정을 도입할 때는 기존 기술 선택을 우선하며, 변경이 필요하면 근거와 영향을 설명한다.
10. 작업 완료 시 변경 파일, 수행한 검증, 남은 위험 또는 다음 단계를 간결하게 보고한다.
11. 각 장의 마지막 작업에서는 `.agents/skills/update-docs/SKILL.md`의 `$update-docs` 워크플로를 사용해 신규·기존 문서를 현재 상태와 동기화하고, 검증된 장 변경 사항을 커밋한다.

## Kubernetes 안전 규칙

모든 `kubectl` 명령에는 예외 없이 다음 컨텍스트를 명시한다.

```bash
kubectl --context gke-sysnet4admin_book_gitaiops ...
```

- 읽기 명령도 컨텍스트를 생략하지 않는다.
- 네임스페이스 리소스에는 `-n <namespace>`를 명시한다.
- 적용 전 대상 클러스터, 네임스페이스, diff 또는 dry-run 결과를 확인한다.
- `kubectl apply` 후 rollout, Pod 상태, 이벤트 등 실제 결과를 확인한다.
- `kubectl delete`, `drain`, `cordon`, `scale --replicas=0` 등 서비스에 영향을 주는 작업은 사용자 확인 없이 실행하지 않는다.

## 권장 검증 명령

```bash
# Go 애플리케이션
cd app
gofmt -w .
go test ./...

# 클러스터 조회
kubectl --context gke-sysnet4admin_book_gitaiops get nodes
kubectl --context gke-sysnet4admin_book_gitaiops get pods -n notiflex

# 매니페스트 사전 검증 예시
kubectl --context gke-sysnet4admin_book_gitaiops apply --dry-run=client -f k8s/smb/
```

`gofmt -w`는 Go 파일을 변경하므로 Go 작업 범위에서만 사용한다. 클러스터 연결이 없거나 필요한 도구가 설치되지 않았다면 검증을 생략한 이유를 결과에 기록한다.

## 문서 유지 규칙

- 구조, 실행 방법, GCP 설정이 바뀌면 `README.md`를 함께 갱신한다.
- 진행 이력을 기록하는 `JOURNEY.md`가 있으면 완료된 사실과 실제 검증 결과만 추가한다.
- 명령 예시는 복사해 실행할 수 있어야 하며 placeholder는 `<VERSION>`처럼 명확히 표시한다.
- 문서와 주석은 기본적으로 한국어로 작성하고, 코드 식별자와 제품명은 원문 표기를 유지한다.
