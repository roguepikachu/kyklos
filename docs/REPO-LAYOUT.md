# Repository Layout

**Project:** Kyklos
**Last Updated:** 2025-10-19
**Owner:** kyklos-orchestrator

This document defines the repository structure, directory purposes, and ownership rules for the Kyklos project.

---

## Directory Structure

```
kyklos/
├── cmd/
│   └── controller/
│       └── main.go              # Controller entry point
├── api/
│   └── v1alpha1/
│       ├── timewindowscaler_types.go
│       ├── groupversion_info.go
│       └── zz_generated.deepcopy.go
├── controllers/
│   ├── timewindowscaler_controller.go
│   ├── timewindowscaler_controller_test.go
│   └── timecalc/                # Time calculation logic package
│       ├── state.go
│       ├── requeue.go
│       └── state_test.go
├── internal/
│   ├── webhook/                 # Admission webhook logic
│   │   ├── validator.go
│   │   └── validator_test.go
│   └── metrics/                 # Metrics and instrumentation
│       ├── metrics.go
│       └── recorder.go
├── config/
│   ├── crd/                     # CRD manifests
│   │   └── bases/
│   │       └── kyklos.io_timewindowscalers.yaml
│   ├── rbac/                    # RBAC manifests
│   │   ├── role.yaml
│   │   ├── clusterrole.yaml
│   │   ├── role_binding.yaml
│   │   └── service_account.yaml
│   ├── webhook/                 # Webhook configuration
│   │   ├── manifests.yaml
│   │   └── service.yaml
│   ├── manager/                 # Controller deployment
│   │   ├── deployment.yaml
│   │   └── kustomization.yaml
│   ├── samples/                 # Sample TimeWindowScaler CRs
│   │   ├── basic.yaml
│   │   ├── cross-midnight.yaml
│   │   ├── with-grace-period.yaml
│   │   └── cross-namespace.yaml
│   └── default/                 # Kustomize default overlay
│       └── kustomization.yaml
├── test/
│   ├── e2e/                     # End-to-end tests
│   │   ├── suite_test.go
│   │   └── timewindowscaler_test.go
│   ├── fixtures/                # Test data (fixed dates, timezones)
│   │   ├── dst-spring-forward.yaml
│   │   └── dst-fall-back.yaml
│   └── utils/                   # Test utilities (time mocking)
│       └── time_mock.go
├── examples/
│   ├── demo/                    # Demo scenario YAMLs
│   │   ├── setup.sh
│   │   └── demo-timewindowscaler.yaml
│   └── production/              # Production-ready examples
│       └── multi-namespace.yaml
├── docs/
│   ├── BRIEF.md
│   ├── DECISIONS.md
│   ├── RACI.md
│   ├── QUALITY-GATES.md
│   ├── ROADMAP.md
│   ├── HANDOFFS-DAY1.md
│   ├── REPO-LAYOUT.md (this file)
│   ├── COMMUNICATION.md
│   ├── RISKS.md
│   ├── CONCEPTS.md              # User-facing conceptual docs
│   ├── TROUBLESHOOTING.md       # Common issues and solutions
│   └── architecture-diagram.png # System architecture visual
├── design/                      # Design phase artifacts
│   ├── api-crd-spec.md
│   ├── reconcile-state-machine.md
│   └── ... (other design docs)
├── scripts/
│   ├── quick-start.sh           # Local cluster setup automation
│   └── time-warp-test.sh        # Time-fast-forward testing helper
├── .github/
│   ├── workflows/
│   │   ├── ci.yaml              # Lint, test, build, smoke test
│   │   └── release.yaml         # Tag-triggered release
│   └── ISSUE_TEMPLATE/
│       ├── bug_report.md
│       └── feature_request.md
├── Makefile                     # Build and test targets
├── Dockerfile                   # Multi-stage controller image build
├── go.mod
├── go.sum
├── README.md                    # User-facing entry point
├── CONTRIBUTING.md
└── LICENSE
```

---

## Directory Purposes and Ownership

### `/cmd/controller/`
**Purpose:** Controller binary entry point
**Owner:** controller-reconcile-designer
**Contents:** main.go with manager setup, flag parsing, webhook/metrics server init
**Rules:**
- Must remain minimal (delegate to controllers/ and internal/)
- No business logic in main.go

---

### `/api/v1alpha1/`
**Purpose:** TimeWindowScaler CRD Go types
**Owner:** api-crd-designer
**Contents:** Struct definitions for Spec and Status, kubebuilder markers
**Rules:**
- All fields must have json/yaml tags and godoc comments
- kubebuilder markers for OpenAPI validation
- Generate CRD YAML via `make manifests`
- No controller logic in this package

---

### `/controllers/`
**Purpose:** Reconcile loop implementation
**Owner:** controller-reconcile-designer
**Contents:** TimeWindowScalerReconciler and supporting packages
**Rules:**
- Reconcile function must be idempotent
- Test coverage >= 80%
- Use timecalc/ subpackage for time logic (facilitates testing)

---

### `/controllers/timecalc/`
**Purpose:** Time calculation and state determination logic
**Owner:** controller-reconcile-designer
**Contents:** Pure functions for state machine, requeue timing, DST handling
**Rules:**
- **No Kubernetes client dependencies** (must be unit testable)
- Accept current time as parameter (enables time mocking)
- 100% test coverage required (critical path)
- Document DST edge cases in godoc

---

### `/internal/webhook/`
**Purpose:** Admission webhook validation logic
**Owner:** api-validation-defaults-designer
**Contents:** ValidatingWebhook implementation, validator functions
**Rules:**
- Validate timezone via time.LoadLocation
- Validate time format via time.Parse
- Return clear error messages for users
- No mutation logic (validating only in v0.1)

---

### `/internal/metrics/`
**Purpose:** Prometheus metrics and event emission
**Owner:** observability-metrics-designer
**Contents:** Metric definitions, recorder wrapper for status updates
**Rules:**
- Register all metrics in init()
- Use consistent label naming (see QUALITY-GATES.md Gate 4)
- Document each metric with help text

---

### `/config/`
**Purpose:** Kubernetes manifests for CRD, RBAC, controller deployment
**Owner:** Multiple (see subdirectories)
**Contents:** Kustomize-based manifest organization
**Rules:**
- Use kustomize for composition (no templating languages)
- Keep base/ minimal, use overlays for variations
- Validate with `kubectl kustomize` before commit

#### `/config/crd/`
**Owner:** api-crd-designer
**Generated:** Yes (via controller-gen from api/)
**Manual Edits:** No (regenerate from Go types)

#### `/config/rbac/`
**Owner:** security-rbac-designer
**Generated:** Partially (role.yaml from markers, manual for ClusterRole)
**Manual Edits:** Yes (review generated permissions, add cross-namespace ClusterRole)

#### `/config/webhook/`
**Owner:** api-validation-defaults-designer
**Generated:** Partially (manifests from markers)
**Manual Edits:** Yes (TLS cert strategy, failure policy)

#### `/config/manager/`
**Owner:** ci-release-designer
**Generated:** No (manually maintained)
**Manual Edits:** Yes (resource limits, image pull policy, args)

#### `/config/samples/`
**Owner:** docs-dx-designer
**Generated:** No (manually crafted examples)
**Manual Edits:** Yes (must match design/api-crd-spec.md)

---

### `/test/`
**Purpose:** Test suites and utilities
**Owner:** testing-strategy-designer
**Contents:** e2e tests, fixtures, mock utilities
**Rules:**
- e2e tests use envtest or real cluster
- fixtures/ contains fixed date/time test data
- utils/ must not depend on controller code (imported by controller tests)

---

### `/examples/`
**Purpose:** User-facing example scenarios
**Owner:** docs-dx-designer
**Contents:** Demo scripts, production-ready YAMLs
**Rules:**
- All examples must be tested (runnable via scripts)
- Include setup and teardown steps
- Document expected outcome in comments

---

### `/docs/`
**Purpose:** Project documentation (planning, design, user guides)
**Owner:** docs-dx-designer (user-facing), kyklos-orchestrator (planning/design)
**Contents:** Planning docs (BRIEF, DECISIONS, etc.), user guides (CONCEPTS, TROUBLESHOOTING)
**Rules:**
- Planning docs (BRIEF, DECISIONS, RACI, etc.) owned by kyklos-orchestrator
- User guides (CONCEPTS, TROUBLESHOOTING) owned by docs-dx-designer
- Keep glossary in BRIEF.md as single source of truth
- All docs must use glossary terms consistently

---

### `/design/`
**Purpose:** Design phase artifacts (Day 1-14)
**Owner:** Respective designers (see RACI.md)
**Contents:** Detailed design documents per work stream
**Rules:**
- File naming: {workstream}-{topic}.md (e.g., api-crd-spec.md)
- Archive to design/archive/ after implementation completes
- Reference from docs/DECISIONS.md ADRs when relevant

---

### `/scripts/`
**Purpose:** Automation scripts for local development and testing
**Owner:** local-workflow-designer
**Contents:** Setup scripts, testing helpers
**Rules:**
- Must be POSIX-compliant shell (sh, not bash-specific)
- Document prerequisites at top of script
- Exit on error (set -e)

---

### `/.github/`
**Purpose:** GitHub-specific configuration (CI, issue templates)
**Owner:** ci-release-designer (workflows), community-launch-designer (templates)
**Contents:** GitHub Actions workflows, issue/PR templates
**Rules:**
- Workflows must use pinned action versions (not @main)
- Secrets via GitHub Secrets (not hardcoded)
- Document required secrets in workflows/README.md

---

## File Naming Conventions

### Go Files
- Types: `{resource}_types.go` (e.g., timewindowscaler_types.go)
- Controllers: `{resource}_controller.go`
- Tests: `{file}_test.go` (e.g., state_test.go)
- Packages: lowercase, single word (no underscores)

### YAML Manifests
- CRDs: `{group}_{plural}.yaml` (e.g., kyklos.io_timewindowscalers.yaml)
- Examples: `{scenario}.yaml` (e.g., cross-midnight.yaml)
- Config: `{resource}.yaml` (e.g., deployment.yaml)

### Documentation
- Planning docs: UPPERCASE.md (e.g., BRIEF.md, DECISIONS.md)
- Design docs: lowercase-with-dashes.md (e.g., api-crd-spec.md)
- User guides: mixed case (e.g., CONCEPTS.md, TROUBLESHOOTING.md)

---

## Ownership and Modification Rules

### Who Can Modify What

| Directory | Primary Owner | Can Modify | Must Consult |
|-----------|---------------|------------|--------------|
| `/api/` | api-crd-designer | api-crd-designer | controller-reconcile-designer, api-validation-defaults-designer |
| `/controllers/` | controller-reconcile-designer | controller-reconcile-designer | observability-metrics-designer |
| `/internal/webhook/` | api-validation-defaults-designer | api-validation-defaults-designer | security-rbac-designer |
| `/config/rbac/` | security-rbac-designer | security-rbac-designer | controller-reconcile-designer |
| `/docs/` planning | kyklos-orchestrator | kyklos-orchestrator | All agents (for updates) |
| `/docs/` user-facing | docs-dx-designer | docs-dx-designer | Relevant domain experts |
| `/design/` | Per work stream | Respective designer | Per RACI.md |

### Modification Protocol
1. **Before Modifying:** Check RACI.md for Accountable party
2. **During Modification:** If cross-cutting, notify Consulted parties
3. **After Modification:** Commit with clear message, notify Accountable for review
4. **If Conflict:** Escalate to kyklos-orchestrator per COMMUNICATION.md

---

## Generated vs Manual Files

### Generated (Do Not Edit Manually)
- `api/v1alpha1/zz_generated.deepcopy.go` - Generated by controller-gen
- `config/crd/bases/*.yaml` - Generated by controller-gen from Go types
- `go.sum` - Generated by go mod

**Regenerate via:**
```bash
make manifests  # CRDs and RBAC
make generate   # deepcopy
```

### Partially Generated (Review and Augment)
- `config/rbac/role.yaml` - Generated base, manually add cross-namespace permissions
- `config/webhook/manifests.yaml` - Generated base, manually configure TLS and failure policy

### Fully Manual
- All `/docs/` and `/design/` markdown files
- All `/examples/` and `/config/samples/` YAMLs
- All scripts in `/scripts/`
- Makefile, Dockerfile, README.md

---

## Testing File Organization

### Unit Tests
- Colocated with source: `{file}_test.go` next to `{file}.go`
- Use table-driven tests for multiple scenarios
- Mock time via interface or build tags

### Envtest (Integration)
- Location: `/controllers/*_test.go`
- Use controller-runtime envtest package
- Test against real API server (no mocks for Kubernetes client)

### E2E Tests
- Location: `/test/e2e/`
- Use Ginkgo/Gomega or standard testing
- Assume cluster exists (setup in CI, not in test code)

### Test Fixtures
- Location: `/test/fixtures/`
- Fixed dates for DST scenarios (e.g., 2025-03-09, 2025-11-02)
- YAML manifests for reproducible test scenarios

---

## CI Artifacts

### Build Outputs
- Binary: `bin/controller` (gitignored)
- Container image: `kyklos/controller:$TAG`
- Release bundle: `kyklos-v0.1.0.yaml` (all-in-one manifest)

### CI-Generated Files
- Test coverage report: `coverage.out`
- Linter output: `golangci-lint.out`
- Vulnerability scan: `trivy-report.json`

All CI artifacts are gitignored, not committed to repository.

---

## Documentation Flow

```
Planning Phase (Day 0-14):
docs/BRIEF.md → design/*.md → docs/DECISIONS.md (ADRs)
                    ↓
            Implementation Phase:
            Code in api/, controllers/, internal/
                    ↓
            User-Facing Docs:
            docs/CONCEPTS.md, README.md, examples/
```

### Documentation Update Triggers
- **API Changes:** Update design/api-crd-spec.md → docs/CONCEPTS.md → README.md examples
- **Behavior Changes:** Update design/reconcile-*.md → docs/CONCEPTS.md → docs/TROUBLESHOOTING.md
- **New Decisions:** Create ADR in docs/DECISIONS.md → Update docs/BRIEF.md if scope changes

---

## Quick Reference: Where to Put New Files

**New CRD Field?** → `/api/v1alpha1/timewindowscaler_types.go` (then `make manifests`)
**New Validation Rule?** → `/internal/webhook/validator.go` + test
**New Metric?** → `/internal/metrics/metrics.go` + update docs/CONCEPTS.md
**New RBAC Permission?** → `/config/rbac/` + justify in design/rbac-permissions.md
**New Test Scenario?** → `/test/e2e/` or `/controllers/*_test.go` depending on scope
**New Example?** → `/examples/` (user-facing) or `/config/samples/` (simple demos)
**New Design Decision?** → Create ADR in `docs/DECISIONS.md`
**New Script?** → `/scripts/` with clear name and documentation

---

## Repository Health Checks

**Before Committing:**
- [ ] Run `make test` (unit tests pass)
- [ ] Run `make manifests` (CRDs up to date)
- [ ] Run `make lint` (no linter errors)
- [ ] Check file is in correct directory per this document
- [ ] If modifying API, update design/api-crd-spec.md
- [ ] If adding field, update docs/BRIEF.md glossary if needed

**Before Merging PR:**
- [ ] CI passes (all checks green)
- [ ] Consulted parties (per RACI) have reviewed
- [ ] Documentation updated if user-facing change
- [ ] DECISIONS.md updated if architectural change

---

## Evolution of This Document

This layout is for v0.1 design and implementation. As project evolves:
- New directories must be documented here before creation
- Ownership changes must update RACI.md and this document in sync
- Directory purpose must remain clear (no "misc" or "utils" dumping grounds)

**Document Owner:** kyklos-orchestrator
**Review Frequency:** At each minor version (v0.2, v0.3, etc.)
**Change Process:** Propose change → Update REPO-LAYOUT.md → Notify all agents → Merge
