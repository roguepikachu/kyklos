# Kyklos E2E Test Plan

## Overview

End-to-end tests validate the complete Kyklos system in a real Kubernetes cluster, focusing on user-visible behavior and integration with actual workloads. Tests use minute-scale time windows for rapid execution while maintaining realistic scenarios.

## Test Environment

### Cluster Setup
```bash
# Local testing with kind
kind create cluster --name kyklos-e2e --config kind-config.yaml

# Install CRDs and controller
kubectl apply -f config/crd/bases/
kubectl apply -f config/rbac/
kubectl apply -f config/manager/
```

### Time Acceleration Strategy
To enable fast E2E tests, we use minute-scale windows:
- Production: Hours (09:00-17:00)
- Testing: Minutes (09:00-09:08 represents 8 hours)
- 1 minute test time = 1 hour real time

## E2E Test Scenarios

### E2E-001: Complete Daily Cycle
```yaml
test_id: E2E-001
category: e2e
description: Validate complete day cycle with multiple windows
duration: 5 minutes
setup:
  - Deploy controller to kyklos-system namespace
  - Create test namespace "e2e-daily"
  - Deploy nginx deployment with initial replicas=1
scenario:
  tws_config: |
    apiVersion: kyklos.io/v1alpha1
    kind: TimeWindowScaler
    metadata:
      name: daily-cycle
      namespace: e2e-daily
    spec:
      targetRef:
        kind: Deployment
        name: nginx
      timezone: UTC
      defaultReplicas: 1
      windows:
      - days: [Mon, Tue, Wed, Thu, Fri]
        start: "09:00"
        end: "09:02"  # 2-minute window (represents 09:00-11:00)
        replicas: 3
      - days: [Mon, Tue, Wed, Thu, Fri]
        start: "09:02"
        end: "09:05"  # 3-minute window (represents 11:00-14:00)
        replicas: 10
      - days: [Mon, Tue, Wed, Thu, Fri]
        start: "09:05"
        end: "09:07"  # 2-minute window (represents 14:00-16:00)
        replicas: 5
validation_timeline:
  - time: "+0s"
    commands:
      - kubectl apply -f tws-daily.yaml
      - kubectl wait --for=condition=Ready tws/daily-cycle -n e2e-daily --timeout=30s
    assertions:
      - deployment.spec.replicas: 1
      - tws.status.currentWindow: "OffHours"
  - time: "+30s"
    commands:
      - kubectl get deployment nginx -n e2e-daily -o jsonpath='{.spec.replicas}'
    assertions:
      - deployment.spec.replicas: 1
  - time: "09:00 UTC"
    commands:
      - kubectl get deployment nginx -n e2e-daily -o jsonpath='{.spec.replicas}'
      - kubectl get tws daily-cycle -n e2e-daily -o jsonpath='{.status.currentWindow}'
    assertions:
      - deployment.spec.replicas: 3
      - tws.status.currentWindow: "09:00-09:02"
  - time: "09:02 UTC"
    assertions:
      - deployment.spec.replicas: 10
      - tws.status.currentWindow: "09:02-09:05"
  - time: "09:05 UTC"
    assertions:
      - deployment.spec.replicas: 5
      - tws.status.currentWindow: "09:05-09:07"
  - time: "09:07 UTC"
    assertions:
      - deployment.spec.replicas: 1
      - tws.status.currentWindow: "OffHours"
cleanup:
  - kubectl delete namespace e2e-daily
success_criteria:
  - All replica transitions occur within 10s of scheduled time
  - No errors in controller logs
  - Events show all scaling operations
```

### E2E-002: Grace Period Validation
```yaml
test_id: E2E-002
category: e2e
description: Verify grace period delays downscaling
duration: 3 minutes
setup:
  - Create namespace "e2e-grace"
  - Deploy nginx with replicas=10
scenario:
  tws_config: |
    apiVersion: kyklos.io/v1alpha1
    kind: TimeWindowScaler
    metadata:
      name: grace-test
      namespace: e2e-grace
    spec:
      targetRef:
        kind: Deployment
        name: nginx
      timezone: UTC
      defaultReplicas: 2
      gracePeriodSeconds: 60  # 1 minute grace
      windows:
      - days: [Mon, Tue, Wed, Thu, Fri]
        start: "10:00"
        end: "10:01"  # 1-minute window
        replicas: 10
validation_timeline:
  - time: "10:00:30 UTC"
    assertions:
      - deployment.spec.replicas: 10
      - tws.status.effectiveReplicas: 10
  - time: "10:01:00 UTC"  # Window ends, grace starts
    assertions:
      - deployment.spec.replicas: 10  # Still 10 due to grace
      - tws.status.gracePeriodExpiry: "~10:02:00"
      - events: contains "Grace period started"
  - time: "10:01:30 UTC"  # Mid-grace
    assertions:
      - deployment.spec.replicas: 10
  - time: "10:02:05 UTC"  # After grace expires
    assertions:
      - deployment.spec.replicas: 2
      - tws.status.gracePeriodExpiry: null
      - events: contains "ScaledDown from 10 to 2"
success_criteria:
  - Grace period delays downscale by exactly 60 seconds
  - Upscaling bypasses grace period
  - Status tracks grace expiry accurately
```

### E2E-003: Manual Intervention Handling
```yaml
test_id: E2E-003
category: e2e
description: System corrects manual changes to deployment
duration: 2 minutes
setup:
  - Create namespace "e2e-drift"
  - Deploy nginx with replicas=5
  - Create TWS with defaultReplicas=5
validation_timeline:
  - time: "+10s"
    commands:
      - kubectl scale deployment nginx -n e2e-drift --replicas=20
    assertions:
      - deployment.spec.replicas: 20  # Immediately after manual change
  - time: "+20s"  # After reconciliation
    assertions:
      - deployment.spec.replicas: 5  # Corrected by controller
      - events: contains "DriftCorrected"
  - time: "+30s"
    commands:
      - kubectl patch tws drift-test -n e2e-drift --type=merge -p '{"spec":{"pause":true}}'
      - kubectl scale deployment nginx -n e2e-drift --replicas=15
    assertions:
      - deployment.spec.replicas: 15  # Not corrected when paused
  - time: "+40s"
    assertions:
      - deployment.spec.replicas: 15  # Still not corrected
      - tws.status.conditions: Ready=False, reason=Paused
success_criteria:
  - Manual changes detected and corrected within 30s
  - Pause mode prevents corrections
  - Clear events explain actions taken
```

### E2E-004: Holiday Mode Integration
```yaml
test_id: E2E-004
category: e2e
description: Holiday mode overrides normal schedule
duration: 2 minutes
setup:
  - Create namespace "e2e-holiday"
  - Create ConfigMap with holiday dates
  - Deploy nginx
scenario:
  configmap: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: holidays
      namespace: e2e-holiday
    data:
      "2025-03-10": "Company Holiday"
  tws_config: |
    apiVersion: kyklos.io/v1alpha1
    kind: TimeWindowScaler
    metadata:
      name: holiday-test
      namespace: e2e-holiday
    spec:
      targetRef:
        kind: Deployment
        name: nginx
      timezone: UTC
      defaultReplicas: 1
      windows:
      - days: [Mon]
        start: "09:00"
        end: "17:00"
        replicas: 10
      holidays:
        mode: treat-as-closed
        sourceRef:
          name: holidays
validation_timeline:
  - time: "Monday 2025-03-10 10:00 UTC"  # Holiday during window
    assertions:
      - deployment.spec.replicas: 1  # Default replicas (closed)
      - tws.status.currentWindow: "Holiday-Closed"
  - time: "Tuesday 2025-03-11 10:00 UTC"  # Normal day
    assertions:
      - deployment.spec.replicas: 0  # No window for Tuesday
success_criteria:
  - Holiday mode correctly overrides window
  - Status indicates holiday state
  - Non-holiday days work normally
```

### E2E-005: Multi-Resource Coordination
```yaml
test_id: E2E-005
category: e2e
description: Multiple TWS resources managing different deployments
duration: 3 minutes
setup:
  - Create namespace "e2e-multi"
  - Deploy frontend, backend, worker deployments
  - Create 3 TWS resources with different schedules
validation_timeline:
  - time: "09:00 UTC"
    assertions:
      - frontend.replicas: 10
      - backend.replicas: 5
      - worker.replicas: 2
  - time: "12:00 UTC"
    assertions:
      - frontend.replicas: 20  # Lunch peak
      - backend.replicas: 10   # Lunch peak
      - worker.replicas: 2     # No change
  - time: "17:00 UTC"
    assertions:
      - frontend.replicas: 2
      - backend.replicas: 2
      - worker.replicas: 10  # Night batch processing
success_criteria:
  - Each TWS manages its target independently
  - No interference between resources
  - Correct scheduling for each deployment
```

### E2E-006: Controller Restart Recovery
```yaml
test_id: E2E-006
category: e2e
description: Controller recovers state after restart
duration: 2 minutes
setup:
  - Create namespace "e2e-restart"
  - Deploy nginx with TWS
  - Set window to scale to 10 replicas
validation_timeline:
  - time: "+10s"
    commands:
      - kubectl rollout restart deployment kyklos-controller -n kyklos-system
    assertions:
      - deployment.spec.replicas: 10  # Before restart
  - time: "+40s"  # After controller restarts
    commands:
      - kubectl wait --for=condition=Available deployment/kyklos-controller -n kyklos-system
    assertions:
      - deployment.spec.replicas: 10  # Maintained after restart
      - tws.status.conditions: Ready=True
  - time: "+60s"
    commands:
      - kubectl scale deployment nginx -n e2e-restart --replicas=5
    assertions:
      - deployment.spec.replicas: 5  # Manual change
  - time: "+70s"
    assertions:
      - deployment.spec.replicas: 10  # Corrected after restart
success_criteria:
  - Controller restarts without losing state
  - Drift correction resumes after restart
  - No duplicate scaling events
```

### E2E-007: Webhook Validation
```yaml
test_id: E2E-007
category: e2e
description: Webhook prevents invalid configurations
duration: 1 minute
setup:
  - Create namespace "e2e-webhook"
test_cases:
  - name: "Invalid timezone"
    yaml: |
      apiVersion: kyklos.io/v1alpha1
      kind: TimeWindowScaler
      metadata:
        name: invalid-tz
        namespace: e2e-webhook
      spec:
        targetRef:
          kind: Deployment
          name: nginx
        timezone: "Invalid/Zone"
        defaultReplicas: 1
    expected_error: "unknown time zone Invalid/Zone"
  - name: "Invalid time format"
    yaml: |
      spec:
        windows:
        - days: [Mon]
          start: "25:00"
          end: "26:00"
          replicas: 5
    expected_error: "start time must match HH:MM format"
  - name: "Equal start and end"
    yaml: |
      spec:
        windows:
        - days: [Mon]
          start: "10:00"
          end: "10:00"
          replicas: 5
    expected_error: "start and end cannot be equal"
success_criteria:
  - Invalid resources rejected at creation time
  - Clear error messages returned
  - Valid resources accepted
```

### E2E-008: Performance Under Load
```yaml
test_id: E2E-008
category: e2e
description: Controller handles many resources efficiently
duration: 5 minutes
setup:
  - Create namespace "e2e-load"
  - Create 50 deployments
  - Create 50 TWS resources
validation:
  - All TWS resources become Ready within 60s
  - Controller memory usage < 100MB
  - Controller CPU usage < 100m average
  - No reconciliation takes > 1s
  - All deployments scaled correctly
metrics_queries:
  - container_memory_usage_bytes{pod=~"kyklos-controller.*"}
  - rate(container_cpu_usage_seconds_total{pod=~"kyklos-controller.*"}[1m])
  - histogram_quantile(0.99, controller_runtime_reconcile_time_seconds_bucket)
success_criteria:
  - P99 reconcile time < 500ms
  - No OOM or CPU throttling
  - All resources reconciled successfully
```

### E2E-009: Upgrade Compatibility
```yaml
test_id: E2E-009
category: e2e
description: Smooth upgrade from previous version
duration: 3 minutes
setup:
  - Install controller v0.1.0
  - Create TWS resources
  - Verify scaling works
upgrade_steps:
  - time: "+30s"
    commands:
      - kubectl apply -f manifests/v0.2.0/
    validations:
      - No TWS resources deleted
      - Deployments maintain replicas
  - time: "+60s"
    validations:
      - All TWS Ready=True
      - New features available
      - Backward compatibility maintained
success_criteria:
  - Zero downtime during upgrade
  - Existing resources continue working
  - No data loss or state corruption
```

### E2E-010: Resource Cleanup
```yaml
test_id: E2E-010
category: e2e
description: Proper cleanup on TWS deletion
duration: 2 minutes
setup:
  - Create namespace "e2e-cleanup"
  - Deploy nginx with 1 replica
  - Create TWS scaling to 10 replicas
validation_timeline:
  - time: "+10s"
    assertions:
      - deployment.spec.replicas: 10
      - tws.metadata.finalizers: ["kyklos.io/finalizer"]
  - time: "+20s"
    commands:
      - kubectl delete tws cleanup-test -n e2e-cleanup
  - time: "+30s"
    assertions:
      - tws: not found
      - deployment.spec.replicas: 10  # Unchanged
      - events: contains "Deleting"
  - time: "+40s"
    commands:
      - kubectl delete namespace e2e-cleanup
    assertions:
      - namespace: terminating
      - No hanging resources
success_criteria:
  - TWS deletion doesn't affect deployment replicas
  - Finalizer ensures clean deletion
  - Namespace deletes cleanly
```

## Test Execution

### Local Execution
```bash
# Run all E2E tests
make test-e2e

# Run specific test
make test-e2e TEST=E2E-001

# Run with verbose output
make test-e2e VERBOSE=true

# Keep cluster after tests
make test-e2e KEEP_CLUSTER=true
```

### CI Pipeline
```yaml
name: E2E Tests
on:
  pull_request:
    types: [opened, synchronize]
  issue_comment:
    types: [created]

jobs:
  e2e:
    if: contains(github.event.comment.body, '/test e2e')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Create kind cluster
        run: |
          kind create cluster --config test/e2e/kind-config.yaml
      - name: Install CRDs
        run: make install
      - name: Run E2E tests
        run: make test-e2e
      - name: Collect logs
        if: failure()
        run: |
          kubectl logs -n kyklos-system -l app=kyklos-controller --tail=1000
```

## Acceptance Criteria

### Functional Requirements
- [ ] All time windows trigger at correct times (±10s tolerance)
- [ ] Grace periods delay downscaling by exact duration
- [ ] Manual changes corrected within one reconciliation cycle
- [ ] Holiday mode overrides work correctly
- [ ] Pause mode prevents all modifications
- [ ] Controller recovers from restarts
- [ ] Resources clean up properly

### Performance Requirements
- [ ] Window transitions complete within 10 seconds
- [ ] Controller uses <100MB memory with 50 resources
- [ ] P99 reconciliation time <500ms
- [ ] No memory leaks over 24-hour run
- [ ] CPU usage <100m average

### Reliability Requirements
- [ ] Zero data loss during upgrades
- [ ] Graceful handling of missing targets
- [ ] Recovery from temporary API server outages
- [ ] No duplicate scaling events
- [ ] Consistent state after controller restarts

## Test Data

### Standard Deployments
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  namespace: e2e-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        resources:
          requests:
            memory: "10Mi"
            cpu: "10m"
          limits:
            memory: "20Mi"
            cpu: "20m"
```

### Minute-Scale TWS
```yaml
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: minute-demo
  namespace: e2e-test
spec:
  targetRef:
    kind: Deployment
    name: test-app
  timezone: UTC
  defaultReplicas: 1
  windows:
  # Each minute represents an hour
  - days: [Mon, Tue, Wed, Thu, Fri]
    start: "10:00"  # 10:00 AM
    end: "10:08"    # 6:00 PM (8 hours = 8 minutes)
    replicas: 10
```

## Success Metrics

### Test Execution
- All tests pass on first run (no flakes)
- Total suite execution time <10 minutes
- Clear error messages on failures
- Logs provide debugging information

### Coverage
- All user scenarios covered
- Edge cases validated
- Error paths tested
- Performance validated

### Documentation
- Each test has clear purpose
- Acceptance criteria documented
- Setup/teardown automated
- Results reproducible