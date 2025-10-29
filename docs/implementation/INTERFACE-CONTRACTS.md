# Kyklos Interface Contracts

**Purpose:** Define public function signatures and behavioral contracts for all major modules without implementation details.

**Last Updated:** 2025-10-29

## Overview

This document specifies the input/output contracts for each major module in Kyklos. These contracts serve as the integration points between modules and define testable boundaries.

## Module: Time Calculation Engine (`/internal/timecalc`)

### Purpose
Pure time mathematics and window evaluation logic with zero Kubernetes dependencies.

### Design Principles
- All functions accept explicit `time.Time` parameters (never call `time.Now()`)
- Return values only (no side effects, no logging)
- 100% deterministic for same inputs
- No global state or caching

---

### Function: ComputeEffectiveReplicas

**Purpose:** Determine the desired replica count at a specific point in time.

**Signature:**
```go
func ComputeEffectiveReplicas(
    windows []TimeWindow,
    defaultReplicas int32,
    localTime time.Time,
) int32
```

**Inputs:**
- `windows`: Array of time windows from spec (days, start, end, replicas)
- `defaultReplicas`: Fallback replica count when no windows match
- `localTime`: Current time in the configured timezone (NOT UTC)

**Outputs:**
- `int32`: The effective replica count at the given local time

**Behavior:**
1. Evaluate each window in array order
2. For each window, check if `localTime` matches window days and time range
3. For cross-midnight windows (end < start), check both current day and previous day
4. Return replicas from the **last matching window** (precedence by position)
5. If no windows match, return `defaultReplicas`

**Guarantees:**
- Deterministic: same inputs always produce same output
- No side effects (no logging, no state mutation)
- Handles cross-midnight windows correctly

**Test Scenarios:**
- In-window: localTime=10:00, window=09:00-17:00 → window replicas
- Out-of-window: localTime=20:00, window=09:00-17:00 → defaultReplicas
- Cross-midnight (in): localTime=23:00, window=22:00-02:00 → window replicas
- Cross-midnight (after midnight): localTime=01:00, window=22:00-02:00 → window replicas
- Cross-midnight (out): localTime=03:00, window=22:00-02:00 → defaultReplicas
- Overlapping windows: two windows match → last window wins
- No windows defined: return defaultReplicas

---

### Function: ComputeNextBoundary

**Purpose:** Calculate when the next window state change will occur.

**Signature:**
```go
func ComputeNextBoundary(
    windows []TimeWindow,
    localTime time.Time,
) time.Time
```

**Inputs:**
- `windows`: Array of time windows from spec
- `localTime`: Current time in the configured timezone

**Outputs:**
- `time.Time`: Timestamp of the next window boundary (start or end)

**Behavior:**
1. Find all window boundaries (start and end times) for the next 24 hours
2. For cross-midnight windows, compute boundaries across day transitions
3. Return the earliest boundary timestamp that is strictly after `localTime`
4. If no boundary found in next 24 hours, return tomorrow at 00:00

**Guarantees:**
- Returned time is always > localTime (never equal)
- Handles cross-midnight boundaries correctly
- Returns time in same timezone as input
- Maximum return value: localTime + 24 hours

**Test Scenarios:**
- Currently in window (14:00, window 09:00-17:00) → 17:00 same day
- Currently out of window (20:00, window 09:00-17:00) → 09:00 next day
- Cross-midnight window (23:00, window 22:00-02:00) → 02:00 next day
- Multiple windows → earliest next boundary
- No windows → tomorrow 00:00

---

### Function: ApplyGracePeriod

**Purpose:** Apply grace period logic to prevent immediate scale-down.

**Signature:**
```go
func ApplyGracePeriod(
    previousReplicas int32,
    desiredReplicas int32,
    gracePeriodSeconds int32,
    gracePeriodExpiry *time.Time,
    now time.Time,
) (finalReplicas int32, newExpiry *time.Time)
```

**Inputs:**
- `previousReplicas`: Last known replica count (from status)
- `desiredReplicas`: Computed replica count from window matching
- `gracePeriodSeconds`: Grace period duration from spec (0 = disabled)
- `gracePeriodExpiry`: Current grace period expiry timestamp (nil if not in grace)
- `now`: Current time (for expiry comparison)

**Outputs:**
- `finalReplicas`: Replica count to apply after grace period logic
- `newExpiry`: Updated grace period expiry timestamp (nil if no grace active)

**Behavior:**
1. **If `desiredReplicas >= previousReplicas`:** (scale-up or no change)
   - Return `(desiredReplicas, nil)` immediately
   - Cancel any active grace period
2. **If `desiredReplicas < previousReplicas` AND `gracePeriodSeconds > 0`:** (scale-down with grace)
   - If `gracePeriodExpiry == nil`: grace period starting now
     - Return `(previousReplicas, now + gracePeriodSeconds)`
   - If `now < gracePeriodExpiry`: still in grace period
     - Return `(previousReplicas, gracePeriodExpiry)` (maintain current replicas)
   - If `now >= gracePeriodExpiry`: grace period expired
     - Return `(desiredReplicas, nil)` (apply scale-down)
3. **If `desiredReplicas < previousReplicas` AND `gracePeriodSeconds == 0`:** (immediate scale-down)
   - Return `(desiredReplicas, nil)`

**Guarantees:**
- Grace only applies to scale-down operations
- Grace period state persists across reconciliations via expiry timestamp
- Scale-up always cancels grace period immediately
- Deterministic: same inputs produce same outputs

**Test Scenarios:**
- Scale-up (3→10): immediate, no grace
- Scale-down first time (10→3, grace=300s): maintain 10, set expiry
- Scale-down within grace (10→3, expiry in future): maintain 10, keep expiry
- Scale-down grace expired (10→3, expiry in past): apply 3, clear expiry
- Scale-up during grace (3→10 while in grace): apply 10, cancel grace
- Grace disabled (grace=0): immediate scale-down

---

### Function: EvaluateHoliday

**Purpose:** Determine if current date is a holiday and compute override replicas.

**Signature:**
```go
func EvaluateHoliday(
    holidayDates map[string]bool,
    localDate string, // YYYY-MM-DD format
    mode string, // "ignore", "treat-as-closed", "treat-as-open"
    windows []TimeWindow,
    defaultReplicas int32,
) (isHoliday bool, overrideReplicas *int32)
```

**Inputs:**
- `holidayDates`: Map of ISO date strings (YYYY-MM-DD) → true
- `localDate`: Current date in YYYY-MM-DD format (from localTime)
- `mode`: Holiday mode from spec.holidays.mode
- `windows`: Array of time windows (for max calculation in treat-as-open)
- `defaultReplicas`: Fallback replica count

**Outputs:**
- `isHoliday`: True if localDate exists in holidayDates map
- `overrideReplicas`: Replica count to use if holiday (nil means use normal window matching)

**Behavior:**
1. **Check holiday:** `isHoliday = holidayDates[localDate]`
2. **If NOT holiday:** Return `(false, nil)` - use normal window matching
3. **If holiday:**
   - **Mode "ignore":** Return `(true, nil)` - holiday noted but use normal windows
   - **Mode "treat-as-closed":** Return `(true, &defaultReplicas)` - use defaultReplicas
   - **Mode "treat-as-open":** Return `(true, &maxReplicas)` where `maxReplicas = max(all window replicas)`

**Guarantees:**
- If `overrideReplicas != nil`, caller must use that value instead of window matching
- If `overrideReplicas == nil`, caller proceeds with normal window matching
- Empty `holidayDates` map means no holidays (always returns `(false, nil)`)

**Test Scenarios:**
- Not holiday: return (false, nil)
- Holiday + ignore mode: return (true, nil)
- Holiday + treat-as-closed: return (true, defaultReplicas)
- Holiday + treat-as-open: return (true, max(window replicas))
- Holiday + empty windows + treat-as-open: return (true, defaultReplicas)

---

## Module: Status Writer (`/internal/statuswriter`)

### Purpose
Atomic status subresource updates with optimistic locking and retry logic.

### Design Principles
- Single atomic update for all status fields
- Automatic retry on conflict (409)
- Condition timestamp management
- No business logic (pure update operations)

---

### Function: UpdateStatus

**Purpose:** Update TimeWindowScaler status subresource with all fields.

**Signature:**
```go
func UpdateStatus(
    ctx context.Context,
    client client.Client,
    tws *kyklosv1alpha1.TimeWindowScaler,
    effectiveReplicas int32,
    targetObservedReplicas int32,
    currentWindow string,
    lastScaleTime *metav1.Time,
    gracePeriodExpiry *metav1.Time,
    conditions []metav1.Condition,
    observedGeneration int64,
) error
```

**Inputs:**
- `ctx`: Context for cancellation
- `client`: Kubernetes client (controller-runtime)
- `tws`: TimeWindowScaler object to update (must have ResourceVersion)
- `effectiveReplicas`: Computed desired replica count
- `targetObservedReplicas`: Last observed replica count of target workload
- `currentWindow`: Label for active window or "OffHours"
- `lastScaleTime`: Timestamp of last scale operation (nil if never scaled)
- `gracePeriodExpiry`: Grace period expiry timestamp (nil if not in grace)
- `conditions`: Array of three conditions (Ready, Reconciling, Degraded)
- `observedGeneration`: spec.generation value that was processed

**Outputs:**
- `error`: nil on success, conflict error if optimistic lock failed, other errors for failures

**Behavior:**
1. Create status patch with all provided fields
2. Set `status.observedGeneration = observedGeneration`
3. For each condition:
   - Set `lastTransitionTime` to current time if status/reason changed
   - Preserve existing `lastTransitionTime` if unchanged
4. Apply patch to status subresource with optimistic locking
5. On conflict (409), return error for caller to retry
6. On success, update in-memory `tws.Status` with new values

**Guarantees:**
- All status fields updated atomically
- Condition timestamps managed correctly
- Optimistic locking enforced via ResourceVersion
- Idempotent: same values produce same result

**Error Handling:**
- Conflict (409): Return error, caller should refetch and retry
- Not Found (404): Return error, object deleted
- Other errors: Return error, caller handles

**Test Scenarios:**
- Successful update: no error, status fields set correctly
- Optimistic lock conflict: returns IsConflict(err) == true
- Condition unchanged: lastTransitionTime preserved
- Condition changed: lastTransitionTime updated to now

---

### Function: BuildConditions

**Purpose:** Build standard condition array from reconciliation state.

**Signature:**
```go
func BuildConditions(
    readyStatus metav1.ConditionStatus,
    readyReason string,
    readyMessage string,
    reconcilingStatus metav1.ConditionStatus,
    reconcilingReason string,
    reconcilingMessage string,
    degradedStatus metav1.ConditionStatus,
    degradedReason string,
    degradedMessage string,
    existingConditions []metav1.Condition,
) []metav1.Condition
```

**Inputs:**
- Three sets of (status, reason, message) for Ready, Reconciling, Degraded conditions
- `existingConditions`: Current conditions from status (for timestamp preservation)

**Outputs:**
- `[]metav1.Condition`: Array with exactly three conditions in order: Ready, Reconciling, Degraded

**Behavior:**
1. For each condition type (Ready, Reconciling, Degraded):
   - Find existing condition in `existingConditions` array
   - If status or reason changed: create new condition with current timestamp
   - If unchanged: preserve existing condition with original timestamp
2. Return array with exactly three conditions

**Guarantees:**
- Always returns exactly three conditions
- Timestamps only updated when state/reason changes
- Condition order: Ready, Reconciling, Degraded

**Test Scenarios:**
- No existing conditions: all get current timestamp
- Condition unchanged: timestamp preserved
- Status changed: new timestamp
- Reason changed: new timestamp

---

## Module: Event Recorder (`/internal/events`)

### Purpose
Kubernetes event emission with deduplication and rate limiting.

### Design Principles
- Deduplicate identical events within 5 minutes
- Rate limit: 20 events/minute per TWS
- Warning events always emitted
- Events include correlation fields

---

### Function: EmitScaleUp

**Purpose:** Emit event when target workload is scaled up.

**Signature:**
```go
func EmitScaleUp(
    recorder record.EventRecorder,
    tws *kyklosv1alpha1.TimeWindowScaler,
    fromReplicas int32,
    toReplicas int32,
    window string,
    reason string, // "WindowEntered", "ManualDriftCorrected", "HolidayOverride"
) error
```

**Inputs:**
- `recorder`: Kubernetes event recorder from controller-runtime
- `tws`: TimeWindowScaler object (event target)
- `fromReplicas`: Previous replica count
- `toReplicas`: New replica count (must be > fromReplicas)
- `window`: Active window name or "OffHours"
- `reason`: Classification of why scale-up occurred

**Outputs:**
- `error`: nil on success, error if emission failed

**Behavior:**
1. Check deduplication cache: if same event emitted within 5 minutes, skip
2. Format message: `"Scaled up from {from} to {to} replicas (window: {window})"`
3. If reason == "ManualDriftCorrected", prepend to message
4. Emit Normal event with reason="ScaledUp"
5. Add to deduplication cache with current timestamp
6. Update rate limit counter

**Guarantees:**
- Deduplicated within 5-minute window
- Event includes all context for debugging
- Rate limited to prevent spam

**Test Scenarios:**
- First scale-up: event emitted
- Duplicate within 5 min: event skipped
- Duplicate after 5 min: event emitted
- Manual drift corrected: message includes "Corrected manual drift"

---

### Function: EmitScaleDown

**Purpose:** Emit event when target workload is scaled down.

**Signature:**
```go
func EmitScaleDown(
    recorder record.EventRecorder,
    tws *kyklosv1alpha1.TimeWindowScaler,
    fromReplicas int32,
    toReplicas int32,
    window string,
    gracePeriodApplied bool,
    gracePeriodSeconds int32,
    reason string, // "WindowExited", "ManualDriftCorrected", "MaintenanceWindow"
) error
```

**Inputs:**
- `recorder`: Kubernetes event recorder
- `tws`: TimeWindowScaler object
- `fromReplicas`: Previous replica count
- `toReplicas`: New replica count (must be < fromReplicas)
- `window`: Active window name or "OffHours"
- `gracePeriodApplied`: True if grace period was applied
- `gracePeriodSeconds`: Grace period duration (if applied)
- `reason`: Classification of scale-down

**Outputs:**
- `error`: nil on success, error if emission failed

**Behavior:**
1. Check deduplication cache
2. Format message: `"Scaled down from {from} to {to} replicas (window: {window})"`
3. If `gracePeriodApplied`, append `" after {duration}s grace period"`
4. If reason == "ManualDriftCorrected", prepend to message
5. Emit Normal event with reason="ScaledDown"
6. Add to deduplication cache

**Test Scenarios:**
- Scale-down without grace: event with basic message
- Scale-down after grace: event includes "after 300s grace period"
- Duplicate event: skipped

---

### Function: EmitScalingSkipped

**Purpose:** Emit event when scaling is computed but not applied.

**Signature:**
```go
func EmitScalingSkipped(
    recorder record.EventRecorder,
    tws *kyklosv1alpha1.TimeWindowScaler,
    currentReplicas int32,
    desiredReplicas int32,
    reason string, // "Paused", "TargetNotFound", "UpdateFailed"
) error
```

**Inputs:**
- `recorder`: Kubernetes event recorder
- `tws`: TimeWindowScaler object
- `currentReplicas`: Current target replica count
- `desiredReplicas`: Computed desired replicas
- `reason`: Why scaling was skipped

**Outputs:**
- `error`: nil on success, error if emission failed

**Behavior:**
1. Format message: `"Scaling skipped due to {reason}: current={current}, desired={desired}"`
2. Emit Normal event with reason="ScalingSkipped"
3. No deduplication (always emit, user should know about skips)

**Test Scenarios:**
- Paused: event with "Scaling skipped due to pause"
- Target not found: event with "TargetNotFound"

---

## Module: Metrics Recorder (`/internal/metrics`)

### Purpose
Prometheus metrics recording for observability.

### Design Principles
- Metrics registered in init()
- Labels: tws_name, namespace, window, state
- No business logic (pure instrumentation)

---

### Function: RecordScaleEvent

**Purpose:** Record scale operation in metrics.

**Signature:**
```go
func RecordScaleEvent(
    twsName string,
    namespace string,
    direction string, // "up" or "down"
    fromReplicas int32,
    toReplicas int32,
)
```

**Inputs:**
- `twsName`: TimeWindowScaler name
- `namespace`: TimeWindowScaler namespace
- `direction`: "up" or "down"
- `fromReplicas`: Previous replica count
- `toReplicas`: New replica count

**Outputs:**
- None (void function)

**Behavior:**
1. Increment `kyklos_scale_events_total` counter with labels
2. Labels: `{tws_name="...", namespace="...", direction="up|down"}`

**Guarantees:**
- Counter increments are atomic
- Never panics (metrics library handles errors)

---

### Function: RecordStateChange

**Purpose:** Update current state gauge.

**Signature:**
```go
func RecordStateChange(
    twsName string,
    namespace string,
    window string,
    effectiveReplicas int32,
    inGracePeriod bool,
)
```

**Inputs:**
- `twsName`: TimeWindowScaler name
- `namespace`: TimeWindowScaler namespace
- `window`: Current window label
- `effectiveReplicas`: Current desired replicas
- `inGracePeriod`: True if in grace period

**Outputs:**
- None (void function)

**Behavior:**
1. Set `kyklos_effective_replicas` gauge to `effectiveReplicas` with labels
2. Set `kyklos_in_grace_period` gauge to 1 (if true) or 0 (if false) with labels
3. Labels: `{tws_name="...", namespace="...", window="..."}`

**Guarantees:**
- Gauge values are current state
- Never panics

---

### Function: RecordReconcileDuration

**Purpose:** Record reconciliation loop duration.

**Signature:**
```go
func RecordReconcileDuration(
    twsName string,
    namespace string,
    durationSeconds float64,
)
```

**Inputs:**
- `twsName`: TimeWindowScaler name
- `namespace`: TimeWindowScaler namespace
- `durationSeconds`: Duration of reconciliation in seconds

**Outputs:**
- None (void function)

**Behavior:**
1. Observe value in `kyklos_reconcile_duration_seconds` histogram with labels
2. Labels: `{tws_name="...", namespace="..."}`

**Guarantees:**
- Histogram buckets: [0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0]
- Never panics

---

## Module: Reconciler (`/controllers`)

### Purpose
Main reconciliation loop orchestrating all modules.

### Design Principles
- Idempotent operations only
- No inline time calculations
- Delegate to specialized modules
- Status updates before return

---

### Function: Reconcile

**Purpose:** Main controller reconciliation function.

**Signature:**
```go
func (r *TimeWindowScalerReconciler) Reconcile(
    ctx context.Context,
    req ctrl.Request,
) (ctrl.Result, error)
```

**Inputs:**
- `ctx`: Context for cancellation and timeout
- `req`: Request with NamespacedName of TimeWindowScaler

**Outputs:**
- `ctrl.Result`: Requeue decision (RequeueAfter duration)
- `error`: Error if reconciliation failed

**Behavior:**
1. **Fetch TimeWindowScaler:** Get TWS from cache; if not found, return no error (deleted)
2. **Validate Timezone:** Load timezone; if invalid, set Degraded condition, return with 5-minute requeue
3. **Load Timezone:** Convert current UTC time to local time using spec.timezone
4. **Load Holiday ConfigMap:** If spec.holidays configured, get ConfigMap; cache dates; handle missing gracefully
5. **Evaluate Holiday:** Call `timecalc.EvaluateHoliday`; if override, skip window matching
6. **Compute Effective Replicas:** Call `timecalc.ComputeEffectiveReplicas` with localTime
7. **Apply Grace Period:** Call `timecalc.ApplyGracePeriod` with status.effectiveReplicas and status.gracePeriodExpiry
8. **Get Target Workload:** Fetch Deployment; if not found, set Ready=False, emit event, return with 30s requeue
9. **Determine Write Need:** Compare targetSpec.replicas with finalReplicas
10. **Handle Pause:** If spec.pause==true, skip step 11, emit ScalingSkipped event
11. **Update Target:** If write needed, PATCH Deployment.spec.replicas; on conflict, requeue immediately
12. **Build Conditions:** Create Ready, Reconciling, Degraded conditions based on reconciliation state
13. **Update Status:** Call `statuswriter.UpdateStatus` with all status fields; on conflict, requeue immediately
14. **Emit Events:** Call eventrecorder functions for scale operations
15. **Record Metrics:** Call metrics recorder functions
16. **Compute Requeue:** Call `timecalc.ComputeNextBoundary`; compute requeue duration with jitter; return

**Guarantees:**
- Idempotent: can be called multiple times safely
- Status always updated before return
- Requeue scheduled for next boundary
- No side effects on error (status may be partial)

**Error Handling:**
- Transient errors: return error, controller-runtime retries
- Semantic errors: set conditions, return success with fixed requeue
- Conflicts: requeue immediately

---

## Integration Contracts

### Controller → TimeCalc
```go
// Controller passes explicit time and spec fields
effectiveReplicas := timecalc.ComputeEffectiveReplicas(
    spec.Windows,
    spec.DefaultReplicas,
    localTime,
)
```

### Controller → StatusWriter
```go
// Controller passes all status fields, statuswriter handles update
err := statuswriter.UpdateStatus(
    ctx,
    r.Client,
    tws,
    effectiveReplicas,
    targetObservedReplicas,
    currentWindow,
    lastScaleTime,
    gracePeriodExpiry,
    conditions,
    tws.Generation,
)
```

### Controller → EventRecorder
```go
// Controller passes event data, eventrecorder handles deduplication
if toReplicas > fromReplicas {
    events.EmitScaleUp(r.Recorder, tws, fromReplicas, toReplicas, window, "WindowEntered")
}
```

### Controller → Metrics
```go
// Controller records metrics after operations
metrics.RecordScaleEvent(tws.Name, tws.Namespace, "up", fromReplicas, toReplicas)
metrics.RecordStateChange(tws.Name, tws.Namespace, window, effectiveReplicas, inGracePeriod)
```

---

## Testing Contracts

### Unit Tests (timecalc)
- Mock time using explicit `time.Time` parameters
- Table-driven tests with fixed dates
- 100% coverage for critical paths

### Integration Tests (controllers)
- Use controller-runtime envtest
- Create real Kubernetes objects
- Verify status and events

### E2E Tests
- Use kind/k3d cluster
- Fast-forward time with short windows
- Verify end-to-end scaling behavior

---

## Versioning and Stability

### Stable Contracts (Do Not Change)
- TimeCalc function signatures (public API)
- Status subresource fields (CRD API)
- Event reason strings (observability API)
- Metric names and labels (monitoring API)

### Evolvable Contracts (Can Add, Not Remove)
- New TimeCalc functions
- New status fields (backwards compatible)
- New event types
- New metrics

### Internal Contracts (Can Change)
- StatusWriter and EventRecorder implementation details
- Controller internal functions
- Test utilities

---

## Summary of Key Contracts

| Module | Primary Contract | Stability |
|--------|-----------------|-----------|
| timecalc.ComputeEffectiveReplicas | (windows, defaultReplicas, localTime) → replicas | Stable |
| timecalc.ComputeNextBoundary | (windows, localTime) → nextBoundary | Stable |
| timecalc.ApplyGracePeriod | (prev, desired, grace, expiry, now) → (final, newExpiry) | Stable |
| statuswriter.UpdateStatus | (ctx, client, tws, ...) → error | Evolvable |
| events.EmitScaleUp/Down | (recorder, tws, from, to, ...) → error | Evolvable |
| metrics.RecordScaleEvent | (name, ns, direction, ...) → void | Stable |
| Reconcile | (ctx, req) → (Result, error) | Framework (controller-runtime) |
