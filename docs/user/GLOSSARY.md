# Glossary

## Core Terms

### TimeWindowScaler (TWS)
A Kubernetes custom resource that defines time-based scaling rules for a target workload. The controller watches TimeWindowScalers and adjusts target replica counts based on time windows, holidays, and calendars.

### Effective Replicas
The replica count Kyklos wants the target deployment to have right now, computed from the current time, active windows, and grace period logic. Shown in `status.effectiveReplicas`.

### Target
The Kubernetes resource being scaled by Kyklos. In v0.1, only Deployments are supported. Referenced via `spec.targetRef`.

### Default Replicas
The replica count used when no time windows match the current time. Set via `spec.defaultReplicas`. This is the "baseline" or "outside hours" capacity.

## Time Windows

### Window
A scheduled time period with a specific replica count. Defined by days of week, start time, end time, and desired replicas.

### Start Time (Inclusive)
The beginning of a time window. At exactly this time, the window becomes active. Format: `HH:MM` (24-hour).

### End Time (Exclusive)
The boundary where a time window ends. The moment before this time is the last instant the window is active. Format: `HH:MM` (24-hour).

### Cross-Midnight Window
A time window where the end time is before the start time (e.g., 22:00-06:00), causing it to span midnight into the next calendar day.

### Window Boundary
A point in time when a window starts or ends. The controller schedules reconciliation at the next boundary to minimize API calls.

### Current Window
The label identifying which window or default state is active right now. Shown in `status.currentWindow`. Values include `BusinessHours`, `OffHours`, or `Custom-<hash>`.

### Last Matching Window
When multiple windows match the current time, the last one in the `spec.windows` array determines the replica count. This provides explicit precedence control.

## Reconciliation

### Reconcile Loop
The controller's main workflow that reads a TimeWindowScaler, computes the desired replica count, updates the target if needed, and schedules the next reconcile.

### Observed Generation
The value of `metadata.generation` that was last processed by the controller. When `status.observedGeneration` matches `metadata.generation`, the status is current.

### Requeue
Scheduling the next reconcile at a specific time. Kyklos requeues at window boundaries plus jitter to minimize unnecessary API calls.

### Manual Drift
When a deployment's replica count differs from what Kyklos expects because of manual scaling operations (e.g., `kubectl scale`). Kyklos detects and corrects drift unless paused.

## Status Conditions

### Ready Condition
Indicates whether the target deployment matches the desired replica count. `True` when aligned, `False` when mismatched or target not found.

### Reconciling Condition
Indicates an ongoing state change. `True` during window transitions or configuration changes, `False` when stable.

### Degraded Condition
Indicates configuration or operational problems. `True` for invalid timezones, missing holiday ConfigMaps, or other errors. `False` during normal operation.

## Scaling Behavior

### Grace Period
An optional delay before downscaling. Configured via `spec.gracePeriodSeconds`. Only applies when reducing replicas (not scale-up). During the grace period, the controller maintains current replicas but shows "grace-period-active" reason in status. The next boundary is set to LastScaleTime + gracePeriodSeconds. Maximum allowed: 3600 seconds (1 hour). Useful for connection draining or workload completion.

### Pause
A mode where Kyklos computes the desired state and updates status but does not modify the target deployment. Set via `spec.pause: true`. Used for manual overrides or testing configurations.

### Scale-Up
Increasing the replica count. Always happens immediately (no grace period).

### Scale-Down
Decreasing the replica count. Subject to grace period if configured.

### Drift Correction
The process of reverting manual scaling changes to match the time-window-based desired state. Emits a `DriftCorrected` event.

## Timezones and Calendars

### IANA Timezone
A timezone identifier from the IANA Time Zone Database (e.g., `America/New_York`, `Asia/Kolkata`, `UTC`). Used for all time calculations with full DST support.

### DST (Daylight Saving Time)
The practice of setting clocks forward in spring and back in fall. Kyklos handles DST automatically using IANA timezone rules.

### Local Time
The time in the TimeWindowScaler's configured timezone after DST adjustment. All window matching happens in local time.

### Holiday
A date defined in a ConfigMap that modifies normal window behavior based on the configured holiday mode.

### Holiday Mode
How holidays affect window matching:
- **ignore** - Normal window processing (default)
- **treat-as-closed** - All windows ignored, uses `defaultReplicas`
- **treat-as-open** - Uses `max(all window replicas)`

### Holiday ConfigMap
A ConfigMap with ISO date keys (YYYY-MM-DD) indicating which dates are holidays. Referenced via `spec.holidays.sourceRef.name`.

## Controller Components

### Controller Manager
The Kubernetes controller process that watches TimeWindowScalers and reconciles them. Runs as a Deployment in the `kyklos-system` namespace.

### Controller-Runtime
The library used to build the Kyklos controller, providing watch mechanisms, caching, and reconcile queue management.

### Metrics Endpoint
An HTTP endpoint (port 8080, path `/metrics`) exposing Prometheus metrics about controller health and scaling operations.

### Health Endpoints
HTTP endpoints for liveness (`/healthz`) and readiness (`/readyz`) probes on port 8081.

## Kubernetes Resources

### CRD (Custom Resource Definition)
The schema that defines the TimeWindowScaler API. Installed cluster-wide and enables creating TimeWindowScaler resources.

### Deployment
A Kubernetes workload resource that manages a replicated set of pods. The target type supported by Kyklos in v0.1.

### Event
A Kubernetes resource recording state changes or notable occurrences. Kyklos emits events like `ScaledUp`, `ScaledDown`, and `DriftCorrected`.

### ClusterRole
A set of permissions defined cluster-wide. Kyklos uses a ClusterRole to grant the controller access to deployments, TimeWindowScalers, events, and ConfigMaps.

### ServiceAccount
A Kubernetes identity for the controller pod. The Kyklos controller runs as the `kyklos-controller` ServiceAccount in `kyklos-system`.

## Observability

### Structured Logs
JSON-formatted log entries with consistent key names, enabling easy filtering and aggregation. Kyklos logs include fields like `tws`, `namespace`, `effectiveReplicas`, and `action`.

### Log Level
The severity of a log message:
- **Info** - Normal operations (scaling, window transitions)
- **Warning** - Recoverable issues (target not found, drift detected)
- **Error** - Problems requiring attention (invalid timezone, update failures)
- **Debug** - Detailed troubleshooting information

### Metric
A time-series measurement exposed via Prometheus format. Examples: `kyklos_scale_operations_total`, `kyklos_reconcile_duration_seconds`.

### Label
A key-value pair attached to metrics for filtering and aggregation. Example: `direction="up"` on scale operation metrics.

## Configuration Fields

### targetRef
The reference to the Kubernetes resource being scaled. Contains `apiVersion`, `kind`, `name`, and optional `namespace`.

### timezone
The IANA timezone identifier used for all time calculations. Required field.

### windows
An array of time window definitions. Each contains `days`, `start`, `end`, and `replicas`.

### days
The days of week when a window applies. Array of three-letter abbreviations: `Mon`, `Tue`, `Wed`, `Thu`, `Fri`, `Sat`, `Sun`.

### replicas
The desired replica count during a specific window or as the default outside all windows.

### gracePeriodSeconds
Optional delay in seconds before applying downscale operations. Only affects replica decreases.

### pause
Boolean flag to suspend target modifications while maintaining status updates. Set to `true` for manual overrides.

### holidays.mode
How to handle dates listed in the holiday ConfigMap: `ignore`, `treat-as-closed`, or `treat-as-open`.

### holidays.sourceRef.name
The name of the ConfigMap containing holiday dates as keys.

## Status Fields

### status.currentWindow
Label identifying the active window or `OffHours` if no windows match.

### status.effectiveReplicas
The replica count Kyklos wants right now based on time and window evaluation.

### status.targetObservedReplicas
The actual replica count of the target deployment as last observed.

### status.lastScaleTime
RFC3339 timestamp of the most recent scaling operation.

### status.observedGeneration
The `metadata.generation` value that was last processed. Used to detect stale status.

### status.conditions
An array of standard Kubernetes condition objects indicating health, reconciliation state, and degradation.

## Operational Terms

### Cluster-wide Installation
Deployment mode where the controller has permissions across all namespaces. Recommended for production.

### Namespaced Installation
Deployment mode where the controller is limited to specific namespaces. Used in multi-tenant clusters.

### RBAC (Role-Based Access Control)
Kubernetes permission system. Kyklos requires permissions to read TimeWindowScalers, update Deployments, create Events, and read ConfigMaps.

### Version Skew
The difference between controller version and CRD version. Kyklos requires controller version >= CRD version.

### Upgrade
Updating the Kyklos controller to a new version. Patch and minor version upgrades require no downtime.

### Rollback
Reverting to a previous Kyklos controller version. Supported within the same minor version (e.g., v0.1.3 → v0.1.2).

### QPS (Queries Per Second)
The rate at which the controller makes API requests to Kubernetes. Configurable via `--kube-api-qps` flag.

### Rate Limiting
Kubernetes API server throttling to prevent excessive requests. Controller respects rate limits and backs off when throttled.

## Development Terms

### Make Targets
Predefined commands in the Makefile for common tasks like `make build`, `make deploy`, `make test`. See [MAKE-TARGETS.md](../MAKE-TARGETS.md).

### Kind Cluster
A local Kubernetes cluster running in Docker, used for development and testing.

### Controller-gen
A tool that generates CRD manifests and RBAC rules from Go type definitions and kubebuilder markers.

### Envtest
A testing framework that starts a local Kubernetes control plane for integration testing without a full cluster.

### Demo Namespace
The `demo` namespace used in quick-start guides and the minute-scale demo. Created by `make demo-setup`.

## Related Terms

### HPA (Horizontal Pod Autoscaler)
Kubernetes built-in autoscaler that scales based on metrics (CPU, memory, custom metrics). Kyklos is complementary, focusing on time-based patterns.

### Cron
Unix time-based job scheduler. Unlike cron-triggered scaling, Kyklos proactively maintains replica counts and corrects drift.

### VPA (Vertical Pod Autoscaler)
Kubernetes autoscaler that adjusts container resource requests/limits. Orthogonal to Kyklos.

### PDB (Pod Disruption Budget)
Kubernetes resource limiting voluntary disruptions during rolling updates or scaling. Works with Kyklos-managed deployments.

### SLO (Service Level Objective)
Performance or availability target. Kyklos helps meet SLOs by ensuring adequate capacity during peak times.

## Acronyms

- **TWS** - TimeWindowScaler
- **DST** - Daylight Saving Time
- **IANA** - Internet Assigned Numbers Authority (maintains timezone database)
- **CRD** - Custom Resource Definition
- **RBAC** - Role-Based Access Control
- **HPA** - Horizontal Pod Autoscaler
- **VPA** - Vertical Pod Autoscaler
- **PDB** - Pod Disruption Budget
- **SLO** - Service Level Objective
- **QPS** - Queries Per Second
- **IST** - India Standard Time (example timezone)
- **UTC** - Coordinated Universal Time

## See Also

- **[Concepts](CONCEPTS.md)** - Detailed explanations of core concepts
- **[FAQ](FAQ.md)** - Common questions with quick answers
- **[API Reference](../api/CRD-SPEC.md)** - Complete field documentation
- **[Operations Guide](OPERATIONS.md)** - Production metrics and monitoring
