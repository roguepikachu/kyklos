# Events Design

## Event Types

### ScaledUp
**When Fired**: Target replica count increased
**Type**: Normal
**Rate Limit**: Deduplicate within 5 minutes for same replica values

**Message Fields**:
- `from`: Previous replica count
- `to`: New replica count
- `window`: Active window name or "OffHours"
- `reason`: Trigger (WindowEntered, ManualDriftCorrected, HolidayOverride)

**Example Messages**:
- "Scaled up from 2 to 10 replicas (window: BusinessHours)"
- "Scaled up from 0 to 3 replicas (window: MorningRampUp)"
- "Corrected manual drift: scaled from 5 to 10 replicas (window: BusinessHours)"
- "Holiday override: scaled to 12 replicas (treat-as-open mode)"

### ScaledDown
**When Fired**: Target replica count decreased
**Type**: Normal
**Rate Limit**: Deduplicate within 5 minutes for same replica values

**Message Fields**:
- `from`: Previous replica count
- `to`: New replica count
- `window`: Active window name or "OffHours"
- `gracePeriod`: If grace period was applied
- `reason`: Trigger (WindowExited, ManualDriftCorrected, MaintenanceWindow)

**Example Messages**:
- "Scaled down from 10 to 2 replicas (window: OffHours)"
- "Scaled down from 10 to 5 replicas after 300s grace period (window: EveningWindDown)"
- "Corrected manual drift: scaled from 15 to 10 replicas (window: BusinessHours)"
- "Maintenance window: reduced from 8 to 2 replicas"

### ScalingSkipped
**When Fired**: Scaling needed but prevented by pause or error
**Type**: Normal
**Rate Limit**: Once per reconcile cycle

**Message Fields**:
- `current`: Current replica count
- `desired`: Computed desired replicas
- `reason`: Why skipped (Paused, TargetNotFound, UpdateFailed)

**Example Messages**:
- "Scaling skipped due to pause: current=7, desired=10"
- "Scaling skipped: target Deployment not found (desired=10)"
- "Scaling skipped after 3 failed attempts: current=5, desired=10"

### WindowOverride
**When Fired**: Holiday mode changes normal window behavior
**Type**: Normal
**Rate Limit**: Once per day per holiday

**Message Fields**:
- `date`: Holiday date (YYYY-MM-DD)
- `mode`: Holiday mode (treat-as-closed, treat-as-open)
- `replicas`: Applied replica count
- `normalWindow`: What window would apply without holiday

**Example Messages**:
- "Holiday 2025-12-25: treating as closed, using 0 replicas (would be BusinessHours)"
- "Holiday 2025-01-01: treating as open, using 12 replicas (maximum configured)"
- "Holiday 2025-07-04: ignoring holiday, using normal window (BusinessHours, 10 replicas)"

### MissingTarget
**When Fired**: Target Deployment not found
**Type**: Warning
**Rate Limit**: Every 5 minutes while missing

**Message Fields**:
- `target`: Deployment namespace/name
- `action`: What controller would do if target existed

**Example Messages**:
- "Target Deployment production/webapp not found (would scale to 10 replicas)"
- "Target Deployment data-pipeline/batch-processor not found (would scale to 0 replicas)"

### InvalidSchedule
**When Fired**: Configuration validation fails
**Type**: Warning
**Rate Limit**: Once per configuration

**Message Fields**:
- `error`: Validation error details
- `field`: Which field failed validation

**Example Messages**:
- "Invalid schedule: window start (25:00) does not match HH:MM format"
- "Invalid schedule: window has start==end (10:00-10:00)"
- "Invalid schedule: timezone 'Mars/Olympus_Mons' not found in IANA database"
- "Invalid schedule: no days specified for window"

### ConfigurationUpdated
**When Fired**: Spec changed (generation incremented)
**Type**: Normal
**Rate Limit**: Once per generation

**Message Fields**:
- `generation`: New generation number
- `changes`: Summary of what changed

**Example Messages**:
- "Configuration updated to generation 5: window times changed"
- "Configuration updated to generation 3: timezone changed to Asia/Kolkata"
- "Configuration updated to generation 7: holidays enabled (treat-as-closed)"
- "Configuration updated to generation 2: grace period set to 300 seconds"

### GracePeriodStarted
**When Fired**: Grace period timer begins
**Type**: Normal
**Rate Limit**: Once per grace period

**Message Fields**:
- `duration`: Grace period in seconds
- `expiresAt`: When grace period ends (local time)
- `targetReplicas`: Replicas after grace expires

**Example Messages**:
- "Grace period started: 300s before scaling to 2 replicas (expires at 17:05:00 IST)"
- "Grace period started: 120s before scaling to 0 replicas (expires at 19:02:00 IST)"

### GracePeriodCancelled
**When Fired**: Grace period interrupted by scale up
**Type**: Normal
**Rate Limit**: Once per cancellation

**Message Fields**:
- `reason`: Why cancelled
- `newReplicas`: New target replica count

**Example Messages**:
- "Grace period cancelled: scaling up to 10 replicas for BusinessHours window"
- "Grace period cancelled: manual override to 15 replicas"

### HolidayDetected
**When Fired**: Holiday affects window evaluation
**Type**: Normal
**Rate Limit**: Once per day

**Message Fields**:
- `date`: Holiday date
- `source`: ConfigMap name
- `mode`: How holiday is handled

**Example Messages**:
- "Holiday detected for 2025-12-25 from company-holidays ConfigMap (mode: treat-as-closed)"
- "Holiday detected for 2025-01-01 from company-holidays ConfigMap (mode: treat-as-open)"

## Event Emission Rules

### Deduplication
- Same event type + same field values within 5 minutes = skip
- Exception: Warning events always emitted

### Batching
- Multiple events in single reconcile: emit all
- Order: Warnings first, then Normal events

### Rate Limiting Guidance
```
Per TimeWindowScaler:
- Maximum 20 events per minute
- Maximum 100 events per hour
- Warning events exempt from rate limiting
```

### Event Correlation
Events include these correlation fields:
- `controller`: "kyklos-controller"
- `tws`: TimeWindowScaler name
- `target`: Target Deployment name
- `namespace`: Namespace

## Example Event Sequences

### Morning Window Entry (09:00 IST)
1. `WindowOverride`: "Holiday 2025-01-26: ignoring holiday, using normal window"
2. `ScaledUp`: "Scaled up from 2 to 10 replicas (window: BusinessHours)"

### Evening Window Exit with Grace (17:00 IST)
1. `GracePeriodStarted`: "Grace period started: 300s before scaling to 2 replicas"
2. (After 300s) `ScaledDown`: "Scaled down from 10 to 2 replicas after 300s grace period"

### Manual Intervention Detected
1. `ScaledUp`: "Corrected manual drift: scaled from 15 to 10 replicas (window: BusinessHours)"

### Holiday Closed Mode
1. `HolidayDetected`: "Holiday detected for 2025-12-25 from company-holidays ConfigMap"
2. `WindowOverride`: "Holiday 2025-12-25: treating as closed, using 0 replicas"
3. `ScaledDown`: "Scaled down from 10 to 0 replicas (window: OffHours)"

### Configuration Change
1. `ConfigurationUpdated`: "Configuration updated to generation 5: window times changed"
2. `ScaledUp`: "Scaled up from 2 to 10 replicas (window: BusinessHours)"

### Target Missing
1. `MissingTarget`: "Target Deployment production/webapp not found (would scale to 10)"
2. (Every 5 minutes) `MissingTarget`: "Target Deployment production/webapp not found"

### Invalid Configuration
1. `InvalidSchedule`: "Invalid schedule: timezone 'Invalid/Zone' not found"
2. `ScalingSkipped`: "Scaling skipped due to invalid configuration"

## Event Storage and Retention

### Event Object Fields
```yaml
involvedObject:
  apiVersion: kyklos.io/v1alpha1
  kind: TimeWindowScaler
  name: webapp-office-hours
  namespace: production
type: Normal  # or Warning
reason: ScaledUp  # Event type
message: "Scaled up from 2 to 10 replicas (window: BusinessHours)"
source:
  component: kyklos-controller
firstTimestamp: "2025-01-26T09:00:00+05:30"
lastTimestamp: "2025-01-26T09:00:00+05:30"
count: 1
```

### Retention
- Events retained per Kubernetes cluster policy (typically 1 hour)
- Controller logs preserve event history for debugging
- Metrics track event emission rates

## Integration with Observability

### Metrics from Events
- `kyklos_events_total{type="ScaledUp|ScaledDown|...", tws="name"}`
- `kyklos_scaling_operations_total{direction="up|down", tws="name"}`
- `kyklos_grace_periods_total{state="started|completed|cancelled"}`

### Log Correlation
Each event emission generates a structured log entry with:
- Event type and fields
- Correlation ID matching reconcile loop
- Timestamp with timezone