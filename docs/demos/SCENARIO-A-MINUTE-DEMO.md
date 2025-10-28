# Scenario A: Minute-Scale Time Window Demo

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** demo-scenario-designer

## Overview

A reproducible 10-minute demonstration showcasing Kyklos' time-based scaling with minute-granularity windows. This scenario uses UTC timezone for maximum reliability and demonstrates the complete scaling lifecycle.

---

## Demo Specifications

| Property | Value |
|----------|-------|
| **Total Duration** | 10 minutes maximum |
| **Timezone** | UTC (no DST complications) |
| **Scale Pattern** | 0 → 2 → 0 replicas |
| **Window Length** | 3 minutes active, remainder idle |
| **Target** | nginx web Deployment |
| **Cluster** | Kind or k3d local cluster |

---

## Prerequisites Checklist

Before starting, verify all components are ready:

```bash
# 1. Complete verification
make verify-all
```

**Expected Output:**
```
✓ Tools: All present and correct versions
✓ Cluster: Reachable and healthy
✓ CRDs: Installed and established
✓ Controller: Running and ready
✓ RBAC: Permissions correctly configured
```

**If verification fails:**
- Tools missing: Follow [LOCAL-DEV-GUIDE.md](/Users/aykumar/personal/kyklos/docs/LOCAL-DEV-GUIDE.md) setup
- Cluster issues: Run `make cluster-down && make cluster-up`
- Controller issues: Run `make redeploy`

---

## Demo Architecture

### Components

**1. webapp-demo (Deployment)**
- Image: nginx:alpine
- Initial replicas: 0 (scaled by TWS)
- Resource limits: 50m CPU, 64Mi memory
- Namespace: demo

**2. webapp-minute-scaler (TimeWindowScaler)**
- Manages webapp-demo
- Timezone: UTC
- Default replicas: 0 (outside windows)
- Active window: T+1min to T+4min (3-minute duration)
- Window replicas: 2

**3. Scaling Timeline**

```
Time    T+0  T+1  T+2  T+3  T+4  T+5  T+6  T+7  T+8  T+9  T+10
        ├────┼────┼────┼────┼────┼────┼────┼────┼────┼────┼────
Replicas 0    2    2    2    0    0    0    0    0    0    0
        │    ▓▓▓▓▓▓▓▓▓▓▓▓▓▓│
Window  │    BusinessHours  │    OffHours
        │                   │
Events  │    ScaleUp        │    ScaleDown
```

---

## Step-by-Step Operator Guide

### T-1: Pre-Demo Setup (30 seconds)

#### Step 0.1: Calculate Window Times

**IMPORTANT:** Calculate exact UTC window times before starting.

```bash
# Get current UTC time
date -u +"%H:%M:%S"

# Calculate T+1min and T+4min
# Example: If current time is 14:37:45 UTC
#   T+1min = 14:38:00 (round up to next minute)
#   T+4min = 14:41:00
```

**Write down your times:**
```
Current UTC: __:__:__
T+1min (start): __:__
T+4min (end): __:__
```

#### Step 0.2: Prepare Demo Manifest

Create the TimeWindowScaler manifest with your calculated times:

```yaml
# Save as /tmp/demo-minute-scaler.yaml
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: webapp-minute-scaler
  namespace: demo
spec:
  targetRef:
    kind: Deployment
    name: webapp-demo

  timezone: UTC
  defaultReplicas: 0

  windows:
  - days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
    start: "14:38"  # REPLACE WITH YOUR T+1min
    end: "14:41"    # REPLACE WITH YOUR T+4min
    replicas: 2
```

**Critical:** Replace start/end times with your calculated values.

---

### T+0: Create Demo Environment (15 seconds)

#### Step 1.1: Create Namespace and Deployment

```bash
# T+0:00 - Create demo namespace
kubectl create namespace demo

# T+0:05 - Create target deployment with 0 replicas
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp-demo
  namespace: demo
  labels:
    app: webapp-demo
spec:
  replicas: 0
  selector:
    matchLabels:
      app: webapp-demo
  template:
    metadata:
      labels:
        app: webapp-demo
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 100m
            memory: 128Mi
EOF
```

**Expected Output:**
```
namespace/demo created
deployment.apps/webapp-demo created
```

#### Step 1.2: Verify Initial State

```bash
# T+0:10 - Verify deployment exists with 0 replicas
kubectl get deploy,pods -n demo
```

**Expected State:**
```
NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   0/0     0            0           5s

No pods found in demo namespace.
```

**CAPTURE POINT 1:** Screenshot initial state showing 0/0 replicas.

---

### T+0:15: Apply TimeWindowScaler (10 seconds)

#### Step 2.1: Apply TWS Manifest

```bash
# T+0:15 - Apply the TimeWindowScaler
kubectl apply -f /tmp/demo-minute-scaler.yaml
```

**Expected Output:**
```
timewindowscaler.kyklos.io/webapp-minute-scaler created
```

#### Step 2.2: Verify TWS Created

```bash
# T+0:20 - Check TWS status
kubectl get tws -n demo
```

**Expected Output:**
```
NAME                    WINDOW     REPLICAS   TARGET        AGE
webapp-minute-scaler    OffHours   0          webapp-demo   5s
```

**Observations:**
- WINDOW shows "OffHours" (we're before T+1min)
- REPLICAS shows 0 (defaultReplicas)
- TARGET correctly references webapp-demo

**CAPTURE POINT 2:** Screenshot TWS in OffHours state.

---

### T+0:25 to T+0:55: Pre-Window Observation (30 seconds)

#### Step 3.1: Monitor Resources

```bash
# T+0:25 - Start watching all demo resources
watch -n 2 'date -u && echo && kubectl get tws,deploy,pods -n demo'
```

**Expected Display (before T+1min):**
```
Mon Oct 28 14:37:45 UTC 2025

NAME                                        WINDOW     REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-minute-scaler OffHours   0          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   0/0     0            0           45s

No pods in demo namespace.
```

**What to Observe:**
- Current UTC time updating every 2 seconds
- WINDOW remains "OffHours"
- REPLICAS remains 0
- No pods present

**Duration:** Continue watching until T+1min boundary approaches.

---

### T+1:00: Window Opens - Scale Up Event (30 seconds)

#### Step 4.1: Observe Scale-Up Transition

**Keep watching** - You will see these changes within 5-10 seconds after T+1min:

**T+1:05 (approximately):**
```
Mon Oct 28 14:38:05 UTC 2025

NAME                                        WINDOW         REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-minute-scaler BusinessHours  2          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   0/2     2            0           1m5s

NAME                              READY   STATUS              RESTARTS   AGE
pod/webapp-demo-7d8f9c5b4-abc12   0/1     ContainerCreating   0          2s
pod/webapp-demo-7d8f9c5b4-def34   0/1     ContainerCreating   0          2s
```

**Key Changes:**
1. WINDOW changed from "OffHours" to "BusinessHours"
2. REPLICAS changed from 0 to 2
3. Deployment UP-TO-DATE shows 2
4. Two new pods appeared with STATUS: ContainerCreating

**T+1:15 (approximately):**
```
Mon Oct 28 14:38:15 UTC 2025

NAME                                        WINDOW         REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-minute-scaler BusinessHours  2          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   2/2     2            2           1m15s

NAME                              READY   STATUS    RESTARTS   AGE
pod/webapp-demo-7d8f9c5b4-abc12   1/1     Running   0          12s
pod/webapp-demo-7d8f9c5b4-def34   1/1     Running   0          12s
```

**Final State:**
- Both pods Running
- Deployment shows 2/2 Ready
- Scale-up complete in ~15 seconds

**CAPTURE POINT 3:** Screenshot showing BusinessHours window with 2/2 Running pods.

#### Step 4.2: Inspect Scale-Up Events

Open a second terminal and capture events:

```bash
# T+1:20 - View recent events
kubectl get events -n demo --sort-by='.lastTimestamp' | tail -10
```

**Expected Events:**
```
LAST SEEN   TYPE     REASON              OBJECT                         MESSAGE
15s         Normal   WindowTransition    timewindowscaler/webapp-...   Entered window: BusinessHours (14:38-14:41)
15s         Normal   ScalingTarget       timewindowscaler/webapp-...   Scaling webapp-demo from 0 to 2 replicas
14s         Normal   ScaledUp            deployment/webapp-demo        Scaled up replica set webapp-demo-7d8f9c5b4 to 2
13s         Normal   SuccessfulCreate    replicaset/webapp-demo-...    Created pod: webapp-demo-7d8f9c5b4-abc12
13s         Normal   SuccessfulCreate    replicaset/webapp-demo-...    Created pod: webapp-demo-7d8f9c5b4-def34
5s          Normal   Pulling             pod/webapp-demo-...           Pulling image "nginx:alpine"
3s          Normal   Pulled              pod/webapp-demo-...           Successfully pulled image
2s          Normal   Started             pod/webapp-demo-...           Started container nginx
```

**CAPTURE POINT 4:** Screenshot of scale-up events.

#### Step 4.3: Examine Controller Logs

```bash
# T+1:25 - Check controller decision logs
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=20 | grep webapp
```

**Expected Log Pattern:**
```
INFO  Reconciling TimeWindowScaler  {"namespace": "demo", "name": "webapp-minute-scaler"}
INFO  Current time in UTC: 2025-10-28T14:38:05Z
INFO  Matched window: BusinessHours (14:38-14:41) -> 2 replicas
INFO  Target deployment has 0 replicas, desired 2 replicas
INFO  Scaling deployment webapp-demo from 0 to 2 replicas
INFO  Successfully scaled deployment  {"deployment": "webapp-demo", "from": 0, "to": 2}
INFO  Requeue scheduled at next window boundary: 2025-10-28T14:41:00Z
```

**Key Observations:**
- Controller detected window entry
- Correctly computed desired replicas as 2
- Performed scale-up action
- Scheduled next reconcile at T+4min (window end)

**CAPTURE POINT 5:** Screenshot of controller logs showing scale-up decision.

---

### T+1:30 to T+3:55: In-Window Steady State (2.5 minutes)

#### Step 5.1: Verify Stable State

Return to the watch terminal. You should see stable state:

```bash
# Watch continues to show:
Mon Oct 28 14:39:30 UTC 2025

NAME                                        WINDOW         REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-minute-scaler BusinessHours  2          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   2/2     2            2           2m30s

NAME                              READY   STATUS    RESTARTS   AGE
pod/webapp-demo-7d8f9c5b4-abc12   1/1     Running   0          1m27s
pod/webapp-demo-7d8f9c5b4-def34   1/1     Running   0          1m27s
```

**Observations:**
- Window remains "BusinessHours"
- Replicas remain 2
- Pods remain Running
- No changes during this period (expected behavior)

#### Step 5.2: Examine TWS Status Detail

In the second terminal:

```bash
# T+2:00 - Examine detailed status
kubectl get tws webapp-minute-scaler -n demo -o yaml | grep -A 20 status:
```

**Expected Status:**
```yaml
status:
  currentWindow: BusinessHours
  effectiveReplicas: 2
  lastScaleTime: "2025-10-28T14:38:05Z"
  targetObservedReplicas: 2
  observedGeneration: 1
  conditions:
  - type: Ready
    status: "True"
    reason: Reconciled
    message: Target deployment matches desired replicas
    lastTransitionTime: "2025-10-28T14:38:15Z"
  - type: Reconciling
    status: "False"
    reason: Stable
    message: No ongoing reconciliation
    lastTransitionTime: "2025-10-28T14:38:15Z"
  - type: Degraded
    status: "False"
    reason: OperationalNormal
    message: No errors detected
    lastTransitionTime: "2025-10-28T14:38:15Z"
```

**Status Interpretation:**
- `currentWindow: BusinessHours` - Active window identified
- `effectiveReplicas: 2` - Desired state
- `targetObservedReplicas: 2` - Actual state
- `Ready=True` - System in sync
- `observedGeneration: 1` - No pending spec changes

**CAPTURE POINT 6:** Screenshot of detailed TWS status showing all conditions.

---

### T+4:00: Window Closes - Scale Down Event (30 seconds)

#### Step 6.1: Observe Scale-Down Transition

**Keep watching** - You will see these changes within 5-10 seconds after T+4min:

**T+4:05 (approximately):**
```
Mon Oct 28 14:41:05 UTC 2025

NAME                                        WINDOW     REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-minute-scaler OffHours   0          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   2/2     0            2           4m5s

NAME                              READY   STATUS        RESTARTS   AGE
pod/webapp-demo-7d8f9c5b4-abc12   1/1     Terminating   0          3m2s
pod/webapp-demo-7d8f9c5b4-def34   1/1     Terminating   0          3m2s
```

**Key Changes:**
1. WINDOW changed from "BusinessHours" to "OffHours"
2. REPLICAS changed from 2 to 0
3. Deployment UP-TO-DATE shows 0
4. Pods show STATUS: Terminating

**T+4:15 (approximately):**
```
Mon Oct 28 14:41:15 UTC 2025

NAME                                        WINDOW     REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-minute-scaler OffHours   0          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   0/0     0            0           4m15s

No pods in demo namespace.
```

**Final State:**
- All pods terminated
- Deployment shows 0/0
- Back to initial state
- Scale-down complete in ~10 seconds

**CAPTURE POINT 7:** Screenshot showing OffHours window with 0/0 replicas.

#### Step 6.2: Inspect Scale-Down Events

```bash
# T+4:20 - View recent events
kubectl get events -n demo --sort-by='.lastTimestamp' | tail -10
```

**Expected Events:**
```
LAST SEEN   TYPE     REASON              OBJECT                         MESSAGE
15s         Normal   WindowTransition    timewindowscaler/webapp-...   Exited window: BusinessHours
15s         Normal   ScalingTarget       timewindowscaler/webapp-...   Scaling webapp-demo from 2 to 0 replicas
14s         Normal   ScaledDown          deployment/webapp-demo        Scaled down replica set webapp-demo-7d8f9c5b4 to 0
12s         Normal   Killing             pod/webapp-demo-...           Stopping container nginx
10s         Normal   Killing             pod/webapp-demo-...           Stopping container nginx
```

**CAPTURE POINT 8:** Screenshot of scale-down events.

#### Step 6.3: Examine Controller Logs

```bash
# T+4:25 - Check controller decision logs
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=20 | grep webapp
```

**Expected Log Pattern:**
```
INFO  Reconciling TimeWindowScaler  {"namespace": "demo", "name": "webapp-minute-scaler"}
INFO  Current time in UTC: 2025-10-28T14:41:05Z
INFO  No matching windows, using defaultReplicas: 0
INFO  Target deployment has 2 replicas, desired 0 replicas
INFO  Scaling deployment webapp-demo from 2 to 0 replicas
INFO  Successfully scaled deployment  {"deployment": "webapp-demo", "from": 2, "to": 0}
INFO  Requeue scheduled at next window boundary: 2025-10-29T14:38:00Z
```

**Key Observations:**
- Controller detected window exit
- Correctly computed desired replicas as 0 (defaultReplicas)
- Performed scale-down action
- Scheduled next reconcile for tomorrow at T+1min

**CAPTURE POINT 9:** Screenshot of controller logs showing scale-down decision.

---

### T+4:30 to T+9:55: Post-Window Verification (5.5 minutes)

#### Step 7.1: Confirm Stable State

Watch terminal should continue showing:

```bash
Mon Oct 28 14:45:00 UTC 2025

NAME                                        WINDOW     REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-minute-scaler OffHours   0          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   0/0     0            0           8m

No pods in demo namespace.
```

**Observations:**
- Window remains "OffHours"
- Replicas remain 0
- No pods present
- System stable and idle

#### Step 7.2: Verify Complete Event History

```bash
# T+5:00 - View full event timeline
kubectl get events -n demo --sort-by='.lastTimestamp'
```

**Expected Full Timeline:**
```
LAST SEEN   TYPE     REASON              OBJECT                         MESSAGE
4m15s       Normal   WindowTransition    timewindowscaler/webapp-...   Entered window: BusinessHours (14:38-14:41)
4m15s       Normal   ScalingTarget       timewindowscaler/webapp-...   Scaling webapp-demo from 0 to 2 replicas
4m14s       Normal   ScaledUp            deployment/webapp-demo        Scaled up to 2
4m13s       Normal   SuccessfulCreate    replicaset/webapp-demo-...    Created pods
1m15s       Normal   WindowTransition    timewindowscaler/webapp-...   Exited window: BusinessHours
1m15s       Normal   ScalingTarget       timewindowscaler/webapp-...   Scaling webapp-demo from 2 to 0 replicas
1m14s       Normal   ScaledDown          deployment/webapp-demo        Scaled down to 0
1m12s       Normal   Killing             pod/webapp-demo-...           Stopping containers
```

**CAPTURE POINT 10:** Screenshot of complete event timeline showing full lifecycle.

---

### T+10:00: Cleanup (15 seconds)

#### Step 8.1: Remove Demo Resources

Press Ctrl+C to stop watching, then:

```bash
# T+10:00 - Clean up all demo resources
kubectl delete namespace demo
```

**Expected Output:**
```
namespace "demo" deleted
```

#### Step 8.2: Verify Cleanup

```bash
# T+10:10 - Verify namespace gone
kubectl get namespace demo
```

**Expected Output:**
```
Error from server (NotFound): namespaces "demo" not found
```

**Demo Complete!**

---

## Success Criteria

The demo is considered successful if ALL of the following are observed:

### Scale-Up Phase (T+1min)
- [ ] WINDOW changed to "BusinessHours" within 10 seconds of T+1min
- [ ] REPLICAS changed from 0 to 2
- [ ] 2 pods created with STATUS: ContainerCreating → Running
- [ ] WindowTransition and ScalingTarget events generated
- [ ] Controller logs show window match and scale decision
- [ ] Transition completed within 20 seconds

### Steady State Phase (T+1min to T+4min)
- [ ] WINDOW remained "BusinessHours" throughout
- [ ] REPLICAS remained 2 throughout
- [ ] Both pods remained Running with 0 restarts
- [ ] TWS status shows Ready=True, effectiveReplicas=2
- [ ] No unexpected events or errors

### Scale-Down Phase (T+4min)
- [ ] WINDOW changed to "OffHours" within 10 seconds of T+4min
- [ ] REPLICAS changed from 2 to 0
- [ ] Pods transitioned to Terminating → removed
- [ ] WindowTransition and ScalingTarget events generated
- [ ] Controller logs show window exit and scale decision
- [ ] Transition completed within 15 seconds

### System Health
- [ ] No controller restarts during demo
- [ ] All TWS conditions remained healthy (no Degraded=True)
- [ ] Events show clear causality chain
- [ ] Logs show deterministic requeue scheduling

---

## Recovery Procedures

### Issue: Window didn't open at T+1min

**Symptoms:**
- WINDOW still shows "OffHours" after T+1min + 15 seconds
- REPLICAS remains 0
- No scale-up events

**Diagnosis:**
```bash
# Check TWS configuration
kubectl get tws webapp-minute-scaler -n demo -o yaml | grep -A 5 windows:

# Check controller logs for errors
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=50 | grep ERROR
```

**Possible Causes:**
1. **Incorrect window times in manifest** - Start/end times don't match current UTC
2. **Controller not running** - Check pod status
3. **Timezone parsing error** - Controller logs will show error

**Fixes:**
```bash
# Fix 1: Recalculate and reapply TWS with correct times
kubectl delete tws webapp-minute-scaler -n demo
# Edit /tmp/demo-minute-scaler.yaml with new times (T+1min from now)
kubectl apply -f /tmp/demo-minute-scaler.yaml

# Fix 2: Restart controller
kubectl rollout restart deployment -n kyklos-system kyklos-controller-manager

# Fix 3: Check timezone
kubectl logs -n kyklos-system -l app=kyklos-controller | grep timezone
```

---

### Issue: Pods stuck in ContainerCreating

**Symptoms:**
- WINDOW shows "BusinessHours"
- REPLICAS shows 2
- Pods stuck in ContainerCreating for > 30 seconds

**Diagnosis:**
```bash
# Check pod events
kubectl describe pods -n demo | grep -A 10 Events

# Check node resources
kubectl describe node
```

**Possible Causes:**
1. **Image pull errors** - nginx:alpine not available
2. **Insufficient resources** - Node has no capacity

**Fixes:**
```bash
# Fix 1: Pre-pull image
docker pull nginx:alpine
kind load docker-image nginx:alpine --name kyklos-dev

# Fix 2: Reduce resource requests in manifest
# Edit deployment to request less CPU/memory
```

---

### Issue: Scale-down didn't happen at T+4min

**Symptoms:**
- WINDOW changed to "OffHours"
- REPLICAS still shows 2 (not 0)
- Pods still Running

**Diagnosis:**
```bash
# Check TWS status
kubectl get tws webapp-minute-scaler -n demo -o yaml | grep -A 5 status:

# Check for manual interference
kubectl get events -n demo | grep manual
```

**Possible Causes:**
1. **defaultReplicas not set to 0** - Check spec
2. **Manual deployment scaling** - Someone scaled manually
3. **Controller permission issue** - Cannot write to deployment

**Fixes:**
```bash
# Fix 1: Verify and update TWS spec
kubectl patch tws webapp-minute-scaler -n demo --type=merge -p '{"spec":{"defaultReplicas":0}}'

# Fix 2: Delete manual scale changes (controller will correct)
# Just wait 10-15 seconds for next reconcile

# Fix 3: Check RBAC
kubectl auth can-i update deployments -n demo --as=system:serviceaccount:kyklos-system:kyklos-controller
```

---

## Timing Variations

### Fast Mode (6 minutes total)

For time-constrained demonstrations:

**Modified Window:**
```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
  start: "14:38"  # T+1min
  end: "14:40"    # T+3min (shortened from T+4min)
  replicas: 2
```

**Timeline:**
- T+0: Setup (15s)
- T+1: Scale up (20s)
- T+1:20 to T+2:55: Observe (1m 35s)
- T+3: Scale down (15s)
- T+3:15 to T+5:50: Verify (2m 35s)
- T+6: Cleanup (15s)

---

### Extended Mode (15 minutes total)

For comprehensive demonstrations with additional testing:

**Additional Steps:**

**T+5:00 - Manual Drift Test:**
```bash
# Manually scale up
kubectl scale deployment webapp-demo -n demo --replicas=3

# Watch controller correct within 15 seconds
# Expected: Scales back to 0 (defaultReplicas)
```

**T+6:00 - Pause Test:**
```bash
# Pause the TWS
kubectl patch tws webapp-minute-scaler -n demo --type=merge -p '{"spec":{"pause":true}}'

# Manually scale
kubectl scale deployment webapp-demo -n demo --replicas=1

# Observe: TWS status shows mismatch but doesn't correct
# Resume
kubectl patch tws webapp-minute-scaler -n demo --type=merge -p '{"spec":{"pause":false}}'

# Observe: Controller corrects to 0 within 15 seconds
```

---

## Handoff to Docs Writer

### Materials to Provide

**1. Screenshots (10 total):**
- CAPTURE POINT 1: Initial state (0/0 replicas)
- CAPTURE POINT 2: TWS in OffHours state
- CAPTURE POINT 3: BusinessHours with 2/2 Running pods
- CAPTURE POINT 4: Scale-up events
- CAPTURE POINT 5: Controller logs - scale-up decision
- CAPTURE POINT 6: TWS detailed status with conditions
- CAPTURE POINT 7: OffHours with 0/0 replicas
- CAPTURE POINT 8: Scale-down events
- CAPTURE POINT 9: Controller logs - scale-down decision
- CAPTURE POINT 10: Complete event timeline

**2. Terminal Recordings:**
- Watch output showing full T+0 to T+10 timeline
- Controller logs with filtered grep for webapp

**3. Manifest Files:**
- /tmp/demo-minute-scaler.yaml (with actual times used)
- Deployment manifest from Step 1.1

**4. Key Outputs:**
- Complete kubectl get tws output (before, during, after)
- Complete kubectl get events output
- TWS status YAML (full detail)

**5. Timing Data:**
- Actual timestamps for each phase
- Delta times between boundaries and actions
- Total demo duration

### Best Screenshots for README

**Primary Hero Shot:** CAPTURE POINT 3
- Shows active BusinessHours window
- 2/2 pods Running
- Clean, complete state

**Supporting Screenshots:**
1. CAPTURE POINT 1 (before) + CAPTURE POINT 7 (after) side-by-side
2. CAPTURE POINT 4 or CAPTURE POINT 8 (events showing causality)
3. CAPTURE POINT 5 or CAPTURE POINT 9 (controller decision logs)

**For Documentation:**
- CAPTURE POINT 6: Status conditions (for troubleshooting guide)
- CAPTURE POINT 10: Event timeline (for concepts explanation)

---

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2025-10-28 | 1.0 | Initial scenario design with 10-minute constraint |
