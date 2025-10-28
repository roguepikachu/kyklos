# 10-Minute Time-Based Scaling Demo

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** local-workflow-designer

This document provides a reproducible demonstration of Kyklos' time-based scaling capabilities using minute-scale windows. The entire demo completes in under 10 minutes and provides immediate, observable feedback.

---

## Demo Overview

**Total Duration:** 8-10 minutes
**Objective:** Observe automatic scaling based on time windows with 1-minute granularity
**Requirements:** Local cluster with Kyklos controller deployed

**What You'll See:**
- Deployment scales UP from 1 to 5 replicas at minute boundaries
- Deployment scales DOWN from 5 to 1 replica after window ends
- Status updates reflect window transitions
- Events show scale operations
- Controller logs explain decisions

---

## Prerequisites

Before starting, ensure your environment is ready:

```bash
# Verify all prerequisites
make verify-all
```

Expected output:
```
✓ Tools: All present and correct versions
✓ Cluster: Reachable and healthy
✓ CRDs: Installed and established
✓ Controller: Running and ready
✓ RBAC: Permissions correctly configured
```

**If any checks fail:**
- See [LOCAL-DEV-GUIDE.md](./LOCAL-DEV-GUIDE.md) for setup instructions
- See [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) for issue resolution

---

## Demo Architecture

### Components

1. **demo-app** (Deployment)
   - Simple nginx deployment
   - Target for time-based scaling
   - Initial replicas: 1

2. **demo-minute-scaler** (TimeWindowScaler)
   - Manages demo-app scaling
   - Uses UTC timezone (no DST complications)
   - Minute-scale windows for fast observation

3. **Scaling Pattern**
   - Even minutes (00, 02, 04...): Scale to 5 replicas
   - Odd minutes (01, 03, 05...): Scale to 1 replica
   - Repeats continuously

### Timeline Visualization

```
Time      00:00  00:01  00:02  00:03  00:04  00:05  00:06  00:07
          ├──────┼──────┼──────┼──────┼──────┼──────┼──────┼──────
Replicas  5      1      5      1      5      1      5      1
          ▓▓▓▓▓  ▓      ▓▓▓▓▓  ▓      ▓▓▓▓▓  ▓      ▓▓▓▓▓  ▓
Window    A      B      A      B      A      B      A      B
```

---

## Step-by-Step Demo

### Phase 1: Setup Demo Environment (1 minute)

#### Step 1.1: Create Demo Namespace and Target

```bash
make demo-setup
```

**Expected Duration:** 10-15 seconds

**What This Does:**
- Creates namespace `demo`
- Deploys `demo-app` with 1 replica
- Waits for pod to be Running

**Expected Output:**
```
Setting up demo environment...
namespace/demo created
deployment.apps/demo-app created
Waiting for demo-app to be ready...
✓ Demo namespace ready
✓ Target deployment created with 1 replica
```

**Verification:**
```bash
kubectl get deploy,pods -n demo
```

Expected state:
```
NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/demo-app   1/1     1            1           15s

NAME                            READY   STATUS    RESTARTS   AGE
pod/demo-app-7d4c8bf5c9-abc12   1/1     Running   0          15s
```

---

### Phase 2: Apply TimeWindowScaler (30 seconds)

#### Step 2.1: Understand the Configuration

Before applying, let's examine the TimeWindowScaler:

```yaml
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: demo-minute-scaler
  namespace: demo
spec:
  # Target the demo-app deployment
  targetRef:
    kind: Deployment
    name: demo-app

  # Use UTC to avoid timezone complications
  timezone: UTC

  # Default to 1 replica outside windows
  defaultReplicas: 1

  # Define minute-scale windows
  windows:
  # Scale to 5 replicas on even minutes (00, 02, 04, ...)
  - days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
    start: "00:00"
    end: "00:01"
    replicas: 5
  - days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
    start: "00:02"
    end: "00:03"
    replicas: 5
  - days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
    start: "00:04"
    end: "00:05"
    replicas: 5
  # Pattern continues for all even minutes...
```

**Key Points:**
- Windows are 1 minute long (e.g., 00:00 to 00:01)
- Windows apply every day
- Start time is inclusive, end time is exclusive
- When in window: 5 replicas
- Outside window: 1 replica (defaultReplicas)

#### Step 2.2: Apply the TimeWindowScaler

```bash
make demo-apply-minute
```

**Expected Duration:** 5 seconds

**Expected Output:**
```
Applying minute-scale demo...
timewindowscaler.kyklos.io/demo-minute-scaler created
✓ Demo TWS applied
✓ Watch for scale changes every minute
```

**Immediate Verification:**
```bash
kubectl get tws -n demo
```

Expected:
```
NAME                  WINDOW    REPLICAS   TARGET      AGE
demo-minute-scaler    OffHours  1          demo-app    5s
```

**Note:** The WINDOW value depends on current time:
- If current second is :00-:59 of an even minute: "BusinessHours" with 5 replicas
- If current second is :00-:59 of an odd minute: "OffHours" with 1 replica

---

### Phase 3: Observe Scaling in Real-Time (6-7 minutes)

This phase demonstrates the core functionality. You'll watch resources scale up and down automatically based on time windows.

#### Step 3.1: Start Watching Resources

```bash
make demo-watch
```

This runs:
```bash
watch -n 2 'kubectl get tws,deploy,pods -n demo'
```

**Expected Display:**
```
Every 2.0s: kubectl get tws,deploy,pods -n demo

NAME                                          WINDOW         REPLICAS   TARGET
timewindowscaler.kyklos.io/demo-minute-scaler BusinessHours  5          demo-app

NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/demo-app   5/5     5            5           2m15s

NAME                            READY   STATUS    RESTARTS   AGE
pod/demo-app-7d4c8bf5c9-abc12   1/1     Running   0          2m15s
pod/demo-app-7d4c8bf5c9-def34   1/1     Running   0          45s
pod/demo-app-7d4c8bf5c9-ghi56   1/1     Running   0          45s
pod/demo-app-7d4c8bf5c9-jkl78   1/1     Running   0          45s
pod/demo-app-7d4c8bf5c9-mno90   1/1     Running   0          45s
```

#### Step 3.2: What to Observe

**At Even Minute Boundaries (e.g., 14:00, 14:02, 14:04):**

Watch for these changes:
1. **WINDOW changes to "BusinessHours"**
2. **REPLICAS changes to 5**
3. **Deployment UP-TO-DATE goes to 5**
4. **New pods appear with STATUS: ContainerCreating → Running**
5. **Transition takes 10-20 seconds**

**At Odd Minute Boundaries (e.g., 14:01, 14:03, 14:05):**

Watch for these changes:
1. **WINDOW changes to "OffHours"**
2. **REPLICAS changes to 1**
3. **Deployment UP-TO-DATE goes to 1**
4. **Pods start Terminating (4 pods removed)**
5. **Transition takes 5-10 seconds**

#### Step 3.3: Typical Observation Window

**Minute 0 (Even - Scale Up):**
```
14:00:05  Window: BusinessHours, Replicas: 5, Ready: 1/5
14:00:08  Window: BusinessHours, Replicas: 5, Ready: 3/5
14:00:12  Window: BusinessHours, Replicas: 5, Ready: 5/5  ✓
```

**Minute 1 (Odd - Scale Down):**
```
14:01:03  Window: OffHours, Replicas: 1, Ready: 5/5
14:01:05  Window: OffHours, Replicas: 1, Ready: 3/5 (Terminating)
14:01:08  Window: OffHours, Replicas: 1, Ready: 1/1  ✓
```

**Minute 2 (Even - Scale Up Again):**
```
14:02:02  Window: BusinessHours, Replicas: 5, Ready: 1/5
14:02:15  Window: BusinessHours, Replicas: 5, Ready: 5/5  ✓
```

**Let this run for 4-5 complete cycles (8-10 minutes) to observe the pattern repeat.**

**Stop watching:** Press `Ctrl+C`

---

### Phase 4: Inspect Events and Logs (2 minutes)

#### Step 4.1: View Scale Events

```bash
kubectl get events -n demo --sort-by='.lastTimestamp' | tail -20
```

**Expected Events:**
```
LAST SEEN   TYPE     REASON         OBJECT                         MESSAGE
45s         Normal   ScaledUp       deployment/demo-app            Scaled up replica set demo-app-7d4c8bf5c9 to 5
2m15s       Normal   ScaledDown     deployment/demo-app            Scaled down replica set demo-app-7d4c8bf5c9 to 1
3m45s       Normal   ScaledUp       deployment/demo-app            Scaled up replica set demo-app-7d4c8bf5c9 to 5
4m15s       Normal   ScaledDown     deployment/demo-app            Scaled down replica set demo-app-7d4c8bf5c9 to 1
...
```

**Kyklos-Specific Events:**
```bash
kubectl get events -n demo --field-selector involvedObject.kind=TimeWindowScaler
```

Expected:
```
REASON              MESSAGE
WindowTransition    Entered window: BusinessHours (00:02-00:03)
ScalingTarget       Scaling demo-app from 1 to 5 replicas
WindowTransition    Exited window: BusinessHours
ScalingTarget       Scaling demo-app from 5 to 1 replicas
```

#### Step 4.2: Examine TimeWindowScaler Status

```bash
kubectl get tws demo-minute-scaler -n demo -o yaml
```

**Key Status Fields:**
```yaml
status:
  currentWindow: BusinessHours  # or OffHours
  effectiveReplicas: 5          # Current desired count
  lastScaleTime: "2025-10-28T14:02:05Z"
  targetObservedReplicas: 5     # What deployment actually has
  observedGeneration: 1         # Matches metadata.generation (no pending changes)
  conditions:
  - type: Ready
    status: "True"
    reason: Reconciled
    message: Target deployment matches desired replicas
  - type: Reconciling
    status: "False"
    reason: Stable
    message: No ongoing reconciliation
  - type: Degraded
    status: "False"
    reason: OperationalNormal
    message: No errors detected
```

**Interpreting Status:**
- `currentWindow`: Which window (or default) is active
- `effectiveReplicas == targetObservedReplicas`: System in sync
- `Ready=True`: Target matches desired state
- `observedGeneration == metadata.generation`: No pending spec changes

#### Step 4.3: View Controller Logs

```bash
make logs-controller | grep demo-app
```

**Expected Log Patterns:**

**Scale-Up Decision:**
```
INFO  Reconciling TimeWindowScaler  {"namespace": "demo", "name": "demo-minute-scaler"}
INFO  Current time in UTC: 2025-10-28T14:02:00Z
INFO  Matched window: BusinessHours (00:02-00:03) -> 5 replicas
INFO  Target deployment has 1 replicas, desired 5 replicas
INFO  Scaling deployment demo-app from 1 to 5 replicas
INFO  Requeue scheduled at next window boundary: 2025-10-28T14:03:00Z
```

**Scale-Down Decision:**
```
INFO  Reconciling TimeWindowScaler  {"namespace": "demo", "name": "demo-minute-scaler"}
INFO  Current time in UTC: 2025-10-28T14:03:05Z
INFO  No matching windows, using defaultReplicas: 1
INFO  Target deployment has 5 replicas, desired 1 replicas
INFO  Scaling deployment demo-app from 5 to 1 replicas
INFO  Requeue scheduled at next window boundary: 2025-10-28T14:04:00Z
```

**Key Observations:**
- Controller reconciles at window boundaries
- Logs explain why each scaling decision was made
- Next reconcile time is predictable (next window boundary)

---

### Phase 5: Demonstrate Manual Drift Correction (2 minutes)

This shows Kyklos correcting manual changes to maintain desired state.

#### Step 5.1: Manually Scale Deployment

While in an "in-window" period (even minute), manually change replicas:

```bash
# Wait for even minute (e.g., 14:04:XX)
# Verify current state
kubectl get tws,deploy -n demo

# Manually scale to 8 replicas
kubectl scale deployment demo-app -n demo --replicas=8
```

**Expected Output:**
```
deployment.apps/demo-app scaled
```

#### Step 5.2: Observe Automatic Correction

Watch Kyklos detect and correct the drift:

```bash
watch kubectl get tws,deploy -n demo
```

**Timeline:**
```
T+0s    Manual change: Replicas=8
T+5s    Controller reconciles (next cycle)
T+5s    Detects drift: observed=8, desired=5
T+5s    Corrects: scales back to 5
T+10s   Deployment stabilizes at 5 replicas
```

**Check Events:**
```bash
kubectl get events -n demo | grep drift
```

Expected event:
```
Normal  DriftCorrected  Corrected manual drift on demo-app from 8 to 5 replicas
```

**Controller Logs:**
```bash
make logs-controller | tail -20
```

Expected:
```
WARN  Manual drift detected  {"deployment": "demo-app", "observed": 8, "desired": 5}
INFO  Correcting drift: scaling demo-app to 5 replicas
INFO  Drift correction complete
```

**Key Takeaway:** Kyklos continuously enforces desired state, correcting manual changes automatically.

---

### Phase 6: Test Pause Functionality (1 minute)

Demonstrate pausing the controller without deleting the TimeWindowScaler.

#### Step 6.1: Pause Scaling

```bash
kubectl patch tws demo-minute-scaler -n demo --type=merge -p '{"spec":{"pause":true}}'
```

**Expected Output:**
```
timewindowscaler.kyklos.io/demo-minute-scaler patched
```

#### Step 6.2: Observe Paused Behavior

```bash
kubectl get tws demo-minute-scaler -n demo -o yaml | grep -A 10 status
```

**Expected Status:**
```yaml
status:
  currentWindow: BusinessHours
  effectiveReplicas: 5          # Still computed
  targetObservedReplicas: 1     # But not applied
  conditions:
  - type: Ready
    status: "False"             # Not ready due to mismatch
    reason: TargetMismatch
    message: "Paused: target has 1 replicas, desired 5 replicas (scaling suspended)"
```

**Key Points:**
- Controller still reconciles and computes desired state
- Status shows what WOULD happen
- No actual scaling occurs while paused
- Ready condition indicates mismatch but explains it's due to pause

#### Step 6.3: Resume Scaling

```bash
kubectl patch tws demo-minute-scaler -n demo --type=merge -p '{"spec":{"pause":false}}'
```

Within 10-15 seconds, scaling resumes and target is corrected.

---

### Phase 7: Cleanup (30 seconds)

#### Step 7.1: Remove Demo Resources

```bash
make demo-cleanup
```

**Expected Duration:** 5-10 seconds

**Expected Output:**
```
Cleaning up demo...
namespace "demo" deleted
✓ Demo resources removed
```

**Verification:**
```bash
kubectl get ns demo
```

Expected:
```
Error from server (NotFound): namespaces "demo" not found
```

---

## Demo Variations

### Variation 1: Add Grace Period

Modify the TimeWindowScaler to add a grace period for downscaling:

```bash
kubectl patch tws demo-minute-scaler -n demo --type=merge -p '{"spec":{"gracePeriodSeconds":30}}'
```

**Observe:**
- Scale-ups happen immediately
- Scale-downs wait 30 seconds after window ends
- Status shows grace period countdown

### Variation 2: Change Default Replicas

Change what happens outside windows:

```bash
kubectl patch tws demo-minute-scaler -n demo --type=merge -p '{"spec":{"defaultReplicas":3}}'
```

**Observe:**
- Odd minutes now scale to 3 instead of 1
- Even minutes still scale to 5

### Variation 3: Continuous High-Frequency Pattern

For extreme testing, apply windows every 30 seconds:

```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
  start: "00:00"
  end: "00:01"
  replicas: 10
```

Then apply at :00, :02, :04... seconds for very rapid scaling cycles.

---

## Metrics to Observe

During the demo, these metrics are available on the controller:

```bash
make port-forward-metrics
# In another terminal:
curl http://localhost:8080/metrics | grep kyklos
```

**Key Metrics:**

```prometheus
# Total reconciliations
kyklos_reconcile_total{controller="timewindowscaler"} 45

# Reconciliation duration
kyklos_reconcile_duration_seconds_sum 2.34
kyklos_reconcile_duration_seconds_count 45

# Scale operations
kyklos_scale_operations_total{direction="up"} 22
kyklos_scale_operations_total{direction="down"} 23

# Window transitions
kyklos_window_transitions_total 45

# Current state
kyklos_effective_replicas{namespace="demo",name="demo-minute-scaler"} 5
```

---

## Troubleshooting the Demo

### Issue: No Scaling Occurs

**Symptoms:**
- Deployment stays at 1 replica
- No events generated
- Controller logs show no activity

**Diagnosis:**
```bash
# Check controller is running
kubectl get pods -n kyklos-system

# Check controller logs for errors
make logs-controller | grep ERROR

# Verify TWS was created
kubectl get tws -n demo
```

**Fixes:**
- Restart controller: `make restart-controller`
- Verify RBAC: `make verify-rbac`
- Check CRD installed: `kubectl get crd timewindowscalers.kyklos.io`

### Issue: Scaling Happens at Wrong Times

**Symptoms:**
- Scales at unexpected minutes
- Pattern doesn't match even/odd

**Diagnosis:**
```bash
# Check current UTC time
date -u

# Check controller's time interpretation
make logs-controller | grep "Current time"
```

**Cause:** Timezone confusion (controller uses UTC, demo uses local time)

**Fix:** Always use UTC timezone in demo configurations

### Issue: Pods Stuck in Pending

**Symptoms:**
- Replicas increase but pods don't become Ready
- Deployment shows "1/5" Ready

**Diagnosis:**
```bash
kubectl describe pods -n demo | grep -A 5 Events
```

**Common Causes:**
- Insufficient cluster resources
- Image pull errors
- Node capacity limits

**Fix:**
- Reduce replicas in windows to 2-3 instead of 5
- Ensure Docker images are available locally

### Issue: Events Not Showing

**Symptoms:**
- Scaling occurs but no events appear

**Diagnosis:**
```bash
# Check controller can create events
kubectl auth can-i create events -n demo --as=system:serviceaccount:kyklos-system:kyklos-controller
```

**Fix:**
- Verify RBAC: `make verify-rbac`
- Check controller logs for permission errors

---

## Expected Outcomes

After completing this demo, you should have observed:

1. **Automatic Scale-Ups** - Deployment scales from 1 to 5 replicas at even minute boundaries
2. **Automatic Scale-Downs** - Deployment scales from 5 to 1 replica at odd minute boundaries
3. **Status Accuracy** - TimeWindowScaler status reflects actual state continuously
4. **Event Generation** - Clear events explain each scaling decision
5. **Controller Logs** - Detailed logs show time evaluation and decisions
6. **Manual Drift Correction** - Manual changes are automatically reverted
7. **Pause Functionality** - Pausing suspends scaling but maintains status updates

---

## Next Steps

**For Further Exploration:**

1. **Real-World Scenarios:**
   - Apply examples from `examples/tws-office-hours.yaml`
   - Create multi-hour windows
   - Test cross-midnight windows

2. **Advanced Features:**
   - Add holiday support with ConfigMap
   - Test grace period behavior in detail
   - Observe DST transitions (use local timezone)

3. **Integration Testing:**
   - Monitor with Prometheus
   - Set up alerting on scaling events
   - Create dashboards for replica trends

4. **Performance Testing:**
   - Scale to 100+ replicas
   - Test with multiple TimeWindowScalers
   - Measure reconcile loop performance

**Documentation References:**
- [LOCAL-DEV-GUIDE.md](./LOCAL-DEV-GUIDE.md) - Setup and workflows
- [VERIFY-CHECKLIST.md](./VERIFY-CHECKLIST.md) - Health checks
- [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) - Issue resolution
- [CRD-SPEC.md](./api/CRD-SPEC.md) - Complete API reference

---

## Demo Timing Summary

| Phase | Duration | Cumulative |
|-------|----------|------------|
| 1. Setup Environment | 1m | 1m |
| 2. Apply TimeWindowScaler | 0.5m | 1.5m |
| 3. Observe Scaling | 6m | 7.5m |
| 4. Inspect Events/Logs | 2m | 9.5m |
| 5. Manual Drift Correction | 2m | 11.5m |
| 6. Test Pause | 1m | 12.5m |
| 7. Cleanup | 0.5m | 13m |

**Minimum Observable Demo:** Phases 1-3 only = 7.5 minutes

**Complete Walkthrough:** All phases = 13 minutes

**For time-constrained demos:** Skip phases 5-6, focus on phases 1-4 for 10-minute window.
