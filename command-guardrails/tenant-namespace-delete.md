# 테넌트 Namespace 삭제

Namespace 삭제는 내부의 워크로드, Secret, ConfigMap, PVC를 함께 제거할 수 있다. GitOps 소유권과 공유 의존성을 먼저 정리하고, 정확한 namespace에 대한 명시적인 승인을 받은 뒤 진행한다.

## 사전 확인

1. 현재 context와 삭제 대상 namespace의 이름·라벨을 확인한다. `default`, `kube-*`, `argocd`, `monitoring`, `kafka` 같은 플랫폼 namespace는 이 절차로 삭제하지 않는다.
2. 모든 namespaced 리소스와 PVC, Secret, ConfigMap, ServiceAccount, RBAC를 조회한다. Secret 값은 출력하지 않는다.
3. PVC 데이터와 Secret 복구 방법을 정하고 필요한 백업을 완료한다.
4. 다른 namespace가 대상 Service, credential, Kafka Consumer Group 또는 공유 Valkey 데이터를 참조하는지 확인한다.
5. 대상 namespace를 관리하는 Argo CD Application과 `k8s/bootstrap/namespaces.yaml` 선언을 확인한다.
6. 사용자에게 정확한 namespace, 삭제될 주요 리소스, 영구 데이터 손실 가능성, 공유 서비스 영향과 복구 방법을 제시하고 승인을 받는다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops get namespace <tenant-namespace> --show-labels
kubectl --context gke-sysnet4admin_book_gitaiops get all,pvc,configmap,secret,serviceaccount,role,rolebinding -n <tenant-namespace>
kubectl --context gke-sysnet4admin_book_gitaiops get applications -n argocd
```

## 실행

1. 관련 Producer, Consumer와 외부 트래픽을 중지해 신규 요청을 차단한다.
2. `argocd/apps/notiflex-<tenant>.yaml`과 `k8s/bootstrap/namespaces.yaml`의 대상 선언을 같은 Git 변경에서 제거한다.
3. diff에서 다른 tenant나 플랫폼 namespace가 포함되지 않았는지 검토한 뒤 커밋·push한다.
4. Argo CD가 Application과 관리 리소스를 정리할 때까지 관찰한다.
5. namespace가 `Terminating`에 남아도 finalizer를 강제로 제거하지 않는다. 원인을 조회해 보고하고 별도 승인을 받는다.
6. GitOps 정리 후에도 namespace가 남는 경우에만, 사용자가 다시 정확한 대상을 승인하면 직접 삭제한다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops delete namespace <tenant-namespace>
```

## 사후 검증

1. Argo CD Application과 namespace가 모두 사라졌는지 확인한다.
2. 남은 Application이 모두 `Synced/Healthy`인지 확인한다.
3. 다른 tenant의 API, 공유 Valkey, Kafka와 관측성 구성에 오류가 없는지 확인한다.
4. 삭제 대상, 승인, 백업, Git 커밋, 실행 시각과 검증 결과를 운영 기록에 남긴다.

## 중단 조건

- 대상이 bootstrap 또는 다른 Application에 계속 선언되어 있으면 삭제하지 않는다.
- 백업·복구 방법이나 공유 의존성이 불명확하면 삭제하지 않는다.
- 빈 문자열, 변수만 있는 값, glob 또는 여러 namespace를 대상으로 삭제하지 않는다.
- 사용자가 정확한 namespace 삭제를 명시적으로 승인하지 않으면 조회 결과만 보고하고 중단한다.
