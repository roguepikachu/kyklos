# Risk Register - Kyklos v0.1

**Last Updated:** 2025-10-19 (Day 0)
**Owner:** kyklos-orchestrator
**Review Frequency:** Daily during design phase, weekly during implementation

This document tracks the top risks for Kyklos v0.1, their likelihood and impact, mitigation strategies, and validation approaches.

---

## Risk Assessment Matrix

| Level | Likelihood | Impact | Action Required |
|-------|------------|--------|-----------------|
| **Critical** | High | High | Immediate mitigation, daily monitoring |
| **High** | High | Medium or Medium | High | Mitigation plan within 1 day |
| **Medium** | Low | High or Medium | Medium | Monitor, mitigate if escalates |
| **Low** | Low | Low or Medium | Accept, document |

---

## Risk 1: DST and Cross-Midnight Correctness

**Risk ID:** RISK-001
**Category:** Technical - Correctness
**Level:** Critical
**Likelihood:** High (DST is inevitable)
**Impact:** High (incorrect scaling breaks user trust)

### Description
Daylight Saving Time transitions and cross-midnight time windows introduce complex edge cases:
- **DST Spring-Forward:** 02:00 → 03:00, skipping an hour. If window is 02:00-04:00, does it start at 03:00 or 01:00 the day before?
- **DST Fall-Back:** 02:00 → 01:00, repeating an hour. If window is 01:30-03:00, does it activate twice?
- **Cross-Midnight:** Window like 22:00-02:00 spans two dates. Current date comparison logic must handle "today 22:00" to "tomorrow 02:00" correctly.

### Impact if Not Mitigated
- Workloads may not scale at expected times
- Users lose confidence in operator reliability
- Support burden increases with timezone-specific issues
- Difficult to debug (timezone-dependent, date-specific failures)

### Mitigation Strategy
1. **Design Phase (Day 2-3):**
   - controller-reconcile-designer must document DST behavior explicitly in design/reconcile-requeue-logic.md
   - Create decision table: for each scenario (spring/fall, cross-midnight), what is expected behavior
   - Define precedence: if ambiguous, prefer user safety (don't scale down unexpectedly)

2. **Testing Phase (Day 4):**
   - testing-strategy-designer must create test fixtures with fixed DST transition dates:
     - 2025-03-09 (US spring-forward example)
     - 2025-11-02 (US fall-back example)
   - Test cases for each combination: DST + cross-midnight, DST + grace period
   - Test with multiple timezones: US (America/New_York), Europe (Europe/London), India (Asia/Kolkata - no DST)

3. **Implementation Phase:**
   - Use Go time.LoadLocation and wall clock time (not duration arithmetic)
   - Add debug logging for state transitions during DST dates
   - Metrics to track DST-adjacent scaling events

4. **Validation Phase (Day 7):**
   - demo-screenshot-designer includes DST scenario in demo (fast-forward to DST date)
   - Verify behavior matches documented decision table

### Validation Approach
- **Automated:** Unit tests with fixed dates covering all DST scenarios
- **Manual:** Run local cluster with system time set to DST transition date, verify scaling
- **Success Criteria:**
  - Zero unexpected scale events during DST transition week
  - Status field accurately reports next transition time across DST boundary

### Monitoring Plan
- **Design Phase:** Daily check that reconcile design addresses DST (Day 2-3)
- **Test Phase:** Verify DST test fixtures exist and pass (Day 4-6)
- **Implementation:** Run DST tests in CI on every commit

### Owner
**Design:** controller-reconcile-designer
**Testing:** testing-strategy-designer
**Validation:** kyklos-orchestrator (verify design addresses all cases)

### Current Status (Day 0)
- Acknowledged in ADR-0003
- Mitigation plan defined above
- Next action: Day 2 reconcile design must include DST decision table

---

## Risk 2: Overlapping Windows Semantics

**Risk ID:** RISK-002
**Category:** Design - Ambiguity
**Level:** High
**Likelihood:** Medium (grace period can cause overlap)
**Impact:** High (unclear behavior confuses users)

### Description
When grace period extends into next active window, or when multiple TimeWindowScalers target same workload, behavior is ambiguous:
- **Grace Overlap:** Window ends at 17:00, grace period is 30 minutes. Next window starts at 17:00 (if multiple TWS resources). Which wins?
- **Multiple TWS on Same Target:** Two TimeWindowScalers reference same Deployment. Which replica count takes precedence?
- **Spec Change During Grace:** User updates TWS spec while in grace period. Does grace cancel or continue?

### Impact if Not Mitigated
- Non-deterministic scaling behavior
- Race conditions between multiple controllers
- User confusion ("why didn't my workload scale when expected?")
- Difficult to troubleshoot (timing-dependent)

### Mitigation Strategy
1. **Design Phase (Day 1-2):**
   - api-crd-designer documents grace period precedence in design/api-field-semantics.md
   - Decision: If grace period overlaps with next active window, cancel grace and activate immediately
   - controller-reconcile-designer documents state machine: spec change resets grace period timer

2. **Scope Limitation (Day 0):**
   - v0.1 does NOT support multiple TimeWindowScalers targeting same workload
   - Validation webhook should reject (or warn) if target is already referenced by another TWS
   - Document in BRIEF.md non-goals: "Multiple TimeWindowScalers per target (v0.2 feature)"

3. **Status Field (Day 1):**
   - Add status.conflictingScalers field: list of other TWS resources referencing same target
   - Controller sets condition type: Scaling=False, Reason=ConflictDetected

4. **Documentation (Day 6):**
   - docs-dx-designer adds "Grace Period Behavior" section to CONCEPTS.md
   - Example scenarios with timeline diagrams

### Validation Approach
- **Automated:** Test case: create two TWS targeting same Deployment, verify one sets ConflictDetected condition
- **Manual:** Demo scenario with grace period extending into next window, verify immediate activation
- **Success Criteria:**
  - Zero ambiguous states in state machine design
  - Conflict detection works within 30 seconds

### Monitoring Plan
- **Design Phase:** Verify grace overlap logic documented (Day 2)
- **API Phase:** Verify status.conflictingScalers field exists (Day 1)
- **Test Phase:** Conflict detection test case exists and passes (Day 4)

### Owner
**Design:** api-crd-designer (precedence), controller-reconcile-designer (state machine)
**Validation:** observability-metrics-designer (conflict metrics)

### Current Status (Day 0)
- Acknowledged in ADR-0004 (grace period semantics)
- Decision: Cancel grace if next window starts
- Next action: Day 1 API design must include conflict detection field
- Next action: Day 2 reconcile design must document overlap handling

---

## Risk 3: Grace Period on Scale-Down Safety

**Risk ID:** RISK-003
**Category:** User Experience - Safety
**Level:** High
**Likelihood:** High (users will expect this)
**Impact:** Medium (data loss or service interruption)

### Description
Grace period delays scale-down, but does not guarantee workload has completed tasks:
- Grace period is controller-level delay, not Pod-level termination grace
- User expects "workload will finish processing before scaling down"
- Reality: Grace period expires, controller scales down, Pods receive SIGTERM with standard terminationGracePeriodSeconds
- If workload task takes longer than terminationGracePeriodSeconds, it is forcefully killed (SIGKILL)

### Impact if Not Mitigated
- Users lose data or requests (e.g., long-running batch job killed mid-process)
- Misunderstanding of grace period purpose leads to support requests
- Reputation: "Operator doesn't safely scale down my workloads"

### Mitigation Strategy
1. **Documentation (Day 6):**
   - docs-dx-designer adds "Grace Period vs Termination Grace" section to CONCEPTS.md
   - Explain: TWS grace period is "when to start scale-down", not "how long Pod shutdown takes"
   - Recommend: Set target Deployment's terminationGracePeriodSeconds to match or exceed expected task duration
   - Example: If task takes up to 10 minutes, set terminationGracePeriodSeconds=600, and TWS gracePeriod=15m (gives 5 min buffer)

2. **API Design (Day 1):**
   - Field comment for gracePeriod includes warning: "This delays scale-down initiation. Ensure target workload's terminationGracePeriodSeconds is sufficient for graceful shutdown."

3. **Status Condition (Day 3):**
   - Add condition type: GracePeriodActive=True with message: "Grace period active, scale-down delayed until {time}"
   - Helps users understand current state

4. **Metrics (Day 3):**
   - Metric: kyklos_grace_period_active (gauge, 1 during grace, 0 otherwise)
   - Allows external alerting on prolonged grace periods

### Validation Approach
- **Documentation Review:** Ensure CONCEPTS.md clearly explains grace period limitations
- **User Testing:** Ask naive user to read grace period docs, verify they understand it's not Pod termination grace
- **Success Criteria:**
  - Documentation includes example with terminationGracePeriodSeconds
  - Status condition shows grace period state

### Monitoring Plan
- **Design Phase:** API field comment includes warning (Day 1)
- **Documentation:** CONCEPTS.md section exists and is clear (Day 6)
- **Review Phase:** Request external review of grace period docs (Day 8)

### Owner
**Design:** api-crd-designer (field comment), controller-reconcile-designer (status condition)
**Docs:** docs-dx-designer (CONCEPTS.md explanation)
**Validation:** kyklos-orchestrator (review for clarity)

### Current Status (Day 0)
- Acknowledged in ADR-0004
- Mitigation plan defined above
- Next action: Day 1 API design must include field comment warning
- Next action: Day 6 docs must explain termination grace difference

---

## Risk 4: Namespace Target References and RBAC

**Risk ID:** RISK-004
**Category:** Security - RBAC
**Level:** Medium
**Likelihood:** Medium (cross-namespace is common request)
**Impact:** High (privilege escalation or permission denial)

### Description
TimeWindowScaler in namespace A may reference target in namespace B:
- **Over-Permissive RBAC:** ClusterRole grants controller access to all namespaces, allows scaling any workload (privilege escalation risk)
- **Under-Permissive RBAC:** Namespaced Role only allows same-namespace, cross-namespace references fail (user confusion)
- **Namespace Not Specified:** If targetRef.namespace is optional and defaults to TWS namespace, cross-namespace intent is unclear

### Impact if Not Mitigated
- Security audit fails due to overly broad permissions
- Users cannot use cross-namespace feature due to RBAC denial
- Inconsistent behavior: works in some clusters, fails in others based on RBAC setup

### Mitigation Strategy
1. **API Design (Day 1):**
   - targetRef.namespace is optional, defaults to TWS metadata.namespace
   - Field comment: "If targeting workload in different namespace, controller must have ClusterRole permissions"

2. **RBAC Design (Day 3):**
   - security-rbac-designer creates two RBAC profiles:
     - **Same-Namespace Mode:** Namespaced Role, only scales workloads in same namespace as TWS
     - **Cross-Namespace Mode:** ClusterRole, can scale workloads in any namespace
   - Default: Same-Namespace Mode (least privilege)
   - Document when to use each mode

3. **Validation (Day 2):**
   - Admission webhook does NOT validate target exists (would require cross-namespace GET permission in webhook)
   - Controller validates target exists during reconcile, sets condition if not found

4. **Status Condition (Day 3):**
   - Condition type: TargetFound=False if target doesn't exist or RBAC denies access
   - Reason: TargetNotFound vs RBACDenied (distinguish between missing target and permission issue)

5. **Documentation (Day 6):**
   - CONCEPTS.md explains same-namespace vs cross-namespace modes
   - Quick Start uses same-namespace (simpler setup)
   - Advanced example shows cross-namespace with ClusterRole setup

### Validation Approach
- **Automated Test:** Create TWS in ns-a targeting Deployment in ns-b, verify with ClusterRole it works, with Role it fails with clear error
- **Manual Test:** Follow Quick Start (same-namespace), verify works without ClusterRole
- **Success Criteria:**
  - Same-namespace works with minimal RBAC
  - Cross-namespace requires explicit ClusterRole (documented)
  - Status condition clearly indicates RBAC denial vs target not found

### Monitoring Plan
- **Design Phase:** API field defaults documented (Day 1), RBAC profiles created (Day 3)
- **Test Phase:** RBAC test cases exist (Day 4)
- **Documentation:** Cross-namespace example with RBAC setup (Day 6)

### Owner
**Design:** api-crd-designer (field default), security-rbac-designer (RBAC profiles)
**Validation:** testing-strategy-designer (RBAC test cases)
**Docs:** docs-dx-designer (RBAC setup guide)

### Current Status (Day 0)
- Acknowledged in ADR-0002 (cross-namespace support)
- Decision: Support both modes with different RBAC profiles
- Next action: Day 1 API design must document namespace default
- Next action: Day 3 RBAC design must create two profiles

---

## Risk 5: Demo Flakiness and Time-Based Testing

**Risk ID:** RISK-005
**Category:** Validation - Testability
**Level:** Medium
**Likelihood:** High (time-based tests are inherently flaky)
**Impact:** Medium (poor first impression, support burden)

### Description
Demo and E2E tests depend on real time passing:
- **Demo:** User follows Quick Start, creates TWS with window "now to now+10min", waits 10 minutes to see scale-down. Flakes if system clock skews or timing is tight.
- **E2E Tests:** Time-based assertions ("after 5 minutes, replica count should be X") fail if cluster is slow or CI is under load.
- **DST Tests:** Cannot easily test DST transitions without mocking time or waiting months.

### Impact if Not Mitigated
- Demo fails during presentations, undermines confidence
- E2E tests are flaky, block merges, team loses trust in CI
- Cannot validate DST behavior until actual DST dates (too late)

### Mitigation Strategy
1. **Local Workflow (Day 5):**
   - local-workflow-designer creates time-warp testing method:
     - Option A: Test build tag that uses mockable time interface
     - Option B: Script to adjust system time (libfaketime or VM snapshot with adjusted clock)
   - Quick Start demo uses immediate window (starts now, ends in 2 minutes) to minimize wait time

2. **Test Design (Day 4):**
   - testing-strategy-designer separates time-sensitive and time-independent tests:
     - **Unit Tests:** Use mocked time (Go interface or fixed time.Now), fully deterministic
     - **Envtest:** Fast-forward time via test harness, not real wall clock waits
     - **E2E Tests:** Use short time windows (1-2 minutes max), add retry logic and generous timeouts

3. **Demo Design (Day 7):**
   - demo-screenshot-designer creates pre-recorded terminal session (asciinema or screenshot sequence) as backup
   - Live demo has "cheat" option: TWS with window that started 1 hour ago (already active), demonstrates scale-up immediately
   - Demo script includes timing notes: "Wait 2 minutes for scale-down (get coffee)"

4. **CI Smoke Test (Day 7):**
   - Smoke test uses short window (window: "now to now+3min")
   - Verify scale-up immediately, wait 3 min with retries, verify scale-down
   - If smoke test takes >10 minutes, treat as CI failure (timing issue)

### Validation Approach
- **Unit Tests:** 100% of time calculation tests use mocked time, zero sleeps/waits
- **Demo:** Run demo script 3 times successfully before Day 7 sign-off
- **Smoke Test:** Smoke test passes 5 consecutive CI runs without flake
- **Success Criteria:**
  - Demo completes in under 15 minutes including wait time
  - E2E tests have <5% flake rate over 20 runs

### Monitoring Plan
- **Design Phase:** Time mock strategy documented (Day 4-5)
- **Testing Phase:** Track E2E test pass rate daily (Day 6+)
- **Review Phase:** Demo dry-run on Day 7, measure actual timing

### Owner
**Design:** testing-strategy-designer (time mock), local-workflow-designer (time-warp)
**Execution:** demo-screenshot-designer (demo script)
**Validation:** ci-release-designer (smoke test stability)

### Current Status (Day 0)
- Risk identified, not yet mitigated
- Next action: Day 4 test plan must include time mocking strategy
- Next action: Day 5 local workflow must document time-warp testing
- Next action: Day 7 demo script must have timing buffer and backup recording

---

## Risk Summary Table

| Risk ID | Risk Name | Level | Owner | Mitigation Deadline | Status |
|---------|-----------|-------|-------|---------------------|--------|
| RISK-001 | DST and Cross-Midnight Correctness | Critical | controller-reconcile-designer | Day 3 | Identified |
| RISK-002 | Overlapping Windows Semantics | High | api-crd-designer, controller-reconcile-designer | Day 2 | Identified |
| RISK-003 | Grace Period Safety | High | docs-dx-designer, api-crd-designer | Day 6 | Identified |
| RISK-004 | Namespace RBAC | Medium | security-rbac-designer | Day 3 | Identified |
| RISK-005 | Demo Flakiness | Medium | testing-strategy-designer, demo-screenshot-designer | Day 7 | Identified |

---

## Risk Escalation Protocol

### When to Escalate
- Risk level increases (e.g., Medium → High)
- Mitigation strategy fails or is insufficient
- New risk discovered that impacts critical path
- Risk deadline missed

### How to Escalate
1. Risk owner updates this document with new status
2. Tag kyklos-orchestrator in commit message
3. kyklos-orchestrator assesses within 2 hours
4. Options: extend deadline, reallocate resources, reduce scope, accept risk
5. Decision logged in DECISIONS.md if scope/timeline changes
6. All affected agents notified

### Acceptance Criteria for Risk Closure
Risk can be marked "Mitigated" or "Closed" when:
- All mitigation actions completed and documented
- Validation approach executed successfully
- Owner and kyklos-orchestrator sign off
- No open questions or unresolved concerns

---

## Daily Risk Review (During Design Phase)

**Time:** End of day (18:00 IST)
**Owner:** kyklos-orchestrator

Review checklist:
- [ ] Are any risks at-risk of missing mitigation deadline?
- [ ] Have any new risks been identified today?
- [ ] Do any risk levels need to be adjusted?
- [ ] Are mitigation actions on track?

If any answer is concerning, trigger escalation protocol.

---

## Risk Metrics

Track weekly:
- Number of Critical/High risks open
- Percentage of risks with mitigation plan on track
- Number of new risks identified vs closed
- Risk escalations per week

Goal: Zero Critical risks at implementation kickoff (Day 15).

---

## Appendix: Risk Brainstorm (Day 0)

Other risks considered but assessed as Low or out of scope for v0.1:

**Operator HA (High Availability):**
- Level: Low (single replica acceptable for v0.1)
- Mitigation: Document as v0.2 feature

**Performance at Scale:**
- Level: Low (assume <100 TWS resources per cluster)
- Mitigation: Note in BRIEF.md constraints

**Timezone Database Updates:**
- Level: Low (Go embeds timezone data, updates with Go version)
- Mitigation: Document Go version dependency

**Cluster Clock Skew:**
- Level: Low (assume NTP synchronized)
- Mitigation: Assume in BRIEF.md, document NTP requirement in prerequisites

**Controller Crash During Scale:**
- Level: Low (Kubernetes retries, scale is idempotent)
- Mitigation: Controller runtime handles restarts

These are acknowledged but not actively tracked. Revisit if evidence suggests higher likelihood or impact.
