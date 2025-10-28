# Kyklos Flake Prevention Policy

## Overview

Test flakes undermine confidence in the test suite and slow development velocity. This document defines policies and practices to prevent, detect, and eliminate flaky tests in the Kyklos project.

## Flake Definition

A test is considered **flaky** if it:
- Passes and fails with the same code
- Produces different results on repeated runs
- Depends on timing, ordering, or external state
- Fails intermittently without code changes

### Flake Categories

1. **Timing Flakes**: Race conditions, sleep-based waits, timeout issues
2. **State Flakes**: Shared state, incomplete cleanup, test ordering
3. **Resource Flakes**: Port conflicts, file system issues, quota limits
4. **Network Flakes**: API timeouts, DNS issues, connection resets
5. **Concurrency Flakes**: Parallel test interference, data races

## Flake Budget

### Maximum Acceptable Flake Rates

| Test Level | Target | Maximum | Action Threshold |
|------------|--------|---------|------------------|
| Unit Tests | 0% | 0% | Any flake = P0 bug |
| Envtest | 0% | 0.05% | >0.05% = Block release |
| E2E Tests | 0.1% | 0.5% | >0.5% = Block release |
| Overall | 0.05% | 0.1% | >0.1% = Block all merges |

### Measurement Period
- Calculated over rolling 7-day window
- Minimum 100 test runs for statistical significance
- Exclude infrastructure failures from flake rate

## Prevention Strategies

### 1. Deterministic Time Control

**Requirement**: All time-dependent tests MUST use controlled time sources.

```go
// GOOD: Controlled time
type TestClock struct {
    mu   sync.Mutex
    now  time.Time
}

func (c *TestClock) Now() time.Time {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.now
}

func (c *TestClock) Set(t time.Time) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.now = t
}

// BAD: System time
func TestSomething(t *testing.T) {
    start := time.Now() // NEVER do this
}
```

### 2. Explicit Waiting

**Requirement**: Use polling with timeout instead of fixed delays.

```go
// GOOD: Poll with timeout
func WaitForReplicas(t *testing.T, expected int32, timeout time.Duration) {
    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for time.Now().Before(deadline) {
        if getReplicas() == expected {
            return
        }
        <-ticker.C
    }
    t.Fatalf("Replicas did not reach %d within %v", expected, timeout)
}

// BAD: Fixed sleep
func TestScale(t *testing.T) {
    scaleDeployment(10)
    time.Sleep(5 * time.Second) // NEVER do this
    assert(getReplicas() == 10)
}
```

### 3. Test Isolation

**Requirement**: Each test MUST be independent and idempotent.

```go
// GOOD: Isolated namespace per test
func TestController(t *testing.T) {
    namespace := fmt.Sprintf("test-%s-%d", t.Name(), time.Now().UnixNano())
    defer deleteNamespace(namespace)
    // test logic
}

// BAD: Shared namespace
var testNamespace = "test" // NEVER share state
```

### 4. Resource Cleanup

**Requirement**: Always clean up resources, even on test failure.

```go
// GOOD: Cleanup in defer
func TestResource(t *testing.T) {
    resource := createResource(t)
    defer func() {
        if err := deleteResource(resource); err != nil {
            t.Logf("Cleanup failed: %v", err)
        }
    }()
    // test logic
}

// BAD: Cleanup at end
func TestResource(t *testing.T) {
    resource := createResource(t)
    // test logic
    deleteResource(resource) // May not run if test fails
}
```

### 5. Retry Logic

**Requirement**: Retry only infrastructure operations, not test assertions.

```go
// GOOD: Retry infrastructure
func createDeploymentWithRetry(t *testing.T, d *Deployment) error {
    var err error
    for i := 0; i < 3; i++ {
        err = createDeployment(d)
        if err == nil || !isRetryable(err) {
            return err
        }
        time.Sleep(time.Second * time.Duration(i+1))
    }
    return fmt.Errorf("failed after 3 attempts: %w", err)
}

// BAD: Retry assertions
func TestSomething(t *testing.T) {
    for i := 0; i < 3; i++ {
        if checkCondition() {
            return // NEVER retry test logic
        }
    }
}
```

## Detection Mechanisms

### 1. Automated Detection

```yaml
# CI configuration for flake detection
name: Flake Detection
on:
  schedule:
    - cron: '0 */6 * * *' # Every 6 hours

jobs:
  detect-flakes:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        iteration: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
    steps:
      - uses: actions/checkout@v3
      - name: Run tests repeatedly
        run: |
          for i in {1..10}; do
            make test || echo "Run $i failed" >> failures.txt
          done
      - name: Check for flakes
        run: |
          if [ -f failures.txt ]; then
            echo "Flaky tests detected!"
            exit 1
          fi
```

### 2. Flake Monitoring Dashboard

Track metrics:
- Test pass rate by test name
- Failure patterns (time of day, day of week)
- Correlation with infrastructure events
- Mean time between failures (MTBF)

### 3. Stress Testing

```bash
# Local flake detection script
#!/bin/bash
TEST_NAME=$1
ITERATIONS=${2:-100}

echo "Running $TEST_NAME $ITERATIONS times..."
FAILURES=0

for i in $(seq 1 $ITERATIONS); do
    if ! go test -run "^${TEST_NAME}$" -count=1 > /dev/null 2>&1; then
        FAILURES=$((FAILURES + 1))
        echo "Failure on iteration $i"
    fi
done

echo "Test failed $FAILURES out of $ITERATIONS times"
if [ $FAILURES -gt 0 ]; then
    echo "FLAKY TEST DETECTED!"
    exit 1
fi
```

## Quarantine Process

### 1. Immediate Quarantine

When a flaky test is detected:

```go
// Mark test as flaky
func TestFlakyFeature(t *testing.T) {
    t.Skip("FLAKY: Issue #123 - Intermittent failures in CI")
    // Original test code commented out
}
```

### 2. Issue Creation

Create GitHub issue with:
- Test name and location
- Failure rate (e.g., "3 failures in 100 runs")
- Error messages and stack traces
- Recent commits that might be related
- Label: `test-flake`, `priority-high`

### 3. Root Cause Analysis

Checklist for investigation:
- [ ] Review test for time.Now() usage
- [ ] Check for shared state between tests
- [ ] Verify resource cleanup
- [ ] Look for hardcoded ports/addresses
- [ ] Check for race conditions
- [ ] Review timeout values
- [ ] Analyze CI logs for patterns

### 4. Fix Validation

Before removing from quarantine:
```bash
# Must pass 1000 consecutive runs
./scripts/stress-test.sh TestName 1000

# Must pass in CI for 7 days
# Must have root cause documented
# Must have prevention measure added
```

## Common Anti-Patterns

### 1. Time-Based Anti-Patterns

```go
// BAD: Using wall clock time
func TestTimeout(t *testing.T) {
    start := time.Now()
    doSomething()
    elapsed := time.Since(start)
    if elapsed > 5*time.Second {
        t.Error("took too long")
    }
}

// GOOD: Use fake clock
func TestTimeout(t *testing.T, clock Clock) {
    start := clock.Now()
    doSomething()
    clock.Advance(6 * time.Second)
    elapsed := clock.Since(start)
    if elapsed > 5*time.Second {
        t.Error("took too long")
    }
}
```

### 2. Ordering Anti-Patterns

```go
// BAD: Depends on map iteration order
func TestMap(t *testing.T) {
    m := map[string]int{"a": 1, "b": 2}
    var result []string
    for k := range m {
        result = append(result, k)
    }
    // This will randomly fail!
    assert.Equal(t, []string{"a", "b"}, result)
}

// GOOD: Sort before comparing
func TestMap(t *testing.T) {
    m := map[string]int{"a": 1, "b": 2}
    var result []string
    for k := range m {
        result = append(result, k)
    }
    sort.Strings(result)
    assert.Equal(t, []string{"a", "b"}, result)
}
```

### 3. Resource Anti-Patterns

```go
// BAD: Hardcoded port
func TestServer(t *testing.T) {
    server := StartServer(":8080") // Will fail if port in use
}

// GOOD: Dynamic port
func TestServer(t *testing.T) {
    server := StartServer(":0") // OS assigns free port
    port := server.Port()
}
```

### 4. Async Anti-Patterns

```go
// BAD: No synchronization
func TestAsync(t *testing.T) {
    go doSomethingAsync()
    checkResult() // Race condition!
}

// GOOD: Proper synchronization
func TestAsync(t *testing.T) {
    done := make(chan struct{})
    go func() {
        doSomethingAsync()
        close(done)
    }()
    <-done
    checkResult()
}
```

## Enforcement

### 1. Pre-commit Hooks

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check for time.Now() in tests
if git diff --cached --name-only | grep '_test\.go$' | xargs grep -l 'time\.Now()'; then
    echo "ERROR: time.Now() detected in test files. Use fake clock instead."
    exit 1
fi

# Check for time.Sleep in tests
if git diff --cached --name-only | grep '_test\.go$' | xargs grep -l 'time\.Sleep'; then
    echo "ERROR: time.Sleep detected in test files. Use polling instead."
    exit 1
fi
```

### 2. CI Gates

```yaml
# Required status checks
protection_rules:
  main:
    required_status_checks:
      - "Unit Tests (No Flakes)"
      - "Envtest (No Flakes)"
      - "E2E Tests (Max 0.5% Flake)"
      - "Flake Detection (Last 7 Days)"
```

### 3. Code Review Checklist

Reviewers must verify:
- [ ] No `time.Now()` in tests
- [ ] No `time.Sleep()` for synchronization
- [ ] Resources cleaned up in defer
- [ ] Tests use unique namespaces/names
- [ ] Assertions have timeouts
- [ ] No shared state between tests
- [ ] Deterministic test data

## Recovery Plan

### When Flake Rate Exceeds Budget

1. **Immediate Actions** (Hour 0-2)
   - Block all non-critical merges
   - Page on-call engineer
   - Begin triage of recent changes

2. **Short Term** (Hour 2-24)
   - Quarantine flaky tests
   - Revert recent suspicious changes
   - Deploy with reduced test coverage if critical

3. **Medium Term** (Day 1-7)
   - Root cause analysis for all flakes
   - Fix or permanently disable flaky tests
   - Add flake detection to affected areas

4. **Long Term** (Week 2+)
   - Post-mortem on flake outbreak
   - Update prevention guidelines
   - Add new detection mechanisms
   - Training on flake prevention

## Success Metrics

### Key Performance Indicators

1. **Flake Rate**: <0.1% across all tests
2. **Mean Time to Fix**: <48 hours from detection
3. **Flake Recurrence**: <5% within 30 days
4. **Test Confidence**: >95% developer trust
5. **CI Time**: No increase due to retries

### Monthly Review

- Number of flakes detected
- Number of flakes fixed
- Time spent on flake-related work
- Impact on deployment velocity
- Trends and patterns identified

## Tools and Resources

### Recommended Tools

1. **gotestsum**: Better test output and flake detection
2. **stress**: Run tests under system load
3. **race detector**: `go test -race`
4. **teststat**: Statistical analysis of test results
5. **pprof**: Profile tests for performance issues

### References

- [Google Testing Blog: Flaky Tests](https://testing.googleblog.com/2016/05/flaky-tests-at-google-and-how-we.html)
- [Eliminating Flaky Tests](https://engineering.atspotify.com/2019/11/18/test-flakiness-methods-for-identifying-and-dealing-with-flaky-tests/)
- [Martin Fowler: Eradicating Non-Determinism](https://martinfowler.com/articles/nonDeterminism.html)

## Appendix: Flake Prevention Checklist

Before merging any test:

- [ ] Uses fake/controlled time source
- [ ] No hardcoded delays or sleeps
- [ ] Polls with timeout for async operations
- [ ] Unique resource names (namespace, deployments)
- [ ] Cleanup in defer blocks
- [ ] No shared state between tests
- [ ] Deterministic test data
- [ ] Passes 100 consecutive local runs
- [ ] Clear failure messages
- [ ] Documented timing assumptions