# Kyklos Envtest Plan

## Overview

This document specifies controller tests using the controller-runtime envtest framework. These tests validate the complete reconciliation loop, status management, and Kubernetes API interactions with a real API server but without a full cluster.

## Test Environment Setup

### Envtest Configuration
```go
testEnv = &envtest.Environment{
    CRDDirectoryPaths: []string{"../../config/crd/bases"},
    ErrorIfCRDPathMissing: true,
    WebhookInstallOptions: envtest.WebhookInstallOptions{
        Paths: []string{"../../config/webhook"},
    },
}

// Use fixed clock for deterministic time
clock := clocktesting.NewFakeClock(time.Date(2025, 3, 10, 14, 0, 0, 0, time.UTC))
```

## Test Scenarios

### ET-001: Happy Path - Business Hours Scaling
```yaml
test_id: ET-001
category: envtest
description: Scale deployment based on business hours window
setup:
  - Create namespace "test-et001"
  - Create Deployment "webapp" with replicas=1
  - Create TimeWindowScaler targeting webapp
  tws_spec:
    timezone: "UTC"
    defaultReplicas: 2
    windows:
      - days: [Mon, Tue, Wed, Thu, Fri]
        start: "09:00"
        end: "17:00"
        replicas: 10
test_sequence:
  - action: Set time to Monday 08:30 UTC
    expected:
      deployment_replicas: 2
      status.effectiveReplicas: 2
      status.currentWindow: "OffHours"
      condition.Ready: "True"
  - action: Advance time to Monday 09:00 UTC
    expected:
      deployment_replicas: 10
      status.effectiveReplicas: 10
      status.currentWindow: "09:00-17:00"
      events: ["ScaledUp from 2 to 10 replicas (entering window)"]
  - action: Advance time to Monday 17:00 UTC
    expected:
      deployment_replicas: 2
      status.effectiveReplicas: 2
      status.currentWindow: "OffHours"
      events: ["ScaledDown from 10 to 2 replicas (exiting window)"]
assertions:
  - Deployment scaled correctly at boundaries
  - Status reflects current state
  - Events emitted for scale operations
```

### ET-002: Grace Period Prevents Immediate Downscale
```yaml
test_id: ET-002
category: envtest
description: Grace period delays downscaling
setup:
  - Create namespace "test-et002"
  - Create Deployment "webapp" with replicas=10
  - Create TimeWindowScaler with gracePeriodSeconds=300
  tws_spec:
    timezone: "UTC"
    defaultReplicas: 2
    gracePeriodSeconds: 300
    windows:
      - days: [Mon]
        start: "09:00"
        end: "17:00"
        replicas: 10
test_sequence:
  - action: Set time to Monday 16:59 UTC (in window)
    expected:
      deployment_replicas: 10
      status.effectiveReplicas: 10
  - action: Advance time to Monday 17:00 UTC (exit window)
    expected:
      deployment_replicas: 10  # Grace period active
      status.effectiveReplicas: 10
      status.gracePeriodExpiry: "2025-03-10T17:05:00Z"
      events: ["Grace period started, will scale down at 17:05:00"]
  - action: Advance time to Monday 17:04 UTC
    expected:
      deployment_replicas: 10  # Still in grace
      status.effectiveReplicas: 10
  - action: Advance time to Monday 17:05:01 UTC
    expected:
      deployment_replicas: 2  # Grace expired
      status.effectiveReplicas: 2
      status.gracePeriodExpiry: null
      events: ["ScaledDown from 10 to 2 replicas (grace period expired)"]
assertions:
  - Grace period delays downscale by exact duration
  - Status tracks grace expiry time
  - Scale happens immediately after grace expires
```

### ET-003: Pause Mode Prevents Scaling
```yaml
test_id: ET-003
category: envtest
description: Pause flag prevents target modifications
setup:
  - Create namespace "test-et003"
  - Create Deployment "webapp" with replicas=5
  - Create TimeWindowScaler with pause=true
  tws_spec:
    timezone: "UTC"
    defaultReplicas: 2
    pause: true
    windows:
      - days: [Mon]
        start: "09:00"
        end: "17:00"
        replicas: 10
test_sequence:
  - action: Set time to Monday 08:30 UTC
    expected:
      deployment_replicas: 5  # Unchanged
      status.effectiveReplicas: 2
      status.targetObservedReplicas: 5
      condition.Ready: "False"
      condition.Ready.reason: "Paused"
      events: ["ScalingSkipped - controller is paused"]
  - action: Advance time to Monday 09:00 UTC
    expected:
      deployment_replicas: 5  # Still unchanged
      status.effectiveReplicas: 10
      events: ["ScalingSkipped - controller is paused"]
  - action: Update TWS set pause=false
    expected:
      deployment_replicas: 10  # Now scales
      status.effectiveReplicas: 10
      condition.Ready: "True"
      events: ["ScaledUp from 5 to 10 replicas (controller resumed)"]
assertions:
  - Pause prevents all scaling operations
  - Status shows desired state even when paused
  - Ready condition reflects pause state
  - Resuming immediately applies pending scale
```

### ET-004: Missing Target Deployment
```yaml
test_id: ET-004
category: envtest
description: Handle missing target gracefully
setup:
  - Create namespace "test-et004"
  - Create TimeWindowScaler targeting non-existent "webapp"
  tws_spec:
    targetRef:
      name: webapp  # Does not exist
    timezone: "UTC"
    defaultReplicas: 5
test_sequence:
  - action: Trigger reconciliation
    expected:
      condition.Ready: "False"
      condition.Ready.reason: "TargetNotFound"
      condition.Degraded: "True"
      condition.Degraded.message: "Deployment webapp not found"
      events: ["TargetNotFound - Deployment webapp not found in namespace test-et004"]
  - action: Create Deployment "webapp" with replicas=1
    expected:
      deployment_replicas: 5
      condition.Ready: "True"
      condition.Degraded: "False"
      events: ["ScaledUp from 1 to 5 replicas (target now available)"]
assertions:
  - Missing target sets Degraded condition
  - Controller retries and recovers when target appears
  - Clear error messages in conditions and events
```

### ET-005: Manual Drift Correction
```yaml
test_id: ET-005
category: envtest
description: Detect and correct manual changes to deployment
setup:
  - Create namespace "test-et005"
  - Create Deployment "webapp" with replicas=10
  - Create TimeWindowScaler
  tws_spec:
    timezone: "UTC"
    defaultReplicas: 10
test_sequence:
  - action: Manually patch deployment to replicas=20
    expected_after_reconcile:
      deployment_replicas: 10
      status.effectiveReplicas: 10
      events: ["DriftCorrected - Reset replicas from 20 to 10 (manual change detected)"]
  - action: Set pause=true, then manually patch to replicas=15
    expected_after_reconcile:
      deployment_replicas: 15  # Not corrected when paused
      status.targetObservedReplicas: 15
      condition.Ready: "False"
      condition.Ready.reason: "Paused"
      events: ["DriftDetected - Target has 15 replicas, expected 10 (correction skipped - paused)"]
assertions:
  - Manual changes detected and corrected when not paused
  - Drift observed but not corrected when paused
  - Events clearly indicate drift and action taken
```

### ET-006: Holiday Mode - Treat As Closed
```yaml
test_id: ET-006
category: envtest
description: Holiday mode overrides window schedule
setup:
  - Create namespace "test-et006"
  - Create ConfigMap "holidays" with data: {"2025-03-10": "Company Day"}
  - Create Deployment "webapp"
  - Create TimeWindowScaler with holiday configuration
  tws_spec:
    timezone: "UTC"
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
test_sequence:
  - action: Set time to Monday March 10, 2025 10:00 UTC (holiday)
    expected:
      deployment_replicas: 1  # Uses default despite being in window
      status.effectiveReplicas: 1
      status.currentWindow: "Holiday-Closed"
      events: ["WindowOverride - Holiday mode (treat-as-closed) using default replicas"]
assertions:
  - Holiday mode overrides normal window
  - Status indicates holiday state
  - Event explains override reason
```

### ET-007: Holiday Mode - Treat As Open
```yaml
test_id: ET-007
category: envtest
description: Holiday uses maximum capacity
setup:
  - Create namespace "test-et007"
  - Create ConfigMap "holidays" with data: {"2025-12-25": "Christmas"}
  - Create Deployment "webapp"
  - Create TimeWindowScaler
  tws_spec:
    timezone: "UTC"
    defaultReplicas: 2
    windows:
      - days: [Mon, Tue, Wed, Thu, Fri]
        start: "08:00"
        end: "12:00"
        replicas: 5
      - days: [Mon, Tue, Wed, Thu, Fri]
        start: "12:00"
        end: "20:00"
        replicas: 15
    holidays:
      mode: treat-as-open
      sourceRef:
        name: holidays
test_sequence:
  - action: Set time to Thursday Dec 25, 2025 03:00 UTC (holiday, off-hours)
    expected:
      deployment_replicas: 15  # Max of all windows
      status.effectiveReplicas: 15
      status.currentWindow: "Holiday-Open"
      events: ["WindowOverride - Holiday mode (treat-as-open) using max replicas"]
assertions:
  - Holiday open mode uses maximum replicas from all windows
  - Applied even during off-hours
```

### ET-008: Cross-Midnight Window
```yaml
test_id: ET-008
category: envtest
description: Handle window that spans midnight
setup:
  - Create namespace "test-et008"
  - Create Deployment "webapp"
  - Create TimeWindowScaler with night shift window
  tws_spec:
    timezone: "UTC"
    defaultReplicas: 2
    windows:
      - days: [Fri]
        start: "22:00"
        end: "06:00"
        replicas: 8
test_sequence:
  - action: Set time to Friday 21:30 UTC
    expected:
      deployment_replicas: 2
      status.currentWindow: "OffHours"
  - action: Advance to Friday 22:00 UTC
    expected:
      deployment_replicas: 8
      status.currentWindow: "22:00-06:00"
  - action: Advance to Saturday 01:00 UTC
    expected:
      deployment_replicas: 8  # Still in window
      status.currentWindow: "22:00-06:00"
  - action: Advance to Saturday 06:00 UTC
    expected:
      deployment_replicas: 2
      status.currentWindow: "OffHours"
assertions:
  - Cross-midnight window activates on listed day
  - Remains active into next day
  - Deactivates at end time on next day
```

### ET-009: Overlapping Windows - Precedence
```yaml
test_id: ET-009
category: envtest
description: Last matching window wins
setup:
  - Create namespace "test-et009"
  - Create Deployment "webapp"
  - Create TimeWindowScaler with overlapping windows
  tws_spec:
    timezone: "UTC"
    defaultReplicas: 1
    windows:
      - days: [Mon]
        start: "08:00"
        end: "18:00"
        replicas: 5
      - days: [Mon]
        start: "12:00"
        end: "14:00"
        replicas: 20  # Lunch rush override
test_sequence:
  - action: Set time to Monday 11:00 UTC
    expected:
      deployment_replicas: 5
      status.currentWindow: "08:00-18:00"
  - action: Advance to Monday 12:00 UTC
    expected:
      deployment_replicas: 20  # Second window takes precedence
      status.currentWindow: "12:00-14:00"
  - action: Advance to Monday 14:00 UTC
    expected:
      deployment_replicas: 5  # Back to first window
      status.currentWindow: "08:00-18:00"
assertions:
  - Later windows in array override earlier ones
  - Window transitions tracked in status
```

### ET-010: Status Conditions Full Lifecycle
```yaml
test_id: ET-010
category: envtest
description: Validate all condition transitions
setup:
  - Create namespace "test-et010"
  - Create TimeWindowScaler (deployment doesn't exist yet)
test_sequence:
  - action: Initial creation
    expected:
      condition.Ready: "False"
      condition.Ready.reason: "TargetNotFound"
      condition.Reconciling: "True"
      condition.Degraded: "True"
  - action: Create target Deployment
    expected:
      condition.Ready: "True"
      condition.Ready.reason: "Reconciled"
      condition.Reconciling: "False"
      condition.Degraded: "False"
  - action: Delete holiday ConfigMap (if configured)
    expected:
      condition.Ready: "True"  # Still ready
      condition.Degraded: "True"
      condition.Degraded.reason: "HolidaySourceMissing"
  - action: Set pause=true
    expected:
      condition.Ready: "False"
      condition.Ready.reason: "Paused"
assertions:
  - Conditions accurately reflect system state
  - Multiple conditions can be true simultaneously
  - Reasons are descriptive and actionable
```

### ET-011: Generation Tracking
```yaml
test_id: ET-011
category: envtest
description: Ensure status.observedGeneration tracks spec changes
setup:
  - Create namespace "test-et011"
  - Create Deployment and TimeWindowScaler
initial_state:
  metadata.generation: 1
  status.observedGeneration: 0
test_sequence:
  - action: First reconciliation
    expected:
      status.observedGeneration: 1
  - action: Update spec.defaultReplicas from 2 to 5
    expected:
      metadata.generation: 2
      status.observedGeneration: 2
  - action: Update spec adding new window
    expected:
      metadata.generation: 3
      status.observedGeneration: 3
assertions:
  - observedGeneration updated on each reconciliation
  - Matches metadata.generation after successful reconcile
```

### ET-012: Timezone DST Transitions
```yaml
test_id: ET-012
category: envtest
description: Handle daylight saving time changes correctly
setup:
  - Create namespace "test-et012"
  - Create Deployment and TimeWindowScaler
  tws_spec:
    timezone: "America/New_York"
    defaultReplicas: 2
    windows:
      - days: [Sun]
        start: "01:00"
        end: "04:00"
        replicas: 10
test_sequence:
  - action: Set time to March 9, 2025 01:30 EST (before DST)
    expected:
      deployment_replicas: 10
      status.currentWindow: "01:00-04:00"
  - action: Advance to March 9, 2025 03:30 EDT (after spring forward)
    expected:
      deployment_replicas: 10  # Still in window
  - action: Advance to March 9, 2025 04:00 EDT
    expected:
      deployment_replicas: 2  # Window ended
assertions:
  - Window duration reduced by 1 hour during spring forward
  - Controller handles missing hour (2:00-3:00) correctly
```

### ET-013: Webhook Validation
```yaml
test_id: ET-013
category: envtest
description: Webhook rejects invalid configurations
setup:
  - Create namespace "test-et013"
test_sequence:
  - action: Create TWS with start="25:00"
    expected:
      error: "validation webhook: start time must match HH:MM format"
  - action: Create TWS with start="14:00", end="14:00"
    expected:
      error: "validation webhook: start and end cannot be equal"
  - action: Create TWS with invalid timezone="Invalid/TZ"
    expected:
      error: "validation webhook: unknown time zone Invalid/TZ"
  - action: Create TWS with negative replicas=-1
    expected:
      error: "validation webhook: replicas must be >= 0"
  - action: Create TWS with empty days array
    expected:
      error: "validation webhook: days cannot be empty"
assertions:
  - Webhook prevents invalid resources from being created
  - Error messages are specific and actionable
```

### ET-014: Finalizer Cleanup
```yaml
test_id: ET-014
category: envtest
description: Ensure clean deletion with finalizer
setup:
  - Create namespace "test-et014"
  - Create Deployment "webapp"
  - Create TimeWindowScaler with finalizer
test_sequence:
  - action: Check TWS has finalizer
    expected:
      metadata.finalizers: ["kyklos.io/finalizer"]
  - action: Delete TWS
    expected:
      events: ["Deleting - Removing TimeWindowScaler"]
  - action: Verify cleanup
    expected:
      tws_exists: false
      deployment_replicas: 10  # Unchanged after TWS deletion
assertions:
  - Finalizer added on creation
  - Cleanup happens before deletion
  - Target deployment not affected by TWS deletion
```

### ET-015: Requeue Timing
```yaml
test_id: ET-015
category: envtest
description: Verify requeue at correct times
setup:
  - Create namespace "test-et015"
  - Create Deployment and TimeWindowScaler
  tws_spec:
    timezone: "UTC"
    defaultReplicas: 2
    windows:
      - days: [Mon]
        start: "09:00"
        end: "17:00"
        replicas: 10
test_sequence:
  - action: Set time to Monday 08:00 UTC
    expected:
      requeue_after: ~3600s  # About 1 hour to 09:00
      requeue_reason: "Next window starts at 09:00"
  - action: Set time to Monday 09:30 UTC
    expected:
      requeue_after: ~27000s  # About 7.5 hours to 17:00
      requeue_reason: "Current window ends at 17:00"
  - action: Set time to Monday 18:00 UTC
    expected:
      requeue_after: ~54000s  # About 15 hours to next 09:00
      requeue_reason: "Next window starts tomorrow"
assertions:
  - Requeue duration matches time to next boundary
  - Includes small jitter (5-25 seconds)
  - Maximum requeue is 24 hours
```

## Test Utilities

### Envtest Helper Functions
```go
// Wait for condition with timeout
func WaitForCondition(ctx context.Context, client client.Client, tws *v1alpha1.TimeWindowScaler, condType string, status metav1.ConditionStatus) error

// Get deployment replicas
func GetDeploymentReplicas(ctx context.Context, client client.Client, name, namespace string) (int32, error)

// Advance fake clock and trigger reconcile
func AdvanceTimeAndReconcile(ctx context.Context, clock *clocktesting.FakeClock, reconciler *Reconciler, duration time.Duration) error

// Assert event was recorded
func AssertEvent(t *testing.T, recorder *record.FakeRecorder, expected string)
```

### Test Fixtures
```go
// Standard TWS for testing
func NewTestTWS(name, namespace string) *v1alpha1.TimeWindowScaler

// Standard Deployment
func NewTestDeployment(name, namespace string, replicas int32) *appsv1.Deployment

// Holiday ConfigMap
func NewHolidayConfigMap(name, namespace string, holidays map[string]string) *corev1.ConfigMap
```

## Execution Strategy

### Test Organization
```
controllers/
├── timewindowscaler_controller_test.go
├── suite_test.go
└── testdata/
    ├── happy_path_test.go
    ├── error_cases_test.go
    ├── pause_test.go
    ├── holiday_test.go
    ├── grace_test.go
    └── status_test.go
```

### Test Isolation
- Each test uses unique namespace
- Resources cleaned up in AfterEach
- No shared state between tests
- Fake clock reset for each test

### Parallel Execution
```go
var _ = Describe("TimeWindowScaler Controller", func() {
    Context("Happy Path", func() {
        It("scales based on time windows", func() {
            // Can run in parallel
        })
    })
})
```

## Coverage Requirements

### Controller Functions
- `Reconcile()`: 95% coverage
- `SetupWithManager()`: 90% coverage
- `determineEffectiveReplicas()`: 100% coverage
- `updateTargetDeployment()`: 95% coverage
- `updateStatus()`: 90% coverage
- `calculateNextRequeue()`: 95% coverage

### Edge Cases Must Test
- Empty window arrays
- Nil/missing optional fields
- Conflicting configurations
- Rapid time changes
- Concurrent updates
- Webhook rejection paths

## Success Criteria

1. All scenarios pass consistently
2. No use of real time.Now()
3. Each test completes in <1 second
4. Zero flakes in 1000 runs
5. Clear logs showing state transitions
6. Descriptive failure messages