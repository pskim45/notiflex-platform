# Notiflex 여정 기록

이 파일은 프로젝트에서 실제로 진행하고 검증한 내용을 기록한다. 각 단계가 완료되면 실제 결과를 기준으로 갱신한다.

## 진행 현황

| 챕터 | 서브챕터 | 상태 | 완료일 | 비고 |
|------|---------|------|--------|------|
| ch2 | 2.2 설치 확인 | ✅ | 2026-08-03 | Git, gcloud, kubectl, Codex 실행 환경 확인 |
| ch2 | 2.3 gcloud 설정 | ✅ | 2026-08-03 | 프로젝트·서울 리전·존 및 활성 계정 확인 |
| ch2 | 2.4 GitHub 저장소 | ✅ | 2026-08-03 | `pskim45/notiflex-platform` private 원격 저장소 연결 |
| ch2 | 2.5 GKE 클러스터 | ✅ | 2026-08-03 | GKE Standard, Spot `e2-medium` 2노드, Gateway API 활성화 |
| ch2 | 2.6 빌드/배포 | ✅ | 2026-08-03 | Cloud Build 테스트 성공, `v0.1.0` 배포 및 API 검증 |
| ch2 | 2.7 첫 커밋 | ✅ | 2026-08-03 | 코드·매니페스트·문서 최초 커밋 및 GitHub 푸시 |
| ch2 | update-docs 스킬 | ✅ | 2026-08-03 | 저장소 문서 동적 탐색·동기화·검증·커밋 워크플로 추가 |
| ch3 | 3.2 GitOps 도구 | ⬜ | | |
| ch3 | 3.3 기능 추가 | ⬜ | | |
| ch3 | 3.4 CI | ⬜ | | |
| ch3 | 3.5 CI-CD 연결 | ⬜ | | |
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
| 이미지 빌드 | Cloud Build | 로컬 Docker 빌드 | 로컬 도구 의존 없이 GCP에서 테스트·빌드·푸시를 일관되게 수행 |
| 문서 동기화 | 저장소 범위 Codex 스킬 | 전역 개인 스킬, 고정 문서 목록 | 팀과 공유하고 이후 장의 신규 문서도 수정 없이 동적으로 처리 |

## 현재 버전

| 컴포넌트 | 버전 | 변경 이력 |
|---------|------|----------|
| Go | 1.25 | 최초 구성 |
| Notiflex 이미지 | `v0.1.0` (`sha256:dad154eff262693903e2d73d8ef0442242060151a15c9bb7171d8141cbb0ccc0`) | 최초 배포 |
| GKE | `1.35.6-gke.1250000` | 최초 클러스터 생성 |
| ArgoCD | 미설치 | ch3 예정 |
| Kafka | 미설치 | ch8 예정 |
| OTel SDK | 미설치 | ch8 예정 |

## 현재 리소스

| 노드풀 | 머신 타입 | 노드 수 | 주요 워크로드 |
|--------|-----------|---------|---------------|
| `default-pool` | `e2-medium` Spot VM | 2 | `notiflex-api` 2 replicas, GKE 시스템 구성 요소 |

| Kubernetes 리소스 | 네임스페이스 | 상태 |
|---------------------|---------------|------|
| Deployment `notiflex-api` | `notiflex` | 2/2 Ready |
| Service `notiflex-api` | `notiflex` | ClusterIP, 80 → 8080 |

## 트러블슈팅 이력

| 챕터 | 문제 | 해결 |
|------|------|------|
| ch2.5 | Kubernetes Engine API가 비활성화되어 클러스터 조회 실패 | `container.googleapis.com` 활성화 후 GKE 클러스터 생성 |
| ch2.5 | 생성 직후 GatewayClass가 보이지 않음 | 컨트롤러 반영을 기다린 뒤 4개 GatewayClass의 `ACCEPTED=True` 확인 |
| ch2.6 | Cloud Build API가 비활성화됨 | `cloudbuild.googleapis.com` 활성화 |
| ch2.6 | 기본 Compute 서비스 계정이 Cloud Build 소스 버킷을 읽지 못함 | 해당 서비스 계정에 `roles/cloudbuild.builds.builder` 부여 |
| ch2.6 | 같은 `v0.1.0` 태그 재빌드 후 실행 중 Pod와 Registry digest 불일치 | `imagePullPolicy: Always` 적용 후 최신 digest로 rollout 및 API 재검증 |
| ch2.6 | Spot VM 노드가 모두 사라져 API Pod 2개가 `Pending` 상태로 유지됨 | `default-pool`을 2대로 resize한 뒤 노드 `Ready`, Deployment 2/2 및 Service 엔드포인트 복구 확인 |
