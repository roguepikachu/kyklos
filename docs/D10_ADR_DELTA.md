# Day 10 ADR Delta: New Architecture Decision Records

**Date:** 2025-10-30
**Coordinator:** kyklos-orchestrator
**Purpose:** Complete text for new ADRs to be added to DECISIONS.md
**Status:** Ready for incorporation

---

## Instructions

Append the following ADRs to `/Users/aykumar/personal/kyklos/docs/DECISIONS.md` in the order specified. Also apply the updates to existing ADRs as shown.

---

## NEW ADR-0005: Holiday Support Scope for v0.1

**Date:** 2025-10-30
**Status:** Accepted
**Deciders:** kyklos-orchestrator, api-crd-designer, controller-reconcile-designer
**Decision Deadline:** 2025-10-30 10:30 IST

### Context

During Day 0 planning, BRIEF.md listed "Calendar integration or holiday awareness" as a non-goal for v0.1 (line 17). However, during Days 1-6 design phase, comprehensive holiday support was designed and documented across multiple files:

- **CRD-SPEC.md** (lines 66-80): Full holiday spec with three modes (ignore, treat-as-closed, treat-as-open)
- **RECONCILE.md** (Step 3): Holiday checking logic with ConfigMap lookup
- **CONCEPTS.md** (lines 226-307): Complete user documentation with examples
- **examples/tws-holidays-closed.yaml**: Working example demonstrating holiday mode

This created critical scope ambiguity. The Day 9 review identified this as **RISK-NEW-001 (Critical)**, blocking scope lock and implementation. The implementation team cannot proceed without knowing whether to build holiday support or remove it from documentation.

**Scope Question:** Should ConfigMap-based holiday support be included in v0.1?

### Decision Options

#### Option A: Include Basic Holiday Support in v0.1 (SELECTED)

**Rationale:**
1. **Already 90% designed:** Complete CRD spec, reconcile logic, user docs, examples
2. **Simple implementation:** ConfigMap-based, no external dependencies
3. **High user value:** Common use case for workload scheduling (closed on holidays)
4. **No breaking changes:** Holiday field is optional, defaults to ignore mode
5. **Clear v0.1/v0.2 boundary:** v0.1 has ConfigMap source, v0.2 adds external calendar sync

**Scope for v0.1:**
- ConfigMap-based holiday list (key = ISO date `yyyy-mm-dd`)
- Three modes: ignore (default), treat-as-closed, treat-as-open
- Same-namespace ConfigMap reference only (cross-namespace in v0.2)
- Static date list (no recurring rules)

**Excluded from v0.1 (deferred to v0.2):**
- External calendar APIs (Google Calendar, Outlook, etc.)
- Recurring holiday rules (e.g., "first Monday of September")
- Cross-namespace ConfigMap references
- Holiday caching/synchronization across multiple TWS resources

#### Option B: Defer All Holiday Support to v0.2 (NOT SELECTED)

**Rationale:**
1. Keep v0.1 minimal and focused on core time window logic
2. Reduce implementation risk and testing surface area
3. Holidays can be added in v0.2 without breaking API changes
4. Focus v0.1 on perfecting DST and cross-midnight logic

**Implementation Impact:**
- Remove holiday sections from 10+ documents
- Move example to examples/future/
- Remove Step 3 from reconcile logic
- Simplifies test matrix

### Final Decision

**Include basic holiday support in v0.1 (Option A).**

### Rationale

1. **Design completeness:** Holiday support is already comprehensively designed, tested, and documented. Removing it would require significant rework and discard valuable design effort.

2. **User value:** Holiday awareness is a common requirement for workload scheduling. Including it in v0.1 provides immediate value to users who need to scale down services on holidays.

3. **Implementation simplicity:** ConfigMap-based holiday lists are simple to implement (watch ConfigMap, parse dates, check membership). This adds minimal complexity compared to the core time window logic.

4. **No scope creep:** The v0.1 implementation is intentionally limited to ConfigMap sources with static date lists. Advanced features (external calendars, recurring rules) are clearly deferred to v0.2.

5. **Optional feature:** The holiday field is optional. Users who don't need holiday support can ignore it entirely. Default mode is `ignore`, which means no behavioral change if holidays aren't configured.

6. **Testing benefit:** Holiday logic exercises the same precedence and window matching code paths as normal windows, providing additional test coverage for the core reconcile loop.

### Consequences

**Positive:**
- v0.1 delivers complete time-based scaling solution including holidays
- All existing design documentation remains valid and accurate
- Users can deploy holiday-aware schedules immediately
- Implementation team has clear, complete specifications
- Test fixtures and scenarios already designed

**Negative:**
- Adds 3-4 days to implementation timeline (ConfigMap watcher, date parsing)
- Increases test matrix size (holiday mode scenarios)
- Adds operational complexity (users must create/maintain ConfigMap)

**Mitigations:**
- ConfigMap is optional, not required
- Default mode (`ignore`) has zero performance impact
- Clear documentation distinguishes v0.1 ConfigMap support from v0.2 external calendars
- Test strategy already includes holiday scenarios

### Implementation Plan

1. **Update BRIEF.md** (kyklos-orchestrator, Oct 30 12:00 IST)
   - Change line 17 from "Calendar integration or holiday awareness" to "Advanced calendar features (recurring patterns, external calendar sync beyond ConfigMap)"
   - This clarifies that basic ConfigMap-based holidays ARE in v0.1, but advanced features are v0.2

2. **Keep all existing holiday documentation** (no removal needed)
   - CRD-SPEC.md holiday section: clarify `ignore` is default mode
   - RECONCILE.md Step 3: keep as-is
   - CONCEPTS.md holiday section: add note "v0.1 supports ConfigMap sources"
   - examples/tws-holidays-closed.yaml: validate and keep

3. **Create holiday test scenarios** (testing-strategy-designer, Oct 31)
   - Add to ENVTEST-PLAN.md: treat-as-closed mode test
   - Add to ENVTEST-PLAN.md: treat-as-open mode test
   - Add to ENVTEST-PLAN.md: ConfigMap missing/invalid test

4. **Implement during implementation phase** (Nov 3-7)
   - ConfigMap watcher in controller
   - Date parsing (Go time.Parse with ISO 8601)
   - Holiday mode logic in reconcile Step 3
   - Events for holiday overrides

### Alternatives Considered

**Option C: Holidays in v0.1 but webhook-validated ConfigMap:**
- Rejected: Adds webhook complexity, ConfigMap validation best done at runtime

**Option D: Holidays in v0.2 but document as "coming soon":**
- Rejected: Confusing to have full documentation for unavailable feature

### Open Questions

None. Decision is final and ready for implementation.

---

## NEW ADR-0006: Validation Strategy for v0.1

**Date:** 2025-10-30
**Status:** Accepted
**Deciders:** api-validation-defaults-designer, k8s-security-rbac-planner
**Decision Deadline:** 2025-10-30 16:00 IST

### Context

The CRD-SPEC.md references "enforced by admission webhook" for validation (line 26), but no webhook design exists. Validation webhooks add significant operational complexity:

**Webhook Complexity:**
- TLS certificate lifecycle management (creation, rotation, expiration)
- Webhook deployment and availability (webhook unavailable blocks CR creation)
- Failure policies (fail-open vs fail-closed security trade-offs)
- Additional RBAC permissions for webhook ServiceAccount
- Webhook registration and lifecycle management
- 5-7 additional days of implementation work

**CRD Validation Capabilities:**
OpenAPI schema validation in CRD can handle:
- Field types, formats, and required constraints
- Enum values and regex patterns
- Numeric min/max constraints
- Array minItems/maxItems

**Runtime Validation Capabilities:**
Controller can validate at reconcile time:
- IANA timezone validity (requires time.LoadLocation)
- Cross-field constraints (start != end)
- Referenced resources exist (ConfigMap, Deployment)

### Decision Options

#### Option A: CRD Validation Only (SELECTED)

**Validation Strategy:**

**Via CRD OpenAPI Schema:**
- `timezone`: string, required
- `windows[].start`: regex pattern `^([0-1][0-9]|2[0-3]):[0-5][0-9]$`
- `windows[].end`: regex pattern `^([0-1][0-9]|2[0-3]):[0-5][0-9]$`
- `windows[].days`: enum [Mon, Tue, Wed, Thu, Fri, Sat, Sun], minItems: 1
- `windows[].replicas`: int32, minimum: 0
- `defaultReplicas`: int32, minimum: 0
- `gracePeriodSeconds`: int32, minimum: 0, maximum: 3600
- `holidays.mode`: enum [ignore, treat-as-closed, treat-as-open]

**Via Runtime Controller Validation:**
- Timezone is valid IANA identifier (validated in Step 1)
- window.start != window.end (validated in Step 1)
- holiday.sourceRef ConfigMap exists (validated in Step 3, degrades gracefully if missing)
- targetRef points to valid Deployment (validated in Step 6)

**Error Reporting:**
- CRD validation errors: immediate API server rejection with clear error message
- Runtime validation errors: Degraded condition with reason (InvalidTimezone, InvalidSchedule, etc.)

#### Option B: Admission Webhook (NOT SELECTED)

**Additional Validations via Webhook:**
- Timezone validity check at creation time
- window.start != window.end enforcement at creation time
- holiday.sourceRef ConfigMap existence check
- Overlapping window warnings (not rejection)

**Why Not Selected:**
- Operational complexity too high for v0.1
- CRD validation + runtime validation covers 95% of error cases
- Webhook can be added in v0.2 without breaking changes if needed

### Final Decision

**Use CRD validation only for v0.1 (Option A).**

Validation webhook is explicitly deferred to v0.2 if user feedback indicates it's needed.

### Rationale

1. **Simplicity:** CRD validation + runtime validation provides adequate protection without operational complexity of webhooks.

2. **User Experience:** Most invalid configurations are caught by CRD validation at `kubectl apply` time. Timezone errors surface quickly via Degraded condition.

3. **Fail-safe:** Runtime validation allows controller to degrade gracefully (use defaultReplicas) rather than blocking resource creation entirely.

4. **Implementation speed:** Eliminates 5-7 days of webhook implementation work, allowing focus on core reconcile logic.

5. **Security:** CRD validation prevents malformed data. Runtime validation prevents logic errors. This is sufficient for v0.1 alpha release.

6. **Future flexibility:** Webhook can be added in v0.2 if needed without API changes.

### Consequences

**Positive:**
- Simpler deployment (no webhook certificates or registration)
- Faster implementation timeline
- No webhook availability concerns
- Clearer error messages from CRD validation
- Graceful degradation for runtime errors

**Negative:**
- Invalid timezone strings can be created (detected at runtime)
- window.start == window.end can be created (detected at runtime)
- ConfigMap references not validated at creation time
- Users see runtime errors in status conditions instead of creation-time rejections

**Mitigations:**
- Clear documentation of validation rules in CRD-SPEC.md and user docs
- Comprehensive error messages in Degraded condition
- Quick feedback loop (reconcile happens within seconds of CR creation)
- Examples in repository are pre-validated
- TROUBLESHOOTING.md covers common validation errors

### Implementation Plan

1. **Update CRD-SPEC.md** (api-crd-designer, Oct 30 13:00 IST)
   - Change line 26 from "enforced by admission webhook" to "enforced by CRD enum validation"
   - Add section documenting runtime validation behavior

2. **Enhance RECONCILE.md Step 1** (controller-reconcile-designer, Oct 30 13:00 IST)
   - Add detailed validation logic with error messages
   - Document Degraded condition reasons for each validation failure

3. **Update TROUBLESHOOTING.md** (docs-dx-designer, Oct 31)
   - Add section on common validation errors
   - Explain difference between CRD validation (immediate) and runtime validation (via status)

4. **No webhook implementation needed** for v0.1

### Alternatives Considered

**Option C: Validating Admission Webhook with cert-manager:**
- Rejected: Adds cert-manager as dependency, increases complexity

**Option D: Static webhook certificates (manual rotation):**
- Rejected: Operational burden for users, certificate expiration risk

### Open Questions

None. Decision is final. Webhook can be reconsidered for v0.2 based on user feedback.

---

## NEW ADR-0007: Field Naming Convention Clarification

**Date:** 2025-10-30
**Status:** Accepted
**Deciders:** api-crd-designer
**Category:** Informational (Decision Already Made on Day 1, Needs Documentation)

### Context

The Day 0 BRIEF.md glossary defined:
- `activeReplicas`: Desired replica count during active window
- `inactiveReplicas`: Desired replica count during inactive window (often 0)

The Day 1 CRD-SPEC.md API design uses:
- `windows[].replicas`: Replica count when this window is active
- `defaultReplicas`: Replica count when no windows match

The runtime status includes:
- `effectiveReplicas`: The replica count computed for right now

This created terminology mismatch between the Day 0 glossary (planning names) and the Day 1 API (actual field names). All documentation from Day 2 onward consistently uses the API field names, but the BRIEF.md glossary was never updated.

The Day 9 review identified this as **RISK-NEW-002 (High Priority)**, causing confusion for anyone reading BRIEF.md and then looking at examples or CRD spec.

### Decision

**Use API field names throughout all documentation. Day 0 glossary terms `activeReplicas` and `inactiveReplicas` are obsolete and should be replaced.**

**Official Field Names:**

**In spec (user-configured):**
- `windows[].replicas` - What to scale to when this specific window matches
- `defaultReplicas` - What to scale to when no windows match

**In status (controller-computed):**
- `effectiveReplicas` - Desired replica count right now (computed from windows/default)
- `targetObservedReplicas` - Actual replica count of target deployment right now

**Never use:** activeReplicas, inactiveReplicas (these terms don't exist in API)

### Rationale

1. **Precision:** "active/inactive" implies binary state, but we support multiple overlapping windows. "windows[].replicas" is clear: the replica count for this specific window.

2. **Clarity:** "defaultReplicas" explicitly means "use this when nothing else matches." "inactiveReplicas" implies workload is inactive, but defaultReplicas=2 for availability is common.

3. **Distinction:** Three separate concepts need three separate terms:
   - Configuration (windows[].replicas, defaultReplicas)
   - Computation (effectiveReplicas)
   - Observation (targetObservedReplicas)

4. **Kubernetes Conventions:** Matches patterns like defaultMode (ConfigMap), effectiveNodeSelector (Pod), etc.

5. **Flexibility:** Terminology scales to future features (multiple windows, window priorities) without confusion.

### Consequences

**Positive:**
- API is self-documenting (field names explain their purpose)
- No confusion between "activeReplicas the concept" vs "activeReplicas the field"
- Clear separation of configuration vs computation vs observation
- Documentation uses consistent terminology across all files

**Negative:**
- Day 0 glossary was incorrect and must be updated
- Any external references to "activeReplicas" must be corrected

**Mitigation:**
- Update BRIEF.md glossary (remove activeReplicas/inactiveReplicas, add all three correct terms)
- Global documentation search to ensure no remaining references to old terms

### Implementation Plan

1. **Update BRIEF.md glossary** (kyklos-orchestrator, Oct 30 12:00 IST)
   - Remove activeReplicas and inactiveReplicas entries
   - Add windows[].replicas, defaultReplicas, effectiveReplicas with clear definitions
   - Add note distinguishing configured vs computed vs observed

2. **Verify terminology cleanup** (kyklos-orchestrator, Oct 30 15:00 IST)
   ```bash
   git grep "activeReplicas" docs/ | grep -v "D9_\|D10_\|windows"
   # Should return 0 results
   ```

3. **Update all user-facing documentation** (docs-dx-designer, Oct 31)
   - CONCEPTS.md: Add terminology clarification section
   - OPERATIONS.md: Use correct field names in all examples
   - FAQ.md: Update any Q&A using old terms
   - GLOSSARY.md: Full alignment with BRIEF.md

### Migration

No API migration needed (v0.1 is first release). This is a documentation-only clarification.

### Alternatives Considered

**Option B: Keep activeReplicas in glossary as "conceptual term":**
- Rejected: Confuses users who look for this field in CRD spec

**Option C: Add activeReplicas as alias in API:**
- Rejected: Multiple names for same concept creates confusion

### Open Questions

None. Naming is finalized and consistent across all Day 1-8 design documents.

---

## UPDATE TO ADR-0002: Cross-Namespace Validation Strategy

**Original ADR:** ADR-0002: Supported Target Kind and Namespace Model
**Update Date:** 2025-10-30
**Reason:** Clarify validation behavior for cross-namespace references

### New Section to Add to ADR-0002

After the "Consequences" section, add:

### Validation Strategy for Cross-Namespace References

**Issue Identified:** CRD-SPEC.md line 28 initially stated "If namespace is specified, it must equal the TimeWindowScaler's namespace." This contradicts the decision to support cross-namespace with explicit RBAC.

**Clarification:**

**CRD Validation Level:**
- `targetRef.namespace` field has NO CRD-level validation constraint
- Users CAN create TimeWindowScaler with cross-namespace reference
- Users CAN create TimeWindowScaler with same-namespace reference
- API server accepts both configurations

**Runtime Controller Validation:**
- If `targetRef.namespace` is empty: use TimeWindowScaler's namespace (same-namespace)
- If `targetRef.namespace` is set: use specified namespace (cross-namespace)
- Controller checks RBAC permissions when accessing target
- If insufficient RBAC: set Degraded condition with reason=InsufficientPermissions

**RBAC Requirements:**
- **Same-namespace:** Controller needs namespaced Role in TimeWindowScaler namespace
  ```yaml
  apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    namespace: default
  rules:
  - apiGroups: ["apps"]
    resources: ["deployments", "deployments/scale"]
    verbs: ["get", "list", "watch", "update", "patch"]
  ```

- **Cross-namespace:** Controller needs ClusterRole (cluster-wide access)
  ```yaml
  apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
    name: kyklos-controller
  rules:
  - apiGroups: ["apps"]
    resources: ["deployments", "deployments/scale"]
    verbs: ["get", "list", "watch", "update", "patch"]
  ```

**Documentation Requirements:**
- RBAC-MATRIX.md must document both models
- TROUBLESHOOTING.md must explain "InsufficientPermissions" error
- Examples directory should include both same-namespace and cross-namespace examples

**Security Considerations:**
- ClusterRole grants access to all namespaces
- Production deployments should use namespace-scoped Role if possible
- Cross-namespace model intended for centralized TWS management use case

### Implementation Impact

- CRD-SPEC.md line 28: Remove same-namespace validation constraint
- RECONCILE.md Step 6: Add RBAC error handling
- RBAC-MATRIX.md: Document both permission models

---

## UPDATE TO ADR-0004: Grace Period Field Naming

**Original ADR:** ADR-0004: Grace Period Semantics
**Update Date:** 2025-10-30
**Reason:** Clarify field naming and add status field specification

### New Section to Add to ADR-0004

After the "Consequences" section, add:

### Field Naming Specification

**Issue Identified:** RECONCILE.md used inconsistent naming (`gracePeriod` vs `gracePeriodSeconds`) and referenced `status.gracePeriodExpiry` field not present in CRD-SPEC.md.

**Clarification:**

**Spec Field (User-Configured):**
```yaml
spec:
  gracePeriodSeconds: 300  # int32, duration in seconds
```
- Field name: `gracePeriodSeconds` (NOT `gracePeriod`)
- Type: int32
- Unit: seconds
- Validation: minimum 0, maximum 3600 (1 hour)
- Default: 0 (no grace period)

**Status Field (Controller-Managed):**
```yaml
status:
  gracePeriodExpiry: "2025-10-30T14:35:00Z"  # RFC3339 timestamp
```
- Field name: `gracePeriodExpiry` (NOT `gracePeriodEnd` or `gracePeriodTimeout`)
- Type: string (RFC3339 format)
- Semantics: Absolute timestamp when grace period expires
- Empty when not in grace period

**Why Two Separate Fields:**

1. **Different Concepts:**
   - `gracePeriodSeconds` = configured duration (e.g., 300 seconds = 5 minutes)
   - `gracePeriodExpiry` = computed expiry timestamp (e.g., "2025-10-30T14:35:00Z")

2. **Controller Restart Safety:**
   - Expiry timestamp persists across controller restarts
   - Controller can check `if now < gracePeriodExpiry` without recalculating

3. **Observability:**
   - Users can see exactly when grace period expires
   - Metrics can track grace period duration accurately

**Grace Period Lifecycle:**

```
State: Active Window (replicas=5)
  status.effectiveReplicas=5
  status.gracePeriodExpiry=""

Event: Window ends (should scale to 0)
  Controller enters grace period
  status.effectiveReplicas=5 (still)
  status.gracePeriodExpiry="2025-10-30T14:35:00Z"
  Status condition: Ready=True, reason=GracePeriodWaiting

Time passes...

Event: now >= gracePeriodExpiry
  Controller applies scale-down
  status.effectiveReplicas=0
  status.gracePeriodExpiry="" (cleared)
  Write replicas=0 to target deployment
```

**Grace Period Cancellation:**

If window becomes active again during grace period:
```
status.effectiveReplicas=5 (window active again)
status.gracePeriodExpiry="" (cleared, grace cancelled)
Status condition: Ready=True, reason=WindowActive
```

### Implementation Impact

- CRD-SPEC.md: Add `status.gracePeriodExpiry` field specification
- RECONCILE.md: Fix all references to use `spec.gracePeriodSeconds`
- STATUS-CONDITIONS.md: Add grace period condition transitions
- CONCEPTS.md: Explain grace period lifecycle with timestamps

---

## Summary of ADR Changes

**New ADRs:**
1. ADR-0005: Holiday Support in v0.1 (holidays included via ConfigMap)
2. ADR-0006: Validation Strategy (CRD validation only, no webhook)
3. ADR-0007: Field Naming Clarification (windows[].replicas, not activeReplicas)

**Updated ADRs:**
1. ADR-0002: Add cross-namespace validation strategy section
2. ADR-0004: Add field naming specification section

**Total Changes:** 5 ADR modifications

---

## Application Instructions

1. Open `/Users/aykumar/personal/kyklos/docs/DECISIONS.md`
2. Append ADR-0005, ADR-0006, ADR-0007 to end of file
3. Navigate to ADR-0002, add new section after "Consequences"
4. Navigate to ADR-0004, add new section after "Consequences"
5. Save file
6. Commit with message: "docs: add ADR-0005, ADR-0006, ADR-0007 and update ADR-0002, ADR-0004"

---

**Prepared by:** kyklos-orchestrator
**Date:** 2025-10-30 10:00 IST
**Status:** Ready for incorporation
**Next Step:** Apply to DECISIONS.md as Phase 1 of D10_MERGE_PLAN.md
