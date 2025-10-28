# Verification Checklist

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** local-workflow-designer

This checklist provides quick verification steps to ensure your Kyklos local development environment is healthy and operational. Use this after setup, before starting development, or when troubleshooting issues.

---

## Table of Contents

1. [Quick Health Check (2 minutes)](#quick-health-check-2-minutes)
2. [Detailed System Verification](#detailed-system-verification)
3. [Pre-Development Checklist](#pre-development-checklist)
4. [Post-Deployment Verification](#post-deployment-verification)
5. [Functional Testing Checklist](#functional-testing-checklist)
6. [Troubleshooting Decision Tree](#troubleshooting-decision-tree)

---

## Quick Health Check (2 minutes)

Run this checklist when you need a fast confirmation that everything is working.

### 1. Prerequisites Check

```bash
make verify-tools
```

**Expected Result:** All tools present and correct versions

**Pass Criteria:**
- [ ] Go 1.21+ installed
- [ ] Docker daemon running
- [ ] kubectl installed and functional
- [ ] Kind or k3d installed

**If Failed:** See [LOCAL-DEV-GUIDE.md - Prerequisites](./LOCAL-DEV-GUIDE.md#prerequisites)

---

### 2. Cluster Connectivity

```bash
kubectl cluster-info
```

**Expected Result:**
```
Kubernetes control plane is running at https://127.0.0.1:xxxxx
CoreDNS is running at https://127.0.0.1:xxxxx/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
```

**Pass Criteria:**
- [ ] Control plane URL shown
- [ ] No connection errors
- [ ] Response time < 2 seconds

**Quick Fix:**
```bash
# If failed, recreate cluster
make cluster-down cluster-up
```

---

### 3. Node Health

```bash
kubectl get nodes
```

**Expected Result:**
```
NAME                        STATUS   ROLES           AGE   VERSION
kyklos-dev-control-plane    Ready    control-plane   5m    v1.28.0
```

**Pass Criteria:**
- [ ] At least 1 node present
- [ ] Status: Ready
- [ ] Age > 1m (cluster stable)

**If Not Ready:**
```bash
kubectl describe node <node-name> | grep -A 10 Conditions
```

---

### 4. CRD Installation

```bash
kubectl get crd timewindowscalers.kyklos.io
```

**Expected Result:**
```
NAME                              CREATED AT
timewindowscalers.kyklos.io       2025-10-28T14:00:00Z
```

**Pass Criteria:**
- [ ] CRD exists
- [ ] No errors returned

**Quick Fix:**
```bash
make install-crds
```

---

### 5. Controller Health

```bash
kubectl get pods -n kyklos-system
```

**Expected Result:**
```
NAME                                        READY   STATUS    RESTARTS   AGE
kyklos-controller-manager-abc123-xyz        1/1     Running   0          2m
```

**Pass Criteria:**
- [ ] Pod status: Running
- [ ] Ready: 1/1
- [ ] Restarts: 0 (or < 3 if recent deploy)
- [ ] Age > 30s

**Quick Fix:**
```bash
# Check logs for errors
make logs-controller

# Restart if needed
make restart-controller
```

---

### 6. Controller Logs (No Errors)

```bash
make logs-controller | grep -i error | head -5
```

**Expected Result:** No output (no errors)

**Pass Criteria:**
- [ ] No ERROR level logs
- [ ] No FATAL logs
- [ ] No panic stack traces

**If Errors Found:**
```bash
# View full logs
make logs-controller

# Check specific error context
make logs-controller | grep -B 5 -A 5 ERROR
```

---

### 7. API Resource Availability

```bash
kubectl api-resources | grep timewindowscaler
```

**Expected Result:**
```
timewindowscalers    tws    kyklos.io/v1alpha1    true    TimeWindowScaler
```

**Pass Criteria:**
- [ ] Resource listed
- [ ] Shortname 'tws' available
- [ ] Namespaced: true

---

### Quick Health Summary

**All checks passed?** You're ready to develop!

**Any checks failed?**
1. Note which step failed
2. Run the Quick Fix command
3. Re-run the failed check
4. If still failing, see [Detailed System Verification](#detailed-system-verification)

---

## Detailed System Verification

Use this section for comprehensive environment validation.

### Build System

#### Check 1: Binary Compilation

```bash
make build
./bin/controller --version
```

**Expected Output:**
```
kyklos-controller version v0.1.0-dev
```

**Pass Criteria:**
- [ ] Binary builds without errors
- [ ] Binary is executable
- [ ] Version information displayed

**Detailed Check:**
```bash
# Check binary size (should be ~10-15 MB)
ls -lh bin/controller

# Verify it's statically linked (no external dependencies)
ldd bin/controller 2>&1 | grep -q "not a dynamic executable" && echo "✓ Static binary" || echo "✗ Dynamic binary"
```

---

#### Check 2: Container Image

```bash
make docker-build
docker images | grep kyklos
```

**Expected Output:**
```
kyklos/controller    dev     abc123def456   5 minutes ago   19.8MB
kyklos/controller    latest  abc123def456   5 minutes ago   19.8MB
```

**Pass Criteria:**
- [ ] Image built successfully
- [ ] Image size < 50 MB
- [ ] Both 'dev' and 'latest' tags present

**Detailed Check:**
```bash
# Inspect image
docker inspect kyklos/controller:dev | jq '.[0].Config.User'
# Expected: "65532:65532" (nonroot user)

# Test image runs
docker run --rm kyklos/controller:dev --version
```

---

#### Check 3: Image Loaded in Cluster

```bash
make verify-image-loaded
```

**For Kind:**
```bash
docker exec -it kyklos-dev-control-plane crictl images | grep kyklos
```

**Expected Output:**
```
docker.io/kyklos/controller    dev    abc123def456   19.8MB
```

**Pass Criteria:**
- [ ] Image present in cluster nodes
- [ ] Image ID matches local build

**Quick Fix:**
```bash
make kind-load
# or
make k3d-load
```

---

### Kubernetes Resources

#### Check 4: Namespace

```bash
kubectl get namespace kyklos-system
```

**Expected Output:**
```
NAME            STATUS   AGE
kyklos-system   Active   5m
```

**Pass Criteria:**
- [ ] Namespace exists
- [ ] Status: Active

---

#### Check 5: ServiceAccount

```bash
kubectl get serviceaccount -n kyklos-system kyklos-controller
```

**Expected Output:**
```
NAME                SECRETS   AGE
kyklos-controller   0         5m
```

**Pass Criteria:**
- [ ] ServiceAccount exists

---

#### Check 6: RBAC - ClusterRole

```bash
kubectl get clusterrole kyklos-controller-role
```

**Expected Output:**
```
NAME                      CREATED AT
kyklos-controller-role    2025-10-28T14:00:00Z
```

**Detailed Check:**
```bash
kubectl describe clusterrole kyklos-controller-role
```

**Required Permissions:**
- [ ] Deployments: get, list, watch, update, patch
- [ ] TimeWindowScalers: get, list, watch, update, patch (status)
- [ ] Events: create, patch
- [ ] ConfigMaps: get, list, watch (for holidays)

**Automated Verification:**
```bash
make verify-rbac
```

---

#### Check 7: RBAC - ClusterRoleBinding

```bash
kubectl get clusterrolebinding kyklos-controller-rolebinding
```

**Expected Output:**
```
NAME                            ROLE                              AGE
kyklos-controller-rolebinding   ClusterRole/kyklos-controller-role   5m
```

**Verify Binding:**
```bash
kubectl describe clusterrolebinding kyklos-controller-rolebinding | grep -A 5 Subjects
```

**Expected:**
```
Subjects:
  Kind            Name               Namespace
  ----            ----               ---------
  ServiceAccount  kyklos-controller  kyklos-system
```

---

#### Check 8: Deployment Configuration

```bash
kubectl get deployment -n kyklos-system kyklos-controller-manager -o yaml
```

**Key Configuration Checks:**

**Replicas:**
```bash
kubectl get deployment -n kyklos-system kyklos-controller-manager -o jsonpath='{.spec.replicas}'
# Expected: 1
```

**Image Pull Policy:**
```bash
kubectl get deployment -n kyklos-system kyklos-controller-manager -o jsonpath='{.spec.template.spec.containers[0].imagePullPolicy}'
# Expected: IfNotPresent (uses local image)
```

**Resource Limits:**
```bash
kubectl get deployment -n kyklos-system kyklos-controller-manager -o jsonpath='{.spec.template.spec.containers[0].resources}'
```

**Expected:**
```json
{
  "limits": {"cpu": "200m", "memory": "256Mi"},
  "requests": {"cpu": "100m", "memory": "128Mi"}
}
```

**Pass Criteria:**
- [ ] Replicas: 1
- [ ] Image: kyklos/controller:dev
- [ ] ImagePullPolicy: IfNotPresent
- [ ] Resource requests/limits defined

---

#### Check 9: Pod Status Details

```bash
kubectl describe pod -n kyklos-system -l app=kyklos-controller
```

**What to Verify:**

**Events Section:**
- [ ] No errors in events
- [ ] Image pulled successfully
- [ ] Container started
- [ ] No crash loops

**Conditions:**
```bash
kubectl get pod -n kyklos-system -l app=kyklos-controller -o jsonpath='{.items[0].status.conditions[?(@.type=="Ready")].status}'
# Expected: True
```

**Container State:**
```bash
kubectl get pod -n kyklos-system -l app=kyklos-controller -o jsonpath='{.items[0].status.containerStatuses[0].state}'
# Expected: {"running":{"startedAt":"..."}}
```

---

#### Check 10: Controller Logs Analysis

```bash
make logs-controller
```

**Look for Success Indicators:**
- [ ] "Starting controller" message
- [ ] "Starting workers" message
- [ ] "Listening for requests" on metrics port
- [ ] No "reconcile error" messages

**Warning Signs:**
- [ ] No "failed to get" errors
- [ ] No "unauthorized" errors
- [ ] No "context deadline exceeded" errors

**Expected Log Pattern:**
```
INFO  setup  Starting controller  {"controller": "timewindowscaler"}
INFO  setup  Starting EventSource  {"controller": "timewindowscaler", "source": "kind source: *v1alpha1.TimeWindowScaler"}
INFO  setup  Starting EventSource  {"controller": "timewindowscaler", "source": "kind source: *v1.Deployment"}
INFO  setup  Starting Controller  {"controller": "timewindowscaler"}
INFO  setup  Starting workers  {"controller": "timewindowscaler", "worker count": 1}
```

---

### CRD Validation

#### Check 11: CRD Schema

```bash
kubectl get crd timewindowscalers.kyklos.io -o yaml | grep -A 20 schema
```

**Pass Criteria:**
- [ ] OpenAPI v3 schema present
- [ ] Required fields defined: targetRef, timezone, windows
- [ ] Validation rules present

**Test Schema Validation:**
```bash
# Try to create invalid TWS (should fail)
kubectl apply -f - <<EOF
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: invalid-test
  namespace: default
spec:
  targetRef:
    kind: StatefulSet  # Invalid in v1alpha1
    name: test
  timezone: UTC
  windows: []  # Invalid: must have at least 1 window
EOF
```

**Expected:** Validation error preventing creation

---

#### Check 12: CRD Status Subresource

```bash
kubectl get crd timewindowscalers.kyklos.io -o jsonpath='{.spec.versions[0].subresources.status}'
```

**Expected Output:**
```json
{}
```

**Pass Criteria:**
- [ ] Status subresource enabled (output is '{}')

---

#### Check 13: CRD Additional Printer Columns

```bash
kubectl get crd timewindowscalers.kyklos.io -o jsonpath='{.spec.versions[0].additionalPrinterColumns}'
```

**Expected Columns:**
- WINDOW (currentWindow)
- REPLICAS (effectiveReplicas)
- TARGET (targetRef.name)
- AGE

**Test Display:**
```bash
# Create sample TWS and check columns
kubectl apply -f config/samples/basic.yaml
kubectl get tws
```

**Expected Output:**
```
NAME          WINDOW     REPLICAS   TARGET        AGE
basic         OffHours   2          webapp        5s
```

---

## Pre-Development Checklist

Run before starting development work to ensure clean state.

### Development Environment

- [ ] **Git Status Clean**
  ```bash
  git status
  # Expected: "nothing to commit, working tree clean"
  ```

- [ ] **On Correct Branch**
  ```bash
  git branch --show-current
  # Expected: Your feature branch name
  ```

- [ ] **Dependencies Up to Date**
  ```bash
  go mod tidy
  git diff go.mod go.sum
  # Expected: No changes
  ```

- [ ] **Generated Files Current**
  ```bash
  make manifests generate
  git status
  # Expected: No changes to api/ or config/crd/
  ```

- [ ] **Tests Pass**
  ```bash
  make test
  # Expected: All tests pass
  ```

- [ ] **Linter Clean**
  ```bash
  make lint
  # Expected: No issues
  ```

---

## Post-Deployment Verification

Run after deploying controller changes to verify correctness.

### Deployment Health

- [ ] **Pod Restarted Successfully**
  ```bash
  kubectl get pods -n kyklos-system -l app=kyklos-controller
  # Check AGE is recent, RESTARTS is 0
  ```

- [ ] **New Image Loaded**
  ```bash
  kubectl get pod -n kyklos-system -l app=kyklos-controller -o jsonpath='{.items[0].spec.containers[0].image}'
  # Expected: kyklos/controller:dev

  kubectl get pod -n kyklos-system -l app=kyklos-controller -o jsonpath='{.items[0].status.containerStatuses[0].imageID}'
  # Verify imageID changed
  ```

- [ ] **No Crash Loops**
  ```bash
  kubectl get pods -n kyklos-system -l app=kyklos-controller -o jsonpath='{.items[0].status.containerStatuses[0].restartCount}'
  # Expected: 0 (or low number if debugging)
  ```

- [ ] **Startup Logs Clean**
  ```bash
  make logs-controller | head -50
  # Verify "Starting controller" message present
  # Verify no ERROR messages
  ```

---

### Functional Verification

- [ ] **Controller Reconciles Existing Resources**
  ```bash
  # If you have existing TWS resources
  kubectl get tws --all-namespaces
  kubectl get events --all-namespaces --field-selector involvedObject.kind=TimeWindowScaler | tail -10
  # Verify recent events show reconciliation
  ```

- [ ] **Can Create New Resources**
  ```bash
  kubectl apply -f config/samples/basic.yaml
  kubectl get tws basic -o yaml | grep -A 5 status
  # Verify status is populated
  ```

- [ ] **Status Updates Working**
  ```bash
  # Check observedGeneration matches metadata.generation
  kubectl get tws basic -o jsonpath='{.metadata.generation} {.status.observedGeneration}'
  # Expected: Both numbers match (e.g., "1 1")
  ```

- [ ] **Events Generated**
  ```bash
  kubectl get events --field-selector involvedObject.kind=TimeWindowScaler
  # Verify events exist and are recent
  ```

---

## Functional Testing Checklist

Verify core Kyklos functionality works correctly.

### Basic Scaling Test

- [ ] **Create Test Environment**
  ```bash
  make demo-setup
  ```

- [ ] **Apply TimeWindowScaler**
  ```bash
  kubectl apply -f config/samples/basic.yaml
  ```

- [ ] **Verify Status Populated**
  ```bash
  kubectl get tws basic -o yaml | grep -A 20 status
  ```

  Check for:
  - currentWindow set
  - effectiveReplicas set
  - conditions present
  - observedGeneration matches spec

- [ ] **Verify Target Scaled**
  ```bash
  kubectl get deployment -n demo demo-app -o jsonpath='{.spec.replicas}'
  # Compare with TWS effectiveReplicas
  ```

- [ ] **Verify Events Created**
  ```bash
  kubectl get events -n demo | grep TimeWindowScaler
  ```

---

### Time Window Evaluation Test

- [ ] **Apply Minute-Scale Demo**
  ```bash
  make demo-apply-minute
  ```

- [ ] **Observe Window Transitions**
  ```bash
  make demo-watch
  # Watch for 2-3 minutes
  # Verify scaling occurs at window boundaries
  ```

- [ ] **Check Window Status Updates**
  ```bash
  # At even minute
  kubectl get tws demo-minute-scaler -n demo -o jsonpath='{.status.currentWindow}'
  # Expected: "BusinessHours"

  # At odd minute
  kubectl get tws demo-minute-scaler -n demo -o jsonpath='{.status.currentWindow}'
  # Expected: "OffHours"
  ```

---

### Manual Drift Correction Test

- [ ] **Create Base State**
  ```bash
  kubectl apply -f config/samples/basic.yaml
  kubectl wait --for=condition=Ready tws/basic --timeout=60s
  ```

- [ ] **Record Desired Replicas**
  ```bash
  DESIRED=$(kubectl get tws basic -o jsonpath='{.status.effectiveReplicas}')
  echo "Desired: $DESIRED"
  ```

- [ ] **Manually Change Deployment**
  ```bash
  kubectl scale deployment -n demo demo-app --replicas=99
  ```

- [ ] **Wait for Correction (should happen within 30 seconds)**
  ```bash
  sleep 30
  ACTUAL=$(kubectl get deployment -n demo demo-app -o jsonpath='{.spec.replicas}')
  echo "Actual: $ACTUAL"
  # Verify ACTUAL == DESIRED
  ```

- [ ] **Check Drift Correction Event**
  ```bash
  kubectl get events -n demo | grep -i drift
  # Should show DriftCorrected event
  ```

---

### Pause Functionality Test

- [ ] **Pause TimeWindowScaler**
  ```bash
  kubectl patch tws basic -n demo --type=merge -p '{"spec":{"pause":true}}'
  ```

- [ ] **Verify Status Shows Paused State**
  ```bash
  kubectl get tws basic -n demo -o yaml | grep -A 10 conditions
  # Should show Ready=False with reason indicating pause
  ```

- [ ] **Manually Scale (Should Not Be Corrected)**
  ```bash
  kubectl scale deployment -n demo demo-app --replicas=7
  sleep 30
  kubectl get deployment -n demo demo-app -o jsonpath='{.spec.replicas}'
  # Expected: Still 7 (not corrected while paused)
  ```

- [ ] **Resume and Verify Correction**
  ```bash
  kubectl patch tws basic -n demo --type=merge -p '{"spec":{"pause":false}}'
  sleep 30
  # Should now correct back to desired replicas
  ```

---

## Troubleshooting Decision Tree

### Problem: Controller Pod Not Starting

**Symptom:** Pod in CrashLoopBackOff or ImagePullBackOff

**Decision Path:**

1. Check pod status:
   ```bash
   kubectl describe pod -n kyklos-system -l app=kyklos-controller
   ```

2. **If ImagePullBackOff:**
   - [ ] Verify image exists: `docker images | grep kyklos`
   - [ ] Load image: `make kind-load`
   - [ ] Check imagePullPolicy: Should be `IfNotPresent`

3. **If CrashLoopBackOff:**
   - [ ] Check logs: `make logs-controller`
   - [ ] Look for panic or fatal error
   - [ ] Verify CRDs installed: `make install-crds`
   - [ ] Verify RBAC: `make verify-rbac`

4. **If Running but not Ready:**
   - [ ] Check readiness probe logs
   - [ ] Verify metrics endpoint: `make port-forward-metrics`

---

### Problem: TimeWindowScaler Not Scaling

**Symptom:** TWS created but deployment doesn't scale

**Decision Path:**

1. Check TWS status:
   ```bash
   kubectl get tws <name> -n <namespace> -o yaml | grep -A 20 status
   ```

2. **If status empty:**
   - [ ] Controller not reconciling
   - [ ] Check controller logs: `make logs-controller`
   - [ ] Verify RBAC permissions: `make verify-rbac`

3. **If status shows effectiveReplicas but deployment not scaled:**
   - [ ] Check deployment exists: `kubectl get deployment <target> -n <namespace>`
   - [ ] Check for manual drift: Compare effectiveReplicas with deployment.spec.replicas
   - [ ] Check if paused: `kubectl get tws <name> -o jsonpath='{.spec.pause}'`
   - [ ] Check controller has update permissions on deployments

4. **If Ready condition False:**
   - [ ] Read condition message for explanation
   - [ ] Check for TargetNotFound, InvalidTimezone, etc.

---

### Problem: Scaling Happens at Wrong Times

**Symptom:** Scaling occurs outside expected windows

**Decision Path:**

1. Verify timezone configuration:
   ```bash
   kubectl get tws <name> -o jsonpath='{.spec.timezone}'
   ```

2. Check current time in that timezone:
   ```bash
   TZ=<timezone> date
   ```

3. Review window definitions:
   ```bash
   kubectl get tws <name> -o yaml | grep -A 50 windows
   ```

4. Check controller's time interpretation:
   ```bash
   make logs-controller | grep "Current time"
   ```

5. **If timezone is correct:**
   - [ ] Verify window days match current day
   - [ ] Check for cross-midnight windows (end < start)
   - [ ] Check for overlapping windows (last match wins)

---

### Problem: Events Not Appearing

**Symptom:** Scaling occurs but no events generated

**Decision Path:**

1. Verify controller can create events:
   ```bash
   kubectl auth can-i create events -n <namespace> --as=system:serviceaccount:kyklos-system:kyklos-controller
   ```

2. Check all events in namespace:
   ```bash
   kubectl get events -n <namespace> --sort-by='.lastTimestamp'
   ```

3. **If permission denied:**
   - [ ] Run `make verify-rbac`
   - [ ] Reapply RBAC: `kubectl apply -f config/rbac/`

4. **If events exist but not showing in kubectl get events:**
   - Events may have expired (default TTL: 1 hour)
   - Check controller logs for event creation attempts

---

### Problem: Tests Failing

**Symptom:** `make test` fails

**Decision Path:**

1. Check which tests failed:
   ```bash
   make test 2>&1 | grep FAIL
   ```

2. **If import errors:**
   - [ ] Run `go mod tidy`
   - [ ] Run `go mod download`
   - [ ] Verify Go version: `go version`

3. **If compilation errors:**
   - [ ] Regenerate code: `make generate`
   - [ ] Clean and rebuild: `make clean build`

4. **If test logic errors:**
   - [ ] Run specific test: `go test -v ./path/to/package -run TestName`
   - [ ] Check test logs for root cause

---

## Verification Automation

### Create Custom Verification Script

Save this as `scripts/verify-all.sh`:

```bash
#!/bin/bash
set -e

echo "=== Kyklos Verification Script ==="
echo ""

FAILED=0

verify_check() {
    local name="$1"
    local command="$2"

    echo -n "Checking $name... "
    if eval "$command" > /dev/null 2>&1; then
        echo "✓"
    else
        echo "✗"
        FAILED=$((FAILED + 1))
    fi
}

verify_check "Go installed" "which go"
verify_check "Docker running" "docker info"
verify_check "kubectl installed" "which kubectl"
verify_check "Cluster reachable" "kubectl cluster-info"
verify_check "CRD installed" "kubectl get crd timewindowscalers.kyklos.io"
verify_check "Controller running" "kubectl get pods -n kyklos-system -l app=kyklos-controller | grep Running"
verify_check "Controller logs clean" "! kubectl logs -n kyklos-system -l app=kyklos-controller --tail=50 | grep ERROR"

echo ""
if [ $FAILED -eq 0 ]; then
    echo "✓ All checks passed!"
    exit 0
else
    echo "✗ $FAILED checks failed"
    exit 1
fi
```

Usage:
```bash
chmod +x scripts/verify-all.sh
./scripts/verify-all.sh
```

---

## Summary

### Minimum Viable Verification (30 seconds)

For a quick check that system is operational:

```bash
make verify-tools && \
kubectl get pods -n kyklos-system -l app=kyklos-controller | grep Running && \
kubectl get crd timewindowscalers.kyklos.io && \
echo "✓ System healthy"
```

### Comprehensive Verification (5 minutes)

For thorough validation before important work:

```bash
make verify-all && \
make test && \
make lint && \
make demo-setup && \
kubectl apply -f config/samples/basic.yaml && \
kubectl wait --for=condition=Ready tws/basic --timeout=60s && \
echo "✓ Full verification complete"
```

### Continuous Verification

Add to your development workflow:

```bash
# After every code change
make test lint

# After every deploy
make verify-controller

# After any API changes
make manifests generate && git diff config/
```

---

For troubleshooting specific issues, see [TROUBLESHOOTING.md](./TROUBLESHOOTING.md).

For setup instructions, see [LOCAL-DEV-GUIDE.md](./LOCAL-DEV-GUIDE.md).
