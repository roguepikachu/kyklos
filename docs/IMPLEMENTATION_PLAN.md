# Kyklos v0.1 Implementation Plan

**Created:** 2025-11-23 | **Owner:** Orchestrator | **Target:** 2025-11-27 23:59 IST

## Overview

This plan details the remaining work to complete Kyklos v0.1, taking the project from 83.8% complete to a fully functional, tested release.

## Current State Analysis

### What's Working
- Core engine: 83.8% test coverage, all time window logic functional
- Controller skeleton: Reconcile loop, status updates, event emission
- CRD: Fully defined with OpenAPI validation
- Documentation: Comprehensive (83+ files covering API, design, testing, operations)
- Build system: Make targets, Docker build, Kind integration

### What's Missing
1. Grace period timing implementation in reconciler
2. Holiday ConfigMap reading logic in controller
3. Metrics implementation (Prometheus endpoints)
4. E2E test scenarios for core functionality
5. Integration verification of all components

## Implementation Tasks

### Task 1: Grace Period Timing Logic
**File:** /Users/aykumar/personal/kyklos/internal/controller/timewindowscaler_controller.go

**Current State:**
- Grace period field exists in Input struct (line 44: GracePeriodSecs)
- LastScaleTime tracking exists (lines 206-208)
- Engine receives grace period data but doesn't use it for timing decisions

**Required Changes:**
1. Modify engine.ComputeEffectiveReplicas() to check:
   - If scaling down (CurrentReplicas > computed replicas)
   - If LastScaleTime + GracePeriodSecs > Now
   - If yes, return CurrentReplicas (maintain current) with reason "grace-period-active"
2. Update Output struct to include GracePeriodRemaining field
3. When grace period active, set NextBoundary to LastScaleTime + GracePeriodSecs
4. Update status to show grace period state in conditions

**Acceptance Criteria:**
- Scale-down delayed by gracePeriodSeconds
- Scale-up happens immediately (no grace period)
- Status shows "GracePeriod" reason when active
- NextBoundary correctly set to grace period end time
- Tests validate grace period behavior

**Estimated Time:** 3 hours

---

### Task 2: Holiday ConfigMap Reading
**File:** /Users/aykumar/personal/kyklos/internal/controller/timewindowscaler_controller.go

**Current State:**
- TODO comment at line 198: `IsHoliday: false, // TODO: check holiday ConfigMap`
- HolidayMode passed to engine but IsHoliday always false
- Engine logic for holiday handling exists (lines 85-111)

**Required Changes:**
1. Add RBAC for ConfigMap reading:
   - Update controller markers for ConfigMap get/list permissions
   - Run `make manifests` to regenerate RBAC
2. Implement readHolidayConfigMap() helper function:
   - Check if tws.Spec.Holidays is configured
   - If yes, fetch ConfigMap by name from appropriate namespace
   - Parse data keys as YYYY-MM-DD dates
   - Return true if today's date (in TWS timezone) matches any key
3. Call readHolidayConfigMap() in buildEngineInput()
4. Handle ConfigMap not found error:
   - Set Degraded condition
   - Log warning
   - Continue with IsHoliday=false

**Acceptance Criteria:**
- Controller reads holiday ConfigMap when configured
- Today's date checked in TWS timezone
- Holiday modes (ignore, closed, open) work correctly
- Missing ConfigMap handled gracefully with status condition
- Events emitted for holiday state changes
- Tests validate all three holiday modes

**Estimated Time:** 4 hours

---

### Task 3: Prometheus Metrics Implementation
**Files:**
- /Users/aykumar/personal/kyklos/internal/controller/timewindowscaler_controller.go
- /Users/aykumar/personal/kyklos/internal/controller/metrics.go (new)

**Current State:**
- Metrics endpoint exists (from kubebuilder scaffold) on port 8443
- No custom metrics defined
- E2E test expects controller_runtime_reconcile_total metric

**Required Changes:**
1. Create metrics.go with Prometheus collectors:
   ```go
   var (
     scaleOperationsTotal = prometheus.NewCounterVec(
       prometheus.CounterOpts{
         Name: "kyklos_scale_operations_total",
         Help: "Total number of scaling operations by direction and result",
       },
       []string{"tws_name", "tws_namespace", "direction", "result"},
     )

     currentEffectiveReplicas = prometheus.NewGaugeVec(
       prometheus.GaugeOpts{
         Name: "kyklos_current_effective_replicas",
         Help: "Current effective replicas for each TimeWindowScaler",
       },
       []string{"tws_name", "tws_namespace"},
     )

     windowTransitionsTotal = prometheus.NewCounterVec(
       prometheus.CounterOpts{
         Name: "kyklos_window_transitions_total",
         Help: "Total window transitions by TWS",
       },
       []string{"tws_name", "tws_namespace", "from_window", "to_window"},
     )

     reconcileDuration = prometheus.NewHistogramVec(
       prometheus.HistogramOpts{
         Name: "kyklos_reconcile_duration_seconds",
         Help: "Time spent in reconcile loop",
       },
       []string{"tws_name", "tws_namespace"},
     )
   )
   ```

2. Register metrics in init() or SetupWithManager()
3. Instrument reconcile loop:
   - Increment scaleOperationsTotal on scale operations
   - Update currentEffectiveReplicas gauge
   - Track window transitions
   - Measure reconcile duration
4. Add metrics documentation to docs/user/OPERATIONS.md

**Acceptance Criteria:**
- Metrics endpoint exposes custom kyklos_* metrics
- Scale operations tracked with direction labels (up/down)
- Current effective replicas gauge accurate
- Reconcile duration histogram populated
- E2E test can scrape metrics successfully
- Metrics documented in operations guide

**Estimated Time:** 5 hours

---

### Task 4: E2E Test Scenarios
**File:** /Users/aykumar/personal/kyklos/test/e2e/e2e_test.go

**Current State:**
- E2E suite scaffolded (lines 45-271)
- Tests for controller deployment and metrics endpoint exist
- TODO comment at line 262 for custom scenarios
- No TimeWindowScaler-specific tests

**Required Changes:**
1. Add test: "should scale deployment based on time window"
   - Create test deployment with 1 replica
   - Create TWS with current time in window (10 replicas)
   - Wait for deployment to scale to 10
   - Update TWS to move window to future
   - Wait for deployment to scale to defaultReplicas

2. Add test: "should respect grace period on scale down"
   - Create deployment with 10 replicas
   - Create TWS with grace period 60s, window just ended
   - Verify replicas stay at 10 for 60+ seconds
   - Verify replicas scale down after grace period

3. Add test: "should handle holiday ConfigMap"
   - Create holiday ConfigMap with today's date
   - Create TWS with holidayMode: treat-as-closed
   - Verify deployment scales to 0
   - Change mode to treat-as-open
   - Verify deployment scales to max window replicas

4. Add test: "should handle cross-midnight windows"
   - Create TWS with window 22:00-02:00
   - Use minute-scale time for fast testing
   - Verify scaling behavior across midnight boundary

5. Add test: "should respect pause mode"
   - Create TWS with pause: true
   - Verify deployment not scaled
   - Verify status shows computed state
   - Set pause: false
   - Verify scaling resumes

**Acceptance Criteria:**
- All 5 E2E tests pass reliably
- Tests use minute-scale windows for speed (complete in <5 min)
- Tests clean up resources properly
- Tests emit useful failure diagnostics
- CI can run tests (add to GitHub Actions workflow)

**Estimated Time:** 8 hours

---

### Task 5: Integration Verification
**Manual Testing Checklist**

**Environment Setup:**
1. Create local Kind cluster: `make cluster-up`
2. Build and load controller: `make build docker-build kind-load`
3. Install CRDs: `make install-crds`
4. Deploy controller: `make deploy`

**Test Scenarios:**
1. Basic scaling:
   - Apply examples/tws-office-hours.yaml
   - Watch deployment scale
   - Verify status updates
   - Check events and logs

2. Holiday ConfigMap:
   - Create holiday ConfigMap with today's date
   - Apply TWS with holiday mode
   - Verify behavior matches mode
   - Check Degraded condition if ConfigMap missing

3. Grace period:
   - Create TWS with grace period 120s
   - Trigger scale-down
   - Verify delay
   - Check status shows grace period reason

4. Cross-midnight:
   - Apply examples/tws-night-shift.yaml
   - Verify window spans midnight correctly

5. Pause mode:
   - Set pause: true on running TWS
   - Manually scale deployment
   - Verify controller doesn't revert
   - Set pause: false
   - Verify drift correction

6. Metrics:
   - Port-forward metrics service
   - Curl metrics endpoint
   - Verify kyklos_* metrics present
   - Check metric values match scaling operations

**Documentation:**
- Record test results in docs/VERIFICATION.md
- Screenshot metrics dashboard (if time permits)
- Capture example logs showing scaling decisions

**Estimated Time:** 4 hours

---

### Task 6: Documentation Updates
**Files to Update:**
- /Users/aykumar/personal/kyklos/README.md (update status section)
- /Users/aykumar/personal/kyklos/docs/user/OPERATIONS.md (add metrics)
- /Users/aykumar/personal/kyklos/docs/ROADMAP.md (mark v0.1 complete)

**Changes:**
1. README.md:
   - Change "Current Implementation Status" checkmarks
   - Remove "structure in place" notes
   - Update version to v0.1.0

2. OPERATIONS.md:
   - Add Prometheus metrics section
   - Document all kyklos_* metrics with examples
   - Add PromQL query examples
   - Add alerting rules suggestions

3. ROADMAP.md:
   - Mark v0.1 features as complete
   - Add v0.2 planning section

**Acceptance Criteria:**
- All status markers accurate
- Metrics fully documented with examples
- No outdated information remains

**Estimated Time:** 2 hours

---

## Task Schedule

### Day 1: 2025-11-24 (Sunday)
**Goal:** Complete grace period and holiday ConfigMap

| Time (IST) | Task | Owner | Deliverable |
|------------|------|-------|-------------|
| 10:00-13:00 | Task 1: Grace Period Logic | Developer | Grace period working, tests pass |
| 14:00-18:00 | Task 2: Holiday ConfigMap | Developer | Holiday modes functional |
| 18:00-19:00 | Daily checkpoint | Orchestrator | Status update |

**Checkpoint:** Grace period and holidays working in controller

---

### Day 2: 2025-11-25 (Monday)
**Goal:** Complete metrics implementation

| Time (IST) | Task | Owner | Deliverable |
|------------|------|-------|-------------|
| 10:00-15:00 | Task 3: Prometheus Metrics | Developer | Metrics exposed and documented |
| 15:00-16:00 | Verify metrics locally | Developer | Metrics scraped successfully |
| 16:00-17:00 | Daily checkpoint | Orchestrator | Status update |

**Checkpoint:** Metrics endpoint functional with all custom metrics

---

### Day 3: 2025-11-26 (Tuesday)
**Goal:** Complete E2E tests and integration verification

| Time (IST) | Task | Owner | Deliverable |
|------------|------|-------|-------------|
| 10:00-18:00 | Task 4: E2E Test Scenarios | Developer | 5 E2E tests passing |
| 18:00-22:00 | Task 5: Integration Verification | Developer | Manual test checklist complete |
| 22:00-23:00 | Daily checkpoint | Orchestrator | Status update |

**Checkpoint:** All tests pass, manual verification complete

---

### Day 4: 2025-11-27 (Wednesday)
**Goal:** Documentation and release preparation

| Time (IST) | Task | Owner | Deliverable |
|------------|------|-------|-------------|
| 10:00-12:00 | Task 6: Documentation Updates | Developer | Docs accurate and complete |
| 12:00-14:00 | Final testing pass | Developer | All tests green |
| 14:00-16:00 | Code review and cleanup | Developer | Code quality verified |
| 16:00-18:00 | Release preparation | Developer | Tag v0.1.0, build artifacts |
| 18:00-20:00 | Buffer for issues | Developer | Issue resolution |
| 20:00-21:00 | Final checkpoint | Orchestrator | Release readiness review |

**Checkpoint:** v0.1.0 release ready

---

## Risk Mitigation

### Risk: Grace Period Logic Complexity
**Impact:** High | **Probability:** Medium

**Mitigation:**
- Start with simplest implementation (fixed delay)
- Add comprehensive unit tests before integration
- Test overlap scenarios (window restarts during grace period)

---

### Risk: Holiday ConfigMap Timezone Confusion
**Impact:** Medium | **Probability:** High

**Mitigation:**
- Document clearly that dates are checked in TWS timezone
- Add explicit test with different system/TWS timezones
- Log timezone used for holiday check

---

### Risk: Metrics Cardinality Explosion
**Impact:** Medium | **Probability:** Low

**Mitigation:**
- Limit label cardinality (use namespace/name, not UID)
- Document metric cardinality in operations guide
- Consider metric aggregation for large deployments

---

### Risk: E2E Tests Flaky on Timing
**Impact:** High | **Probability:** Medium

**Mitigation:**
- Use minute-scale windows (not second-scale)
- Add generous timeouts with Eventually()
- Mock time in controller for deterministic tests
- Run tests multiple times to detect flakes

---

### Risk: Time Runs Out
**Impact:** Critical | **Probability:** Low

**Mitigation:**
- Daily checkpoints catch delays early
- Buffer time built into Day 4
- Minimum viable: Tasks 1-3 for basic functionality
- Tasks 4-6 can slip to v0.1.1 if necessary

---

## Success Metrics

### Code Quality
- [ ] All unit tests pass (maintain 83.8%+ coverage)
- [ ] E2E tests pass reliably (5/5 scenarios)
- [ ] No critical or high severity linter warnings
- [ ] Code reviewed for clarity and maintainability

### Functionality
- [ ] Grace period delays downscaling correctly
- [ ] Holiday ConfigMap integration works for all modes
- [ ] Metrics endpoint exposes all expected metrics
- [ ] Cross-midnight windows work correctly
- [ ] Pause mode and drift correction functional

### Documentation
- [ ] README status accurate
- [ ] Metrics fully documented with examples
- [ ] Manual testing checklist complete
- [ ] No TODOs remaining in critical paths

### Deployment
- [ ] Controller builds and deploys successfully
- [ ] Examples work on fresh Kind cluster
- [ ] Quick Start guide validated end-to-end

---

## Handoff Protocol

### Daily Handoffs
At end of each day (19:00-23:00 IST):
1. Update task status (completed/in-progress/blocked)
2. Commit all changes with descriptive messages
3. Log any decisions made in DECISIONS.md
4. Note any blockers or risks for next day
5. Push to remote repository

### Final Handoff (2025-11-27 21:00 IST)
Deliverables:
1. All code committed and pushed
2. Git tag v0.1.0 created
3. Docker image built and tagged
4. Release notes drafted
5. VERIFICATION.md completed with test results
6. This plan updated with actual completion times

---

## Communication

### Status Updates
- Daily checkpoint times: 19:00 IST (days 1-2), 22:00 IST (day 3), 21:00 IST (day 4)
- Format: Completed tasks, in-progress tasks, blockers, next steps

### Escalation
If any task takes >150% of estimated time:
1. Reassess scope - can we simplify?
2. Check for blockers - do we need help?
3. Consider moving to v0.1.1 if not critical path

### Decision Log
All architectural or scope decisions logged in:
- /Users/aykumar/personal/kyklos/docs/DECISIONS.md (using ADR format)

---

## References

### Key Files
- Controller: /Users/aykumar/personal/kyklos/internal/controller/timewindowscaler_controller.go
- Engine: /Users/aykumar/personal/kyklos/internal/engine/schedule.go
- E2E Tests: /Users/aykumar/personal/kyklos/test/e2e/e2e_test.go
- CRD: /Users/aykumar/personal/kyklos/api/v1alpha1/timewindowscaler_types.go

### Documentation
- Project Brief: /Users/aykumar/personal/kyklos/docs/PROJECT_BRIEF.md
- Glossary: /Users/aykumar/personal/kyklos/docs/user/GLOSSARY.md
- Design Docs: /Users/aykumar/personal/kyklos/docs/design/

### Tools
- Build: `make help` for all targets
- Testing: `make test`, `make test-engine`, `make test-controller`
- Deployment: `make cluster-up`, `make deploy`, `make demo-setup`
