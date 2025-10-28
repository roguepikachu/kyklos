# Kyklos Assertion Patterns

## Overview

This document defines canonical assertion patterns for the Kyklos test suite. Consistent assertions improve test readability, debugging, and maintenance.

## Core Principles

1. **Descriptive Failures**: Every assertion should produce a clear error message
2. **Context Included**: Include relevant values in failure messages
3. **Type Safety**: Use type-specific assertions over generic comparisons
4. **Eventual Consistency**: Use Eventually/Consistently for async operations
5. **No Magic Numbers**: Define constants for expected values

## Unit Test Assertions

### Time Window Matching

```go
// Good: Descriptive assertion with context
func AssertInWindow(t *testing.T, window Window, checkTime time.Time, expected bool) {
    t.Helper()
    actual := IsInWindow(window, checkTime)
    if actual != expected {
        t.Errorf("IsInWindow(%v, %s) = %v, want %v\nWindow: %s-%s on %v\nTime: %s (%s)",
            window, checkTime.Format(time.RFC3339), actual, expected,
            window.Start, window.End, window.Days,
            checkTime.Format("15:04"), checkTime.Weekday())
    }
}

// Usage
AssertInWindow(t, window, testTime, true)
```

### Replica Count Validation

```go
// Good: Type-specific with tolerance
func AssertReplicas(t *testing.T, got, want int32, context string) {
    t.Helper()
    if got != want {
        t.Errorf("%s: replicas = %d, want %d", context, got, want)
    }
}

// With tolerance for timing-sensitive tests
func AssertReplicasWithinTolerance(t *testing.T, got, want, tolerance int32, context string) {
    t.Helper()
    diff := got - want
    if diff < 0 {
        diff = -diff
    }
    if diff > tolerance {
        t.Errorf("%s: replicas = %d, want %d ± %d (difference: %d)",
            context, got, want, tolerance, got-want)
    }
}
```

### Holiday Evaluation

```go
// Good: Clear boolean assertion
func AssertHoliday(t *testing.T, date string, holidays map[string]string, expected bool) {
    t.Helper()
    actual := IsHoliday(date, holidays)
    if actual != expected {
        if expected {
            t.Errorf("Expected %s to be a holiday, but it wasn't. Holidays: %v", date, holidays)
        } else {
            t.Errorf("Expected %s NOT to be a holiday, but it was. Found: %s", date, holidays[date])
        }
    }
}
```

### Grace Period State

```go
// Good: Complex state assertion
func AssertGracePeriod(t *testing.T, state GracePeriodState, expectedActive bool, expectedExpiry *time.Time) {
    t.Helper()

    if state.Active != expectedActive {
        t.Errorf("Grace period active = %v, want %v", state.Active, expectedActive)
    }

    if expectedExpiry == nil && state.Expiry != nil {
        t.Errorf("Grace period expiry = %v, want nil", state.Expiry)
    } else if expectedExpiry != nil && state.Expiry == nil {
        t.Errorf("Grace period expiry = nil, want %v", expectedExpiry)
    } else if expectedExpiry != nil && state.Expiry != nil {
        diff := state.Expiry.Sub(*expectedExpiry).Abs()
        if diff > time.Second {
            t.Errorf("Grace period expiry = %v, want %v (diff: %v)",
                state.Expiry, expectedExpiry, diff)
        }
    }
}
```

### Next Boundary Calculation

```go
// Good: Time assertion with tolerance
func AssertNextBoundary(t *testing.T, actual, expected time.Time, tolerance time.Duration, context string) {
    t.Helper()
    diff := actual.Sub(expected).Abs()
    if diff > tolerance {
        t.Errorf("%s: next boundary = %v, want %v ± %v (diff: %v)",
            context,
            actual.Format(time.RFC3339),
            expected.Format(time.RFC3339),
            tolerance, diff)
    }
}
```

### Validation Errors

```go
// Good: Multi-error assertion
func AssertValidationErrors(t *testing.T, errs []error, expectedMessages ...string) {
    t.Helper()

    if len(errs) != len(expectedMessages) {
        t.Fatalf("Got %d errors, want %d\nErrors: %v\nExpected: %v",
            len(errs), len(expectedMessages), errs, expectedMessages)
    }

    for i, err := range errs {
        if err == nil {
            t.Errorf("Error[%d] = nil, want %q", i, expectedMessages[i])
            continue
        }
        if !strings.Contains(err.Error(), expectedMessages[i]) {
            t.Errorf("Error[%d] = %q, want to contain %q", i, err.Error(), expectedMessages[i])
        }
    }
}
```

## Envtest Assertions

### Deployment State

```go
import (
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
)

// Good: Eventually pattern for async operations
func AssertDeploymentReplicas(ctx context.Context, client client.Client, name, namespace string, expected int32) {
    Eventually(func(g Gomega) {
        deployment := &appsv1.Deployment{}
        err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, deployment)
        g.Expect(err).NotTo(HaveOccurred(), "Failed to get deployment %s/%s", namespace, name)
        g.Expect(deployment.Spec.Replicas).To(Equal(&expected),
            "Deployment %s/%s should have %d replicas but has %d",
            namespace, name, expected, *deployment.Spec.Replicas)
    }, 30*time.Second, 1*time.Second).Should(Succeed())
}
```

### Status Conditions

```go
// Good: Comprehensive condition assertion
func AssertCondition(t *testing.T, tws *v1alpha1.TimeWindowScaler,
    condType string, status metav1.ConditionStatus, reason string) {

    cond := meta.FindStatusCondition(tws.Status.Conditions, condType)
    if cond == nil {
        t.Fatalf("Condition %s not found in status. Available: %v",
            condType, tws.Status.Conditions)
    }

    if cond.Status != status {
        t.Errorf("Condition %s status = %v, want %v\nMessage: %s\nReason: %s",
            condType, cond.Status, status, cond.Message, cond.Reason)
    }

    if reason != "" && cond.Reason != reason {
        t.Errorf("Condition %s reason = %q, want %q", condType, cond.Reason, reason)
    }
}

// Gomega style
func AssertConditionGomega(g Gomega, tws *v1alpha1.TimeWindowScaler,
    condType string, status metav1.ConditionStatus) {

    cond := meta.FindStatusCondition(tws.Status.Conditions, condType)
    g.Expect(cond).NotTo(BeNil(), "Condition %s should exist", condType)
    g.Expect(cond.Status).To(Equal(status),
        "Condition %s should be %v", condType, status)
}
```

### Event Verification

```go
// Good: Event assertion with pattern matching
func AssertEvent(t *testing.T, recorder *record.FakeRecorder,
    eventType, reason, messagePattern string) {

    select {
    case event := <-recorder.Events:
        parts := strings.SplitN(event, " ", 3)
        if len(parts) != 3 {
            t.Fatalf("Malformed event: %q", event)
        }

        actualType, actualReason, actualMessage := parts[0], parts[1], parts[2]

        if actualType != eventType {
            t.Errorf("Event type = %q, want %q", actualType, eventType)
        }
        if actualReason != reason {
            t.Errorf("Event reason = %q, want %q", actualReason, reason)
        }
        if matched, _ := regexp.MatchString(messagePattern, actualMessage); !matched {
            t.Errorf("Event message = %q, want pattern %q", actualMessage, messagePattern)
        }
    case <-time.After(5 * time.Second):
        t.Errorf("No event received, expected %s %s", eventType, reason)
    }
}
```

### Status Field Validation

```go
// Good: Structured status assertion
func AssertTWSStatus(t *testing.T, tws *v1alpha1.TimeWindowScaler, expected TWSStatusExpectation) {
    t.Helper()

    if expected.EffectiveReplicas != nil {
        if tws.Status.EffectiveReplicas != *expected.EffectiveReplicas {
            t.Errorf("EffectiveReplicas = %d, want %d",
                tws.Status.EffectiveReplicas, *expected.EffectiveReplicas)
        }
    }

    if expected.CurrentWindow != "" {
        if tws.Status.CurrentWindow != expected.CurrentWindow {
            t.Errorf("CurrentWindow = %q, want %q",
                tws.Status.CurrentWindow, expected.CurrentWindow)
        }
    }

    if expected.ObservedGeneration != nil {
        if tws.Status.ObservedGeneration != *expected.ObservedGeneration {
            t.Errorf("ObservedGeneration = %d, want %d",
                tws.Status.ObservedGeneration, *expected.ObservedGeneration)
        }
    }
}

type TWSStatusExpectation struct {
    EffectiveReplicas  *int32
    CurrentWindow      string
    ObservedGeneration *int64
    GracePeriodExpiry  *time.Time
}
```

### Reconciliation Result

```go
// Good: Reconcile result assertion
func AssertReconcileSuccess(t *testing.T, result ctrl.Result, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("Reconcile failed: %v", err)
    }
    if result.Requeue {
        t.Errorf("Unexpected requeue requested")
    }
}

func AssertReconcileRequeue(t *testing.T, result ctrl.Result, err error,
    expectedDuration time.Duration, tolerance time.Duration) {
    t.Helper()

    if err != nil {
        t.Fatalf("Reconcile failed: %v", err)
    }

    if !result.Requeue && result.RequeueAfter == 0 {
        t.Errorf("Expected requeue, but got none")
        return
    }

    diff := (result.RequeueAfter - expectedDuration).Abs()
    if diff > tolerance {
        t.Errorf("RequeueAfter = %v, want %v ± %v (diff: %v)",
            result.RequeueAfter, expectedDuration, tolerance, diff)
    }
}
```

## E2E Test Assertions

### Kubectl Output

```go
// Good: Command output assertion
func AssertKubectlOutput(t *testing.T, cmd, expected string) {
    t.Helper()
    output, err := exec.Command("kubectl", strings.Split(cmd, " ")...).Output()
    if err != nil {
        t.Fatalf("kubectl command failed: %v\nCommand: kubectl %s", err, cmd)
    }

    actual := strings.TrimSpace(string(output))
    if actual != expected {
        t.Errorf("kubectl %s\nGot:  %q\nWant: %q", cmd, actual, expected)
    }
}
```

### Resource State

```go
// Good: JSONPath assertion
func AssertResourceField(t *testing.T, resource, name, namespace, jsonPath, expected string) {
    t.Helper()

    cmd := fmt.Sprintf("get %s %s -n %s -o jsonpath='{%s}'",
        resource, name, namespace, jsonPath)
    output, err := kubectl(cmd)
    if err != nil {
        t.Fatalf("Failed to get %s/%s: %v", resource, name, err)
    }

    if output != expected {
        t.Errorf("%s %s/%s field %s = %q, want %q",
            resource, namespace, name, jsonPath, output, expected)
    }
}

// Usage
AssertResourceField(t, "deployment", "nginx", "default",
    ".spec.replicas", "10")
```

### Eventually Consistent State

```go
// Good: Retry with timeout
func AssertEventuallyReplicas(t *testing.T, namespace, deployment string,
    expected int32, timeout time.Duration) {
    t.Helper()

    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for time.Now().Before(deadline) {
        actual := getDeploymentReplicas(namespace, deployment)
        if actual == expected {
            return // Success
        }

        select {
        case <-ticker.C:
            t.Logf("Waiting for replicas: current=%d, expected=%d", actual, expected)
        case <-time.After(timeout):
            t.Fatalf("Timeout waiting for replicas to be %d (current: %d)", expected, actual)
        }
    }

    t.Fatalf("Deployment %s/%s replicas never reached %d within %v",
        namespace, deployment, expected, timeout)
}
```

### Log Verification

```go
// Good: Log pattern matching
func AssertControllerLogs(t *testing.T, pattern string, shouldExist bool) {
    t.Helper()

    logs := getControllerLogs()
    matched, _ := regexp.MatchString(pattern, logs)

    if shouldExist && !matched {
        t.Errorf("Controller logs should contain pattern %q but didn't.\nLogs:\n%s",
            pattern, logs)
    } else if !shouldExist && matched {
        t.Errorf("Controller logs should NOT contain pattern %q but did.\nLogs:\n%s",
            pattern, logs)
    }
}
```

## Common Patterns

### Table-Driven Tests

```go
// Good: Structured test cases
func TestWindowMatching(t *testing.T) {
    tests := []struct {
        name     string
        window   Window
        checkTime time.Time
        want     bool
    }{
        {
            name: "in_business_hours",
            window: Window{Days: []string{"Mon"}, Start: "09:00", End: "17:00"},
            checkTime: time.Date(2025, 3, 10, 14, 30, 0, 0, time.UTC),
            want: true,
        },
        {
            name: "after_hours",
            window: Window{Days: []string{"Mon"}, Start: "09:00", End: "17:00"},
            checkTime: time.Date(2025, 3, 10, 18, 0, 0, 0, time.UTC),
            want: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := IsInWindow(tt.window, tt.checkTime)
            if got != tt.want {
                t.Errorf("IsInWindow() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Helper Functions

```go
// Good: Reusable test helpers
func NewTestWindow(start, end string, days []string, replicas int32) Window {
    return Window{
        Start:    start,
        End:      end,
        Days:     days,
        Replicas: replicas,
    }
}

func MustParseTime(t *testing.T, layout, value string) time.Time {
    t.Helper()
    parsed, err := time.Parse(layout, value)
    if err != nil {
        t.Fatalf("Failed to parse time %q: %v", value, err)
    }
    return parsed
}

func IntPtr(i int32) *int32 { return &i }
func TimePtr(t time.Time) *time.Time { return &t }
```

### Cleanup Verification

```go
// Good: Ensure cleanup happens
func TestWithCleanup(t *testing.T) {
    namespace := createTestNamespace(t)
    defer func() {
        if err := deleteNamespace(namespace); err != nil {
            t.Errorf("Failed to cleanup namespace: %v", err)
        }
    }()

    // Test logic here
}
```

## Anti-Patterns to Avoid

### Bad: Vague Assertions
```go
// Bad: No context on failure
if got != want {
    t.Fail()
}

// Bad: Generic message
if err != nil {
    t.Error("test failed")
}
```

### Bad: Magic Values
```go
// Bad: What does 10 mean?
if replicas != 10 {
    t.Error("wrong replicas")
}

// Good: Named constant
const expectedBusinessHoursReplicas = 10
if replicas != expectedBusinessHoursReplicas {
    t.Errorf("replicas = %d, want %d (business hours)",
        replicas, expectedBusinessHoursReplicas)
}
```

### Bad: Sleep-Based Waiting
```go
// Bad: Fixed sleep
time.Sleep(5 * time.Second)
if getReplicas() != 10 {
    t.Error("not scaled")
}

// Good: Poll with timeout
require.Eventually(t, func() bool {
    return getReplicas() == 10
}, 30*time.Second, 1*time.Second, "Deployment should scale to 10")
```

### Bad: Comparing Complex Objects
```go
// Bad: Unclear what's different
if !reflect.DeepEqual(got, want) {
    t.Errorf("got %v, want %v", got, want)
}

// Good: Use diff library
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("Deployment mismatch (-want +got):\n%s", diff)
}
```

## Best Practices

1. **Use t.Helper()**: Mark assertion functions as helpers
2. **Include Context**: Always include relevant values in error messages
3. **Be Specific**: Test one thing per assertion
4. **Use Eventually**: For async operations, poll don't sleep
5. **Clean Failures**: Make failures actionable with clear messages
6. **Type Safety**: Use type-specific assertions over interface{}
7. **Consistent Format**: Follow team patterns for similar assertions
8. **Fail Fast**: Use t.Fatal for setup failures, t.Error for test assertions