# Kyklos Sanity Tests

Quick tests to verify Kyklos is functioning correctly in a live cluster.

## Tests Available

### 1. Smoke Test (30 seconds) ⚡
**Fastest way to verify basic functionality**

```bash
./test/sanity/smoke-test.sh
# OR
make smoke-test
```

**What it does:**
- Creates a deployment and TimeWindowScaler with an active window
- Verifies the deployment scales to the expected replica count
- Checks that status conditions are set correctly
- Validates scaling events are emitted
- Auto-cleanup after completion

**Pass Criteria:**
- ✓ Deployment scales to expected replicas (5)
- ✓ TWS status shows Ready=True
- ✓ ScaledUp event is emitted

**When to use:**
- Quick smoke test after deployment
- CI/CD pipeline validation
- Before making changes to verify baseline

---

### 2. Sanity Test (3 minutes) 🔍
**Comprehensive test with time-based window transitions**

```bash
./test/sanity/run-sanity-test.sh
# OR
make sanity-test
```

**What it does:**
- Creates deployment starting with 1 replica
- Sets up TWS with minute-scale windows (transitions in ~1-3 minutes)
- Monitors scaling through window transitions:
  - Default (1 replica)
  - Window 1 (3 replicas)
  - Window 2 (5 replicas)
  - Back to default (1 replica)
- Shows real-time scaling behavior
- Optional cleanup at the end

**Pass Criteria:**
- ✓ Deployment scales up when window becomes active
- ✓ Deployment scales correctly through multiple window transitions
- ✓ Deployment returns to default after windows expire
- ✓ Status reflects current window accurately

**When to use:**
- Verify time-based scaling logic
- Test window transitions
- Demonstrate Kyklos behavior
- Debug timing issues

---

## Prerequisites

Both tests require:
- Kubernetes cluster (Kind, k3d, minikube, or real cluster)
- Kyklos controller deployed and running
- `kubectl` configured to access the cluster

## Quick Start

### Using Make Targets

```bash
# Run smoke test (30 seconds)
make smoke-test

# Run full sanity test (3 minutes)
make sanity-test
```

### Manual Execution

```bash
# Smoke test
cd test/sanity
./smoke-test.sh

# Sanity test
cd test/sanity
./run-sanity-test.sh
```

## Expected Output

### Smoke Test Success
```
========================================
Kyklos Smoke Test (30 seconds)
========================================

Creating active window: 10:35 - 10:45
Expected: Deployment should scale to 5 replicas immediately

[1/4] Creating test namespace...
[2/4] Creating deployment (1 replica)...
[3/4] Creating TimeWindowScaler...
[4/4] Waiting for scaling (max 20s)...
✓ SUCCESS: Deployment scaled to 5 replicas in 3s

========================================
Results:
========================================
Deployment Replicas: 5/5
TWS Current Window: active-now
TWS Effective Replicas: 5
TWS Ready Status: True

✓✓✓ SMOKE TEST PASSED ✓✓✓
Kyklos is functioning correctly!
```

### Sanity Test Success
```
========================================
Kyklos Quick Sanity Test
========================================

Current UTC time: 10:36

Test windows:
  Window 1 (scale to 3): 10:37 - 10:38
  Window 2 (scale to 5): 10:38 - 10:39
  Default replicas: 1

[1/5] Creating namespace...
[2/5] Creating test deployment...
[3/5] Waiting for deployment to be ready...
✓ Deployment ready with 1 replica(s)
[4/5] Creating TimeWindowScaler with minute-scale windows...
✓ TimeWindowScaler created
[5/5] Monitoring scaling behavior for 3 minutes...

[10:36:15] Replicas: 1, Window: Default
[10:37:05] Replicas: 3, Window: window-1
[10:38:05] Replicas: 5, Window: window-2
[10:39:05] Replicas: 1, Window: Default

========================================
Sanity Test Complete!
========================================
```

## Troubleshooting

### Smoke test fails
1. Check controller is running: `kubectl get pods -n kyklos-system`
2. Check controller logs: `kubectl logs -n kyklos-system deployment/kyklos-controller-manager`
3. Verify CRDs are installed: `kubectl get crd timewindowscalers.kyklos.kyklos.io`

### Sanity test doesn't scale
1. Verify current UTC time aligns with test windows
2. Check TWS status: `kubectl get tws -n kyklos-sanity -o yaml`
3. Check events: `kubectl get events -n kyklos-sanity`
4. Increase monitoring duration in script if needed

### Permission errors
Ensure your kubeconfig has permissions to:
- Create namespaces
- Create deployments
- Create TimeWindowScalers
- Get events

## Cleanup

Both tests clean up automatically, but if needed:

```bash
# Remove smoke test resources
kubectl delete namespace kyklos-smoke

# Remove sanity test resources
kubectl delete namespace kyklos-sanity
```

## Integration with CI/CD

### GitHub Actions Example
```yaml
- name: Run Kyklos Smoke Test
  run: make smoke-test

- name: Run Kyklos Sanity Test
  run: make sanity-test
  if: github.event_name == 'pull_request'
```

### GitLab CI Example
```yaml
smoke-test:
  script:
    - make smoke-test
  timeout: 1m

sanity-test:
  script:
    - make sanity-test
  timeout: 5m
  when: manual
```
