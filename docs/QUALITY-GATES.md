# Quality Gates - Kyklos v0.1

Each work stream must meet these acceptance criteria before handoff to next phase.

## Gate 1: API Spec Semantics and Precedence

**Owner:** api-crd-designer
**Due:** Day 1 End (2025-10-20 18:00 IST)

### Acceptance Criteria
1. CRD YAML defines TimeWindowScaler with all fields documented
2. Spec includes: timezone, windowStart, windowEnd, activeReplicas, inactiveReplicas, gracePeriod, targetRef
3. Status includes: currentState (Active/Inactive/GracePeriod), observedGeneration, lastScaleTime, conditions array
4. OpenAPI schema validation enforces: timezone non-empty, time format HH:MM, replicas >= 0, gracePeriod <= 60m
5. Field comments explain precedence (e.g., gracePeriod=0 means immediate scale-down)
6. Cross-midnight windows explicitly supported via documentation

### Verification Steps
- Run `kubectl explain timewindowscaler.spec` and see all fields
- Validate invalid YAML is rejected (bad time format, negative replicas)
- Generate API reference docs from CRD comments

### Exit Criteria
- api-validation-defaults-designer and controller-reconcile-designer sign off
- No ambiguous field semantics
- All glossary terms from BRIEF.md used consistently

---

## Gate 2: Validation and Defaults Mapping

**Owner:** api-validation-defaults-designer
**Due:** Day 2 End (2025-10-21 18:00 IST)

### Acceptance Criteria
1. Document default values: gracePeriod=0, inactiveReplicas=0
2. Validation logic prevents: start=end time, invalid IANA timezone, cross-field conflicts
3. Admission webhook design document covering: validating webhook config, TLS cert strategy, failure policy
4. Test matrix for validation cases: valid, invalid timezone, invalid time, negative replicas, grace > 60m
5. Decision on whether targetRef namespace defaults to TWS namespace or requires explicit value

### Verification Steps
- Table mapping each validation rule to rejection message
- Pseudo-code or logic flow for each validator
- Test case list with input/output examples

### Exit Criteria
- controller-reconcile-designer confirms validation catches all error states
- security-rbac-designer confirms webhook security model
- No validation gaps that could cause runtime panics

---

## Gate 3: Reconcile Design with Requeue Logic

**Owner:** controller-reconcile-designer
**Due:** Day 3 End (2025-10-22 18:00 IST)

### Acceptance Criteria
1. State machine diagram showing: Inactive -> Active -> GracePeriod -> Inactive transitions
2. Requeue timing logic: calculate next transition, add jitter, handle errors
3. Pseudo-code for reconcile loop showing: fetch target, determine current state, scale if needed, update status
4. Error handling: target not found, scale API failure, timezone error
5. Edge case handling: DST transitions, cross-midnight windows, grace period overlap with next window
6. Idempotency guarantee: repeated reconciles don't cause thrashing

### Verification Steps
- Walk through state machine for full 24-hour cycle
- Trace DST spring-forward and fall-back scenarios
- Confirm requeue timing for each state transition
- Verify no infinite loops or missed transitions

### Exit Criteria
- observability-metrics-designer confirms all state transitions are observable
- testing-strategy-designer confirms design is testable with mocked time
- No race conditions or consistency gaps identified

---

## Gate 4: Metrics Guidance

**Owner:** observability-metrics-designer
**Due:** Day 3 End (2025-10-22 18:00 IST)

### Acceptance Criteria
1. Define metrics: kyklos_window_state (gauge), kyklos_scale_total (counter), kyklos_scale_errors_total (counter), kyklos_reconcile_duration_seconds (histogram)
2. Metric labels: timewindowscaler_name, timewindowscaler_namespace, target_kind, target_name, target_namespace
3. Status condition types: Ready, Scaling, TimezoneValid, TargetFound
4. Event types: Normal (ScaledUp, ScaledDown, GracePeriodStarted), Warning (ScaleFailed, TargetNotFound)
5. Logging levels: Info for state transitions, Debug for requeue calculations, Error for failures

### Verification Steps
- List each metric with help text and label set
- Map each status condition to triggering scenario
- Confirm metrics cover success and failure paths

### Exit Criteria
- controller-reconcile-designer confirms metrics align with reconcile logic
- docs-dx-designer confirms metrics are documentable
- All state transitions and errors are observable

---

## Gate 5: RBAC Least Privilege

**Owner:** security-rbac-designer
**Due:** Day 4 End (2025-10-23 18:00 IST)

### Acceptance Criteria
1. Controller ServiceAccount with minimal permissions
2. Role for same-namespace mode: get/list/watch TWS, update status, get/patch Deployments/StatefulSets/ReplicaSets scale subresource
3. ClusterRole for cross-namespace mode: same as above but cluster-scoped
4. Webhook ServiceAccount with only webhook-related permissions
5. No unnecessary permissions (e.g., delete, escalate, impersonate)
6. Documentation of permission matrix with justification for each verb

### Verification Steps
- RBAC YAML files for both namespace and cluster modes
- Permission audit table showing: resource, verbs, justification
- Test plan to verify controller fails gracefully without excess permissions

### Exit Criteria
- api-validation-defaults-designer confirms webhook RBAC is sufficient
- controller-reconcile-designer confirms reconcile RBAC is sufficient
- No privilege escalation paths identified

---

## Gate 6: Local Workflow (15 Minutes to Visible Scale Change)

**Owner:** local-workflow-designer
**Due:** Day 5 End (2025-10-24 18:00 IST)

### Acceptance Criteria
1. Quick start script: kind cluster up, install CRD, deploy controller, create sample TWS, verify scaling
2. Total time from clone to visible scale event: under 15 minutes
3. Sample TimeWindowScaler YAML with immediate window (starts now, ends in 10 minutes)
4. Troubleshooting steps: check controller logs, verify RBAC, validate time window
5. Time-warp testing method: override system time or use test-friendly time injection

### Verification Steps
- Run script on clean machine (no pre-existing cluster)
- Time each step and identify bottlenecks
- Verify user sees: TWS created, Deployment scaled up, metrics updated, status shows Active

### Exit Criteria
- testing-strategy-designer confirms test scenarios are runnable locally
- docs-dx-designer confirms workflow is documentable in README
- No manual steps that block automation

---

## Gate 7: Test Plan Coverage

**Owner:** testing-strategy-designer
**Due:** Day 6 End (2025-10-25 18:00 IST)

### Acceptance Criteria
1. Unit tests: time calculation, state machine logic, validation functions
2. Envtest scenarios: reconcile with mocked target, status updates, error cases
3. E2E test design: real cluster, fast-forwarded time windows, verify scale events
4. DST test cases: spring-forward (skip hour), fall-back (repeat hour), cross-midnight
5. Test data: fixed dates for reproducibility (e.g., 2025-03-09 for DST spring-forward)
6. Coverage target: 80% for controller logic, 100% for time calculation and state machine

### Verification Steps
- Test plan document listing: test type, scenario, expected outcome, automation status
- Test data fixtures for DST edge cases
- Mock time strategy (interface or build tags)

### Exit Criteria
- controller-reconcile-designer confirms tests cover all reconcile paths
- local-workflow-designer confirms tests are runnable in local cluster
- No critical scenarios missing from test matrix

---

## Gate 8: CI with Unit/Envtest/Smoke

**Owner:** ci-release-designer
**Due:** Day 7 End (2025-10-26 18:00 IST)

### Acceptance Criteria
1. GitHub Actions workflow: lint, unit test, envtest, build image, smoke test
2. Smoke test: kind cluster, deploy controller, create TWS, verify scale in 5 minutes
3. Pre-merge checks: all tests pass, code coverage report, no linter errors
4. Container image: multi-stage build, minimal base (distroless or alpine), tagged with commit SHA
5. Release process: tag triggers build, push to registry, create GitHub release with YAML bundle

### Verification Steps
- Trigger CI on feature branch and verify all steps pass
- Check image size and vulnerabilities (grype or trivy scan)
- Verify smoke test completes in under 10 minutes

### Exit Criteria
- testing-strategy-designer confirms CI runs test plan
- security-rbac-designer confirms image security scan
- No flaky tests or timing-dependent failures

---

## Gate 9: README with 5-Minute Quick Start

**Owner:** docs-dx-designer
**Due:** Day 8 End (2025-10-27 18:00 IST)

### Acceptance Criteria
1. README.md sections: What is Kyklos, Quick Start, Concepts, Examples, Development, Contributing
2. Quick Start: install CRD, deploy controller, create sample TWS, verify scaling (copy-pasteable commands)
3. Concepts: explain time windows, DST handling, grace periods, state machine
4. Examples: basic window, cross-midnight, with grace period, cross-namespace
5. Troubleshooting: common errors and solutions
6. Time from following Quick Start to visible result: under 5 minutes (assumes running cluster)

### Verification Steps
- Fresh user follows README on new cluster
- Verify each command works without modification
- Check for broken links or missing prerequisites

### Exit Criteria
- local-workflow-designer confirms Quick Start matches tested workflow
- community-launch-designer confirms README is launch-ready
- No ambiguous instructions or unexplained jargon

---

## Summary: Definition of Done for v0.1

A gate is considered DONE when:
1. All acceptance criteria are met and verifiable
2. Required Consulted parties (per RACI) have reviewed and approved
3. Artifacts are committed to git with clear filenames
4. Handoff document is updated with completion status
5. kyklos-orchestrator has validated against this quality gate

Failure to meet a gate blocks downstream work. If a gate is at risk of missing deadline, owner must escalate to kyklos-orchestrator at least 12 hours before due time.
