# Kyklos Core Logic Pseudocode

**Purpose:** Readable pseudocode for critical algorithms without implementation details.

**Last Updated:** 2025-10-29

## Overview

This document provides human-readable pseudocode for the core time window logic. These algorithms are implemented in the `/internal/timecalc` package.

---

## Algorithm 1: ComputeEffectiveReplicas

**Purpose:** Determine the desired replica count at a specific moment in time.

**Inputs:**
- `windows`: Array of TimeWindow objects
- `defaultReplicas`: Fallback replica count (int32)
- `localTime`: Current time in configured timezone (Time)

**Output:**
- `effectiveReplicas`: int32

**Pseudocode:**

```
FUNCTION ComputeEffectiveReplicas(windows, defaultReplicas, localTime):
    // Extract current day and time components
    currentDay = DayOfWeek(localTime)  // e.g., "Mon", "Tue", ...
    currentTime = TimeOfDay(localTime) // e.g., HH:MM as comparable value
    currentDate = DateOnly(localTime)  // For day boundary checks

    // Track last matching window (precedence by position)
    lastMatchingWindow = NULL

    // Evaluate each window in order
    FOR EACH window IN windows:
        isMatch = FALSE

        // Check if current day is in window's days list
        IF currentDay IN window.days:
            // Handle cross-midnight windows
            IF window.end < window.start:
                // Window spans to next day
                // Example: 22:00-02:00
                // Matches if: time >= 22:00 OR time < 02:00
                IF currentTime >= window.start OR currentTime < window.end:
                    isMatch = TRUE
            ELSE:
                // Normal window within same day
                // Example: 09:00-17:00
                // Matches if: time >= 09:00 AND time < 17:00
                IF currentTime >= window.start AND currentTime < window.end:
                    isMatch = TRUE

        // Check if yesterday's cross-midnight window extends to today
        ELSE:
            yesterday = DayOfWeek(localTime - 24 hours)
            IF yesterday IN window.days:
                // Only matters if window crosses midnight
                IF window.end < window.start:
                    // Check if current time is before window end
                    // Example: Window is Friday 22:00-02:00
                    //          Saturday 01:00 matches (before 02:00)
                    IF currentTime < window.end:
                        isMatch = TRUE

        // Update last matching window if this one matches
        IF isMatch:
            lastMatchingWindow = window

    // Return replicas from last matching window or default
    IF lastMatchingWindow != NULL:
        RETURN lastMatchingWindow.replicas
    ELSE:
        RETURN defaultReplicas
END FUNCTION
```

**Example Execution:**

```
Input:
  windows = [
    {days: [Mon,Tue,Wed,Thu,Fri], start: 09:00, end: 17:00, replicas: 10},
    {days: [Sat,Sun], start: 00:00, end: 23:59, replicas: 2}
  ]
  defaultReplicas = 0
  localTime = "2025-01-27 14:30:00 IST" (Monday)

Execution:
  currentDay = "Mon"
  currentTime = 14:30

  Window 1:
    "Mon" IN [Mon,Tue,Wed,Thu,Fri] → YES
    end (17:00) >= start (09:00) → Normal window
    14:30 >= 09:00 AND 14:30 < 17:00 → TRUE
    lastMatchingWindow = Window 1

  Window 2:
    "Mon" IN [Sat,Sun] → NO
    yesterday = "Sun"
    "Sun" IN [Sat,Sun] → YES
    end (23:59) >= start (00:00) → Normal window (not cross-midnight)
    Skip (yesterday check only for cross-midnight)

  Result: lastMatchingWindow.replicas = 10
```

---

## Algorithm 2: ComputeNextBoundary

**Purpose:** Calculate when the next window state change will occur.

**Inputs:**
- `windows`: Array of TimeWindow objects
- `localTime`: Current time in configured timezone (Time)

**Output:**
- `nextBoundary`: Time (timestamp of next boundary)

**Pseudocode:**

```
FUNCTION ComputeNextBoundary(windows, localTime):
    // Collect all possible boundaries in next 24 hours
    boundaries = []

    currentDay = DayOfWeek(localTime)
    currentTime = TimeOfDay(localTime)
    currentDate = DateOnly(localTime)

    FOR EACH window IN windows:
        FOR EACH day IN window.days:
            // Calculate when this window starts on 'day'
            startBoundary = ComputeBoundaryTime(day, window.start, localTime)

            // Calculate when this window ends
            IF window.end < window.start:
                // Cross-midnight: end is on next calendar day
                nextDay = DayAfter(day)
                endBoundary = ComputeBoundaryTime(nextDay, window.end, localTime)
            ELSE:
                // Normal window: end is same day as start
                endBoundary = ComputeBoundaryTime(day, window.end, localTime)

            // Only include boundaries in the future
            IF startBoundary > localTime:
                ADD startBoundary TO boundaries
            IF endBoundary > localTime:
                ADD endBoundary TO boundaries

    // Find earliest boundary
    IF boundaries is not empty:
        RETURN MIN(boundaries)
    ELSE:
        // No boundaries in next 24 hours, default to tomorrow midnight
        tomorrowMidnight = localTime + 24 hours, rounded to 00:00
        RETURN tomorrowMidnight
END FUNCTION

HELPER FUNCTION ComputeBoundaryTime(targetDay, time, fromTime):
    // Convert day name and time to absolute timestamp
    // Accounts for wrapping to next week if needed

    currentDay = DayOfWeek(fromTime)
    daysUntilTarget = DaysUntil(currentDay, targetDay)

    // Build timestamp for target day at specified time
    targetDate = DateOnly(fromTime) + daysUntilTarget days
    boundaryTime = CombineDateTime(targetDate, time)

    RETURN boundaryTime
END HELPER

HELPER FUNCTION DaysUntil(fromDay, toDay):
    // Calculate days from fromDay to next occurrence of toDay
    // Mon=0, Tue=1, ..., Sun=6

    fromIndex = DayIndex(fromDay)
    toIndex = DayIndex(toDay)

    IF toIndex >= fromIndex:
        RETURN toIndex - fromIndex
    ELSE:
        // Wrap to next week
        RETURN 7 - fromIndex + toIndex
END HELPER
```

**Example Execution:**

```
Input:
  windows = [
    {days: [Mon,Tue,Wed,Thu,Fri], start: 09:00, end: 17:00, replicas: 10}
  ]
  localTime = "2025-01-27 14:30:00 IST" (Monday)

Execution:
  currentDay = "Mon"
  currentTime = 14:30

  Window 1:
    days = [Mon,Tue,Wed,Thu,Fri]

    For day=Mon:
      startBoundary = Mon 09:00 → 2025-01-27 09:00 (PAST, skip)
      endBoundary = Mon 17:00 → 2025-01-27 17:00 (FUTURE, add)

    For day=Tue:
      startBoundary = Tue 09:00 → 2025-01-28 09:00 (FUTURE, add)
      endBoundary = Tue 17:00 → 2025-01-28 17:00 (FUTURE, add)

    ... (similar for Wed, Thu, Fri)

  boundaries = [
    2025-01-27 17:00,  // Today end
    2025-01-28 09:00,  // Tomorrow start
    2025-01-28 17:00,  // Tomorrow end
    ...
  ]

  Result: MIN(boundaries) = 2025-01-27 17:00 (today at 17:00)
```

---

## Algorithm 3: ApplyGracePeriod

**Purpose:** Apply grace period logic to delay scale-down operations.

**Inputs:**
- `previousReplicas`: int32 (last known replica count)
- `desiredReplicas`: int32 (computed from window matching)
- `gracePeriodSeconds`: int32 (grace duration from spec)
- `gracePeriodExpiry`: *Time (current expiry timestamp, NULL if not in grace)
- `now`: Time (current timestamp)

**Outputs:**
- `finalReplicas`: int32 (replica count to apply)
- `newExpiry`: *Time (updated expiry timestamp)

**Pseudocode:**

```
FUNCTION ApplyGracePeriod(previousReplicas, desiredReplicas, gracePeriodSeconds, gracePeriodExpiry, now):

    // CASE 1: Scale-up or no change → No grace needed
    IF desiredReplicas >= previousReplicas:
        // Cancel any active grace period
        RETURN (desiredReplicas, NULL)

    // CASE 2: Scale-down detected
    ELSE:
        // CASE 2a: Grace period disabled
        IF gracePeriodSeconds == 0:
            // Immediate scale-down
            RETURN (desiredReplicas, NULL)

        // CASE 2b: Grace period enabled
        ELSE:
            // CASE 2b-i: Grace period starting now
            IF gracePeriodExpiry == NULL:
                // First time detecting scale-down
                // Start grace period, maintain previous replicas
                newExpiry = now + gracePeriodSeconds seconds
                RETURN (previousReplicas, newExpiry)

            // CASE 2b-ii: Already in grace period
            ELSE:
                // Check if grace period has expired
                IF now >= gracePeriodExpiry:
                    // Grace period expired, apply scale-down
                    RETURN (desiredReplicas, NULL)
                ELSE:
                    // Still within grace period, maintain previous replicas
                    RETURN (previousReplicas, gracePeriodExpiry)
END FUNCTION
```

**State Diagram:**

```
                    desiredReplicas >= previousReplicas
    ┌───────────────────────────────────────────────────┐
    │                                                   │
    │  Apply desiredReplicas immediately                │
    │  newExpiry = NULL                                 │
    │                                                   │
    └───────────────────────────────────────────────────┘
                              │
                              │
                              ▼
    ┌───────────────────────────────────────────────────┐
    │                                                   │
    │  desiredReplicas < previousReplicas               │
    │  (Scale-down detected)                            │
    │                                                   │
    └───────────────────────────────────────────────────┘
                              │
                              │
                ┌─────────────┴─────────────┐
                │                           │
                ▼                           ▼
    ┌───────────────────────┐   ┌───────────────────────┐
    │ gracePeriodSeconds=0  │   │ gracePeriodSeconds>0  │
    │                       │   │                       │
    │ Apply desiredReplicas │   │ Check expiry          │
    │ newExpiry = NULL      │   │                       │
    └───────────────────────┘   └───────────────────────┘
                                            │
                                            │
                        ┌───────────────────┴───────────────────┐
                        │                                       │
                        ▼                                       ▼
            ┌───────────────────────┐           ┌───────────────────────┐
            │ gracePeriodExpiry=NULL│           │ gracePeriodExpiry set │
            │                       │           │                       │
            │ Start grace period    │           │ Check time            │
            │ finalReplicas=previous│           │                       │
            │ newExpiry=now+grace   │           │                       │
            └───────────────────────┘           └───────────────────────┘
                                                            │
                                                            │
                                    ┌───────────────────────┴───────────────────┐
                                    │                                           │
                                    ▼                                           ▼
                        ┌───────────────────────┐               ┌───────────────────────┐
                        │ now >= expiry         │               │ now < expiry          │
                        │                       │               │                       │
                        │ Apply desiredReplicas │               │ Maintain previous     │
                        │ newExpiry = NULL      │               │ newExpiry = expiry    │
                        └───────────────────────┘               └───────────────────────┘
```

**Example Execution:**

```
Scenario: Scale-down with 300-second grace period

Initial State:
  previousReplicas = 10
  desiredReplicas = 10
  gracePeriodExpiry = NULL
  now = 17:00:00

Time 17:00:00 - Window exits, desiredReplicas becomes 2:
  Input: (10, 2, 300, NULL, 17:00:00)
  10 > 2 → Scale-down detected
  300 > 0 → Grace enabled
  expiry == NULL → Start grace
  Output: (10, 17:05:00)  // Maintain 10 replicas until 17:05

Time 17:02:30 - Reconcile again:
  Input: (10, 2, 300, 17:05:00, 17:02:30)
  10 > 2 → Scale-down detected
  300 > 0 → Grace enabled
  expiry != NULL → Check expiry
  17:02:30 < 17:05:00 → Still in grace
  Output: (10, 17:05:00)  // Still maintain 10 replicas

Time 17:05:15 - Reconcile after expiry:
  Input: (10, 2, 300, 17:05:00, 17:05:15)
  10 > 2 → Scale-down detected
  300 > 0 → Grace enabled
  expiry != NULL → Check expiry
  17:05:15 >= 17:05:00 → Grace expired
  Output: (2, NULL)  // Apply scale-down to 2 replicas
```

---

## Algorithm 4: EvaluateHoliday

**Purpose:** Determine if current date is a holiday and compute override replicas.

**Inputs:**
- `holidayDates`: Map[string]bool (ISO dates like "2025-12-25" → true)
- `localDate`: string (current date in YYYY-MM-DD format)
- `mode`: string ("ignore", "treat-as-closed", "treat-as-open")
- `windows`: Array of TimeWindow objects
- `defaultReplicas`: int32

**Outputs:**
- `isHoliday`: bool
- `overrideReplicas`: *int32 (NULL means use normal window matching)

**Pseudocode:**

```
FUNCTION EvaluateHoliday(holidayDates, localDate, mode, windows, defaultReplicas):

    // Check if current date is a holiday
    isHoliday = holidayDates[localDate]

    IF NOT isHoliday:
        // Not a holiday, use normal window matching
        RETURN (FALSE, NULL)

    // Current date is a holiday
    // Behavior depends on mode
    SWITCH mode:
        CASE "ignore":
            // Holiday detected but use normal windows
            RETURN (TRUE, NULL)

        CASE "treat-as-closed":
            // Use defaultReplicas (typically 0 or 2)
            RETURN (TRUE, &defaultReplicas)

        CASE "treat-as-open":
            // Use maximum configured replicas
            IF windows is empty:
                // No windows defined, fall back to defaultReplicas
                RETURN (TRUE, &defaultReplicas)
            ELSE:
                // Find max replicas across all windows
                maxReplicas = MAX(window.replicas FOR window IN windows)
                RETURN (TRUE, &maxReplicas)

        DEFAULT:
            // Unknown mode, treat as "ignore"
            RETURN (TRUE, NULL)
END FUNCTION
```

**Example Execution:**

```
Scenario 1: Holiday with treat-as-closed mode

Input:
  holidayDates = {"2025-12-25": true}
  localDate = "2025-12-25"
  mode = "treat-as-closed"
  windows = [{days:[Mon-Fri], start:09:00, end:17:00, replicas:10}]
  defaultReplicas = 2

Execution:
  isHoliday = holidayDates["2025-12-25"] → TRUE
  mode = "treat-as-closed"
  Result: (TRUE, &2)

Caller behavior: Use overrideReplicas=2, skip window matching

---

Scenario 2: Holiday with treat-as-open mode

Input:
  holidayDates = {"2025-01-01": true}
  localDate = "2025-01-01"
  mode = "treat-as-open"
  windows = [
    {days:[Mon-Fri], start:09:00, end:17:00, replicas:10},
    {days:[Sat-Sun], start:00:00, end:23:59, replicas:3}
  ]
  defaultReplicas = 0

Execution:
  isHoliday = TRUE
  mode = "treat-as-open"
  maxReplicas = MAX(10, 3) → 10
  Result: (TRUE, &10)

Caller behavior: Use overrideReplicas=10, skip window matching

---

Scenario 3: Not a holiday

Input:
  holidayDates = {"2025-12-25": true}
  localDate = "2025-06-15"
  mode = "treat-as-closed"
  windows = [...]
  defaultReplicas = 2

Execution:
  isHoliday = holidayDates["2025-06-15"] → FALSE
  Result: (FALSE, NULL)

Caller behavior: Proceed with normal window matching
```

---

## Algorithm 5: Main Reconcile Flow

**Purpose:** Orchestrate all logic in the reconciliation loop.

**Inputs:**
- `ctx`: Context
- `req`: Request with NamespacedName

**Outputs:**
- `Result`: Requeue decision
- `error`: Error if reconciliation failed

**Pseudocode:**

```
FUNCTION Reconcile(ctx, req):
    startTime = Now()

    // ─────────────────────────────────────────────────────────
    // STEP 1: Fetch TimeWindowScaler
    // ─────────────────────────────────────────────────────────
    tws = GetTimeWindowScaler(ctx, req.NamespacedName)
    IF tws == NULL:
        // Object deleted, nothing to do
        RETURN (NoRequeue, NULL)

    // ─────────────────────────────────────────────────────────
    // STEP 2: Validate and Load Timezone
    // ─────────────────────────────────────────────────────────
    timezone = LoadTimezone(tws.spec.timezone)
    IF timezone is invalid:
        conditions = [
            Ready: (FALSE, "ConfigurationInvalid", "Invalid timezone"),
            Reconciling: (FALSE, "Stable", "Waiting for valid config"),
            Degraded: (TRUE, "InvalidTimezone", "Cannot load timezone")
        ]
        UpdateStatus(ctx, tws, status.effectiveReplicas, ..., conditions)
        EmitEvent(tws, "InvalidSchedule", "Invalid timezone")
        RETURN (RequeueAfter: 5 minutes, NULL)

    // ─────────────────────────────────────────────────────────
    // STEP 3: Convert Current Time to Local Time
    // ─────────────────────────────────────────────────────────
    utcNow = Now()
    localTime = ConvertToTimezone(utcNow, timezone)
    localDate = FormatDate(localTime, "YYYY-MM-DD")

    // ─────────────────────────────────────────────────────────
    // STEP 4: Load Holiday ConfigMap (if configured)
    // ─────────────────────────────────────────────────────────
    holidayDates = {}
    IF tws.spec.holidays != NULL AND tws.spec.holidays.sourceRef != NULL:
        configMap = GetConfigMap(ctx, tws.spec.holidays.sourceRef.name)
        IF configMap == NULL:
            conditions = [
                Ready: (FALSE, "ConfigurationInvalid", "Holiday ConfigMap missing"),
                Reconciling: (FALSE, "Stable", "Waiting for ConfigMap"),
                Degraded: (TRUE, "HolidaySourceMissing", "ConfigMap not found")
            ]
            UpdateStatus(ctx, tws, status.effectiveReplicas, ..., conditions)
            RETURN (RequeueAfter: 5 minutes, NULL)
        ELSE:
            holidayDates = ParseHolidayConfigMap(configMap)

    // ─────────────────────────────────────────────────────────
    // STEP 5: Evaluate Holiday
    // ─────────────────────────────────────────────────────────
    (isHoliday, overrideReplicas) = EvaluateHoliday(
        holidayDates,
        localDate,
        tws.spec.holidays.mode,
        tws.spec.windows,
        tws.spec.defaultReplicas
    )

    IF isHoliday:
        EmitEvent(tws, "HolidayDetected", "Holiday detected for {localDate}")

    // ─────────────────────────────────────────────────────────
    // STEP 6: Compute Effective Replicas
    // ─────────────────────────────────────────────────────────
    IF overrideReplicas != NULL:
        // Holiday overrides window matching
        desiredReplicas = *overrideReplicas
        currentWindow = "Holiday"
        EmitEvent(tws, "WindowOverride", "Holiday override active")
    ELSE:
        // Normal window matching
        desiredReplicas = ComputeEffectiveReplicas(
            tws.spec.windows,
            tws.spec.defaultReplicas,
            localTime
        )
        currentWindow = DetermineWindowLabel(tws.spec.windows, localTime)

    // ─────────────────────────────────────────────────────────
    // STEP 7: Apply Grace Period
    // ─────────────────────────────────────────────────────────
    (finalReplicas, newGraceExpiry) = ApplyGracePeriod(
        tws.status.effectiveReplicas,
        desiredReplicas,
        tws.spec.gracePeriodSeconds,
        tws.status.gracePeriodExpiry,
        utcNow
    )

    inGracePeriod = (newGraceExpiry != NULL)

    IF inGracePeriod AND newGraceExpiry != tws.status.gracePeriodExpiry:
        // Grace period starting
        EmitEvent(tws, "GracePeriodStarted", "Grace period started")

    IF NOT inGracePeriod AND tws.status.gracePeriodExpiry != NULL:
        // Grace period ended (either expired or cancelled)
        IF finalReplicas > tws.status.effectiveReplicas:
            EmitEvent(tws, "GracePeriodCancelled", "Scale-up cancelled grace")
        // Else: grace expired naturally (ScaledDown event emitted later)

    // ─────────────────────────────────────────────────────────
    // STEP 8: Get Target Workload
    // ─────────────────────────────────────────────────────────
    targetRef = tws.spec.targetRef
    targetDeployment = GetDeployment(ctx, targetRef.namespace, targetRef.name)

    IF targetDeployment == NULL:
        conditions = [
            Ready: (FALSE, "TargetNotFound", "Deployment not found"),
            Reconciling: (FALSE, "Stable", "Waiting for target"),
            Degraded: (FALSE, "OperationalNormal", "No issues")
        ]
        UpdateStatus(ctx, tws, finalReplicas, 0, currentWindow, ..., conditions)
        EmitEvent(tws, "MissingTarget", "Target Deployment not found")
        RETURN (RequeueAfter: 30 seconds, NULL)

    targetSpecReplicas = targetDeployment.spec.replicas
    targetObservedReplicas = targetDeployment.status.replicas

    // ─────────────────────────────────────────────────────────
    // STEP 9: Determine Write Need
    // ─────────────────────────────────────────────────────────
    writeNeeded = (targetSpecReplicas != finalReplicas)

    // ─────────────────────────────────────────────────────────
    // STEP 10: Handle Pause Mode
    // ─────────────────────────────────────────────────────────
    IF tws.spec.pause == TRUE:
        // Compute state but don't write
        conditions = [
            Ready: (targetSpecReplicas == finalReplicas, "TargetMismatch", "..."),
            Reconciling: (FALSE, "Stable", "Paused"),
            Degraded: (FALSE, "OperationalNormal", "No issues")
        ]
        UpdateStatus(ctx, tws, finalReplicas, targetObservedReplicas, currentWindow, ..., conditions)
        EmitEvent(tws, "ScalingSkipped", "Scaling skipped due to pause")

        // Requeue at next boundary
        nextBoundary = ComputeNextBoundary(tws.spec.windows, localTime)
        requeueDuration = ComputeRequeueDuration(nextBoundary, localTime)
        RETURN (RequeueAfter: requeueDuration, NULL)

    // ─────────────────────────────────────────────────────────
    // STEP 11: Update Target Workload
    // ─────────────────────────────────────────────────────────
    lastScaleTime = tws.status.lastScaleTime

    IF writeNeeded:
        err = PatchDeploymentReplicas(ctx, targetDeployment, finalReplicas)
        IF err is Conflict:
            // Optimistic lock failed, requeue immediately
            RETURN (Requeue: true, NULL)
        ELSE IF err != NULL:
            // Other error
            conditions = [
                Ready: (FALSE, "TargetUpdateFailed", "Failed to update target"),
                Reconciling: (TRUE, "ScaleInProgress", "Retrying scale"),
                Degraded: (TRUE, "TargetUpdateFailed", "Update failed")
            ]
            UpdateStatus(ctx, tws, finalReplicas, targetObservedReplicas, currentWindow, ..., conditions)
            RETURN (RequeueAfter: 30 seconds, err)

        // Update succeeded
        lastScaleTime = Now()

        // Emit scale events
        IF finalReplicas > targetSpecReplicas:
            EmitScaleUp(tws, targetSpecReplicas, finalReplicas, currentWindow, "WindowEntered")
        ELSE IF finalReplicas < targetSpecReplicas:
            gracePeriodApplied = (tws.status.gracePeriodExpiry != NULL AND newGraceExpiry == NULL)
            EmitScaleDown(tws, targetSpecReplicas, finalReplicas, currentWindow, gracePeriodApplied, tws.spec.gracePeriodSeconds)

    // ─────────────────────────────────────────────────────────
    // STEP 12: Build Status Conditions
    // ─────────────────────────────────────────────────────────
    conditions = BuildConditions(
        ready: (targetSpecReplicas == finalReplicas, "Reconciled", "Target matches desired"),
        reconciling: (FALSE, "Stable", "Waiting until {nextBoundary}"),
        degraded: (FALSE, "OperationalNormal", "No issues")
    )

    // ─────────────────────────────────────────────────────────
    // STEP 13: Update Status
    // ─────────────────────────────────────────────────────────
    err = UpdateStatus(
        ctx,
        tws,
        finalReplicas,
        targetObservedReplicas,
        currentWindow,
        lastScaleTime,
        newGraceExpiry,
        conditions,
        tws.metadata.generation
    )
    IF err is Conflict:
        // Status update conflict, requeue immediately
        RETURN (Requeue: true, NULL)
    ELSE IF err != NULL:
        RETURN (RequeueAfter: 30 seconds, err)

    // ─────────────────────────────────────────────────────────
    // STEP 14: Record Metrics
    // ─────────────────────────────────────────────────────────
    IF writeNeeded:
        direction = (finalReplicas > targetSpecReplicas) ? "up" : "down"
        RecordScaleEvent(tws.name, tws.namespace, direction, targetSpecReplicas, finalReplicas)

    RecordStateChange(tws.name, tws.namespace, currentWindow, finalReplicas, inGracePeriod)

    reconcileDuration = Now() - startTime
    RecordReconcileDuration(tws.name, tws.namespace, reconcileDuration.Seconds())

    // ─────────────────────────────────────────────────────────
    // STEP 15: Compute Requeue Duration
    // ─────────────────────────────────────────────────────────
    nextBoundary = ComputeNextBoundary(tws.spec.windows, localTime)
    requeueDuration = ComputeRequeueDuration(nextBoundary, localTime)

    RETURN (RequeueAfter: requeueDuration, NULL)
END FUNCTION

// ═════════════════════════════════════════════════════════════
// HELPER: ComputeRequeueDuration
// ═════════════════════════════════════════════════════════════
FUNCTION ComputeRequeueDuration(nextBoundary, now):
    baseDuration = nextBoundary - now

    // Add jitter to prevent thundering herd
    jitterSeconds = RandomInt(5, 25)
    durationWithJitter = baseDuration + jitterSeconds seconds

    // Quantize to 10-second boundaries
    quantized = RoundToNearest(durationWithJitter, 10 seconds)

    // Enforce minimum and maximum
    final = MAX(30 seconds, MIN(24 hours, quantized))

    RETURN final
END FUNCTION
```

---

## Testing Pseudocode

### Unit Test Example: ComputeEffectiveReplicas

```
TEST "ComputeEffectiveReplicas with in-window time":
    // Setup
    windows = [
        {days: [Mon,Tue,Wed,Thu,Fri], start: 09:00, end: 17:00, replicas: 10}
    ]
    defaultReplicas = 2
    localTime = "2025-01-27 14:30:00" (Monday)

    // Execute
    result = ComputeEffectiveReplicas(windows, defaultReplicas, localTime)

    // Assert
    EXPECT result == 10
END TEST

TEST "ComputeEffectiveReplicas with out-of-window time":
    // Setup
    windows = [
        {days: [Mon,Tue,Wed,Thu,Fri], start: 09:00, end: 17:00, replicas: 10}
    ]
    defaultReplicas = 2
    localTime = "2025-01-27 20:00:00" (Monday)

    // Execute
    result = ComputeEffectiveReplicas(windows, defaultReplicas, localTime)

    // Assert
    EXPECT result == 2
END TEST

TEST "ComputeEffectiveReplicas with cross-midnight window":
    // Setup
    windows = [
        {days: [Fri], start: 22:00, end: 02:00, replicas: 5}
    ]
    defaultReplicas = 2
    localTime = "2025-01-25 01:00:00" (Saturday)

    // Execute
    result = ComputeEffectiveReplicas(windows, defaultReplicas, localTime)

    // Assert
    EXPECT result == 5  // Saturday 01:00 matches Friday 22:00-02:00 window
END TEST
```

---

## Summary

This pseudocode provides implementation-independent algorithms for:

1. **ComputeEffectiveReplicas** - Window matching with cross-midnight support
2. **ComputeNextBoundary** - Next state change calculation
3. **ApplyGracePeriod** - Grace period state machine
4. **EvaluateHoliday** - Holiday override logic
5. **Reconcile** - Complete orchestration flow

These algorithms can be implemented in any language following these contracts.
