# Sequence Diagrams

## Scale Up Sequence (Entering Window)

### Actors
- **Kubernetes API**: API server and etcd
- **Controller**: Kyklos reconcile loop
- **Target**: Deployment being scaled
- **Cache**: Controller-runtime client cache
- **EventRecorder**: Kubernetes event system
- **Metrics**: Prometheus metrics endpoint

### Sequence: Morning Window Entry at 09:00 IST

```
Time: 09:00:00 IST triggers reconciliation
```

1. **Controller → Cache**: Get TimeWindowScaler "webapp-office-hours"
   - Action: Read from cache
   - Response: TWS object with spec and status

2. **Controller → Internal**: Load timezone "Asia/Kolkata"
   - Action: time.LoadLocation()
   - Response: Location object

3. **Controller → Internal**: Convert UTC to local time
   - Action: time.Now().In(location)
   - Response: 2025-01-26T09:00:00+05:30

4. **Controller → Internal**: Evaluate windows
   - Action: Check each window against current time and day
   - Decision: Monday 09:00 matches window [Mon-Fri 09:00-17:00]
   - Response: effectiveReplicas = 10

5. **Controller → Cache**: Get Deployment "webapp"
   - Action: Read from cache
   - Response: Deployment with status.replicas=2, spec.replicas=2

6. **Controller → Internal**: Compare replicas
   - Action: effectiveReplicas(10) != targetSpec(2)
   - Decision: Scale needed

7. **Controller → Kubernetes API**: Patch Deployment
   - Action: PATCH /apis/apps/v1/namespaces/production/deployments/webapp
   - Body: `{"spec": {"replicas": 10}}`
   - Response: 200 OK, updated Deployment

8. **Controller → EventRecorder**: Emit ScaledUp event
   - Action: Create Event object
   - Message: "Scaled up from 2 to 10 replicas (window: BusinessHours)"

9. **Controller → Kubernetes API**: Update TWS status
   - Action: PATCH /apis/kyklos.io/v1alpha1/.../status
   - Body: Updated status with effectiveReplicas=10, currentWindow="BusinessHours"
   - Response: 200 OK

10. **Controller → Metrics**: Update metrics
    - Action: Set gauge kyklos_effective_replicas{tws="webapp-office-hours"} = 10
    - Action: Increment counter kyklos_scaling_operations_total{direction="up"}

11. **Controller → Internal**: Calculate next boundary
    - Action: Find next window end at 17:00 IST
    - Response: 8 hours until boundary

12. **Controller → Internal**: Schedule requeue
    - Action: Return reconcile.Result{RequeueAfter: 8h + jitter}

### Observable Side Effects
- Deployment spec.replicas changed from 2 to 10
- Deployment controller begins creating 8 new pods
- Event appears in `kubectl get events`
- TWS status shows currentWindow="BusinessHours", effectiveReplicas=10
- Metrics reflect new state
- Log entry: "Scaling up deployment"

---

## Scale Down Sequence (Exiting Window with Grace Period)

### Sequence: Evening Window Exit at 17:00 IST with 300s Grace

```
Time: 17:00:00 IST triggers reconciliation
```

1. **Controller → Cache**: Get TimeWindowScaler "webapp-office-hours"
   - Action: Read from cache
   - Response: TWS with gracePeriodSeconds=300

2. **Controller → Internal**: Load timezone and convert time
   - Action: Get local time 17:00:00 IST
   - Response: Current time just after window end

3. **Controller → Internal**: Evaluate windows
   - Action: Check windows, none match current time
   - Response: effectiveReplicas = 2 (defaultReplicas)

4. **Controller → Internal**: Check grace period
   - Action: effectiveReplicas(2) < status.effectiveReplicas(10)
   - Action: gracePeriodSeconds > 0
   - Decision: Start grace period

5. **Controller → Cache**: Get Deployment "webapp"
   - Action: Read from cache
   - Response: Deployment with status.replicas=10

6. **Controller → Internal**: Apply grace logic
   - Action: Set gracePeriodExpiry = now + 300s
   - Decision: Maintain current replicas during grace
   - Response: effectiveReplicas remains 10 (temporarily)

7. **Controller → EventRecorder**: Emit GracePeriodStarted event
   - Action: Create Event object
   - Message: "Grace period started: 300s before scaling to 2 replicas"

8. **Controller → Kubernetes API**: Update TWS status
   - Action: PATCH status with gracePeriodExpiry="2025-01-26T17:05:00+05:30"
   - Response: 200 OK

9. **Controller → Internal**: Schedule requeue for grace expiry
   - Action: Return reconcile.Result{RequeueAfter: 300s + small_jitter}

```
Time: 17:05:00 IST (grace period expired)
```

10. **Controller → Internal**: Check grace expiry
    - Action: now >= gracePeriodExpiry
    - Decision: Apply pending scale down

11. **Controller → Kubernetes API**: Patch Deployment
    - Action: PATCH with replicas=2
    - Response: 200 OK

12. **Controller → EventRecorder**: Emit ScaledDown event
    - Message: "Scaled down from 10 to 2 replicas after 300s grace period"

13. **Controller → Kubernetes API**: Update TWS status
    - Action: Clear gracePeriodExpiry, set effectiveReplicas=2
    - Response: 200 OK

14. **Controller → Metrics**: Update metrics
    - Action: Set gauge kyklos_effective_replicas = 2
    - Action: Increment counter kyklos_scaling_operations_total{direction="down"}

### Observable Side Effects
- 17:00: No immediate change to Deployment (grace period active)
- 17:00: Event shows grace period started
- 17:00: TWS status shows gracePeriodExpiry timestamp
- 17:05: Deployment spec.replicas changed from 10 to 2
- 17:05: Deployment controller begins terminating 8 pods
- 17:05: Event shows scale down after grace period
- Metrics show grace period duration

---

## Manual Drift Correction Sequence

### Sequence: User manually scales to 15 replicas at 14:00 IST

```
Initial state: In BusinessHours window, effectiveReplicas=10
User action: kubectl scale deployment webapp --replicas=15
```

1. **Informer → Controller**: Deployment update notification
   - Action: Informer detects Deployment change
   - Trigger: Reconcile queued immediately

2. **Controller → Cache**: Get TimeWindowScaler
   - Response: TWS expecting 10 replicas for BusinessHours

3. **Controller → Cache**: Get Deployment "webapp"
   - Response: spec.replicas=15, status.replicas=15

4. **Controller → Internal**: Detect drift
   - Action: Compare targetSpec(15) != effective(10)
   - Decision: Correction needed (pause=false)

5. **Controller → Kubernetes API**: Patch Deployment
   - Action: PATCH replicas back to 10
   - Response: 200 OK

6. **Controller → EventRecorder**: Emit ScaledDown event
   - Message: "Corrected manual drift: scaled from 15 to 10 replicas"

7. **Controller → Kubernetes API**: Update TWS status
   - Action: Update lastScaleTime
   - Response: 200 OK

8. **Controller → Metrics**: Increment drift counter
   - Action: Increment kyklos_manual_drift_corrections_total

### Observable Side Effects
- Deployment briefly has 15 replicas
- Within seconds, scaled back to 10 replicas
- Event shows drift correction
- Metrics track drift occurrence

---

## Holiday Override Sequence

### Sequence: Holiday Detected with treat-as-closed Mode

```
Date: 2025-12-25 (Thursday, normally a business day)
Configuration: holidays.mode = "treat-as-closed"
```

1. **Controller → Cache**: Get TimeWindowScaler
   - Response: TWS with holiday configuration

2. **Controller → Cache**: Get ConfigMap "company-holidays"
   - Action: Read ConfigMap from cache
   - Response: ConfigMap with key "2025-12-25"

3. **Controller → Internal**: Check holiday status
   - Action: Current date exists in ConfigMap
   - Decision: Today is a holiday

4. **Controller → Internal**: Apply holiday mode
   - Action: mode=="treat-as-closed"
   - Decision: Use defaultReplicas regardless of windows
   - Response: effectiveReplicas = 0

5. **Controller → EventRecorder**: Emit HolidayDetected event
   - Message: "Holiday detected for 2025-12-25 (mode: treat-as-closed)"

6. **Controller → EventRecorder**: Emit WindowOverride event
   - Message: "Holiday 2025-12-25: treating as closed, using 0 replicas"

7. **Controller → Kubernetes API**: Patch Deployment
   - Action: PATCH replicas to 0
   - Response: 200 OK

8. **Controller → EventRecorder**: Emit ScaledDown event
   - Message: "Scaled down from 10 to 0 replicas (holiday: treat-as-closed)"

9. **Controller → Internal**: Schedule requeue
   - Action: Next boundary = tomorrow 00:00 (end of holiday)
   - Response: Requeue in ~7 hours

### Observable Side Effects
- Deployment scaled to 0 despite being Thursday
- Multiple events explain holiday handling
- Status shows currentWindow="OffHours" (holiday override)
- Metrics show holiday override active

---

## Pause Handling Sequence

### Sequence: Scaling Needed but Controller Paused

```
Current: In window, should have 10 replicas
Actual: Deployment has 5 replicas (manual change)
TWS: spec.pause = true
```

1. **Controller → Cache**: Get TimeWindowScaler
   - Response: TWS with pause=true

2. **Controller → Internal**: Compute effective replicas
   - Action: Normal window evaluation
   - Response: effectiveReplicas = 10

3. **Controller → Cache**: Get Deployment
   - Response: spec.replicas=5, status.replicas=5

4. **Controller → Internal**: Check pause flag
   - Action: Detect pause=true
   - Decision: Skip deployment update

5. **Controller → EventRecorder**: Emit ScalingSkipped event
   - Message: "Scaling skipped due to pause: current=5, desired=10"

6. **Controller → Kubernetes API**: Update TWS status only
   - Action: Set effectiveReplicas=10, Ready=False
   - Reason: TargetMismatch
   - Message: "Target has 5 replicas but desired is 10 (pause=true)"

7. **Controller → Metrics**: Update pause metrics
   - Action: Set kyklos_pause_active{tws="webapp-office-hours"} = 1
   - Action: Set kyklos_replica_drift = 5 (difference)

8. **Controller → Internal**: Normal requeue
   - Action: Calculate next window boundary
   - Response: Requeue as normal (pause doesn't affect schedule)

### Observable Side Effects
- Deployment remains at 5 replicas (no correction)
- TWS status shows mismatch
- Ready condition is False
- Event explains why scaling skipped
- Metrics show pause active and drift amount

---

## Error Recovery Sequence

### Sequence: Target Not Found, Then Recreated

```
Initial: Deployment deleted externally
```

1. **Controller → Cache**: Get Deployment "webapp"
   - Action: Read from cache
   - Response: 404 Not Found

2. **Controller → EventRecorder**: Emit MissingTarget event
   - Message: "Target Deployment production/webapp not found"

3. **Controller → Kubernetes API**: Update TWS status
   - Action: Set Ready=False, reason=TargetNotFound
   - Response: 200 OK

4. **Controller → Internal**: Schedule error requeue
   - Action: Requeue in 30 seconds
   - Response: reconcile.Result{RequeueAfter: 30s}

```
User recreates deployment with 1 replica
```

5. **Informer → Controller**: Deployment created notification
   - Trigger: Immediate reconcile

6. **Controller → Cache**: Get Deployment "webapp"
   - Response: Deployment exists with replicas=1

7. **Controller → Internal**: Compute and compare
   - Action: effectiveReplicas(10) != targetSpec(1)
   - Decision: Scale needed

8. **Controller → Kubernetes API**: Patch Deployment
   - Action: PATCH replicas to 10
   - Response: 200 OK

9. **Controller → EventRecorder**: Emit ScaledUp event
   - Message: "Scaled up from 1 to 10 replicas (window: BusinessHours)"

10. **Controller → Kubernetes API**: Update TWS status
    - Action: Set Ready=True, reason=Reconciled
    - Response: 200 OK

### Observable Side Effects
- Period of Ready=False while target missing
- Warning events during missing period
- Automatic recovery when target recreated
- Correct replicas applied immediately on recreation