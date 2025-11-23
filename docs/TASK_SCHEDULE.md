# Kyklos v0.1 Task Schedule

**Created:** 2025-11-23 20:30 IST
**Target Completion:** 2025-11-27 23:59 IST
**Status:** ACTIVE

## Task Registry

### Task 1: Grace Period Timing Logic
- **ID:** KYKLOS-001
- **Owner:** Developer
- **Status:** PENDING
- **Priority:** P0 (Critical Path)
- **Estimated:** 3 hours
- **Deadline:** 2025-11-24 13:00 IST
- **Dependencies:** None
- **Files:**
  - /Users/aykumar/personal/kyklos/internal/engine/schedule.go
  - /Users/aykumar/personal/kyklos/internal/engine/schedule_test.go
- **Acceptance Criteria:**
  - [ ] Engine checks if scaling down AND LastScaleTime + GracePeriodSecs > Now
  - [ ] Returns CurrentReplicas with reason "grace-period-active" during grace period
  - [ ] NextBoundary set to grace period end time
  - [ ] Scale-up bypasses grace period (immediate)
  - [ ] Unit tests cover all scenarios
  - [ ] Integration test shows delayed downscaling

---

### Task 2: Holiday ConfigMap Reading
- **ID:** KYKLOS-002
- **Owner:** Developer
- **Status:** PENDING
- **Priority:** P0 (Critical Path)
- **Dependencies:** KYKLOS-001
- **Estimated:** 4 hours
- **Deadline:** 2025-11-24 18:00 IST
- **Files:**
  - /Users/aykumar/personal/kyklos/internal/controller/timewindowscaler_controller.go
  - /Users/aykumar/personal/kyklos/config/rbac/role.yaml (auto-generated)
- **Acceptance Criteria:**
  - [ ] RBAC markers added for ConfigMap get/list
  - [ ] readHolidayConfigMap() helper function implemented
  - [ ] Date formatted as YYYY-MM-DD in TWS timezone
  - [ ] Missing ConfigMap sets Degraded condition
  - [ ] IsHoliday correctly passed to engine
  - [ ] All three holiday modes tested (ignore, closed, open)
  - [ ] Unit tests mock ConfigMap client

---

### Task 3: Prometheus Metrics Implementation
- **ID:** KYKLOS-003
- **Owner:** Developer
- **Status:** PENDING
- **Priority:** P1 (Required for Release)
- **Dependencies:** KYKLOS-002
- **Estimated:** 5 hours
- **Deadline:** 2025-11-25 15:00 IST
- **Files:**
  - /Users/aykumar/personal/kyklos/internal/controller/metrics.go (new)
  - /Users/aykumar/personal/kyklos/internal/controller/timewindowscaler_controller.go
- **Acceptance Criteria:**
  - [ ] metrics.go created with 4 metric definitions
  - [ ] Metrics registered in controller setup
  - [ ] scaleOperationsTotal incremented on scale operations
  - [ ] currentEffectiveReplicas gauge updated in reconcile
  - [ ] windowTransitionsTotal tracks window changes
  - [ ] reconcileDuration histogram measures reconcile time
  - [ ] Metrics endpoint accessible (:8443/metrics)
  - [ ] curl test confirms metrics present

---

### Task 4: E2E Test Scenarios
- **ID:** KYKLOS-004
- **Owner:** Developer
- **Status:** PENDING
- **Priority:** P1 (Required for Release)
- **Dependencies:** KYKLOS-003
- **Estimated:** 8 hours
- **Deadline:** 2025-11-26 18:00 IST
- **Files:**
  - /Users/aykumar/personal/kyklos/test/e2e/e2e_test.go
- **Test Cases:**
  1. [ ] Time window scaling (window active/inactive)
  2. [ ] Grace period delays scale-down
  3. [ ] Holiday ConfigMap integration (all modes)
  4. [ ] Cross-midnight window behavior
  5. [ ] Pause mode prevents scaling
- **Acceptance Criteria:**
  - [ ] All 5 test cases pass reliably
  - [ ] Tests use minute-scale windows (<5 min total runtime)
  - [ ] Tests clean up resources
  - [ ] Failure diagnostics emitted
  - [ ] Can run via `make test-e2e`

---

### Task 5: Integration Verification
- **ID:** KYKLOS-005
- **Owner:** Developer
- **Status:** PENDING
- **Priority:** P1 (Required for Release)
- **Dependencies:** KYKLOS-004
- **Estimated:** 4 hours
- **Deadline:** 2025-11-26 22:00 IST
- **Deliverable:** /Users/aykumar/personal/kyklos/docs/VERIFICATION.md
- **Manual Test Checklist:**
  - [ ] Fresh Kind cluster created
  - [ ] Controller builds and deploys successfully
  - [ ] Basic scaling scenario works (office hours example)
  - [ ] Holiday ConfigMap tested (all modes)
  - [ ] Grace period verified (delay observed)
  - [ ] Cross-midnight window tested
  - [ ] Pause mode and drift correction validated
  - [ ] Metrics scraped and verified
  - [ ] Logs show expected scaling decisions
  - [ ] Status conditions accurate
  - [ ] Events emitted correctly

---

### Task 6: Documentation Updates
- **ID:** KYKLOS-006
- **Owner:** Developer
- **Status:** PENDING
- **Priority:** P2 (Release Quality)
- **Dependencies:** KYKLOS-005
- **Estimated:** 2 hours
- **Deadline:** 2025-11-27 12:00 IST
- **Files:**
  - /Users/aykumar/personal/kyklos/README.md
  - /Users/aykumar/personal/kyklos/docs/user/OPERATIONS.md
  - /Users/aykumar/personal/kyklos/docs/ROADMAP.md
- **Acceptance Criteria:**
  - [ ] README status checkmarks updated
  - [ ] No "structure in place" notes remain
  - [ ] OPERATIONS.md includes full metrics section
  - [ ] PromQL examples provided
  - [ ] Alerting rules suggested
  - [ ] ROADMAP.md marks v0.1 complete
  - [ ] No outdated information

---

## Daily Checkpoint Schedule

### Checkpoint 1: Day 1 End (2025-11-24 19:00 IST)
**Expected Status:**
- KYKLOS-001: COMPLETE
- KYKLOS-002: COMPLETE
- Grace period and holiday ConfigMap working

**Actions:**
- [ ] Commit all changes
- [ ] Update task status
- [ ] Log any decisions in DECISIONS.md
- [ ] Note blockers for Day 2

---

### Checkpoint 2: Day 2 End (2025-11-25 17:00 IST)
**Expected Status:**
- KYKLOS-003: COMPLETE
- Metrics endpoint functional

**Actions:**
- [ ] Commit all changes
- [ ] Test metrics locally
- [ ] Update task status
- [ ] Note blockers for Day 3

---

### Checkpoint 3: Day 3 End (2025-11-26 22:00 IST)
**Expected Status:**
- KYKLOS-004: COMPLETE
- KYKLOS-005: COMPLETE
- All tests passing, manual verification complete

**Actions:**
- [ ] Commit all changes
- [ ] VERIFICATION.md complete
- [ ] Update task status
- [ ] Prepare for Day 4 documentation

---

### Checkpoint 4: Day 4 End (2025-11-27 21:00 IST)
**Expected Status:**
- KYKLOS-006: COMPLETE
- All tasks COMPLETE
- Release ready

**Actions:**
- [ ] Commit all changes
- [ ] Create git tag v0.1.0
- [ ] Build release artifacts
- [ ] Final release readiness review

---

## Progress Tracking

### Day 1: 2025-11-24
| Time (IST) | Planned Activity | Actual Activity | Status |
|------------|------------------|-----------------|--------|
| 10:00-13:00 | KYKLOS-001 | TBD | - |
| 14:00-18:00 | KYKLOS-002 | TBD | - |
| 19:00 | Checkpoint 1 | TBD | - |

### Day 2: 2025-11-25
| Time (IST) | Planned Activity | Actual Activity | Status |
|------------|------------------|-----------------|--------|
| 10:00-15:00 | KYKLOS-003 | TBD | - |
| 15:00-16:00 | Metrics verification | TBD | - |
| 17:00 | Checkpoint 2 | TBD | - |

### Day 3: 2025-11-26
| Time (IST) | Planned Activity | Actual Activity | Status |
|------------|------------------|-----------------|--------|
| 10:00-18:00 | KYKLOS-004 | TBD | - |
| 18:00-22:00 | KYKLOS-005 | TBD | - |
| 22:00 | Checkpoint 3 | TBD | - |

### Day 4: 2025-11-27
| Time (IST) | Planned Activity | Actual Activity | Status |
|------------|------------------|-----------------|--------|
| 10:00-12:00 | KYKLOS-006 | TBD | - |
| 12:00-14:00 | Final testing | TBD | - |
| 14:00-16:00 | Code review | TBD | - |
| 16:00-18:00 | Release prep | TBD | - |
| 21:00 | Final checkpoint | TBD | - |

---

## Risk Register

### Active Risks

| ID | Risk | Impact | Probability | Mitigation | Status |
|----|------|--------|-------------|------------|--------|
| R1 | Grace period logic complex | High | Medium | Start simple, add tests first | OPEN |
| R2 | Holiday timezone confusion | Medium | High | Document clearly, test multiple TZ | OPEN |
| R3 | Metrics cardinality explosion | Medium | Low | Limit labels to name/namespace | OPEN |
| R4 | E2E tests flaky on timing | High | Medium | Use minute-scale, generous timeouts | OPEN |
| R5 | Timeline slip | Critical | Low | Daily checkpoints, buffer on Day 4 | OPEN |

---

## Blockers and Issues

### Current Blockers
*None yet - to be updated during implementation*

### Resolved Blockers
*To be filled as blockers are resolved*

---

## Definition of Done

A task is COMPLETE when:
1. Code changes committed and pushed
2. All acceptance criteria checked off
3. Unit tests pass (`make test`)
4. Integration tests pass (if applicable)
5. Code reviewed for clarity
6. No compiler warnings or linter errors
7. Documentation updated if needed
8. Checkpoint status updated in this file

---

## Communication Protocol

### Status Update Format
```
Task: KYKLOS-XXX
Status: PENDING | IN_PROGRESS | BLOCKED | COMPLETE
Progress: X% or "Started" | "50%" | "Testing" | "Done"
Blockers: None or description
Next Steps: What's next
ETA: Estimated completion time
```

### Escalation Triggers
- Task takes >150% of estimated time
- Blocker cannot be resolved within 2 hours
- Test coverage drops below 80%
- Critical bug discovered

---

## Handoff Package Contents

On completion (2025-11-27 21:00 IST), deliver:
1. All code committed to main branch
2. Git tag v0.1.0 created
3. This file updated with actual completion times
4. VERIFICATION.md completed with test results
5. DECISIONS.md updated with any new ADRs
6. README.md and docs accurate
7. Release notes drafted

---

## Appendix: Quick Reference

### Key Commands
```bash
# Build and test
make build
make test
make test-engine
make test-controller

# Local deployment
make cluster-up
make docker-build kind-load
make install-crds
make deploy

# Demo and verification
make demo-setup
make demo-apply-minute
make demo-watch

# Cleanup
make cluster-down
make undeploy
make uninstall
```

### Key Files
- Controller: /Users/aykumar/personal/kyklos/internal/controller/timewindowscaler_controller.go
- Engine: /Users/aykumar/personal/kyklos/internal/engine/schedule.go
- Tests: /Users/aykumar/personal/kyklos/test/e2e/e2e_test.go
- CRD: /Users/aykumar/personal/kyklos/api/v1alpha1/timewindowscaler_types.go

### Documentation
- Project Brief: /Users/aykumar/personal/kyklos/docs/PROJECT_BRIEF.md
- Implementation Plan: /Users/aykumar/personal/kyklos/docs/IMPLEMENTATION_PLAN.md
- Decisions: /Users/aykumar/personal/kyklos/docs/DECISIONS.md
- Glossary: /Users/aykumar/personal/kyklos/docs/user/GLOSSARY.md
