# Day 9 Review: Quick Verification Procedure

**Purpose:** Fast coherence check for Kyklos documentation after fixes applied
**Time Required:** 15 minutes
**Owner:** Any agent or kyklos-orchestrator
**When to Run:** After each fix sprint, before scope lock, before sign-off

---

## Quick Start

```bash
# From repository root
cd /Users/aykumar/personal/kyklos

# Run full verification
./docs/D9_VERIFICATION_README.md --run-checks

# Or follow manual steps below
```

---

## Verification Checklist

### Phase 1: Critical Fixes (5 minutes)

#### 1.1 Holiday Scope Decision Made?

```bash
# Check for ADR-0005
grep -q "ADR-0005" docs/DECISIONS.md && echo "✅ Decision documented" || echo "❌ ADR-0005 missing"

# Verify BRIEF updated
grep -q "Calendar integration or holiday awareness" docs/BRIEF.md && echo "⚠️ Still listed as non-goal - check if correct" || echo "✅ Non-goals updated"

# Check consistency
if grep -q "Holiday" docs/api/CRD-SPEC.md; then
    echo "Holidays in CRD - verifying complete design..."
    [ -f docs/design/holiday-logic.md ] || grep -q "Holiday" docs/design/RECONCILE.md && echo "✅ Logic exists" || echo "❌ Missing holiday logic"
else
    echo "Holidays not in CRD - verifying cleanup..."
    ! grep -q "Holiday" docs/user/CONCEPTS.md && echo "✅ Removed from user docs" || echo "❌ Still in CONCEPTS"
fi
```

**Expected:** One of two states - holidays fully in v0.1 OR fully removed from v0.1 docs.

---

#### 1.2 DST Test Fixtures Created?

```bash
# Check for DST fixtures
ls -la test/fixtures/dst-*.yaml 2>/dev/null | wc -l | xargs -I {} echo "Found {} DST fixtures (expect 3)"

# Verify content
for file in test/fixtures/dst-*.yaml; do
    [ -f "$file" ] && grep -q "2025" "$file" && echo "✅ $file has 2025 date" || echo "❌ $file missing or wrong format"
done
```

**Expected:** 3 files (spring, fall, cross-midnight) with 2025 dates.

---

#### 1.3 GitHub Workflows Created?

```bash
# Check CI workflow
[ -f .github/workflows/ci.yml ] && echo "✅ CI workflow exists" || echo "❌ ci.yml missing"

# Basic validation
[ -f .github/workflows/ci.yml ] && grep -q "jobs:" .github/workflows/ci.yml && echo "✅ Has jobs" || echo "❌ Invalid structure"
```

**Expected:** At minimum ci.yml exists with lint/test/build jobs.

---

#### 1.4 Glossary Updated?

```bash
# Check for obsolete terms
git grep "activeReplicas" docs/ | grep -v "D9_" | grep -v "\.git" && echo "❌ activeReplicas still exists" || echo "✅ No activeReplicas found"
git grep "inactiveReplicas" docs/ | grep -v "D9_" | grep -v "\.git" && echo "❌ inactiveReplicas still exists" || echo "✅ No inactiveReplicas found"

# Check for new terms
grep -q "effectiveReplicas" docs/BRIEF.md && echo "✅ effectiveReplicas in glossary" || echo "❌ effectiveReplicas missing from glossary"
```

**Expected:** No activeReplicas or inactiveReplicas in docs (except review docs). effectiveReplicas in glossary.

---

#### 1.5 Cross-Namespace Validation Fixed?

```bash
# Check CRD-SPEC for blocking constraint
grep -q "must equal the TimeWindowScaler's namespace" docs/api/CRD-SPEC.md && echo "❌ Cross-namespace still blocked" || echo "✅ Cross-namespace allowed"

# Verify ADR-0002 alignment
grep -q "Cross-namespace support" docs/DECISIONS.md && echo "✅ ADR-0002 still supports cross-namespace" || echo "❌ ADR-0002 changed"
```

**Expected:** CRD does not block cross-namespace, ADR-0002 unchanged.

---

### Phase 2: High Priority Fixes (5 minutes)

#### 2.1 Grace Period Field Consistency?

```bash
# Check for inconsistent usage (should all be gracePeriodSeconds)
git grep -n "gracePeriod[^S]" docs/design/RECONCILE.md | grep -v "gracePeriodExpiry" | head -5

# Count occurrences
INCONSISTENT=$(git grep "gracePeriod[^S]" docs/ | grep -v "gracePeriodSeconds" | grep -v "gracePeriodExpiry" | grep -v "D9_" | wc -l)
[ "$INCONSISTENT" -eq 0 ] && echo "✅ No inconsistent grace period field names" || echo "❌ Found $INCONSISTENT inconsistent uses"
```

**Expected:** All uses are gracePeriodSeconds (spec) or gracePeriodExpiry (status).

---

#### 2.2 gracePeriodExpiry in CRD Status?

```bash
# Check if field added to CRD-SPEC
grep -A 50 "## Status Definition" docs/api/CRD-SPEC.md | grep -q "gracePeriodExpiry" && echo "✅ gracePeriodExpiry in CRD" || echo "❌ gracePeriodExpiry missing"
```

**Expected:** gracePeriodExpiry listed in CRD-SPEC.md status fields.

---

#### 2.3 Pause Semantics Complete?

```bash
# Check RECONCILE.md Step 7 for detailed pause logic
PAUSE_LINES=$(grep -A 20 "Step 7: Determine Write Need" docs/design/RECONCILE.md | wc -l)
[ "$PAUSE_LINES" -gt 30 ] && echo "✅ Pause semantics detailed (>30 lines)" || echo "⚠️ Pause semantics may be incomplete"

# Check for specific keywords
grep -A 30 "Step 7" docs/design/RECONCILE.md | grep -q "Emit event describing what WOULD happen" && echo "✅ Pause event logic present" || echo "❌ Missing pause event logic"
```

**Expected:** Step 7 has comprehensive pause behavior (15+ lines), includes event emission.

---

#### 2.4 MAKE-TARGETS.md Created?

```bash
[ -f docs/MAKE-TARGETS.md ] && echo "✅ MAKE-TARGETS.md exists" || echo "❌ MAKE-TARGETS.md missing"

# Check for reasonable content
[ -f docs/MAKE-TARGETS.md ] && [ $(wc -l < docs/MAKE-TARGETS.md) -gt 50 ] && echo "✅ Has substantial content" || echo "⚠️ May be incomplete"
```

**Expected:** File exists with 50+ lines documenting make targets.

---

#### 2.5 Examples Validate?

```bash
# Requires cluster and CRD (skip if not available)
if kubectl get crd timewindowscalers.kyklos.io 2>/dev/null; then
    for example in examples/*.yaml; do
        kubectl apply --dry-run=client -f "$example" && echo "✅ $example validates" || echo "❌ $example has errors"
    done
else
    echo "⚠️ Skipping example validation (no CRD installed)"
fi
```

**Expected:** All examples validate without errors.

---

### Phase 3: Documentation Quality (5 minutes)

#### 3.1 No Broken Links?

```bash
# Find all markdown links
echo "Checking documentation links..."
find docs -name "*.md" -exec grep -Ho '\[.*\](.*\.md)' {} \; | grep -v "D9_" | sort -u > /tmp/kyklos-links.txt

# Check a few critical ones manually
grep -q "user/MINUTE-DEMO.md" /tmp/kyklos-links.txt && echo "✅ Links to user/MINUTE-DEMO.md" || echo "⚠️ No links to MINUTE-DEMO (may be okay)"
grep -q "MAKE-TARGETS.md" /tmp/kyklos-links.txt && echo "✅ Links to MAKE-TARGETS.md" || echo "⚠️ No links to MAKE-TARGETS (check LOCAL-DEV-GUIDE)"

# Manual verification needed for:
echo "Manual check needed: Open docs/LOCAL-DEV-GUIDE.md and verify line 764 links work"
```

**Expected:** No obviously broken links. Manual spot-check passes.

---

#### 3.2 Terminology Consistency Spot Check?

```bash
# Random sampling of key terms
echo "Checking terminology consistency..."

# windows[].replicas should appear
grep -q "windows\[\]\.replicas" docs/api/CRD-SPEC.md && echo "✅ windows[].replicas in CRD" || echo "❌ Missing windows[].replicas"

# defaultReplicas should appear
grep -q "defaultReplicas" docs/api/CRD-SPEC.md && echo "✅ defaultReplicas in CRD" || echo "❌ Missing defaultReplicas"

# effectiveReplicas should appear in status
grep -A 30 "Status Definition" docs/api/CRD-SPEC.md | grep -q "effectiveReplicas" && echo "✅ effectiveReplicas in status" || echo "❌ Missing effectiveReplicas"
```

**Expected:** All three key replica terms present in correct locations.

---

#### 3.3 Version Information Centralized?

```bash
# Check if version section added to BRIEF
grep -A 10 "Version Requirements" docs/BRIEF.md | grep -q "Kubernetes" && echo "✅ Version requirements in BRIEF" || echo "⚠️ Version info may be scattered"
```

**Expected:** BRIEF.md has version requirements section.

---

#### 3.4 ADR-0005 Exists and is Accepted?

```bash
# Check ADR-0005 status
grep "ADR-0005" docs/DECISIONS.md | grep -q "Accepted" && echo "✅ ADR-0005 Accepted" || echo "❌ ADR-0005 not accepted yet"
```

**Expected:** ADR-0005 has status "Accepted" with clear decision on holidays.

---

## Full Consistency Re-Check

After all fixes, re-run consistency matrix:

```bash
echo "=== CONSISTENCY SCORE ==="
echo "Run these manual checks from D9_CONSISTENCY_MATRIX.md:"
echo ""
echo "1. Feature Scope Consistency (Row 1: Holidays)"
echo "   - Check BRIEF, CRD-SPEC, RECONCILE, CONCEPTS all agree"
echo ""
echo "2. Field Name Consistency (Table: Field Name Consistency)"
echo "   - Check no obsolete terms exist"
echo "   - Check all field names match across docs"
echo ""
echo "3. Status Conditions (Table: Status Condition Consistency)"
echo "   - Should be green (was already consistent)"
echo ""
echo "4. Test Coverage (Table: Test Coverage Consistency)"
echo "   - Check DST fixtures created"
echo "   - Check test scenarios added to plans"
```

---

## Quality Gate Status Check

Run through D9_OVERVIEW.md quality gate table:

```bash
echo "=== QUALITY GATE STATUS ==="
echo ""
echo "Gate 1 (CRD): Check D9_REDLINE_NOTES fixes 1-8 applied"
echo "Gate 2 (Validation): Check ADR-0006 decision made"
echo "Gate 3 (Reconcile): Check pause semantics complete"
echo "Gate 7 (Testing): Check DST fixtures exist"
echo "Gate 8 (CI/CD): Check workflows exist"
echo ""
echo "Manual verification required for each gate."
```

---

## Issue Board Status Check

```bash
# Count open critical issues
echo "=== ISSUE BOARD STATUS ==="
if [ -f docs/D9_ISSUES_BOARD.csv ]; then
    CRITICAL_OPEN=$(grep "P0-Critical,Open" docs/D9_ISSUES_BOARD.csv | wc -l)
    echo "Critical issues open: $CRITICAL_OPEN (expect 0 after fixes)"

    HIGH_OPEN=$(grep "P1-High,Open" docs/D9_ISSUES_BOARD.csv | wc -l)
    echo "High priority issues open: $HIGH_OPEN"
else
    echo "⚠️ D9_ISSUES_BOARD.csv not found"
fi
```

**Expected:** 0 critical issues open before scope lock.

---

## Sign-Off Checklist

Before marking Day 9 review complete:

```bash
cat << 'EOF'
=== SIGN-OFF CHECKLIST ===

Critical Fixes (Must All Pass):
[ ] ADR-0005 holiday decision made and documented
[ ] DST test fixtures created (3 files)
[ ] GitHub CI workflow created
[ ] Glossary updated (no obsolete terms)
[ ] Cross-namespace validation fixed

High Priority Fixes (Must All Pass):
[ ] Grace period fields consistent (gracePeriodSeconds, gracePeriodExpiry)
[ ] Pause semantics complete in RECONCILE.md
[ ] MAKE-TARGETS.md created
[ ] All examples validate with kubectl
[ ] Terminology standardized across docs

Medium Priority (Should Pass):
[ ] Documentation links fixed
[ ] Version section in BRIEF.md
[ ] Test scenarios added for DST
[ ] Validation strategy decided (ADR-0006)

Documentation Quality:
[ ] No broken internal links
[ ] No conflicting information
[ ] Consistent terminology
[ ] Complete status field definitions

Manual Verification:
[ ] Opened 3 random docs, read for consistency
[ ] Tried kubectl apply on 2 examples
[ ] Searched for "activeReplicas" - found none
[ ] Checked DECISIONS.md has 7 ADRs (0001-0007)

Final Check:
[ ] Re-read D9_OVERVIEW.md executive summary
[ ] Confidence level: ___ % (must be >90% for sign-off)
[ ] Ready for scope lock: YES / NO
[ ] Estimated remaining fix time: ___ hours

Signed: ___________________
Date: ___________________
EOF
```

---

## Quick Command Reference

```bash
# Search for term across all docs
git grep "TERM" docs/ | grep -v "D9_"

# Count occurrences
git grep "TERM" docs/ | wc -l

# Find files mentioning term
git grep -l "TERM" docs/

# Check file exists
[ -f path/to/file ] && echo "✅ Exists" || echo "❌ Missing"

# Validate YAML
kubectl apply --dry-run=client -f file.yaml

# Check ADR status
grep "ADR-00XX" docs/DECISIONS.md

# List all examples
ls -la examples/*.yaml

# Count lines in file
wc -l file.md

# Check for pattern in section
grep -A 20 "Section Title" file.md | grep "pattern"
```

---

## Verification Results Template

```markdown
## Verification Results - [DATE]

**Verifier:** [NAME]
**Duration:** [MINUTES]

### Phase 1: Critical Fixes
- Holiday scope: ✅/❌
- DST fixtures: ✅/❌
- CI workflows: ✅/❌
- Glossary: ✅/❌
- Cross-namespace: ✅/❌

### Phase 2: High Priority
- Grace fields: ✅/❌
- Pause semantics: ✅/❌
- MAKE-TARGETS: ✅/❌
- Examples: ✅/❌

### Phase 3: Documentation Quality
- Links: ✅/❌
- Terminology: ✅/❌
- Versions: ✅/❌
- ADRs: ✅/❌

### Overall Status
- Critical issues remaining: [NUMBER]
- High priority issues remaining: [NUMBER]
- Ready for scope lock: YES/NO
- Confidence: [PERCENTAGE]%

### Blockers
[List any remaining blockers]

### Recommendations
[Any recommendations before proceeding]
```

---

## Automation (Future Enhancement)

```bash
# Save this as verify.sh for future use
#!/bin/bash
# Kyklos Documentation Verification Script
# Usage: ./verify.sh [--full|--quick|--critical-only]

# Set this script to executable:
# chmod +x verify.sh

# Run all checks and generate report
# Future: Parse D9_ISSUES_BOARD.csv and automate status checks
```

---

## When to Run This Check

1. **After Critical Fix Sprint** (Oct 29 EOD) - Verify critical issues resolved
2. **After High Priority Sprint** (Oct 30 EOD) - Verify high priority issues resolved
3. **Before Scope Lock** (Nov 2 morning) - Final verification
4. **Before Sign-Off** (Nov 3 morning) - Ultimate verification
5. **After Any Major Doc Update** - Spot check for regressions

---

## Expected Results

### After Critical Fix Sprint (Oct 29)
- Phase 1: 5/5 passing
- Phase 2: 2-3/5 passing (in progress)
- Phase 3: 2-3/4 passing (in progress)
- **Ready for Day 10 work:** YES

### After High Priority Sprint (Oct 30)
- Phase 1: 5/5 passing
- Phase 2: 5/5 passing
- Phase 3: 3-4/4 passing
- **Ready for medium priority:** YES

### Before Scope Lock (Nov 2)
- Phase 1: 5/5 passing
- Phase 2: 5/5 passing
- Phase 3: 4/4 passing
- **Ready for scope lock:** YES

---

**Prepared by:** kyklos-tws-reviewer
**Date:** 2025-10-29 18:00 IST
**Version:** 1.0
**Next Update:** Add automation scripts after Day 10
