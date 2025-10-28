# Kyklos Test Strategy

## Executive Summary

The Kyklos test suite ensures the Time Window Scaler controller operates correctly across all time-based scheduling scenarios, edge cases, and failure modes. Our testing philosophy emphasizes **determinism**, **speed**, and **comprehensive coverage** while preventing flakiness.

## Testing Goals

### Primary Goals
1. **Correctness**: Validate all time window calculations, state transitions, and scaling operations
2. **Determinism**: Every test produces identical results across runs using controlled time sources
3. **Speed**: Complete test suite execution in under 10 minutes for rapid feedback
4. **Coverage**: Achieve 95% code coverage for critical paths, 80% overall
5. **Reliability**: Maintain less than 0.1% test flake rate

### Non-Goals
- Performance benchmarking (separate benchmark suite)
- Load testing (handled by dedicated stress tests)
- UI/CLI testing (out of scope for controller tests)
- Multi-cluster scenarios (v1alpha1 is single-cluster)

## Testing Scope

### In Scope
- Time window boundary calculations
- Holiday mode behaviors
- Grace period mechanics
- Pause/resume operations
- Timezone and DST handling
- Controller reconciliation loops
- Status condition management
- Event emission
- Manual drift correction
- Resource lifecycle management
- Error recovery paths
- Webhook validation

### Out of Scope
- Kubernetes API server implementation
- Go standard library time functions
- Third-party operator SDK functionality
- Network layer reliability
- Hardware failures

## Test Levels

### Level 1: Unit Tests
**Purpose**: Validate pure functions and business logic without Kubernetes dependencies

**Scope**:
- Time mathematics (window matching, boundary computation)
- Holiday evaluation logic
- Grace period calculations
- Precedence rules for overlapping windows
- Timezone conversions
- Cron expression parsing

**Characteristics**:
- No Kubernetes API calls
- Use `time.FakeClock` for deterministic time
- Sub-millisecond execution per test
- 100% deterministic
- Run on every commit

### Level 2: Envtest Tests
**Purpose**: Validate controller behavior with a real Kubernetes API server

**Scope**:
- Full reconciliation loops
- Resource CRUD operations
- Status updates and conditions
- Event emission
- Webhook validation
- Error handling and recovery
- Multi-resource interactions

**Characteristics**:
- Uses controller-runtime envtest
- Embedded etcd and API server
- Sub-second execution per test
- Controlled time advancement
- Run on every pull request

### Level 3: E2E Tests
**Purpose**: Validate end-to-end scenarios in a real Kubernetes cluster

**Scope**:
- Complete user workflows
- Multi-minute time progressions
- Resource cleanup verification
- Upgrade scenarios
- Integration with actual workloads
- Observability validation

**Characteristics**:
- Uses kind/k3d clusters
- Minute-scale time windows for speed
- Under 10 seconds per scenario
- Run before releases
- Acceptance criteria validation

## Determinism Guidelines

### Time Control
```go
// All tests MUST use controlled time
type TestClock interface {
    Now() time.Time
    Advance(d time.Duration)
    Set(t time.Time)
}

// Example usage
clock := NewFakeClock(time.Date(2025, 3, 15, 14, 30, 0, 0, time.UTC))
```

### Fixed Inputs
- Use static UUIDs: `test-uuid-0001`, `test-uuid-0002`
- Fixed random seeds: `rand.Seed(42)`
- Deterministic generation counts: Start at 1, increment by 1
- Stable resource names: `test-deployment-1`, `test-tws-1`

### Ordering Guarantees
- Sort arrays before comparison
- Use ordered maps where iteration matters
- Explicit event ordering in assertions
- Deterministic reconcile queue processing

## Coverage Goals

### Target Coverage Levels
| Component | Target | Minimum | Rationale |
|-----------|--------|---------|-----------|
| Time calculations | 100% | 95% | Core business logic |
| Reconcile loop | 95% | 90% | Critical path |
| Status updates | 90% | 85% | User-visible state |
| Event emission | 85% | 80% | Observability |
| Error handling | 90% | 85% | Resilience |
| Validation | 95% | 90% | API contract |
| Overall | 85% | 80% | Project baseline |

### Coverage Exclusions
- Generated code (deepcopy, clientsets)
- Panic recovery (catastrophic failures)
- Main function bootstrapping
- Unreachable defensive code

## Flake Budget

### Definition
A test is considered flaky if it fails and then passes on retry without code changes.

### Budget Allocation
- **Unit tests**: 0% flake rate (must be 100% deterministic)
- **Envtest**: Maximum 0.05% flake rate
- **E2E tests**: Maximum 0.5% flake rate
- **Overall suite**: Maximum 0.1% flake rate

### Flake Prevention
1. Always use controlled time sources
2. Avoid hardcoded sleep statements
3. Use explicit waits with conditions
4. Clean up resources after each test
5. Isolate tests with unique namespaces
6. Retry only at infrastructure level, not test level

## Test Execution Strategy

### Local Development
```bash
# Fast feedback loop
make test-unit        # <5 seconds
make test-envtest     # <30 seconds
make test-e2e-quick   # <2 minutes (subset)
```

### CI Pipeline
```bash
# Pull request validation
make test-unit        # Runs always
make test-envtest     # Runs always
make test-e2e         # Runs on /test comment

# Release validation
make test-all         # Full suite
make test-upgrade     # Upgrade scenarios
make test-chaos       # Failure injection
```

### Test Parallelization
- Unit tests: Parallel by default (`go test -parallel=8`)
- Envtest: Parallel with separate namespaces
- E2E: Sequential to avoid resource conflicts

## Test Data Management

### Time Test Cases
Use significant dates that expose edge cases:
- `2025-03-09 02:30:00` - Spring DST forward (becomes 03:30)
- `2025-11-02 02:30:00` - Fall DST back (happens twice)
- `2025-02-28 23:59:59` - Non-leap year boundary
- `2024-02-29 12:00:00` - Leap year date
- `2025-12-31 23:59:59` - Year boundary
- `2025-01-01 00:00:00` - New year start

### Timezone Test Set
- `UTC` - Baseline, no DST
- `America/New_York` - Eastern time with DST
- `Asia/Kolkata` - IST, no DST, +05:30 offset
- `Pacific/Auckland` - Southern hemisphere DST
- `Europe/London` - GMT/BST transitions

### Window Configurations
Standard test windows for consistency:
- `09:00-17:00` - Business hours
- `22:00-06:00` - Cross-midnight
- `00:00-23:59` - Full day
- `12:00-12:30` - Short window
- `23:45-00:15` - Midnight spanning

## Quality Gates

### Pre-commit
- Unit tests must pass
- Coverage must not decrease
- No new linting issues

### Pull Request
- All unit tests pass
- All envtest scenarios pass
- Coverage ≥ 80%
- No flaky tests in last 10 runs

### Release
- Full test suite passes
- E2E acceptance tests pass
- Upgrade tests pass
- Performance benchmarks within bounds
- Security scan clean

## Test Maintenance

### Test Review Checklist
- [ ] Test uses controlled time source
- [ ] Resources are cleaned up
- [ ] Assertions have clear failure messages
- [ ] No hardcoded delays
- [ ] Test is independent of others
- [ ] Edge cases are covered
- [ ] Test name describes scenario

### Quarantine Process
1. Flaky test detected (fails 2+ times in 100 runs)
2. Mark with `[Flaky]` tag
3. Create issue with reproduction steps
4. Fix within 2 sprints or remove
5. Validate fix with 1000 runs

## Anti-Patterns to Avoid

### Time-Related
- ❌ Using `time.Now()` directly
- ❌ Hardcoded sleep durations
- ❌ Assuming specific test execution time
- ❌ Timezone-dependent assertions

### Resource-Related
- ❌ Sharing resources between tests
- ❌ Assuming resource creation order
- ❌ Not cleaning up after failures
- ❌ Hardcoded cluster endpoints

### Assertion-Related
- ❌ Vague error messages ("test failed")
- ❌ Multiple unrelated assertions per test
- ❌ Asserting on timestamps without tolerance
- ❌ Order-dependent array comparisons

## Success Metrics

### KPIs
- Test execution time: <10 minutes for full suite
- Flake rate: <0.1% over 30 days
- Coverage: >80% overall, >95% critical paths
- Test failures caught bugs: >90% of bugs found by tests
- Time to diagnose failure: <5 minutes average

### Monitoring
- Track flake rates in CI dashboard
- Alert on coverage drops >2%
- Monitor test execution time trends
- Report on test effectiveness quarterly

## Migration Path

### Phase 1: Foundation (Week 1-2)
- Set up test infrastructure
- Create time control utilities
- Establish envtest harness
- Write first 10 unit tests

### Phase 2: Core Coverage (Week 3-4)
- Cover all time calculations
- Test reconcile loop paths
- Add status update tests
- Implement E2E framework

### Phase 3: Edge Cases (Week 5-6)
- DST transition tests
- Holiday mode scenarios
- Grace period edge cases
- Error injection tests

### Phase 4: Hardening (Week 7-8)
- Chaos testing
- Performance benchmarks
- Upgrade scenarios
- Documentation

## Related Documents

- [UNIT-PLAN.md](UNIT-PLAN.md) - Detailed unit test specifications
- [ENVTEST-PLAN.md](ENVTEST-PLAN.md) - Controller test scenarios
- [E2E-PLAN.md](E2E-PLAN.md) - End-to-end test plans
- [ASSERTIONS.md](ASSERTIONS.md) - Standard assertion patterns
- [FLAKE-POLICY.md](FLAKE-POLICY.md) - Flake prevention details