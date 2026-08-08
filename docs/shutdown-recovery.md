# GCP 실습 환경 종료 및 복구

이 문서는 비용을 중단하기 위해 Notiflex 실습 인프라를 삭제한 뒤 Git 저장소로 재구축하는 절차다. 인프라 구성은 복구되지만, 삭제된 Loki 과거 로그와 기존 외부 IP는 복구되지 않는다.

## 종료 기준 상태

- 프로젝트: `project-10edc337-9677-4dfc-91a`
- 리전/존: `asia-northeast3` / `asia-northeast3-a`
- 클러스터: GKE Standard `notiflex-cluster`
- 노드풀: `default-pool` Spot `e2-medium` 2대와 역할별 `api-pool`·`worker-pool`·`ops-pool` 각 1대
- 애플리케이션: `v0.3.2`, 이미지 태그 `sha-059f3ab`
- 배포: Argo CD + Argo Rollouts Canary
- 외부 진입점: GKE 리전 외부 Gateway API
- 관측성: kube-prometheus-stack, Loki, Fluent Bit
- 상태·시크릿: Valkey, GCP Secret Manager, Workload Identity, GKE managed CSI

## 복구 전 로컬 설정

```powershell
gcloud config set project project-10edc337-9677-4dfc-91a
gcloud config set compute/region asia-northeast3
gcloud config set compute/zone asia-northeast3-a
gcloud auth configure-docker asia-northeast3-docker.pkg.dev
```

## 1. Artifact Registry와 이미지 복구

```powershell
gcloud artifacts repositories create notiflex `
  --project project-10edc337-9677-4dfc-91a `
  --location asia-northeast3 `
  --repository-format docker `
  --description "Notiflex container images"

gh workflow run CI --repo pskim45/notiflex-platform --ref main
$runId = gh run list --repo pskim45/notiflex-platform --workflow CI --limit 1 --json databaseId --jq '.[0].databaseId'
gh run watch $runId --repo pskim45/notiflex-platform --exit-status
git pull --ff-only origin main
```

GitHub Actions가 새 SHA 이미지 태그를 게시하고 `k8s/smb/rollout.yaml`을 자동 갱신한다. 종료 전에 사용하던 이미지는 삭제됐으므로 새 CI 실행 없이 이전 매니페스트를 배포하면 `ImagePullBackOff`가 발생한다.

## 2. GKE 클러스터 복구

```powershell
gcloud container clusters create notiflex-cluster `
  --project project-10edc337-9677-4dfc-91a `
  --zone asia-northeast3-a `
  --machine-type e2-medium `
  --num-nodes 2 `
  --spot `
  --gateway-api standard `
  --workload-pool project-10edc337-9677-4dfc-91a.svc.id.goog `
  --enable-secret-manager `
  --disk-size 30

gcloud container clusters get-credentials notiflex-cluster `
  --project project-10edc337-9677-4dfc-91a `
  --zone asia-northeast3-a
```

생성된 컨텍스트 이름을 확인한 뒤 `gke-sysnet4admin_book_gitaiops`로 맞춘다. 같은 이름의 오래된 컨텍스트가 있으면 먼저 삭제한다.

```powershell
kubectl config get-contexts
kubectl config delete-context gke-sysnet4admin_book_gitaiops
kubectl config rename-context (kubectl config current-context) gke-sysnet4admin_book_gitaiops
kubectl --context gke-sysnet4admin_book_gitaiops get nodes
```

노드풀의 Workload Identity metadata mode도 확인한다. 비어 있으면 업데이트가 완료될 때까지 기다린 뒤 다음 단계로 진행한다.

```powershell
gcloud container node-pools update default-pool `
  --cluster notiflex-cluster `
  --project project-10edc337-9677-4dfc-91a `
  --zone asia-northeast3-a `
  --workload-metadata GKE_METADATA
```

역할별 노드풀을 생성한다. 모든 풀에 Workload Identity metadata를 설정해야 API의 Secret Manager CSI 마운트가 동작한다.

```powershell
gcloud container node-pools create api-pool `
  --cluster notiflex-cluster --project project-10edc337-9677-4dfc-91a `
  --zone asia-northeast3-a --machine-type e2-medium `
  --disk-type pd-standard --disk-size 50 --num-nodes 1 --spot `
  --workload-metadata GKE_METADATA

gcloud container node-pools create worker-pool `
  --cluster notiflex-cluster --project project-10edc337-9677-4dfc-91a `
  --zone asia-northeast3-a --machine-type e2-standard-2 `
  --disk-type pd-standard --disk-size 50 --num-nodes 1 --spot `
  --workload-metadata GKE_METADATA

gcloud container node-pools create ops-pool `
  --cluster notiflex-cluster --project project-10edc337-9677-4dfc-91a `
  --zone asia-northeast3-a --machine-type e2-small `
  --disk-type pd-standard --disk-size 50 --num-nodes 1 --spot `
  --workload-metadata GKE_METADATA
```

각 생성 명령 전에는 `gcloud container operations list --zone asia-northeast3-a --filter="status=RUNNING"` 출력이 없는지 확인한다. 노드풀 생성은 동시에 실행하지 않는다.

## 3. Argo Rollouts 설치

애플리케이션 매니페스트를 동기화하기 전에 Rollout CRD가 먼저 있어야 한다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops create namespace argo-rollouts
kubectl --context gke-sysnet4admin_book_gitaiops apply --server-side `
  -n argo-rollouts `
  -f https://github.com/argoproj/argo-rollouts/releases/download/v1.9.1/install.yaml
kubectl --context gke-sysnet4admin_book_gitaiops wait `
  --for=condition=Available deployment/argo-rollouts `
  -n argo-rollouts --timeout=180s
```

## 4. Argo CD와 애플리케이션 복구

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops create namespace argocd
kubectl --context gke-sysnet4admin_book_gitaiops apply `
  -n argocd --server-side --force-conflicts `
  -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
kubectl --context gke-sysnet4admin_book_gitaiops wait `
  --for=condition=Available deployment/argocd-server `
  -n argocd --timeout=180s
```

저장소가 private으로 바뀌었다면 Argo CD 저장소 자격 증명을 클러스터에 직접 등록한다. 토큰을 YAML이나 Git에 저장하지 않는다.
`root-app`은 Valkey와 Grafana의 기존 Secret을 참조하므로 아직 적용하지 않고 7단계에서 bootstrap credential을 만든 뒤 적용한다.

## 5. Gateway 선행 조건 복구

클러스터 생성 시 Gateway API를 활성화했으므로 GatewayClass가 준비될 때까지 기다린다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops get gatewayclass
gcloud compute networks subnets describe proxy-only-subnet `
  --project project-10edc337-9677-4dfc-91a `
  --region asia-northeast3
```

`proxy-only-subnet`이 없을 때만 생성한다.

```powershell
gcloud compute networks subnets create proxy-only-subnet `
  --project project-10edc337-9677-4dfc-91a `
  --purpose REGIONAL_MANAGED_PROXY `
  --role ACTIVE `
  --region asia-northeast3 `
  --network default `
  --range 172.16.0.0/23
```

Argo CD App of Apps가 `notiflex-smb`를 동기화하면 Gateway, HTTPRoute, HealthCheckPolicy와 Notiflex Rollout이 자동 생성된다.

## 6. 관측성 복구

관측성 구성은 별도 Helm CLI 명령으로 설치하지 않는다. 7단계에서 `root-app`을 적용하면 `kube-prometheus`·`loki`·`fluent-bit` Application과 `monitoring-config` Application이 chart와 추가 YAML을 자동 동기화한다.

새 Loki PVC와 빈 데이터 디스크가 만들어진다. 과거 로그는 돌아오지 않는다.

## 7. Valkey와 Secret Manager 복구

Argo CD가 참조할 Valkey와 Grafana bootstrap Secret을 먼저 만들고 Secret Manager API와 `notiflex-sa`를 복구한다. 비밀번호 값은 콘솔에 출력하거나 Git에 저장하지 않는다.

```powershell
$projectId = "project-10edc337-9677-4dfc-91a"
$gsa = "notiflex-sa@$projectId.iam.gserviceaccount.com"

gcloud services enable secretmanager.googleapis.com --project=$projectId
gcloud iam service-accounts create notiflex-sa `
  --project=$projectId `
  --display-name="Notiflex workload identity"
gcloud secrets create valkey-password `
  --project=$projectId `
  --replication-policy=automatic

$valkeyBytes = New-Object byte[] 32
$grafanaBytes = New-Object byte[] 32
$rng = [Security.Cryptography.RandomNumberGenerator]::Create()
$rng.GetBytes($valkeyBytes)
$rng.GetBytes($grafanaBytes)
$valkeyPassword = [Convert]::ToBase64String($valkeyBytes)
$grafanaPassword = [Convert]::ToBase64String($grafanaBytes)
$valkeyFile = New-TemporaryFile
$grafanaFile = New-TemporaryFile
try {
  [IO.File]::WriteAllText($valkeyFile.FullName, $valkeyPassword)
  [IO.File]::WriteAllText($grafanaFile.FullName, $grafanaPassword)
  kubectl --context gke-sysnet4admin_book_gitaiops create namespace notiflex
  kubectl --context gke-sysnet4admin_book_gitaiops create namespace monitoring
  kubectl --context gke-sysnet4admin_book_gitaiops create secret generic valkey `
    -n notiflex --from-file="valkey-password=$($valkeyFile.FullName)"
  kubectl --context gke-sysnet4admin_book_gitaiops create secret generic kube-prometheus-grafana `
    -n monitoring --from-literal="admin-user=admin" `
    --from-file="admin-password=$($grafanaFile.FullName)"
  gcloud secrets versions add valkey-password `
    --project=$projectId `
    --data-file=$valkeyFile.FullName
} finally {
  $rng.Dispose()
  Remove-Item -LiteralPath $valkeyFile.FullName,$grafanaFile.FullName -Force
  Remove-Variable valkeyBytes,grafanaBytes,valkeyPassword,grafanaPassword
}

gcloud secrets add-iam-policy-binding valkey-password `
  --project=$projectId `
  --member="serviceAccount:$gsa" `
  --role=roles/secretmanager.secretAccessor
gcloud iam service-accounts add-iam-policy-binding $gsa `
  --project=$projectId `
  --member="serviceAccount:$projectId.svc.id.goog[notiflex/notiflex-api]" `
  --role=roles/iam.workloadIdentityUser
```

IAM 전파 후 `root-app` 하나를 적용한다. root가 API와 네 Helm 앱, 모니터링 설정을 모두 등록한다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops apply -f argocd/root-app.yaml
kubectl --context gke-sysnet4admin_book_gitaiops get application -n argocd
kubectl --context gke-sysnet4admin_book_gitaiops get `
  secretproviderclass,secretproviderclasspodstatus -n notiflex
```

## 8. 최종 검증

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops get nodes
kubectl --context gke-sysnet4admin_book_gitaiops get pods -A
kubectl --context gke-sysnet4admin_book_gitaiops get application notiflex-smb -n argocd
kubectl --context gke-sysnet4admin_book_gitaiops get rollout notiflex-api -n notiflex
kubectl --context gke-sysnet4admin_book_gitaiops get gateway,httproute -n notiflex
kubectl --context gke-sysnet4admin_book_gitaiops get pvc -A
```

Gateway의 새 외부 IP를 조회해 API를 검증한다.

```powershell
$gatewayIp = kubectl --context gke-sysnet4admin_book_gitaiops get gateway notiflex-gateway -n notiflex -o jsonpath='{.status.addresses[0].value}'
Invoke-RestMethod "http://$gatewayIp/health"
Invoke-RestMethod "http://$gatewayIp/version"
```

## 종료 후 남겨도 과금되지 않는 구성

- GitHub 저장소의 코드·매니페스트·문서와 CI 워크플로
- Workload Identity Federation 설정과 IAM 바인딩
- 활성화된 Google Cloud API
- 기본 VPC 및 `proxy-only-subnet` 자체

복구 시점에는 Google Cloud 가격 정책과 각 구성 요소의 최신 호환 버전을 다시 확인한다.
