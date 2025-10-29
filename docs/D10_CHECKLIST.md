# Day 10 Verification Checklist

**Date:** 2025-10-30
**Coordinator:** kyklos-orchestrator
**Purpose:** Pass/fail checks to confirm all Day 9 findings resolved
**Status:** Ready for execution

---

## Instructions

Run each check in sequence after completing all edits from D10_EDIT_PACK.md. Each check must PASS before proceeding to sign-off.

**Pass Criteria:** Check returns expected result (0 errors, files exist, etc.)
**Fail Criteria:** Check returns unexpected result → trace to phase and fix

---

## CHECK-01: Glossary Terminology Cleanup

**Purpose:** Verify obsolete terms removed from documentation
**Owner:** kyklos-orchestrator
**Run After:** Phase 2 complete (BRIEF.md updated)

### Commands:

```bash
# Check 1a: No activeReplicas except in context of windows[].replicas
git grep "activeReplicas" docs/ | grep -v "D9_\|D10_\|windows"
```
**Expected:** 0 results
**If Fail:** activeReplicas still present → review EDIT-001

```bash
# Check 1b: No inactiveReplicas
git grep "inactiveReplicas" docs/ | grep -v "D9_\|D10_"
```
**Expected:** 0 results
**If Fail:** inactiveReplicas still present → review EDIT-001

```bash
# Check 1c: effectiveReplicas present in glossary
grep "effectiveReplicas" docs/BRIEF.md
```
**Expected:** At least 1 match in glossary section
**If Fail:** New term not added → review EDIT-001

**Status:** [ ] PASS [ ] FAIL

---

## CHECK-02: Grace Period Field Consistency

**Purpose:** Verify consistent use of gracePeriodSeconds (not gracePeriod)
**Owner:** api-crd-designer
**Run After:** Phase 3 and 4 complete

### Commands:

```bash
# Check 2a: No bare gracePeriod references (except compound words)
git grep 'gracePeriod[^S]' docs/ | grep -v "D9_\|D10_\|GracePeriod"
```
**Expected:** 0 results
**If Fail:** Inconsistent field name → review EDIT-007

```bash
# Check 2b: gracePeriodExpiry field exists in CRD spec
grep "gracePeriodExpiry" docs/api/CRD-SPEC.md
```
**Expected:** Multiple matches in status section
**If Fail:** Missing status field → review EDIT-005

```bash
# Check 2c: spec.gracePeriodSeconds used in reconcile
grep "spec\.gracePeriodSeconds" docs/design/RECONCILE.md | wc -l
```
**Expected:** At least 2 matches
**If Fail:** Wrong field reference → review EDIT-007

**Status:** [ ] PASS [ ] FAIL

---

## CHECK-03: Cross-Namespace Validation Fixed

**Purpose:** Verify same-namespace constraint removed from CRD
**Owner:** api-crd-designer
**Run After:** Phase 3 complete

### Commands:

```bash
# Check 3a: No same-namespace constraint
grep "must equal the TimeWindowScaler's namespace" docs/api/CRD-SPEC.md
```
**Expected:** 0 results
**If Fail:** Constraint still present → review EDIT-004

```bash
# Check 3b: Cross-namespace mentioned as allowed
grep -A 2 "namespace may differ" docs/api/CRD-SPEC.md
```
**Expected:** At least 1 match mentioning ClusterRole
**If Fail:** New text not present → review EDIT-004

**Status:** [ ] PASS [ ] FAIL

---

## CHECK-04: DST Test Fixtures Exist

**Purpose:** Verify critical DST test artifacts created
**Owner:** testing-strategy-designer
**Run After:** Phase 5 complete

### Commands:

```bash
# Check 4a: All three fixtures exist
ls -1 test/fixtures/dst-*.yaml | wc -l
```
**Expected:** 3 (spring, fall, cross-midnight)
**If Fail:** Fixtures missing → review EDIT-010, EDIT-011, EDIT-012

```bash
# Check 4b: Spring forward fixture has correct date
grep "2025-03-09" test/fixtures/dst-spring-2025.yaml
```
**Expected:** At least 1 match
**If Fail:** Wrong date or file malformed → review EDIT-010

```bash
# Check 4c: Fall back fixture has correct date
grep "2025-11-02" test/fixtures/dst-fall-2025.yaml
```
**Expected:** At least 1 match
**If Fail:** Wrong date or file malformed → review EDIT-011

**Status:** [ ] PASS [ ] FAIL

---

## CHECK-05: CI Workflow Created

**Purpose:** Verify GitHub Actions workflow file exists
**Owner:** ci-release-designer
**Run After:** Phase 5 complete

### Commands:

```bash
# Check 5a: Workflow file exists
test -f .github/workflows/ci.yml && echo "PASS" || echo "FAIL"
```
**Expected:** PASS
**If Fail:** Workflow not created → review EDIT-013

```bash
# Check 5b: Workflow has basic jobs
grep -E "(lint|test-unit|verify)" .github/workflows/ci.yml | wc -l
```
**Expected:** At least 3 (3 job names)
**If Fail:** Workflow incomplete → review EDIT-013

```bash
# Check 5c: Workflow syntax is valid YAML
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))" && echo "PASS" || echo "FAIL"
```
**Expected:** PASS (or use `yamllint` if installed)
**If Fail:** YAML syntax error → fix syntax

**Status:** [ ] PASS [ ] FAIL

---

## CHECK-06: MAKE-TARGETS.md Created

**Purpose:** Verify Makefile documentation exists
**Owner:** local-workflow-designer
**Run After:** Phase 6 complete

### Commands:

```bash
# Check 6a: File exists
test -f docs/MAKE-TARGETS.md && echo "PASS" || echo "FAIL"
```
**Expected:** PASS
**If Fail:** File not created → review EDIT-015

```bash
# Check 6b: Contains target categories
grep -E "(Setup|Build|Deploy|Testing|Demo)" docs/MAKE-TARGETS.md | wc -l
```
**Expected:** At least 5 (5 category headings)
**If Fail:** File incomplete → review EDIT-015

**Status:** [ ] PASS [ ] FAIL

---

## CHECK-07: Documentation Links Valid

**Purpose:** Verify all internal markdown links resolve
**Owner:** local-workflow-designer
**Run After:** Phase 6 complete

### Commands:

```bash
# Check 7a: MINUTE-DEMO link correct in LOCAL-DEV-GUIDE
grep "\[MINUTE-DEMO\.md\](./user/MINUTE-DEMO\.md)" docs/LOCAL-DEV-GUIDE.md
```
**Expected:** At least 1 match
**If Fail:** Link not fixed → review EDIT-014

```bash
# Check 7b: MAKE-TARGETS link present
grep "MAKE-TARGETS\.md" docs/LOCAL-DEV-GUIDE.md
```
**Expected:** At least 1 match
**If Fail:** Link not added or file doesn't exist → check Phase 6

**Status:** [ ] PASS [ ] FAIL

---

## CHECK-08: Examples Validate

**Purpose:** Verify example YAMLs are syntactically valid
**Owner:** docs-dx-designer
**Run After:** Phase 6 complete

### Commands:

```bash
# Check 8a: Office hours example validates
kubectl apply --dry-run=client -f examples/tws-office-hours.yaml && echo "PASS" || echo "FAIL"
```
**Expected:** PASS (created dry run)
**If Fail:** YAML invalid → fix example

```bash
# Check 8b: Night shift example validates
kubectl apply --dry-run=client -f examples/tws-night-shift.yaml && echo "PASS" || echo "FAIL"
```
**Expected:** PASS
**If Fail:** YAML invalid → fix example

```bash
# Check 8c: Holiday example handled correctly
# If holidays IN v0.1:
kubectl apply --dry-run=client -f examples/tws-holidays-closed.yaml && echo "PASS" || echo "FAIL"

# If holidays NOT in v0.1:
test -f examples/future/tws-holidays-closed.yaml && echo "PASS (moved)" || echo "FAIL"
```
**Expected:** PASS (validates or moved to future/)
**If Fail:** Not handled per ADR-0005 → review EDIT-019-A or EDIT-019-B

**Status:** [ ] PASS [ ] FAIL

---

## CHECK-09: Pause Semantics Complete

**Purpose:** Verify pause logic fully detailed in reconcile
**Owner:** controller-reconcile-designer
**Run After:** Phase 4 complete

### Commands:

```bash
# Check 9a: Pause early return documented
grep "Return early, do not proceed to Step 8" docs/design/RECONCILE.md
```
**Expected:** 1 match in Step 7
**If Fail:** Pause logic not expanded → review EDIT-008

```bash
# Check 9b: Pause status update documented
grep "Update all status fields" docs/design/RECONCILE.md
```
**Expected:** 1 match in pause section
**If Fail:** Status handling not detailed → review EDIT-008

```bash
# Check 9c: ScalingSkipped event documented
grep "ScalingSkipped" docs/design/RECONCILE.md
```
**Expected:** 1 match in pause section
**If Fail:** Event not documented → review EDIT-008

**Status:** [ ] PASS [ ] FAIL

---

## CHECK-10: ADRs Complete

**Purpose:** Verify all new ADRs added to DECISIONS.md
**Owner:** kyklos-orchestrator
**Run After:** Phase 1 complete

### Commands:

```bash
# Check 10a: ADR-0005 exists (Holiday Scope)
grep "ADR-0005:" docs/DECISIONS.md
```
**Expected:** 1 match
**If Fail:** ADR not added → review Phase 1 Step 1.1

```bash
# Check 10b: ADR-0006 exists (Validation Strategy)
grep "ADR-0006:" docs/DECISIONS.md
```
**Expected:** 1 match
**If Fail:** ADR not added → review Phase 1 Step 1.2

```bash
# Check 10c: ADR-0007 exists (Field Naming)
grep "ADR-0007:" docs/DECISIONS.md
```
**Expected:** 1 match
**If Fail:** ADR not added → review Phase 1 Step 1.3

```bash
# Check 10d: ADR-0002 updated (Cross-namespace)
grep -A 5 "ADR-0002:" docs/DECISIONS.md | grep "Validation Strategy"
```
**Expected:** 1 match
**If Fail:** ADR not updated → review Phase 1 Step 1.4

```bash
# Check 10e: ADR-0004 updated (Grace period fields)
grep -A 5 "ADR-0004:" docs/DECISIONS.md | grep "Field Naming"
```
**Expected:** 1 match
**If Fail:** ADR not updated → review Phase 1 Step 1.5

**Status:** [ ] PASS [ ] FAIL

---

## FINAL SUMMARY CHECK

### Prerequisites
All checks CHECK-01 through CHECK-10 must PASS before final summary.

### Final Validation

```bash
# Overall consistency check - no critical terms should remain
TERMS=("activeReplicas" "inactiveReplicas" "gracePeriod[^S]")
for term in "${TERMS[@]}"; do
  echo "Checking: $term"
  git grep "$term" docs/ | grep -v "D9_\|D10_\|windows\|GracePeriod" || echo "  ✓ Clean"
done
```

### Quality Gates Status After Fixes

| Gate | Before | After | Status |
|------|--------|-------|--------|
| 1: CRD Schema | FAIL | PASS | ✓ |
| 2: Validation | PENDING | PASS | ✓ |
| 3: Reconcile | CONDITIONAL | PASS | ✓ |
| 4: Observability | PASS | PASS | ✓ |
| 5: RBAC | PASS | PASS | ✓ |
| 6: Local Workflow | CONDITIONAL | PASS | ✓ |
| 7: Testing | FAIL | PASS | ✓ |
| 8: CI/CD | FAIL | PASS | ✓ |
| 9: Docs | CONDITIONAL | PASS | ✓ |

**Expected:** 9/9 gates PASS

### Consistency Score

| Area | Before | After | Status |
|------|--------|-------|--------|
| Feature Scope | 22/100 | 100/100 | ✓ |
| Field Names | 56/100 | 100/100 | ✓ |
| Status Conditions | 100/100 | 100/100 | ✓ |
| Terminology | 71/100 | 100/100 | ✓ |
| Logic | 78/100 | 95/100 | ✓ |
| Observability | 95/100 | 95/100 | ✓ |
| Testing | 40/100 | 100/100 | ✓ |
| Examples | 33/100 | 100/100 | ✓ |
| Cross-References | 75/100 | 100/100 | ✓ |
| **Overall** | **73/100** | **98/100** | ✓ |

**Expected:** Overall score >= 95/100

---

## Sign-Off Checklist

After all checks pass:

- [ ] All 10 verification checks PASS
- [ ] All 9 quality gates PASS
- [ ] Consistency score >= 95/100
- [ ] No broken documentation links
- [ ] All examples validate
- [ ] All test fixtures exist
- [ ] CI workflow functional
- [ ] All ADRs documented

**Sign-Off:**
- **Reviewer:** kyklos-orchestrator
- **Date:** 2025-10-30 ______ IST
- **Status:** APPROVED FOR SCOPE LOCK

**Signature Line:**
```
I verify that all Day 9 review findings have been resolved and all verification
checks pass. The Kyklos Time Window Scaler v0.1 design is consistent, complete,
and ready for scope lock on Day 13.

Signed: ___________________
Date: 2025-10-30 ______ IST
```

---

## Troubleshooting Failed Checks

### If CHECK-01 Fails (Terminology)
**Root Cause:** BRIEF.md glossary not fully updated
**Fix:** Re-apply EDIT-001, verify all terms replaced

### If CHECK-02 Fails (Grace Period)
**Root Cause:** Field name inconsistencies remain
**Fix:** Search/replace all `gracePeriod ` → `gracePeriodSeconds ` (note space)

### If CHECK-03 Fails (Cross-Namespace)
**Root Cause:** CRD validation constraint not updated
**Fix:** Re-apply EDIT-004, ensure old text completely removed

### If CHECK-04 Fails (DST Fixtures)
**Root Cause:** Test fixtures not created or wrong location
**Fix:** Create `test/fixtures/` directory, re-apply EDIT-010, EDIT-011, EDIT-012

### If CHECK-05 Fails (CI Workflow)
**Root Cause:** Workflow file missing or invalid YAML
**Fix:** Create `.github/workflows/` directory, re-apply EDIT-013, check YAML syntax

### If CHECK-06 Fails (MAKE-TARGETS)
**Root Cause:** Documentation file not created
**Fix:** Re-apply EDIT-015

### If CHECK-07 Fails (Links)
**Root Cause:** Documentation cross-references not updated
**Fix:** Re-apply EDIT-014, verify target files exist

### If CHECK-08 Fails (Examples)
**Root Cause:** Example YAML invalid or not moved per ADR-0005
**Fix:** Validate YAML syntax, apply EDIT-019-A or EDIT-019-B per decision

### If CHECK-09 Fails (Pause Semantics)
**Root Cause:** RECONCILE.md pause section not expanded
**Fix:** Re-apply EDIT-008, ensure full replacement not partial

### If CHECK-10 Fails (ADRs)
**Root Cause:** DECISIONS.md not updated with all ADRs
**Fix:** Review Phase 1 steps, append all ADRs from D10_ADR_DELTA.md

---

**Prepared by:** kyklos-orchestrator
**Date:** 2025-10-30 10:00 IST
**Status:** Ready for execution
**Run After:** All phases of D10_MERGE_PLAN.md complete
