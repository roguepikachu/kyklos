# Status Conditions Design

## Condition Types

### Ready
Indicates whether the TimeWindowScaler has successfully applied the desired replica count to the target.

**True States**:
- `Reconciled`: Target matches desired state, no pending changes
- `GracePeriodActive`: Target maintaining higher replicas during grace period

**False States**:
- `TargetMismatch`: Target replica count differs from desired (pause or manual drift)
- `TargetNotFound`: Target Deployment does not exist
- `ConfigurationInvalid`: Spec validation failed

### Reconciling
Indicates whether the controller is actively processing changes.

**True States**:
- `ConfigurationChange`: Processing spec generation change
- `WindowTransition`: Transitioning between time windows
- `ScaleInProgress`: Actively updating target replicas
- `GracePeriodWaiting`: Waiting for grace period to expire

**False States**:
- `Stable`: No ongoing reconciliation, waiting for next boundary

### Degraded
Indicates operational problems that prevent normal function.

**True States**:
- `InvalidTimezone`: Timezone cannot be loaded from IANA database
- `HolidaySourceMissing`: Holiday ConfigMap not found but required
- `InvalidConfiguration`: Window validation failed (e.g., start==end)
- `TargetUpdateFailed`: Repeated failures updating target

**False States**:
- `OperationalNormal`: No degradation detected

## Reason Catalog

### Ready Condition Reasons

| Reason | Message Template | When Set |
|--------|-----------------|----------|
| `Reconciled` | "Target replicas match desired state ({replicas})" | Target.spec.replicas == effectiveReplicas |
| `GracePeriodActive` | "Maintaining {current} replicas during grace period (target: {desired})" | Grace period preventing downscale |
| `TargetMismatch` | "Target has {observed} replicas but desired is {effective} (pause={pauseState})" | Mismatch detected |
| `TargetNotFound` | "Deployment {namespace}/{name} not found" | GET returns 404 |
| `ConfigurationInvalid` | "Invalid configuration: {error}" | Validation failed |

### Reconciling Condition Reasons

| Reason | Message Template | When Set |
|--------|-----------------|----------|
| `ConfigurationChange` | "Processing configuration update (generation {gen})" | observedGen != metadata.gen |
| `WindowTransition` | "Transitioning from {oldWindow} to {newWindow}" | Window boundary crossed |
| `ScaleInProgress` | "Scaling from {current} to {desired} replicas" | PATCH in flight |
| `GracePeriodWaiting` | "Grace period expires at {time}" | Delaying downscale |
| `Stable` | "Waiting until {nextBoundary}" | No active reconciliation |

### Degraded Condition Reasons

| Reason | Message Template | When Set |
|--------|-----------------|----------|
| `InvalidTimezone` | "Cannot load timezone {tz}: {error}" | time.LoadLocation fails |
| `HolidaySourceMissing` | "ConfigMap {name} not found in namespace {ns}" | Holiday ConfigMap GET fails |
| `InvalidConfiguration` | "Window validation failed: {details}" | start==end or pattern invalid |
| `TargetUpdateFailed` | "Failed to update target after {retries} attempts: {error}" | Multiple PATCH failures |
| `OperationalNormal` | "No issues detected" | Everything working |

## State Transition Table

### Ready Transitions

| From State | To State | Trigger Event |
|------------|----------|---------------|
| nil | Reconciled | Initial successful reconcile |
| Reconciled | TargetMismatch | Manual scaling detected |
| Reconciled | GracePeriodActive | Window exit with grace period |
| TargetMismatch | Reconciled | Drift corrected or pause disabled |
| GracePeriodActive | Reconciled | Grace period expired and scale applied |
| Reconciled | TargetNotFound | Target deleted |
| TargetNotFound | Reconciled | Target recreated and scaled |
| * | ConfigurationInvalid | Validation fails |
| ConfigurationInvalid | Reconciled | Configuration fixed |

### Reconciling Transitions

| From State | To State | Trigger Event |
|------------|----------|---------------|
| nil | Stable | Initial state |
| Stable | ConfigurationChange | Spec updated |
| Stable | WindowTransition | Time boundary reached |
| ConfigurationChange | ScaleInProgress | Starting scale operation |
| WindowTransition | ScaleInProgress | Starting scale operation |
| WindowTransition | GracePeriodWaiting | Downscale with grace |
| ScaleInProgress | Stable | Scale completed |
| GracePeriodWaiting | ScaleInProgress | Grace expired |
| GracePeriodWaiting | Stable | Grace cancelled (scale up) |
| * | Stable | Reconcile completed |

### Degraded Transitions

| From State | To State | Trigger Event |
|------------|----------|---------------|
| nil | OperationalNormal | Initial state |
| OperationalNormal | InvalidTimezone | Timezone load fails |
| OperationalNormal | HolidaySourceMissing | ConfigMap not found |
| OperationalNormal | InvalidConfiguration | Validation fails |
| OperationalNormal | TargetUpdateFailed | Multiple PATCH failures |
| InvalidTimezone | OperationalNormal | Timezone fixed |
| HolidaySourceMissing | OperationalNormal | ConfigMap created or holidays disabled |
| InvalidConfiguration | OperationalNormal | Configuration corrected |
| TargetUpdateFailed | OperationalNormal | Successful update |

## Terminal vs Transient States

### Terminal States (Stable)
- Ready=True, Reconciling=False, Degraded=False
  - System at desired state, waiting for next event
- Ready=False, Reconciling=False, Degraded=True
  - Permanent error requiring user intervention

### Transient States (Active)
- Reconciling=True
  - Active processing, will transition when complete
- Ready=False with Degraded=False
  - Temporary mismatch, will self-correct

## Example Condition Sets

### Scenario 1: Normal Operation During Window
```yaml
conditions:
- type: Ready
  status: "True"
  reason: Reconciled
  message: "Target replicas match desired state (10)"
- type: Reconciling
  status: "False"
  reason: Stable
  message: "Waiting until 17:00 IST"
- type: Degraded
  status: "False"
  reason: OperationalNormal
  message: "No issues detected"
```

### Scenario 2: Grace Period After Window Exit
```yaml
conditions:
- type: Ready
  status: "True"
  reason: GracePeriodActive
  message: "Maintaining 10 replicas during grace period (target: 2)"
- type: Reconciling
  status: "True"
  reason: GracePeriodWaiting
  message: "Grace period expires at 17:05:00 IST"
- type: Degraded
  status: "False"
  reason: OperationalNormal
  message: "No issues detected"
```

### Scenario 3: Paused with Manual Drift
```yaml
conditions:
- type: Ready
  status: "False"
  reason: TargetMismatch
  message: "Target has 7 replicas but desired is 10 (pause=true)"
- type: Reconciling
  status: "False"
  reason: Stable
  message: "Waiting until 17:00 IST"
- type: Degraded
  status: "False"
  reason: OperationalNormal
  message: "No issues detected"
```

### Scenario 4: Invalid Timezone
```yaml
conditions:
- type: Ready
  status: "False"
  reason: ConfigurationInvalid
  message: "Invalid configuration: timezone Mars/Olympus_Mons not found"
- type: Reconciling
  status: "False"
  reason: Stable
  message: "Waiting for valid configuration"
- type: Degraded
  status: "True"
  reason: InvalidTimezone
  message: "Cannot load timezone Mars/Olympus_Mons: unknown time zone"
```

### Scenario 5: Holiday with Closed Mode
```yaml
conditions:
- type: Ready
  status: "True"
  reason: Reconciled
  message: "Target replicas match desired state (0)"
- type: Reconciling
  status: "False"
  reason: Stable
  message: "Holiday mode active, next evaluation tomorrow 00:00 IST"
- type: Degraded
  status: "False"
  reason: OperationalNormal
  message: "No issues detected"
```

## Condition Update Rules

1. **Atomic Updates**: All three conditions updated together in single PATCH
2. **Timestamp Precision**: Use RFC3339 with timezone for lastTransitionTime
3. **Message Fields**: Always populate {replicas}, {window}, {time} placeholders
4. **Reason Stability**: Don't change reason without state change
5. **Degraded Priority**: Degraded=True overrides Ready state interpretation

## Observability Integration

### Metrics Labels from Conditions
- `ready_state`: true/false from Ready condition
- `ready_reason`: reason field from Ready condition
- `reconciling`: true/false from Reconciling condition
- `degraded`: true/false from Degraded condition
- `degraded_reason`: reason field when Degraded=True

### Alert Rules Based on Conditions
- **Critical**: Degraded=True for > 10 minutes
- **Warning**: Ready=False for > 30 minutes
- **Info**: Reconciling=True for > 5 minutes (possible stuck reconcile)