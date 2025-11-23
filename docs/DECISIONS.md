# Architecture Decision Records

## ADR-0001: Name and Scope Decisions for v0.1

**Date:** 2025-10-19
**Status:** Accepted
**Deciders:** kyklos-orchestrator

### Context
Need to establish project name, primary CRD name, and scope boundaries for initial release.

### Decision
- **Project Name:** Kyklos (Greek for "cycle/circle", representing recurring time patterns)
- **CRD Name:** TimeWindowScaler (abbrev: TWS)
- **API Group:** kyklos.io
- **Version:** v1alpha1
- **Scope:** Single daily time window per resource, scales one target workload

### Consequences
**Positive:**
- Clear, memorable name with meaningful etymology
- Explicit in naming (not generic like "autoscaler")
- API group avoids conflicts with existing operators
- Alpha version signals evolving API

**Negative:**
- Single window limitation may require multiple CRs for complex schedules
- Greek name may need explanation for some users

**Alternatives Considered:**
- TimeScaler: Too generic, conflicts with TimescaleDB mental model
- WindowedScaler: Verbose, unclear purpose
- ScheduledScaler: Implies cron-like complexity we don't support

### Open Questions
- Should we allow multiple windows in v0.2? If so, precedence rules?

---

## ADR-0002: Supported Target Kind and Namespace Model

**Date:** 2025-10-19
**Status:** Accepted
**Deciders:** kyklos-orchestrator

### Context
Controller must scale various workload types. Must decide which kinds to support and whether to allow cross-namespace references.

### Decision
- **Supported Kinds:** Deployment, StatefulSet, ReplicaSet
- **Namespace Model:** Support same-namespace (default) and cross-namespace with explicit RBAC
- **Reference Format:** Standard corev1.ObjectReference (apiVersion, kind, name, namespace)
- **Scale Mechanism:** Use Scale subresource API (apps/v1 scale)

### Consequences
**Positive:**
- Scale subresource provides uniform interface across kinds
- Cross-namespace support enables centralized TimeWindowScaler management
- ObjectReference is standard Kubernetes pattern

**Negative:**
- Cross-namespace requires ClusterRole instead of namespaced Role
- Must validate target exists and is scalable before reconcile
- Security boundary crossing requires careful RBAC design

**Alternatives Considered:**
- Same-namespace only: Too restrictive for multi-tenant use cases
- Support DaemonSet: Has no replica count, would need special handling
- Support custom resources: Too complex for v0.1, requires discovery

### Open Questions
- Should we support scaling Jobs or CronJobs in future versions?
- Do we need admission webhook to validate target reference on create?

---

## ADR-0003: Timezone and DST Handling Approach

**Date:** 2025-10-19
**Status:** Accepted
**Deciders:** kyklos-orchestrator

### Context
Time windows must work correctly across timezones and DST transitions. Need approach that's correct, testable, and understandable.

### Decision
- **Timezone Specification:** IANA timezone string (e.g., "Asia/Kolkata")
- **DST Handling:** Use Go time.LoadLocation and wall clock time (not duration arithmetic)
- **Controller Clock:** System time (time.Now()) assumed synchronized via NTP
- **Requeue Strategy:** Calculate next state transition time, requeue with jitter
- **Cross-Midnight:** Support windows like "22:00-02:00" by date comparison logic

### Consequences
**Positive:**
- Go standard library handles DST transitions correctly
- IANA timezone database is standard and maintained
- Wall clock time matches user mental model
- Explicit timezone in CR makes behavior clear

**Negative:**
- Testing requires mocking time or using fixed test dates
- DST spring-forward can cause 1-hour window to be skipped
- DST fall-back can cause window to occur twice (2 AM repeated)
- Invalid timezone strings cause runtime errors

**Mitigations:**
- Validation webhook rejects invalid timezones
- Status condition reports timezone load errors
- Document DST edge cases in user guide
- Add test cases for spring-forward and fall-back dates

**Alternatives Considered:**
- UTC-only with offset: Loses DST awareness, user-unfriendly
- Duration-based: Fails on DST transitions
- Calendar library: Adds dependency, overkill for daily windows

### Open Questions
- Should we emit warning metrics on DST transition dates?
- Do we need "skip DST transition days" option for critical workloads?

---

## ADR-0004: Grace Period Semantics

**Date:** 2025-10-19
**Status:** Accepted
**Deciders:** kyklos-orchestrator

### Context
Scale-down operations may interrupt running workloads. Need grace period to allow graceful shutdown.

### Decision
- **Grace Period Field:** Optional duration field (e.g., "15m")
- **Application:** Only applies to scale-down (active to inactive transition)
- **State:** Introduce "GracePeriod" state between Active and Inactive
- **Behavior:** During grace period, maintain activeReplicas but show GracePeriod status
- **Maximum:** 60 minutes enforced by validation
- **Default:** 0 (immediate scale-down if not specified)

### Consequences
**Positive:**
- Predictable delay before scale-down
- Status clearly indicates grace period state
- Metrics can track grace period duration
- Zero default preserves simple use cases

**Negative:**
- Adds complexity to state machine
- Grace period may extend beyond next window start (overlap handling needed)
- Does not guarantee workload actually shuts down gracefully

**Alternatives Considered:**
- No grace period: Too aggressive for stateful workloads
- Pod-level graceful termination only: Not visible in operator logic
- Percentage-based: Unclear semantics, harder to predict

### Open Questions
- Should grace period be cancellable if window becomes active again?
- Do we need separate grace periods for different target kinds?

---

## ADR-0005: Implementation Priorities for v0.1 Completion

**Date:** 2025-11-23
**Status:** Accepted
**Deciders:** kyklos-orchestrator

### Context
Project is 83.8% complete with core engine working but missing critical integration components. Need to prioritize remaining work to reach v0.1 release by 2025-11-27 23:59 IST.

### Decision
Implement in this order:
1. **Grace period timing logic** - Core scaling behavior, highest user value
2. **Holiday ConfigMap reading** - Key differentiator feature
3. **Prometheus metrics** - Required for production observability
4. **E2E test scenarios** - Quality gate for release
5. **Integration verification** - Manual validation of all components
6. **Documentation updates** - Ensure accuracy before release

Timeline: 4 days (2025-11-24 to 2025-11-27)

### Consequences
**Positive:**
- Clear task sequence prevents paralysis
- Daily checkpoints enable early risk detection
- Metrics implementation ensures production readiness
- E2E tests provide confidence for release

**Negative:**
- Aggressive timeline may require scope cuts
- No buffer for major blockers
- Risk of incomplete testing if tasks slip

**Mitigations:**
- Daily progress checkpoints at 19:00-23:00 IST
- Minimum viable: Tasks 1-3 complete, 4-6 can slip to v0.1.1
- Simplest implementations first (iterate in minor versions)
- Comprehensive risk register in implementation plan

**Alternatives Considered:**
- Parallel implementation: Risk of integration issues
- Focus on tests first: No code to test
- Skip metrics: Unacceptable for production use

### Open Questions
- Should we add clock mocking for deterministic E2E tests?
- Do we need admission webhook in v0.1 or defer to v0.2?

---

## ADR-0006: Holiday ConfigMap Design

**Date:** 2025-11-23
**Status:** Accepted
**Deciders:** kyklos-orchestrator

### Context
Holiday handling is implemented in engine (3 modes: ignore, closed, open) but controller never sets IsHoliday=true. Need ConfigMap format and lookup logic.

### Decision
**ConfigMap Format:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: company-holidays
data:
  "2025-01-01": "New Year"
  "2025-07-04": "Independence Day"
  "2025-12-25": "Christmas"
```

**Lookup Logic:**
1. Check if spec.holidays.sourceRef.name is set
2. Fetch ConfigMap from TWS namespace (same namespace as TWS)
3. Format today's date as YYYY-MM-DD in TWS timezone
4. Check if date exists as key in ConfigMap data
5. Return true if exists, false otherwise
6. On error (ConfigMap not found), set Degraded condition and continue with IsHoliday=false

**RBAC:** Add ConfigMap get/list permissions to controller ClusterRole

### Consequences
**Positive:**
- Simple, declarative format
- Standard Kubernetes primitive
- Easy to update via kubectl/GitOps
- Human-readable date format
- Values can be descriptive names

**Negative:**
- No validation of date format in ConfigMap (malformed dates ignored)
- No holiday recurrence patterns (must list each year)
- ConfigMap must exist in same namespace as TWS

**Mitigations:**
- Document date format clearly in CRD spec comments
- Add validation example in user guide
- Consider admission webhook in v0.2 for date validation

**Alternatives Considered:**
- JSON array in annotation: Less GitOps-friendly
- Separate Holiday CRD: Over-engineered for v0.1
- External calendar API: Increases complexity and failure modes

### Open Questions
- Should we support cross-namespace ConfigMap references?
- Do we need ConfigMap watch to detect holiday changes immediately?

---

## ADR-0007: Prometheus Metrics Design

**Date:** 2025-11-23
**Status:** Accepted
**Deciders:** kyklos-orchestrator

### Context
Controller needs observability for production use. Must expose metrics for scaling operations, current state, and performance.

### Decision
**Metrics to Implement:**

1. `kyklos_scale_operations_total` (counter)
   - Labels: tws_name, tws_namespace, direction (up/down), result (success/error)
   - Tracks all scaling attempts

2. `kyklos_current_effective_replicas` (gauge)
   - Labels: tws_name, tws_namespace
   - Current desired replica count for each TWS

3. `kyklos_window_transitions_total` (counter)
   - Labels: tws_name, tws_namespace, from_window, to_window
   - Tracks window changes (e.g., "Default" -> "BusinessHours")

4. `kyklos_reconcile_duration_seconds` (histogram)
   - Labels: tws_name, tws_namespace
   - Time spent in reconcile loop

**Implementation:**
- Use prometheus/client_golang
- Register metrics in init() or SetupWithManager()
- Instrument at key points in reconcile loop
- Follow Prometheus best practices (counter for events, gauge for state)

### Consequences
**Positive:**
- Standard Prometheus format
- Low cardinality labels (namespace/name only)
- Covers scaling operations, state, and performance
- Enables alerting on scale failures or stuck reconciliation

**Negative:**
- Increases reconcile loop complexity slightly
- Adds dependency on prometheus client library
- Metrics persist after TWS deletion (until controller restart)

**Mitigations:**
- Keep instrumentation simple and focused
- Document metric lifecycle in operations guide
- Consider metric cleanup in future version

**Alternatives Considered:**
- OpenTelemetry: Too heavy for v0.1
- Custom metrics endpoint: Reinventing wheel
- No metrics: Unacceptable for production

### Open Questions
- Should we expose grace period remaining as gauge?
- Do we need holiday state metric?

---

## Template for Future ADRs

```markdown
## ADR-XXXX: Title

**Date:** YYYY-MM-DD
**Status:** Proposed | Accepted | Deprecated | Superseded
**Deciders:** agent-id(s)

### Context
What is the issue we're trying to solve?

### Decision
What did we decide?

### Consequences
**Positive:**
**Negative:**
**Mitigations:**

**Alternatives Considered:**

### Open Questions
```
