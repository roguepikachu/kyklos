# Sample Validation Scenarios

**Purpose:** Three complete end-to-end scenarios for validating Kyklos implementation.

**Last Updated:** 2025-10-29

## Overview

These scenarios validate core functionality with realistic use cases. Each scenario includes setup, execution steps, expected outcomes, and validation commands.

---

## Scenario A: Office Hours Scaling

**Duration:** 15 minutes (with time-warp)
**Complexity:** Simple
**Validates:** Basic window matching, scale up/down, status updates

### Business Case

A web application needs 10 replicas during business hours (09:00-17:00 EST) for user traffic, but only 2 replicas outside business hours for background tasks.

### Prerequisites

```bash
# 1. Kubernetes cluster (kind/minikube) running
# 2. Kyklos controller installed
# 3. kubectl configured

# Create test namespace
kubectl create namespace demo-office-hours
```

### Setup

**Step 1: Deploy Target Workload**

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
  namespace: demo-office-hours
spec:
  replicas: 2
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: nginx
        image: nginx:1.25
        resources:
          requests:
            cpu: 100m
            memory: 64Mi
          limits:
            cpu: 200m
            memory: 128Mi
```

```bash
kubectl apply -f deployment.yaml
kubectl wait --for=condition=available --timeout=60s deployment/webapp -n demo-office-hours
```

**Step 2: Create TimeWindowScaler**

```yaml
# timewindowscaler.yaml
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: webapp-office-hours
  namespace: demo-office-hours
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: webapp
  timezone: "America/New_York"
  defaultReplicas: 2
  windows:
  - days: ["Mon", "Tue", "Wed", "Thu", "Fri"]
    start: "09:00"
    end: "17:00"
    replicas: 10
```

```bash
kubectl apply -f timewindowscaler.yaml
```

### Execution Steps

**Test 1: Outside Business Hours (e.g., 20:00 EST)**

Simulate current time: Monday, 2025-01-27, 20:00:00 EST

```bash
# Expected: Deployment scales to 2 replicas (defaultReplicas)

# Wait for reconciliation (30-60 seconds)
sleep 60

# Verify replica count
kubectl get deployment webapp -n demo-office-hours -o jsonpath='{.spec.replicas}'
# Expected output: 2

# Verify TWS status
kubectl get tws webapp-office-hours -n demo-office-hours -o yaml
```

**Expected Status:**

```yaml
status:
  currentWindow: "OffHours"
  effectiveReplicas: 2
  targetObservedReplicas: 2
  observedGeneration: 1
  conditions:
  - type: Ready
    status: "True"
    reason: Reconciled
    message: "Target replicas match desired state (2)"
  - type: Reconciling
    status: "False"
    reason: Stable
    message: "Waiting until 09:00 EST"
  - type: Degraded
    status: "False"
    reason: OperationalNormal
    message: "No issues detected"
```

**Test 2: Enter Business Hours (simulate 09:00 EST)**

If using real time, wait until 09:00 EST. For testing, use time-warp (fast minutes):

```bash
# Time-warp: Update TWS to use minute-scale windows for testing
# Replace window: 09:00-17:00 with 00:09-00:17 (9th to 17th minute of hour)

kubectl patch tws webapp-office-hours -n demo-office-hours --type=merge -p '
spec:
  windows:
  - days: ["Mon", "Tue", "Wed", "Thu", "Fri"]
    start: "00:09"
    end: "00:17"
    replicas: 10
'

# Wait until 9th minute of current hour
# Example: If it's 14:05, wait until 14:09
```

**Expected Behavior:**

```bash
# At 14:09 (window starts)
# Controller reconciles and scales Deployment to 10 replicas

# Verify scaling
kubectl get deployment webapp -n demo-office-hours -o jsonpath='{.spec.replicas}'
# Expected output: 10

# Verify events
kubectl get events -n demo-office-hours --field-selector involvedObject.name=webapp-office-hours --sort-by=.lastTimestamp
# Expected: "ScaledUp" event: "Scaled up from 2 to 10 replicas (window: BusinessHours)"
```

**Test 3: Exit Business Hours (simulate 17:00 EST)**

```bash
# Wait until 17th minute (window ends)
# Example: If it's 14:09, wait until 14:17

# Expected: Deployment scales down to 2 replicas

kubectl get deployment webapp -n demo-office-hours -o jsonpath='{.spec.replicas}'
# Expected output: 2

# Verify scale-down event
kubectl get events -n demo-office-hours --field-selector involvedObject.name=webapp-office-hours --sort-by=.lastTimestamp
# Expected: "ScaledDown" event: "Scaled down from 10 to 2 replicas (window: OffHours)"
```

### Validation Checklist

- [ ] Deployment starts with 2 replicas
- [ ] At window start (09:00), scales to 10 replicas within 60 seconds
- [ ] Status.effectiveReplicas shows 10 during window
- [ ] Status.currentWindow shows "BusinessHours" during window
- [ ] ScaledUp event emitted at window entry
- [ ] At window end (17:00), scales to 2 replicas within 60 seconds
- [ ] Status.effectiveReplicas shows 2 outside window
- [ ] Status.currentWindow shows "OffHours" outside window
- [ ] ScaledDown event emitted at window exit
- [ ] Ready condition stays True throughout

### Metrics Validation

```bash
# Port-forward to controller metrics endpoint
kubectl port-forward -n kyklos-system deployment/kyklos-controller 8080:8080 &

# Check metrics
curl -s http://localhost:8080/metrics | grep kyklos

# Expected metrics:
# kyklos_scale_events_total{tws_name="webapp-office-hours",namespace="demo-office-hours",direction="up"} 1
# kyklos_scale_events_total{tws_name="webapp-office-hours",namespace="demo-office-hours",direction="down"} 1
# kyklos_effective_replicas{tws_name="webapp-office-hours",namespace="demo-office-hours",window="OffHours"} 2
```

### Cleanup

```bash
kubectl delete namespace demo-office-hours
```

---

## Scenario B: Cross-Midnight Window

**Duration:** 20 minutes (with time-warp)
**Complexity:** Moderate
**Validates:** Cross-midnight logic, day boundary handling, weekend windows

### Business Case

A batch processing system runs intensive jobs overnight (22:00-02:00) on weeknights, requiring 8 replicas. During the day, only 2 replicas are needed for monitoring.

### Prerequisites

```bash
kubectl create namespace demo-cross-midnight
```

### Setup

**Step 1: Deploy Target Workload**

```yaml
# batch-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: batch-processor
  namespace: demo-cross-midnight
spec:
  replicas: 2
  selector:
    matchLabels:
      app: batch-processor
  template:
    metadata:
      labels:
        app: batch-processor
    spec:
      containers:
      - name: processor
        image: busybox:1.36
        command: ["sh", "-c", "while true; do echo Processing...; sleep 30; done"]
        resources:
          requests:
            cpu: 100m
            memory: 64Mi
```

```bash
kubectl apply -f batch-deployment.yaml
kubectl wait --for=condition=available --timeout=60s deployment/batch-processor -n demo-cross-midnight
```

**Step 2: Create TimeWindowScaler with Cross-Midnight Window**

```yaml
# cross-midnight-tws.yaml
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: batch-processor-night
  namespace: demo-cross-midnight
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: batch-processor
  timezone: "UTC"
  defaultReplicas: 2
  windows:
  - days: ["Mon", "Tue", "Wed", "Thu", "Fri"]
    start: "22:00"
    end: "02:00"
    replicas: 8
```

```bash
kubectl apply -f cross-midnight-tws.yaml
```

### Execution Steps

**Test 1: Before Window Start (e.g., Friday 21:00 UTC)**

```bash
# Expected: Outside window, 2 replicas

kubectl get deployment batch-processor -n demo-cross-midnight -o jsonpath='{.spec.replicas}'
# Expected: 2

kubectl get tws batch-processor-night -n demo-cross-midnight -o jsonpath='{.status.currentWindow}'
# Expected: "OffHours"
```

**Test 2: Window Start (Friday 22:00 UTC)**

For time-warp testing, patch window to minute-scale:

```yaml
# Patch to use 00:22-00:02 (22nd minute to 2nd minute)
windows:
- days: ["Mon", "Tue", "Wed", "Thu", "Fri"]
  start: "00:22"
  end: "00:02"
  replicas: 8
```

```bash
# Wait until 22nd minute of current hour
# Example: If it's 15:20, wait until 15:22

# At 15:22 (window starts)
kubectl get deployment batch-processor -n demo-cross-midnight -o jsonpath='{.spec.replicas}'
# Expected: 8

kubectl get tws batch-processor-night -n demo-cross-midnight -o jsonpath='{.status.currentWindow}'
# Expected: "Night" or "Custom-..."
```

**Test 3: After Midnight (Saturday 01:00 UTC)**

```bash
# Still within window (crosses midnight)
# Wait until 1st minute of next hour (e.g., 16:01)

kubectl get deployment batch-processor -n demo-cross-midnight -o jsonpath='{.spec.replicas}'
# Expected: 8 (still in window)

# Verify status shows window still active
kubectl get tws batch-processor-night -n demo-cross-midnight -o yaml | grep currentWindow
# Expected: currentWindow shows active window
```

**Test 4: Window End (Saturday 02:00 UTC)**

```bash
# Wait until 2nd minute (e.g., 16:02)

kubectl get deployment batch-processor -n demo-cross-midnight -o jsonpath='{.spec.replicas}'
# Expected: 2 (scaled down)

# Verify events
kubectl get events -n demo-cross-midnight --field-selector involvedObject.name=batch-processor-night --sort-by=.lastTimestamp
# Expected: ScaledDown event at window exit
```

**Test 5: Weekend (Saturday 22:00 UTC)**

```bash
# Saturday is not in window days [Mon-Fri]
# Should NOT scale up at 22:00 on Saturday

# Wait until Saturday 22nd minute
# (In time-warp: wait for next occurrence of 22nd minute on a different day simulation)

kubectl get deployment batch-processor -n demo-cross-midnight -o jsonpath='{.spec.replicas}'
# Expected: 2 (no window on Saturday)
```

### Validation Checklist

- [ ] Friday 22:00: Scales to 8 replicas
- [ ] Friday 23:00: Still 8 replicas (in window)
- [ ] Saturday 00:00: Still 8 replicas (window extends to Saturday morning)
- [ ] Saturday 01:00: Still 8 replicas (before 02:00)
- [ ] Saturday 02:00: Scales to 2 replicas (window ends)
- [ ] Saturday 22:00: Stays at 2 replicas (not in window days)
- [ ] Monday 22:00: Scales to 8 replicas (window starts again)

### Edge Case Tests

**Day Boundary Check:**

```bash
# Verify TWS correctly identifies that Saturday 01:00 matches Friday's window

kubectl get tws batch-processor-night -n demo-cross-midnight -o yaml
# Check status.currentWindow at Saturday 01:00
# Should show active window (not "OffHours")
```

**Weekend Exclusion:**

```bash
# Verify Saturday and Sunday nights do NOT trigger the window

# Simulate Saturday 22:00
# Expected: replicas=2, currentWindow="OffHours"
```

### Cleanup

```bash
kubectl delete namespace demo-cross-midnight
```

---

## Scenario C: Pause and Manual Drift Correction

**Duration:** 10 minutes
**Complexity:** Simple
**Validates:** Pause functionality, manual drift detection and correction

### Business Case

An SRE needs to temporarily prevent Kyklos from scaling a deployment during a maintenance window, while still allowing the controller to compute and report desired state.

### Prerequisites

```bash
kubectl create namespace demo-pause
```

### Setup

**Step 1: Deploy Target Workload**

```yaml
# api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-service
  namespace: demo-pause
spec:
  replicas: 10
  selector:
    matchLabels:
      app: api-service
  template:
    metadata:
      labels:
        app: api-service
    spec:
      containers:
      - name: api
        image: nginx:1.25
        resources:
          requests:
            cpu: 100m
            memory: 64Mi
```

```bash
kubectl apply -f api-deployment.yaml
kubectl wait --for=condition=available --timeout=60s deployment/api-service -n demo-pause
```

**Step 2: Create TimeWindowScaler (Active Window)**

```yaml
# api-tws.yaml
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: api-service-scaler
  namespace: demo-pause
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-service
  timezone: "UTC"
  defaultReplicas: 3
  windows:
  - days: ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]
    start: "00:00"
    end: "23:59"
    replicas: 10
  pause: false
```

```bash
kubectl apply -f api-tws.yaml

# Wait for reconciliation
sleep 30
```

### Execution Steps

**Test 1: Verify Normal Operation**

```bash
# Check that TWS is managing the deployment
kubectl get deployment api-service -n demo-pause -o jsonpath='{.spec.replicas}'
# Expected: 10 (within window)

kubectl get tws api-service-scaler -n demo-pause -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
# Expected: "True"
```

**Test 2: Enable Pause**

```bash
# SRE enables pause before maintenance
kubectl patch tws api-service-scaler -n demo-pause --type=merge -p '
spec:
  pause: true
'

# Wait for reconciliation
sleep 30

# Verify status updated
kubectl get tws api-service-scaler -n demo-pause -o yaml | grep pause
# Expected: spec.pause: true
```

**Test 3: Manual Scale While Paused**

```bash
# SRE manually scales deployment during maintenance
kubectl scale deployment api-service -n demo-pause --replicas=5

# Wait for scale to complete
sleep 10

# Verify deployment scaled
kubectl get deployment api-service -n demo-pause -o jsonpath='{.spec.replicas}'
# Expected: 5

# Check TWS status
kubectl get tws api-service-scaler -n demo-pause -o yaml
```

**Expected Status:**

```yaml
status:
  effectiveReplicas: 10  # TWS still computes desired state
  targetObservedReplicas: 5  # But observes actual drift
  currentWindow: "BusinessHours"
  conditions:
  - type: Ready
    status: "False"  # Drift detected
    reason: TargetMismatch
    message: "Target has 5 replicas but desired is 10 (pause=true)"
  - type: Reconciling
    status: "False"
    reason: Stable
    message: "Paused"
  - type: Degraded
    status: "False"
    reason: OperationalNormal
```

**Test 4: Verify No Auto-Correction While Paused**

```bash
# Wait 2 minutes to ensure controller doesn't correct drift
sleep 120

# Verify deployment still at 5 replicas (no correction)
kubectl get deployment api-service -n demo-pause -o jsonpath='{.spec.replicas}'
# Expected: 5

# Verify ScalingSkipped event
kubectl get events -n demo-pause --field-selector involvedObject.name=api-service-scaler --sort-by=.lastTimestamp
# Expected: "ScalingSkipped" event: "Scaling skipped due to pause: current=5, desired=10"
```

**Test 5: Disable Pause and Verify Auto-Correction**

```bash
# SRE completes maintenance and disables pause
kubectl patch tws api-service-scaler -n demo-pause --type=merge -p '
spec:
  pause: false
'

# Wait for reconciliation (should correct drift)
sleep 60

# Verify deployment corrected to 10 replicas
kubectl get deployment api-service -n demo-pause -o jsonpath='{.spec.replicas}'
# Expected: 10

# Verify Ready condition restored
kubectl get tws api-service-scaler -n demo-pause -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'
# Expected: "True"

# Verify correction event
kubectl get events -n demo-pause --field-selector involvedObject.name=api-service-scaler --sort-by=.lastTimestamp
# Expected: "ScaledUp" event: "Corrected manual drift: scaled from 5 to 10 replicas"
```

**Test 6: Manual Scale-Up Without Pause**

```bash
# Test drift correction in non-paused mode
kubectl scale deployment api-service -n demo-pause --replicas=15

# Wait for controller to detect and correct drift
sleep 60

# Verify corrected back to 10
kubectl get deployment api-service -n demo-pause -o jsonpath='{.spec.replicas}'
# Expected: 10 (corrected immediately)

# Verify event
kubectl get events -n demo-pause --sort-by=.lastTimestamp | tail -n 5
# Expected: ScaledDown event showing correction from 15 to 10
```

### Validation Checklist

- [ ] Normal operation: TWS manages deployment at 10 replicas
- [ ] Pause enabled: spec.pause=true
- [ ] Manual scale to 5: deployment scaled successfully
- [ ] While paused: effectiveReplicas=10, targetObservedReplicas=5, Ready=False
- [ ] While paused: No auto-correction after 2 minutes
- [ ] ScalingSkipped event emitted while paused
- [ ] Pause disabled: drift corrected within 60 seconds
- [ ] Post-unpause: deployment back to 10 replicas, Ready=True
- [ ] Manual scale-up without pause: corrected immediately

### Status Transition Validation

```bash
# Capture status at each stage
for stage in normal paused drift-detected unpause-corrected; do
  echo "=== Stage: $stage ==="
  kubectl get tws api-service-scaler -n demo-pause -o jsonpath='{.status}' | jq
  sleep 30
done
```

### Cleanup

```bash
kubectl delete namespace demo-pause
```

---

## Automated Test Script

**`/test/e2e/run-scenarios.sh`:**

```bash
#!/bin/bash
set -e

SCENARIOS=("office-hours" "cross-midnight" "pause")

echo "Running Kyklos E2E Scenarios"
echo "============================"

for scenario in "${SCENARIOS[@]}"; do
  echo ""
  echo "Running Scenario: $scenario"
  echo "----------------------------"

  # Run scenario-specific script
  ./test/e2e/scenario-${scenario}.sh

  if [ $? -eq 0 ]; then
    echo "✓ Scenario $scenario PASSED"
  else
    echo "✗ Scenario $scenario FAILED"
    exit 1
  fi
done

echo ""
echo "All scenarios PASSED ✓"
```

**Usage:**

```bash
# From repository root
./test/e2e/run-scenarios.sh
```

---

## Summary

These three scenarios validate:

1. **Office Hours (A):** Basic window matching, scale up/down, events, status
2. **Cross-Midnight (B):** Cross-midnight logic, day boundaries, weekend exclusion
3. **Pause (C):** Pause behavior, manual drift detection, auto-correction

All scenarios can be executed in under 30 minutes total using time-warp testing (minute-scale windows).
