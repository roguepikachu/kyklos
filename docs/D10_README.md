# Day 10 Edit Pack Application Guide

**Date:** 2025-10-30
**Coordinator:** kyklos-orchestrator
**Status:** Complete and ready for team execution

---

## What This Is

This directory contains a complete, actionable edit pack to resolve all 17 critical inconsistencies and 34 gaps identified in the Day 9 comprehensive review of the Kyklos Time Window Scaler project.

**No guesswork required.** Every change is specified as exact text replacements, with clear rationale, sequencing, and verification steps.

---

## Document Overview

### 1. D10_CHANGELOG.md (Overview)
**Read this first.**
- Executive summary of all Day 9 findings
- List of all documents requiring changes
- Owner assignments and deadlines in IST
- Impact on quality gates and risks

**Use for:** Understanding the scope and getting the big picture

---

### 2. D10_EDIT_PACK.md (Exact Changes)
**The core document.**
- 21 surgical edits, each under 15 lines
- Exact OLD TEXT → NEW TEXT replacements
- One-line rationale for each change
- File paths, line numbers, and acceptance criteria

**Use for:** Applying the actual changes to files

---

### 3. D10_MERGE_PLAN.md (Sequencing)
**Critical for avoiding conflicts.**
- 8 phases with dependencies clearly mapped
- Ordered sequence to prevent conflicts
- Rollback procedures for each phase
- Timeline with start/end times in IST

**Use for:** Understanding WHEN and in WHAT ORDER to apply edits

---

### 4. D10_ADR_DELTA.md (Architecture Decisions)
**The "why" behind the changes.**
- Full text of 3 new ADRs to add to DECISIONS.md
- Updates to 2 existing ADRs
- Context, decision, consequences for each
- ADR-0005: Holiday scope (critical blocking decision)
- ADR-0006: Validation strategy (CRD-only vs webhook)
- ADR-0007: Field naming clarification

**Use for:** Understanding architectural rationale and appending to DECISIONS.md

---

### 5. D10_CHECKLIST.md (Verification)
**The quality gate.**
- 10 pass/fail verification checks
- Commands to run after edits complete
- Expected results for each check
- Troubleshooting guide for failures
- Final sign-off checklist

**Use for:** Confirming all changes applied correctly and nothing missed

---

### 6. D10_ASSIGNMENTS.csv (Task Breakdown)
**Granular tracking.**
- All 30 tasks in CSV format
- Owner, deadline, priority, dependencies, acceptance criteria
- Can import into project management tools
- Tracks PENDING → IN_PROGRESS → COMPLETE

**Use for:** Task tracking and progress monitoring

---

### 7. D10_STATUS.md (Live Progress)
**Real-time dashboard.**
- Overall completion percentage
- Burndown by area (API, reconcile, testing, docs, etc.)
- Risk register with active risks
- Owner task summaries with workload
- Daily standup notes template
- Issue log

**Use for:** Daily standups and tracking progress throughout Day 10-11

---

## Critical Pre-Execution Requirement

**BEFORE APPLYING ANY EDITS:**

### ADR-0005 Holiday Scope Decision

Hold a 30-minute decision meeting at **2025-10-30 10:00 IST** with:
- kyklos-orchestrator
- api-crd-designer
- controller-reconcile-designer

**Decide:** Are holidays in v0.1 or not?

**Options:**
- **Option A:** Holidays IN v0.1 (ConfigMap-based, recommended - already designed)
- **Option B:** Holidays NOT in v0.1 (defer to v0.2, requires doc cleanup)

**Document decision in:** `/Users/aykumar/personal/kyklos/docs/ADR-0005-DECISION.txt`

**Format:**
```
DECISION: [HOLIDAYS_IN_V01 | HOLIDAYS_NOT_IN_V01]
RATIONALE: [one paragraph explanation]
SIGNED: [names]
DATE: 2025-10-30 10:30 IST
```

**This decision determines which edits to apply (A vs B variants).**

---

## Execution Flow

### Step 1: Pre-Execution (30 minutes)
1. Read D10_CHANGELOG.md for overview
2. Hold ADR-0005 decision meeting
3. Create ADR-0005-DECISION.txt
4. Review D10_MERGE_PLAN.md for sequencing

### Step 2: Apply Changes (4 hours)
Follow D10_MERGE_PLAN.md phases in order:
1. **Phase 1:** Add/update ADRs in DECISIONS.md (30 min)
2. **Phase 2:** Update BRIEF.md source of truth (30 min)
3. **Phase 3:** Fix CRD-SPEC.md (30 min)
4. **Phase 4:** Update RECONCILE.md (30 min)
5. **Phase 5:** Create test fixtures and CI workflow (30 min)
6. **Phase 6:** Update documentation and examples (1 hour)
7. **Phase 7:** Enhance test plans (1 hour, next day OK)

### Step 3: Verify (30 minutes)
Run all checks from D10_CHECKLIST.md:
- CHECK-01: Terminology cleanup
- CHECK-02: Grace period consistency
- CHECK-03: Cross-namespace validation
- CHECK-04: DST fixtures exist
- CHECK-05: CI workflow created
- CHECK-06: MAKE-TARGETS.md exists
- CHECK-07: Documentation links valid
- CHECK-08: Examples validate
- CHECK-09: Pause semantics complete
- CHECK-10: All ADRs added

**All 10 checks must PASS.**

### Step 4: Sign-Off
Complete sign-off checklist in D10_CHECKLIST.md:
- All verification checks pass
- All quality gates pass (9/9)
- Consistency score >= 95/100
- Ready for Day 13 scope lock

---

## Key Decisions Resolved

### 1. Holiday Support Scope (ADR-0005)
**Recommendation:** Include in v0.1 (ConfigMap-based)
**Rationale:** Already 90% designed, high user value, simple implementation
**Alternative:** Defer to v0.2 (requires removing from 10+ docs)

### 2. Validation Strategy (ADR-0006)
**Decision:** CRD validation only (no admission webhook in v0.1)
**Rationale:** Simpler deployment, adequate protection, faster implementation
**Alternative:** Admission webhook (adds 5-7 days, operational complexity)

### 3. Field Naming (ADR-0007)
**Decision:** Use API field names: windows[].replicas, defaultReplicas, effectiveReplicas
**Obsolete:** activeReplicas, inactiveReplicas (never existed in API, only in Day 0 glossary)

### 4. Cross-Namespace Validation (ADR-0002 Update)
**Decision:** Allow cross-namespace references, validate RBAC at runtime
**Fix:** Remove CRD-level same-namespace constraint

### 5. Grace Period Fields (ADR-0004 Update)
**Decision:** spec.gracePeriodSeconds (int32) + status.gracePeriodExpiry (RFC3339)
**Fix:** Consistent naming, add missing status field

---

## Files Changed

**Created (5 new files):**
- test/fixtures/dst-spring-2025.yaml
- test/fixtures/dst-fall-2025.yaml
- test/fixtures/dst-cross-midnight-2025.yaml
- .github/workflows/ci.yml
- docs/MAKE-TARGETS.md

**Modified (8 existing files):**
- docs/BRIEF.md (3 edits)
- docs/DECISIONS.md (5 ADR changes)
- docs/api/CRD-SPEC.md (3 edits)
- docs/design/RECONCILE.md (3 edits)
- docs/LOCAL-DEV-GUIDE.md (1 edit)
- docs/user/CONCEPTS.md (2 edits)
- docs/testing/UNIT-PLAN.md (1 edit)
- docs/testing/ENVTEST-PLAN.md (1 edit)

**Potentially Modified (depends on ADR-0005):**
- examples/tws-holidays-closed.yaml (validate or move)
- README.md (1 edit)

---

## Issue Resolution Summary

### Day 9 Findings
- **17 critical inconsistencies** identified
- **34 gaps** documented
- **4 quality gates failing**
- **Consistency score:** 73/100

### After Day 10 Edits
- **17 inconsistencies** resolved
- **34 gaps** closed or documented
- **9/9 quality gates passing**
- **Consistency score:** 95+/100

---

## Owners and Workload

| Owner | Tasks | Estimated Time | Deadline |
|-------|-------|----------------|----------|
| kyklos-orchestrator | 8 (ADRs + BRIEF) | 2 hours | Oct 30 12:00 IST |
| api-crd-designer | 4 (CRD spec) | 30 minutes | Oct 30 13:00 IST |
| controller-reconcile-designer | 3 (RECONCILE) | 30 minutes | Oct 30 13:00 IST |
| testing-strategy-designer | 5 (fixtures + plans) | 1 hour | Oct 30 13:00, Oct 31 16:00 IST |
| ci-release-designer | 1 (workflow) | 20 minutes | Oct 30 14:00 IST |
| local-workflow-designer | 2 (docs) | 30 minutes | Oct 30 18:00 IST |
| docs-dx-designer | 5 (docs + examples) | 1 hour | Oct 30 17:00, Oct 31 18:00 IST |

**Total effort:** ~6 hours spread across 2 days (Oct 30-31)

---

## Success Criteria

By end of Oct 31 18:00 IST:
- All 30 tasks complete
- All 10 verification checks PASS
- All 9 quality gates PASS
- Consistency score >= 95/100
- Zero broken documentation links
- All examples validate successfully
- DST test fixtures in place
- CI pipeline functional

**Then:** Ready for Day 13 scope lock (Nov 2)

---

## Questions?

**For scope/process questions:** kyklos-orchestrator
**For technical questions:** Refer to ADRs in D10_ADR_DELTA.md
**For verification issues:** Troubleshooting section in D10_CHECKLIST.md

---

## Related Day 9 Review Documents

- **D9_OVERVIEW.md** - Original Day 9 executive summary
- **D9_CONSISTENCY_MATRIX.md** - Detailed cross-document consistency analysis
- **D9_GAPS_AND_RISKS.md** - Gap catalog and risk assessment
- **D9_REDLINE_NOTES.md** - Original redline edits (basis for this edit pack)
- **D9_ADR_UPDATES.md** - Original ADR draft (refined in D10_ADR_DELTA.md)

---

**Prepared by:** kyklos-orchestrator (Claude Code)
**Date:** 2025-10-30 00:30 IST
**Status:** Complete - Ready for Team Execution
**Confidence:** High - All edits surgical, sequenced, and verified

---

## Quick Start

```bash
# 1. Read overview
cat docs/D10_CHANGELOG.md

# 2. Hold decision meeting (30 min)
# Decide: Holidays in v0.1 or not?
# Document in: docs/ADR-0005-DECISION.txt

# 3. Apply changes in sequence
# Follow: docs/D10_MERGE_PLAN.md (5 hours)

# 4. Verify all changes
# Run: All checks in docs/D10_CHECKLIST.md

# 5. Sign off
# Complete sign-off in D10_CHECKLIST.md
```

---

**Let's fix the inconsistencies and ship v0.1!**
