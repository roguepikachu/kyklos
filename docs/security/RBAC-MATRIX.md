# RBAC Permission Matrix

## Purpose
This document defines the exact RBAC permissions required by the Kyklos controller to perform time-based scaling operations. All permissions follow the principle of least privilege.

## Controller Identity
- **ServiceAccount Name**: `kyklos-controller`
- **Namespace (Namespaced Mode)**: Same namespace as TimeWindowScaler resources
- **Namespace (Cluster Mode)**: `kyklos-system` (or operator-chosen namespace)

## Deployment Modes

### Namespaced Mode
Controller watches and operates only within a single namespace. Requires Role + RoleBinding.

**Use Case**: Multi-tenant environments where operators want namespace-isolated controllers.

**Cache Scope**: Single namespace only. Controller cannot see resources in other namespaces.

### Cluster Mode
Controller watches and operates across all namespaces. Requires ClusterRole + ClusterRoleBinding.

**Use Case**: Platform teams managing workloads across many namespaces from central operator.

**Cache Scope**: Cluster-wide. Controller can see all resources of permitted types.

---

## Namespaced Mode Permissions

### Core Workload Access

| API Group | Resources | Verbs | Rationale |
|-----------|-----------|-------|-----------|
| `apps` | `deployments` | `get`, `list`, `watch` | Read Deployment status to compare current replicas with desired state |
| `apps` | `deployments` | `patch` | Update spec.replicas field only. Uses strategic merge patch to minimize conflict risk |
| `apps` | `deployments/status` | `get` | Read observed replicas and readiness to detect drift |
| `apps` | `statefulsets` | `get`, `list`, `watch` | (Future: v1beta1) Read StatefulSet status for scaling decisions |
| `apps` | `statefulsets` | `patch` | (Future: v1beta1) Update spec.replicas field |
| `apps` | `statefulsets/status` | `get` | (Future: v1beta1) Read observed replicas |

**Why PATCH and not UPDATE?**
- PATCH with strategic merge patch only touches spec.replicas
- Reduces optimistic lock conflicts
- Does not require full Deployment spec round-trip
- More resilient to concurrent modifications by other controllers

**Why READ on /status subresource?**
- Separate read permission follows Kubernetes RBAC best practices
- Allows auditing of status-only reads vs spec modifications
- Required for detecting manual drift (status.replicas != spec.replicas)

### CRD Access

| API Group | Resources | Verbs | Rationale |
|-----------|-----------|-------|-----------|
| `kyklos.io` | `timewindowscalers` | `get`, `list`, `watch` | Primary resources managed by controller. List + watch for cache initialization |
| `kyklos.io` | `timewindowscalers/status` | `patch`, `update` | Update status fields (effectiveReplicas, conditions, observedGeneration). Uses status subresource to avoid spec conflicts |

**Why both PATCH and UPDATE for status?**
- PATCH preferred for incremental status updates (single condition change)
- UPDATE required by some controller-runtime operations
- Both operate on /status subresource only, cannot modify spec

### Configuration Access (Conditional)

| API Group | Resources | Verbs | Rationale | Required When |
|-----------|-----------|-------|-----------|---------------|
| `` (core) | `configmaps` | `get`, `list`, `watch` | Read holiday dates when spec.holidays.sourceRef is configured | Any TimeWindowScaler in namespace has holidays configured |

**Conditional Permission Logic**:
- If NO TimeWindowScaler resources use holidays: omit ConfigMap permissions entirely
- If ANY TimeWindowScaler uses holidays: include ConfigMap read permissions
- Controller must handle missing ConfigMap gracefully (set Degraded condition, continue operation)

**Security Note**: Read-only access to ConfigMaps. Controller never writes or modifies ConfigMaps.

### Event Recording

| API Group | Resources | Verbs | Rationale |
|-----------|-----------|-------|-----------|
| `` (core) | `events` | `create`, `patch` | Emit events for scaling operations (ScaledUp, ScaledDown, etc.) |

**Why CREATE and PATCH?**
- CREATE for new event
- PATCH for updating count/lastTimestamp on deduplicated events
- Follows standard Kubernetes event recorder pattern

**Alternative: No Events Permission**
- Controller will log warnings: "Failed to create event: forbidden"
- Functionality unaffected; events are observability-only
- Consider omitting in high-security environments where event write access is restricted

### Coordination (Leader Election)

| API Group | Resources | Verbs | Rationale |
|-----------|-----------|-------|-----------|
| `coordination.k8s.io` | `leases` | `get`, `create`, `update` | Acquire and renew leader election lease for controller HA |

**Lease Naming**: `kyklos-controller-leader` (single lease per namespace in namespaced mode)

**Alternative: Single-Replica Deployment**
- Omit lease permissions if running single replica without HA
- Set controller flag: `--leader-elect=false`

---

## Cluster Mode Permissions

### Core Workload Access

| API Group | Resources | Verbs | Rationale |
|-----------|-----------|-------|-----------|
| `apps` | `deployments` | `get`, `list`, `watch` | Read Deployments across all namespaces |
| `apps` | `deployments` | `patch` | Update spec.replicas across all namespaces |
| `apps` | `deployments/status` | `get` | Read observed replicas across all namespaces |
| `apps` | `statefulsets` | `get`, `list`, `watch` | (Future) Read StatefulSets across all namespaces |
| `apps` | `statefulsets` | `patch` | (Future) Update spec.replicas across all namespaces |
| `apps` | `statefulsets/status` | `get` | (Future) Read observed replicas across all namespaces |

**Scope**: All namespaces. Controller can target Deployments in any namespace via spec.targetRef.namespace.

### CRD Access

| API Group | Resources | Verbs | Rationale |
|-----------|-----------|-------|-----------|
| `kyklos.io` | `timewindowscalers` | `get`, `list`, `watch` | Watch TimeWindowScaler resources in all namespaces |
| `kyklos.io` | `timewindowscalers/status` | `patch`, `update` | Update status across all namespaces |

### Configuration Access (Conditional)

| API Group | Resources | Verbs | Rationale | Required When |
|-----------|-----------|-------|-----------|---------------|
| `` (core) | `configmaps` | `get`, `list`, `watch` | Read holiday ConfigMaps from any namespace | Any TimeWindowScaler cluster-wide has holidays configured |

**Cross-Namespace ConfigMap Access**:
- Controller can read ConfigMap in namespace specified by TimeWindowScaler
- If TimeWindowScaler in ns-A references ConfigMap in ns-B: requires cluster-wide ConfigMap read
- Recommendation: Store holiday ConfigMaps in controller's own namespace and document convention

### Event Recording

| API Group | Resources | Verbs | Rationale |
|-----------|-----------|-------|-----------|
| `` (core) | `events` | `create`, `patch` | Emit events in namespace of each TimeWindowScaler |

**Multi-Namespace Events**: Events are created in the namespace of the TimeWindowScaler resource, not controller namespace.

### Coordination (Leader Election)

| API Group | Resources | Verbs | Rationale |
|-----------|-----------|-------|-----------|
| `coordination.k8s.io` | `leases` | `get`, `create`, `update` | Acquire cluster-wide leader election lease |

**Lease Naming**: `kyklos-controller-leader` in controller's own namespace (e.g., `kyklos-system`)

---

## What is NOT Allowed

### Secrets Access
**Denied**: No access to `secrets` resource in any API group.

**Rationale**: Controller does not need sensitive data. If future features require credentials (e.g., external metrics), use Workload Identity or mounted service account tokens, not Secret reads.

### Pod Direct Manipulation
**Denied**: No `delete`, `create` on `pods` resource.

**Rationale**: Controller modifies Deployment/StatefulSet spec.replicas. ReplicaSet/StatefulSet controllers handle pod lifecycle.

### Destructive Workload Operations
**Denied**: No `delete` on `deployments` or `statefulsets`.

**Rationale**: Controller never removes target workloads, only scales them.

### Node or Namespace Operations
**Denied**: No access to `nodes`, `namespaces`, `persistentvolumes`.

**Rationale**: Controller operates at workload scope only.

### Admission Control
**Denied**: No `create`, `update` on `validatingwebhookconfigurations` or `mutatingwebhookconfigurations`.

**Rationale**: Webhook configuration is deployment-time operation, not runtime controller responsibility.

### Custom Resource Definitions
**Denied**: No access to `customresourcedefinitions` resource.

**Rationale**: CRD installation is Helm/operator deployment phase, not runtime.

### RBAC Self-Modification
**Denied**: No access to `roles`, `rolebindings`, `clusterroles`, `clusterrolebindings`, `serviceaccounts`.

**Rationale**: Prevents privilege escalation. RBAC is managed by cluster administrators.

### Wildcard Permissions
**Denied**: No use of `*` in apiGroups, resources, or verbs.

**Rationale**: Explicit permission lists only. Each permission must have documented rationale.

---

## Permission Reduction Strategies

### Minimal Events Strategy
If you want to eliminate Event write permissions:

1. Remove `events` permissions from Role/ClusterRole
2. Set controller flag: `--enable-events=false`
3. Events will be logged only (structured logs capture same data)
4. Metrics still available for observability

### No ConfigMap Access
If holidays feature is unused:

1. Remove `configmaps` permissions
2. Document that holiday feature requires manual RBAC extension
3. Controller logs warning if TimeWindowScaler references ConfigMap without permission
4. Sets Degraded=True condition with reason `HolidaySourcePermissionDenied`

### No Leader Election
If running single replica:

1. Remove `leases` permissions
2. Set controller flag: `--leader-elect=false`
3. Deploy with `replicas: 1` in Deployment spec
4. Not recommended for production (no HA)

---

## Migration: Namespaced to Cluster Mode

### Step 1: Create ClusterRole
```plaintext
Create ClusterRole with cluster-wide permissions (see Cluster Mode table)
```

### Step 2: Create ClusterRoleBinding
```plaintext
Bind ClusterRole to existing ServiceAccount in controller namespace
```

### Step 3: Update Controller Deployment
```plaintext
Set environment variable or flag: WATCH_NAMESPACE=""
Empty string = cluster-wide watch
```

### Step 4: Verify Permissions
```plaintext
kubectl auth can-i list deployments --all-namespaces --as=system:serviceaccount:kyklos-system:kyklos-controller
Should return: yes
```

### Step 5: Remove Old RoleBindings
```plaintext
Delete per-namespace Role and RoleBinding resources (if applicable)
```

---

## Migration: Cluster to Namespaced Mode

### Step 1: Choose Target Namespace
```plaintext
Decide which namespace controller will operate in
```

### Step 2: Create Namespace-Scoped Role
```plaintext
Create Role in target namespace with namespaced permissions (see Namespaced Mode table)
```

### Step 3: Create RoleBinding
```plaintext
Bind Role to controller ServiceAccount
```

### Step 4: Update Controller Deployment
```plaintext
Set environment variable: WATCH_NAMESPACE=<target-namespace>
Move controller Deployment to target namespace if not already present
```

### Step 5: Remove ClusterRole
```plaintext
Delete ClusterRole and ClusterRoleBinding
Verify controller logs show: "Watching namespace: <target-namespace>"
```

---

## Validation Commands

### Check Namespaced Permissions
```bash
# Replace with your namespace and service account
NS=production
SA=kyklos-controller

# Deployments
kubectl auth can-i get deployments --namespace=$NS --as=system:serviceaccount:$NS:$SA
kubectl auth can-i patch deployments --namespace=$NS --as=system:serviceaccount:$NS:$SA
kubectl auth can-i delete deployments --namespace=$NS --as=system:serviceaccount:$NS:$SA
# Should return: yes, yes, no

# TimeWindowScalers
kubectl auth can-i get timewindowscalers --namespace=$NS --as=system:serviceaccount:$NS:$SA
kubectl auth can-i patch timewindowscalers/status --namespace=$NS --as=system:serviceaccount:$NS:$SA
# Should return: yes, yes

# ConfigMaps (if holidays enabled)
kubectl auth can-i get configmaps --namespace=$NS --as=system:serviceaccount:$NS:$SA
kubectl auth can-i update configmaps --namespace=$NS --as=system:serviceaccount:$NS:$SA
# Should return: yes, no

# Events
kubectl auth can-i create events --namespace=$NS --as=system:serviceaccount:$NS:$SA
# Should return: yes (or no if events disabled)

# Secrets (should fail)
kubectl auth can-i get secrets --namespace=$NS --as=system:serviceaccount:$NS:$SA
# Should return: no
```

### Check Cluster Permissions
```bash
SA_NAMESPACE=kyklos-system
SA=kyklos-controller

# Cluster-wide Deployments
kubectl auth can-i get deployments --all-namespaces --as=system:serviceaccount:$SA_NAMESPACE:$SA
kubectl auth can-i patch deployments --all-namespaces --as=system:serviceaccount:$SA_NAMESPACE:$SA
# Should return: yes, yes

# Cluster-wide TimeWindowScalers
kubectl auth can-i list timewindowscalers --all-namespaces --as=system:serviceaccount:$SA_NAMESPACE:$SA
# Should return: yes

# Cross-namespace events
kubectl auth can-i create events --namespace=production --as=system:serviceaccount:$SA_NAMESPACE:$SA
# Should return: yes
```

---

## Audit Trail

### RBAC Changes Require
1. Git commit with rationale for permission addition/removal
2. Security review for any new API group access
3. Documentation update in this matrix
4. Validation command verification

### Annual RBAC Review
- Review all permissions against actual controller behavior
- Remove unused permissions (grep codebase for API calls)
- Check for new least-privilege alternatives (e.g., subresource-only access)
- Verify wildcard permissions not introduced

---

## Compliance Mapping

### CIS Kubernetes Benchmark
- **5.1.3**: Minimize wildcard use in Roles and ClusterRoles (PASS: no wildcards)
- **5.1.5**: Ensure default service account is not used (PASS: dedicated ServiceAccount)
- **5.1.6**: Ensure service accounts tokens are only mounted where necessary (PASS: controller pod only)

### Pod Security Standards
- Controller requires no privileged permissions
- No hostPath, hostNetwork, or hostPID access
- Compliant with Restricted Pod Security Standard

---

## Example RBAC Snippets

### Namespaced Role (Core Permissions)
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: kyklos-controller
  namespace: production
rules:
# Deployments
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch", "patch"]
- apiGroups: ["apps"]
  resources: ["deployments/status"]
  verbs: ["get"]

# TimeWindowScalers
- apiGroups: ["kyklos.io"]
  resources: ["timewindowscalers"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["kyklos.io"]
  resources: ["timewindowscalers/status"]
  verbs: ["patch", "update"]

# Events
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create", "patch"]

# Leader election
- apiGroups: ["coordination.k8s.io"]
  resources: ["leases"]
  verbs: ["get", "create", "update"]
```

### Namespaced Role (With Holidays)
```yaml
# Add to above rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch"]
```

### ClusterRole (Cluster Mode)
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kyklos-controller
rules:
# Same rules as Namespaced Role, but applies cluster-wide
# (Copy rules from above examples)
```

---

## Security Implications

### Namespaced Mode Attack Surface
**Risk**: Compromised controller can modify Deployments in single namespace only.

**Blast Radius**: Limited to namespace where controller runs.

**Mitigation**: Use namespaced mode in multi-tenant environments.

### Cluster Mode Attack Surface
**Risk**: Compromised controller can modify Deployments across all namespaces.

**Blast Radius**: Entire cluster.

**Mitigation**:
- Use NetworkPolicy to restrict controller egress
- Enable audit logging for all PATCH operations by controller ServiceAccount
- Monitor for unexpected scaling operations via metrics
- Consider namespace exclusions via validating webhook (future enhancement)

### ConfigMap Permission Risk
**Risk**: Read access to ConfigMaps could expose non-sensitive but private data.

**Mitigation**:
- Store only holiday dates in ConfigMaps (no sensitive data)
- Use dedicated ConfigMap for Kyklos (not shared with other apps)
- Document naming convention: `kyklos-holidays-*`
- Consider future enhancement: dedicated CRD for holidays

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-20 | Initial RBAC matrix for v0.1 |

---

## References
- Kubernetes RBAC Documentation: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
- CIS Kubernetes Benchmark: https://www.cisecurity.org/benchmark/kubernetes
- Kyklos CRD Spec: `/Users/aykumar/personal/kyklos/docs/api/CRD-SPEC.md`
- Kyklos Reconcile Design: `/Users/aykumar/personal/kyklos/docs/design/RECONCILE.md`
