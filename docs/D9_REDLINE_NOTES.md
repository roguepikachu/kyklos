# Day 9 Review: Redline Notes - Specific Text Changes

**Review Date:** 2025-10-29
**Reviewer:** kyklos-tws-reviewer
**Purpose:** Exact text changes required to fix consistency issues
**Format:** File → Line → Change

---

## Instructions

For each file listed below:
1. Open the file
2. Navigate to the specified line number
3. Apply the exact change shown (DELETE → ADD or REPLACE WITH)
4. Save and commit

Use `git grep` to find additional occurrences if patterns repeat.

---

## CRITICAL CHANGES - Must Fix Before Implementation

### 1. BRIEF.md - Holiday Scope Decision

**DECISION REQUIRED FIRST:** Is holiday support in v0.1 or v0.2?

**Option A: Holidays ARE in v0.1** (Recommended - already fully designed)

```diff
File: docs/BRIEF.md
Line: 17

- - Calendar integration or holiday awareness
+ - Advanced calendar features (recurring patterns, external calendar sync)
```

**Option B: Holidays NOT in v0.1** (Requires removing from many docs)

If choosing Option B, follow instructions in Section 10 below to remove all holiday references.

---

### 2. BRIEF.md - Fix Glossary Terms

```diff
File: docs/BRIEF.md
Lines: 72-75

DELETE:
- **activeReplicas**: Desired replica count during active window
-
- **inactiveReplicas**: Desired replica count during inactive window (often 0)

ADD:
+ **windows[].replicas**: Desired replica count when this window is active
+
+ **defaultReplicas**: Replica count when no windows match (often 2, not 0 for availability)
+
+ **effectiveReplicas**: The computed replica count right now (shown in status)
+
+ **pause**: When true, controller computes state but doesn't modify target workload
```

---

### 3. CRD-SPEC.md - Fix Cross-Namespace Validation

```diff
File: docs/api/CRD-SPEC.md
Lines: 26-28

DELETE:
- **Validation Rules**:
- - `kind` must equal `Deployment` (enforced by admission webhook)
- - `name` must be non-empty
- - If `namespace` is specified, it must equal the TimeWindowScaler's namespace

REPLACE WITH:
+ **Validation Rules**:
+ - `kind` must equal `Deployment` in v1alpha1 (enforced by CRD validation)
+ - `name` must be non-empty (enforced by Kubernetes)
+ - `namespace` may differ from TimeWindowScaler namespace (cross-namespace requires ClusterRole)
```

---

### 4. CRD-SPEC.md - Add Missing Status Field

```diff
File: docs/api/CRD-SPEC.md
Line: 125 (after lastScaleTime)

ADD NEW SECTION:
+ ### status.gracePeriodExpiry
+ | Field | Type | Description |
+ |-------|------|-------------|
+ | `gracePeriodExpiry` | string | RFC3339 timestamp when grace period expires |
+
+ **Semantics**: Set when entering grace period, cleared when grace expires or cancelled.
```

---

### 5. RECONCILE.md - Fix Grace Period Field Name

```diff
File: docs/design/RECONCILE.md
Line: 73

- - If !status.gracePeriodExpiry: set expiry = now + gracePeriodSeconds
+ - If !status.gracePeriodExpiry: set expiry = now + spec.gracePeriodSeconds
```

Also search for all occurrences of `gracePeriod` (without Seconds) in RECONCILE.md and replace with `gracePeriodSeconds`.

---

### 6. RECONCILE.md - Add Pause Semantics Detail

```diff
File: docs/design/RECONCILE.md
Lines: 89-95

REPLACE SECTION:
- ### Step 7: Determine Write Need
- **Preconditions**: effectiveReplicas computed, target status known
- **Actions**:
- 1. If spec.pause==true: skip write, set Ready based on alignment
- 2. If targetSpecReplicas != effectiveReplicas: write needed
- 3. If manual drift detected (observedReplicas != targetSpecReplicas != effectiveReplicas): write needed

WITH DETAILED VERSION:
+ ### Step 7: Determine Write Need
+ **Preconditions**: effectiveReplicas computed, target status known
+ **Actions**:
+ 1. **If spec.pause==true**:
+    - Skip all writes to target
+    - Continue computing effectiveReplicas normally
+    - Update all status fields (effectiveReplicas, targetObservedReplicas, etc.)
+    - Set Ready condition:
+      - Ready=True if targetObservedReplicas == effectiveReplicas
+      - Ready=False with reason=TargetMismatch if different
+    - Emit event describing what WOULD happen if not paused
+    - **Return early, do not proceed to Step 8**
+ 2. If targetSpecReplicas != effectiveReplicas: write needed
+ 3. If manual drift detected: write needed
```

---

### 7. LOCAL-DEV-GUIDE.md - Fix Broken Link

```diff
File: docs/LOCAL-DEV-GUIDE.md
Line: 764

- - Follow [MINUTE-DEMO.md](./MINUTE-DEMO.md) for a 10-minute walkthrough
+ - Follow [MINUTE-DEMO.md](./user/MINUTE-DEMO.md) for a 10-minute walkthrough
```

---

### 8. CRD-SPEC.md - Clarify Validation Method

```diff
File: docs/api/CRD-SPEC.md
Line: 26

- - `kind` must equal `Deployment` (enforced by admission webhook)
+ - `kind` must equal `Deployment` (enforced by CRD enum validation)
```

Note: If validation webhook is later added, update this. For v0.1, use CRD validation only.

---

## HIGH PRIORITY CHANGES - Fix During Sprint

### 9. CONCEPTS.md - Add Holiday Scope Note

**Only if holidays ARE in v0.1:**

```diff
File: docs/user/CONCEPTS.md
Line: 227 (before ## Holiday Handling)

ADD NOTE:
+ > **Note:** Holiday support is available in v0.1 with ConfigMap-based sources. External calendar sync is planned for v0.2.
```

**If holidays NOT in v0.1, add different note:**

```diff
File: docs/user/CONCEPTS.md
Line: 227

ADD NOTE:
+ > **Note:** Holiday support is coming in v0.2. This section describes future functionality for planning purposes.
```

---

### 10. Remove Holiday References (If NOT in v0.1)

**Only execute if decision is "Holidays NOT in v0.1":**

#### 10a. CRD-SPEC.md

```diff
File: docs/api/CRD-SPEC.md
Lines: 66-80

DELETE ENTIRE SECTION:
- ### spec.holidays (optional)
- [entire section through line 80]
```

#### 10b. RECONCILE.md

```diff
File: docs/design/RECONCILE.md
Lines: 33-41

DELETE STEP 3:
- ### Step 3: Check Holiday Status (if configured)
- [entire section]
```

And update Step 4:

```diff
Lines: 43-44

- ### Step 4: Compute Effective Replicas
- **Preconditions**: Local time available, holiday status determined

+ ### Step 3: Compute Effective Replicas (renumber from Step 4)
+ **Preconditions**: Local time available
```

#### 10c. CONCEPTS.md

```diff
File: docs/user/CONCEPTS.md
Lines: 226-307

DELETE ENTIRE SECTION:
- ## Holiday Handling
- [entire section through line 307]
```

#### 10d. Remove Example File

```bash
git rm examples/tws-holidays-closed.yaml
```

#### 10e. Remove from Decision Table

```diff
File: docs/design/RECONCILE.md
Lines: 60-67

DELETE ROWS WITH HOLIDAY:
[Remove all rows starting with "true" in Holiday column]
```

---

### 11. Create MAKE-TARGETS.md

```bash
# Create new file
cat > docs/MAKE-TARGETS.md << 'EOF'
# Makefile Targets Reference

## Setup and Verification
- `make tools` - Install development tools (controller-gen, golangci-lint)
- `make verify-tools` - Check prerequisites are installed
- `make verify-all` - Complete system verification

## Cluster Management
- `make cluster-up` - Create Kind cluster
- `make cluster-down` - Delete Kind cluster
- `make cluster-up-k3d` - Create k3d cluster (alternative)

## Build
- `make build` - Build controller binary
- `make docker-build` - Build container image
- `make manifests` - Generate CRD manifests
- `make generate` - Generate deepcopy code

## Deploy
- `make install-crds` - Install CRDs to cluster
- `make deploy` - Deploy controller
- `make kind-load` - Load image into Kind cluster
- `make undeploy` - Remove controller
- `make uninstall-crds` - Remove CRDs

## Testing
- `make test` - Run unit tests
- `make test-integration` - Run envtest integration tests
- `make test-e2e` - Run end-to-end tests
- `make test-coverage` - Generate coverage report

## Development
- `make run-local` - Run controller locally
- `make logs-controller` - Show controller logs
- `make logs-controller-follow` - Follow controller logs
- `make restart-controller` - Restart controller pod

## Demo
- `make demo-setup` - Create demo namespace and deployment
- `make demo-apply-minute` - Apply minute-scale demo TWS
- `make demo-watch` - Watch demo resources
- `make demo-cleanup` - Clean up demo resources

## Verification
- `make verify-controller` - Check controller health
- `make verify-demo` - Verify demo status

## Cleanup
- `make clean` - Remove build artifacts
- `make clean-all` - Complete cleanup including cluster
- `make reset-env` - Full reset and rebuild

See LOCAL-DEV-GUIDE.md for usage examples.
EOF
```

---

### 12. Fix CRD-SPEC Holiday Mode Default

**Only if holidays ARE in v0.1:**

```diff
File: docs/api/CRD-SPEC.md
Line: 75

- - `ignore`: Process windows normally on holidays
+ - `ignore` (default): Process windows normally on holidays, no special handling
```

---

## MEDIUM PRIORITY CHANGES - Before Release

### 13. Add Version Section to BRIEF.md

```diff
File: docs/BRIEF.md
Line: 3 (after status line)

ADD NEW SECTION:
+
+ ## Version Requirements
+ - **Project Version:** 0.1.0 (alpha)
+ - **API Version:** kyklos.io/v1alpha1
+ - **Kubernetes:** 1.25+ (tested on 1.28)
+ - **Go:** 1.21+ for building controller
+ - **Docker:** 24.0+ for building images
```

---

### 14. Add effectiveReplicas to Glossary

```diff
File: docs/BRIEF.md
Line: 75 (after inactiveReplicas, which should be replaced per #2)

ADD:
+ **effectiveReplicas**: The replica count computed by controller for right now, shown in status
```

---

### 15. CONCEPTS.md - Clarify Terminology

```diff
File: docs/user/CONCEPTS.md
Lines: 87-89

REPLACE SECTION TITLE AND INTRO:
- ## Effective Replicas
-
- The effective replica count is the number of replicas Kyklos wants your deployment to have right now.

WITH:
+ ## Effective Replicas (Current Desired State)
+
+ The **effectiveReplicas** field in status shows the number of replicas Kyklos has computed as correct for right now, based on current time and window matching. This is the replica count Kyklos will enforce on the target deployment.
+
+ **Terminology clarification:**
+ - `windows[].replicas` - Configured in spec, what you want during each window
+ - `defaultReplicas` - Configured in spec, what you want when no windows match
+ - `effectiveReplicas` - Computed in status, what controller wants RIGHT NOW
+ - `targetObservedReplicas` - Observed in status, what the deployment actually has
```

---

### 16. README.md - Add Test Verification

```diff
File: README.md
Line: 48

ADD AFTER CONTROLLER VERIFICATION:
+
+ 4. **Run smoke test** (optional but recommended)
+ ```bash
+ make test-unit
+ ```
```

---

### 17. Fix Example if Holidays NOT in v0.1

**Only if holidays NOT in v0.1:**

```bash
# Move holiday example to future directory
mkdir -p examples/future
git mv examples/tws-holidays-closed.yaml examples/future/
```

Add README to examples/future/:

```bash
cat > examples/future/README.md << 'EOF'
# Future Features Examples

These examples demonstrate features planned for v0.2 and later.
They are provided for planning purposes but will not work with v0.1.

- `tws-holidays-closed.yaml` - Holiday handling (v0.2)
EOF
```

---

### 18. CRD-SPEC - Add Pause to Status

**If pause affects status fields:**

```diff
File: docs/api/CRD-SPEC.md
Line: 125 (in status section, after observedGeneration)

ADD NEW STATUS FIELD:
+ ### status.paused
+ | Field | Type | Description |
+ |-------|------|-------------|
+ | `paused` | bool | Reflects spec.pause value |
+
+ **Semantics**: Convenience field mirroring spec.pause for quick status checks.
```

Note: This is optional optimization. Status can be checked via spec.pause directly.

---

## DOCUMENTATION LINK FIXES

### 19. Fix All Documentation Cross-References

Run these commands to verify and fix links:

```bash
# Find all markdown links
find docs -name "*.md" -exec grep -H '\[.*\](.*\.md)' {} \;

# Verify each link target exists
# Fix these specific broken links:
```

**Specific fixes:**

```diff
File: docs/LOCAL-DEV-GUIDE.md
Line: 777

- - See [MAKE-TARGETS.md](./MAKE-TARGETS.md) for complete target reference
+ - See [MAKE-TARGETS.md](MAKE-TARGETS.md) for complete target reference
```

(After creating MAKE-TARGETS.md per #11)

---

## VALIDATION

### 20. Validate All Example Files

Run these commands:

```bash
# Dry-run validation for each example
kubectl apply --dry-run=client -f examples/tws-office-hours.yaml
kubectl apply --dry-run=client -f examples/tws-night-shift.yaml

# If holidays in v0.1:
kubectl apply --dry-run=client -f examples/tws-holidays-closed.yaml

# Fix any validation errors before sign-off
```

Expected output: No errors, just "created (dry run)"

---

## TESTING FIXTURES

### 21. Create DST Test Fixtures

```bash
# Create test fixtures directory
mkdir -p test/fixtures

# Spring Forward Fixture
cat > test/fixtures/dst-spring-2025.yaml << 'EOF'
# DST Spring Forward Test Case
# Date: 2025-03-09 (Second Sunday of March)
# Timezone: America/New_York
# Transition: 02:00 AM EST → 03:00 AM EDT (clock jumps forward 1 hour)
# Test: Window spanning 01:00-04:00 should be 2 hours (01:00-01:59, then 03:00-03:59)

apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: dst-spring-test
  namespace: test
spec:
  targetRef:
    kind: Deployment
    name: test-deployment
  timezone: America/New_York
  defaultReplicas: 1
  windows:
  - days: [Sun]
    start: "01:00"
    end: "04:00"
    replicas: 5
  # This window will be shortened by 1 hour on 2025-03-09
  # Verify: Window active 01:00-01:59, skips 02:00-02:59, active 03:00-03:59
EOF

# Fall Back Fixture
cat > test/fixtures/dst-fall-2025.yaml << 'EOF'
# DST Fall Back Test Case
# Date: 2025-11-02 (First Sunday of November)
# Timezone: America/New_York
# Transition: 02:00 AM EDT → 01:00 AM EST (clock falls back 1 hour)
# Test: Window spanning 01:00-04:00 should be 4 hours (includes 01:00-01:59 twice)

apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: dst-fall-test
  namespace: test
spec:
  targetRef:
    kind: Deployment
    name: test-deployment
  timezone: America/New_York
  defaultReplicas: 1
  windows:
  - days: [Sun]
    start: "01:00"
    end: "04:00"
    replicas: 5
  # This window will be extended by 1 hour on 2025-11-02
  # Verify: Window active 01:00-01:59 (first), 01:00-01:59 (second), 02:00-02:59, 03:00-03:59
EOF

# Cross-Midnight + DST Fixture
cat > test/fixtures/dst-cross-midnight-2025.yaml << 'EOF'
# DST + Cross-Midnight Test Case
# Date: 2025-03-08 to 2025-03-09 (Saturday night to Sunday morning)
# Timezone: America/New_York
# Window: 22:00 Saturday to 06:00 Sunday
# DST transition at 02:00 Sunday
# Test: Window should remain active across midnight and DST transition

apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: dst-cross-midnight-test
  namespace: test
spec:
  targetRef:
    kind: Deployment
    name: test-deployment
  timezone: America/New_York
  defaultReplicas: 1
  windows:
  - days: [Sat]
    start: "22:00"
    end: "06:00"
    replicas: 5
  # Saturday 22:00 - Sunday 05:59
  # DST spring forward at 02:00 Sunday
  # Verify: Window active Saturday 22:00-23:59, Sunday 00:00-01:59, 03:00-05:59
  # Hour 02:00-02:59 does not exist
EOF
```

---

## GITHUB WORKFLOWS

### 22. Create Basic CI Workflow

```bash
mkdir -p .github/workflows

cat > .github/workflows/ci.yml << 'EOF'
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Run lint
      run: make lint

  test-unit:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Run unit tests
      run: make test
    - name: Upload coverage
      uses: actions/upload-artifact@v3
      with:
        name: coverage
        path: coverage.out

  test-envtest:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Run envtest
      run: make test-integration

  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Build controller
      run: make build

  smoke-test:
    runs-on: ubuntu-latest
    needs: [lint, test-unit, test-envtest, build]
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Create kind cluster
      uses: helm/kind-action@v1.5.0
    - name: Build and load image
      run: |
        make docker-build kind-load
    - name: Deploy controller
      run: |
        make install-crds deploy
    - name: Run smoke test
      run: |
        make demo-setup demo-apply-minute
        sleep 300  # Wait for 5 minutes
        make verify-demo
EOF
```

---

## VERIFICATION CHECKLIST

After applying all changes, verify:

```bash
# 1. Glossary updated
grep -A 20 "## Glossary" docs/BRIEF.md

# 2. No activeReplicas/inactiveReplicas in docs
git grep "activeReplicas" docs/
git grep "inactiveReplicas" docs/
# Should find NONE (except in this redline document)

# 3. All examples validate
kubectl apply --dry-run=client -f examples/*.yaml

# 4. All links work
find docs -name "*.md" -exec grep -H '\[.*\](.*\.md)' {} \; | while read line; do
  # Manual verification needed
  echo "$line"
done

# 5. Grace period field consistent
git grep "gracePeriod[^S]" docs/
# Should find NONE (all should be gracePeriodSeconds)

# 6. Test fixtures exist
ls -la test/fixtures/dst-*.yaml
# Should show 3 files

# 7. Workflows exist
ls -la .github/workflows/*.yml
# Should show at least ci.yml
```

---

## NOTES

1. **Holiday Decision Impact:** If holidays are removed from v0.1, approximately 15 documentation sections must be updated or removed. This is a 2-hour task.

2. **Field Name Changes:** Use global find-replace carefully. Some occurrences may be in code comments or examples that need different handling.

3. **Test Fixtures:** The DST fixtures use specific 2025 dates. Update dates if testing in a different year.

4. **Workflow:** The basic CI workflow provided is minimal. Full workflow per PIPELINE.md requires additional jobs and caching configuration.

5. **Cross-Reference Validation:** Some links are intentionally external (to tools, Kubernetes docs). Only fix internal doc links.

---

**Prepared by:** kyklos-tws-reviewer
**Date:** 2025-10-29 16:30 IST
**Estimated Application Time:** 3-4 hours for all changes
**Priority Order:** Critical (1-8) → High (9-18) → Medium (19-22)
