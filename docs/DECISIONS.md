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
