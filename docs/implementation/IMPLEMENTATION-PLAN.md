# Kyklos v0.1 Implementation Plan

**Target Delivery:** One working slice in 7 days
**Last Updated:** 2025-10-29
**Status:** Ready for Implementation

## Executive Summary

This document defines the implementation scope, module boundaries, and milestones for Kyklos v0.1. The plan focuses on delivering a **minimal working slice** that can scale a Deployment based on time windows within one week, with subsequent milestones adding robustness, observability, and production readiness.

## Version Scope

### What's IN Scope for v0.1

**Core Functionality:**
- TimeWindowScaler CRD with v1alpha1 API
- Single time window per day (Mon-Sun)
- Timezone-aware scheduling with IANA database
- Cross-midnight window support (22:00-02:00)
- Grace period for scale-down operations
- Pause functionality (compute but don't apply)
- Manual drift correction
- Status conditions (Ready, Reconciling, Degraded)
- Event emission for scale operations
- Prometheus metrics for observability

**Target Workloads:**
- Deployment (apps/v1) only in v0.1
- Same-namespace and cross-namespace scaling

**Holiday Support:**
- ConfigMap-based holiday definitions
- Three modes: ignore, treat-as-closed, treat-as-open
- ISO date format (YYYY-MM-DD)

**Validation:**
- CRD-based OpenAPI validation only
- No admission webhooks in v0.1 (deferred to v0.2)

### What's OUT of Scope for v0.1

**Deferred to v0.2+:**
- Admission webhook for runtime validation
- StatefulSet and DaemonSet support
- Multiple windows per day
- HPA/VPA integration
- Cron expression syntax
- Defaulting webhook
- Conversion webhooks for API versioning

**Explicitly Excluded:**
- Web UI or dashboard
- Multi-cluster support
- Cost estimation features
- External calendar integration (beyond ConfigMap)
- Notification systems (Slack, email)

## Module Boundaries

### Module 1: API Types (`/api/v1alpha1`)
**Purpose:** Go struct definitions for TimeWindowScaler CRD
**Dependencies:** None (pure types)
**Owned By:** API designer
**Interface:** Public Go types with JSON/YAML tags

**Components:**
- `timewindowscaler_types.go` - Spec and Status structs
- `groupversion_info.go` - API group registration
- `zz_generated.deepcopy.go` - Generated (controller-gen)

**Key Decisions:**
- No business logic in this package
- All validation via kubebuilder markers
- Field names match CRD-SPEC.md exactly

### Module 2: Time Calculation Engine (`/internal/timecalc`)
**Purpose:** Pure functions for time window logic
**Dependencies:** Go standard library only (time, math)
**Owned By:** Controller designer
**Interface:** Public functions with explicit time input

**Components:**
- `matcher.go` - Window matching algorithm
- `boundary.go` - Next boundary computation
- `grace.go` - Grace period state machine
- `holiday.go` - Holiday evaluation logic

**Key Decisions:**
- Zero Kubernetes dependencies (unit testable)
- All functions accept `time.Time` parameter (no `time.Now()` calls)
- 100% test coverage required (critical path)

### Module 3: Reconciler (`/controllers`)
**Purpose:** Main reconciliation loop
**Dependencies:** controller-runtime, timecalc, statuswriter, eventrecorder
**Owned By:** Controller designer
**Interface:** controller-runtime Reconciler interface

**Components:**
- `timewindowscaler_controller.go` - Reconcile function
- `controller_test.go` - Envtest integration tests

**Key Decisions:**
- Idempotent operations only
- No inline time calculations (delegate to timecalc)
- Status updates via statuswriter module
- Events via eventrecorder module

### Module 4: Status Writer (`/internal/statuswriter`)
**Purpose:** Status subresource updates with retry logic
**Dependencies:** controller-runtime client
**Owned By:** Controller designer
**Interface:** `UpdateStatus(tws, effectiveReplicas, conditions, observedGen) error`

**Components:**
- `writer.go` - Status update with optimistic locking
- `conditions.go` - Condition builder helpers

**Key Decisions:**
- Single atomic update for all status fields
- Automatic retry on conflict (409)
- Condition timestamp management

### Module 5: Event Recorder (`/internal/events`)
**Purpose:** Kubernetes event emission with deduplication
**Dependencies:** controller-runtime recorder
**Owned By:** Observability designer
**Interface:** `EmitScaleUp/Down/Skipped(...) error`

**Components:**
- `recorder.go` - Event emission facade
- `dedup.go` - 5-minute deduplication logic

**Key Decisions:**
- Rate limit: 20 events/minute per TWS
- Warning events always emitted
- Normal events deduplicated within 5 minutes

### Module 6: Metrics (`/internal/metrics`)
**Purpose:** Prometheus metrics exposition
**Dependencies:** prometheus client_golang
**Owned By:** Observability designer
**Interface:** `RecordScaleEvent/StateChange/ReconcileDuration(...)`

**Components:**
- `metrics.go` - Metric definitions and registration
- `recorder.go` - Metrics recording facade

**Key Decisions:**
- Metrics registered in init()
- Labels: tws_name, namespace, window, state
- Counters for scale events, gauges for state

### Module 7: Config Manifests (`/config`)
**Purpose:** Kustomize-based Kubernetes manifests
**Dependencies:** None (declarative YAML)
**Owned By:** Multiple (see subdirectories)
**Interface:** `kubectl kustomize config/default | kubectl apply`

**Structure:**
```
/config
├── crd/bases/              # Generated CRD YAML
├── rbac/                   # Role, RoleBinding, ServiceAccount
├── manager/                # Controller Deployment
├── samples/                # Example TimeWindowScaler CRs
└── default/                # Kustomize overlay
```

**Key Decisions:**
- Base manifests in respective directories
- Overlays for dev/staging/prod
- No Helm or templating in v0.1

## First Working Slice Definition

**Goal:** Demonstrate basic time-based scaling within 7 days

**Acceptance Criteria:**
1. Can deploy TimeWindowScaler CRD to cluster
2. Can create TWS resource with single window (09:00-17:00)
3. Controller detects time boundary and scales Deployment
4. Status.effectiveReplicas reflects computed state
5. `kubectl get tws` shows current window status
6. Basic logs visible via `kubectl logs`

**What Works:**
- Basic window matching (no cross-midnight)
- Scale up and scale down
- Status updates with observedGeneration
- Single condition: Ready (true/false)

**What's Minimal/Missing:**
- No grace period (added in M2)
- No holiday support (added in M3)
- No events (added in M2)
- No metrics (added in M4)
- No cross-midnight windows (added in M2)
- No pause support (added in M2)

**Time Budget:**
- Day 1-2: API types + CRD generation (M1)
- Day 3-4: Time calculation engine (M1)
- Day 5-6: Basic reconciler (M1)
- Day 7: Integration + smoke test (M1)

## Milestones

### Milestone 1 (M1): Minimal Viable Controller
**Duration:** 7 days (Day 1-7)
**Goal:** Working controller that scales based on simple time windows

**Deliverables:**
- [ ] TimeWindowScaler CRD with basic fields
- [ ] Time calculation engine (window matching only)
- [ ] Basic reconciler (no grace, no holidays, no pause)
- [ ] Status updates with Ready condition
- [ ] RBAC for same-namespace scaling
- [ ] Basic unit tests (60% coverage)
- [ ] Smoke test script
- [ ] README with quick start

**Acceptance Criteria:**
- Deploys to local kind cluster
- Scales Deployment up during window
- Scales Deployment down outside window
- Status shows current effectiveReplicas
- Controller restarts without losing state
- Single time window works (09:00-17:00 in one timezone)

**Risk Mitigation:**
- No complex features (defer to M2)
- Focus on happy path only
- Manual testing acceptable (no full test suite yet)

### Milestone 2 (M2): Production Features
**Duration:** 5 days (Day 8-12)
**Goal:** Add grace period, cross-midnight, pause, events

**Deliverables:**
- [ ] Grace period implementation
- [ ] Cross-midnight window support
- [ ] Pause functionality
- [ ] Event emission (ScaledUp/Down/Skipped)
- [ ] Enhanced status conditions (Reconciling, Degraded)
- [ ] Requeue scheduling with jitter
- [ ] Envtest integration tests
- [ ] Enhanced RBAC for cross-namespace

**Acceptance Criteria:**
- Grace period delays scale-down correctly
- Windows spanning midnight work (22:00-02:00)
- Pause prevents target updates
- Events visible via `kubectl get events`
- All status conditions populated
- Envtest suite passes
- Unit test coverage >80%

**Risk Mitigation:**
- Grace period state persisted in status.gracePeriodExpiry
- Cross-midnight tested with fixed date fixtures
- Pause tested with manual drift scenarios

### Milestone 3 (M3): Holiday Support
**Duration:** 3 days (Day 13-15)
**Goal:** ConfigMap-based holiday handling

**Deliverables:**
- [ ] Holiday ConfigMap loading
- [ ] Three modes: ignore, treat-as-closed, treat-as-open
- [ ] Holiday override events
- [ ] Holiday test fixtures
- [ ] Holiday documentation

**Acceptance Criteria:**
- ConfigMap with ISO dates works
- treat-as-closed uses defaultReplicas
- treat-as-open uses max(window replicas)
- Missing ConfigMap handled gracefully
- Holiday events emitted

**Risk Mitigation:**
- ConfigMap loading cached (no API call per reconcile)
- Missing ConfigMap = Degraded condition but continues

### Milestone 4 (M4): Observability
**Duration:** 3 days (Day 16-18)
**Goal:** Metrics, structured logging, operational visibility

**Deliverables:**
- [ ] Prometheus metrics (state, scale events, reconcile duration)
- [ ] Structured logging with correlation IDs
- [ ] Metrics documentation
- [ ] Grafana dashboard JSON
- [ ] Alerting rules examples

**Acceptance Criteria:**
- Metrics exposed on :8080/metrics
- Can scrape with Prometheus
- Dashboard shows current state
- Logs parseable as JSON
- Alert rules trigger on Degraded state

**Risk Mitigation:**
- Metrics registered in init() (no registration failures)
- Logs use logr interface (controller-runtime standard)

### Milestone 5 (M5): Hardening and Release
**Duration:** 4 days (Day 19-22)
**Goal:** E2E tests, documentation, release artifacts

**Deliverables:**
- [ ] E2E test suite (3 scenarios)
- [ ] DST transition tests
- [ ] Manual drift correction tests
- [ ] Comprehensive README
- [ ] TROUBLESHOOTING.md
- [ ] Release container image
- [ ] All-in-one install YAML
- [ ] GitHub release with notes

**Acceptance Criteria:**
- E2E tests pass in kind cluster
- DST tests cover spring forward and fall back
- Manual scale-up corrected automatically
- README 5-minute quick start works
- Container image published
- Release tagged v0.1.0

**Risk Mitigation:**
- E2E tests use time-warp (fast minutes, not real hours)
- DST tests use fixed dates (2025-03-09, 2025-11-02)
- Release checklist from RELEASE-POLICY.md

## Implementation Phases

### Phase 1: Foundation (M1 Days 1-3)
**Focus:** API and time calculation engine

**Tasks:**
1. Create Go module and directory structure
2. Define TimeWindowScaler types in `/api/v1alpha1`
3. Add kubebuilder markers for CRD generation
4. Generate CRD manifests with `make manifests`
5. Implement time calculation functions in `/internal/timecalc`
6. Write unit tests for time calculations (100% coverage)
7. Validate CRD installs to test cluster

**Success:** CRD installs, time calculations pass unit tests

### Phase 2: Basic Reconciler (M1 Days 4-6)
**Focus:** Reconciliation loop

**Tasks:**
1. Implement TimeWindowScalerReconciler
2. Add RBAC markers for controller permissions
3. Wire up manager in `cmd/controller/main.go`
4. Implement status updates (Ready condition only)
5. Add requeue logic (simple: check every minute)
6. Write controller integration test (envtest)
7. Build controller binary and container image

**Success:** Controller scales Deployment in kind cluster

### Phase 3: Integration and Smoke Test (M1 Day 7)
**Focus:** End-to-end validation

**Tasks:**
1. Deploy controller to kind cluster
2. Create sample TimeWindowScaler CR
3. Verify Deployment scales up/down
4. Check status updates
5. Test controller restart (state preserved)
6. Write smoke test script
7. Document quick start in README

**Success:** 15-minute quick start works from clone to scale

### Phase 4: Enhanced Features (M2 Days 8-12)
**Focus:** Grace period, cross-midnight, pause, events

**Tasks:**
1. Implement grace period state machine
2. Add gracePeriodExpiry to status
3. Implement cross-midnight window detection
4. Add pause field and conditional write logic
5. Implement event recorder module
6. Enhance status conditions (Reconciling, Degraded)
7. Improve requeue scheduling with jitter
8. Add envtest scenarios for new features
9. Update RBAC for cross-namespace

**Success:** All M2 acceptance criteria pass

### Phase 5: Holiday Support (M3 Days 13-15)
**Focus:** ConfigMap-based holiday handling

**Tasks:**
1. Implement holiday ConfigMap loader with caching
2. Add holiday evaluation to time calculation engine
3. Emit holiday events
4. Test with sample ConfigMap
5. Handle missing ConfigMap gracefully
6. Add holiday test fixtures
7. Document holiday feature

**Success:** All three holiday modes work correctly

### Phase 6: Observability (M4 Days 16-18)
**Focus:** Metrics, logging, dashboards

**Tasks:**
1. Define Prometheus metrics
2. Implement metrics recorder
3. Expose metrics endpoint
4. Add structured logging with correlation IDs
5. Create Grafana dashboard
6. Write alerting rules
7. Document metrics and logs

**Success:** Dashboard shows live state, alerts fire

### Phase 7: Hardening (M5 Days 19-22)
**Focus:** E2E tests, documentation, release

**Tasks:**
1. Write E2E test scenarios
2. Test DST transitions
3. Test manual drift correction
4. Complete README and troubleshooting docs
5. Build and push container image
6. Generate all-in-one install YAML
7. Create GitHub release
8. Announce release

**Success:** v0.1.0 release published and usable

## Critical Path

The critical path determines minimum time to completion:

```
Day 1-3: API + Time Engine (M1) → Cannot parallelize (foundation)
  ↓
Day 4-6: Basic Reconciler (M1) → Depends on API
  ↓
Day 7: Integration Test (M1) → Depends on reconciler
  ↓
Day 8-12: Enhanced Features (M2) → Can partially parallelize:
  - Grace period (sequential with M1)
  - Events (parallel with grace period)
  - Cross-midnight (depends on time engine)
  ↓
Day 13-15: Holiday Support (M3) → Depends on reconciler
  ↓
Day 16-18: Observability (M4) → Parallel with M3 (can start earlier)
  ↓
Day 19-22: Hardening (M5) → Depends on all features
```

**Total Duration:** 22 days (minimum)
**First Slice:** 7 days (M1 only)
**Production Ready:** 22 days (M1-M5 complete)

## Dependencies

### External Dependencies
- Kubernetes 1.25+ cluster (kind/minikube for local)
- Go 1.21+ with module support
- controller-runtime v0.16+
- kubebuilder tools (controller-gen, kustomize)
- Docker/Podman for image building

### Internal Dependencies
```
timecalc → (no dependencies)
  ↓
api types → (depends on kubebuilder markers)
  ↓
statuswriter → (depends on api types)
eventrecorder → (depends on api types)
metrics → (depends on api types)
  ↓
reconciler → (depends on all above)
  ↓
main.go → (depends on reconciler)
```

### Documentation Dependencies
- CRD-SPEC.md → guides API types
- RECONCILE.md → guides reconciler logic
- STATUS-CONDITIONS.md → guides statuswriter
- EVENTS.md → guides eventrecorder
- TEST-STRATEGY.md → guides test implementation

## Resource Requirements

### Development Environment
- 1 developer (full-time for M1)
- Local Kubernetes cluster (kind: 4 CPU, 8GB RAM)
- Go development tools
- Test cluster for integration tests

### CI/CD Requirements
- GitHub Actions runners (standard tier)
- Container registry (ghcr.io or Docker Hub)
- Artifact storage for test reports

### Time Budget per Module
| Module | Unit Tests | Integration | E2E | Total |
|--------|-----------|-------------|-----|-------|
| API Types | 1 day | - | - | 1 day |
| Time Engine | 2 days | - | - | 2 days |
| Reconciler | 2 days | 2 days | 1 day | 5 days |
| Status Writer | 1 day | 1 day | - | 2 days |
| Event Recorder | 1 day | 1 day | - | 2 days |
| Metrics | 1 day | 1 day | - | 2 days |
| Config/RBAC | 1 day | - | - | 1 day |
| Documentation | - | - | - | 3 days |
| **Total** | 9 days | 5 days | 1 day | **18 days** |

*Note: 18 days is optimistic; buffer brings to 22 days*

## Risk Register

### High Risk: Time Calculation Complexity
**Impact:** Bugs in window matching affect all users
**Mitigation:**
- 100% unit test coverage for timecalc package
- Use fixed date test fixtures
- Test all DST transitions explicitly
- Code review by second developer

### Medium Risk: Cross-Midnight Edge Cases
**Impact:** Windows spanning midnight may not trigger correctly
**Mitigation:**
- Dedicated test suite for cross-midnight scenarios
- Test at 23:59, 00:00, 00:01 boundaries
- Document algorithm in PSEUDOCODE.md

### Medium Risk: Grace Period State Loss
**Impact:** Controller restart could lose grace period state
**Mitigation:**
- Persist gracePeriodExpiry in status subresource
- Reconciler checks expiry timestamp, not timer
- Test controller restart during grace period

### Low Risk: Hot Reconcile Loop
**Impact:** Tight loop causes excessive API calls
**Mitigation:**
- Minimum requeue 30 seconds
- Add jitter to prevent thundering herd
- Rate limiting metrics to detect hot loops

### Low Risk: RBAC Gaps
**Impact:** Controller lacks permissions for operations
**Mitigation:**
- Generate RBAC from kubebuilder markers
- Test same-namespace and cross-namespace modes
- Document required permissions in RBAC-MATRIX.md

## Success Metrics

### M1 Success Criteria
- [ ] Can create TWS resource successfully
- [ ] Controller logs show reconciliation
- [ ] Deployment scales during window
- [ ] Deployment scales outside window
- [ ] Status.effectiveReplicas is correct
- [ ] Controller survives restart
- [ ] Unit tests pass
- [ ] 15-minute quick start documented

### M2 Success Criteria
- [ ] All M1 criteria still pass
- [ ] Grace period delays scale-down
- [ ] Cross-midnight windows work
- [ ] Pause prevents scaling
- [ ] Events appear in `kubectl get events`
- [ ] Status conditions populated correctly
- [ ] Envtest suite passes
- [ ] Unit coverage >80%

### M3 Success Criteria
- [ ] Holiday ConfigMap loads successfully
- [ ] treat-as-closed mode works
- [ ] treat-as-open mode works
- [ ] ignore mode works
- [ ] Missing ConfigMap handled gracefully
- [ ] Holiday events emitted

### M4 Success Criteria
- [ ] Metrics exposed on /metrics
- [ ] Prometheus can scrape metrics
- [ ] Grafana dashboard displays state
- [ ] Logs parseable as JSON
- [ ] Alert rules provided

### M5 Success Criteria
- [ ] E2E tests pass in kind cluster
- [ ] DST tests pass (spring forward, fall back)
- [ ] Manual drift corrected automatically
- [ ] README 5-minute quick start works
- [ ] Container image published
- [ ] Release v0.1.0 tagged
- [ ] TROUBLESHOOTING.md complete

## Post-v0.1 Roadmap

### v0.2 (Next Release)
- Admission webhook for validation
- Defaulting webhook
- Multiple windows per day
- StatefulSet support

### v0.3 (Future)
- Conversion webhook for API versioning
- v1beta1 API graduation
- DaemonSet support
- Enhanced metrics and dashboard

### v1.0 (Stable)
- v1 API
- Production hardening
- Performance optimization
- Multi-cluster support (stretch)

## Related Documents

- [CRD-SPEC.md](/Users/aykumar/personal/kyklos/docs/api/CRD-SPEC.md) - API specification
- [RECONCILE.md](/Users/aykumar/personal/kyklos/docs/design/RECONCILE.md) - Reconcile design
- [STATUS-CONDITIONS.md](/Users/aykumar/personal/kyklos/docs/design/STATUS-CONDITIONS.md) - Status design
- [EVENTS.md](/Users/aykumar/personal/kyklos/docs/design/EVENTS.md) - Event design
- [TEST-STRATEGY.md](/Users/aykumar/personal/kyklos/docs/testing/TEST-STRATEGY.md) - Testing approach
- [TASKS.csv](/Users/aykumar/personal/kyklos/docs/implementation/TASKS.csv) - Granular task breakdown
- [INTERFACE-CONTRACTS.md](/Users/aykumar/personal/kyklos/docs/implementation/INTERFACE-CONTRACTS.md) - Module contracts
- [PSEUDOCODE.md](/Users/aykumar/personal/kyklos/docs/implementation/PSEUDOCODE.md) - Core algorithms
