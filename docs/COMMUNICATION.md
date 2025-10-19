# Communication Protocols

**Project:** Kyklos
**Last Updated:** 2025-10-19
**Owner:** kyklos-orchestrator

This document defines how agents communicate, resolve conflicts, track decisions, and manage artifact versioning throughout the project lifecycle.

---

## Communication Channels

### Primary: Git Repository
- **Design Documents:** `/design/*.md` (markdown files)
- **Planning Documents:** `/docs/BRIEF.md`, `/docs/DECISIONS.md`, etc.
- **Comments:** Inline markdown comments or commit messages
- **Decisions:** Architecture Decision Records in `/docs/DECISIONS.md`

### Commit Message Format
```
[Component] Brief description

Detailed explanation if needed.

Related: QUALITY-GATE-N, ADR-XXXX
Owner: agent-name
```

**Examples:**
```
[api] Define TimeWindowScaler CRD schema

Complete spec and status fields with OpenAPI validation.
Status subresource uses observedGeneration pattern.

Related: QUALITY-GATE-1, ADR-0001
Owner: api-crd-designer
```

```
[reconcile] Add state machine design with DST handling

State transitions: Inactive -> Active -> GracePeriod.
DST spring-forward and fall-back edge cases documented.

Related: QUALITY-GATE-3, ADR-0003
Owner: controller-reconcile-designer
```

---

## Daily Check-In Schedule

### Design Phase (Days 1-14)

**Morning Sync (10:00 IST) - Async via Commit Messages or Comments**
- Update format in commit message or design doc comment:
  ```
  Agent: {agent-id}
  Status: On track | At risk | Blocked
  Today's Goal: {one sentence}
  Blockers: {if any}
  ```

**Midday Check (14:00 IST) - Optional, for Cross-Dependencies**
- Triggered when one agent's work unblocks another
- Example: api-crd-designer completes spec, notifies validation and reconcile designers
- Format: Commit with message tagging dependent agents
  ```
  git commit -m "[api] CRD spec complete - @api-validation-defaults-designer @controller-reconcile-designer ready for Day 2 work"
  ```

**End of Day (18:00 IST) - Mandatory Status Update**
- Update ROADMAP.md with day completion status
- Commit all work-in-progress (even if incomplete)
- Tag kyklos-orchestrator if quality gate needs verification
- Format:
  ```
  Day N Status for {agent-id}:
  - Deliverables: [list with completion %]
  - Blockers: [none or describe]
  - Tomorrow's Plan: [one sentence]
  - Handoff To: [next agent(s) if applicable]
  ```

---

## Artifact Versioning and Ownership

### Document Lifecycle States
1. **Draft** - Work in progress, not ready for review
2. **Review** - Complete, awaiting feedback from Consulted parties
3. **Approved** - Reviewed and approved, ready for implementation
4. **Implemented** - Code written based on this design
5. **Archived** - Superseded by newer version or implementation complete

### Version Header in Design Documents
All design documents must include this header:
```markdown
# {Document Title}

**Version:** 0.1.0
**Status:** Draft | Review | Approved | Implemented | Archived
**Owner:** {agent-id}
**Last Updated:** YYYY-MM-DD
**Approvers:** {list of agent-ids who approved}

---
```

### Version Numbering
- **Major (X.0.0):** Breaking change to design (requires re-review)
- **Minor (0.X.0):** Significant addition (new section, major clarification)
- **Patch (0.0.X):** Minor edits (typo fixes, formatting, small clarifications)

**Examples:**
- Add new field to API spec: Minor version bump (0.1.0 → 0.2.0)
- Fix typo in field description: Patch version bump (0.1.0 → 0.1.1)
- Change state machine fundamentally: Major version bump (0.1.0 → 1.0.0)

### Ownership Rules
**Single Owner per Document:**
- Owner is Accountable party from RACI.md
- Only owner can approve status changes (Draft → Review → Approved)
- Others can suggest changes via comments

**Consulted Parties:**
- Must review before document moves to Approved status
- Document "Approvers:" field tracks who signed off
- Consulted agent comment format:
  ```markdown
  <!-- REVIEW by {agent-id} on {date}:
  Status: Approved | Changes Requested
  Comments: {feedback}
  -->
  ```

---

## Comment Resolution Protocol

### Comment Types
1. **Question:** Requires clarification from owner
   ```markdown
   <!-- QUESTION by {agent-id}:
   {question text}
   -->
   ```

2. **Suggestion:** Optional improvement
   ```markdown
   <!-- SUGGESTION by {agent-id}:
   Consider {suggestion}
   Rationale: {why}
   -->
   ```

3. **Issue:** Must be addressed before approval
   ```markdown
   <!-- ISSUE by {agent-id}:
   {problem description}
   Must fix because: {reason}
   -->
   ```

### Resolution Timeline
- **Questions:** Must be answered within 4 hours during working day (09:00-18:00 IST)
- **Suggestions:** Owner decides to accept/reject, document reasoning
- **Issues:** Must be resolved before document moves to Approved status

### Resolution Format
Owner responds inline:
```markdown
<!-- QUESTION by api-validation-defaults-designer:
What happens if timezone is invalid?
-->

<!-- ANSWER by api-crd-designer:
Admission webhook will reject the CR with validation error.
Documented in field semantics under timezone field.
-->
```

### Unresolved Comments
If comment cannot be resolved within 1 day:
1. Owner escalates to kyklos-orchestrator
2. kyklos-orchestrator assesses: scope change, needs ADR, needs clarification
3. Decision logged in DECISIONS.md
4. All parties notified within 2 hours of decision

---

## Conflict Resolution

### Conflict Types and Escalation

**Type 1: Design Disagreement**
- **Example:** Two agents disagree on field naming
- **Resolution:**
  1. Agents discuss and document both options
  2. If no consensus in 2 hours, escalate to kyklos-orchestrator
  3. kyklos-orchestrator decides based on BRIEF.md goals and consistency
  4. Create ADR in DECISIONS.md documenting decision and alternatives
  5. Both agents notified, work continues

**Type 2: Scope Ambiguity**
- **Example:** Agent unsure if feature is in or out of scope
- **Resolution:**
  1. Agent checks BRIEF.md goals/non-goals
  2. If unclear, post question as comment tagging kyklos-orchestrator
  3. kyklos-orchestrator clarifies within 2 hours
  4. Update BRIEF.md if needed (with version bump)
  5. Notify all agents if scope changes

**Type 3: Timeline Risk**
- **Example:** Quality gate at risk of missing deadline
- **Resolution:**
  1. Owner notifies kyklos-orchestrator at least 12 hours before deadline
  2. Options:
     - Reduce scope (move items to v0.2)
     - Extend deadline (impact dependent work)
     - Reallocate resources (another agent assists)
  3. kyklos-orchestrator decides, updates ROADMAP.md
  4. All affected agents notified immediately

**Type 4: Cross-Document Inconsistency**
- **Example:** API spec and reconcile design have conflicting information
- **Resolution:**
  1. Agent discovering conflict posts issue comment in both documents
  2. Tag both document owners and kyklos-orchestrator
  3. Owners coordinate to resolve (one or both documents updated)
  4. If technical decision, create ADR in DECISIONS.md
  5. kyklos-orchestrator verifies consistency restored

### Escalation Path
```
Agent (Responsible) → Document Owner (Accountable) → kyklos-orchestrator → All Agents (if scope/timeline impact)
```

### Escalation SLA
- **Agent to Owner:** Immediate (via comment)
- **Owner Response:** Within 4 hours
- **Owner to kyklos-orchestrator:** Within 2 hours if unable to resolve
- **kyklos-orchestrator Decision:** Within 2 hours of escalation
- **Notification to All:** Within 1 hour of decision

---

## Decision Change Process

### When to Create an ADR
Create a new ADR when:
- Choosing between multiple technical approaches
- Scope boundary decisions (what's in/out of v0.1)
- API design decisions that affect user experience
- Security or RBAC policy decisions
- Testing strategy decisions

### ADR Template Location
See `/docs/DECISIONS.md` for template at end of file.

### ADR Numbering
- Sequential: ADR-0001, ADR-0002, etc.
- No gaps (even if ADR is later superseded)

### ADR Lifecycle
1. **Proposed:** Draft ADR with decision, rationale, alternatives
2. **Under Review:** Consulted parties provide feedback (2-day review period)
3. **Accepted:** Decision is final, document status changes to Accepted
4. **Deprecated:** Later superseded by newer ADR (link to replacement)

### Amending an Accepted ADR
**Do NOT edit Accepted ADRs.** Instead:
1. Create new ADR superseding previous one
2. Link from new ADR: "Supersedes: ADR-XXXX"
3. Link from old ADR: "Superseded by: ADR-YYYY"
4. Update BRIEF.md if decision changes scope or goals

---

## Handoff Protocol

### Pre-Handoff Checklist
Before handing off to next agent, responsible agent must:
- [ ] All deliverables committed to git
- [ ] Quality gate checklist completed (in design doc or separate file)
- [ ] Consulted parties have reviewed and approved
- [ ] Comments resolved or escalated
- [ ] Document status changed to Approved
- [ ] Next agent tagged in commit message

### Handoff Package Contents
1. **Completed Artifacts:** Design docs, decision records
2. **References:** Link to BRIEF.md, relevant ADRs, QUALITY-GATES.md section
3. **Open Questions:** List of unresolved questions (if any) with escalation status
4. **Dependencies:** List of inputs required from other agents (with status: ready/pending)
5. **Acceptance Criteria:** Copy from QUALITY-GATES.md for receiving agent's gate

### Handoff Notification Format
Commit message or comment:
```
HANDOFF to {agent-id}

Completed:
- [x] {deliverable 1}
- [x] {deliverable 2}

References:
- design/api-crd-spec.md (v1.0.0, Approved)
- docs/DECISIONS.md (ADR-0001, ADR-0002)
- QUALITY-GATES.md Gate 2 (receiving agent's acceptance criteria)

Open Questions:
- Q1: {question} - Escalated to kyklos-orchestrator on {date}
- (none if all resolved)

Dependencies:
- Requires: design/api-crd-spec.md from api-crd-designer (Status: Ready)

Next: {receiving agent} proceeds with {next workstream}
```

### Receiving Agent Acknowledgment
Receiving agent must acknowledge within 4 hours:
```
ACK by {agent-id} on {timestamp}

Handoff received. Starting work on {workstream}.
Estimated completion: {date time IST}
Blockers: {none or describe}
```

If receiving agent identifies handoff gap (missing info, unclear requirement):
1. Post comment describing gap
2. Tag sender and kyklos-orchestrator
3. Sender responds within 4 hours
4. If gap blocks work, timeline escalation protocol applies

---

## Working Hours and Response Times

### Standard Working Hours
**09:00 - 18:00 IST (Asia/Kolkata)** on working days

### Response Time SLAs
| Request Type | Response Time | Applies To |
|--------------|---------------|------------|
| Clarification question | 4 hours | All agents |
| Review request | 1 day | Consulted parties |
| Escalation | 2 hours | kyklos-orchestrator |
| Handoff acknowledgment | 4 hours | Receiving agent |
| Conflict resolution | 2 hours (decision) | kyklos-orchestrator |

### After-Hours Policy
- No expectation of responses outside 09:00-18:00 IST
- Emergency escalations (critical blocker): Use dedicated channel (TBD)
- Timeline adjustments account for working hours

---

## Notification Methods

### When to Tag Agents
**Tag in Commit Message:**
```
git commit -m "[api] CRD spec complete

@api-validation-defaults-designer: Ready for validation design
@controller-reconcile-designer: Status fields defined as discussed

Related: QUALITY-GATE-1"
```

**Tag in Inline Comment:**
```markdown
<!-- @security-rbac-designer: Please review webhook RBAC permissions in this section -->
```

**Tag in Design Doc:**
```markdown
## Open Questions

1. Should webhook use ClusterRole or namespaced Role?
   - **Assigned to:** @security-rbac-designer
   - **Due:** Day 3 midday (2025-10-22 14:00 IST)
```

### Who to Notify When
| Event | Notify | Method |
|-------|--------|--------|
| Document moves to Review | Consulted parties (RACI) | Commit message |
| Quality gate complete | kyklos-orchestrator | Commit message or comment |
| Handoff ready | Next agent | Commit message |
| Blocker encountered | Owner and kyklos-orchestrator | Comment with BLOCKED tag |
| Decision needed | kyklos-orchestrator | Comment with DECISION-NEEDED tag |
| ADR proposed | All agents | Commit with ADR-REVIEW in message |

---

## Status Reporting Format

### Daily Status (End of Day 18:00 IST)
Update in commit message or ROADMAP.md:
```
Day N - {agent-id} Status

Work Completed:
- {item 1} - 100%
- {item 2} - 70%

Blockers:
- {blocker description} - Escalated to {whom} at {time}
- (none if no blockers)

Tomorrow's Plan:
- Complete {item 2}
- Begin {item 3}

On Track: Yes | At Risk | Blocked
```

### Weekly Rollup (Every Friday 17:00 IST)
kyklos-orchestrator posts:
```
Week {N} Summary

Completed:
- Gate 1: API CRD Design ✓
- Gate 2: Validation Design ✓

In Progress:
- Gate 3: Reconcile Design (80%)
- Gate 4: Metrics Design (50%)

Risks:
- DST test scenarios need more time (Medium risk)

Next Week:
- Complete Gates 3-5
- Begin test plan design
```

---

## Document Review Cycles

### Review Request
Owner posts:
```markdown
## Review Request

**Document:** design/api-crd-spec.md v1.0.0
**Status:** Draft → Review
**Reviewers:** @api-validation-defaults-designer @controller-reconcile-designer
**Due:** 2025-10-20 16:00 IST (Day 1 afternoon)

Please review for:
- Field completeness
- Validation rule coverage
- Consistency with BRIEF.md glossary
```

### Review Completion
Reviewer posts:
```markdown
<!-- REVIEW by api-validation-defaults-designer on 2025-10-20 15:30 IST

Status: Approved with minor suggestions

Strengths:
- Clear field definitions
- Good OpenAPI validation coverage

Suggestions:
- Add example for cross-midnight window
- Clarify gracePeriod default (suggest making explicit)

Blockers: None
-->
```

### Moving to Approved
After all reviewers approve:
1. Owner addresses suggestions or documents why declined
2. Owner updates document header: `Status: Approved`
3. Owner commits with message: `[api] CRD spec approved - ready for implementation`
4. Owner notifies kyklos-orchestrator for quality gate verification

---

## Communication Anti-Patterns (Avoid These)

**Don't:** Make assumptions without documenting
**Do:** Post question comment, tag relevant agent

**Don't:** Edit another agent's document without permission
**Do:** Suggest changes via comment, let owner incorporate

**Don't:** Skip handoff notification
**Do:** Use formal handoff format even if next agent is "aware"

**Don't:** Resolve conflicts by implementing your preference
**Do:** Escalate to owner, then kyklos-orchestrator if needed

**Don't:** Leave comments unresolved indefinitely
**Do:** Follow up within 1 day, escalate if no response

**Don't:** Change scope without updating BRIEF.md
**Do:** Propose scope change, get approval, update BRIEF.md and notify all agents

---

## Glossary of Communication Terms

**Handoff:** Transfer of work and context from one agent to another
**Escalation:** Raising issue to higher authority (kyklos-orchestrator) for decision
**Blocker:** Issue preventing progress, requires immediate attention
**At Risk:** Work may miss deadline without intervention
**Quality Gate:** Verification checkpoint before moving to next phase
**ADR:** Architecture Decision Record, documents significant decisions
**Consulted Party:** Agent who must review work per RACI before approval
**Owner:** Accountable agent for a document or workstream

---

## Communication Health Metrics

Track weekly:
- Average response time to questions (target: < 4 hours)
- Percentage of comments resolved within 1 day (target: > 90%)
- Number of escalations (lower is better, indicates clear requirements)
- Handoff acknowledgment time (target: < 4 hours)

kyklos-orchestrator reviews metrics and adjusts protocol if SLAs are frequently missed.

---

## Evolution of Communication Protocol

This protocol is for v0.1 design and implementation. Adjust if:
- Team size changes (more/fewer agents)
- Distributed timezone collaboration needed
- Faster iteration required (shorten review cycles)

**Change Process:**
1. Propose change to this document
2. All agents review (2-day review period)
3. kyklos-orchestrator approves
4. Update version, notify all agents
5. Apply new protocol from next day
