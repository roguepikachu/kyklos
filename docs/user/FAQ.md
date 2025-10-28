# Frequently Asked Questions

Quick answers to common questions about Kyklos time-based scaling.

## General Concepts

### Why is end time exclusive but start time inclusive?

This follows standard interval notation `[start, end)` used in programming and mathematics.

**Benefit:** Adjacent windows connect perfectly with no gaps or overlaps.

```yaml
# These windows connect seamlessly at noon
- start: "09:00"  # Includes 09:00:00
  end: "12:00"    # Excludes 12:00:00
- start: "12:00"  # Includes 12:00:00
  end: "17:00"    # Excludes 17:00:00
```

At 12:00:00 exactly, only the second window matches.

### How do overlapping windows work?

**Last matching window wins.**

When multiple windows match the current time, Kyklos uses the replica count from the last matching window in the array.

```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri]
  start: "09:00"
  end: "17:00"
  replicas: 10
- days: [Wed]
  start: "13:00"
  end: "17:00"
  replicas: 5  # This wins on Wednesday afternoons
```

On Wednesday at 14:00, both windows match, but replica count is 5 (from the second window).

### What happens if no windows match?

Kyklos uses `defaultReplicas`.

```yaml
defaultReplicas: 2
```

This applies when:
- Current time is outside all defined windows
- Current day is not in any window's days list
- Holiday mode is `treat-as-closed`

## Cross-Midnight Windows

### How do I create a window that spans midnight?

Set end time before start time.

```yaml
- days: [Fri]
  start: "22:00"  # Friday 10 PM
  end: "06:00"    # Saturday 6 AM
  replicas: 5
```

This window is active:
- Friday 22:00-23:59
- Saturday 00:00-05:59

### Which day do I specify for cross-midnight windows?

**The starting day only.**

For a Friday night to Saturday morning window, use `days: [Fri]`.

```yaml
- days: [Fri]  # Starting day
  start: "22:00"
  end: "06:00"
  replicas: 5
```

Don't include both days:
```yaml
# WRONG - don't do this
- days: [Fri, Sat]
  start: "22:00"
  end: "06:00"
```

This would create two separate windows (Friday 22:00-06:00 and Saturday 22:00-06:00).

### What if I want continuous night coverage?

List each starting day separately:

```yaml
# Sunday night through Saturday morning
- days: [Sun, Mon, Tue, Wed, Thu, Fri]
  start: "22:00"
  end: "06:00"
  replicas: 3
```

This gives you night shift coverage every night Sunday-Saturday.

## Timezones and DST

### How does Kyklos handle DST transitions?

Automatically using IANA timezone data.

**Spring forward (2:00 becomes 3:00):**
- The skipped hour never matches any window
- Windows spanning the transition are 1 hour shorter

**Fall back (2:00 occurs twice):**
- The repeated hour matches windows twice
- Windows spanning the transition are 1 hour longer

### Should I use UTC or local timezone?

**Use local timezone for business hours:**
```yaml
timezone: America/New_York
```

Benefits:
- Windows stay aligned with local business hours
- DST handled automatically
- Matches team expectations

**Use UTC for global services:**
```yaml
timezone: UTC
```

Benefits:
- No DST surprises
- Predictable across regions
- Simpler for distributed teams

### What happens during the DST skip hour?

Windows during the skipped hour never match.

**Example:** Spring forward at 2:00 AM becomes 3:00 AM.

Window `01:30-02:30` on that day:
- 01:30-01:59: Active (30 minutes)
- 02:00-02:59: Skipped (doesn't exist)
- Result: Window is active for only 30 minutes instead of 60

## Manual Scaling and Drift

### Why does Kyklos keep reverting my manual changes?

**This is intentional.** Kyklos continuously enforces the desired replica count based on time windows.

When you manually scale a deployment, Kyklos detects drift and corrects it on the next reconcile (typically within 30 seconds).

### How do I manually scale temporarily?

**Option 1: Pause the TimeWindowScaler**
```bash
kubectl patch tws my-scaler -n production \
  --type=merge -p '{"spec":{"pause":true}}'

# Manually scale deployment
kubectl scale deployment my-app --replicas=20

# Resume when ready
kubectl patch tws my-scaler -n production \
  --type=merge -p '{"spec":{"pause":false}}'
```

**Option 2: Delete the TimeWindowScaler**
```bash
kubectl delete tws my-scaler -n production

# Deployment no longer managed
kubectl scale deployment my-app --replicas=20
```

**Option 3: Update the TimeWindowScaler**
```bash
# Change defaultReplicas or window replicas to match what you want
kubectl patch tws my-scaler -n production \
  --type=merge -p '{"spec":{"defaultReplicas":20}}'
```

### What does pause actually do?

**Pause suspends target modifications while maintaining observability.**

When paused:
- Status continues updating (effectiveReplicas computed correctly)
- Events are emitted ("would scale to X replicas")
- No writes to the target deployment
- Ready condition shows `TargetMismatch` if drift exists

Use cases:
- Incident response (manual override needed)
- Testing window configurations
- Maintenance windows

## Grace Periods

### What does grace period delay?

**Only downscaling** (reducing replicas).

```yaml
gracePeriodSeconds: 300  # 5 minutes
```

**Applies to:**
- Leaving a high-replica window: 10 → 2 replicas (delayed 5 minutes)
- Transitioning to lower window: 10 → 5 replicas (delayed 5 minutes)

**Does NOT apply to:**
- Scaling up: 2 → 10 replicas (immediate)
- Same replica count: 10 → 10 replicas (no change)

### Why isn't my grace period working?

**Grace only applies when replicas decrease.**

Check:
1. Are you scaling down? `status.effectiveReplicas < status.targetObservedReplicas`
2. Is grace period configured? `spec.gracePeriodSeconds > 0`
3. Check controller logs for grace period messages

**During grace period:**
```yaml
status:
  effectiveReplicas: 10  # Maintaining during grace
  conditions:
  - message: "Grace period active, expires at 17:05:00"
```

**After grace expires:**
```yaml
status:
  effectiveReplicas: 2   # Applied after grace
  lastScaleTime: "2025-10-28T17:05:00Z"
```

## Holidays

### How do I configure holidays?

Create a ConfigMap with ISO date keys:

```bash
kubectl create configmap company-holidays -n production \
  --from-literal='2025-12-25'='Christmas' \
  --from-literal='2025-01-01'='New Year' \
  --from-literal='2025-07-04'='Independence Day'
```

Reference in TimeWindowScaler:

```yaml
spec:
  holidays:
    mode: treat-as-closed
    sourceRef:
      name: company-holidays
```

### What do the holiday modes mean?

**ignore (default):**
Normal window matching. Holidays have no effect.

**treat-as-closed:**
All windows ignored on holidays. Uses `defaultReplicas`.

Use for: Business closed on holidays.

**treat-as-open:**
Creates synthetic window with `max(all window replicas)` on holidays.

Use for: High-traffic holidays (Black Friday, sales events).

### Can I use different ConfigMaps for different TimeWindowScalers?

Yes. Each TimeWindowScaler can reference its own ConfigMap.

```yaml
# US team
holidays:
  mode: treat-as-closed
  sourceRef:
    name: us-holidays

# EU team
holidays:
  mode: treat-as-closed
  sourceRef:
    name: eu-holidays
```

### What if the holiday ConfigMap is missing?

Controller sets `Degraded=True` condition with reason `HolidaySourceMissing`.

**Behavior:** Falls back to `mode: ignore` (normal window matching).

**Fix:** Create the ConfigMap:
```bash
kubectl create configmap company-holidays -n production \
  --from-literal='2025-12-25'=''
```

## Status and Conditions

### How do I know if my TimeWindowScaler is working?

Check the `Ready` condition:

```bash
kubectl get tws my-scaler -o jsonpath='{.status.conditions[?(@.type=="Ready")]}'
```

**Ready=True:**
```json
{
  "type": "Ready",
  "status": "True",
  "reason": "Reconciled",
  "message": "Target deployment matches desired replicas"
}
```

Your TimeWindowScaler is working correctly.

**Ready=False:**
```json
{
  "type": "Ready",
  "status": "False",
  "reason": "TargetNotFound",
  "message": "Deployment 'my-app' not found"
}
```

There's a problem. Check the reason and message.

### What does observedGeneration mean?

It tracks whether status is current with spec.

```yaml
metadata:
  generation: 5  # Incremented on spec changes
status:
  observedGeneration: 5  # Last processed generation
```

**When they match:** Status is up-to-date.

**When they differ:** A reconcile is pending (spec changed but not yet processed).

### What's the difference between effectiveReplicas and targetObservedReplicas?

**effectiveReplicas:** What Kyklos wants right now (based on windows and time).

**targetObservedReplicas:** What the deployment actually has.

**When they match:** System is aligned.

**When they differ:** Drift exists (manual change, scaling in progress, or pause active).

## Deployment Targets

### Can Kyklos scale StatefulSets or DaemonSets?

**Not in v0.1.** Only Deployments are supported.

```yaml
targetRef:
  kind: Deployment  # Only supported kind
  name: my-app
```

Future versions (v1beta1+) will support:
- StatefulSets
- ReplicaSets
- Custom resources with scale subresource

### Can I scale deployments in different namespaces?

**Not directly.** TimeWindowScaler must be in the same namespace as its target.

```yaml
# TimeWindowScaler in 'production' namespace
metadata:
  namespace: production
spec:
  targetRef:
    kind: Deployment
    name: webapp
    namespace: production  # Optional, defaults to same namespace
```

To scale targets in multiple namespaces, create one TimeWindowScaler per namespace.

### What happens if the target deployment doesn't exist?

Controller sets `Ready=False` with reason `TargetNotFound`.

```yaml
conditions:
- type: Ready
  status: "False"
  reason: TargetNotFound
  message: "Deployment 'my-app' not found in namespace 'production'"
```

**Behavior:**
- Status continues updating (shows what replicas WOULD be)
- Events emitted ("would scale to X replicas")
- No errors or crash loops

**Fix:** Create the deployment or update the TimeWindowScaler target reference.

## Troubleshooting

### Scaling happens at the wrong times

**Likely cause:** Timezone mismatch.

**Check:**
1. What timezone is configured? `kubectl get tws my-scaler -o jsonpath='{.spec.timezone}'`
2. What time does the controller see? Check logs for "Current time in <timezone>"
3. Is your local time matching?

**Fix:**
```bash
# Update to correct timezone
kubectl patch tws my-scaler -n production \
  --type=merge -p '{"spec":{"timezone":"America/New_York"}}'
```

### No scaling is happening at all

**Common causes:**

1. **Controller not running:**
```bash
kubectl get pods -n kyklos-system
```

2. **TimeWindowScaler paused:**
```bash
kubectl get tws my-scaler -o jsonpath='{.spec.pause}'
# If true, set to false
```

3. **Target not found:**
```bash
kubectl get tws my-scaler -o jsonpath='{.status.conditions[?(@.type=="Ready")]}'
# Check reason
```

4. **Window configuration wrong:**
```bash
# Check current time is actually in a window
kubectl get tws my-scaler -o yaml | grep -A 20 windows
```

### Events aren't showing up

**Check RBAC permissions:**
```bash
kubectl auth can-i create events --as=system:serviceaccount:kyklos-system:kyklos-controller -n production
```

Should return `yes`.

**Check controller logs:**
```bash
kubectl logs -n kyklos-system -l app=kyklos-controller | grep -i event
```

Look for event emission messages or permission errors.

## Best Practices

### How many windows should I define?

**Keep it simple.** Fewer windows are easier to understand and maintain.

**Good:**
```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri]
  start: "09:00"
  end: "17:00"
  replicas: 10
```

**Too complex:**
```yaml
windows:
- days: [Mon]
  start: "09:00"
  end: "09:30"
  replicas: 5
- days: [Mon]
  start: "09:30"
  end: "10:00"
  replicas: 8
# ... 20 more windows
```

### Should I set defaultReplicas to 0?

**Only if you truly want zero traffic capacity outside windows.**

Usually better to set a safe minimum:

```yaml
defaultReplicas: 2  # Can handle basic traffic
```

This ensures:
- Health checks pass
- Service is reachable
- Can handle unexpected off-hours traffic

### How do I test a new TimeWindowScaler?

**Start with pause enabled:**

```yaml
spec:
  pause: true
  # ... rest of config
```

**Steps:**
1. Apply with `pause: true`
2. Check `status.effectiveReplicas` matches expectations
3. Wait through a few window boundaries
4. Verify `status.currentWindow` changes correctly
5. When confident, set `pause: false`

This lets you validate the configuration without affecting the deployment.

## Next Steps

- [Concepts](CONCEPTS.md) - Deep dive into window matching and computation
- [Operations](OPERATIONS.md) - Production monitoring and alerts
- [Troubleshooting](TROUBLESHOOTING.md) - Detailed symptom-based solutions
- [API Reference](../api/CRD-SPEC.md) - Complete field documentation
