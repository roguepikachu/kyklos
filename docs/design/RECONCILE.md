# Reconcile Loop Design

## Purpose
The reconcile loop ensures the target workload's replica count matches the time-window-based desired state while maintaining idempotency, minimizing API writes, and handling time boundaries correctly.

## Inputs
- **Primary**: TimeWindowScaler resource (spec + metadata.generation)
- **Secondary**: Current UTC time, target Deployment status
- **Cached**: Holiday ConfigMap (if configured), timezone data

## Reconcile Steps

### Step 1: Validate Configuration
**Preconditions**: TimeWindowScaler object received from queue
**Actions**:
1. Parse and validate timezone string against IANA database
2. Validate all window time formats (HH:MM pattern)
3. Ensure no window has start==end
4. Verify targetRef.kind=="Deployment" (v1alpha1 constraint)

**Postconditions**: Valid configuration or return error with Degraded condition
**On Error**: Set Degraded=True, reason=InvalidConfiguration, requeue with exponential backoff

### Step 2: Load Timezone
**Preconditions**: Valid timezone string
**Actions**:
1. Load timezone using time.LoadLocation(spec.timezone)
2. Convert current UTC to local time in specified timezone

**Postconditions**: Local time available for window calculations
**On Error**: Set Degraded=True, reason=InvalidTimezone, use defaultReplicas, requeue in 5 minutes

### Step 3: Check Holiday Status (if configured)
**Preconditions**: spec.holidays configured
**Actions**:
1. Get ConfigMap from cache (name=spec.holidays.sourceRef.name)
2. Extract current date as YYYY-MM-DD in local timezone
3. Check if date exists as key in ConfigMap data

**Postconditions**: Holiday status determined (true/false) or error
**On Error**: Log warning, set Degraded=True, reason=HolidaySourceMissing, continue with holiday=false

### Step 4: Compute Effective Replicas
**Preconditions**: Local time available, holiday status determined
**Actions**:
1. If holiday && mode==treat-as-closed: return defaultReplicas
2. If holiday && mode==treat-as-open: return max(all window.replicas)
3. Otherwise, evaluate windows:
   - For each window in spec.windows array order:
     - Check if current day in window.days
     - For cross-midnight windows (end < start):
       - Check today: localTime >= start OR localTime < end
       - Check yesterday: was yesterday in days AND localTime < end
     - For normal windows: localTime >= start AND localTime < end
   - Keep last matching window
4. Return matching window replicas or defaultReplicas

**Postconditions**: effectiveReplicas determined
**Decision Table**:
| Holiday | Mode | In Window | Result |
|---------|------|-----------|--------|
| true | treat-as-closed | * | defaultReplicas |
| true | treat-as-open | * | max(window.replicas) |
| true | ignore | yes | window.replicas |
| true | ignore | no | defaultReplicas |
| false | * | yes | window.replicas |
| false | * | no | defaultReplicas |

### Step 5: Apply Grace Period Logic
**Preconditions**: effectiveReplicas computed, status.effectiveReplicas available
**Actions**:
1. If effectiveReplicas >= status.effectiveReplicas: no grace needed
2. If effectiveReplicas < status.effectiveReplicas && spec.gracePeriodSeconds > 0:
   - If !status.gracePeriodExpiry: set expiry = now + spec.gracePeriodSeconds
   - If now < status.gracePeriodExpiry: maintain previous replicas
   - If now >= status.gracePeriodExpiry: apply new replicas, clear expiry

**Postconditions**: Final effectiveReplicas determined with grace applied

### Step 6: Get Target Status
**Preconditions**: targetRef specified
**Actions**:
1. Get Deployment using cached client (namespace=targetRef.namespace || object.namespace)
2. Extract deployment.status.replicas
3. Extract deployment.spec.replicas

**Postconditions**: targetObservedReplicas and targetSpecReplicas known
**On Error**: Set Ready=False, reason=TargetNotFound, requeue in 30 seconds

### Step 7: Determine Write Need
**Preconditions**: effectiveReplicas computed, target status known
**Actions**:
1. **If spec.pause==true**:
   - Skip all writes to target workload
   - Continue computing effectiveReplicas normally (show what WOULD happen)
   - Update all status fields: effectiveReplicas, targetObservedReplicas, currentWindow, gracePeriodExpiry
   - Set Ready condition:
     - Ready=True if targetObservedReplicas == effectiveReplicas (aligned)
     - Ready=False with reason=TargetMismatch if different (drift while paused)
   - Emit ScalingSkipped event with message describing what would happen if not paused
   - **Return early, do not proceed to Step 8**
2. If targetSpecReplicas != effectiveReplicas: write needed
3. If manual drift detected (observedReplicas != targetSpecReplicas != effectiveReplicas): write needed

**Postconditions**: Write decision made or early return if paused

### Step 8: Update Target (if needed)
**Preconditions**: Write needed, pause==false
**Actions**:
1. Create patch: {"spec": {"replicas": effectiveReplicas}}
2. Apply patch to Deployment with optimistic locking
3. Record lastScaleTime in status

**Postconditions**: Target updated or conflict error
**On Conflict**: Requeue immediately (no backoff) for retry
**Idempotency**: Patch only if targetSpec != effective, making operation safe to retry

### Step 9: Compute Next Boundary
**Preconditions**: Current window state known
**Actions**:
1. Find all window boundaries for next 24 hours:
   - For each window, calculate next start and end in local time
   - Handle cross-midnight windows specially
2. Select earliest boundary > now
3. If no boundary in 24 hours: use tomorrow 00:00 local

**Postconditions**: nextBoundary timestamp determined
**Examples**:
- Currently in window (09:00-17:00), now 14:30: next = today 17:00
- Currently out of window, next window tomorrow 09:00: next = tomorrow 09:00
- Cross-midnight window (22:00-06:00), now 23:30: next = tomorrow 06:00

### Step 10: Update Status
**Preconditions**: All computations complete
**Actions**:
1. Update status fields:
   - currentWindow = label for active window or "OffHours"
   - effectiveReplicas = computed value
   - targetObservedReplicas = from Deployment
   - observedGeneration = metadata.generation
   - lastScaleTime = if scale occurred
2. Update conditions:
   - Ready: based on alignment and errors
   - Reconciling: based on ongoing changes
   - Degraded: based on errors encountered
3. Use PATCH with optimistic concurrency

**Postconditions**: Status reflects current state
**On Conflict**: Requeue for retry

### Step 11: Emit Events
**Preconditions**: State changes detected
**Actions**:
1. If scaled up: emit ScaledUp event
2. If scaled down: emit ScaledDown event
3. If pause prevented scaling: emit ScalingSkipped event
4. If window override due to holiday: emit WindowOverride event
5. If manual drift corrected: include in scale event reason

**Postconditions**: Events recorded for observability

### Step 12: Calculate Requeue
**Preconditions**: nextBoundary computed, no errors
**Actions**:
1. Base duration = nextBoundary - now
2. Add jitter: rand(5, 25) seconds
3. Quantize to 10-second boundaries
4. Minimum requeue: 30 seconds
5. Maximum requeue: 24 hours

**Postconditions**: Requeue scheduled
**Formula**: `requeueAfter = max(30s, min(24h, quantize(nextBoundary - now + jitter, 10s)))`

## Write Policy

### Create vs Update
- Always use PATCH for Deployment updates (never PUT)
- Use strategic merge patch for spec.replicas
- Include resourceVersion for optimistic concurrency

### Status Updates
- Use status subresource PATCH
- Include resourceVersion from last read
- Retry on conflict with fresh read

### Conflict Handling
- On 409 Conflict: requeue immediately
- On 429 RateLimit: exponential backoff
- On 500-503: exponential backoff with max 5 minutes

## Idempotency Guarantees

1. **Read-Compute-Write**: Always read current state before computing
2. **Conditional Writes**: Only write if change needed
3. **Status Convergence**: Status always reflects observed state
4. **Event Deduplication**: Same transition within 5 minutes suppressed
5. **Deterministic Compute**: Same inputs always produce same effectiveReplicas

## Manual Drift Correction

**Detection**: targetObservedReplicas != effectiveReplicas
**Correction**:
- If pause==false: update target on next reconcile
- If pause==true: observe drift, update status, skip correction
- Emit event describing drift and action taken

## Failure Modes

### Transient Errors (Retry)
- Network timeouts
- API server unavailable (503)
- Rate limiting (429)
- Optimistic lock conflicts (409)
- **Action**: Exponential backoff, max 5 minutes

### Semantic Errors (Report)
- Invalid timezone
- Target not found
- Invalid configuration
- Holiday ConfigMap missing
- **Action**: Set Degraded condition, use safe defaults, fixed requeue

## Requeue Strategy

### Normal Operation
- Requeue at next window boundary + jitter(5-25s)
- Minimum 30 seconds to prevent tight loops
- Maximum 24 hours for distant boundaries

### Error Conditions
- Transient errors: exponential backoff (30s, 1m, 2m, 5m)
- Semantic errors: fixed 5-minute interval
- Configuration changes: immediate (via watch)

### Jitter Application
```
jitter = rand(5, 25) // seconds
quantized = floor((baseDelay + jitter) / 10) * 10
final = max(30, quantized)
```

## Cross-Midnight Window Examples

### Friday 22:00 to 02:00, Current Time Scenarios
- Friday 21:00: Not in window (before start)
- Friday 23:00: In window (after start on listed day)
- Saturday 01:00: In window (before end on next day)
- Saturday 03:00: Not in window (after end)

### Calculation Logic
```
if window.end < window.start:
  // Cross-midnight
  if currentDay in window.days:
    inWindow = (localTime >= window.start OR localTime < window.end)
  else if yesterday in window.days:
    inWindow = (localTime < window.end)
else:
  // Normal window
  if currentDay in window.days:
    inWindow = (localTime >= window.start AND localTime < window.end)
```

## DST Transition Handling

### Spring Forward (2:00 → 3:00)
- Window 01:00-03:00: Duration reduced by 1 hour
- Window 23:00-05:00: Unaffected (still 6 hours wall clock)

### Fall Back (2:00 → 1:00)
- Window 01:00-03:00: Duration extended by 1 hour
- Window 23:00-05:00: Unaffected (still 6 hours wall clock)

**Implementation**: Use time.Location for all calculations, it handles DST automatically