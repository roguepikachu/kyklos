# Day 1 Handoff Specifications

**Date:** 2025-10-20
**Handoff From:** kyklos-orchestrator
**Handoff To:** api-crd-designer, api-validation-defaults-designer, controller-reconcile-designer

---

## Handoff 1: API and CRD Design

**To:** api-crd-designer
**Due:** 2025-10-20 18:00 IST
**Priority:** Critical Path

### Scope
Design the complete TimeWindowScaler Custom Resource Definition including spec fields, status subresource, and OpenAPI validation schema.

### Context from Planning Phase
- **Project Name:** Kyklos (time-based scaling operator)
- **CRD Name:** TimeWindowScaler (API group: kyklos.io/v1alpha1)
- **Core Function:** Scale Deployments/StatefulSets/ReplicaSets based on daily time windows
- **Key Decisions:**
  - Single time window per resource (ADR-0001)
  - Support Deployment, StatefulSet, ReplicaSet via Scale subresource (ADR-0002)
  - IANA timezone with DST handling (ADR-0003)
  - Grace period for scale-down (ADR-0004)

### Required Deliverables
1. **design/api-crd-spec.md**: Complete CRD schema with:
   - Spec fields: timezone, windowStart, windowEnd, activeReplicas, inactiveReplicas, gracePeriod (optional), targetRef
   - Status fields: currentState, observedGeneration, lastScaleTime, nextTransitionTime, conditions
   - OpenAPI validation: timezone non-empty, time format HH:MM 24-hour, replicas >= 0, gracePeriod <= 60m
   - Field tags: json, yaml, validation markers

2. **design/api-field-semantics.md**: Document each field:
   - Purpose and user-facing description
   - Default value (if applicable)
   - Validation rules and error messages
   - Examples of valid and invalid values
   - Precedence rules (e.g., what happens if window overlaps with grace period)

3. **design/api-status-design.md**: Status subresource strategy:
   - Condition types: Ready, Scaling, TimezoneValid, TargetFound
   - observedGeneration pattern for spec changes
   - lastScaleTime and nextTransitionTime calculation responsibility
   - Status update frequency and reconcile coordination

### Acceptance Criteria (Quality Gate 1)
- [ ] All required spec and status fields defined with types
- [ ] OpenAPI validation rules cover all constraints from BRIEF.md
- [ ] Field comments use glossary terms from docs/BRIEF.md consistently
- [ ] Cross-midnight window support is explicitly documented
- [ ] Status conditions are sufficient for user troubleshooting
- [ ] No ambiguous field semantics that could cause reconcile errors

### Inputs Available
- docs/BRIEF.md - Project goals, glossary, success criteria
- docs/DECISIONS.md - ADR-0001 through ADR-0004 with context
- docs/QUALITY-GATES.md - Gate 1 detailed requirements

### Consulted Agents (RACI)
- api-validation-defaults-designer: Review validation rules for completeness
- controller-reconcile-designer: Confirm status fields support reconcile logic

### Risks and Dependencies
**Risk:** Field naming conflicts with Kubernetes built-in types
**Mitigation:** Use "window" prefix for time fields, "target" prefix for workload reference

**Risk:** Ambiguous grace period semantics when overlapping with next window
**Mitigation:** Document explicit precedence: if grace period extends into next active window, cancel grace and activate immediately

**Dependency:** None (can start immediately)

### Handoff Package Includes
- This specification document
- docs/BRIEF.md (glossary reference)
- docs/DECISIONS.md (ADR context)
- docs/QUALITY-GATES.md (acceptance criteria)

### Exit Criteria for Handoff
When design/api-crd-spec.md is complete:
1. Commit to main branch
2. Notify api-validation-defaults-designer and controller-reconcile-designer
3. Update ROADMAP.md Day 1 status to "Complete"
4. Tag kyklos-orchestrator for Gate 1 verification

---

## Handoff 2: Validation and Defaults Design (Pre-Work)

**To:** api-validation-defaults-designer
**Start:** After api-crd-designer commits spec (expected 2025-10-20 14:00 IST)
**Due:** 2025-10-21 18:00 IST
**Priority:** Critical Path

### Scope
Design validation logic, default values, and admission webhook architecture to ensure only valid TimeWindowScaler resources are created and updated.

### Context from Planning Phase
- **Validation Approach:** Combination of OpenAPI schema (CRD-level) and admission webhook (cross-field logic)
- **Default Values:** Minimal - only set defaults that improve user experience without magic behavior
- **Webhook Strategy:** Validating webhook only (no mutation in v0.1)

### Required Deliverables (Day 2)
1. **design/validation-rules.md**: Complete validation logic:
   - Timezone: Must be valid IANA timezone (loadable via Go time.LoadLocation)
   - Time format: HH:MM 24-hour, start and end must differ
   - Replicas: activeReplicas and inactiveReplicas both >= 0
   - Grace period: If specified, must be > 0 and <= 60 minutes
   - TargetRef: apiVersion, kind, name required; namespace optional (default to TWS namespace)
   - Cross-field: windowStart != windowEnd

2. **design/default-values.md**: Default value strategy:
   - gracePeriod: Default to 0 (no grace period)
   - inactiveReplicas: Default to 0 (scale to zero)
   - targetRef.namespace: Default to metadata.namespace of TWS
   - Document rationale for each default

3. **design/admission-webhook-design.md**: Webhook architecture:
   - Webhook configuration: validating, failurePolicy (Fail or Ignore), timeout
   - TLS certificate strategy: cert-manager or self-signed bootstrap
   - Validation logic: when to reject, error message format
   - Update vs Create differences (if any)

### Acceptance Criteria (Quality Gate 2)
- [ ] All validation rules documented with test cases
- [ ] Default values specified for optional fields with rationale
- [ ] Webhook design includes TLS and failure policy with security review
- [ ] Test matrix: valid baseline, each invalid case, boundary conditions
- [ ] Decision on targetRef.namespace default behavior

### Inputs Available
- design/api-crd-spec.md (from api-crd-designer)
- docs/DECISIONS.md (validation patterns)
- Go time.LoadLocation documentation for timezone validation

### Consulted Agents (RACI)
- api-crd-designer: Confirm validation rules match CRD design intent
- security-rbac-designer: Review webhook security model and TLS strategy

### Risks and Dependencies
**Risk:** Invalid timezone strings cause runtime errors in controller
**Mitigation:** Webhook must validate timezone is loadable before accepting

**Risk:** Webhook TLS certificate management complexity
**Mitigation:** Document both cert-manager (production) and self-signed (local dev) approaches

**Dependency:** Requires design/api-crd-spec.md from api-crd-designer

### Day 1 Pre-Work
While waiting for api-crd-designer:
1. Research Go time.LoadLocation error cases for timezone validation
2. List all IANA timezones that should be valid
3. Draft test matrix structure (fill in after API spec is available)
4. Review Kubernetes admission webhook best practices

### Exit Criteria for Handoff
When validation design is complete:
1. Commit all three design docs to main branch
2. Notify security-rbac-designer for webhook security review
3. Notify controller-reconcile-designer (validation gaps affect reconcile error handling)
4. Tag kyklos-orchestrator for Gate 2 verification

---

## Handoff 3: Reconcile Design (Pre-Work)

**To:** controller-reconcile-designer
**Start:** After api-crd-designer commits spec (expected 2025-10-20 14:00 IST)
**Due:** 2025-10-22 18:00 IST (Day 3, parallel with metrics design)
**Priority:** Critical Path

### Scope
Design the controller reconcile loop state machine, requeue timing logic, and error handling to ensure correct scaling behavior across all time scenarios.

### Context from Planning Phase
- **State Machine:** Three states: Inactive, Active, GracePeriod
- **Requeue Pattern:** Calculate next transition time, add jitter, use requeueAfter
- **Idempotency:** Repeated reconciles must not cause scaling thrash
- **Edge Cases:** DST transitions, cross-midnight windows, grace period overlap

### Required Deliverables (Day 2-3)
1. **design/reconcile-state-machine.md**: Complete state transition logic:
   - State definitions: Inactive (outside window, inactiveReplicas), Active (inside window, activeReplicas), GracePeriod (post-window, activeReplicas maintained)
   - Transitions: How and when state changes occur
   - State diagram: Visual representation of transitions
   - Transition triggers: Time-based vs spec change vs target state change

2. **design/reconcile-requeue-logic.md**: Timing calculation:
   - Algorithm to determine current state given: current time, timezone, windowStart, windowEnd
   - Calculate next transition time: when will state next change
   - Handle cross-midnight windows: today vs tomorrow logic
   - DST handling: spring-forward (skip hour), fall-back (repeat hour)
   - Jitter: Add small random delay to avoid thundering herd

3. **design/reconcile-error-handling.md**: Failure scenarios:
   - Target not found: Requeue with backoff, emit event, set condition
   - Scale API failure: Retry with backoff, preserve status
   - Timezone load error: Set condition, do not requeue indefinitely
   - Invalid spec: Should be caught by validation, but handle gracefully

4. **design/reconcile-pseudo-code.md**: Pseudo-code for reconcile function:
   - Fetch TimeWindowScaler resource
   - Load timezone and parse window times
   - Determine current state (Inactive/Active/GracePeriod)
   - Fetch target workload and current replica count
   - If state requires different replicas, scale target
   - Update status: currentState, lastScaleTime, nextTransitionTime, conditions
   - Calculate and return requeue time

### Acceptance Criteria (Quality Gate 3)
- [ ] State machine covers all transitions with clear trigger conditions
- [ ] Requeue timing logic handles DST spring-forward and fall-back
- [ ] Cross-midnight windows correctly determine current state
- [ ] Error handling specifies retry strategy for each failure type
- [ ] Idempotency: reconcile(reconcile(state)) == reconcile(state)
- [ ] Pseudo-code is detailed enough to implement without ambiguity

### Inputs Available
- design/api-crd-spec.md (from api-crd-designer)
- docs/DECISIONS.md (ADR-0003 timezone handling, ADR-0004 grace period)
- docs/BRIEF.md (glossary for state terms)

### Consulted Agents (RACI)
- api-crd-designer: Confirm status fields support reconcile needs
- observability-metrics-designer: Ensure state transitions are observable

### Risks and Dependencies
**Risk:** DST spring-forward causes window to be skipped entirely
**Mitigation:** Document behavior, test with fixed dates (2025-03-09 spring-forward example)

**Risk:** Grace period overlaps with next window start (e.g., grace=30m, window starts 15m after previous ends)
**Mitigation:** Decide precedence: cancel grace and activate immediately, document in ADR

**Risk:** Cross-midnight window logic is complex and error-prone
**Mitigation:** Extensive testing with fixed dates, clear pseudo-code, state machine diagram

**Dependency:** Requires design/api-crd-spec.md from api-crd-designer

### Day 1 Pre-Work
While waiting for api-crd-designer:
1. Research Go time package: time.Now().In(location), time.Parse for HH:MM
2. Sketch state machine diagram on paper
3. List all edge cases: DST spring/fall, cross-midnight, grace overlap
4. Review controller-runtime requeue patterns and best practices

### Exit Criteria for Handoff
When reconcile design is complete:
1. Commit all four design docs to main branch
2. Notify observability-metrics-designer (metrics must cover all states)
3. Notify testing-strategy-designer (test plan must cover all edge cases)
4. Tag kyklos-orchestrator for Gate 3 verification

---

## Communication Protocol for Day 1

### Check-In Schedule
- **10:00 IST:** Morning sync (async via comments or brief standup)
  - Each agent reports: starting on X, blockers, expected completion time
- **14:00 IST:** Midday check
  - api-crd-designer: Status update, ETA for spec completion
  - Other agents: Report on pre-work progress
- **18:00 IST:** End-of-day wrap-up
  - api-crd-designer: Gate 1 verification request or escalation
  - All agents: Summary of what's ready for Day 2

### Collaboration Tools
- **Documents:** All design docs in design/ directory, markdown format
- **Comments:** Use GitHub PR review comments or inline markdown comments
- **Decisions:** Any new decisions go in DECISIONS.md as draft ADRs
- **Questions:** Tag kyklos-orchestrator in commit messages or comments for clarification

### Escalation Triggers
Escalate to kyklos-orchestrator if:
- Any deliverable is at risk of missing 18:00 IST deadline
- Conflicting requirements discovered between agents
- Ambiguity in BRIEF.md or DECISIONS.md blocks progress
- Need scope clarification or priority change

### Definition of Done for Day 1
- [ ] design/api-crd-spec.md committed and reviewed
- [ ] design/api-field-semantics.md committed and reviewed
- [ ] design/api-status-design.md committed and reviewed
- [ ] Quality Gate 1 checklist verified by kyklos-orchestrator
- [ ] Handoffs to Day 2 agents completed (validation and reconcile designers notified)
- [ ] ROADMAP.md Day 1 status updated to "Complete"

---

## Resources and References

### Must-Read Before Starting
1. docs/BRIEF.md - Glossary and success criteria
2. docs/DECISIONS.md - ADR-0001 through ADR-0004
3. docs/QUALITY-GATES.md - Gate 1 detailed requirements
4. docs/RACI.md - Responsibility matrix

### Kubernetes API Reference
- [Custom Resource Definitions](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
- [OpenAPI Validation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#validation)
- [Scale Subresource](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#scale-subresource)
- [Status Subresource](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#status-subresource)

### Go Time Package
- [time.LoadLocation](https://pkg.go.dev/time#LoadLocation)
- [time.Parse](https://pkg.go.dev/time#Parse)
- [IANA Time Zone Database](https://www.iana.org/time-zones)

### Controller Runtime Patterns
- [Reconcile Return Values](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/reconcile)
- [Requeue Strategies](https://sdk.operatorframework.io/docs/building-operators/golang/references/reconcile-response/)

---

## Success Criteria for Day 1 Handoffs

Day 1 handoffs are successful when:
1. api-crd-designer has clear, unambiguous scope and acceptance criteria
2. All required inputs (BRIEF, DECISIONS, QUALITY-GATES) are accessible
3. Consulted agents (validation, reconcile designers) know when to engage
4. Risks are identified with mitigations or escalation paths
5. Communication protocol is clear (check-in times, escalation triggers)
6. Definition of done is measurable and verifiable by kyklos-orchestrator

If any handoff is unclear or incomplete, agents must request clarification from kyklos-orchestrator before starting work.
