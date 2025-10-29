# Implementation Risks and Mitigations

**Purpose:** Identify implementation risks for Kyklos v0.1 with detection strategies and rollback plans.

**Last Updated:** 2025-10-29

## Risk Assessment Framework

**Risk Levels:**
- **Critical:** Blocks release, requires immediate resolution
- **High:** Major impact on functionality or user experience
- **Medium:** Moderate impact, workaround available
- **Low:** Minor impact, can be deferred to v0.2

**Impact Categories:**
- **Functional:** Core features don't work
- **Performance:** Unacceptable latency or resource usage
- **Security:** Privilege escalation or data exposure
- **Operational:** Difficult to deploy or maintain
- **User Experience:** Confusing or frustrating to use

---

## Risk Register

### Risk 1: Time Calculation Bugs

**Risk Level:** Critical
**Category:** Functional
**Probability:** High (complex time math, DST, cross-midnight)

#### Description

Incorrect time window calculations could result in:
- Scaling at wrong times
- Missing window boundaries
- Incorrect cross-midnight handling
- DST transition errors

#### Impact

- User workloads scaled incorrectly (over-provision or under-provision)
- Loss of trust in controller reliability
- Potential production incidents (insufficient replicas during peak)

#### Detection Strategies

**During Development:**
```bash
# 1. Unit test coverage requirement: 100% for timecalc package
make test-unit
go test ./internal/timecalc/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep timecalc

# 2. Table-driven tests with fixed dates
# See: /test/fixtures/time-test-cases.yaml

# 3. DST transition tests with specific dates
# 2025-03-09 02:00:00 (spring forward)
# 2025-11-02 02:00:00 (fall back)
```

**During Testing:**
```bash
# 1. E2E tests with time-warp (minute-scale windows)
make test-e2e

# 2. Manual verification with real timezones
# Deploy to cluster in different timezone
# Verify scaling at expected local times

# 3. Cross-midnight scenario validation
./test/e2e/scenario-cross-midnight.sh
```

**In Production:**
```prometheus
# Monitor for unexpected scaling
kyklos_scale_events_total{direction="up"}
kyklos_scale_events_total{direction="down"}

# Alert on mismatched effective replicas
abs(kyklos_effective_replicas - kyklos_target_observed_replicas) > 0
```

#### Mitigation Plan

**Preventive:**
1. Use Go standard library `time` package (battle-tested)
2. IANA timezone database (system-provided)
3. Extensive unit tests with edge cases:
   - All day boundaries (Mon-Sun)
   - Cross-midnight windows (22:00-02:00)
   - DST transitions (spring forward, fall back)
   - Leap years (2024-02-29)
   - Year boundaries (2025-12-31 23:59:59)

**Detective:**
1. Unit tests run on every commit (CI)
2. E2E tests run before release
3. Manual testing in multiple timezones
4. Metrics monitoring in staging

**Corrective:**
1. If bug found in production:
   - Pause affected TimeWindowScalers (spec.pause=true)
   - Deploy hotfix within 24 hours
   - Backfill missed scaling operations manually

**Rollback Plan:**
```bash
# 1. Pause all TimeWindowScalers
kubectl get tws -A -o name | xargs -I {} kubectl patch {} --type=merge -p '{"spec":{"pause":true}}'

# 2. Rollback controller to previous version
kubectl set image deployment/kyklos-controller -n kyklos-system controller=ghcr.io/aykumar/kyklos-controller:v0.0.9

# 3. Verify previous version running
kubectl rollout status deployment/kyklos-controller -n kyklos-system

# 4. Unpause TimeWindowScalers after verification
kubectl get tws -A -o name | xargs -I {} kubectl patch {} --type=merge -p '{"spec":{"pause":false}}'
```

---

### Risk 2: Hot Reconcile Loop

**Risk Level:** High
**Category:** Performance
**Probability:** Medium (can happen with incorrect requeue logic)

#### Description

Controller enters tight reconcile loop causing:
- Excessive API server load
- High CPU usage in controller pod
- Increased etcd pressure
- Rate limiting errors

#### Impact

- Controller pod CPU throttled or OOMKilled
- API server performance degradation
- Other controllers affected by rate limiting
- Delayed reconciliation for all TimeWindowScalers

#### Detection Strategies

**During Development:**
```go
// Add minimum requeue enforcement in code
const MinRequeueDelay = 30 * time.Second

func computeRequeueDuration(nextBoundary time.Time, now time.Time) time.Duration {
    baseDuration := nextBoundary.Sub(now)
    // Always enforce minimum
    if baseDuration < MinRequeueDelay {
        return MinRequeueDelay
    }
    // ...
}
```

**During Testing:**
```bash
# Monitor requeue intervals in logs
kubectl logs -n kyklos-system -l app=kyklos-controller | grep "Requeue"

# Should see intervals >= 30 seconds
# If seeing < 30 seconds, hot loop detected
```

**In Production:**
```prometheus
# Monitor reconciliation rate
rate(kyklos_reconcile_total[5m]) > 2  # More than 2/second indicates hot loop

# Monitor reconcile duration
histogram_quantile(0.99, rate(kyklos_reconcile_duration_seconds_bucket[5m])) > 5

# Alert rule
- alert: KyklosHotReconcileLoop
  expr: rate(kyklos_reconcile_total{tws_name="*"}[1m]) > 1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Kyklos controller in hot reconcile loop"
```

#### Mitigation Plan

**Preventive:**
1. Enforce minimum requeue delay (30 seconds)
2. Add jitter to prevent thundering herd (5-25 seconds)
3. Quantize requeue times to 10-second boundaries
4. Maximum requeue delay (24 hours) to prevent distant boundaries

**Detective:**
1. Reconcile rate metrics exposed
2. Log requeue durations at debug level
3. CI tests verify requeue intervals

**Corrective:**
1. If hot loop detected:
   - Scale controller to 0 replicas (stop reconciliation)
   - Investigate logs for root cause
   - Deploy patched version with increased minimum requeue
   - Gradually scale back up

**Rollback Plan:**
```bash
# 1. Stop controller immediately
kubectl scale deployment kyklos-controller -n kyklos-system --replicas=0

# 2. Analyze logs
kubectl logs -n kyklos-system -l app=kyklos-controller --previous > hotloop-logs.txt

# 3. Deploy emergency patch
# Edit deployment to add --min-requeue-delay=60s arg
kubectl edit deployment kyklos-controller -n kyklos-system

# 4. Scale back to 1 replica
kubectl scale deployment kyklos-controller -n kyklos-system --replicas=1

# 5. Monitor reconcile rate
watch "kubectl top pod -n kyklos-system"
```

---

### Risk 3: Grace Period State Loss

**Risk Level:** High
**Category:** Functional
**Probability:** Medium (controller restart during grace period)

#### Description

Controller restart could lose grace period state if not persisted:
- In-memory grace timer lost
- Replicas scaled down immediately instead of waiting
- Users expect grace period to be honored across restarts

#### Impact

- Unexpected immediate scale-down on controller restart
- Grace period feature unreliable
- Potential disruption to workloads

#### Detection Strategies

**During Development:**
```go
// Persist grace period expiry in status (not in-memory timer)
type TimeWindowScalerStatus struct {
    // ...
    GracePeriodExpiry *metav1.Time `json:"gracePeriodExpiry,omitempty"`
}

// Reconciler checks status, not timer
if status.GracePeriodExpiry != nil && now.Before(status.GracePeriodExpiry.Time) {
    // Still in grace period
}
```

**During Testing:**
```bash
# Integration test: Controller restart during grace
# test/e2e/grace-period-restart.sh

# 1. Create TWS with grace period
# 2. Trigger scale-down (start grace period)
# 3. Restart controller pod
# 4. Verify grace period still honored
# 5. Wait for expiry
# 6. Verify scale-down applied
```

**In Production:**
```prometheus
# Monitor grace period cancellations
kyklos_grace_periods_total{state="cancelled"} > 0

# Should only cancel on scale-up, not on restart
```

#### Mitigation Plan

**Preventive:**
1. Store `gracePeriodExpiry` timestamp in status subresource (persisted in etcd)
2. Reconciler reads expiry from status, not in-memory timer
3. Integration test validates restart behavior

**Detective:**
1. Test includes controller restart scenario
2. Monitor GracePeriodCancelled events

**Corrective:**
1. If grace period lost on restart:
   - No immediate fix (already scaled down)
   - Hotfix controller to persist expiry
   - Deploy updated controller
   - Document known issue in v0.1.0 release notes

**Rollback Plan:**
```bash
# If this is discovered post-release:

# 1. Document workaround
# Users should not restart controller during grace period
# Or set grace period longer (e.g., 600s instead of 300s)

# 2. Prioritize fix in v0.1.1 patch release

# 3. No rollback needed (not breaking existing functionality)
```

---

### Risk 4: RBAC Gaps

**Risk Level:** High
**Category:** Security
**Probability:** Low (good tooling, but easy to miss)

#### Description

Insufficient RBAC permissions cause:
- Controller unable to read TimeWindowScalers
- Controller unable to update Deployments
- Controller unable to create Events
- Silent failures with cryptic error messages

#### Impact

- Controller non-functional
- Difficult to diagnose (permission errors in logs)
- Blocked deployments for users

#### Detection Strategies

**During Development:**
```bash
# Generate RBAC from markers
make manifests

# Verify generated RBAC
cat config/rbac/role.yaml
```

**During Testing:**
```bash
# Test RBAC in fresh cluster
make kind-cluster
make deploy

# Attempt operations
kubectl apply -f config/samples/basic.yaml

# Check for permission errors
kubectl logs -n kyklos-system -l app=kyklos-controller | grep "forbidden"
```

**In Production:**
```bash
# Monitor for permission errors in logs
kubectl logs -n kyklos-system -l app=kyklos-controller | grep -i "forbidden\|unauthorized"
```

#### Mitigation Plan

**Preventive:**
1. Use kubebuilder RBAC markers (auto-generated)
2. Document required permissions in RBAC-MATRIX.md
3. Test in fresh cluster (no cluster-admin)
4. Provide both Role (same-namespace) and ClusterRole (cross-namespace)

**Detective:**
1. CI tests deploy to fresh cluster
2. E2E tests verify all operations succeed
3. Log analysis for permission errors

**Corrective:**
1. If RBAC gap found:
   - Add missing permission to role.yaml
   - Apply updated RBAC: `kubectl apply -f config/rbac/`
   - No controller restart needed (permissions take effect immediately)

**Rollback Plan:**
```bash
# If overly permissive RBAC deployed:

# 1. Apply restrictive RBAC immediately
kubectl apply -f config/rbac/role-minimal.yaml

# 2. Verify controller still functional
kubectl logs -n kyklos-system -l app=kyklos-controller

# 3. Test operations
kubectl apply -f config/samples/basic.yaml

# 4. Gradually add back permissions as needed
```

**RBAC Audit Checklist:**
```bash
# Before release, verify all permissions:

# TimeWindowScalers: get, list, watch
kubectl auth can-i get timewindowscalers --as=system:serviceaccount:kyklos-system:kyklos-controller

# TimeWindowScalers/status: get, update, patch
kubectl auth can-i update timewindowscalers/status --as=system:serviceaccount:kyklos-system:kyklos-controller

# Deployments: get, list, watch, update, patch
kubectl auth can-i update deployments --as=system:serviceaccount:kyklos-system:kyklos-controller

# ConfigMaps: get, list, watch (for holidays)
kubectl auth can-i get configmaps --as=system:serviceaccount:kyklos-system:kyklos-controller

# Events: create, patch
kubectl auth can-i create events --as=system:serviceaccount:kyklos-system:kyklos-controller
```

---

### Risk 5: Cross-Midnight Edge Cases

**Risk Level:** Medium
**Category:** Functional
**Probability:** Medium (complex logic, easy to get wrong)

#### Description

Cross-midnight windows (e.g., 22:00-02:00) have subtle edge cases:
- Day boundary confusion (Friday 22:00 vs Saturday 02:00)
- Weekend windows not excluded correctly
- Next boundary calculation wrong

#### Impact

- Incorrect scaling on cross-midnight boundaries
- Weekend scaling when not intended
- User confusion about window behavior

#### Detection Strategies

**During Development:**
```go
// Dedicated test suite for cross-midnight
func TestComputeEffectiveReplicas_CrossMidnight(t *testing.T) {
    tests := []struct {
        name      string
        window    TimeWindow{Days: []string{"Fri"}, Start: "22:00", End: "02:00"}
        localTime time.Time
        want      int32
    }{
        {"Friday 21:00", friday2100, 2},    // Before window
        {"Friday 23:00", friday2300, 10},   // In window (same day)
        {"Saturday 01:00", saturday0100, 10}, // In window (next day)
        {"Saturday 03:00", saturday0300, 2},  // After window
        {"Saturday 23:00", saturday2300, 2},  // Not in window (different day)
    }
    // ...
}
```

**During Testing:**
```bash
# E2E test with cross-midnight scenario
./test/e2e/scenario-cross-midnight.sh

# Verify behavior at each hour across midnight
```

**In Production:**
```bash
# User reports unexpected scaling on weekend
# Check logs for window evaluation

kubectl logs -n kyklos-system -l app=kyklos-controller | grep "window matching"
```

#### Mitigation Plan

**Preventive:**
1. Comprehensive unit tests for cross-midnight logic
2. Clear documentation in CRD-SPEC.md with examples
3. E2E scenario validates cross-midnight behavior
4. Pseudocode documented in PSEUDOCODE.md

**Detective:**
1. Unit tests cover all edge cases
2. E2E tests verify Friday→Saturday boundary
3. User documentation includes cross-midnight examples

**Corrective:**
1. If cross-midnight bug found:
   - Identify incorrect logic in timecalc/matcher.go
   - Add regression test
   - Fix logic (ensure yesterday check only applies to cross-midnight)
   - Deploy hotfix

**Rollback Plan:**
```bash
# If cross-midnight logic is broken:

# 1. Advise users to avoid cross-midnight windows temporarily
# Add to KNOWN-ISSUES.md in v0.1.0 release notes

# 2. Users can work around with two separate windows:
# - Window 1: Mon-Fri 22:00-23:59 (10 replicas)
# - Window 2: Tue-Sat 00:00-02:00 (10 replicas)
# (Overlapping is okay, last window wins)

# 3. Fix in v0.1.1 patch release within 1 week
```

---

### Risk 6: Holiday ConfigMap Not Found

**Risk Level:** Low
**Category:** Operational
**Probability:** High (user error, ConfigMap not created)

#### Description

User configures holiday support but forgets to create ConfigMap:
- Controller sets Degraded condition
- Scaling still works (falls back to normal windows)
- User confused about why holiday not applied

#### Impact

- Minor: Scaling continues with ignore mode
- User confusion
- Support burden

#### Detection Strategies

**During Development:**
```go
// Handle missing ConfigMap gracefully
configMap, err := r.Get(ctx, holidayConfigMapName, configMap)
if errors.IsNotFound(err) {
    // Set Degraded condition but continue
    conditions = append(conditions, metav1.Condition{
        Type:    "Degraded",
        Status:  metav1.ConditionTrue,
        Reason:  "HolidaySourceMissing",
        Message: fmt.Sprintf("ConfigMap %s not found", holidayConfigMapName),
    })
    // Continue with ignore mode
    holidayDates = map[string]bool{}
} else if err != nil {
    return ctrl.Result{}, err
}
```

**During Testing:**
```bash
# Test missing ConfigMap scenario
# Deploy TWS with holidays but no ConfigMap
kubectl apply -f - <<EOF
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: test-missing-configmap
spec:
  # ...
  holidays:
    mode: treat-as-closed
    sourceRef:
      name: nonexistent-configmap
EOF

# Verify Degraded condition set
kubectl get tws test-missing-configmap -o jsonpath='{.status.conditions[?(@.type=="Degraded")]}'
```

**In Production:**
```bash
# User reports holiday not working
# Check Degraded condition
kubectl get tws -A -o json | jq '.items[] | select(.status.conditions[] | select(.type=="Degraded" and .reason=="HolidaySourceMissing"))'
```

#### Mitigation Plan

**Preventive:**
1. Clear documentation: holiday ConfigMap must be created first
2. Example YAMLs include both ConfigMap and TWS
3. Validation webhook (v0.2+) checks ConfigMap exists

**Detective:**
1. Degraded condition with HolidaySourceMissing reason
2. Event: "HolidaySourceMissing" emitted
3. Controller logs warning

**Corrective:**
1. User creates missing ConfigMap:
   ```bash
   kubectl apply -f config/samples/holidays-configmap.yaml
   ```
2. Controller reconciles within 5 minutes (requeue on Degraded)
3. Degraded condition clears automatically

**Rollback Plan:**
```bash
# Not applicable (no rollback needed, user action required)

# Workaround:
# 1. Disable holiday support temporarily
kubectl patch tws <name> --type=merge -p '{"spec":{"holidays":null}}'

# 2. Create ConfigMap
kubectl create configmap company-holidays --from-literal="2025-12-25=Christmas"

# 3. Re-enable holiday support
kubectl patch tws <name> --type=merge -p '{"spec":{"holidays":{"mode":"treat-as-closed","sourceRef":{"name":"company-holidays"}}}}'
```

---

### Risk 7: Timezone Typo

**Risk Level:** Low
**Category:** User Experience
**Probability:** High (user input error)

#### Description

User specifies invalid timezone (e.g., "America/NewYork" instead of "America/New_York"):
- Controller cannot load timezone
- Degraded condition set
- No scaling occurs

#### Impact

- No functional scaling until fixed
- User frustration
- Support burden

#### Detection Strategies

**During Development:**
```go
// Validate timezone on load
location, err := time.LoadLocation(spec.Timezone)
if err != nil {
    conditions = append(conditions, metav1.Condition{
        Type:    "Degraded",
        Status:  metav1.ConditionTrue,
        Reason:  "InvalidTimezone",
        Message: fmt.Sprintf("Cannot load timezone %s: %v", spec.Timezone, err),
    })
    // Use defaultReplicas as fallback
}
```

**During Testing:**
```bash
# Test invalid timezone handling
kubectl apply -f - <<EOF
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: test-invalid-tz
spec:
  targetRef:
    kind: Deployment
    name: test
  timezone: "Invalid/Timezone"
  defaultReplicas: 2
  windows:
  - days: ["Mon"]
    start: "09:00"
    end: "17:00"
    replicas: 10
EOF

# Verify Degraded condition
kubectl get tws test-invalid-tz -o jsonpath='{.status.conditions[?(@.type=="Degraded")]}'
```

**In Production:**
```bash
# User reports "controller not working"
# Check for InvalidTimezone
kubectl get tws -A -o json | jq '.items[] | select(.status.conditions[] | select(.reason=="InvalidTimezone"))'
```

#### Mitigation Plan

**Preventive:**
1. CRD validation with enum (v0.2+: list common timezones)
2. Documentation: list common IANA timezones
3. Admission webhook validates timezone (v0.2+)

**Detective:**
1. Degraded condition with InvalidTimezone reason
2. Event: "InvalidSchedule" with timezone error
3. Controller logs error

**Corrective:**
1. User fixes timezone:
   ```bash
   kubectl patch tws <name> --type=merge -p '{"spec":{"timezone":"America/New_York"}}'
   ```
2. Controller reconciles immediately (watches for spec changes)
3. Degraded condition clears

**Rollback Plan:**
```bash
# No rollback needed (user action)

# Helper: List valid timezones
timedatectl list-timezones | grep America
```

---

## Risk Monitoring Dashboard

### Prometheus Queries

```yaml
# Hot reconcile loop detection
rate(kyklos_reconcile_total[1m]) > 1

# Degraded TimeWindowScalers
kyklos_degraded_total > 0

# Failed scale operations
rate(kyklos_scale_failures_total[5m]) > 0

# High reconcile duration
histogram_quantile(0.99, rate(kyklos_reconcile_duration_seconds_bucket[5m])) > 5

# Grace period anomalies
rate(kyklos_grace_periods_total{state="cancelled"}[5m]) > rate(kyklos_scale_events_total{direction="up"}[5m])
```

### Alerting Rules

```yaml
groups:
- name: kyklos-controller
  rules:
  - alert: KyklosControllerDown
    expr: up{job="kyklos-controller"} == 0
    for: 5m
    severity: critical
    annotations:
      summary: "Kyklos controller is down"

  - alert: KyklosHotLoop
    expr: rate(kyklos_reconcile_total[1m]) > 1
    for: 5m
    severity: critical
    annotations:
      summary: "Kyklos controller in hot reconcile loop"

  - alert: KyklosDegraded
    expr: kyklos_degraded_total > 0
    for: 10m
    severity: warning
    annotations:
      summary: "One or more TimeWindowScalers degraded"

  - alert: KyklosScaleFailed
    expr: rate(kyklos_scale_failures_total[5m]) > 0
    for: 5m
    severity: warning
    annotations:
      summary: "Scale operations failing"
```

---

## Emergency Response Playbook

### Scenario: Controller Completely Broken

```bash
# 1. Immediate: Pause all TimeWindowScalers
kubectl get tws -A -o json | jq -r '.items[] | "\(.metadata.namespace)/\(.metadata.name)"' | \
  xargs -I {} sh -c 'kubectl patch tws {} --type=merge -p "{\"spec\":{\"pause\":true}}"'

# 2. Scale controller to 0 (stop reconciliation)
kubectl scale deployment kyklos-controller -n kyklos-system --replicas=0

# 3. Manually scale Deployments to safe replica counts
kubectl get deploy -A -o json | jq -r '.items[] | "\(.metadata.namespace) \(.metadata.name)"' | \
  xargs -n2 sh -c 'kubectl scale deployment $1 -n $0 --replicas=3'

# 4. Rollback to previous version
kubectl set image deployment/kyklos-controller -n kyklos-system \
  controller=ghcr.io/aykumar/kyklos-controller:v0.0.9

# 5. Scale controller back to 1
kubectl scale deployment kyklos-controller -n kyklos-system --replicas=1

# 6. Monitor logs
kubectl logs -n kyklos-system -l app=kyklos-controller -f

# 7. Gradually unpause TimeWindowScalers
# (manually, one at a time, verifying behavior)
```

### Scenario: Time Calculation Bug Discovered

```bash
# 1. Document the bug
# - What time/timezone/window causes incorrect behavior?
# - What is expected vs actual replica count?

# 2. Pause affected TimeWindowScalers only
kubectl patch tws <affected-tws> -n <namespace> --type=merge -p '{"spec":{"pause":true}}'

# 3. Create hotfix branch
git checkout -b hotfix/v0.1.1-time-calc

# 4. Fix logic in internal/timecalc/matcher.go
# 5. Add regression test
# 6. Build and push patched image
# 7. Deploy hotfix

# 8. Verify fix in staging
kubectl kustomize config/overlays/staging | kubectl apply -f -

# 9. Deploy to production
kubectl set image deployment/kyklos-controller -n kyklos-system \
  controller=ghcr.io/aykumar/kyklos-controller:v0.1.1

# 10. Unpause TimeWindowScalers
kubectl patch tws <affected-tws> -n <namespace> --type=merge -p '{"spec":{"pause":false}}'
```

---

## Post-Incident Review Template

After any production incident:

1. **What happened?** (timeline of events)
2. **Root cause?** (technical explanation)
3. **Detection:** How was it discovered?
4. **Impact:** What was affected and for how long?
5. **Resolution:** What fixed it?
6. **Action items:**
   - Prevent recurrence
   - Improve detection
   - Update documentation

---

## Summary

**Critical Risks (Must address before v0.1.0):**
1. Time calculation bugs → 100% unit test coverage
2. Hot reconcile loop → Enforce minimum requeue delay
3. Grace period state loss → Persist expiry in status
4. RBAC gaps → Test in fresh cluster

**High Risks (Address in v0.1.x patches):**
5. Cross-midnight edge cases → Comprehensive test suite

**Medium/Low Risks (Acceptable for v0.1.0):**
6. Holiday ConfigMap not found → Clear documentation
7. Timezone typo → Helpful error messages

All risks have defined detection strategies and rollback plans.
