# Day 9 Review: Required ADR Updates

**Review Date:** 2025-10-29
**Reviewer:** kyklos-tws-reviewer
**Purpose:** New Architecture Decision Records needed based on review findings
**Status:** 3 New ADRs Required, 2 ADR Updates Needed

---

## New ADRs Required

### ADR-0005: Holiday Support Scope for v0.1

**Status:** DRAFT - CRITICAL DECISION NEEDED
**Deciders:** kyklos-orchestrator, api-crd-designer, controller-reconcile-designer
**Date:** 2025-10-29
**Deadline:** 2025-10-29 18:00 IST

#### Context

BRIEF.md lists "Calendar integration or holiday awareness" as a non-goal for v0.1 (line 17). However, comprehensive holiday support has been designed and documented across multiple files:
- CRD-SPEC.md: Full spec with three modes (lines 66-80)
- RECONCILE.md: Holiday checking logic (Step 3)
- CONCEPTS.md: Complete user documentation (lines 226-307)
- Examples: tws-holidays-closed.yaml exists

This creates scope ambiguity. Implementation team doesn't know whether to build holiday support or not.

#### Decision Options

**Option A: Include Holidays in v0.1 (RECOMMENDED)**

**Rationale:**
- Already fully designed (90% complete)
- ConfigMap-based, simple implementation
- No external dependencies
- High user value (common use case)
- Test strategy exists

**Consequences:**
- Add 3-4 days to implementation (ConfigMap watcher, date parsing)
- Need holiday test scenarios
- Keep all existing documentation
- Update BRIEF.md non-goals to remove holiday mention

**Implementation Impact:**
- Add ConfigMap watch to controller
- Implement Step 3 of reconcile logic
- Create holiday envtest scenarios
- Validate tws-holidays-closed.yaml example

---

**Option B: Defer Holidays to v0.2**

**Rationale:**
- Keep v0.1 minimal
- Focus on core time window logic
- Holidays can be added later without API breaking changes
- Reduces implementation risk

**Consequences:**
- Remove holiday sections from 5+ documents
- Move example to examples/future/
- Document as "coming in v0.2"
- 10+ documentation updates required

**Implementation Impact:**
- Remove holiday logic from reconcile design
- No ConfigMap watch needed
- Simpler test matrix
- Clear v0.1/v0.2 boundary

---

**Decision:** [TO BE DECIDED BY TEAM]

**Rationale:** [TO BE FILLED]

**Mitigations:** [TO BE FILLED]

**Follow-up Actions:**
- [ ] Update BRIEF.md non-goals section
- [ ] Update or remove holiday sections per decision
- [ ] Update test strategy for holiday scenarios
- [ ] Update or move example files
- [ ] Document decision in ROADMAP.md

---

### ADR-0006: Validation Strategy for v0.1

**Status:** DRAFT - DECISION NEEDED
**Deciders:** api-validation-defaults-designer, k8s-security-rbac-planner
**Date:** 2025-10-29
**Deadline:** 2025-10-30 18:00 IST

#### Context

CRD-SPEC.md mentions "enforced by admission webhook" (line 26) but no webhook design exists. Day 2 deliverables included validation design, but it's unclear whether webhook is v0.1 or not.

Validation webhooks add complexity:
- TLS certificate management
- Webhook deployment and lifecycle
- Failure modes and policies
- Additional RBAC permissions

CRD validation (OpenAPI schema) can handle most validations:
- Field types and formats
- Required fields
- Enum values
- Regex patterns
- Min/max constraints

#### Decision Options

**Option A: CRD Validation Only (RECOMMENDED)**

**Validations via CRD OpenAPI:**
- timezone: string, required
- windows[].start/end: regex pattern `^([0-1][0-9]|2[0-3]):[0-5][0-9]$`
- windows[].days: enum [Mon, Tue, Wed, Thu, Fri, Sat, Sun], minItems: 1
- windows[].replicas: int32, minimum: 0
- defaultReplicas: int32, minimum: 0
- gracePeriodSeconds: int32, minimum: 0, maximum: 3600

**Cannot Validate via CRD:**
- Timezone is valid IANA identifier (checked at runtime)
- window.start != window.end (checked at runtime)
- ConfigMap referenced in holidays exists (checked at runtime)

**Runtime Validation:**
- Controller checks timezone validity in Step 2
- Controller checks window.start != window.end in Step 1
- Sets Degraded condition if validation fails

**Pros:**
- Simpler implementation (no webhook server)
- No TLS certificate management
- Faster feedback (API server validates immediately)
- No webhook failure modes

**Cons:**
- Can't validate timezone is valid IANA name until runtime
- Can't validate start != end until runtime
- Less defensive (some invalid resources can be created)

---

**Option B: Admission Webhook for Complex Validation**

**Additional Validations via Webhook:**
- Timezone is valid IANA identifier (time.LoadLocation check)
- window.start != window.end
- holiday.sourceRef ConfigMap exists (if specified)
- Overlapping windows are intentional (warning, not rejection)

**Pros:**
- Fail fast at creation time
- Better user experience (errors at kubectl apply)
- Can provide detailed validation messages
- Can validate cross-field constraints

**Cons:**
- Additional operational complexity
- TLS certificate lifecycle management
- Webhook availability affects cluster operations
- Requires webhook deployment before CRD installation
- Additional 5-7 days implementation time

---

**Decision:** [TO BE DECIDED]

**Rationale:** [TO BE FILLED]

**Implementation Plan:**

If Option A:
- [ ] Update CRD-SPEC.md to remove webhook references
- [ ] Add runtime validation section to RECONCILE.md
- [ ] Document validation error messages
- [ ] Update RBAC (no webhook permissions needed)

If Option B:
- [ ] Create design/validation-webhook.md
- [ ] Design TLS certificate strategy
- [ ] Design webhook deployment model
- [ ] Update RBAC for webhook permissions
- [ ] Add webhook to CI/CD pipeline

---

### ADR-0007: Field Naming Convention - replicas vs activeReplicas

**Status:** DRAFT - INFORMATIONAL (Decision Already Made, Needs Documentation)
**Deciders:** api-crd-designer
**Date:** 2025-10-29

#### Context

Day 0 glossary (BRIEF.md) defined:
- `activeReplicas`: Desired replica count during active window
- `inactiveReplicas`: Desired replica count during inactive window

Day 1 API design (CRD-SPEC.md) uses:
- `windows[].replicas`: Replica count when this window is active
- `defaultReplicas`: Replica count when no windows match

Runtime computed value:
- `effectiveReplicas` (in status): The replica count right now

This creates terminology mismatch between glossary and actual API.

#### Decision

**Use API field names, not glossary terms.**

**Spec fields:**
- `windows[].replicas` - What to scale to when this window matches
- `defaultReplicas` - What to scale to when no windows match

**Status fields:**
- `effectiveReplicas` - Computed desired replicas right now
- `targetObservedReplicas` - Actual replica count of target deployment

**No fields named:** activeReplicas, inactiveReplicas

#### Rationale

- API field names are more precise and flexible
- "active/inactive" implies binary state, but we have multiple windows
- `defaultReplicas` clearly communicates "default when nothing else matches"
- `effectiveReplicas` clearly means "what we computed for right now"
- Matches Kubernetes conventions (defaultMode, effectiveNodeSelector, etc.)

#### Consequences

**Positive:**
- API is self-documenting
- No confusion between "activeReplicas" the concept vs the field
- Flexible enough for future multi-window support

**Negative:**
- Glossary from Day 0 is now outdated
- Documentation using old terms must be updated

#### Implementation

- [ ] Update BRIEF.md glossary (remove activeReplicas/inactiveReplicas, add effectiveReplicas)
- [ ] Global find-replace in all docs: activeReplicas → windows[].replicas (context-dependent)
- [ ] Global find-replace in all docs: inactiveReplicas → defaultReplicas
- [ ] Add note explaining why effectiveReplicas is computed, not in spec

**Migration:** No API migration needed (v0.1 is first release).

---

## ADR Updates Required

### ADR-0002 Update: Cross-Namespace Validation Constraint

**Original ADR:** ADR-0002: Supported Target Kind and Namespace Model
**Update Required:** Clarify validation behavior
**Date:** 2025-10-29

#### Issue

CRD-SPEC.md line 28 states:
> "If namespace is specified, it must equal the TimeWindowScaler's namespace"

This contradicts ADR-0002 which explicitly supports cross-namespace with RBAC.

#### Required Update

Add section to ADR-0002:

**Validation Strategy:**
- CRD validation does NOT enforce same-namespace constraint
- Controller validates RBAC at runtime
- Cross-namespace requires ClusterRole (documented in RBAC design)
- Same-namespace uses namespaced Role (simpler RBAC)

**Implementation:**
- If targetRef.namespace is empty: use TimeWindowScaler namespace
- If targetRef.namespace differs: verify RBAC, set Degraded if insufficient
- No API validation preventing cross-namespace

#### Consequences

- Users can create cross-namespace references
- Controller detects insufficient RBAC and reports via Degraded condition
- Documentation must clarify RBAC requirements for cross-namespace

#### Action Items

- [ ] Update CRD-SPEC.md validation rules (remove same-namespace constraint)
- [ ] Add RBAC validation section to RECONCILE.md Step 6
- [ ] Document cross-namespace RBAC requirements
- [ ] Update ADR-0002 with validation strategy

---

### ADR-0004 Update: Grace Period Field Name and Status

**Original ADR:** ADR-0004: Grace Period Semantics
**Update Required:** Clarify field naming and status field
**Date:** 2025-10-29

#### Issue

RECONCILE.md uses `status.gracePeriodExpiry` field (line 73) but this field is not in CRD-SPEC.md status definition.

Also, RECONCILE.md sometimes uses `gracePeriod` and sometimes `gracePeriodSeconds`.

#### Required Update

Add section to ADR-0004:

**Field Names:**
- **Spec field:** `gracePeriodSeconds` (int32) - Duration in seconds
- **Status field:** `gracePeriodExpiry` (string) - RFC3339 timestamp when grace expires

**Semantics:**
- `gracePeriodSeconds` is the configured duration (e.g., 300 for 5 minutes)
- `gracePeriodExpiry` is set to `now + gracePeriodSeconds` when entering grace period
- `gracePeriodExpiry` is cleared when grace expires or is cancelled
- Controller uses `gracePeriodExpiry` to determine if still in grace

**Why separate fields:**
- Configured duration (spec) vs active expiry time (status) are different concepts
- Expiry timestamp allows controller restarts to maintain grace period correctly
- Clear when grace is active (field is non-empty) vs not active (field is empty)

#### Consequences

- CRD-SPEC.md must add gracePeriodExpiry to status fields
- All docs must use gracePeriodSeconds (not gracePeriod) for spec field
- Controller must set/clear gracePeriodExpiry correctly

#### Action Items

- [ ] Add gracePeriodExpiry to CRD-SPEC.md status section
- [ ] Update ADR-0004 with field naming clarification
- [ ] Global replace gracePeriod → gracePeriodSeconds in RECONCILE.md
- [ ] Document grace period lifecycle in STATUS-CONDITIONS.md

---

## ADR Template for Future

All future ADRs should follow this format:

```markdown
## ADR-XXXX: Title

**Date:** YYYY-MM-DD
**Status:** Proposed | Accepted | Deprecated | Superseded
**Deciders:** agent-id(s)
**Deadline:** YYYY-MM-DD HH:MM TZ (if decision is time-sensitive)

### Context
What is the issue we're trying to solve?

### Decision Options
List all considered options with pros/cons

### Decision
What did we decide?

### Rationale
Why this option?

### Consequences
**Positive:**
**Negative:**
**Mitigations:**

### Implementation Plan
Specific action items with owners

### Alternatives Considered
What else was on the table?

### Open Questions
What remains to be decided?
```

---

## ADR Tracking

| ADR | Title | Status | Decider | Deadline | Blocker? |
|-----|-------|--------|---------|----------|----------|
| ADR-0001 | Name and Scope | Accepted | kyklos-orchestrator | Done | No |
| ADR-0002 | Target Kinds | Update Needed | api-crd-designer | Oct 29 | No |
| ADR-0003 | Timezone and DST | Accepted | kyklos-orchestrator | Done | No |
| ADR-0004 | Grace Period | Update Needed | api-crd-designer | Oct 29 | No |
| ADR-0005 | Holiday Scope | Draft | kyklos-orchestrator | Oct 29 18:00 | **YES** |
| ADR-0006 | Validation Strategy | Draft | api-validation-defaults-designer | Oct 30 18:00 | No |
| ADR-0007 | Field Naming | Draft | api-crd-designer | Oct 29 | No |

**Critical Path:** ADR-0005 must be decided before any Day 10 work can proceed.

---

## Implementation Order

1. **Oct 29 18:00** - ADR-0005 (Holiday Scope) - Blocks all holiday-related work
2. **Oct 29 EOD** - ADR-0007 (Field Naming) - Enables terminology updates
3. **Oct 29 EOD** - ADR-0002 Update - Fixes cross-namespace contradiction
4. **Oct 29 EOD** - ADR-0004 Update - Fixes grace period field confusion
5. **Oct 30 18:00** - ADR-0006 (Validation) - Determines validation implementation

---

**Prepared by:** kyklos-tws-reviewer
**Date:** 2025-10-29 17:30 IST
**Next Review:** After ADR-0005 decision (Oct 29 18:00)
