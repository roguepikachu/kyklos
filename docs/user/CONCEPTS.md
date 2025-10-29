# Kyklos Concepts

This guide explains the core concepts behind Kyklos time-based scaling.

## Time Windows

A time window defines when your application should have a specific replica count.

### Basic Window Structure

```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri]
  start: "09:00"
  end: "17:00"
  replicas: 10
```

This window means "scale to 10 replicas from 9 AM to 5 PM on weekdays."

### Start Inclusive, End Exclusive

Windows follow standard interval notation: `[start, end)`.

- **Start time is included** - At exactly 09:00:00, the window is active
- **End time is excluded** - At 17:00:00, the window has ended

**Why this matters:**
```yaml
# No gap or overlap between these windows
- start: "09:00"  # Active at 09:00:00.000
  end: "12:00"    # Ends just before 12:00:00.000
- start: "12:00"  # Active at 12:00:00.000
  end: "17:00"    # Ends just before 17:00:00.000
```

At noon, only the second window matches. No ambiguity.

### Days of Week

Use three-letter abbreviations: `Mon`, `Tue`, `Wed`, `Thu`, `Fri`, `Sat`, `Sun`.

A window can apply to multiple days:
```yaml
- days: [Mon, Wed, Fri]
  start: "10:00"
  end: "14:00"
  replicas: 5
```

This window only matches on Mondays, Wednesdays, and Fridays.

## Cross-Midnight Windows

When the end time is before the start time, the window crosses midnight into the next day.

### Example: Night Shift

```yaml
- days: [Mon, Tue, Wed, Thu, Fri]
  start: "22:00"  # 10 PM
  end: "06:00"    # 6 AM next day
  replicas: 3
```

**What this matches:**
- Monday 22:00 - 23:59 (Monday night)
- Tuesday 00:00 - 05:59 (Tuesday early morning)
- Repeats Tuesday-Wednesday, Wednesday-Thursday, etc.

**Key rule:** The `days` list specifies the starting day only.

### Midnight Crossing Calculation

For a Friday 22:00 - 02:00 window:

1. **Listed day (Friday):**
   - Active from Friday 22:00 until Friday 23:59
2. **Next day (Saturday):**
   - Active from Saturday 00:00 until Saturday 01:59
   - Ends at Saturday 02:00

The window does NOT match Saturday 22:00 (that would need `days: [Sat]`).

## Effective Replicas (Current Desired State)

The **effectiveReplicas** field in status shows the number of replicas Kyklos has computed as correct for right now, based on current time and window matching. This is the replica count Kyklos will write to the target deployment.

**Terminology clarification:**
- `windows[].replicas` - Configured in spec, what you want during each window
- `defaultReplicas` - Configured in spec, what you want when no windows match
- `effectiveReplicas` - Computed in status, what controller wants RIGHT NOW
- `targetObservedReplicas` - Observed in status, what the deployment actually has

### Computation Flow

1. **Check current time in specified timezone**
2. **Check if today is a holiday** (if configured)
3. **Evaluate all windows in order**
4. **Return the last matching window's replicas**
5. **If no windows match, use defaultReplicas**

### Example Computation

Given this configuration at Tuesday 11:30 AM:

```yaml
timezone: America/New_York
defaultReplicas: 2
windows:
- days: [Mon, Tue, Wed, Thu, Fri]
  start: "09:00"
  end: "12:00"
  replicas: 5
- days: [Tue, Thu]
  start: "11:00"
  end: "13:00"
  replicas: 8
```

**Evaluation:**
1. Current time: Tuesday 11:30 AM
2. First window matches: 09:00-12:00 on Tuesday (5 replicas)
3. Second window also matches: 11:00-13:00 on Tuesday (8 replicas)
4. **Last matching window wins: 8 replicas**

## Window Matching Precedence

When multiple windows match the current time, **the last one in the array wins**.

This gives you explicit control over precedence:

```yaml
windows:
# Base pattern for all weekdays
- days: [Mon, Tue, Wed, Thu, Fri]
  start: "09:00"
  end: "17:00"
  replicas: 10

# Override for Wednesday afternoons
- days: [Wed]
  start: "13:00"
  end: "17:00"
  replicas: 5  # Only 5 replicas needed
```

On Wednesday at 14:00, both windows match, but the second wins (5 replicas).

## Default Replicas

The replica count when no windows match.

```yaml
defaultReplicas: 2
```

This applies:
- Outside all defined windows
- On days not listed in any window
- When in `treat-as-closed` holiday mode

Set this to your "safe minimum" capacity.

## Timezones

Kyklos uses IANA timezone identifiers with full DST support.

### Timezone Examples

```yaml
timezone: America/New_York  # US Eastern Time
timezone: Europe/London     # UK time
timezone: Asia/Kolkata      # Indian Standard Time
timezone: UTC               # No DST, fixed offset
```

### How DST Works

The Go standard library automatically handles DST transitions.

**Spring Forward (clocks advance):**
```
2:00 AM becomes 3:00 AM (1 hour skipped)
```
- Windows during skipped hour never match
- Windows spanning the transition are shortened by 1 hour

**Fall Back (clocks retreat):**
```
2:00 AM occurs twice (1 hour repeated)
```
- Windows during repeated hour match twice
- Windows spanning the transition are extended by 1 hour

### DST Example

Window: `01:00-04:00` in `America/New_York`

**Normal day:** 3 hours (01:00, 02:00, 03:00)

**Spring forward:** 2 hours (01:00, 03:00) - 02:00 is skipped

**Fall back:** 4 hours (01:00, 02:00-first, 02:00-second, 03:00)

Kyklos computes this correctly without manual intervention.

## Next Boundary Computation

The controller schedules reconciliation at the next window boundary to minimize API calls.

### Boundary Types

1. **Window start** - When a new window begins
2. **Window end** - When a window ends
3. **Midnight** - Default boundary if no windows in next 24 hours

### Computation Algorithm

1. Find all window starts and ends for next 24 hours
2. Convert each to local time in the configured timezone
3. Select the earliest boundary after current time
4. Add small jitter (5-25 seconds)
5. Schedule reconciliation

**Example:** Current time Tuesday 14:00, windows at 09:00-17:00 and 17:00-22:00

- Next boundaries: Tuesday 17:00, Tuesday 22:00, Wednesday 09:00
- **Controller requeues at Tuesday 17:00** (earliest after now)

> **Note:** Holiday support is available in v0.1 with ConfigMap-based sources. External calendar sync and advanced recurring patterns are planned for v0.2.

## Holiday Handling

Holidays override normal window matching based on the configured mode.

### Holiday Source

Holidays come from a ConfigMap:

```yaml
holidays:
  mode: treat-as-closed
  sourceRef:
    name: company-holidays
```

The ConfigMap format:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: company-holidays
data:
  "2025-12-25": "Christmas Day"
  "2025-07-04": "Independence Day"
  "2025-11-28": "Thanksgiving"
```

Keys must be ISO dates (`YYYY-MM-DD`). Values are ignored but can document the holiday.

### Holiday Modes

#### ignore (default)

Normal window matching. Holidays have no effect.

```yaml
holidays:
  mode: ignore
```

#### treat-as-closed

All windows are ignored on holidays. Uses `defaultReplicas`.

```yaml
holidays:
  mode: treat-as-closed
  sourceRef:
    name: company-holidays
```

**Use case:** Business closed on holidays.

On Christmas Day with `defaultReplicas: 1`, the deployment scales to 1 replica regardless of windows.

#### treat-as-open

Creates a synthetic window with maximum capacity on holidays.

```yaml
holidays:
  mode: treat-as-open
  sourceRef:
    name: company-holidays
```

The replica count is `max(all window.replicas values)`.

**Example:**
```yaml
windows:
- replicas: 5
- replicas: 10
- replicas: 3
holidays:
  mode: treat-as-open
```

On a holiday: uses 10 replicas (the maximum).

**Use case:** High traffic holidays like Black Friday.

## Grace Period

An optional delay before downscaling.

```yaml
gracePeriodSeconds: 300  # 5 minutes
```

### When Grace Applies

**Only when replicas decrease:**
- Leaving a window with 10 replicas to defaultReplicas of 2
- Transitioning from high window to low window

**Does NOT apply:**
- Scaling up
- Same replica count
- Initial resource creation

### Grace Period Flow

1. **Window ends at 17:00** - Should scale from 10 to 2 replicas
2. **17:00-17:05** - Grace period active, maintains 10 replicas
3. **17:05** - Grace expires, scales down to 2 replicas

Status reflects grace state:
```yaml
status:
  effectiveReplicas: 10  # Maintaining during grace
  targetObservedReplicas: 10
```

After grace expires:
```yaml
status:
  effectiveReplicas: 2   # Applied after grace
  lastScaleTime: "2025-10-28T17:05:00Z"
```

## Manual Drift Correction

Kyklos continuously enforces the desired replica count, reverting manual changes.

### Drift Detection

On each reconcile:
1. Compare deployment's actual replicas with effective replicas
2. If different: drift detected

### Drift Correction

**When pause is false:**
```
User scales to 15 → Controller detects drift → Scales back to 10
```

Event emitted:
```
Normal  DriftCorrected  Corrected manual drift from 15 to 10 replicas
```

**When pause is true:**
```
User scales to 15 → Controller detects drift → Updates status only
```

Status shows mismatch:
```yaml
status:
  effectiveReplicas: 10  # What Kyklos wants
  targetObservedReplicas: 15  # What deployment has
  conditions:
  - type: Ready
    status: "False"
    reason: TargetMismatch
    message: "Paused: deployment has 15 replicas, desired 10"
```

## Pause Semantics

The `pause` field suspends target modifications while maintaining observability.

```yaml
spec:
  pause: true
```

### Paused Behavior

1. **Computation continues** - effectiveReplicas still calculated
2. **Status updates** - All status fields populated correctly
3. **Events emitted** - Describes what would happen
4. **No writes to target** - Deployment is not modified
5. **Conditions reflect state** - Ready condition shows alignment

### Pause Use Cases

**Incident response:**
```bash
kubectl patch tws my-scaler -p '{"spec":{"pause":true}}'
# Manually scale deployment for incident
# Kyklos won't fight you
```

**Testing configurations:**
```bash
# Apply new TimeWindowScaler with pause: true
# Watch status to verify effectiveReplicas computation
# When confident, set pause: false
```

**Maintenance windows:**
```bash
# Pause before maintenance
# Manually scale as needed
# Resume after maintenance
```

## Status Conditions

Kyklos uses standard Kubernetes conditions to report health.

### Ready Condition

Indicates whether the target matches desired state.

**Ready=True:**
```yaml
conditions:
- type: Ready
  status: "True"
  reason: Reconciled
  message: "Target deployment matches desired replicas"
```

**Ready=False (target mismatch):**
```yaml
conditions:
- type: Ready
  status: "False"
  reason: TargetMismatch
  message: "Deployment has 5 replicas, desired 10"
```

**Ready=False (target not found):**
```yaml
conditions:
- type: Ready
  status: "False"
  reason: TargetNotFound
  message: "Deployment 'webapp' not found in namespace 'prod'"
```

### Reconciling Condition

Indicates ongoing state changes.

**Reconciling=True:**
```yaml
conditions:
- type: Reconciling
  status: "True"
  reason: WindowTransition
  message: "Transitioning from OffHours to BusinessHours window"
```

**Reconciling=False:**
```yaml
conditions:
- type: Reconciling
  status: "False"
  reason: Stable
  message: "No ongoing reconciliation"
```

### Degraded Condition

Indicates configuration or operational problems.

**Degraded=True (invalid timezone):**
```yaml
conditions:
- type: Degraded
  status: "True"
  reason: InvalidTimezone
  message: "Failed to load timezone 'America/Invalid': unknown time zone"
```

**Degraded=True (holiday ConfigMap missing):**
```yaml
conditions:
- type: Degraded
  status: "True"
  reason: HolidaySourceMissing
  message: "ConfigMap 'company-holidays' not found"
```

**Degraded=False:**
```yaml
conditions:
- type: Degraded
  status: "False"
  reason: OperationalNormal
  message: "No errors detected"
```

## Status Fields

### currentWindow

Label identifying the active window.

```yaml
status:
  currentWindow: BusinessHours
```

Possible values:
- `BusinessHours` - Standard daytime window
- `OffHours` - No window matches, using defaultReplicas
- `Custom-<hash>` - Unique window identified by configuration hash

### effectiveReplicas

The replica count Kyklos wants right now.

```yaml
status:
  effectiveReplicas: 10
```

This is the result of window evaluation and grace period logic.

### targetObservedReplicas

The actual replica count of the target deployment.

```yaml
status:
  targetObservedReplicas: 10
```

When this differs from effectiveReplicas, drift exists.

### lastScaleTime

Timestamp of the last scale operation.

```yaml
status:
  lastScaleTime: "2025-10-28T14:30:15Z"
```

Useful for:
- Calculating time since last scale
- Grace period expiration tracking
- Debugging scale frequency

### observedGeneration

The spec generation last processed.

```yaml
metadata:
  generation: 5
status:
  observedGeneration: 5
```

When these match, the status is current. When they differ, a reconcile is pending.

## Window Labels

Kyklos assigns semantic labels to common window patterns for easier status reading.

### BusinessHours

Monday-Friday 09:00-17:00 patterns.

```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri]
  start: "09:00"
  end: "17:00"
  replicas: 10
```

Status shows:
```yaml
status:
  currentWindow: BusinessHours
```

### OffHours

When no windows match.

```yaml
status:
  currentWindow: OffHours
  effectiveReplicas: 2  # Uses defaultReplicas
```

### Custom Windows

Non-standard patterns get a hash-based label:

```yaml
windows:
- days: [Tue, Thu]
  start: "14:00"
  end: "16:00"
  replicas: 7
```

Status shows:
```yaml
status:
  currentWindow: Custom-a3b2c1
```

The hash ensures consistent labeling across reconciles.

## Best Practices

### Timezone Selection

**Use local timezone for business hours:**
```yaml
timezone: America/New_York  # For New York office
timezone: Europe/London     # For London office
```

**Use UTC for global services:**
```yaml
timezone: UTC  # Predictable, no DST surprises
```

### Window Design

**Prefer non-overlapping windows:**
```yaml
# Clear, easy to understand
- start: "09:00"
  end: "17:00"
  replicas: 10
- start: "17:00"
  end: "22:00"
  replicas: 5
```

**Use overlaps only for overrides:**
```yaml
# Base pattern
- days: [Mon, Tue, Wed, Thu, Fri]
  start: "09:00"
  end: "17:00"
  replicas: 10
# Friday afternoon exception
- days: [Fri]
  start: "13:00"
  end: "17:00"
  replicas: 5
```

### Default Replicas

**Set to safe minimum:**
```yaml
defaultReplicas: 2  # Enough to handle basic traffic
```

Not zero (unless truly zero traffic expected).

### Grace Periods

**Match workload drain time:**
```yaml
gracePeriodSeconds: 300  # 5 minutes for connection draining
```

Consider:
- Time to drain connections
- Time for in-flight requests to complete
- Kubernetes termination grace period

## Next Topics

- **[Operations Guide](OPERATIONS.md)** - Metrics, alerts, and production patterns
- **[FAQ](FAQ.md)** - Common questions about window behavior
- **[Troubleshooting](TROUBLESHOOTING.md)** - Fix issues with window matching
- **[API Reference](../api/CRD-SPEC.md)** - Complete field documentation
