# Kyklos Unit Test Plan

## Overview

This document specifies unit tests for the Kyklos Time Window Scaler's pure business logic functions. All tests use controlled time sources and require no Kubernetes API interaction.

## Test Categories

### 1. Time Window Matching
Tests for determining if a given time falls within a configured window.

### 2. Holiday Evaluation
Tests for holiday mode logic and ConfigMap parsing.

### 3. Grace Period Calculations
Tests for grace period state management and expiry.

### 4. Timezone Handling
Tests for timezone conversions and DST transitions.

### 5. Boundary Computation
Tests for calculating next window transition times.

## Detailed Test Cases

### TM-001: Basic Window Match
```yaml
test_id: TM-001
category: unit
function: IsInWindow
description: Match time within regular business hours window
input:
  window:
    days: [Mon, Tue, Wed, Thu, Fri]
    start: "09:00"
    end: "17:00"
  current_time: "2025-03-10T14:30:00"  # Monday 2:30 PM
  timezone: "UTC"
expected:
  result: true
  reason: "14:30 is between 09:00 and 17:00 on Monday"
```

### TM-002: Window Boundary Start (Inclusive)
```yaml
test_id: TM-002
category: unit
function: IsInWindow
description: Test inclusive start boundary
input:
  window:
    days: [Mon]
    start: "09:00"
    end: "17:00"
  current_time: "2025-03-10T09:00:00"  # Exactly 9 AM
  timezone: "UTC"
expected:
  result: true
  reason: "Start time is inclusive"
```

### TM-003: Window Boundary End (Exclusive)
```yaml
test_id: TM-003
category: unit
function: IsInWindow
description: Test exclusive end boundary
input:
  window:
    days: [Mon]
    start: "09:00"
    end: "17:00"
  current_time: "2025-03-10T17:00:00"  # Exactly 5 PM
  timezone: "UTC"
expected:
  result: false
  reason: "End time is exclusive"
```

### TM-004: Cross-Midnight Window Same Day
```yaml
test_id: TM-004
category: unit
function: IsInWindow
description: Cross-midnight window, checking after start
input:
  window:
    days: [Fri]
    start: "22:00"
    end: "06:00"
  current_time: "2025-03-14T23:30:00"  # Friday 11:30 PM
  timezone: "UTC"
expected:
  result: true
  reason: "23:30 is after 22:00 on listed day"
```

### TM-005: Cross-Midnight Window Next Day
```yaml
test_id: TM-005
category: unit
function: IsInWindow
description: Cross-midnight window, checking before end on next day
input:
  window:
    days: [Fri]
    start: "22:00"
    end: "06:00"
  current_time: "2025-03-15T05:30:00"  # Saturday 5:30 AM
  timezone: "UTC"
expected:
  result: true
  reason: "05:30 is before 06:00, Friday window extends to Saturday"
```

### TM-006: Cross-Midnight Window After End
```yaml
test_id: TM-006
category: unit
function: IsInWindow
description: Cross-midnight window, after end time next day
input:
  window:
    days: [Fri]
    start: "22:00"
    end: "06:00"
  current_time: "2025-03-15T06:30:00"  # Saturday 6:30 AM
  timezone: "UTC"
expected:
  result: false
  reason: "06:30 is after 06:00 end time"
```

### TM-007: Wrong Day of Week
```yaml
test_id: TM-007
category: unit
function: IsInWindow
description: Time within hours but wrong day
input:
  window:
    days: [Mon, Wed, Fri]
    start: "09:00"
    end: "17:00"
  current_time: "2025-03-11T14:30:00"  # Tuesday
  timezone: "UTC"
expected:
  result: false
  reason: "Tuesday not in configured days"
```

### TM-008: Full Day Window
```yaml
test_id: TM-008
category: unit
function: IsInWindow
description: Nearly 24-hour window
input:
  window:
    days: [Sat, Sun]
    start: "00:00"
    end: "23:59"
  current_time: "2025-03-15T12:00:00"  # Saturday noon
  timezone: "UTC"
expected:
  result: true
  reason: "Within full day window"
```

### WP-001: Window Precedence - Last Match Wins
```yaml
test_id: WP-001
category: unit
function: GetEffectiveReplicas
description: Multiple overlapping windows, last match wins
input:
  windows:
    - days: [Mon]
      start: "08:00"
      end: "18:00"
      replicas: 5
    - days: [Mon]
      start: "09:00"
      end: "17:00"
      replicas: 10
    - days: [Mon]
      start: "12:00"
      end: "14:00"
      replicas: 15
  current_time: "2025-03-10T13:00:00"  # Monday 1 PM
  timezone: "UTC"
  default_replicas: 1
expected:
  result: 15
  matched_window: "12:00-14:00"
  reason: "All three windows match, last one (15 replicas) wins"
```

### WP-002: Window Precedence - No Match Uses Default
```yaml
test_id: WP-002
category: unit
function: GetEffectiveReplicas
description: No window matches, use default
input:
  windows:
    - days: [Mon, Tue, Wed]
      start: "09:00"
      end: "17:00"
      replicas: 10
  current_time: "2025-03-10T08:30:00"  # Monday 8:30 AM (before window)
  timezone: "UTC"
  default_replicas: 2
expected:
  result: 2
  matched_window: null
  reason: "No window matches, using default"
```

### H-001: Holiday Mode - Treat As Closed
```yaml
test_id: H-001
category: unit
function: ApplyHolidayMode
description: Holiday with treat-as-closed mode
input:
  is_holiday: true
  mode: "treat-as-closed"
  window_replicas: 10
  default_replicas: 1
expected:
  result: 1
  reason: "Holiday with treat-as-closed uses default replicas"
```

### H-002: Holiday Mode - Treat As Open
```yaml
test_id: H-002
category: unit
function: ApplyHolidayMode
description: Holiday with treat-as-open mode
input:
  is_holiday: true
  mode: "treat-as-open"
  all_window_replicas: [3, 5, 10, 8]
  default_replicas: 1
expected:
  result: 10
  reason: "Holiday with treat-as-open uses max window replicas"
```

### H-003: Holiday Mode - Ignore
```yaml
test_id: H-003
category: unit
function: ApplyHolidayMode
description: Holiday with ignore mode uses normal logic
input:
  is_holiday: true
  mode: "ignore"
  window_replicas: 8
  default_replicas: 1
expected:
  result: 8
  reason: "Holiday with ignore mode uses normal window match"
```

### H-004: Parse Holiday ConfigMap
```yaml
test_id: H-004
category: unit
function: IsHoliday
description: Check if date exists in holiday ConfigMap
input:
  configmap_data:
    "2025-01-01": "New Year's Day"
    "2025-07-04": "Independence Day"
    "2025-12-25": "Christmas"
  check_date: "2025-07-04"
  timezone: "America/New_York"
expected:
  result: true
  reason: "2025-07-04 exists in ConfigMap"
```

### GP-001: Grace Period - Scale Up (No Grace)
```yaml
test_id: GP-001
category: unit
function: ApplyGracePeriod
description: Scaling up bypasses grace period
input:
  current_replicas: 5
  new_replicas: 10
  grace_seconds: 300
  grace_expiry: null
  now: "2025-03-10T14:00:00Z"
expected:
  effective_replicas: 10
  new_grace_expiry: null
  reason: "Scale up happens immediately"
```

### GP-002: Grace Period - Scale Down Starts Grace
```yaml
test_id: GP-002
category: unit
function: ApplyGracePeriod
description: First scale down starts grace period
input:
  current_replicas: 10
  new_replicas: 5
  grace_seconds: 300
  grace_expiry: null
  now: "2025-03-10T17:00:00Z"
expected:
  effective_replicas: 10
  new_grace_expiry: "2025-03-10T17:05:00Z"
  reason: "Grace period started, maintaining current replicas"
```

### GP-003: Grace Period - During Grace
```yaml
test_id: GP-003
category: unit
function: ApplyGracePeriod
description: Within grace period, maintain replicas
input:
  current_replicas: 10
  new_replicas: 5
  grace_seconds: 300
  grace_expiry: "2025-03-10T17:05:00Z"
  now: "2025-03-10T17:02:00Z"
expected:
  effective_replicas: 10
  new_grace_expiry: "2025-03-10T17:05:00Z"
  reason: "Still in grace period (3 minutes remaining)"
```

### GP-004: Grace Period - After Expiry
```yaml
test_id: GP-004
category: unit
function: ApplyGracePeriod
description: Grace period expired, apply scale down
input:
  current_replicas: 10
  new_replicas: 5
  grace_seconds: 300
  grace_expiry: "2025-03-10T17:05:00Z"
  now: "2025-03-10T17:05:01Z"
expected:
  effective_replicas: 5
  new_grace_expiry: null
  reason: "Grace period expired, scaling down"
```

### GP-005: Grace Period - Zero Grace
```yaml
test_id: GP-005
category: unit
function: ApplyGracePeriod
description: Zero grace period means immediate scale
input:
  current_replicas: 10
  new_replicas: 5
  grace_seconds: 0
  grace_expiry: null
  now: "2025-03-10T17:00:00Z"
expected:
  effective_replicas: 5
  new_grace_expiry: null
  reason: "Zero grace period, immediate scale down"
```

### TZ-001: Timezone Conversion - EST to UTC
```yaml
test_id: TZ-001
category: unit
function: ConvertToUTC
description: Convert Eastern time window to UTC
input:
  window:
    start: "09:00"
    end: "17:00"
  date: "2025-03-10"  # Not DST
  timezone: "America/New_York"
expected:
  utc_start: "2025-03-10T14:00:00Z"
  utc_end: "2025-03-10T22:00:00Z"
  reason: "EST is UTC-5"
```

### TZ-002: Timezone Conversion - EDT to UTC
```yaml
test_id: TZ-002
category: unit
function: ConvertToUTC
description: Convert Eastern time during DST to UTC
input:
  window:
    start: "09:00"
    end: "17:00"
  date: "2025-07-10"  # During DST
  timezone: "America/New_York"
expected:
  utc_start: "2025-07-10T13:00:00Z"
  utc_end: "2025-07-10T21:00:00Z"
  reason: "EDT is UTC-4"
```

### TZ-003: DST Spring Forward
```yaml
test_id: TZ-003
category: unit
function: GetWindowDuration
description: Window containing DST spring forward
input:
  window:
    start: "01:00"
    end: "04:00"
  date: "2025-03-09"  # DST starts at 2 AM
  timezone: "America/New_York"
expected:
  duration_hours: 2
  reason: "2 AM becomes 3 AM, window is 1 hour shorter"
```

### TZ-004: DST Fall Back
```yaml
test_id: TZ-004
category: unit
function: GetWindowDuration
description: Window containing DST fall back
input:
  window:
    start: "01:00"
    end: "04:00"
  date: "2025-11-02"  # DST ends at 2 AM
  timezone: "America/New_York"
expected:
  duration_hours: 4
  reason: "2 AM happens twice, window is 1 hour longer"
```

### TZ-005: India Standard Time (No DST)
```yaml
test_id: TZ-005
category: unit
function: ConvertToUTC
description: IST timezone with 30-minute offset
input:
  window:
    start: "09:30"
    end: "18:30"
  date: "2025-03-10"
  timezone: "Asia/Kolkata"
expected:
  utc_start: "2025-03-10T04:00:00Z"
  utc_end: "2025-03-10T13:00:00Z"
  reason: "IST is UTC+5:30, no DST"
```

### BC-001: Next Boundary - In Window
```yaml
test_id: BC-001
category: unit
function: GetNextBoundary
description: Currently in window, next boundary is window end
input:
  windows:
    - days: [Mon]
      start: "09:00"
      end: "17:00"
  current_time: "2025-03-10T14:30:00"  # Monday 2:30 PM
  timezone: "UTC"
expected:
  next_boundary: "2025-03-10T17:00:00"
  boundary_type: "window_end"
  reason: "In window until 5 PM today"
```

### BC-002: Next Boundary - Between Windows Same Day
```yaml
test_id: BC-002
category: unit
function: GetNextBoundary
description: Between windows on same day
input:
  windows:
    - days: [Mon]
      start: "09:00"
      end: "12:00"
    - days: [Mon]
      start: "13:00"
      end: "17:00"
  current_time: "2025-03-10T12:30:00"  # Monday 12:30 PM
  timezone: "UTC"
expected:
  next_boundary: "2025-03-10T13:00:00"
  boundary_type: "window_start"
  reason: "Next window starts at 1 PM"
```

### BC-003: Next Boundary - After All Windows Today
```yaml
test_id: BC-003
category: unit
function: GetNextBoundary
description: After all windows for today, next is tomorrow
input:
  windows:
    - days: [Mon, Tue]
      start: "09:00"
      end: "17:00"
  current_time: "2025-03-10T18:00:00"  # Monday 6 PM
  timezone: "UTC"
expected:
  next_boundary: "2025-03-11T09:00:00"
  boundary_type: "window_start"
  reason: "Next window is tomorrow at 9 AM"
```

### BC-004: Next Boundary - Cross-Midnight Active
```yaml
test_id: BC-004
category: unit
function: GetNextBoundary
description: In cross-midnight window
input:
  windows:
    - days: [Fri]
      start: "22:00"
      end: "06:00"
  current_time: "2025-03-14T23:30:00"  # Friday 11:30 PM
  timezone: "UTC"
expected:
  next_boundary: "2025-03-15T06:00:00"
  boundary_type: "window_end"
  reason: "Cross-midnight window ends at 6 AM Saturday"
```

### BC-005: Next Boundary - Weekend to Monday
```yaml
test_id: BC-005
category: unit
function: GetNextBoundary
description: Weekend with no windows, next is Monday
input:
  windows:
    - days: [Mon, Tue, Wed, Thu, Fri]
      start: "09:00"
      end: "17:00"
  current_time: "2025-03-15T10:00:00"  # Saturday 10 AM
  timezone: "UTC"
expected:
  next_boundary: "2025-03-17T09:00:00"
  boundary_type: "window_start"
  reason: "No weekend windows, next is Monday 9 AM"
```

### BC-006: Next Boundary - No Windows
```yaml
test_id: BC-006
category: unit
function: GetNextBoundary
description: No windows defined, use midnight
input:
  windows: []
  current_time: "2025-03-10T14:30:00"
  timezone: "UTC"
expected:
  next_boundary: "2025-03-11T00:00:00"
  boundary_type: "midnight"
  reason: "No windows, check again at midnight"
```

### V-001: Validate Window Times
```yaml
test_id: V-001
category: unit
function: ValidateWindow
description: Validate correct window format
input:
  window:
    start: "09:00"
    end: "17:00"
expected:
  valid: true
  errors: []
```

### V-002: Invalid Time Format
```yaml
test_id: V-002
category: unit
function: ValidateWindow
description: Invalid time format
input:
  window:
    start: "9:00"  # Missing leading zero
    end: "17:00"
expected:
  valid: false
  errors: ["start time must match HH:MM format"]
```

### V-003: Start Equals End
```yaml
test_id: V-003
category: unit
function: ValidateWindow
description: Start and end cannot be equal
input:
  window:
    start: "09:00"
    end: "09:00"
expected:
  valid: false
  errors: ["start and end times cannot be equal"]
```

### V-004: Invalid Timezone
```yaml
test_id: V-004
category: unit
function: ValidateTimezone
description: Invalid IANA timezone
input:
  timezone: "Invalid/Timezone"
expected:
  valid: false
  error: "unknown time zone Invalid/Timezone"
```

### V-005: Valid IANA Timezone
```yaml
test_id: V-005
category: unit
function: ValidateTimezone
description: Valid IANA timezone
input:
  timezone: "Europe/London"
expected:
  valid: true
  location: "Europe/London"
```

## Test Utilities

### FakeClock Implementation
```go
type FakeClock struct {
    current time.Time
    mu      sync.Mutex
}

func (f *FakeClock) Now() time.Time {
    f.mu.Lock()
    defer f.mu.Unlock()
    return f.current
}

func (f *FakeClock) Set(t time.Time) {
    f.mu.Lock()
    defer f.mu.Unlock()
    f.current = t
}

func (f *FakeClock) Advance(d time.Duration) {
    f.mu.Lock()
    defer f.mu.Unlock()
    f.current = f.current.Add(d)
}
```

### Test Helper Functions
```go
// Create deterministic test windows
func TestWindow(days []string, start, end string, replicas int32) Window

// Parse test time with location
func TestTime(str string, loc *time.Location) time.Time

// Assert with descriptive messages
func AssertReplicas(t *testing.T, got, want int32, context string)
```

## Execution Strategy

### Test Organization
```
pkg/timewindow/
├── matcher_test.go      # Window matching logic
├── holiday_test.go      # Holiday evaluation
├── grace_test.go        # Grace period logic
├── timezone_test.go     # Timezone handling
├── boundary_test.go     # Boundary calculations
├── validation_test.go   # Input validation
└── testutil/
    ├── clock.go         # FakeClock implementation
    └── fixtures.go      # Common test data
```

### Parallel Execution
All unit tests run in parallel:
```go
func TestWindowMatch(t *testing.T) {
    t.Parallel()
    // test implementation
}
```

### Benchmark Tests
Include benchmarks for critical paths:
```go
func BenchmarkWindowMatch(b *testing.B) {
    // benchmark window matching performance
}

func BenchmarkBoundaryCalculation(b *testing.B) {
    // benchmark next boundary computation
}
```

## Coverage Requirements

### Minimum Coverage by Package
- `pkg/timewindow`: 100% coverage
- `pkg/holiday`: 95% coverage
- `pkg/grace`: 95% coverage
- `pkg/validation`: 90% coverage

### Edge Cases Must Cover
- All DST transitions (spring/fall)
- Leap years and February 29
- Year boundaries
- Midnight boundaries
- Empty/nil inputs
- Maximum/minimum values

## DST Test Scenarios (Using Fixed Test Dates)

### DST-1: Spring Forward Transition
**Test Date:** 2025-03-09 (Sunday)
**Timezone:** America/New_York
**Fixture:** test/fixtures/dst-spring-2025.yaml
**Scenario:**
- Window: 01:00-04:00 on Sunday
- At 01:30 EST: window should be active
- At 02:30: this time does not exist (jumped to 03:30 EDT)
- At 03:30 EDT: window should be active
- Verify: Window duration is 2 hours, not 3

### DST-2: Fall Back Transition
**Test Date:** 2025-11-02 (Sunday)
**Timezone:** America/New_York
**Fixture:** test/fixtures/dst-fall-2025.yaml
**Scenario:**
- Window: 01:00-04:00 on Sunday
- At 01:30 EDT (first occurrence): window active
- At 01:30 EST (second occurrence after fallback): window still active
- Verify: Window duration is 4 hours, not 3

### DST-3: Cross-Midnight with Spring Forward
**Test Date:** 2025-03-08 22:00 to 2025-03-09 06:00
**Timezone:** America/New_York
**Fixture:** test/fixtures/dst-cross-midnight-2025.yaml
**Scenario:**
- Window: 22:00 Saturday to 06:00 Sunday
- Window spans midnight and DST transition
- Verify: Window remains active across both boundaries
- Verify: Total duration is 7 hours (lost 1 hour to DST)

## Success Criteria

1. All test cases pass with controlled time
2. No test uses `time.Now()` directly
3. Test execution under 1 second total
4. Zero flakes in 10,000 runs
5. Clear failure messages identifying issue