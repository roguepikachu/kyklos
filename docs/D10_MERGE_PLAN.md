# Day 10 Merge Plan: Ordered Edit Application Strategy

**Date:** 2025-10-30
**Coordinator:** kyklos-orchestrator
**Purpose:** Sequenced application to avoid conflicts and enable rollback
**Status:** Ready for execution

---

## Critical Pre-Execution Decision

**BEFORE APPLYING ANY EDITS:**

Hold emergency decision meeting (30 minutes):
- **Time:** Oct 30 10:00-10:30 IST
- **Participants:** kyklos-orchestrator, api-crd-designer, controller-reconcile-designer
- **Decision:** ADR-0005 - Are holidays in v0.1 or not?
- **Output:** Create `/Users/aykumar/personal/kyklos/docs/ADR-0005-DECISION.txt` with:
  ```
  DECISION: [HOLIDAYS_IN_V01 | HOLIDAYS_NOT_IN_V01]
  RATIONALE: [one paragraph]
  SIGNED: [names]
  DATE: 2025-10-30 10:30 IST
  ```

**Edit application cannot proceed until this file exists.**

---

## Merge Strategy Overview

### Phases
1. **ADR Phase** - Document decisions first (30 min)
2. **Source of Truth Phase** - BRIEF.md and DECISIONS.md (30 min)
3. **API Phase** - CRD specification (30 min)
4. **Logic Phase** - Reconcile design (30 min)
5. **Artifact Phase** - Test fixtures and workflows (30 min)
6. **Documentation Phase** - User docs and examples (1 hour)
7. **Test Plan Phase** - Test scenario details (1 hour)
8. **Verification Phase** - Validate all changes (30 min)

### Principles
1. **Decisions before implementation** - ADRs must be created before applying changes
2. **Source of truth first** - BRIEF.md defines scope and terminology
3. **API before logic** - CRD spec defines what reconcile operates on
4. **Critical path priority** - Blocking issues before nice-to-haves
5. **Create before modify** - New files before edits to existing files
6. **Validate incrementally** - Check consistency after each phase

---

## Phase 1: ADR Documentation (10:30-11:00 IST)

**Purpose:** Document all architectural decisions before changing code/docs
**Owner:** kyklos-orchestrator
**Duration:** 30 minutes

### Step 1.1: Create ADR-0005 (Holiday Scope)
**File:** `/Users/aykumar/personal/kyklos/docs/DECISIONS.md`
**Action:** Append ADR-0005 from D10_ADR_DELTA.md
**Dependency:** ADR-0005-DECISION.txt must exist
**Validation:** Check ADR-0005 section exists in DECISIONS.md

### Step 1.2: Create ADR-0006 (Validation Strategy)
**File:** `/Users/aykumar/personal/kyklos/docs/DECISIONS.md`
**Action:** Append ADR-0006 from D10_ADR_DELTA.md
**Decision:** CRD validation only (recommended for v0.1)
**Validation:** Check ADR-0006 section exists

### Step 1.3: Create ADR-0007 (Field Naming)
**File:** `/Users/aykumar/personal/kyklos/docs/DECISIONS.md`
**Action:** Append ADR-0007 from D10_ADR_DELTA.md
**Validation:** Check ADR-0007 section exists

### Step 1.4: Update ADR-0002
**File:** `/Users/aykumar/personal/kyklos/docs/DECISIONS.md`
**Action:** Add cross-namespace validation section per D10_ADR_DELTA.md
**Validation:** Verify ADR-0002 has validation strategy section

### Step 1.5: Update ADR-0004
**File:** `/Users/aykumar/personal/kyklos/docs/DECISIONS.md`
**Action:** Add field naming section per D10_ADR_DELTA.md
**Validation:** Verify ADR-0004 has gracePeriodSeconds clarification

**Phase 1 Checkpoint:**
```bash
grep "ADR-0005:" docs/DECISIONS.md
grep "ADR-0006:" docs/DECISIONS.md
grep "ADR-0007:" docs/DECISIONS.md
# All 3 must return results
```

---

## Phase 2: Source of Truth Updates (11:00-11:30 IST)

**Purpose:** Update BRIEF.md to reflect decided scope and correct terminology
**Owner:** kyklos-orchestrator
**Duration:** 30 minutes

### Step 2.1: Apply EDIT-002 (Version Requirements)
**File:** `/Users/aykumar/personal/kyklos/docs/BRIEF.md`
**Action:** Insert version section after line 3
**Dependency:** None
**Rollback:** Delete lines 5-10 if needed

### Step 2.2: Apply EDIT-001 (Glossary Terms)
**File:** `/Users/aykumar/personal/kyklos/docs/BRIEF.md`
**Action:** Replace lines 72-81 with updated glossary
**Dependency:** None
**Rollback:** Restore original 10 lines from git

**Critical:** This changes terminology used throughout project

### Step 2.3: Apply EDIT-003-A or EDIT-003-B (Holiday Scope)
**File:** `/Users/aykumar/personal/kyklos/docs/BRIEF.md`
**Action:**
- If HOLIDAYS_IN_V01: Apply EDIT-003-A (modify line 17)
- If HOLIDAYS_NOT_IN_V01: No change needed
**Dependency:** ADR-0005-DECISION.txt
**Rollback:** Restore line 17 from git

**Phase 2 Checkpoint:**
```bash
# Verify glossary no longer has activeReplicas
grep "activeReplicas" docs/BRIEF.md | grep -v "windows"
# Should return 0 results

# Verify version section exists
grep "Version Requirements" docs/BRIEF.md
# Should return 1 result
```

---

## Phase 3: API Specification Updates (11:30-12:00 IST)

**Purpose:** Update CRD spec to match decisions and add missing fields
**Owner:** api-crd-designer
**Duration:** 30 minutes

### Step 3.1: Apply EDIT-004 (Validation Method)
**File:** `/Users/aykumar/personal/kyklos/docs/api/CRD-SPEC.md`
**Action:** Replace lines 25-28 with new validation rules
**Dependency:** ADR-0006 (validation strategy decided)
**Rollback:** Restore 4 lines from git

### Step 3.2: Apply EDIT-005 (Grace Period Expiry Field)
**File:** `/Users/aykumar/personal/kyklos/docs/api/CRD-SPEC.md`
**Action:** Insert new field section after line 125
**Dependency:** ADR-0004 updated
**Rollback:** Delete inserted lines (search for "gracePeriodExpiry")

### Step 3.3: Apply EDIT-006-A or EDIT-006-B (Holiday Section)
**File:** `/Users/aykumar/personal/kyklos/docs/api/CRD-SPEC.md`
**Action:**
- If HOLIDAYS_IN_V01: Apply EDIT-006-A (update line 75 with default)
- If HOLIDAYS_NOT_IN_V01: Apply EDIT-006-B (delete lines 66-80)
**Dependency:** ADR-0005-DECISION.txt
**Rollback:**
- If A: Restore line 75
- If B: Restore deleted section from git

**Phase 3 Checkpoint:**
```bash
# Verify gracePeriodExpiry field exists
grep "gracePeriodExpiry" docs/api/CRD-SPEC.md
# Should return multiple results

# Verify cross-namespace allowed
grep "must equal the TimeWindowScaler's namespace" docs/api/CRD-SPEC.md
# Should return 0 results
```

---

## Phase 4: Reconcile Logic Updates (12:00-12:30 IST)

**Purpose:** Update reconcile design to match CRD spec and add detail
**Owner:** controller-reconcile-designer
**Duration:** 30 minutes

### Step 4.1: Apply EDIT-007 (Grace Period Field Name)
**File:** `/Users/aykumar/personal/kyklos/docs/design/RECONCILE.md`
**Action:** Fix field references in lines 73-74
**Dependency:** Phase 3 complete (CRD has gracePeriodExpiry)
**Rollback:** Restore 2 lines from git

### Step 4.2: Apply EDIT-008 (Expand Pause Semantics)
**File:** `/Users/aykumar/personal/kyklos/docs/design/RECONCILE.md`
**Action:** Replace lines 90-97 with expanded version
**Dependency:** None
**Rollback:** Restore original 8 lines from git

### Step 4.3: Apply EDIT-009-A or EDIT-009-B (Holiday Logic)
**File:** `/Users/aykumar/personal/kyklos/docs/design/RECONCILE.md`
**Action:**
- If HOLIDAYS_IN_V01: No change (EDIT-009-A)
- If HOLIDAYS_NOT_IN_V01: Delete Step 3, renumber, update preconditions (EDIT-009-B)
**Dependency:** ADR-0005-DECISION.txt, Phase 3 complete
**Rollback:**
- If B: Restore deleted Step 3 and restore numbering

**Phase 4 Checkpoint:**
```bash
# Verify pause semantics detailed
grep "Return early, do not proceed to Step 8" docs/design/RECONCILE.md
# Should return 1 result

# Verify grace field consistency
grep 'gracePeriodSeconds' docs/design/RECONCILE.md | wc -l
# Should return multiple results

grep 'spec\.gracePeriodSeconds' docs/design/RECONCILE.md
# Should return at least 2 results
```

---

## Phase 5: Artifact Creation (12:30-13:00 IST)

**Purpose:** Create missing test fixtures and CI workflow
**Owner:** testing-strategy-designer, ci-release-designer
**Duration:** 30 minutes
**Parallelizable:** These can be done concurrently

### Step 5.1: Create DST Fixtures (Parallel Track A)
**Owner:** testing-strategy-designer

#### 5.1a: Apply EDIT-010 (Spring Forward Fixture)
**File:** `/Users/aykumar/personal/kyklos/test/fixtures/dst-spring-2025.yaml`
**Action:** Create new file with content from edit pack
**Dependency:** Directory test/fixtures/ must exist (create if needed)
**Rollback:** `rm test/fixtures/dst-spring-2025.yaml`

#### 5.1b: Apply EDIT-011 (Fall Back Fixture)
**File:** `/Users/aykumar/personal/kyklos/test/fixtures/dst-fall-2025.yaml`
**Action:** Create new file with content from edit pack
**Dependency:** Directory test/fixtures/ must exist
**Rollback:** `rm test/fixtures/dst-fall-2025.yaml`

#### 5.1c: Apply EDIT-012 (Cross-Midnight Fixture)
**File:** `/Users/aykumar/personal/kyklos/test/fixtures/dst-cross-midnight-2025.yaml`
**Action:** Create new file with content from edit pack
**Dependency:** Directory test/fixtures/ must exist
**Rollback:** `rm test/fixtures/dst-cross-midnight-2025.yaml`

### Step 5.2: Create CI Workflow (Parallel Track B)
**Owner:** ci-release-designer

#### 5.2a: Apply EDIT-013 (CI Workflow)
**File:** `/Users/aykumar/personal/kyklos/.github/workflows/ci.yml`
**Action:** Create new file with content from edit pack
**Dependency:** Directory .github/workflows/ must exist (create if needed)
**Rollback:** `rm .github/workflows/ci.yml`

**Phase 5 Checkpoint:**
```bash
# Verify fixtures exist
ls -la test/fixtures/dst-*.yaml | wc -l
# Should return 3

# Verify workflow exists
test -f .github/workflows/ci.yml && echo "OK" || echo "MISSING"
# Should return OK
```

---

## Phase 6: Documentation Updates (13:00-14:00 IST)

**Purpose:** Update user-facing documentation and examples
**Owner:** docs-dx-designer, local-workflow-designer
**Duration:** 1 hour

### Step 6.1: Create MAKE-TARGETS.md
**Owner:** local-workflow-designer
**File:** `/Users/aykumar/personal/kyklos/docs/MAKE-TARGETS.md`
**Action:** Apply EDIT-015 (create new file)
**Dependency:** None
**Rollback:** `rm docs/MAKE-TARGETS.md`

### Step 6.2: Fix LOCAL-DEV-GUIDE.md Links
**Owner:** local-workflow-designer
**File:** `/Users/aykumar/personal/kyklos/docs/LOCAL-DEV-GUIDE.md`
**Action:** Apply EDIT-014 (fix link path)
**Dependency:** Step 6.1 complete (MAKE-TARGETS.md exists)
**Rollback:** Restore original link from git
**Note:** May have multiple occurrences, use find-replace

### Step 6.3: Update CONCEPTS.md Terminology
**Owner:** docs-dx-designer
**File:** `/Users/aykumar/personal/kyklos/docs/user/CONCEPTS.md`
**Action:** Apply EDIT-016 (expand effectiveReplicas explanation)
**Dependency:** Phase 2 complete (glossary updated)
**Rollback:** Restore original 3 lines from git

### Step 6.4: Add CONCEPTS.md Holiday Note
**Owner:** docs-dx-designer
**File:** `/Users/aykumar/personal/kyklos/docs/user/CONCEPTS.md`
**Action:**
- If HOLIDAYS_IN_V01: Apply EDIT-017-A (v0.1 note)
- If HOLIDAYS_NOT_IN_V01: Apply EDIT-017-B (v0.2 note)
**Dependency:** ADR-0005-DECISION.txt
**Rollback:** Delete inserted note

### Step 6.5: Handle Examples
**Owner:** docs-dx-designer
**Files:** `examples/*.yaml`
**Action:**
- If HOLIDAYS_IN_V01: Apply EDIT-019-A (validate all)
- If HOLIDAYS_NOT_IN_V01: Apply EDIT-019-B (move holiday example)
**Dependency:** ADR-0005-DECISION.txt
**Rollback:**
- If A: No changes if validation passed
- If B: `git mv examples/future/tws-holidays-closed.yaml examples/`

**Phase 6 Checkpoint:**
```bash
# Verify MAKE-TARGETS.md exists
test -f docs/MAKE-TARGETS.md && echo "OK" || echo "MISSING"

# Verify CONCEPTS.md has holiday note
grep "Holiday support" docs/user/CONCEPTS.md
# Should return at least 1 result

# Verify examples validate
kubectl apply --dry-run=client -f examples/*.yaml
# All should succeed
```

---

## Phase 7: Test Plan Expansion (14:00-15:00 IST)

**Purpose:** Add detailed test scenarios to test plans
**Owner:** testing-strategy-designer
**Duration:** 1 hour

### Step 7.1: Apply EDIT-020 (DST Unit Test Scenarios)
**File:** `/Users/aykumar/personal/kyklos/docs/testing/UNIT-PLAN.md`
**Action:** Append DST scenarios from edit pack
**Dependency:** Phase 5 complete (fixtures exist)
**Rollback:** Delete appended section

### Step 7.2: Apply EDIT-021 (Pause Envtest Scenarios)
**File:** `/Users/aykumar/personal/kyklos/docs/testing/ENVTEST-PLAN.md`
**Action:** Append pause scenarios from edit pack
**Dependency:** Phase 4 complete (pause semantics defined)
**Rollback:** Delete appended section

### Step 7.3: Add README Test Step (Optional)
**File:** `/Users/aykumar/personal/kyklos/README.md`
**Action:** Apply EDIT-018 (add test step to quick start)
**Dependency:** None
**Rollback:** Delete added step

**Phase 7 Checkpoint:**
```bash
# Verify DST scenarios in UNIT-PLAN
grep "DST-1: Spring Forward" docs/testing/UNIT-PLAN.md
# Should return 1 result

# Verify pause scenarios in ENVTEST-PLAN
grep "PAUSE-1: Pause During Active Window" docs/testing/ENVTEST-PLAN.md
# Should return 1 result
```

---

## Phase 8: Final Verification (15:00-15:30 IST)

**Purpose:** Validate all changes are consistent and correct
**Owner:** kyklos-orchestrator
**Duration:** 30 minutes

### Step 8.1: Run Terminology Checks
```bash
# Should return 0 results (except in D9/D10 docs)
git grep "activeReplicas" docs/ | grep -v "D9_\|D10_\|windows"

# Should return 0 results
git grep "inactiveReplicas" docs/ | grep -v "D9_\|D10_"

# Should return 0 results (except GracePeriod compound words)
git grep 'gracePeriod[^S]' docs/ | grep -v "D9_\|D10_\|GracePeriod"
```

### Step 8.2: Validate Cross-References
```bash
# Check all internal markdown links
find docs -name "*.md" -exec grep -H '\[.*\](.*\.md)' {} \; | \
  while read link; do
    # Manual validation needed - ensure all links resolve
    echo "$link"
  done
```

### Step 8.3: Run Example Validation
```bash
# All examples must validate
kubectl apply --dry-run=client -f examples/*.yaml
```

### Step 8.4: Check Artifact Existence
```bash
# Test fixtures
ls -la test/fixtures/dst-*.yaml | wc -l  # Should be 3

# CI workflow
test -f .github/workflows/ci.yml && echo "OK" || echo "MISSING"

# Documentation
test -f docs/MAKE-TARGETS.md && echo "OK" || echo "MISSING"
```

### Step 8.5: Verify ADR Completeness
```bash
# All new ADRs should exist
grep "ADR-0005:" docs/DECISIONS.md  # Holiday scope
grep "ADR-0006:" docs/DECISIONS.md  # Validation strategy
grep "ADR-0007:" docs/DECISIONS.md  # Field naming
```

**Phase 8 Checkpoint:**
All verification commands must pass. If any fail, trace back to the phase that should have fixed it and reapply.

---

## Rollback Procedures

### Full Rollback (Emergency)
```bash
# Discard all uncommitted changes
git checkout docs/
git clean -fd docs/
git clean -fd test/
git clean -fd .github/
```

### Partial Rollback (Phase-Level)

**Rollback Phase 1 (ADRs):**
```bash
git checkout docs/DECISIONS.md
```

**Rollback Phase 2 (BRIEF):**
```bash
git checkout docs/BRIEF.md
```

**Rollback Phase 3 (CRD Spec):**
```bash
git checkout docs/api/CRD-SPEC.md
```

**Rollback Phase 4 (Reconcile):**
```bash
git checkout docs/design/RECONCILE.md
```

**Rollback Phase 5 (Artifacts):**
```bash
rm -f test/fixtures/dst-*.yaml
rm -f .github/workflows/ci.yml
```

**Rollback Phase 6 (Documentation):**
```bash
git checkout docs/LOCAL-DEV-GUIDE.md
git checkout docs/user/CONCEPTS.md
git checkout README.md
rm -f docs/MAKE-TARGETS.md
git checkout examples/
rm -rf examples/future/
```

**Rollback Phase 7 (Test Plans):**
```bash
git checkout docs/testing/UNIT-PLAN.md
git checkout docs/testing/ENVTEST-PLAN.md
```

### Individual Edit Rollback
See "Rollback:" line in each step above for specific rollback commands.

---

## Conflict Prevention Strategy

### Why This Order?
1. **ADRs first:** Decisions documented before changes prevent reverting decisions later
2. **BRIEF next:** Source of truth must be updated before derived docs
3. **CRD before reconcile:** API spec defines what logic operates on
4. **Create before modify:** New files can't conflict with edits
5. **Critical before nice-to-have:** Blocking issues resolved first

### Dependencies Graph
```
Phase 1 (ADRs)
  ↓
Phase 2 (BRIEF) ← depends on ADR-0005
  ↓
Phase 3 (CRD) ← depends on ADR-0006, ADR-0007
  ↓
Phase 4 (Reconcile) ← depends on Phase 3
  ↓
Phase 5 (Artifacts) ← independent, can run parallel
  ↓
Phase 6 (Docs) ← depends on Phase 2, ADR-0005
  ↓
Phase 7 (Test Plans) ← depends on Phase 4, Phase 5
  ↓
Phase 8 (Verification) ← depends on all
```

### Parallel Opportunities
- Phase 5.1 (fixtures) and Phase 5.2 (workflow) can run simultaneously
- Phase 6.1 (MAKE-TARGETS) and Phase 6.3 (CONCEPTS) can run simultaneously
- Phase 7.1 (UNIT-PLAN) and Phase 7.2 (ENVTEST-PLAN) can run simultaneously

---

## Git Commit Strategy

### Option A: Single Atomic Commit (Recommended)
After all phases complete and verification passes:
```bash
git add docs/ test/ .github/ examples/
git commit -m "fix: resolve Day 9 review findings - 17 critical issues

- Update BRIEF.md glossary to match API field names
- Add ADR-0005 (holiday scope), ADR-0006 (validation), ADR-0007 (naming)
- Fix CRD-SPEC cross-namespace validation and add gracePeriodExpiry
- Expand RECONCILE pause semantics and fix grace field references
- Create DST test fixtures (spring, fall, cross-midnight)
- Add GitHub Actions CI workflow
- Create MAKE-TARGETS.md documentation
- Update CONCEPTS, LOCAL-DEV-GUIDE, examples

Resolves 17 critical inconsistencies and 34 gaps identified in D9 review.
Consistency score improves from 73/100 to 95+/100.

See docs/D10_CHANGELOG.md for complete change list.
"
```

### Option B: Phase-Based Commits
Commit after each phase completes:
```bash
# After Phase 1
git add docs/DECISIONS.md
git commit -m "docs: add ADR-0005, ADR-0006, ADR-0007 and update ADR-0002, ADR-0004"

# After Phase 2
git add docs/BRIEF.md
git commit -m "docs: update BRIEF glossary and add version requirements"

# ... etc for each phase
```

**Recommendation:** Use Option A for clean history, Option B if incremental review needed.

---

## Timeline Summary

| Phase | Duration | Start | End | Owner |
|-------|----------|-------|-----|-------|
| Pre-Decision | 30 min | 10:00 | 10:30 | All |
| Phase 1: ADRs | 30 min | 10:30 | 11:00 | kyklos-orchestrator |
| Phase 2: BRIEF | 30 min | 11:00 | 11:30 | kyklos-orchestrator |
| Phase 3: CRD | 30 min | 11:30 | 12:00 | api-crd-designer |
| Phase 4: Reconcile | 30 min | 12:00 | 12:30 | controller-reconcile-designer |
| Phase 5: Artifacts | 30 min | 12:30 | 13:00 | testing/ci designers |
| Phase 6: Docs | 1 hour | 13:00 | 14:00 | docs/local designers |
| Phase 7: Test Plans | 1 hour | 14:00 | 15:00 | testing-strategy-designer |
| Phase 8: Verify | 30 min | 15:00 | 15:30 | kyklos-orchestrator |
| **Total** | **5 hours** | **10:00** | **15:30** | **All** |

---

## Success Criteria

At 15:30 IST, all of the following must be true:
- [ ] ADR-0005-DECISION.txt exists with clear decision
- [ ] All 21 edits from D10_EDIT_PACK.md applied
- [ ] No broken documentation links
- [ ] All examples validate with kubectl
- [ ] Terminology checks pass (0 results for old terms)
- [ ] All new files created (3 fixtures, 1 workflow, 1 doc)
- [ ] Verification commands pass
- [ ] Git working tree is clean (all committed)

---

**Prepared by:** kyklos-orchestrator
**Date:** 2025-10-30 09:45 IST
**Status:** Ready for execution
**Next Step:** Hold decision meeting at 10:00 IST
