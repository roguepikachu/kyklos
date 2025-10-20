# Structured Logging Design

## Log Keys

### Core Identification Keys
- `controller`: Always "kyklos-controller"
- `tws`: TimeWindowScaler name (e.g., "webapp-office-hours")
- `namespace`: Object namespace
- `generation`: Spec generation being processed

### Target Keys
- `target`: Target Deployment name
- `targetNamespace`: Target namespace (if different)
- `targetKind`: Always "Deployment" in v1alpha1

### Time Context Keys
- `tz`: IANA timezone (e.g., "Asia/Kolkata")
- `nowLocal`: Current local time (e.g., "2025-01-26T14:30:00+05:30")
- `nowUTC`: Current UTC time (e.g., "2025-01-26T09:00:00Z")
- `nextBoundary`: Next window boundary in local time
- `boundaryType`: "window_start", "window_end", or "day_boundary"

### Replica State Keys
- `effectiveReplicas`: Computed desired replica count
- `observedReplicas`: Current target replica count
- `previousReplicas`: Previous effective replica count
- `targetSpecReplicas`: Target's spec.replicas value

### Window Keys
- `currentWindow`: Active window label ("BusinessHours", "OffHours", etc.)
- `matchedWindow`: Index of matched window in spec (-1 if none)
- `windowDays`: Days for current window (e.g., "[Mon,Tue,Wed,Thu,Fri]")
- `windowStart`: Start time of current window
- `windowEnd`: End time of current window

### Decision Keys
- `reason`: Why action taken/skipped
- `action`: "scale_up", "scale_down", "maintain", "skip"
- `pause`: Boolean, is controller paused
- `holiday`: Boolean, is today a holiday
- `holidayMode`: "ignore", "treat-as-closed", "treat-as-open"

### Grace Period Keys
- `gracePeriod`: Configured grace period in seconds
- `graceExpiry`: When grace period expires (RFC3339)
- `graceRemaining`: Seconds remaining in grace period

### Error Keys
- `error`: Error message
- `errorType`: "transient", "semantic", "validation"
- `retryCount`: Number of retries attempted
- `backoffSeconds`: Next retry delay

## Log Level Policy

### Info Level
- Successful scale operations
- Window transitions
- Grace period starts/completions
- Holiday detection
- Configuration updates
- Reconcile completion with next requeue time

### Warning Level
- Target not found
- Holiday ConfigMap missing
- Manual drift detected
- Configuration validation issues (non-fatal)
- Rate limiting encountered

### Error Level
- Invalid timezone
- Invalid configuration (fatal)
- Repeated update failures (after retries)
- Unexpected API errors

### Debug Level
- Window evaluation details
- Time calculation steps
- Cache hits/misses
- Patch/update details
- Requeue calculations

## Sample Log Lines

### Reconcile Start
```json
{
  "level": "info",
  "msg": "Starting reconcile",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "generation": 5,
  "tz": "Asia/Kolkata",
  "nowLocal": "2025-01-26T09:00:00+05:30"
}
```

### Window Evaluation
```json
{
  "level": "debug",
  "msg": "Evaluating time windows",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "nowLocal": "2025-01-26T09:00:00+05:30",
  "dayOfWeek": "Sunday",
  "windowCount": 4,
  "matchedWindow": 1,
  "currentWindow": "BusinessHours",
  "effectiveReplicas": 10
}
```

### Scale Up Decision
```json
{
  "level": "info",
  "msg": "Scaling up deployment",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "target": "webapp",
  "action": "scale_up",
  "observedReplicas": 2,
  "effectiveReplicas": 10,
  "currentWindow": "BusinessHours",
  "reason": "Entered BusinessHours window"
}
```

### Scale Down with Grace
```json
{
  "level": "info",
  "msg": "Starting grace period before scale down",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "target": "webapp",
  "action": "maintain",
  "observedReplicas": 10,
  "effectiveReplicas": 2,
  "gracePeriod": 300,
  "graceExpiry": "2025-01-26T17:05:00+05:30",
  "reason": "Window exit with grace period"
}
```

### Holiday Override
```json
{
  "level": "info",
  "msg": "Holiday detected, overriding window",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "holiday": true,
  "holidayDate": "2025-12-25",
  "holidayMode": "treat-as-closed",
  "normalWindow": "BusinessHours",
  "effectiveReplicas": 0,
  "reason": "Holiday forces closed mode"
}
```

### Manual Drift Correction
```json
{
  "level": "warning",
  "msg": "Correcting manual drift",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "target": "webapp",
  "action": "scale_down",
  "observedReplicas": 15,
  "targetSpecReplicas": 15,
  "effectiveReplicas": 10,
  "reason": "Manual scaling detected"
}
```

### Pause Active
```json
{
  "level": "info",
  "msg": "Scaling skipped due to pause",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "pause": true,
  "observedReplicas": 7,
  "effectiveReplicas": 10,
  "action": "skip",
  "reason": "Controller paused"
}
```

### Target Not Found
```json
{
  "level": "warning",
  "msg": "Target deployment not found",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "target": "webapp",
  "targetNamespace": "production",
  "error": "deployments.apps \"webapp\" not found",
  "errorType": "semantic"
}
```

### Invalid Configuration
```json
{
  "level": "error",
  "msg": "Invalid timezone configuration",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "tz": "Mars/Olympus_Mons",
  "error": "unknown time zone Mars/Olympus_Mons",
  "errorType": "validation",
  "effectiveReplicas": 2,
  "reason": "Using defaultReplicas as fallback"
}
```

### Cross-Midnight Window
```json
{
  "level": "debug",
  "msg": "Evaluating cross-midnight window",
  "controller": "kyklos-controller",
  "tws": "batch-processor-night",
  "namespace": "data-pipeline",
  "windowStart": "22:00",
  "windowEnd": "06:00",
  "nowLocal": "2025-01-26T23:30:00+05:30",
  "inWindow": true,
  "reason": "Current time after start on listed day"
}
```

### Requeue Scheduled
```json
{
  "level": "info",
  "msg": "Reconcile complete, scheduling next run",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "nextBoundary": "2025-01-26T17:00:00+05:30",
  "boundaryType": "window_end",
  "requeueAfter": "7h45m17s",
  "jitter": 17,
  "effectiveReplicas": 10,
  "observedReplicas": 10
}
```

### Update Conflict
```json
{
  "level": "debug",
  "msg": "Update conflict, retrying",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "target": "webapp",
  "error": "Operation cannot be fulfilled on deployments.apps \"webapp\": the object has been modified",
  "errorType": "transient",
  "retryCount": 1,
  "backoffSeconds": 0
}
```

### Rate Limited
```json
{
  "level": "warning",
  "msg": "Rate limited by API server",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "error": "429 Too Many Requests",
  "errorType": "transient",
  "retryCount": 2,
  "backoffSeconds": 30
}
```

### DST Transition
```json
{
  "level": "info",
  "msg": "DST transition detected",
  "controller": "kyklos-controller",
  "tws": "webapp-office-hours",
  "namespace": "production",
  "tz": "America/New_York",
  "transition": "spring_forward",
  "nowLocal": "2025-03-09T03:00:00-04:00",
  "note": "2:00 AM skipped to 3:00 AM"
}
```

## Structured Logging Best Practices

### Consistency Rules
1. Always include `controller`, `tws`, `namespace` in every log
2. Use consistent key names across all log entries
3. Include correlation ID for request tracing
4. Use RFC3339 format for all timestamps

### Performance Considerations
1. Use log sampling for high-frequency debug logs
2. Batch log writes when possible
3. Avoid logging sensitive data (secrets, tokens)
4. Limit string concatenation in hot paths

### Integration with Observability
1. Structure allows easy metric extraction
2. Keys map to Prometheus labels
3. Enable JSON output for log aggregation
4. Include trace IDs when available

## Log Aggregation Queries

### Find All Scale Operations
```
controller="kyklos-controller" AND action=~"scale_.*"
```

### Find Grace Period Activities
```
controller="kyklos-controller" AND graceExpiry=~".+"
```

### Find Configuration Errors
```
controller="kyklos-controller" AND errorType="validation"
```

### Find Manual Drift Corrections
```
controller="kyklos-controller" AND msg="Correcting manual drift"
```

### Find Holiday Overrides
```
controller="kyklos-controller" AND holiday=true
```