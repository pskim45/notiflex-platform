# CronJob 수동 실행

수동 실행은 원본 CronJob을 수정하지 않고 별도의 일회성 Job을 만든다. 외부 호출, 데이터 변경, 알림 중복 같은 부작용이 있을 수 있으므로 실행 전에 영향과 중복 여부를 확인한다.

## 사전 확인

1. 클러스터·namespace·CronJob 이름, schedule, suspend, lastScheduleTime을 확인한다.
2. 현재 실행 중인 Job과 다음 예약 시각을 확인해 자동 실행과 겹치지 않는지 판단한다.
3. Job의 명령, 대상 시스템, credential, nodeSelector와 데이터 변경 여부를 확인한다.
4. 일회성 Job 이름을 `<cronjob-name>-manual-<yyyyMMdd-HHmmss>` 형식으로 정하고 기존 리소스와 충돌하지 않는지 확인한다.
5. 데이터 변경이나 외부 알림 등 부작용이 있으면 정확한 영향 범위를 제시하고 사용자 승인을 받는다. 읽기 전용 헬스체크는 요청 범위 안에서 실행할 수 있다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops get cronjob <cronjob-name> -n <namespace> -o wide
kubectl --context gke-sysnet4admin_book_gitaiops get jobs -n <namespace>
kubectl --context gke-sysnet4admin_book_gitaiops get cronjob <cronjob-name> -n <namespace> -o yaml
```

## 실행

1. CronJob 템플릿으로 일회성 Job을 생성한다.
2. Job 상태와 Pod 로그를 관찰한다. timeout을 정하고 무한 대기하지 않는다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops create job <manual-job-name> --from=cronjob/<cronjob-name> -n <namespace>
kubectl --context gke-sysnet4admin_book_gitaiops wait --for=condition=complete job/<manual-job-name> -n <namespace> --timeout=120s
kubectl --context gke-sysnet4admin_book_gitaiops logs job/<manual-job-name> -n <namespace>
```

## 사후 검증

1. Job이 `Complete`인지, 재시작·실패 횟수가 예상 범위인지 확인한다.
2. 로그뿐 아니라 대상 시스템의 실제 결과도 확인한다.
3. 원본 CronJob의 schedule과 jobTemplate이 변경되지 않았는지 확인한다.
4. 수동 Job은 CronJob의 history 정리 대상이 아니다. 로그와 결과를 보존한 뒤, 정확한 Job 이름을 보여주고 사용자 승인을 받아 삭제한다.

```powershell
kubectl --context gke-sysnet4admin_book_gitaiops delete job <manual-job-name> -n <namespace>
```

## 중단 조건

- 동일 CronJob의 Job이 실행 중이거나 다음 자동 실행과 겹칠 가능성이 있으면 실행하지 않는다.
- 명령의 부작용이나 대상 credential을 확인할 수 없으면 실행하지 않는다.
- 수동 Job 삭제는 별도의 파괴적 작업이므로 승인 없이 수행하지 않는다.
