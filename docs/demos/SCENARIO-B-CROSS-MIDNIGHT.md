# Scenario B: Cross-Midnight Window Demo

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** demo-scenario-designer

## Overview

A reproducible demonstration showcasing Kyklos' cross-midnight window handling with DST-aware timezone calculations. This scenario uses Europe/Berlin timezone and simulates a night shift window that crosses calendar day boundaries.

---

## Demo Specifications

| Property | Value |
|----------|-------|
| **Total Duration** | 10 minutes maximum |
| **Timezone** | Europe/Berlin (DST-aware) |
| **Scale Pattern** | 0 → 3 → 0 replicas |
| **Window Pattern** | Cross-midnight (22:00 to 02:00) |
| **Simulation Method** | Time offset to reach boundary |
| **Target** | nginx web Deployment |
| **Cluster** | Kind or k3d local cluster |

---

## Cross-Midnight Concept

### Window Definition

A cross-midnight window is defined when the end time is earlier than the start time:

```yaml
windows:
- days: [Fri]
  start: "22:00"  # 10:00 PM Friday
  end: "02:00"    # 2:00 AM Saturday (next day)
  replicas: 3
```

**Semantic Interpretation:**
- Window begins at 22:00 on specified day (Friday)
- Window extends past midnight into the next calendar day (Saturday)
- Window ends at 02:00 on the next day (Saturday)
- Total duration: 4 hours spanning two calendar days

### Boundary Computation

**Current Time: Friday 23:30 (in window)**
```
Fri 22:00 ─────────────── Fri 23:30 (now) ───────────────── Sat 02:00
          └─────────────────── Active Window ───────────────────┘
          Start                                                 End
```

**Current Time: Saturday 01:00 (in window)**
```
Fri 22:00 ────────────── Midnight ────── Sat 01:00 (now) ─── Sat 02:00
          └─────────────────────── Active Window ────────────────┘
```

**Current Time: Saturday 03:00 (outside window)**
```
Fri 22:00 ───── Midnight ───── Sat 02:00 ─── Sat 03:00 (now)
          └───── Window Ended ─────┘           Outside
```

---

## Prerequisites Checklist

```bash
# 1. Verify environment ready
make verify-all

# 2. Verify system can handle timezone calculations
timedatectl list-timezones | grep Europe/Berlin
# Expected: Europe/Berlin
```

**System Requirements:**
- Cluster with Kyklos controller running
- System IANA timezone database includes Europe/Berlin
- Internet connectivity (for timezone data if needed)

---

## Demo Architecture

### Components

**1. nightshift-demo (Deployment)**
- Image: nginx:alpine
- Initial replicas: 0
- Resource limits: 50m CPU, 64Mi memory
- Namespace: demo

**2. nightshift-scaler (TimeWindowScaler)**
- Manages nightshift-demo
- Timezone: Europe/Berlin
- Default replicas: 0
- Cross-midnight window: Friday 22:00 to Saturday 02:00
- Window replicas: 3

### Timeline Visualization

```
Day     Thu  Fri 21:00  Fri 22:00  Fri 23:00  Sat 00:00  Sat 01:00  Sat 02:00  Sat 03:00
        ────┼──────────┼──────────┼──────────┼──────────┼──────────┼──────────┼──────────
Replicas 0   0          3          3          3          3          0          0
                        ├──────────────────────────────────┤
Window                  │   Night Shift (Fri→Sat)          │
Events                  ScaleUp                             ScaleDown
Calendar                Friday                   Saturday
```

---

## Simulation Strategy

Since we cannot wait until Friday 22:00 Berlin time, we use time offset calculations to simulate the cross-midnight boundary.

### Offset Calculation Method

**Goal:** Create a window that will become active within 1 minute and cross midnight within the demo timeframe.

**Approach:**
We'll calculate times relative to the current hour to simulate the cross-midnight effect.

**Example Calculation:**
```
Current Berlin Time: Tuesday 14:37 CET

Simulation Strategy:
- Treat current hour as "22:00" (pre-midnight hour)
- Window start: current_hour:00 (e.g., 14:00)
- Window end: (current_hour+2):00 but specify as 02:00 pattern

To simulate crossing midnight:
- Window start: "14:37" (T+0, simulating 22:00)
- Window end: "16:37" (T+120min, simulating 02:00 next day)
```

**Important Note:** For true cross-midnight demonstration, we'll use a window that genuinely crosses the 00:00 boundary by running the demo in the evening.

---

## Two Demo Modes

### Mode 1: Simulated Midnight Cross (Daytime Demo)

**For demonstrations that must run during daytime hours.**

**Window Configuration:**
```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
  start: "14:38"  # Current hour + 1 minute
  end: "14:41"    # Current hour + 4 minutes
  replicas: 3
```

**Limitation:** Does not actually cross midnight, but demonstrates the window mechanics. Use SCENARIO-A for this case instead.

---

### Mode 2: Actual Midnight Cross (Evening Demo)

**For authentic cross-midnight demonstration.**

**Timing Requirement:** Start demo between 21:30 and 23:00 local time (Europe/Berlin).

**Window Configuration:**
```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
  start: "23:00"
  end: "01:00"    # Next day
  replicas: 3
```

**Timeline:**
- T+0 (23:00): Setup
- T+1 (23:01): Window opens, scale to 3
- T+61 (00:01): Crosses midnight while in window
- T+121 (01:01): Window closes, scale to 0

**This guide focuses on Mode 2 - Actual Midnight Cross.**

---

## Step-by-Step Operator Guide (Mode 2)

### Prerequisites for Evening Run

**Check Current Time:**
```bash
# Show current time in Europe/Berlin
TZ=Europe/Berlin date +"%Y-%m-%d %H:%M:%S %Z"
```

**Expected Output:**
```
2025-10-28 22:45:30 CET
```

**Requirement:** Current hour must be between 21 and 23 (inclusive).

If current time is outside this range, this demo cannot authentically demonstrate cross-midnight behavior. Use Mode 1 or wait for evening.

---

### T-1: Pre-Demo Setup (1 minute)

#### Step 0.1: Calculate Window Times

```bash
# Get current Berlin time
TZ=Europe/Berlin date +"%H:%M"

# Example output: 22:47

# Calculate windows:
# If current time is 22:47, use:
#   start: "22:48" (next minute, or round up)
#   end: "00:50" (2 hours and 2 minutes later, crossing midnight)
```

**Write down your times:**
```
Current Berlin Time: __:__
Window Start: __:__
Window End (next day): __:__
Midnight Offset (minutes from now): __
```

**Example:**
```
Current Berlin Time: 22:47
Window Start: 22:48
Window End (next day): 00:50
Midnight Offset: 73 minutes
```

#### Step 0.2: Prepare Demo Manifest

```yaml
# Save as /tmp/nightshift-scaler.yaml
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: nightshift-scaler
  namespace: demo
spec:
  targetRef:
    kind: Deployment
    name: nightshift-demo

  timezone: Europe/Berlin
  defaultReplicas: 0

  windows:
  # Cross-midnight window: Evening to Morning
  - days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
    start: "22:48"  # REPLACE with your calculated start
    end: "00:50"    # REPLACE with your calculated end (next day)
    replicas: 3
```

**Critical Verification:**
- `end` time has earlier hour value than `start` time (e.g., 00:XX < 22:XX)
- This signals to Kyklos that window crosses midnight

---

### T+0: Create Demo Environment (20 seconds)

#### Step 1.1: Create Namespace and Deployment

```bash
# T+0:00 - Create demo namespace
kubectl create namespace demo

# T+0:05 - Create target deployment
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nightshift-demo
  namespace: demo
  labels:
    app: nightshift-demo
spec:
  replicas: 0
  selector:
    matchLabels:
      app: nightshift-demo
  template:
    metadata:
      labels:
        app: nightshift-demo
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

#### Step 1.2: Verify Initial State

```bash
# Show Berlin time and resources
TZ=Europe/Berlin date && kubectl get deploy,pods -n demo
```

**Expected Output:**
```
Tue Oct 28 22:47:30 CET 2025

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nightshift-demo   0/0     0            0           5s

No pods found in demo namespace.
```

**CAPTURE POINT 1:** Screenshot showing Berlin time and 0/0 replicas.

---

### T+0:20: Apply TimeWindowScaler (15 seconds)

#### Step 2.1: Apply TWS with Cross-Midnight Window

```bash
# T+0:20 - Apply TimeWindowScaler
kubectl apply -f /tmp/nightshift-scaler.yaml
```

**Expected Output:**
```
timewindowscaler.kyklos.io/nightshift-scaler created
```

#### Step 2.2: Verify TWS Configuration

```bash
# T+0:25 - Check TWS and show Berlin time
TZ=Europe/Berlin date && kubectl get tws -n demo
```

**Expected Output:**
```
Tue Oct 28 22:47:35 CET 2025

NAME                 WINDOW     REPLICAS   TARGET            AGE
nightshift-scaler    OffHours   0          nightshift-demo   5s
```

#### Step 2.3: Examine Cross-Midnight Window in Spec

```bash
# T+0:30 - Verify cross-midnight window specification
kubectl get tws nightshift-scaler -n demo -o yaml | grep -A 10 windows:
```

**Expected Output:**
```yaml
windows:
- days:
  - Mon
  - Tue
  - Wed
  - Thu
  - Fri
  - Sat
  - Sun
  start: "22:48"
  end: "00:50"
  replicas: 3
```

**Key Observation:** `end: "00:50"` is less than `start: "22:48"` - this indicates cross-midnight to Kyklos.

**CAPTURE POINT 2:** Screenshot showing window specification with cross-midnight pattern.

---

### T+0:35 to T+0:55: Pre-Window Observation (20 seconds)

#### Step 3.1: Watch Resources with Berlin Time

```bash
# T+0:35 - Watch with timezone-aware timestamp
watch -n 2 'TZ=Europe/Berlin date +"%a %Y-%m-%d %H:%M:%S %Z" && echo && kubectl get tws,deploy,pods -n demo'
```

**Expected Display (before window start):**
```
Tue 2025-10-28 22:47:45 CET

NAME                                     WINDOW     REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift-scaler OffHours   0          nightshift-demo

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nightshift-demo   0/0     0            0           50s

No pods in demo namespace.
```

**Observations:**
- Berlin time updates every 2 seconds
- WINDOW shows "OffHours" (before 22:48)
- REPLICAS remains 0
- No pods present

---

### T+1:00 (22:48): Window Opens - Scale Up (30 seconds)

#### Step 4.1: Observe Scale-Up at Window Start

**Keep watching.** At 22:48 Berlin time:

**T+1:05 (22:48:05 approximately):**
```
Tue 2025-10-28 22:48:05 CET

NAME                                     WINDOW         REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift-scaler NightShift     3          nightshift-demo

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nightshift-demo   0/3     3            0           1m10s

NAME                                  READY   STATUS              RESTARTS   AGE
pod/nightshift-demo-7f8c9b5d-abc12    0/1     ContainerCreating   0          3s
pod/nightshift-demo-7f8c9b5d-def34    0/1     ContainerCreating   0          3s
pod/nightshift-demo-7f8c9b5d-ghi56    0/1     ContainerCreating   0          3s
```

**Key Changes:**
1. WINDOW changed to "NightShift"
2. REPLICAS changed to 3
3. Three pods appeared in ContainerCreating state

**T+1:20 (22:48:20 approximately):**
```
Tue 2025-10-28 22:48:20 CET

NAME                                     WINDOW         REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift-scaler NightShift     3          nightshift-demo

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nightshift-demo   3/3     3            3           1m25s

NAME                                  READY   STATUS    RESTARTS   AGE
pod/nightshift-demo-7f8c9b5d-abc12    1/1     Running   0          18s
pod/nightshift-demo-7f8c9b5d-def34    1/1     Running   0          18s
pod/nightshift-demo-7f8c9b5d-ghi56    1/1     Running   0          18s
```

**Final State:** All 3 pods Running, scale-up complete.

**CAPTURE POINT 3:** Screenshot showing NightShift window with 3/3 Running pods before midnight.

#### Step 4.2: Examine Controller Cross-Midnight Logic

Open second terminal:

```bash
# View controller logs for cross-midnight handling
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=30 | grep -A 5 nightshift
```

**Expected Log Pattern:**
```
INFO  Reconciling TimeWindowScaler  {"namespace": "demo", "name": "nightshift-scaler"}
INFO  Current time in Europe/Berlin: 2025-10-28T22:48:05+01:00
INFO  Evaluating cross-midnight window  {"start": "22:48", "end": "00:50", "currentDay": "Tuesday"}
INFO  Cross-midnight calculation: window extends into next day
INFO  Matched window: NightShift (22:48 Tue → 00:50 Wed) -> 3 replicas
INFO  Target deployment has 0 replicas, desired 3 replicas
INFO  Scaling deployment nightshift-demo from 0 to 3 replicas
INFO  Requeue scheduled at window end: 2025-10-29T00:50:00+01:00
```

**Key Observations:**
- Controller explicitly identifies cross-midnight window
- Shows calculation extending into next calendar day
- Window notation shows day transition (Tue → Wed)
- Requeue scheduled for 00:50 tomorrow

**CAPTURE POINT 4:** Controller logs showing cross-midnight detection and boundary calculation.

---

### T+13:00 (23:00 to 23:59): Pre-Midnight Steady State (60 minutes)

**Note:** This section represents the hour before midnight. Watch continues to run.

#### Step 5.1: Monitor Stability Before Midnight

**T+15:00 (23:00):**
```
Tue 2025-10-28 23:00:00 CET

NAME                                     WINDOW         REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift-scaler NightShift     3          nightshift-demo

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nightshift-demo   3/3     3            3           13m

NAME                                  READY   STATUS    RESTARTS   AGE
pod/nightshift-demo-7f8c9b5d-abc12    1/1     Running   0          12m
pod/nightshift-demo-7f8c9b5d-def34    1/1     Running   0          12m
pod/nightshift-demo-7f8c9b5d-ghi56    1/1     Running   0          12m
```

**Observations:**
- Window remains "NightShift"
- Replicas remain 3
- Still Tuesday (before midnight)
- Pods stable with no restarts

#### Step 5.2: Verify Cross-Midnight Window Status

```bash
# Check status showing active cross-midnight window
kubectl get tws nightshift-scaler -n demo -o yaml | grep -A 15 status:
```

**Expected Status:**
```yaml
status:
  currentWindow: NightShift
  effectiveReplicas: 3
  lastScaleTime: "2025-10-28T22:48:05+01:00"
  targetObservedReplicas: 3
  observedGeneration: 1
  windowMetadata:
    crossesMidnight: true
    windowStart: "22:48"
    windowEnd: "00:50"
    windowEndDay: "Wednesday"
  conditions:
  - type: Ready
    status: "True"
    reason: Reconciled
    message: Target deployment matches desired replicas
```

**Key Status Fields:**
- `windowMetadata.crossesMidnight: true` - Explicitly marked
- `windowEndDay: "Wednesday"` - Shows day transition
- Ready condition True

**CAPTURE POINT 5:** TWS status showing cross-midnight metadata before midnight.

---

### T+73:00 (00:00 to 00:49): Post-Midnight Active Window (50 minutes)

**THIS IS THE CRITICAL DEMONSTRATION PERIOD**

#### Step 6.1: Observe Midnight Transition

**T+73:00 (00:00:00 approximately):**
```
Wed 2025-10-29 00:00:00 CET

NAME                                     WINDOW         REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift-scaler NightShift     3          nightshift-demo

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nightshift-demo   3/3     3            3           73m

NAME                                  READY   STATUS    RESTARTS   AGE
pod/nightshift-demo-7f8c9b5d-abc12    1/1     Running   0          72m
pod/nightshift-demo-7f8c9b5d-def34    1/1     Running   0          72m
pod/nightshift-demo-7f8c9b5d-ghi56    1/1     Running   0          72m
```

**CRITICAL OBSERVATIONS:**
1. **Date changed from Tuesday to Wednesday** - Calendar day rolled over
2. **WINDOW still shows "NightShift"** - Window remains active despite day change
3. **REPLICAS still 3** - No scale change at midnight
4. **Pods still Running** - No disruption at day boundary

**This proves cross-midnight handling works correctly!**

**CAPTURE POINT 6:** Screenshot showing Wednesday date with window still active.

#### Step 6.2: Verify Controller Handled Midnight Correctly

```bash
# Check controller logs around midnight
kubectl logs -n kyklos-system -l app=kyklos-controller --since=5m | grep -A 3 "00:00"
```

**Expected Logs (or absence of logs):**
```
INFO  Reconciling TimeWindowScaler  {"namespace": "demo", "name": "nightshift-scaler"}
INFO  Current time in Europe/Berlin: 2025-10-29T00:00:30+01:00
INFO  Cross-midnight window still active  {"end": "00:50", "minutesRemaining": 49}
INFO  Target deployment matches desired state, no action needed
INFO  Requeue scheduled at window end: 2025-10-29T00:50:00+01:00
```

**Key Observations:**
- Controller reconciled around midnight (may be triggered by watch or scheduled requeue)
- Recognized window is still active despite day change
- Correctly calculated remaining window time
- No scaling action taken (correct behavior)

**CAPTURE POINT 7:** Controller logs confirming midnight crossed without scaling action.

#### Step 6.3: Examine Cross-Day Window Calculation

```bash
# Get detailed TWS status post-midnight
kubectl get tws nightshift-scaler -n demo -o jsonpath='{.status.windowMetadata}' | jq
```

**Expected Output:**
```json
{
  "crossesMidnight": true,
  "currentDay": "Wednesday",
  "windowStart": "22:48",
  "windowStartDay": "Tuesday",
  "windowEnd": "00:50",
  "windowEndDay": "Wednesday",
  "activeFor": "1h 12m",
  "remainingTime": "49m"
}
```

**Status Interpretation:**
- Window started on Tuesday at 22:48
- Will end on Wednesday at 00:50
- Currently Wednesday, so we're in the "next day" portion
- Remaining time calculated correctly

**CAPTURE POINT 8:** Window metadata showing cross-day state.

---

### T+123:00 (00:50): Window Closes - Scale Down (15 seconds)

#### Step 7.1: Observe Scale-Down at Window End

**Keep watching.** At 00:50 Berlin time:

**T+123:05 (00:50:05 approximately):**
```
Wed 2025-10-29 00:50:05 CET

NAME                                     WINDOW     REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift-scaler OffHours   0          nightshift-demo

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nightshift-demo   3/3     0            3           123m

NAME                                  READY   STATUS        RESTARTS   AGE
pod/nightshift-demo-7f8c9b5d-abc12    1/1     Terminating   0          122m
pod/nightshift-demo-7f8c9b5d-def34    1/1     Terminating   0          122m
pod/nightshift-demo-7f8c9b5d-ghi56    1/1     Terminating   0          122m
```

**Key Changes:**
1. WINDOW changed to "OffHours"
2. REPLICAS changed to 0
3. Pods entered Terminating state
4. Day is Wednesday (window end day)

**T+123:15 (00:50:15 approximately):**
```
Wed 2025-10-29 00:50:15 CET

NAME                                     WINDOW     REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift-scaler OffHours   0          nightshift-demo

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nightshift-demo   0/0     0            0           123m

No pods in demo namespace.
```

**Final State:** Scale-down complete, back to 0 replicas on Wednesday morning.

**CAPTURE POINT 9:** Screenshot showing OffHours state after cross-midnight window closed.

#### Step 7.2: Examine Controller Window Exit Logs

```bash
# View scale-down decision logs
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=20 | grep nightshift
```

**Expected Log Pattern:**
```
INFO  Reconciling TimeWindowScaler  {"namespace": "demo", "name": "nightshift-scaler"}
INFO  Current time in Europe/Berlin: 2025-10-29T00:50:05+01:00
INFO  Cross-midnight window ended  {"start": "22:48 Tue", "end": "00:50 Wed", "duration": "2h 2m"}
INFO  No matching windows, using defaultReplicas: 0
INFO  Target deployment has 3 replicas, desired 0 replicas
INFO  Scaling deployment nightshift-demo from 3 to 0 replicas
INFO  Requeue scheduled at next window start: 2025-10-29T22:48:00+01:00
```

**Key Observations:**
- Controller identified window end on Wednesday
- Calculated total window duration correctly (crossed midnight)
- Scheduled next occurrence for tonight (Wednesday 22:48)

**CAPTURE POINT 10:** Controller logs showing cross-midnight window end calculation.

---

### T+125:00: Post-Demo Verification (5 minutes)

#### Step 8.1: Review Complete Event Timeline

```bash
# View all events to see complete cross-midnight lifecycle
kubectl get events -n demo --sort-by='.lastTimestamp'
```

**Expected Event Timeline:**
```
LAST SEEN   TYPE     REASON              OBJECT                         MESSAGE
2h3m        Normal   WindowTransition    timewindowscaler/nightshift    Entered cross-midnight window: NightShift (22:48 Tue → 00:50 Wed)
2h3m        Normal   ScalingTarget       timewindowscaler/nightshift    Scaling nightshift-demo from 0 to 3 replicas
2h3m        Normal   ScaledUp            deployment/nightshift-demo     Scaled up to 3
2m          Normal   MidnightCrossed     timewindowscaler/nightshift    Window crossed midnight boundary, remains active
2m          Normal   WindowTransition    timewindowscaler/nightshift    Exited window: NightShift (00:50 Wed)
2m          Normal   ScalingTarget       timewindowscaler/nightshift    Scaling nightshift-demo from 3 to 0 replicas
2m          Normal   ScaledDown          deployment/nightshift-demo     Scaled down to 0
```

**Key Event:** `MidnightCrossed` event shows Kyklos explicitly acknowledged the day boundary crossing.

**CAPTURE POINT 11:** Complete event timeline showing cross-midnight lifecycle.

---

### T+130:00: Cleanup (15 seconds)

```bash
# Stop watching
# Press Ctrl+C

# Clean up demo resources
kubectl delete namespace demo
```

**Expected Output:**
```
namespace "demo" deleted
```

**Demo Complete!**

---

## Success Criteria

The demo successfully demonstrates cross-midnight handling if ALL criteria are met:

### Pre-Midnight Phase (22:48 to 23:59)
- [ ] Window opened at calculated start time (22:48)
- [ ] WINDOW changed to "NightShift"
- [ ] Scaled from 0 to 3 replicas
- [ ] Controller logs show cross-midnight window detection
- [ ] Status shows `windowMetadata.crossesMidnight: true`
- [ ] Pods remained stable throughout pre-midnight period
- [ ] Date shown as Tuesday in all outputs

### Midnight Transition (00:00)
- [ ] **Date changed to Wednesday in watch output**
- [ ] **WINDOW remained "NightShift" (did not reset)**
- [ ] **REPLICAS remained 3 (no scale change)**
- [ ] **Pods continued Running without disruption**
- [ ] Controller logs show "window still active" or equivalent
- [ ] No scaling events generated at midnight
- [ ] Status shows correct cross-day metadata

### Post-Midnight Phase (00:01 to 00:49)
- [ ] Window remained active throughout
- [ ] Replicas remained 3
- [ ] Status shows Wednesday as current day
- [ ] Status shows Tuesday as window start day
- [ ] Remaining time calculated correctly

### Window End (00:50)
- [ ] Window closed at calculated end time (00:50 Wednesday)
- [ ] WINDOW changed to "OffHours"
- [ ] Scaled from 3 to 0 replicas
- [ ] Controller logs show "cross-midnight window ended"
- [ ] Total duration calculated correctly (~2 hours)
- [ ] Next window scheduled for tonight (Wednesday 22:48)

### System Health
- [ ] No controller restarts during entire demo
- [ ] All conditions remained healthy
- [ ] Events show clear cross-midnight lifecycle
- [ ] No timezone errors or warnings in logs

---

## Key Demonstration Points

### What This Proves

**1. Calendar Day Independence**
- Window activation is based on time-of-day, not calendar day
- Crossing midnight does not terminate active window
- Day-of-week specification applies to window start, not the entire duration

**2. Correct Boundary Calculation**
- Start boundary computed on specified day (Tuesday 22:48)
- End boundary computed on next calendar day (Wednesday 00:50)
- Duration calculated correctly across day boundary

**3. DST Awareness**
- Using Europe/Berlin timezone (CET/CEST)
- Timezone offset shown correctly in logs (+01:00 or +02:00)
- Controller uses IANA timezone rules for calculations

**4. Requeue Scheduling**
- Before midnight: Next requeue scheduled for 00:50 Wednesday
- After midnight: Next requeue still correct for 00:50 Wednesday
- After window end: Next requeue scheduled for 22:48 Wednesday (same day)

---

## Recovery Procedures

### Issue: Window closed at midnight

**Symptoms:**
- At 00:00, WINDOW changed to "OffHours"
- Replicas scaled down to 0
- Window did not extend past midnight

**Diagnosis:**
```bash
# Check window specification
kubectl get tws nightshift-scaler -n demo -o yaml | grep -A 5 windows:
```

**Possible Cause:** Window end time is NOT less than start time.

**Example of INCORRECT specification:**
```yaml
windows:
- start: "22:00"
  end: "23:59"  # WRONG: end > start, does not cross midnight
```

**Fix:**
```yaml
windows:
- start: "22:00"
  end: "02:00"  # CORRECT: end < start, crosses midnight
```

**Recovery:**
```bash
# Delete and recreate TWS with correct times
kubectl delete tws nightshift-scaler -n demo
kubectl apply -f /tmp/nightshift-scaler.yaml  # with corrected times
```

---

### Issue: Timezone errors in controller logs

**Symptoms:**
- Logs show "invalid timezone" or "timezone not found"
- Window never activates
- Status shows Degraded=True

**Diagnosis:**
```bash
# Check controller logs for timezone errors
kubectl logs -n kyklos-system -l app=kyklos-controller | grep -i timezone
```

**Expected Error:**
```
ERROR Failed to load timezone  {"timezone": "Europe/Berlin", "error": "unknown time zone"}
```

**Possible Causes:**
1. Controller container missing IANA timezone database
2. Typo in timezone name
3. Outdated timezone data

**Fix:**
```bash
# Fix 1: Verify timezone name (case-sensitive)
# Correct: Europe/Berlin
# Wrong: europe/berlin, Europe/berlin

# Fix 2: Check if timezone database is available in controller
kubectl exec -n kyklos-system deploy/kyklos-controller-manager -- ls -la /usr/share/zoneinfo/Europe/Berlin

# Fix 3: Rebuild controller with timezone data
# Add to Dockerfile:
# COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
```

---

### Issue: Demo ran too long (> 2 hours)

**Symptoms:**
- Window opened correctly but demo exceeded planned time
- Need to complete demo faster

**Solution:** Use shorter window duration.

**Modified Window (30-minute cross-midnight):**
```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
  start: "23:45"
  end: "00:15"    # Only 30 minutes
  replicas: 3
```

**Timeline:**
- T+0 (23:45): Setup
- T+1 (23:46): Window opens
- T+16 (00:01): Cross midnight while active
- T+31 (00:16): Window closes
- Total: 31 minutes

---

## Alternative: Rapid Cross-Midnight Test

For testing cross-midnight logic without waiting for evening:

### Mock Midnight Approach

**Strategy:** Use current minute as "midnight" and verify boundary logic.

**Window Configuration:**
```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
  start: "14:57"  # 3 minutes before "mock midnight" (15:00)
  end: "15:02"    # 2 minutes after "mock midnight"
  replicas: 3
```

**Timeline:**
- 14:57: Window opens (simulate 23:00)
- 14:58-14:59: Pre-"midnight" (simulate 23:01-23:59)
- 15:00: "Midnight" transition (simulate 00:00)
- 15:01: Post-"midnight" (simulate 00:01)
- 15:02: Window closes (simulate 00:02)

**Limitation:** This does NOT test actual calendar day transition, only the time arithmetic. Use Mode 2 for authentic cross-midnight demonstration.

---

## Handoff to Docs Writer

### Materials to Provide

**1. Screenshots (11 total):**
- CAPTURE POINT 1: Berlin time and initial 0/0 state
- CAPTURE POINT 2: Window spec showing cross-midnight (end < start)
- CAPTURE POINT 3: NightShift active on Tuesday with 3/3 pods
- CAPTURE POINT 4: Controller logs detecting cross-midnight window
- CAPTURE POINT 5: Status showing cross-midnight metadata (Tuesday)
- CAPTURE POINT 6: **KEY SHOT** - Wednesday date with window still active
- CAPTURE POINT 7: Controller logs confirming midnight crossed
- CAPTURE POINT 8: Window metadata showing cross-day state
- CAPTURE POINT 9: OffHours state on Wednesday after window end
- CAPTURE POINT 10: Controller logs calculating window end
- CAPTURE POINT 11: Complete event timeline with MidnightCrossed event

**2. Terminal Recordings:**
- Watch output from 22:47 to 00:52 (full midnight transition)
- Controller logs filtered for nightshift with timestamps

**3. Key Demonstration Materials:**
- Side-by-side comparison: Tuesday 23:59 vs Wednesday 00:01 (same window)
- Window metadata JSON before and after midnight
- Event timeline showing no scaling at midnight

**4. Concept Diagrams:**
- Timeline showing window spanning two calendar days
- Boundary calculation flowchart for cross-midnight logic

### Best Screenshots for Documentation

**Primary Hero Shot:** CAPTURE POINT 6
- Shows Wednesday date with window still active
- Proves midnight crossing without disruption
- Most compelling evidence of cross-midnight handling

**Supporting Screenshots:**
1. CAPTURE POINT 2: Window spec (explains the configuration)
2. CAPTURE POINT 6: Midnight transition (proves the behavior)
3. CAPTURE POINT 8: Metadata (shows internal state)

**For Technical Documentation:**
- CAPTURE POINT 4 + 7 + 10: Controller decision log sequence
- CAPTURE POINT 11: Event timeline for troubleshooting guide

---

## Comparison with Scenario A

| Aspect | Scenario A (Minute Demo) | Scenario B (Cross-Midnight) |
|--------|--------------------------|----------------------------|
| **Duration** | 10 minutes | 2+ hours (can be shortened) |
| **Timezone** | UTC (no DST) | Europe/Berlin (DST-aware) |
| **Window Pattern** | Within same hour | Crosses midnight boundary |
| **Scale Pattern** | 0→2→0 | 0→3→0 |
| **Key Demonstration** | Basic time windows | Cross-midnight logic |
| **Complexity** | Low | Medium |
| **Run Anytime** | Yes | Evening only (authentic) |
| **Primary Audience** | First-time users | Advanced users |

---

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2025-10-28 | 1.0 | Initial cross-midnight scenario design |
