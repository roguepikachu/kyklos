# Day 10 Edit Pack: Exact Text Replacements

**Date:** 2025-10-30
**Coordinator:** kyklos-orchestrator
**Format:** Surgical edits under 15 lines each
**Status:** Ready for application

---

## How to Use This Edit Pack

For each edit:
1. Open the file at the specified path
2. Locate the OLD TEXT (exact match required)
3. Replace with NEW TEXT (exact as specified)
4. Verify the change makes sense in context
5. Mark as completed in D10_ASSIGNMENTS.csv

**Important:** Apply edits in the order specified in D10_MERGE_PLAN.md to avoid conflicts.

---

## CRITICAL DECISION REQUIRED FIRST

Before applying any edits, decide on ADR-0005: Holiday Support in v0.1

**If HOLIDAYS IN v0.1:** Apply edits marked [OPTION A]
**If HOLIDAYS NOT IN v0.1:** Apply edits marked [OPTION B]

---

## Edit Group 1: Glossary and Terminology (CRITICAL)

### EDIT-001: BRIEF.md - Update Glossary Terms
**File:** `/Users/aykumar/personal/kyklos/docs/BRIEF.md`
**Owner:** kyklos-orchestrator
**Deadline:** Oct 30 12:00 IST
**Rationale:** Align glossary with actual API field names

**OLD TEXT (lines 72-81):**
```
**activeReplicas**: Desired replica count during active window

**inactiveReplicas**: Desired replica count during inactive window (often 0)

**crossMidnight**: Window spanning two calendar days (e.g., 22:00-02:00)

**windowStart**: HH:MM time when active window begins

**windowEnd**: HH:MM time when active window ends
```

**NEW TEXT:**
```
**windows[].replicas**: Desired replica count when this window is active (configured in spec)

**defaultReplicas**: Replica count when no windows match (often 2 for availability, not 0)

**effectiveReplicas**: The computed replica count right now (shown in status)

**pause**: When true, controller computes state but doesn't modify target workload

**crossMidnight**: Window spanning two calendar days (e.g., 22:00-02:00)

**windowStart**: HH:MM time when active window begins

**windowEnd**: HH:MM time when active window ends
```

---

### EDIT-002: BRIEF.md - Add Version Requirements
**File:** `/Users/aykumar/personal/kyklos/docs/BRIEF.md`
**Owner:** kyklos-orchestrator
**Deadline:** Oct 30 12:00 IST
**Rationale:** Centralize version information

**INSERT AFTER line 3 (after Status line):**
```

## Version Requirements
- **Project Version:** 0.1.0 (alpha)
- **API Version:** kyklos.io/v1alpha1
- **Kubernetes:** 1.25+ (tested on 1.28)
- **Go:** 1.21+ for building controller
- **Docker:** 24.0+ for building images
```

---

### EDIT-003-A: BRIEF.md - Holiday Scope (If IN v0.1)
**File:** `/Users/aykumar/personal/kyklos/docs/BRIEF.md`
**Owner:** kyklos-orchestrator
**Deadline:** Oct 30 12:00 IST
**Rationale:** Clarify holiday support is in v0.1
**Apply if:** ADR-0005 decides holidays IN v0.1

**OLD TEXT (line 17):**
```
- Calendar integration or holiday awareness
```

**NEW TEXT:**
```
- Advanced calendar features (recurring patterns, external calendar sync beyond ConfigMap)
```

---

### EDIT-003-B: BRIEF.md - Holiday Scope (If NOT in v0.1)
**File:** `/Users/aykumar/personal/kyklos/docs/BRIEF.md`
**Owner:** kyklos-orchestrator
**Deadline:** Oct 30 12:00 IST
**Rationale:** Keep holiday as non-goal
**Apply if:** ADR-0005 decides holidays NOT in v0.1

**NO CHANGE NEEDED** - Line 17 is already correct

---

## Edit Group 2: CRD Specification (CRITICAL)

### EDIT-004: CRD-SPEC.md - Fix Validation Method
**File:** `/Users/aykumar/personal/kyklos/docs/api/CRD-SPEC.md`
**Owner:** api-crd-designer
**Deadline:** Oct 30 13:00 IST
**Rationale:** Clarify validation strategy (per ADR-0006)

**OLD TEXT (lines 25-28):**
```
**Validation Rules**:
- `kind` must equal `Deployment` (enforced by admission webhook)
- `name` must be non-empty
- If `namespace` is specified, it must equal the TimeWindowScaler's namespace
```

**NEW TEXT:**
```
**Validation Rules**:
- `kind` must equal `Deployment` in v1alpha1 (enforced by CRD enum validation)
- `name` must be non-empty (enforced by Kubernetes)
- `namespace` may differ from TimeWindowScaler namespace (cross-namespace requires ClusterRole, see ADR-0002)
```

---

### EDIT-005: CRD-SPEC.md - Add Grace Period Expiry Field
**File:** `/Users/aykumar/personal/kyklos/docs/api/CRD-SPEC.md`
**Owner:** api-crd-designer
**Deadline:** Oct 30 13:00 IST
**Rationale:** Add missing status field used in reconcile logic

**INSERT AFTER line 125 (in status section, after lastScaleTime):**
```

### status.gracePeriodExpiry
| Field | Type | Description |
|-------|------|-------------|
| `gracePeriodExpiry` | string | RFC3339 timestamp when grace period expires (empty if not in grace) |

**Semantics:**
- Set to `now + spec.gracePeriodSeconds` when entering grace period
- Cleared when grace expires or is cancelled by new window activation
- Controller uses this to determine if still within grace period across restarts
```

---

### EDIT-006-A: CRD-SPEC.md - Holiday Section (If IN v0.1)
**File:** `/Users/aykumar/personal/kyklos/docs/api/CRD-SPEC.md`
**Owner:** api-crd-designer
**Deadline:** Oct 30 13:00 IST
**Rationale:** Clarify default mode
**Apply if:** ADR-0005 decides holidays IN v0.1

**OLD TEXT (line 75):**
```
- `ignore`: Process windows normally on holidays
```

**NEW TEXT:**
```
- `ignore` (default): Process windows normally on holidays, no special handling
```

---

### EDIT-006-B: CRD-SPEC.md - Holiday Section (If NOT in v0.1)
**File:** `/Users/aykumar/personal/kyklos/docs/api/CRD-SPEC.md`
**Owner:** api-crd-designer
**Deadline:** Oct 30 13:00 IST
**Rationale:** Remove holiday section entirely
**Apply if:** ADR-0005 decides holidays NOT in v0.1

**DELETE ENTIRE SECTION (lines 66-80):**
```
### spec.holidays (optional)
[Delete from line 66 through line 80]
```

---

## Edit Group 3: Reconcile Logic (CRITICAL)

### EDIT-007: RECONCILE.md - Fix Grace Period Field Name
**File:** `/Users/aykumar/personal/kyklos/docs/design/RECONCILE.md`
**Owner:** controller-reconcile-designer
**Deadline:** Oct 30 13:00 IST
**Rationale:** Consistent field naming

**OLD TEXT (line 73):**
```
2. If effectiveReplicas < status.effectiveReplicas && grace > 0:
```

**NEW TEXT:**
```
2. If effectiveReplicas < status.effectiveReplicas && spec.gracePeriodSeconds > 0:
```

**AND OLD TEXT (line 74):**
```
   - If !status.gracePeriodExpiry: set expiry = now + gracePeriodSeconds
```

**NEW TEXT:**
```
   - If !status.gracePeriodExpiry: set expiry = now + spec.gracePeriodSeconds
```

---

### EDIT-008: RECONCILE.md - Expand Pause Semantics
**File:** `/Users/aykumar/personal/kyklos/docs/design/RECONCILE.md`
**Owner:** controller-reconcile-designer
**Deadline:** Oct 30 13:00 IST
**Rationale:** Complete pause implementation detail

**OLD TEXT (lines 90-97):**
```
### Step 7: Determine Write Need
**Preconditions**: effectiveReplicas computed, target status known
**Actions**:
1. If spec.pause==true: skip write, set Ready based on alignment
2. If targetSpecReplicas != effectiveReplicas: write needed
3. If manual drift detected (observedReplicas != targetSpecReplicas != effectiveReplicas): write needed

**Postconditions**: Write decision made
```

**NEW TEXT:**
```
### Step 7: Determine Write Need
**Preconditions**: effectiveReplicas computed, target status known
**Actions**:
1. **If spec.pause==true**:
   - Skip all writes to target workload
   - Continue computing effectiveReplicas normally (show what WOULD happen)
   - Update all status fields: effectiveReplicas, targetObservedReplicas, currentWindow, gracePeriodExpiry
   - Set Ready condition:
     - Ready=True if targetObservedReplicas == effectiveReplicas (aligned)
     - Ready=False with reason=TargetMismatch if different (drift while paused)
   - Emit ScalingSkipped event with message describing what would happen if not paused
   - **Return early, do not proceed to Step 8**
2. If targetSpecReplicas != effectiveReplicas: write needed
3. If manual drift detected (observedReplicas != targetSpecReplicas != effectiveReplicas): write needed

**Postconditions**: Write decision made or early return if paused
```

---

### EDIT-009-A: RECONCILE.md - Holiday Logic (If IN v0.1)
**File:** `/Users/aykumar/personal/kyklos/docs/design/RECONCILE.md`
**Owner:** controller-reconcile-designer
**Deadline:** Oct 30 13:00 IST
**Rationale:** Keep holiday logic
**Apply if:** ADR-0005 decides holidays IN v0.1

**NO CHANGE NEEDED** - Holiday logic in Step 3 is already correct

---

### EDIT-009-B: RECONCILE.md - Holiday Logic (If NOT in v0.1)
**File:** `/Users/aykumar/personal/kyklos/docs/design/RECONCILE.md`
**Owner:** controller-reconcile-designer
**Deadline:** Oct 30 13:00 IST
**Rationale:** Remove holiday logic entirely
**Apply if:** ADR-0005 decides holidays NOT in v0.1

**DELETE STEP 3 (lines 33-41):**
```
### Step 3: Check Holiday Status (if configured)
[Delete entire step]
```

**AND RENUMBER STEPS:** Step 4 becomes Step 3, etc.

**AND UPDATE Step 4 Preconditions (line 44):**

**OLD TEXT:**
```
**Preconditions**: Local time available, holiday status determined
```

**NEW TEXT:**
```
**Preconditions**: Local time available
```

**AND UPDATE Decision Table (lines 60-67) - Remove all holiday rows**

---

## Edit Group 4: Test Fixtures (CRITICAL)

### EDIT-010: Create DST Spring Forward Fixture
**File:** `/Users/aykumar/personal/kyklos/test/fixtures/dst-spring-2025.yaml`
**Owner:** testing-strategy-designer
**Deadline:** Oct 30 13:00 IST
**Rationale:** Enable DST spring forward testing

**CREATE NEW FILE:**
```yaml
# DST Spring Forward Test Case
# Date: 2025-03-09 (Second Sunday of March)
# Timezone: America/New_York
# Transition: 02:00 AM EST → 03:00 AM EDT (clock jumps forward 1 hour)
# Test: Window spanning 01:00-04:00 should be 2 hours (01:00-01:59, then 03:00-03:59)
# Hour 02:00-02:59 does not exist on this date

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
  # Expected behavior on 2025-03-09:
  # - Window active 01:00-01:59 EST (1 hour)
  # - Clock jumps from 01:59:59 EST to 03:00:00 EDT
  # - Window active 03:00-03:59 EDT (1 hour)
  # - Total active duration: 2 hours instead of 3
```

---

### EDIT-011: Create DST Fall Back Fixture
**File:** `/Users/aykumar/personal/kyklos/test/fixtures/dst-fall-2025.yaml`
**Owner:** testing-strategy-designer
**Deadline:** Oct 30 13:00 IST
**Rationale:** Enable DST fall back testing

**CREATE NEW FILE:**
```yaml
# DST Fall Back Test Case
# Date: 2025-11-02 (First Sunday of November)
# Timezone: America/New_York
# Transition: 02:00 AM EDT → 01:00 AM EST (clock falls back 1 hour)
# Test: Window spanning 01:00-04:00 should be 4 hours (includes 01:00-01:59 twice)
# Hour 01:00-01:59 occurs twice on this date

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
  # Expected behavior on 2025-11-02:
  # - Window active 01:00-01:59 EDT (1 hour, first occurrence)
  # - Clock falls back from 01:59:59 EDT to 01:00:00 EST
  # - Window active 01:00-01:59 EST (1 hour, second occurrence)
  # - Window active 02:00-03:59 EST (2 hours)
  # - Total active duration: 4 hours instead of 3
```

---

### EDIT-012: Create DST Cross-Midnight Fixture
**File:** `/Users/aykumar/personal/kyklos/test/fixtures/dst-cross-midnight-2025.yaml`
**Owner:** testing-strategy-designer
**Deadline:** Oct 30 13:00 IST
**Rationale:** Enable combined DST + cross-midnight testing

**CREATE NEW FILE:**
```yaml
# DST + Cross-Midnight Test Case
# Date: 2025-03-08 to 2025-03-09 (Saturday night to Sunday morning)
# Timezone: America/New_York
# Window: 22:00 Saturday to 06:00 Sunday
# DST transition at 02:00 Sunday (spring forward)
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
  # Expected behavior on 2025-03-08 22:00 to 2025-03-09 06:00:
  # - Window active Saturday 22:00-23:59 EST (2 hours)
  # - Midnight crosses to Sunday
  # - Window active Sunday 00:00-01:59 EST (2 hours)
  # - DST spring forward at 02:00 (clock jumps to 03:00)
  # - Window active Sunday 03:00-05:59 EDT (3 hours)
  # - Total active duration: 7 hours instead of 8 hours
  # - Window ends at 06:00 EDT (equivalent to 05:00 EST)
```

---

## Edit Group 5: CI Workflow (CRITICAL)

### EDIT-013: Create Basic CI Workflow
**File:** `/Users/aykumar/personal/kyklos/.github/workflows/ci.yml`
**Owner:** ci-release-designer
**Deadline:** Oct 30 14:00 IST
**Rationale:** Enable automated testing

**CREATE NEW FILE:**
```yaml
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
      run: make lint || echo "Lint target not yet implemented"

  test-unit:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Run unit tests
      run: make test || echo "Test target not yet implemented"

  verify:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Verify documentation links
      run: |
        echo "Checking for broken internal links..."
        find docs -name "*.md" -exec grep -H '\[.*\](.*\.md)' {} \; || true
        echo "Link verification complete"
```

---

## Edit Group 6: Documentation (HIGH PRIORITY)

### EDIT-014: LOCAL-DEV-GUIDE.md - Fix Broken Link
**File:** `/Users/aykumar/personal/kyklos/docs/LOCAL-DEV-GUIDE.md`
**Owner:** local-workflow-designer
**Deadline:** Oct 30 18:00 IST
**Rationale:** Fix documentation cross-reference

**Search for and replace:**
**OLD TEXT:**
```
[MINUTE-DEMO.md](./MINUTE-DEMO.md)
```

**NEW TEXT:**
```
[MINUTE-DEMO.md](./user/MINUTE-DEMO.md)
```

---

### EDIT-015: Create MAKE-TARGETS.md
**File:** `/Users/aykumar/personal/kyklos/docs/MAKE-TARGETS.md`
**Owner:** local-workflow-designer
**Deadline:** Oct 30 16:00 IST
**Rationale:** Document all Makefile targets

**CREATE NEW FILE:**
```markdown
# Makefile Targets Reference

**Note:** This is the planned target structure for v0.1 implementation. Actual targets will be created during implementation phase.

## Setup and Verification
- `make tools` - Install development tools (controller-gen, golangci-lint, kind)
- `make verify-tools` - Check prerequisites are installed
- `make verify-all` - Complete system verification

## Cluster Management
- `make cluster-up` - Create Kind cluster (default)
- `make cluster-down` - Delete Kind cluster
- `make cluster-up-k3d` - Create k3d cluster (alternative)

## Build
- `make build` - Build controller binary
- `make docker-build` - Build container image
- `make manifests` - Generate CRD manifests from Go types
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
- `make test-e2e` - Run end-to-end tests (requires cluster)
- `make test-coverage` - Generate coverage report

## Development
- `make run-local` - Run controller locally (outside cluster)
- `make logs-controller` - Show controller logs
- `make logs-controller-follow` - Follow controller logs in real-time
- `make restart-controller` - Restart controller pod

## Demo
- `make demo-setup` - Create demo namespace and deployment
- `make demo-apply-minute` - Apply minute-scale demo TWS
- `make demo-watch` - Watch demo resources
- `make demo-cleanup` - Clean up demo resources

## Linting and Verification
- `make lint` - Run golangci-lint
- `make fmt` - Format Go code
- `make vet` - Run go vet
- `make verify-controller` - Check controller health

## Cleanup
- `make clean` - Remove build artifacts
- `make clean-all` - Complete cleanup including cluster
- `make reset-env` - Full reset and rebuild

## Target Dependencies

```
install-crds → manifests
deploy → docker-build, install-crds
demo-apply-minute → demo-setup, deploy
test-e2e → cluster-up, deploy
```

## Usage Examples

**Quick local development:**
```bash
make cluster-up
make docker-build kind-load
make install-crds deploy
make demo-setup demo-apply-minute demo-watch
```

**Run tests:**
```bash
make test
make test-integration
```

**Full CI simulation:**
```bash
make lint
make test
make test-integration
make build
make docker-build
```

See [LOCAL-DEV-GUIDE.md](LOCAL-DEV-GUIDE.md) for detailed workflows.
```

---

### EDIT-016: CONCEPTS.md - Clarify Terminology
**File:** `/Users/aykumar/personal/kyklos/docs/user/CONCEPTS.md`
**Owner:** docs-dx-designer
**Deadline:** Oct 30 17:00 IST
**Rationale:** Explain effective vs configured replicas

**OLD TEXT (lines 87-89):**
```
## Effective Replicas

The effective replica count is the number of replicas Kyklos wants your deployment to have right now.
```

**NEW TEXT:**
```
## Effective Replicas (Current Desired State)

The **effectiveReplicas** field in status shows the number of replicas Kyklos has computed as correct for right now, based on current time and window matching. This is the replica count Kyklos will write to the target deployment.

**Terminology clarification:**
- `windows[].replicas` - Configured in spec, what you want during each window
- `defaultReplicas` - Configured in spec, what you want when no windows match
- `effectiveReplicas` - Computed in status, what controller wants RIGHT NOW
- `targetObservedReplicas` - Observed in status, what the deployment actually has
```

---

### EDIT-017-A: CONCEPTS.md - Holiday Note (If IN v0.1)
**File:** `/Users/aykumar/personal/kyklos/docs/user/CONCEPTS.md`
**Owner:** docs-dx-designer
**Deadline:** Oct 30 17:00 IST
**Rationale:** Clarify holiday support status
**Apply if:** ADR-0005 decides holidays IN v0.1

**INSERT BEFORE line 227 (before ## Holiday Handling):**
```

> **Note:** Holiday support is available in v0.1 with ConfigMap-based sources. External calendar sync and advanced recurring patterns are planned for v0.2.

```

---

### EDIT-017-B: CONCEPTS.md - Holiday Note (If NOT in v0.1)
**File:** `/Users/aykumar/personal/kyklos/docs/user/CONCEPTS.md`
**Owner:** docs-dx-designer
**Deadline:** Oct 30 17:00 IST
**Rationale:** Mark holiday section as future
**Apply if:** ADR-0005 decides holidays NOT in v0.1

**INSERT BEFORE line 227 (before ## Holiday Handling):**
```

> **Note:** Holiday support is coming in v0.2. This section describes future functionality for planning purposes. The holiday fields in the CRD spec will be added in v0.2.

```

---

### EDIT-018: README.md - Add Test Step
**File:** `/Users/aykumar/personal/kyklos/README.md`
**Owner:** docs-dx-designer
**Deadline:** Oct 31 18:00 IST
**Rationale:** Include testing in quick start

**Search for verification step (around line 48) and add after:**
```

4. **Run smoke test** (optional but recommended)
   ```bash
   make test || echo "Tests will be available in implementation phase"
   ```
```

---

## Edit Group 7: Examples (HIGH PRIORITY)

### EDIT-019-A: Examples - Validate and Keep Holidays (If IN v0.1)
**File:** Multiple example files
**Owner:** docs-dx-designer
**Deadline:** Oct 30 17:00 IST
**Rationale:** Ensure examples are valid
**Apply if:** ADR-0005 decides holidays IN v0.1

**ACTION:**
```bash
# Validate all examples
kubectl apply --dry-run=client -f examples/tws-office-hours.yaml
kubectl apply --dry-run=client -f examples/tws-night-shift.yaml
kubectl apply --dry-run=client -f examples/tws-holidays-closed.yaml

# If validation fails, fix the YAML
# No text edits needed if validation passes
```

---

### EDIT-019-B: Examples - Move Holiday Example (If NOT in v0.1)
**File:** `examples/tws-holidays-closed.yaml`
**Owner:** docs-dx-designer
**Deadline:** Oct 30 17:00 IST
**Rationale:** Move future feature to separate directory
**Apply if:** ADR-0005 decides holidays NOT in v0.1

**ACTION:**
```bash
# Create future features directory
mkdir -p examples/future

# Move holiday example
git mv examples/tws-holidays-closed.yaml examples/future/

# Create README explaining future features
cat > examples/future/README.md << 'EOF'
# Future Features Examples

These examples demonstrate features planned for v0.2 and later.
They are provided for planning purposes but will not work with v0.1.

- `tws-holidays-closed.yaml` - Holiday handling (planned for v0.2)
EOF

# Validate remaining examples
kubectl apply --dry-run=client -f examples/tws-office-hours.yaml
kubectl apply --dry-run=client -f examples/tws-night-shift.yaml
```

---

## Edit Group 8: Testing Documentation (MEDIUM PRIORITY)

### EDIT-020: UNIT-PLAN.md - Add DST Scenarios
**File:** `/Users/aykumar/personal/kyklos/docs/testing/UNIT-PLAN.md`
**Owner:** testing-strategy-designer
**Deadline:** Oct 31 14:00 IST
**Rationale:** Specify DST test scenarios

**INSERT after existing scenarios:**
```

## DST Test Scenarios (Using Fixed Test Dates)

### DST-1: Spring Forward Transition
**Test Date:** 2025-03-09 (Sunday)
**Timezone:** America/New_York
**Fixture:** test/fixtures/dst-spring-2025.yaml
**Scenario:**
- Window: 01:00-04:00 on Sunday
- At 01:30 EST: window should be active
- At 02:30: this time does not exist (jumped to 03:30 EDT)
- At 03:30 EDT: window should be active
- Verify: Window duration is 2 hours, not 3

### DST-2: Fall Back Transition
**Test Date:** 2025-11-02 (Sunday)
**Timezone:** America/New_York
**Fixture:** test/fixtures/dst-fall-2025.yaml
**Scenario:**
- Window: 01:00-04:00 on Sunday
- At 01:30 EDT (first occurrence): window active
- At 01:30 EST (second occurrence after fallback): window still active
- Verify: Window duration is 4 hours, not 3

### DST-3: Cross-Midnight with Spring Forward
**Test Date:** 2025-03-08 22:00 to 2025-03-09 06:00
**Timezone:** America/New_York
**Fixture:** test/fixtures/dst-cross-midnight-2025.yaml
**Scenario:**
- Window: 22:00 Saturday to 06:00 Sunday
- Window spans midnight and DST transition
- Verify: Window remains active across both boundaries
- Verify: Total duration is 7 hours (lost 1 hour to DST)
```

---

### EDIT-021: ENVTEST-PLAN.md - Add Pause Scenarios
**File:** `/Users/aykumar/personal/kyklos/docs/testing/ENVTEST-PLAN.md`
**Owner:** testing-strategy-designer
**Deadline:** Oct 31 16:00 IST
**Rationale:** Test pause functionality

**INSERT after existing scenarios:**
```

## Pause Functionality Scenarios

### PAUSE-1: Pause During Active Window
**Setup:** Window active, replicas=5
**Action:** Set spec.pause=true
**Expected:**
- Controller computes effectiveReplicas=5
- No write to target deployment
- status.effectiveReplicas=5
- Ready=True if target already at 5
- ScalingSkipped event emitted

### PAUSE-2: Pause During Grace Period
**Setup:** Grace period active (transitioning 5→0)
**Action:** Set spec.pause=true during grace
**Expected:**
- Grace period timer stops
- Replicas remain at 5
- gracePeriodExpiry field cleared
- ScalingSkipped event emitted

### PAUSE-3: Resume from Pause
**Setup:** Paused with replicas=5, window now inactive (should be 0)
**Action:** Set spec.pause=false
**Expected:**
- Controller immediately writes replicas=0
- ScaledDown event emitted
- Ready condition updated
```

---

## Verification Commands

After applying all edits, run these commands:

```bash
# 1. Check terminology cleanup
git grep "activeReplicas" docs/ | grep -v "D9_\|D10_"
# Expected: 0 results

# 2. Check grace period field consistency
git grep 'gracePeriod[^S]' docs/ | grep -v "D9_\|D10_\|GracePeriod"
# Expected: 0 results (all should be gracePeriodSeconds or GracePeriodExpiry)

# 3. Verify test fixtures exist
ls -la test/fixtures/dst-*.yaml
# Expected: 3 files

# 4. Verify workflow exists
ls -la .github/workflows/ci.yml
# Expected: file exists

# 5. Validate examples
kubectl apply --dry-run=client -f examples/*.yaml
# Expected: all succeed

# 6. Check MAKE-TARGETS.md exists
test -f docs/MAKE-TARGETS.md && echo "EXISTS" || echo "MISSING"
# Expected: EXISTS
```

---

## Edit Statistics

**Total Edits:** 21 primary edits (some conditional on ADR-0005)
**Critical:** 13 edits
**High Priority:** 6 edits
**Medium Priority:** 2 edits

**Files Created:** 5 new files
**Files Modified:** 8 existing files
**Files Potentially Moved:** 1 (if holidays not in v0.1)

**Total Lines Changed:** ~250 lines across all files
**Average Edit Size:** 12 lines per edit

---

**Prepared by:** kyklos-orchestrator
**Date:** 2025-10-30 09:30 IST
**Status:** Ready for application
**Next Step:** See D10_MERGE_PLAN.md for application order
