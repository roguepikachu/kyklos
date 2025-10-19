# Kyklos v0.1 Roadmap

**Planning Start:** 2025-10-19 (Day 0)
**Design Phase:** 2025-10-20 to 2025-10-26 (Days 1-7)
**Review Phase:** 2025-10-27 to 2025-10-31 (Days 8-12)
**Lock Phase:** 2025-11-01 to 2025-11-02 (Days 13-14)
**Target Sign-Off:** 2025-11-02 18:00 IST

All times in IST (Asia/Kolkata timezone).

---

## Day 0: Planning Package Complete (2025-10-19)

**Status:** In Progress
**Owner:** kyklos-orchestrator

### Deliverables
- [x] docs/BRIEF.md - One-page project brief with glossary
- [x] docs/DECISIONS.md - ADR log with initial decisions
- [x] docs/RACI.md - RACI matrix for all workstreams
- [x] docs/QUALITY-GATES.md - Acceptance criteria for each gate
- [x] docs/ROADMAP.md - Two-week timeline (this file)
- [x] docs/HANDOFFS-DAY1.md - Day 1 handoff specifications
- [x] docs/REPO-LAYOUT.md - Repository structure definition
- [x] docs/COMMUNICATION.md - Communication protocols
- [x] docs/RISKS.md - Top 5 risks with mitigations

### Completion Criteria
- All 9 planning documents committed to main branch
- Day 1 agents have clear handoff specifications
- No placeholder content or open questions blocking Day 1 work

---

## Day 1: API and CRD Draft (2025-10-20)

**Owner:** api-crd-designer
**Consulted:** api-validation-defaults-designer, controller-reconcile-designer

### Deliverables
- design/api-crd-spec.md - Complete CRD schema design
- design/api-field-semantics.md - Field-by-field documentation
- design/api-status-design.md - Status subresource specification
- Quality Gate 1 completion checklist

### Acceptance Criteria
- TimeWindowScaler CRD schema with all fields defined
- OpenAPI validation rules documented
- Status conditions and observedGeneration strategy specified
- Cross-midnight window support explained
- Glossary terms from BRIEF.md used consistently

### Handoff To
- api-validation-defaults-designer (for validation rules)
- controller-reconcile-designer (for reconcile logic)

### Risks
- Field naming conflicts with existing Kubernetes types
- Ambiguous semantics for grace period overlap scenarios

---

## Day 2: Reconcile Design and Metrics Draft (2025-10-21)

**Parallel Tracks:**

### Track A: Validation and Defaults
**Owner:** api-validation-defaults-designer
**Consulted:** api-crd-designer, security-rbac-designer

**Deliverables:**
- design/validation-rules.md - Complete validation logic
- design/default-values.md - Default value strategy
- design/admission-webhook-design.md - Webhook architecture
- Quality Gate 2 completion checklist

**Acceptance Criteria:**
- All validation rules documented with test cases
- Default values specified for optional fields
- Webhook design includes TLS and failure policy
- Test matrix for valid and invalid inputs

### Track B: Reconcile Design
**Owner:** controller-reconcile-designer
**Consulted:** api-crd-designer, observability-metrics-designer

**Deliverables:**
- design/reconcile-state-machine.md - State transition logic
- design/reconcile-requeue-logic.md - Timing calculations
- design/reconcile-error-handling.md - Error and retry strategy
- Quality Gate 3 completion checklist

**Acceptance Criteria:**
- State machine covers all transitions (Inactive/Active/GracePeriod)
- Requeue timing logic handles DST and cross-midnight
- Error handling specified for target not found, scale failures
- Idempotency guarantees documented

---

## Day 3: Security RBAC and Metrics Plan (2025-10-22)

**Parallel Tracks:**

### Track A: RBAC Design
**Owner:** security-rbac-designer
**Consulted:** api-validation-defaults-designer, controller-reconcile-designer

**Deliverables:**
- design/rbac-permissions.md - Permission matrix with justifications
- design/rbac-same-namespace.md - Same-namespace RBAC model
- design/rbac-cross-namespace.md - Cross-namespace RBAC model
- Quality Gate 5 completion checklist

**Acceptance Criteria:**
- Minimal permissions for controller ServiceAccount
- Separate RBAC for same-namespace and cross-namespace modes
- Webhook ServiceAccount permissions specified
- No privilege escalation paths

### Track B: Observability Design
**Owner:** observability-metrics-designer
**Consulted:** controller-reconcile-designer, docs-dx-designer

**Deliverables:**
- design/metrics-specification.md - Prometheus metrics with labels
- design/status-conditions.md - Condition types and transitions
- design/events-logging.md - Event emission and log levels
- Quality Gate 4 completion checklist

**Acceptance Criteria:**
- Metrics cover state, scale events, errors, reconcile duration
- Status conditions map to user-visible states
- Events for all scale operations and errors
- Logging levels specified for debug and production

---

## Day 4: Test Plan (2025-10-23)

**Owner:** testing-strategy-designer
**Consulted:** controller-reconcile-designer, local-workflow-designer

### Deliverables
- design/test-plan-unit.md - Unit test scenarios and coverage
- design/test-plan-envtest.md - Envtest integration scenarios
- design/test-plan-e2e.md - E2E test design with time mocking
- design/test-plan-dst.md - DST edge case test data
- Quality Gate 7 completion checklist

### Acceptance Criteria
- Unit tests for time calculation and state machine logic
- Envtest scenarios for reconcile loop with mocked targets
- E2E test design with time-warp or fast-forward method
- DST test cases for spring-forward and fall-back
- Test data with fixed dates for reproducibility
- Coverage target 80% for controller, 100% for critical paths

### Handoff To
- ci-release-designer (for CI test integration)
- local-workflow-designer (for local test execution)

---

## Day 5: CI and Release Plan (2025-10-24)

**Parallel Tracks:**

### Track A: Local Workflow
**Owner:** local-workflow-designer
**Consulted:** testing-strategy-designer, docs-dx-designer

**Deliverables:**
- design/local-setup.md - Kind/minikube quick start
- design/local-testing.md - Time-warp testing methodology
- scripts/quick-start.sh - Automated setup script
- examples/sample-timewindowscaler.yaml - Demo resource
- Quality Gate 6 completion checklist

**Acceptance Criteria:**
- Quick start script completes in under 15 minutes
- Sample TWS demonstrates immediate visible scaling
- Troubleshooting guide for common local issues
- Time-warp testing allows fast validation of time windows

### Track B: CI Design
**Owner:** ci-release-designer
**Consulted:** testing-strategy-designer, security-rbac-designer

**Deliverables:**
- design/ci-pipeline.md - GitHub Actions workflow design
- design/ci-smoke-test.md - Smoke test specification
- design/release-process.md - Versioning and artifact publishing
- Quality Gate 8 completion checklist

**Acceptance Criteria:**
- CI runs lint, unit tests, envtest, build, smoke test
- Smoke test completes in under 10 minutes
- Container image build with security scanning
- Release process for tagging and publishing

---

## Day 6: README and Concepts Skeleton (2025-10-25)

**Owner:** docs-dx-designer
**Consulted:** local-workflow-designer, api-crd-designer, community-launch-designer

### Deliverables
- README.md - Complete user-facing README
- docs/CONCEPTS.md - Detailed conceptual documentation
- docs/TROUBLESHOOTING.md - Common issues and solutions
- examples/ - Directory with multiple example scenarios
- Quality Gate 9 completion checklist

### Acceptance Criteria
- README Quick Start in under 5 minutes (assumes cluster exists)
- Concepts explain time windows, DST, grace periods, state machine
- Examples: basic, cross-midnight, grace period, cross-namespace
- Troubleshooting covers controller logs, RBAC, time window validation
- All commands are copy-pasteable and tested

### Handoff To
- community-launch-designer (for launch content)
- demo-screenshot-designer (for visual content)

---

## Day 7: Demo Scenarios (2025-10-26)

**Owner:** demo-screenshot-designer
**Consulted:** docs-dx-designer, local-workflow-designer

### Deliverables
- design/demo-scenarios.md - Demo script with timing
- examples/demo/ - Demo YAML and setup scripts
- docs/architecture-diagram.png - System architecture visual
- Terminal recordings or GIFs for README

### Acceptance Criteria
- Demo shows: create TWS, see scale-up, wait for scale-down, verify metrics
- Demo runs in under 10 minutes with fast-forward time
- Architecture diagram shows controller, CRD, target workload flow
- Terminal recordings are clear and well-timed
- Demo is reproducible on local cluster

---

## Days 8-12: Review, Tighten, Resolve Comments (2025-10-27 to 2025-10-31)

**Owner:** All agents (round-robin review)
**Coordinator:** kyklos-orchestrator

### Daily Check-Ins
- 10:00 IST: Status update (blockers, progress, risks)
- 18:00 IST: Day wrap-up (completed reviews, open comments)

### Activities
- **Day 8 (Oct 27):** Cross-review all design documents
  - Each agent reviews 2-3 documents outside their domain
  - Log review comments in design/reviews/ directory

- **Day 9 (Oct 28):** Address critical gaps and inconsistencies
  - Focus on cross-document conflicts
  - Update DECISIONS.md with any new ADRs

- **Day 10 (Oct 29):** Implementation feasibility check
  - Verify designs are implementable in 2-week code sprint
  - Flag any design decisions that increase complexity

- **Day 11 (Oct 30):** User experience review
  - Walk through entire workflow from user perspective
  - Ensure documentation matches design reality

- **Day 12 (Oct 31):** Final polish and gap closure
  - All open comments resolved or escalated
  - All quality gates verified as achievable

### Completion Criteria
- Zero critical unresolved comments
- All quality gates have clear pass/fail criteria
- Documentation is consistent across all files
- No scope creep beyond BRIEF.md non-goals

---

## Day 13: Scope Lock and Issue Board (2025-11-01)

**Owner:** kyklos-orchestrator
**Participants:** All agents (2-hour working session 14:00-16:00 IST)

### Deliverables
- docs/SCOPE-LOCK.md - Frozen scope statement
- GitHub Issues created for all implementation tasks
- GitHub Milestones: v0.1-implementation, v0.1-testing, v0.1-docs
- GitHub Project board with columns: Backlog, In Progress, Review, Done

### Activities
- Review all design documents and extract implementation tasks
- Create Issues with labels: api, controller, testing, docs, ci
- Assign Issues to appropriate agents (implementation phase)
- Estimate effort for each Issue (S/M/L)
- Sequence Issues by dependencies
- Identify critical path and potential bottlenecks

### Completion Criteria
- Scope lock document signed off by all agents
- All design work is captured as Issues
- Project board shows clear 2-week implementation plan
- No ambiguous or un-scoped work items

---

## Day 14: Design Package Sign-Off (2025-11-02)

**Owner:** kyklos-orchestrator
**Final Approval:** 18:00 IST

### Deliverables
- docs/DESIGN-SIGNOFF.md - Formal approval document
- All design documents marked as "Approved for Implementation"
- Implementation kickoff scheduled for Day 15 (Nov 3)

### Sign-Off Checklist
- [ ] All 9 quality gates verified
- [ ] BRIEF.md reflects current understanding
- [ ] DECISIONS.md has no unresolved ADRs
- [ ] RACI.md has no ownership gaps
- [ ] ROADMAP.md shows all work completed on schedule
- [ ] RISKS.md top risks have mitigations
- [ ] Implementation Issue board is ready
- [ ] All agents confirm readiness to begin coding

### Post-Sign-Off
- Archive design phase artifacts
- Transition to implementation phase (Day 15+)
- Daily standups begin (10:00 IST)
- Weekly demos to stakeholders

---

## Risk Mitigation Timeline

| Risk | Mitigation Checkpoint | Owner |
|------|----------------------|-------|
| DST edge cases not testable | Day 4: Test plan must include fixed date fixtures | testing-strategy-designer |
| Cross-midnight complexity | Day 2: Reconcile design must show state machine trace | controller-reconcile-designer |
| 15-minute local setup too ambitious | Day 5: Measure actual time and adjust if > 15 min | local-workflow-designer |
| RBAC too permissive | Day 3: Security review with permission justification | security-rbac-designer |
| Demo flakiness | Day 7: Demo must run 3 times successfully | demo-screenshot-designer |

---

## Escalation Path

If any day's work is at risk of missing deadline:
1. Owner notifies kyklos-orchestrator at least 12 hours before due time
2. kyklos-orchestrator assesses: scope reduction, timeline extension, or resource reallocation
3. Decision logged in DECISIONS.md
4. All affected agents notified within 2 hours
5. ROADMAP.md updated to reflect change

---

## Success Metrics for Design Phase

At end of Day 14, we should have:
- 100% of quality gates passed
- Zero unresolved critical design questions
- Complete Issue board for 2-week implementation
- Confidence level > 80% that v0.1 is achievable in 2 weeks
- README and examples ready for first user
- CI pipeline design ready to implement

If any metric is not met, implementation phase does not begin until gaps are closed.
