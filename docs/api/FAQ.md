# TimeWindowScaler FAQ

## How are overlapping windows resolved?

**Last matching window wins.** When multiple windows in the `spec.windows` array match the current time, the controller selects the **last matching window** in the array. This gives users explicit control over precedence through ordering.

Example:
```yaml
windows:
- days: [Mon]
  start: "09:00"
  end: "12:00"
  replicas: 2
- days: [Mon]
  start: "11:00"  # Overlaps with first window
  end: "13:00"
  replicas: 4
```

At 11:30 on Monday, both windows match, but the second window (4 replicas) is applied because it appears last in the array.

## How is DST (Daylight Saving Time) handled?

**Full IANA timezone rules apply.** The controller uses the Go standard library's time package with IANA timezone data, which automatically handles DST transitions.

Key behaviors:
- Windows are evaluated in local time after DST adjustment
- During "spring forward", the skipped hour never matches any window
- During "fall back", the repeated hour matches windows twice
- Cross-midnight windows adjust their effective duration during DST changes

Example: A window from 22:00-06:00 in `America/New_York`:
- Normally 8 hours
- On spring DST change: 7 hours (2 AM becomes 3 AM)
- On fall DST change: 9 hours (2 AM occurs twice)

## What happens if a user manually scales the Deployment?

**Manual changes are corrected on the next reconciliation** unless `pause: true`.

Behavior sequence:
1. User runs `kubectl scale deployment/webapp --replicas=20`
2. Deployment scales to 20 replicas
3. On next reconcile (typically within 30 seconds):
   - Controller detects drift: `targetObservedReplicas: 20` vs `effectiveReplicas: 5`
   - If `pause: false`: Scales back to 5, emits event "Corrected manual drift from 20 to 5 replicas"
   - If `pause: true`: Observes drift, updates status, but takes no action

This ensures the TimeWindowScaler maintains control unless explicitly paused.

## How do holiday modes work?

### treat-as-closed
**Behaves as if no windows match on holiday dates.**
- Uses `defaultReplicas` for the entire day
- Completely ignores all defined windows
- Useful for businesses closed on holidays

### treat-as-open
**Creates a synthetic window with maximum capacity.**
- Calculates `max(all window.replicas values)`
- Applies this max value for the entire holiday
- If no windows defined, uses `defaultReplicas`
- Useful for high-traffic holidays (Black Friday, etc.)

### ignore (default)
**Normal window processing.**
- Holidays have no special effect
- Windows match normally based on day and time

## What does pause mean in practice?

**Pause suspends target modifications while maintaining full observability.**

When `pause: true`:
1. Controller continues computing `effectiveReplicas` based on current time/windows
2. All status fields update normally
3. Events are emitted describing what *would* happen
4. **No writes to the target Deployment**
5. Ready condition reflects alignment:
   - If target matches computed state: `Ready=True`
   - If drift exists: `Ready=False` with reason `TargetMismatch`

Use cases:
- Temporary manual override for incidents
- Maintenance windows
- Testing window configurations without impact
- Gradual rollout by observing computed values first

## What does the grace period delay and where does it apply?

**Grace period delays downscaling only.**

Rules:
- **Only applies when `effectiveReplicas` decreases**
- Timer starts when leaving a higher-replica state
- During grace, maintains previous higher replica count
- After expiration, applies new lower replica count

Example with `gracePeriodSeconds: 300` (5 minutes):
- 17:00: Window ends, should scale from 10 to 2 replicas
- 17:00-17:05: Maintains 10 replicas (grace period)
- 17:05: Scales down to 2 replicas

Does NOT apply for:
- Scaling up (happens immediately)
- Same replica count transitions
- Initial deployment creation

## Why is end exclusive and start inclusive?

**Follows standard programming interval notation [start, end).**

Benefits:
- **No gaps or overlaps** when defining adjacent windows
- **Precise midnight handling** without ambiguity
- **Industry standard** matching cron, business hours systems

Example of seamless adjacent windows:
```yaml
- start: "09:00"  # Includes 09:00:00.000
  end: "12:00"    # Excludes 12:00:00.000
- start: "12:00"  # Includes 12:00:00.000
  end: "17:00"    # Excludes 17:00:00.000
```

No gap or overlap at noon - exactly one window matches at any time.

## Why does the last matching window win?

**Provides explicit, predictable precedence control.**

Advantages:
- **Deterministic behavior** - no ambiguity about which window applies
- **Override capability** - add specific exceptions at the end
- **Maintainable** - precedence visible in YAML structure
- **Debugging-friendly** - can trace through array to find active window

Example override pattern:
```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri]  # Base weekday pattern
  start: "09:00"
  end: "17:00"
  replicas: 10
- days: [Wed]  # Wednesday override (last wins)
  start: "09:00"
  end: "12:00"
  replicas: 5  # Half capacity Wednesday mornings
```

## Why only Deployment support in v1alpha1?

**Focused initial scope for production stability.**

Current state (v1alpha1):
- Only `kind: Deployment` supported
- Covers 90% of time-based scaling use cases
- Simpler controller logic, easier to verify correctness
- Fast iteration on core scheduling features

Future expansion (v1beta1):
- **StatefulSet**: Ordered scaling for stateful workloads
- **ReplicaSet**: Direct replica control
- **Custom resources**: Via duck-typing scale subresource

Migration path:
- v1alpha1 resources continue working
- Automatic conversion to v1beta1 schema
- No breaking changes to existing fields

## What happens during timezone resolution failures?

**Degraded condition with fallback behavior.**

When timezone cannot be resolved:
1. Status condition: `Degraded=True`, reason: `InvalidTimezone`
2. Fallback: Uses `defaultReplicas`
3. Event: "Failed to load timezone 'Mars/Olympus_Mons': unknown time zone"
4. Continues attempting resolution on each reconcile

Common causes:
- Typo in timezone name
- Missing tzdata in container image
- Outdated IANA timezone database

## How do cross-midnight windows work exactly?

**Window extends into the next calendar day when end < start.**

Key rules:
- Days list refers to the **start day** only
- Window is active from start time on listed day until end time on **next calendar day**

Example:
```yaml
- days: [Fri]
  start: "22:00"  # Friday 10 PM
  end: "02:00"    # Saturday 2 AM
```

Matches:
- Friday 22:00 - 23:59 ✓
- Saturday 00:00 - 01:59 ✓
- Saturday 02:00 ✗ (window ended)

Does NOT match on Thursday 22:00 or Saturday 22:00.

## What's the recommended testing workflow?

1. **Start with pause: true**
   - Deploy TimeWindowScaler with `pause: true`
   - Verify `effectiveReplicas` in status matches expectations
   - Check conditions and events

2. **Test window transitions**
   - Temporarily adjust system time or wait for natural transitions
   - Confirm `currentWindow` changes correctly
   - Verify grace periods by watching status.lastScaleTime

3. **Enable with confidence**
   - Set `pause: false`
   - Monitor first few transitions
   - Check Deployment events for scale operations

4. **Holiday testing**
   - Add test date to holiday ConfigMap
   - Verify holiday mode behavior
   - Remove test date when complete