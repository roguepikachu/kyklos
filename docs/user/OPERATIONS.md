# Operations Guide

This guide covers running Kyklos in production environments.

## Installation Modes

### Cluster-wide (Recommended)

The controller runs with cluster-wide permissions and can manage TimeWindowScalers in any namespace.

**Install:**
```bash
kubectl apply -f https://github.com/your-org/kyklos/releases/v0.1.0/install.yaml
```

This creates:
- Namespace: `kyklos-system`
- ClusterRole: Permissions for deployments, events, TimeWindowScalers
- ServiceAccount and ClusterRoleBinding
- Deployment: Single replica controller

**Resource requirements:**
- CPU: 100m request, 200m limit
- Memory: 128Mi request, 256Mi limit

### Namespaced

The controller runs with permissions limited to specific namespaces.

**Use case:** Multi-tenant clusters where cluster-wide permissions are restricted.

**Install:**
```bash
# Create in target namespace
kubectl create namespace production
kubectl apply -f install-namespaced.yaml -n production
```

**Limitations:**
- Must deploy controller per namespace
- Cannot watch TimeWindowScalers across namespaces
- Higher resource overhead with multiple controllers

## Key Metrics

Kyklos exposes Prometheus metrics on port 8080 at `/metrics`.

### Controller Health Metrics

| Metric | Type | Description | Alert Threshold |
|--------|------|-------------|-----------------|
| `kyklos_controller_up` | Gauge | Controller running (1=up, 0=down) | `< 1 for 2m` |
| `kyklos_reconcile_duration_seconds` | Histogram | Time to complete reconcile loop | `p99 > 5s` |
| `kyklos_reconcile_errors_total` | Counter | Failed reconciliations | `rate > 0.5/min` |
| `workqueue_depth` | Gauge | Pending reconcile requests | `> 50` |
| `workqueue_unfinished_work_seconds` | Gauge | Time items spend in queue | `> 30s` |

### Scaling Operations Metrics

| Metric | Type | Description | Alert Threshold |
|--------|------|-------------|-----------------|
| `kyklos_scale_operations_total{direction="up"}` | Counter | Successful scale-ups | Rate analysis |
| `kyklos_scale_operations_total{direction="down"}` | Counter | Successful scale-downs | Rate analysis |
| `kyklos_scale_failures_total` | Counter | Failed scale operations | `> 0` |
| `kyklos_effective_replicas` | Gauge | Current desired replicas per TWS | Informational |
| `kyklos_target_observed_replicas` | Gauge | Actual target replicas | Compare with effective |

### Window and Time Metrics

| Metric | Type | Description | Alert Threshold |
|--------|------|-------------|-----------------|
| `kyklos_window_transitions_total` | Counter | Window boundary crossings | Rate analysis |
| `kyklos_grace_periods_active` | Gauge | Active grace periods | Informational |
| `kyklos_holiday_overrides_total` | Counter | Holiday mode activations | Informational |
| `kyklos_manual_drift_corrections_total` | Counter | Manual scaling reverted | `rate > 1/hour` |

### Configuration and Status Metrics

| Metric | Type | Description | Alert Threshold |
|--------|------|-------------|-----------------|
| `kyklos_tws_total{namespace}` | Gauge | Total TimeWindowScalers | Informational |
| `kyklos_degraded_tws_total` | Gauge | TWS in Degraded state | `> 0` |
| `kyklos_paused_tws_total` | Gauge | TWS with pause=true | Informational |

## Suggested Alerts

### Critical Alerts

**Controller Down:**
```yaml
alert: KyklosControllerDown
expr: kyklos_controller_up == 0
for: 2m
severity: critical
annotations:
  summary: "Kyklos controller is not running"
  description: "No scaling operations will occur"
```

**Scale Operations Failing:**
```yaml
alert: KyklosScaleFailures
expr: rate(kyklos_scale_failures_total[5m]) > 0
for: 5m
severity: critical
annotations:
  summary: "Kyklos cannot scale deployments"
  description: "{{ $value }} scale operations failing per second"
```

### Warning Alerts

**High Reconcile Latency:**
```yaml
alert: KyklosSlowReconcile
expr: histogram_quantile(0.99, kyklos_reconcile_duration_seconds) > 5
for: 10m
severity: warning
annotations:
  summary: "Kyklos reconcile loop is slow"
  description: "P99 reconcile time: {{ $value }}s"
```

**Degraded TimeWindowScalers:**
```yaml
alert: KyklosDegradedResources
expr: kyklos_degraded_tws_total > 0
for: 15m
severity: warning
annotations:
  summary: "{{ $value }} TimeWindowScalers in Degraded state"
  description: "Check invalid configurations or missing targets"
```

**Excessive Manual Drift:**
```yaml
alert: KyklosFrequentDriftCorrection
expr: rate(kyklos_manual_drift_corrections_total[1h]) > 1
for: 30m
severity: warning
annotations:
  summary: "Frequent manual scaling conflicts detected"
  description: "{{ $value }} drift corrections per hour"
```

**Deep Work Queue:**
```yaml
alert: KyklosWorkQueueBacklog
expr: workqueue_depth{name="timewindowscaler"} > 50
for: 5m
severity: warning
annotations:
  summary: "Kyklos work queue is backing up"
  description: "{{ $value }} items pending reconciliation"
```

### Info Alerts

**High Pause Rate:**
```yaml
alert: KyklosManyPaused
expr: kyklos_paused_tws_total / kyklos_tws_total > 0.3
for: 1h
severity: info
annotations:
  summary: "Over 30% of TimeWindowScalers are paused"
  description: "{{ $value }} paused resources, check if intentional"
```

## Health Endpoints

### Liveness Probe

**Endpoint:** `http://localhost:8081/healthz`

**Returns:**
- `200 OK` if controller manager is running
- `500 Internal Server Error` if unhealthy

**Kubernetes config:**
```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8081
  initialDelaySeconds: 15
  periodSeconds: 20
```

### Readiness Probe

**Endpoint:** `http://localhost:8081/readyz`

**Returns:**
- `200 OK` if controller is ready to reconcile
- `503 Service Unavailable` if not ready (caches syncing, leader election pending)

**Kubernetes config:**
```yaml
readinessProbe:
  httpGet:
    path: /readyz
    port: 8081
  initialDelaySeconds: 5
  periodSeconds: 10
```

## Upgrades

### In-Place Upgrade

**For patch and minor versions** (e.g., v0.1.0 → v0.1.1):

```bash
# Update controller image
kubectl set image deployment/kyklos-controller-manager \
  manager=kyklos/controller:v0.1.1 \
  -n kyklos-system

# Wait for rollout
kubectl rollout status deployment/kyklos-controller-manager -n kyklos-system
```

**Downtime:** None. Rolling update with zero downtime.

**Compatibility:** v1alpha1 API remains stable across v0.x releases.

### CRD Update Required

**For releases with CRD schema changes:**

```bash
# Apply new CRDs first
kubectl apply -f https://github.com/your-org/kyklos/releases/v0.2.0/crds.yaml

# Then update controller
kubectl set image deployment/kyklos-controller-manager \
  manager=kyklos/controller:v0.2.0 \
  -n kyklos-system
```

**Check release notes** for breaking changes and migration steps.

### Version Skew Policy

**Controller and CRD:**
- Controller version must be >= CRD version
- CRD version must be >= any TWS resource version

**Supported skew:**
- Controller can be 1 minor version ahead of CRDs
- CRDs must match or precede controller version

**Example:**
- Controller v0.2.0 + CRD v0.1.0 = supported
- Controller v0.1.0 + CRD v0.2.0 = unsupported (upgrade controller)

## Resource Limits

### Controller Pod

**Recommended production limits:**

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

**Scaling guidance:**

| TimeWindowScalers | CPU Request | Memory Request |
|-------------------|-------------|----------------|
| 1-50 | 100m | 128Mi |
| 51-200 | 200m | 256Mi |
| 201-500 | 500m | 512Mi |
| 500+ | Consider multiple controllers |

### Rate Limiting

Kyklos respects Kubernetes API rate limits:
- Default: 5 QPS burst 10
- Increase for large deployments

**Configure via controller args:**
```yaml
args:
- --kube-api-qps=10
- --kube-api-burst=20
```

## Common Operator Tasks

### View All TimeWindowScalers

```bash
kubectl get tws --all-namespaces
```

### Check Status of Specific TWS

```bash
kubectl get tws my-scaler -n production -o yaml
```

Look at `status.conditions` for health:
- `Ready=True` - Operating normally
- `Degraded=True` - Configuration or operational issue
- `Reconciling=True` - Active state change in progress

### View Scaling Events

```bash
kubectl get events -n production \
  --field-selector involvedObject.kind=TimeWindowScaler
```

### Pause All Scaling Temporarily

```bash
# Pause a specific TWS
kubectl patch tws my-scaler -n production \
  --type=merge -p '{"spec":{"pause":true}}'

# Pause all in namespace
kubectl get tws -n production -o name | xargs -I {} \
  kubectl patch {} --type=merge -p '{"spec":{"pause":true}}'
```

**Resume:**
```bash
kubectl patch tws my-scaler -n production \
  --type=merge -p '{"spec":{"pause":false}}'
```

### Check Controller Logs

```bash
# Recent logs
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=100

# Follow logs
kubectl logs -n kyklos-system -l app=kyklos-controller -f

# Filter for errors
kubectl logs -n kyklos-system -l app=kyklos-controller \
  | grep -i error
```

### View Metrics

```bash
# Port-forward to metrics endpoint
kubectl port-forward -n kyklos-system \
  deployment/kyklos-controller-manager 8080:8080

# In another terminal, query metrics
curl http://localhost:8080/metrics | grep kyklos
```

### Validate TimeWindowScaler Before Apply

```bash
# Dry-run
kubectl apply -f my-scaler.yaml --dry-run=server

# Check for validation errors
kubectl apply -f my-scaler.yaml --validate=true
```

### Force Immediate Reconciliation

Delete the controller pod to trigger immediate reconcile on restart:

```bash
kubectl delete pod -n kyklos-system -l app=kyklos-controller
```

New pod reconciles all TimeWindowScalers on startup.

### Backup TimeWindowScalers

```bash
# Export all TWS resources
kubectl get tws --all-namespaces -o yaml > tws-backup.yaml

# Restore
kubectl apply -f tws-backup.yaml
```

## Performance Tuning

### Reduce Reconcile Frequency

If controller is over-utilized, increase minimum requeue time:

**Default:** 30 seconds minimum between reconciles

**Increase (requires code change):**
```go
const MinRequeueAfter = 60 * time.Second
```

### Optimize Window Count

Fewer windows = faster evaluation.

**Inefficient:**
```yaml
windows:
- days: [Mon]
  start: "09:00"
  end: "17:00"
  replicas: 10
- days: [Tue]
  start: "09:00"
  end: "17:00"
  replicas: 10
# ... repeat for each day
```

**Efficient:**
```yaml
windows:
- days: [Mon, Tue, Wed, Thu, Fri]
  start: "09:00"
  end: "17:00"
  replicas: 10
```

### Cache Configuration

Controller uses client-go caching. For large clusters:

**Increase cache size:**
```yaml
args:
- --cache-sync-timeout=5m
```

## Security Considerations

### RBAC Permissions

The controller requires these permissions:

**TimeWindowScalers:**
- `get`, `list`, `watch` - Read resources
- `update`, `patch` - Update status

**Deployments:**
- `get`, `list`, `watch` - Monitor targets
- `update`, `patch` - Scale targets

**Events:**
- `create`, `patch` - Emit events

**ConfigMaps:**
- `get`, `list`, `watch` - Read holiday data

### Network Policies

If using network policies, allow:

**Egress from controller:**
- Kubernetes API server (typically 443)

**Ingress to controller:**
- Metrics port 8080 (from Prometheus)
- Health port 8081 (from kubelet)

**Example policy:**
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kyklos-controller
  namespace: kyklos-system
spec:
  podSelector:
    matchLabels:
      app: kyklos-controller
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443
```

### Pod Security

**Controller runs as:**
- Non-root user (UID 65532)
- Read-only root filesystem
- No privilege escalation
- Dropped all capabilities

**Pod Security Standard:** Baseline compliant.

## Disaster Recovery

### Controller Failure

**Symptom:** Controller pod not running.

**Impact:** No scaling operations occur. Existing deployments remain at current replica count.

**Recovery:**
```bash
# Check pod status
kubectl get pods -n kyklos-system

# Check events
kubectl get events -n kyklos-system

# Force restart
kubectl delete pod -n kyklos-system -l app=kyklos-controller
```

### CRD Deletion

**Symptom:** TimeWindowScalers deleted.

**Impact:** All TWS resources deleted. Deployments remain at last scaled replica count.

**Recovery:**
```bash
# Reinstall CRD
kubectl apply -f crds.yaml

# Restore TWS resources from backup
kubectl apply -f tws-backup.yaml
```

### Holiday ConfigMap Missing

**Symptom:** `Degraded=True` condition with `HolidaySourceMissing`.

**Impact:** Controller treats every day as non-holiday (mode reverts to `ignore`).

**Recovery:**
```bash
# Recreate ConfigMap
kubectl create configmap company-holidays -n production \
  --from-literal='2025-12-25'='Christmas'
```

## Monitoring Dashboards

### Grafana Dashboard

**Key panels:**

1. **Controller Health**
   - Uptime
   - Reconcile rate
   - Error rate

2. **Scaling Activity**
   - Scale operations (up/down) over time
   - Current effective replicas by TWS
   - Drift corrections

3. **Performance**
   - Reconcile duration (p50, p95, p99)
   - Work queue depth
   - API call rate

4. **Configuration**
   - Total TWS count
   - Degraded count
   - Paused count

### Sample Queries

**Reconcile rate:**
```promql
rate(kyklos_reconcile_total[5m])
```

**Scale-up frequency:**
```promql
rate(kyklos_scale_operations_total{direction="up"}[1h])
```

**Drift correction rate by TWS:**
```promql
rate(kyklos_manual_drift_corrections_total[1h]) > 0
```

**Effective replicas timeline:**
```promql
kyklos_effective_replicas
```

## Troubleshooting Production Issues

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for detailed symptom-based solutions.

**Quick checks:**
1. Controller pod running: `kubectl get pods -n kyklos-system`
2. Recent errors: `kubectl logs -n kyklos-system -l app=kyklos-controller | grep ERROR`
3. TWS status: `kubectl get tws --all-namespaces`
4. Metrics: `curl http://localhost:8080/metrics` (after port-forward)

## Further Reading

- [Concepts](CONCEPTS.md) - Understanding window matching and computation
- [FAQ](FAQ.md) - Common questions about behavior
- [Troubleshooting](TROUBLESHOOTING.md) - Symptom-based issue resolution
- [API Reference](../api/CRD-SPEC.md) - Complete CRD specification
