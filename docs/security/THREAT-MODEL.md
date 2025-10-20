# Threat Model

## Scope
This threat model covers the Kyklos TimeWindowScaler controller v0.1 operating within a Kubernetes cluster. It identifies potential attack vectors, trust boundaries, and mitigations.

**In Scope**:
- Kyklos controller pod and RBAC permissions
- TimeWindowScaler custom resources
- Target Deployments/StatefulSets
- ConfigMap-based holiday configuration
- Controller-to-API-server communication

**Out of Scope**:
- Kubernetes control plane vulnerabilities
- Container runtime exploits (containerd, CRI-O)
- etcd encryption and backup security
- Network CNI plugin vulnerabilities
- Node-level compromise

---

## System Overview

### Components
1. **Kyklos Controller**: Go binary running in pod, watches TimeWindowScaler CRs, modifies Deployment replicas
2. **Kubernetes API Server**: Trusted intermediary for all resource operations
3. **TimeWindowScaler CRDs**: User-defined time windows and scaling configurations
4. **Target Workloads**: Deployments or StatefulSets being scaled
5. **Holiday ConfigMaps**: Optional calendar data for holiday-aware scaling
6. **RBAC**: Roles and bindings limiting controller permissions

### Trust Boundaries
- **Boundary 1**: User → Kubernetes API (authenticated via kubeconfig/token)
- **Boundary 2**: Kubernetes API → Controller pod (authenticated via ServiceAccount token)
- **Boundary 3**: Controller pod → Target workloads (authorized via RBAC)
- **Boundary 4**: Controller pod → ConfigMaps (read-only, RBAC-enforced)

### Data Flow
```
User creates TimeWindowScaler CR
    ↓
API server validates and stores CR
    ↓
Controller watches CR via informer
    ↓
Controller computes desired replicas
    ↓
Controller patches target Deployment spec.replicas
    ↓
ReplicaSet controller scales pods
```

---

## Threat Categories

### T1: Privilege Escalation
**Attack Vector**: Attacker gains access to controller pod or ServiceAccount token and uses excessive RBAC permissions to compromise cluster.

**Scenarios**:
- **T1.1**: Overbroad RBAC grants controller access to Secrets, allowing credential theft
- **T1.2**: Controller has `delete` permissions on Deployments, enabling workload destruction
- **T1.3**: Controller has wildcard (`*`) permissions on core API groups
- **T1.4**: Attacker modifies RBAC bindings if controller has self-modification rights

**Impact**: High. Compromised controller could read Secrets, delete workloads, or create malicious pods.

**Mitigations**:
- **M1.1**: RBAC grants only `get`, `list`, `watch`, `patch` on Deployments (no `delete`, no `create`)
- **M1.2**: No Secrets access granted (see RBAC-MATRIX.md)
- **M1.3**: No wildcard permissions in any API group or verb
- **M1.4**: No access to RBAC resources (roles, rolebindings)
- **M1.5**: ServiceAccount token projected with 1-hour expiration
- **M1.6**: NetworkPolicy blocks egress except API server (prevents exfiltration)

**Residual Risk**: Low. If controller pod is compromised, attacker can only modify replica counts of Deployments.

---

### T2: Denial of Service
**Attack Vector**: Attacker causes controller to consume excessive resources or disrupt target workloads.

**Scenarios**:
- **T2.1**: Malicious TimeWindowScaler CR with rapid window transitions (e.g., every 1 minute)
- **T2.2**: Hot-looping reconcile logic overwhelming API server with PATCH requests
- **T2.3**: Controller memory exhaustion from caching thousands of resources in cluster mode
- **T2.4**: Attacker scales target to 0 replicas via crafted time window

**Impact**: Medium. Controller crash or API server rate limiting. Workload unavailability if scaled to 0.

**Mitigations**:
- **M2.1**: Rate limiting built into controller-runtime (default 10 QPS, burst 100)
- **M2.2**: Reconcile loop idempotency: only PATCH if current != desired (no writes if already correct)
- **M2.3**: Resource limits on controller pod (CPU 500m, memory 256Mi in namespaced mode)
- **M2.4**: Minimum requeue interval of 30 seconds prevents tight loops
- **M2.5**: Jitter (5-25s) in requeue calculations spreads load
- **M2.6**: Admission webhook (future) can reject windows with < 5-minute duration
- **M2.7**: PodDisruptionBudget on target workloads prevents zero availability

**Residual Risk**: Medium. Malicious user with TimeWindowScaler create permissions can disrupt their own workloads but not others' (namespace isolation).

---

### T3: Data Tampering
**Attack Vector**: Attacker modifies TimeWindowScaler CRs, ConfigMaps, or target workloads to cause unintended scaling.

**Scenarios**:
- **T3.1**: Attacker with namespace edit rights modifies TimeWindowScaler to scale production to 0
- **T3.2**: Attacker modifies holiday ConfigMap to inject fake holiday dates
- **T3.3**: Manual Deployment replica change causes drift, controller reverts without notice
- **T3.4**: Time zone poisoning: attacker sets invalid timezone causing controller to use defaultReplicas

**Impact**: Medium. Unintended scaling could cause service disruption or over-provisioning costs.

**Mitigations**:
- **M3.1**: RBAC controls who can create/update TimeWindowScaler CRs (namespace admins only)
- **M3.2**: Audit logging enabled for all TimeWindowScaler modifications
- **M3.3**: Controller emits events when correcting manual drift (observability)
- **M3.4**: Pause field allows safe testing without affecting targets
- **M3.5**: Timezone validation: controller sets Degraded=True for invalid zones, uses safe defaultReplicas
- **M3.6**: ConfigMap changes watched: controller reacts within requeue interval
- **M3.7**: GitOps workflows (ArgoCD/Flux) enforce declarative CR management with change review

**Residual Risk**: Low. Requires namespace-level edit permissions. Audit logs provide forensics.

---

### T4: Information Disclosure
**Attack Vector**: Attacker gains access to sensitive information via controller logs, metrics, or status fields.

**Scenarios**:
- **T4.1**: Controller logs expose target workload names and namespaces (privacy concern in multi-tenant)
- **T4.2**: Metrics endpoint leaks TimeWindowScaler names and replica counts
- **T4.3**: Status fields reveal operational patterns (e.g., when offices are closed)
- **T4.4**: Holiday ConfigMap readable by attacker reveals company calendar

**Impact**: Low. Information is operational metadata, not credentials or user data.

**Mitigations**:
- **M4.1**: Metrics bind to localhost only (127.0.0.1:8080) by default
- **M4.2**: NetworkPolicy blocks ingress to controller pod
- **M4.3**: Logs do not contain user data or credentials
- **M4.4**: ConfigMap RBAC limits read access to controller ServiceAccount and namespace admins
- **M4.5**: Audit logs track who reads ConfigMaps
- **M4.6**: Status fields are namespace-scoped (RBAC enforced)

**Residual Risk**: Low. Attacker needs namespace read permissions to view TimeWindowScaler resources.

---

### T5: Supply Chain Attacks
**Attack Vector**: Attacker compromises controller image, dependencies, or build pipeline.

**Scenarios**:
- **T5.1**: Malicious code injected into controller binary during build
- **T5.2**: Compromised base image (distroless) contains backdoor
- **T5.3**: Vulnerable Go dependency with known CVE
- **T5.4**: Attacker replaces image in registry with malicious version using same tag

**Impact**: Critical. Full cluster compromise possible if malicious controller has RBAC permissions.

**Mitigations**:
- **M5.1**: Image digest pinning (sha256 hash) prevents tag-based poisoning
- **M5.2**: Image signing with cosign or Notary, verified by admission controller
- **M5.3**: SBOM generation for all releases, published alongside image
- **M5.4**: Trivy/Grype scanning in CI/CD pipeline, blocks HIGH/CRITICAL CVEs
- **M5.5**: Dependabot or Renovate for automated dependency updates
- **M5.6**: Use distroless base image (minimal attack surface, no shell)
- **M5.7**: Build pipeline runs on trusted infrastructure with SLSA provenance
- **M5.8**: Multi-stage Docker build discards build tools from final image

**Residual Risk**: Low with all mitigations applied. Medium if image scanning or signing is skipped.

---

### T6: Configuration Errors
**Attack Vector**: Misconfiguration by operator causes security or availability issues.

**Scenarios**:
- **T6.1**: Operator grants cluster-wide RBAC when namespaced mode is intended
- **T6.2**: Pod runs as root due to missing SecurityContext
- **T6.3**: No NetworkPolicy defined, controller can egress to internet
- **T6.4**: Overlapping time windows cause unpredictable scaling
- **T6.5**: Grace period too long (hours) delays critical scale-down

**Impact**: Medium. Security weakening or operational instability.

**Mitigations**:
- **M6.1**: Helm chart templates include SecurityContext by default
- **M6.2**: Helm values clearly document namespaced vs cluster mode
- **M6.3**: Admission webhook (future) validates TimeWindowScaler for overlapping windows
- **M6.4**: Pod Security Admission enforces Restricted standard (rejects root pods)
- **M6.5**: NetworkPolicy example provided in installation docs
- **M6.6**: Pre-flight checks in controller startup log RBAC permissions

**Residual Risk**: Medium. Operators can disable security features if determined.

---

### T7: Insider Threats
**Attack Vector**: Malicious insider with cluster access abuses Kyklos for disruption.

**Scenarios**:
- **T7.1**: Insider creates TimeWindowScaler scaling production to 0 during business hours
- **T7.2**: Insider with RBAC admin rights grants excessive permissions to controller
- **T7.3**: Insider disables audit logging before making malicious changes

**Impact**: High. Insider with cluster admin can bypass all controls.

**Mitigations**:
- **M7.1**: Role-based access control limits who can create TimeWindowScaler CRs
- **M7.2**: Audit logging immutable (sent to external SIEM, not editable in cluster)
- **M7.3**: Change management process for RBAC modifications
- **M7.4**: Alerts on TimeWindowScaler creation/modification in production namespaces
- **M7.5**: GitOps enforces declarative config with PR review
- **M7.6**: Pause field allows testing in non-prod before prod rollout

**Residual Risk**: High. Cluster admin privileges bypass technical controls. Rely on organizational policy.

---

### T8: Time-Based Attacks
**Attack Vector**: Attacker exploits time zone handling or DST transitions.

**Scenarios**:
- **T8.1**: Attacker creates window crossing midnight with incorrect timezone, causing unintended 24-hour window
- **T8.2**: DST transition causes window to skip or double (Spring forward/Fall back)
- **T8.3**: Controller clock skew vs API server causes premature scaling

**Impact**: Low. Scaling happens at wrong time but no security boundary violated.

**Mitigations**:
- **M8.1**: Controller uses IANA timezone database (Go time.LoadLocation) with DST rules
- **M8.2**: All time calculations in local time, converted from UTC
- **M8.3**: Cross-midnight window logic explicitly tested (see CRD-SPEC.md examples)
- **M8.4**: Controller syncs time with API server (trusted time source)
- **M8.5**: Degraded condition set if timezone invalid
- **M8.6**: Admission webhook (future) can warn on suspicious windows (e.g., 23:59-00:01)

**Residual Risk**: Low. DST edge cases are well-tested. Operator can use pause field if uncertain.

---

### T9: API Server Compromise
**Attack Vector**: Attacker compromises Kubernetes API server.

**Scenarios**:
- **T9.1**: Attacker with API server access reads controller ServiceAccount token
- **T9.2**: Attacker modifies TimeWindowScaler CRs directly in etcd
- **T9.3**: Attacker intercepts controller-API communication (MITM)

**Impact**: Critical. Full cluster compromise.

**Mitigations**:
- **M9.1**: TLS encryption for all API server communication (default in Kubernetes)
- **M9.2**: ServiceAccount tokens are short-lived (1 hour expiration)
- **M9.3**: etcd encryption at rest (cluster-level feature)
- **M9.4**: API server audit logging tracks all resource modifications
- **M9.5**: Controller validates API server certificate (trusted CA bundle)

**Residual Risk**: High if API server is compromised. Out of scope for Kyklos.

---

## Attack Trees

### Attacker Goal: Scale Production Workload to 0

```
Scale Production Workload to 0
├─ Modify TimeWindowScaler CR
│  ├─ Gain namespace edit permissions [M3.1: RBAC]
│  └─ Exploit admission webhook bypass [M6.3: future webhook]
├─ Compromise Controller Pod
│  ├─ Exploit container vulnerability [M5.4: image scanning]
│  └─ Steal ServiceAccount token [M9.2: short-lived tokens]
└─ Direct Deployment Modification
   ├─ Gain apps/v1 write permissions [Out of scope: RBAC]
   └─ Manual kubectl scale [M3.3: controller reverts drift]
```

**Highest Risk Path**: Gain namespace edit permissions → Modify TimeWindowScaler

**Primary Defense**: RBAC limiting TimeWindowScaler create/update to trusted users

---

### Attacker Goal: Exfiltrate Data from Controller

```
Exfiltrate Data from Controller
├─ Network Exfiltration
│  ├─ Compromise pod and open outbound connection [M1.6: NetworkPolicy]
│  └─ Use DNS tunneling [M1.6: blocks non-kube-system DNS]
├─ Log Scraping
│  ├─ Read pod logs with kubectl logs [Requires namespace read permission]
│  └─ Access node filesystem [Out of scope: node security]
└─ Metrics Scraping
   └─ Access metrics endpoint [M4.1: localhost bind, M4.2: no ingress]
```

**Highest Risk Path**: Compromise pod → Blocked by NetworkPolicy egress rules

**Primary Defense**: NetworkPolicy restricting egress to API server only

---

## Assumptions and Dependencies

### Assumptions
1. Kubernetes API server is trusted and secured by cluster administrators
2. RBAC is enforced cluster-wide (not disabled)
3. Node OS and container runtime are patched and hardened
4. Network policies are supported by CNI plugin (Calico, Cilium, etc.)
5. Users creating TimeWindowScaler CRs understand scaling implications
6. Audit logging is enabled and monitored

### Dependencies
1. **controller-runtime**: Framework security (watches, caching, leader election)
2. **Kubernetes API server**: Authentication, authorization, admission control
3. **etcd**: CRD storage integrity
4. **CNI Plugin**: NetworkPolicy enforcement
5. **Time synchronization**: NTP on nodes for accurate window calculations

---

## Security Requirements for v0.1

| Requirement | Implemented | Verified By |
|-------------|-------------|-------------|
| Non-root container | Yes | Pod Security Admission |
| Read-only root filesystem | Yes | SecurityContext |
| No Secrets access | Yes | RBAC |
| No wildcard RBAC | Yes | RBAC-MATRIX.md |
| Image digest pinning | Recommended | Helm chart |
| Image scanning | Recommended | CI/CD pipeline |
| NetworkPolicy egress restriction | Recommended | SECURITY-CHECKLIST.md |
| Audit logging | External dependency | Cluster admin |
| Leader election token security | Yes | Projected token with expiration |

---

## Residual Risks for v0.1

### Accepted Risks
1. **No admission webhook**: Overlapping windows and invalid configurations accepted by API server. Validated at runtime by controller.
   - **Why Accepted**: Webhook infrastructure is complex for v0.1. Runtime validation sufficient.
   - **Mitigation Timing**: v0.2 will include validating webhook

2. **Namespace admin can disrupt own workloads**: User with TimeWindowScaler create permissions can scale workloads to 0.
   - **Why Accepted**: Equivalent to directly editing Deployment replicas (same RBAC level).
   - **Mitigation**: RBAC controls limit who has namespace edit rights

3. **No multi-factor approval for scaling operations**: Controller automatically applies scaling without manual approval.
   - **Why Accepted**: Use case is automated scheduling, not sensitive operations.
   - **Mitigation**: Pause field allows review before enabling

4. **No rate limiting on TimeWindowScaler creation**: Attacker could create many CRs causing controller load.
   - **Why Accepted**: Namespace ResourceQuota can limit CR count.
   - **Mitigation Timing**: v0.2 may add admission webhook rate limiting

5. **Hot-loop reconcile risk from bad configuration**: Window transitions every 30 seconds could cause excessive API calls.
   - **Why Accepted**: Controller rate limiting (10 QPS) provides backpressure.
   - **Mitigation**: Admission webhook in v0.2 will enforce minimum window duration

### Monitoring Recommendations
- Alert on reconcile error rate > 10%
- Alert on controller pod restarts
- Alert on RBAC denied errors in logs
- Alert on memory usage > 80% of limit
- Dashboard showing current replicas vs desired replicas per TimeWindowScaler

---

## Threat Model Maintenance

### Review Triggers
- New feature additions (e.g., StatefulSet support in v0.2)
- Newly disclosed CVEs in dependencies
- Kubernetes security advisories
- Customer security incidents
- Annual security review

### Update Process
1. Security team reviews threat model quarterly
2. New threats added with TXXX numbering
3. Mitigations tracked in SECURITY-CHECKLIST.md
4. Residual risks documented with acceptance rationale
5. Version history updated

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-20 | Initial threat model for v0.1 |

---

## References
- STRIDE Threat Modeling: https://learn.microsoft.com/en-us/azure/security/develop/threat-modeling-tool
- Kubernetes Security Best Practices: https://kubernetes.io/docs/concepts/security/
- OWASP Threat Modeling: https://owasp.org/www-community/Threat_Modeling
- NIST Cybersecurity Framework: https://www.nist.gov/cyberframework
- Kyklos RBAC Matrix: `/Users/aykumar/personal/kyklos/docs/security/RBAC-MATRIX.md`
- Kyklos Pod Security Baseline: `/Users/aykumar/personal/kyklos/docs/security/POD-SECURITY-BASELINE.md`
