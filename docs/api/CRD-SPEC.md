# TimeWindowScaler CRD Specification

## Resource Identity

- **Group**: `kyklos.io`
- **Kind**: `TimeWindowScaler`
- **Version**: `v1alpha1`
- **Scope**: Namespaced
- **Plural**: `timewindowscalers`
- **Singular**: `timewindowscaler`
- **ShortNames**: `tws`

## Spec Definition

### spec.targetRef (required)
Reference to the target resource to scale.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `apiVersion` | string | `apps/v1` | API version of the target resource |
| `kind` | string | required | Resource kind. Must be `Deployment` in v1alpha1 |
| `name` | string | required | Name of the target resource |
| `namespace` | string | object namespace | Namespace of the target resource |

**Validation Rules**:
- `kind` must equal `Deployment` (enforced by admission webhook)
- `name` must be non-empty
- If `namespace` is specified, it must equal the TimeWindowScaler's namespace

### spec.timezone (required)
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `timezone` | string | required | IANA timezone identifier (e.g., `America/New_York`) |

**Semantics**: All time calculations use this timezone with full DST awareness via IANA rules.

### spec.defaultReplicas
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `defaultReplicas` | int32 | `0` | Replica count when no windows match |

**Validation**: Must be >= 0

### spec.windows (required)
Array of time windows defining when and how to scale.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `days` | []string | required | Days of week when window applies |
| `start` | string | required | Start time in HH:MM format (inclusive) |
| `end` | string | required | End time in HH:MM format (exclusive) |
| `replicas` | int32 | required | Desired replica count during this window |

**Validation Rules**:
- `windows` array must have at least 1 element
- `days` must contain at least one valid day enum: `Mon`, `Tue`, `Wed`, `Thu`, `Fri`, `Sat`, `Sun`
- `start` and `end` must match pattern `^([0-1][0-9]|2[0-3]):[0-5][0-9]$`
- `start` must not equal `end` (rejected at runtime)
- `replicas` must be >= 0

**Semantics**:
- Start time is inclusive, end time is exclusive
- If `end` < `start`, window crosses midnight into the next calendar day
- Overlapping windows allowed; last matching window in array wins (precedence by position)

### spec.holidays (optional)
Holiday handling configuration.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mode` | string | `ignore` | How to handle holidays |
| `sourceRef.name` | string | optional | ConfigMap name containing holiday dates |

**Mode Enum Values**:
- `ignore`: Process windows normally on holidays
- `treat-as-closed`: No windows match on holiday dates (uses defaultReplicas)
- `treat-as-open`: Synthetic window with replicas = max(all defined window replicas)

**ConfigMap Format**: Keys must be ISO dates `yyyy-mm-dd`, values are ignored

### spec.gracePeriodSeconds
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `gracePeriodSeconds` | int32 | `0` | Delay before applying downscale |

**Semantics**: Only applies when transitioning to fewer replicas. Timer starts when leaving a window state.

### spec.pause
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `pause` | bool | `false` | Suspend target modifications |

**Semantics**: When true, controller computes desired state and updates status but never writes to target.

## Status Definition

### status.currentWindow
| Field | Type | Description |
|-------|------|-------------|
| `currentWindow` | string | Label identifying active window |

**Values**:
- `BusinessHours`: Standard daytime window
- `OffHours`: Outside all windows
- `Custom-<hash>`: Custom window identified by hash of configuration

### status.effectiveReplicas
| Field | Type | Description |
|-------|------|-------------|
| `effectiveReplicas` | int32 | Currently desired replica count |

### status.lastScaleTime
| Field | Type | Description |
|-------|------|-------------|
| `lastScaleTime` | string | RFC3339 timestamp of last scale operation |

### status.targetObservedReplicas
| Field | Type | Description |
|-------|------|-------------|
| `targetObservedReplicas` | int32 | Last observed replica count of target |

### status.observedGeneration
| Field | Type | Description |
|-------|------|-------------|
| `observedGeneration` | int64 | Last processed spec generation |

**Semantics**: Increments when spec changes. Status is stale if != metadata.generation.

### status.conditions
Standard Kubernetes condition array.

**Condition Types**:

| Type | Status | Reason | Description |
|------|--------|--------|-------------|
| `Ready` | True | `Reconciled` | Target matches desired state |
| `Ready` | False | `TargetMismatch` | Target differs from desired (manual drift or pause) |
| `Ready` | False | `TargetNotFound` | Target resource doesn't exist |
| `Reconciling` | True | `ConfigurationChange` | Processing spec change |
| `Reconciling` | True | `WindowTransition` | Transitioning between windows |
| `Reconciling` | False | `Stable` | No ongoing reconciliation |
| `Degraded` | True | `InvalidTimezone` | Timezone cannot be resolved |
| `Degraded` | True | `HolidaySourceMissing` | ConfigMap for holidays not found |
| `Degraded` | True | `InvalidConfiguration` | Configuration validation failed |
| `Degraded` | False | `OperationalNormal` | No degradation |

## Deterministic Behavioral Rules

### Window Matching Algorithm
1. Convert current UTC time to local time using IANA timezone rules
2. Evaluate each window in spec.windows array order
3. For each window:
   - Check if current day matches any day in window.days
   - Check if current time >= start AND < end (accounting for midnight crossing)
   - If match found, continue checking remaining windows
4. Return last matching window or defaultReplicas if none match

### Cross-Midnight Handling
When `end` < `start`:
- Window spans from start time on listed day to end time on following calendar day
- Example: Friday 22:00 to 02:00 matches Friday 22:00-23:59 and Saturday 00:00-01:59

### Holiday Processing
1. Check if current date exists as key in holiday ConfigMap
2. If holiday detected:
   - `ignore`: Continue normal window matching
   - `treat-as-closed`: Return defaultReplicas immediately
   - `treat-as-open`: Return max(all window.replicas values)

### Grace Period Application
1. Grace only applies when effectiveReplicas decreases
2. Timer starts when leaving higher-replica state
3. During grace, maintain previous higher replica count
4. After grace expires, apply new lower replica count

### Manual Drift Correction
1. On each reconcile, compare targetObservedReplicas with effectiveReplicas
2. If different and pause=false, update target to effectiveReplicas
3. If pause=true, observe drift but take no action

### Pause Semantics
When pause=true:
1. Continue computing effectiveReplicas
2. Update all status fields
3. Set Ready condition based on alignment
4. Skip actual target modification
5. Emit events describing what would happen

## Edge Case Behaviors

1. **Cross-midnight Friday 22:00-02:00, current Saturday 01:00**
   - Matches because window extends into Saturday morning

2. **Overlapping windows 09:00-12:00 (2 replicas) and 11:00-13:00 (4 replicas)**
   - At 11:30, both match but second window (4 replicas) wins

3. **Start equals end (10:00-10:00)**
   - Runtime validation error: "Invalid window: start must not equal end"

4. **Invalid timezone "Mars/Olympus_Mons"**
   - Degraded=True with reason InvalidTimezone
   - Uses defaultReplicas as fallback

5. **Holiday with treat-as-closed on matching weekday**
   - Holiday takes precedence, uses defaultReplicas

6. **Holiday with treat-as-open, no windows defined**
   - Uses defaultReplicas (max of empty set defaults to defaultReplicas)

7. **Grace period: leaving window at 14:00, grace=120s**
   - 14:00-14:02: Maintain in-window replicas
   - 14:02+: Apply out-of-window replicas

8. **Pause=true while inside matching window**
   - Status shows correct effectiveReplicas
   - Ready=False with reason TargetMismatch if drift exists
   - No writes to target

9. **Manual scale to 7 while effective=3, pause=false**
   - Next reconcile corrects to 3
   - Event: "Corrected manual drift from 7 to 3 replicas"

10. **Last window crosses midnight, compute next boundary**
    - If in window: next boundary is window.end on next calendar day
    - If before window: next boundary is window.start on same day

## Forward/Backward Compatibility

### v1alpha1 → v1beta1 Migration Path
- **New in v1beta1**:
  - Support for StatefulSet, DaemonSet via targetRef.kind
  - Multiple timezone support per window
  - Percentage-based scaling
  - External metrics integration

- **Deprecations**:
  - None planned; v1alpha1 fields remain stable

- **Conversion Strategy**:
  - Conversion webhook will default new fields
  - v1alpha1 resources auto-upgrade on read
  - No data loss during round-trip conversion

### Storage Version
- v1alpha1 is storage version
- Future versions will require migration controller