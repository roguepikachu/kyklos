# Requeue Schedule Design

## Overview
The requeue algorithm determines when the controller should next evaluate the TimeWindowScaler to minimize unnecessary reconciliations while ensuring timely responses to window boundaries.

## Next Boundary Calculation

### Algorithm Steps
1. **Load Current Time**: Get current time in configured timezone
2. **Identify Current State**: Determine if currently in a window or not
3. **Find Next Boundaries**: Calculate upcoming window transitions
4. **Select Nearest**: Choose the earliest boundary after now

### Boundary Types
- **Window Start**: When a window becomes active
- **Window End**: When a window becomes inactive
- **Grace Expiry**: When grace period completes
- **Day Boundary**: Midnight for daily re-evaluation

### Cross-Midnight Window Handling
For windows where end < start (e.g., 22:00-06:00):
```
if in_window:
  next_boundary = end_time_tomorrow
else:
  if current_time < end_time_today:
    next_boundary = end_time_today
  else:
    next_boundary = start_time_today_or_tomorrow
```

### Examples with Asia/Kolkata Timezone

#### Example 1: In Business Hours Window
```
Current: Monday 14:30 IST
Window: Mon-Fri 09:00-17:00
State: In window
Next boundary: Monday 17:00 IST (window end)
```

#### Example 2: Before Morning Window
```
Current: Tuesday 07:30 IST
Window: Mon-Fri 09:00-17:00
State: Out of window
Next boundary: Tuesday 09:00 IST (window start)
```

#### Example 3: After Evening Window
```
Current: Wednesday 19:00 IST
Window: Mon-Fri 09:00-17:00
State: Out of window
Next boundary: Thursday 09:00 IST (next day window start)
```

#### Example 4: Cross-Midnight Window Active
```
Current: Friday 23:30 IST
Window: Fri 22:00-02:00 (crosses midnight)
State: In window
Next boundary: Saturday 02:00 IST (window end on next calendar day)
```

#### Example 5: Weekend with No Windows
```
Current: Saturday 10:00 IST
Windows: Mon-Fri 09:00-17:00 only
State: Out of window
Next boundary: Monday 09:00 IST (first window next week)
```

## Jitter Application

### Purpose
Prevent thundering herd when multiple TimeWindowScalers have identical schedules.

### Algorithm
```go
baseDelay = nextBoundary - now
jitter = random(5, 25) // seconds
withJitter = baseDelay + jitter
```

### Quantization
Round to 10-second boundaries for predictability:
```go
quantized = floor(withJitter / 10) * 10
```

### Constraints
```go
minimum = 30 seconds  // Prevent tight loops
maximum = 24 hours    // Prevent excessive delays
final = max(minimum, min(maximum, quantized))
```

## Requeue Examples

### Normal Window Transition
```
Next boundary: 17:00:00 IST
Current time: 09:15:23 IST
Base delay: 7h 44m 37s
Jitter: 17 seconds
With jitter: 7h 44m 54s
Quantized: 7h 44m 50s (27890 seconds)
Final: 7h 44m 50s
```

### Near Boundary
```
Next boundary: 09:00:00 IST
Current time: 08:59:45 IST
Base delay: 15 seconds
Jitter: 12 seconds
With jitter: 27 seconds
Quantized: 20 seconds
Final: 30 seconds (minimum applied)
```

### Distant Boundary
```
Next boundary: Monday 09:00 IST
Current time: Friday 17:30 IST
Base delay: 63h 30m
Jitter: 21 seconds
Final: 24 hours (maximum applied)
```

## Error Backoff Settings

### Transient Errors
Exponential backoff with jitter:
```
Attempt 1: 30 seconds
Attempt 2: 1 minute
Attempt 3: 2 minutes
Attempt 4: 4 minutes
Attempt 5+: 5 minutes (capped)
```

### Semantic Errors
Fixed interval:
```
All attempts: 5 minutes
```

### Conflict Errors
Immediate retry:
```
All attempts: 0 seconds (immediate requeue)
```

## Grace Period Integration

### During Grace Period
```
if graceExpiryTime > now:
  nextRequeue = graceExpiryTime + jitter(0, 5)
else:
  nextRequeue = nextWindowBoundary + jitter(5, 25)
```

### Example
```
Grace expires: 17:05:00 IST
Current time: 17:00:30 IST
Requeue: 4m 30s + 3s jitter = 4m 30s (quantized)
```

## Holiday Considerations

### Holiday with treat-as-closed
```
if holiday && mode == "treat-as-closed":
  # Skip all windows today
  nextRequeue = tomorrow_00:00 + jitter
```

### Holiday with treat-as-open
```
if holiday && mode == "treat-as-open":
  # Check again tomorrow
  nextRequeue = tomorrow_00:00 + jitter
```

## DST Transition Handling

### Spring Forward (2:00 → 3:00)
```
Window: 01:00-04:00
At 00:30: Next boundary = 01:00 (30 minutes)
At 01:30: Next boundary = 04:00 (2.5 hours wall clock, 1.5 hours actual)
```

### Fall Back (2:00 → 1:00)
```
Window: 01:00-03:00
At 00:30: Next boundary = 01:00 (30 minutes)
At 01:30 (first): Next boundary = 03:00 (2.5 hours)
At 01:30 (second): Next boundary = 03:00 (1.5 hours)
```

## Implementation Pseudocode

```
function calculateRequeue(tws, now):
  timezone = loadLocation(tws.spec.timezone)
  localNow = now.In(timezone)

  # Find current window state
  currentWindow = evaluateWindows(tws, localNow)
  effectiveReplicas = getEffectiveReplicas(currentWindow)

  # Check grace period
  if tws.status.gracePeriodExpiry > now:
    return tws.status.gracePeriodExpiry - now + smallJitter()

  # Find next boundary
  boundaries = []
  for window in tws.spec.windows:
    boundaries.append(nextStart(window, localNow))
    boundaries.append(nextEnd(window, localNow))

  if boundaries.empty():
    # No windows, check again tomorrow
    boundaries.append(tomorrow_midnight)

  nextBoundary = min(boundaries)

  # Apply jitter and constraints
  baseDelay = nextBoundary - now
  jitter = random(5, 25)
  withJitter = baseDelay + jitter
  quantized = floor(withJitter / 10) * 10

  return max(30, min(24*hour, quantized))
```

## Optimization Strategies

### Window Coalescing
If multiple boundaries occur within 1 minute:
```
Coalesce to earliest boundary
Single requeue handles all transitions
```

### Holiday Prefetch
```
If approaching midnight:
  Prefetch tomorrow's holiday status
  Cache for smooth transition
```

### Steady State Detection
```
If no windows for next 7 days:
  Requeue daily at midnight + jitter
  Avoid computing distant boundaries
```

## Monitoring and Metrics

### Requeue Metrics
- `kyklos_requeue_duration_seconds{tws}`: Time until next reconcile
- `kyklos_requeue_reason{reason}`: Why requeued (boundary, error, grace)
- `kyklos_requeue_jitter_seconds{tws}`: Applied jitter amount

### Alerting Thresholds
- Requeue > 25 hours: Possible calculation error
- Requeue < 10 seconds repeatedly: Possible loop
- Error backoff at maximum: Persistent failure

## Testing Scenarios

### Scenario 1: Multiple Windows Same Day
```
Windows: 09:00-12:00, 14:00-17:00
Current: 10:00
Next: 12:00 (first window end)
```

### Scenario 2: Overlapping Windows
```
Windows: 09:00-15:00, 12:00-17:00
Current: 13:00 (in both)
Next: 15:00 (first window end, still in second)
```

### Scenario 3: No Windows Match
```
Windows: Mon-Fri 09:00-17:00
Current: Sat 14:00
Next: Mon 09:00 (with 24-hour cap)
```

### Scenario 4: Grace Period Active
```
Current: 17:00:30 (just left window)
Grace: 300 seconds
Next: 17:05:30 (grace expiry)
```

### Scenario 5: Rapid Window Changes
```
Windows: Every hour different replicas
Current: 14:59:50
Next: 15:00:00 (with 30-second minimum)
```