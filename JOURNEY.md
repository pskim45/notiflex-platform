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
| ch4 | 4.2 메트릭 모니터링 | ⬜ | | |
| ch4 | 4.3 로그 수집 | ⬜ | | |
| ch4 | 4.4 알림 | ⬜ | | |
| ch5 | 5.2 트래픽 관리 | ⬜ | | |
| ch5 | 5.3 무중단 배포 | ⬜ | | |
| ch6 | 6.1 캐시 | ⬜ | | |
| ch6 | 6.2 시크릿 관리 | ⬜ | | |
| ch6 | 6.3 Canary 전환 | ⬜ | | |
| ch7 | 7.2 멀티 노드풀 | ⬜ | | |
| ch7 | 7.3 App of Apps | ⬜ | | |
| ch7 | 7.4 멀티테넌시 | ⬜ | | |
| ch8 | 8.1 메시징 | ⬜ | | |
| ch8 | 8.2 트레이싱 | ⬜ | | |
| ch8 | 8.3 CronJob | ⬜ | | |
| ch9 | 9.1 저장소 분석 | ⬜ | | |
| ch9 | 9.2 회고 | ⬜ | | |
| ch9 | 9.3 온보딩 문서 | ⬜ | | |
| ch9 | 9.4 GitAIOps 분석 | ⬜ | | |
| ch9 | 9.5 마무리 | ⬜ | | |

## 도구 선택 기록

| 영역 | 선택 | 검토한 대안 | 선택 이유 |
|------|------|-------------|----------|
| 컨테이너 런타임 | GKE Standard (Zonal) | GKE Autopilot | 노드 구성과 단계별 플랫폼 확장을 직접 실습하기 위해 선택 |
| 이미지 저장소 | Artifact Registry | Docker Hub | GCP IAM 및 Cloud Build와의 네이티브 통합 |
| 이미지 빌드 | GitHub Actions Docker build | Cloud Build, 로컬 Docker 빌드 | 코드 push 시 테스트·SHA 이미지 빌드·Artifact Registry 게시를 자동화 |
| CI | GitHub Actions + Workload Identity Federation | Cloud Build Trigger, Jenkins, 서비스 계정 키 | GitHub push와 직접 연동하고 장기 키 없이 최소 권한으로 Artifact Registry에 게시 |
| 문서 동기화 | 저장소 범위 Codex 스킬 | 전역 개인 스킬, 고정 문서 목록 | 팀과 공유하고 이후 장의 신규 문서도 수정 없이 동적으로 처리 |

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | 1.25.12 | 실행 중 API `/version` 응답으로 확인 |
| Notiflex 이미지 | `sha-066d7dd` (`sha256:720b4442c6555f7f5c93effd2e0ede60ec056ddf52711b79e976b01b866b3aaf`) | API `v0.1.3`, 코드-only push 후 CI-CD 자동 롤링 업데이트 |
| GKE | `1.35.6-gke.1250000` | 최초 클러스터 생성 |
| ArgoCD | `v3.3.6` | 최초 설치 및 `notiflex-smb` Application 연결 |
| Kafka | 미설치 | ch8 예정 |
| OTel SDK | 미설치 | ch8 예정 |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|-----------|---------|---------------|
| `default-pool` | `e2-medium` Spot VM | 2 | `notiflex-api` 2 replicas, GKE 시스템 구성 요소 |

| Kubernetes 리소스 | 네임스페이스 | 상태 |
|---------------------|---------------|------|
| Deployment `notiflex-api` | `notiflex` | `sha-066d7dd`, 2/2 Ready |
| Service `notiflex-api` | `notiflex` | ClusterIP, 80 → 8080 |
| Application `notiflex-smb` | `argocd` | Synced, Healthy, auto-sync/prune/selfHeal 활성화 |

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
