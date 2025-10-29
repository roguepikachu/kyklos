# Contributing to Kyklos Implementation

**Purpose:** Onboarding guide for developers contributing to Kyklos v0.1 implementation.

**Last Updated:** 2025-10-29

## Welcome!

This guide helps you set up your development environment, understand the codebase structure, run tests, and submit contributions to Kyklos.

---

## Prerequisites

### Required Tools

| Tool | Version | Purpose | Installation |
|------|---------|---------|-------------|
| Go | 1.21+ | Build controller | https://golang.org/dl/ |
| Docker | 24.0+ | Build container images | https://docs.docker.com/get-docker/ |
| kubectl | 1.25+ | Interact with Kubernetes | https://kubernetes.io/docs/tasks/tools/ |
| kind | 0.20+ | Local Kubernetes cluster | https://kind.sigs.k8s.io/docs/user/quick-start/ |
| make | 3.81+ | Build automation | Included in most systems |

### Optional Tools

| Tool | Purpose | Installation |
|------|---------|-------------|
| kubebuilder | Controller scaffolding | https://book.kubebuilder.io/quick-start.html |
| golangci-lint | Code linting | https://golangci-lint.run/usage/install/ |
| jq | JSON processing | https://stedolan.github.io/jq/download/ |

### Knowledge Prerequisites

**Required:**
- Go programming (interfaces, error handling, testing)
- Kubernetes basics (Pods, Deployments, Services)
- Git workflow (branch, commit, PR)

**Helpful:**
- controller-runtime library
- Kustomize
- Time zone concepts (IANA database, DST)

---

## Quick Start (5 Minutes)

```bash
# 1. Clone repository
git clone https://github.com/aykumar/kyklos.git
cd kyklos

# 2. Verify prerequisites
make prereq-check

# 3. Download dependencies
go mod download

# 4. Generate code and manifests
make generate
make manifests

# 5. Run unit tests
make test-unit

# 6. Build controller binary
make build

# Success! You're ready to develop.
```

---

## Repository Structure

```
kyklos/
├── api/v1alpha1/           # CRD types (DO NOT edit generated files)
├── controllers/            # Reconciler implementation
├── internal/
│   ├── timecalc/          # Time calculation engine (CRITICAL PATH - 100% coverage)
│   ├── statuswriter/      # Status update logic
│   ├── events/            # Event recorder
│   └── metrics/           # Prometheus metrics
├── config/                # Kustomize manifests
├── test/
│   ├── e2e/              # End-to-end tests
│   └── fixtures/         # Test data (fixed dates)
├── docs/
│   ├── implementation/   # Implementation planning (this directory)
│   └── design/           # Design documents
├── Makefile              # Build targets
└── go.mod                # Go module definition
```

**Golden Rule:** Follow REPO-LAYOUT.md for file placement.

---

## Development Workflow

### Step 1: Pick a Task

```bash
# View available tasks
cat docs/implementation/TASKS.csv

# Or check GitHub issues
gh issue list --label "good-first-issue"
```

**Task Selection Tips:**
- Start with unit tests (T007-T010, T032-T033, T047-T048)
- Then move to controller logic (T011-T015)
- Avoid changing API types until familiar with controller-gen

### Step 2: Create Feature Branch

```bash
# Create branch from main
git checkout main
git pull origin main
git checkout -b feature/T007-window-matching

# Branch naming convention:
# - feature/TXXX-description (new functionality)
# - fix/issue-NNN-description (bug fix)
# - test/TXXX-description (test-only changes)
# - docs/topic-description (documentation)
```

### Step 3: Make Changes

**Before Coding:**

1. Read relevant design docs:
   - `/docs/api/CRD-SPEC.md` for API fields
   - `/docs/design/RECONCILE.md` for reconcile logic
   - `/docs/implementation/INTERFACE-CONTRACTS.md` for function signatures
   - `/docs/implementation/PSEUDOCODE.md` for algorithms

2. Understand acceptance criteria from TASKS.csv

3. Check for existing tests in similar modules

**While Coding:**

```bash
# Run tests frequently (fast feedback)
make test-unit

# Run linter
make lint

# Generate code if you changed API types
make generate

# Generate manifests if you changed kubebuilder markers
make manifests
```

### Step 4: Write Tests

**Test Coverage Requirements:**

| Module | Minimum Coverage | Target Coverage |
|--------|-----------------|-----------------|
| timecalc | 95% | 100% |
| controllers | 85% | 90% |
| statuswriter | 80% | 85% |
| events | 75% | 80% |
| metrics | 70% | 75% |

**Test Patterns:**

**Unit Test Example (timecalc):**

```go
// internal/timecalc/matcher_test.go
func TestComputeEffectiveReplicas_InWindow(t *testing.T) {
    // Setup
    windows := []TimeWindow{
        {
            Days:     []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
            Start:    "09:00",
            End:      "17:00",
            Replicas: 10,
        },
    }
    defaultReplicas := int32(2)
    localTime := time.Date(2025, 1, 27, 14, 30, 0, 0, time.UTC) // Monday 14:30

    // Execute
    result := ComputeEffectiveReplicas(windows, defaultReplicas, localTime)

    // Assert
    if result != 10 {
        t.Errorf("Expected 10 replicas, got %d", result)
    }
}
```

**Table-Driven Test Example:**

```go
func TestComputeEffectiveReplicas_TableDriven(t *testing.T) {
    tests := []struct {
        name            string
        windows         []TimeWindow
        defaultReplicas int32
        localTime       time.Time
        want            int32
    }{
        {
            name:            "in window",
            windows:         businessHoursWindow,
            defaultReplicas: 2,
            localTime:       monday1430,
            want:            10,
        },
        {
            name:            "out of window",
            localTime:       monday2000,
            want:            2,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ComputeEffectiveReplicas(tt.windows, tt.defaultReplicas, tt.localTime)
            if got != tt.want {
                t.Errorf("got %d, want %d", got, tt.want)
            }
        })
    }
}
```

**Integration Test Example (controllers):**

```go
// controllers/timewindowscaler_controller_test.go
var _ = Describe("TimeWindowScalerController", func() {
    Context("When reconciling a TimeWindowScaler", func() {
        It("Should scale Deployment to window replicas", func() {
            // Setup
            ctx := context.Background()
            tws := createTimeWindowScaler("test-tws", "default")
            deployment := createDeployment("test-deployment", "default", 2)

            // Create resources
            Expect(k8sClient.Create(ctx, tws)).To(Succeed())
            Expect(k8sClient.Create(ctx, deployment)).To(Succeed())

            // Wait for reconciliation
            Eventually(func() int32 {
                Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)).To(Succeed())
                return deployment.Spec.Replicas
            }, timeout, interval).Should(Equal(int32(10)))

            // Verify status
            Eventually(func() int32 {
                Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(tws), tws)).To(Succeed())
                return tws.Status.EffectiveReplicas
            }, timeout, interval).Should(Equal(int32(10)))
        })
    })
})
```

### Step 5: Run Full Test Suite

```bash
# Run all tests
make test-all

# Check coverage
make test-coverage
open coverage.html

# Run linter
make lint

# Verify manifests are up to date
make manifests
git status  # Should show no changes
```

### Step 6: Test Locally

```bash
# Create kind cluster
make kind-cluster

# Install CRD
make install

# Run controller locally (against kind cluster)
make run

# In another terminal, create test resources
kubectl apply -f config/samples/basic.yaml

# Watch logs
# Controller logs appear in first terminal

# Verify scaling
kubectl get deployment -n default
kubectl get tws -n default

# Cleanup
make kind-delete
```

### Step 7: Commit Changes

```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat(timecalc): implement window matching algorithm

- Implement ComputeEffectiveReplicas function
- Add support for cross-midnight windows
- Add unit tests with 100% coverage
- Closes #42 (T007)"

# Push to origin
git push origin feature/T007-window-matching
```

**Commit Message Format:**

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `test`: Test-only changes
- `docs`: Documentation
- `refactor`: Code refactoring
- `chore`: Build/tooling changes

**Scopes:**
- `api`: API types
- `timecalc`: Time calculation engine
- `controller`: Reconciler
- `statuswriter`: Status updates
- `events`: Event emission
- `metrics`: Metrics
- `config`: Kustomize manifests

### Step 8: Create Pull Request

```bash
# Create PR using GitHub CLI
gh pr create --title "feat(timecalc): implement window matching algorithm" \
             --body "Implements T007: Window matching with cross-midnight support. See commit message for details."

# Or create PR via GitHub web UI
```

**PR Checklist:**

- [ ] Title follows commit message format
- [ ] Description links to issue/task (e.g., "Implements T007")
- [ ] All tests pass locally
- [ ] Coverage meets minimum requirements
- [ ] No linter errors
- [ ] Manifests regenerated if API changed
- [ ] Documentation updated if user-facing change

---

## Testing Guide

### Running Tests

```bash
# Unit tests (fast, no Kubernetes)
make test-unit

# Integration tests (envtest, embedded API server)
make test-envtest

# E2E tests (requires kind cluster)
make test-e2e

# All tests
make test-all

# Specific package
go test ./internal/timecalc/... -v

# Specific test
go test ./internal/timecalc/... -run TestComputeEffectiveReplicas_InWindow -v

# With coverage
go test ./internal/timecalc/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Conventions

**DO:**
- Use table-driven tests for multiple scenarios
- Use fixed dates (2025-01-27, not time.Now())
- Mock time via function parameters, not global state
- Clean up resources in AfterEach/cleanup functions
- Use meaningful test names (TestFunction_Scenario_ExpectedResult)

**DON'T:**
- Use time.Now() in tests (non-deterministic)
- Use time.Sleep() for synchronization (flaky)
- Share state between tests
- Hardcode cluster-specific values (endpoints, IPs)

### Time Testing Pattern

```go
// ✓ GOOD: Explicit time parameter
func TestWithFixedTime(t *testing.T) {
    fixedTime := time.Date(2025, 1, 27, 14, 30, 0, 0, time.UTC)
    result := ComputeEffectiveReplicas(windows, defaultReplicas, fixedTime)
    // ...
}

// ✗ BAD: Uses current time (non-deterministic)
func TestWithCurrentTime(t *testing.T) {
    result := ComputeEffectiveReplicas(windows, defaultReplicas, time.Now())
    // ...
}
```

---

## Code Standards

### Go Style

Follow [Effective Go](https://golang.org/doc/effective_go.html) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

**Key Conventions:**

1. **Error Handling:**
   ```go
   // ✓ GOOD: Explicit error check
   if err != nil {
       return nil, fmt.Errorf("failed to get deployment: %w", err)
   }

   // ✗ BAD: Ignored error
   deployment, _ := getDeployment()
   ```

2. **Interfaces:**
   ```go
   // ✓ GOOD: Small, focused interface
   type TimeCalculator interface {
       ComputeEffectiveReplicas(windows []TimeWindow, defaultReplicas int32, localTime time.Time) int32
   }

   // ✗ BAD: Large interface with many methods
   ```

3. **Contexts:**
   ```go
   // ✓ GOOD: Pass context as first parameter
   func Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)

   // ✗ BAD: Store context in struct field
   ```

4. **Naming:**
   - Packages: lowercase, single word (timecalc, not timeCalc)
   - Exported: PascalCase (ComputeEffectiveReplicas)
   - Unexported: camelCase (computeNextBoundary)
   - Constants: PascalCase or SCREAMING_SNAKE_CASE

### Comments

```go
// ✓ GOOD: Explains why, not what
// ComputeEffectiveReplicas determines the desired replica count by evaluating
// all time windows against the current local time. Cross-midnight windows
// are handled by checking both the current day and previous day.
func ComputeEffectiveReplicas(...) int32 {
    // ...
}

// ✗ BAD: Restates code
// ComputeEffectiveReplicas computes effective replicas
func ComputeEffectiveReplicas(...) int32 {
    // ...
}
```

### Logging

```go
// Use structured logging with logr
log := ctrl.LoggerFrom(ctx)

// ✓ GOOD: Structured with key-value pairs
log.Info("Scaling deployment", "deployment", targetName, "from", currentReplicas, "to", desiredReplicas)

// ✗ BAD: Concatenated strings
log.Info(fmt.Sprintf("Scaling %s from %d to %d", targetName, currentReplicas, desiredReplicas))
```

---

## Common Tasks

### Adding a New Field to CRD

1. **Edit API types:**
   ```bash
   vim api/v1alpha1/timewindowscaler_types.go
   ```

2. **Add field with tags:**
   ```go
   // +kubebuilder:validation:Minimum=0
   // +kubebuilder:validation:Maximum=3600
   MaxScaleUpDelay int32 `json:"maxScaleUpDelay,omitempty"`
   ```

3. **Regenerate:**
   ```bash
   make manifests
   make generate
   ```

4. **Update tests:**
   ```bash
   vim controllers/timewindowscaler_controller_test.go
   ```

5. **Update docs:**
   ```bash
   vim docs/api/CRD-SPEC.md
   ```

### Adding a New Metric

1. **Define metric:**
   ```go
   // internal/metrics/metrics.go
   var (
       scaleOperationDuration = prometheus.NewHistogramVec(
           prometheus.HistogramOpts{
               Name: "kyklos_scale_operation_duration_seconds",
               Help: "Duration of scale operations",
           },
           []string{"tws_name", "namespace", "direction"},
       )
   )
   ```

2. **Register in init():**
   ```go
   func init() {
       metrics.Registry.MustRegister(scaleOperationDuration)
   }
   ```

3. **Add recorder function:**
   ```go
   func RecordScaleOperationDuration(twsName, namespace, direction string, duration float64) {
       scaleOperationDuration.WithLabelValues(twsName, namespace, direction).Observe(duration)
   }
   ```

4. **Document:**
   ```bash
   vim docs/user/OPERATIONS.md
   ```

### Debugging Locally

```bash
# Run controller with debug logging
make run LOG_LEVEL=debug

# Or use delve debugger
dlv debug ./cmd/controller/main.go -- --metrics-bind-address=:8080

# In another terminal, create test resource
kubectl apply -f config/samples/basic.yaml

# Watch reconciliation in debugger
```

---

## Troubleshooting

### Issue: Tests Fail with "connection refused"

**Cause:** Envtest API server not started

**Solution:**
```bash
# Install envtest binaries
make envtest-setup

# Re-run tests
make test-envtest
```

### Issue: CRD generation fails

**Cause:** kubebuilder markers incorrect

**Solution:**
```bash
# Check marker syntax
go doc -all ./api/v1alpha1

# Verify controller-gen version
controller-gen --version

# Reinstall if needed
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
```

### Issue: Linter errors

**Cause:** Code style violations

**Solution:**
```bash
# Auto-fix many issues
make lint-fix

# Review remaining issues
make lint

# Ignore false positives (use sparingly)
// nolint:errcheck // Reason for ignoring
```

### Issue: Test flakiness

**Cause:** Non-deterministic time or resource creation

**Solution:**
- Use fixed dates, never time.Now()
- Use Eventually() with proper timeout for async operations
- Clean up resources in test teardown
- Isolate tests with unique namespaces

---

## Getting Help

### Documentation

- Implementation Plan: `/docs/implementation/IMPLEMENTATION-PLAN.md`
- API Spec: `/docs/api/CRD-SPEC.md`
- Reconcile Design: `/docs/design/RECONCILE.md`
- Interface Contracts: `/docs/implementation/INTERFACE-CONTRACTS.md`
- Pseudocode: `/docs/implementation/PSEUDOCODE.md`

### Community

- GitHub Issues: https://github.com/aykumar/kyklos/issues
- Discussions: https://github.com/aykumar/kyklos/discussions
- Slack: #kyklos on Kubernetes Slack (future)

### Questions?

- For design questions: Check `/docs/design/`
- For API questions: Check `/docs/api/CRD-SPEC.md`
- For implementation questions: Check `/docs/implementation/`
- Still stuck? Open a discussion on GitHub

---

## Contribution Checklist

Before submitting a PR:

- [ ] Code compiles: `make build`
- [ ] Tests pass: `make test-all`
- [ ] Coverage meets requirements: `make test-coverage`
- [ ] Linter clean: `make lint`
- [ ] Manifests updated: `make manifests` (no git changes)
- [ ] Documentation updated (if user-facing change)
- [ ] Commit messages follow format
- [ ] PR description links to issue/task
- [ ] Branch rebased on latest main

---

## Maintainer Notes

### Review Priorities

When reviewing PRs:

1. **Correctness:** Does it work as intended?
2. **Tests:** Are there tests with good coverage?
3. **Documentation:** Is it documented?
4. **Style:** Does it follow conventions?
5. **Performance:** Are there obvious bottlenecks?

### Merge Criteria

PRs must meet:

- All CI checks passing
- At least one approval from maintainer
- No unresolved comments
- Code coverage not decreased
- Documentation updated

---

## Thank You!

Your contributions make Kyklos better for everyone. Welcome to the team!
