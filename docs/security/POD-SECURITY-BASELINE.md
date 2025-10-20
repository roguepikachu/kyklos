# Pod Security Baseline

## Purpose
This document defines the security hardening requirements for the Kyklos controller pod. All settings align with the Kubernetes Restricted Pod Security Standard with additional defense-in-depth controls.

---

## Security Context Requirements

### Pod-Level Security Context

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65532  # nonroot user (distroless convention)
  runAsGroup: 65532
  fsGroup: 65532
  seccompProfile:
    type: RuntimeDefault
```

**Rationale**:
- `runAsNonRoot`: Enforces non-root execution. Container will fail to start if image specifies USER 0
- `runAsUser: 65532`: Explicit non-privileged UID. Matches distroless/static nonroot user
- `fsGroup`: Ensures mounted volumes are readable by controller process
- `seccompProfile: RuntimeDefault`: Applies default seccomp filtering to reduce syscall attack surface

**Validation**: Pod will be rejected by Pod Security Admission if `runAsNonRoot` is false under Restricted policy.

### Container-Level Security Context

```yaml
securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  runAsNonRoot: true
  runAsUser: 65532
  capabilities:
    drop:
      - ALL
  seccompProfile:
    type: RuntimeDefault
```

**Rationale**:
- `allowPrivilegeEscalation: false`: Prevents gaining more privileges than parent process via setuid binaries
- `readOnlyRootFilesystem: true`: Immutable container filesystem. Prevents runtime code injection
- `drop: ALL`: Removes all Linux capabilities. Controller needs no elevated capabilities
- Container-level settings override pod-level for defense in depth

**Writable Paths**:
Controller requires writable directories for temporary files:

```yaml
volumeMounts:
- name: tmp
  mountPath: /tmp
  readOnly: false
- name: cache
  mountPath: /.cache
  readOnly: false

volumes:
- name: tmp
  emptyDir:
    sizeLimit: 100Mi
- name: cache
  emptyDir:
    sizeLimit: 50Mi
```

**Why /tmp and /.cache**:
- Go runtime may write temporary files during execution
- Leader election client writes lock metadata
- Explicit emptyDir volumes with size limits prevent disk exhaustion
- Ephemeral: cleared on pod restart

---

## AppArmor and SELinux

### AppArmor Profile (Recommended)
```yaml
metadata:
  annotations:
    container.apparmor.security.beta.kubernetes.io/kyklos-controller: runtime/default
```

**Effect**: Applies default AppArmor profile if AppArmor is enabled on node.

**Fallback**: Annotation ignored if AppArmor unavailable (no error).

### SELinux (If Enforcing Mode)
```yaml
securityContext:
  seLinuxOptions:
    type: spc_t  # Or use default (no override)
```

**Recommendation**: Do not override SELinux context. Let Kubernetes assign default.

**If Required**: Use `container_t` or cluster-specific type. Coordinate with cluster security policy.

---

## Resource Limits and Requests

### Conservative Defaults
```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi
```

**Rationale**:
- `cpu.requests`: Controller is event-driven, low CPU when idle
- `cpu.limits`: Allows burst for reconcile operations with many windows
- `memory.requests`: Base memory for controller-runtime cache
- `memory.limits`: Headroom for informer cache growth in cluster mode

**Sizing Guidance**:
| Deployment Mode | Watched Resources | Recommended Memory |
|-----------------|-------------------|-------------------|
| Namespaced | < 50 Deployments | 128Mi request / 256Mi limit |
| Namespaced | 50-200 Deployments | 256Mi request / 512Mi limit |
| Cluster-wide | < 500 Deployments | 512Mi request / 1Gi limit |
| Cluster-wide | 500+ Deployments | 1Gi request / 2Gi limit |

**Monitoring**: Alert if memory usage exceeds 80% of limit (indicates cache pressure or leak).

### Quality of Service
With requests = limits on memory and requests < limits on CPU:
- **QoS Class**: Burstable
- **Eviction Risk**: Medium (only evicted under memory pressure if exceeding requests)

**For Guaranteed QoS** (lowest eviction risk):
```yaml
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 200m
    memory: 256Mi
```

---

## Network Security

### Network Policy (Egress Only to API Server)
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kyklos-controller-netpol
  namespace: kyklos-system
spec:
  podSelector:
    matchLabels:
      app: kyklos-controller
  policyTypes:
  - Egress
  egress:
  # Allow DNS resolution
  - to:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
    ports:
    - protocol: UDP
      port: 53
  # Allow Kubernetes API server
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 6443
    - protocol: TCP
      port: 443
  # Block all other egress
```

**Rationale**:
- Controller only communicates with Kubernetes API server
- No external HTTP calls or database connections
- Blocks potential data exfiltration if controller is compromised

**DNS Exception**: Required for service name resolution (`kubernetes.default.svc`).

**Ingress Policy**: No ingress required. Controller exposes metrics on localhost only.

### Metrics Exposure
```yaml
# Metrics port bound to localhost only
args:
  - --metrics-bind-address=127.0.0.1:8080
  - --health-probe-bind-address=127.0.0.1:8081
```

**Rationale**: Prevents external metric scraping. Use sidecar or service mesh for metrics export.

**Alternative for Prometheus**:
```yaml
args:
  - --metrics-bind-address=:8080  # All interfaces

# NetworkPolicy ingress rule
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring
    ports:
    - protocol: TCP
      port: 8080
```

---

## Health Probes

### Liveness Probe
```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8081
    scheme: HTTP
  initialDelaySeconds: 15
  periodSeconds: 20
  timeoutSeconds: 5
  failureThreshold: 3
```

**Rationale**:
- Detects deadlocked controller (e.g., informer cache stuck)
- 3 consecutive failures (60 seconds total) triggers restart
- `initialDelaySeconds: 15` allows controller startup (cache sync)

**Endpoint Behavior**: `/healthz` returns 200 if controller leader election and cache are healthy.

### Readiness Probe
```yaml
readinessProbe:
  httpGet:
    path: /readyz
    port: 8081
    scheme: HTTP
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 2
```

**Rationale**:
- Indicates controller is ready to process events
- Service traffic gated by readiness (if Service is configured)
- Faster failure detection (20 seconds vs 60 for liveness)

**Endpoint Behavior**: `/readyz` returns 200 if cache is synced and leader election is stable.

### Startup Probe (Optional)
```yaml
startupProbe:
  httpGet:
    path: /readyz
    port: 8081
  initialDelaySeconds: 0
  periodSeconds: 5
  failureThreshold: 12  # 60 seconds total
```

**Use Case**: Large clusters where cache sync takes > 15 seconds. Prevents liveness probe from killing controller during startup.

---

## Image Security

### Base Image Recommendations
**Preferred**: Distroless images (gcr.io/distroless/static:nonroot)

**Rationale**:
- No shell, package manager, or binaries except controller
- Minimal attack surface (< 10 MB image)
- Built-in nonroot user (UID 65532)
- Regular security patches via Google

**Alternative**: Alpine-based images if dynamic linking required (e.g., CGO-enabled builds).

**Avoid**: Debian/Ubuntu full images (100+ MB, unnecessary attack surface).

### Image Digest Pinning
```yaml
image: gcr.io/kyklos/controller@sha256:abcdef1234567890...
```

**Rationale**:
- Immutable reference. Tag-based images can be overwritten
- Reproducible deployments
- Prevents supply chain attacks via tag poisoning

**Process**:
1. Build and push image with tag (e.g., `v0.1.0`)
2. Retrieve digest: `docker inspect --format='{{index .RepoDigests 0}}' gcr.io/kyklos/controller:v0.1.0`
3. Update Helm chart or manifest with digest reference

### Image Scanning
**CI/CD Integration**:
- Scan with Trivy, Grype, or Snyk before pushing to registry
- Block builds with HIGH or CRITICAL CVEs
- Re-scan weekly for new vulnerabilities

**Registry Configuration**:
- Enable image signing (cosign or Notary)
- Require signed images via admission controller (Kyverno/OPA)

### SBOM Generation
```bash
# Generate SBOM with Syft
syft packages gcr.io/kyklos/controller:v0.1.0 -o cyclonedx-json > sbom.json
```

**Attachment**: Store SBOM in OCI registry alongside image:
```bash
oras push gcr.io/kyklos/controller:v0.1.0-sbom sbom.json
```

**Usage**: SBOM enables rapid vulnerability correlation during zero-day disclosures.

---

## Service Account Configuration

### Service Account Token Projection
```yaml
serviceAccountName: kyklos-controller
automountServiceAccountToken: true  # Required for API access

volumes:
- name: kube-api-access
  projected:
    sources:
    - serviceAccountToken:
        expirationSeconds: 3600  # 1 hour token lifetime
        path: token
    - configMap:
        name: kube-root-ca.crt
        items:
        - key: ca.crt
          path: ca.crt
    - downwardAPI:
        items:
        - path: namespace
          fieldRef:
            fieldPath: metadata.namespace
```

**Rationale**:
- Short-lived tokens (1 hour) reduce risk of stolen token abuse
- Automatic rotation by kubelet
- Projected volume provides token, CA cert, and namespace in single mount

**Default Behavior**: Kubernetes 1.25+ automatically uses projected tokens. Explicit configuration shown for clarity.

### Disable Token Automount for Unrelated Pods
```yaml
# In other pods in same namespace
automountServiceAccountToken: false
```

**Rationale**: Prevents accidental token exposure in pods that don't need API access.

---

## Namespace Hardening

### Pod Security Admission
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: kyklos-system
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

**Effect**:
- **Enforce**: Rejects pods that violate Restricted standard (includes all settings in this doc)
- **Audit**: Logs violations to audit log
- **Warn**: Returns warning messages to kubectl users

**Validation**: Controller pod must comply with all Restricted requirements or deployment will fail.

### Resource Quotas
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: kyklos-quota
  namespace: kyklos-system
spec:
  hard:
    requests.cpu: "1"
    requests.memory: "512Mi"
    limits.cpu: "2"
    limits.memory: "1Gi"
    pods: "5"
```

**Rationale**: Prevents resource exhaustion if controller is misconfigured or compromised.

### Network Policies (Default Deny)
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: kyklos-system
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
```

**Effect**: Blocks all traffic unless explicitly allowed by additional NetworkPolicy (see earlier egress policy).

---

## Monitoring and Alerting

### Security Metrics to Track
```prometheus
# Unauthorized API calls (should be 0)
controller_runtime_client_requests_total{code="403"}

# Rate limiting (indicates API pressure)
controller_runtime_client_requests_total{code="429"}

# Reconcile errors
controller_runtime_reconcile_errors_total

# Memory usage approaching limit
container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.8
```

### Security Events to Alert On
- Pod restart (liveness probe failure)
- Image pull from unexpected registry
- Container running as root (should never happen)
- High reconcile error rate (> 10% error ratio)
- RBAC permission denied errors in logs

---

## Deployment Hardening Checklist

### Pre-Deployment
- [ ] Image scanned with zero HIGH/CRITICAL CVEs
- [ ] Image digest pinned (not using `:latest` or mutable tags)
- [ ] SBOM generated and stored
- [ ] Resource limits defined
- [ ] SecurityContext set (runAsNonRoot, readOnlyRootFilesystem, drop ALL)
- [ ] ServiceAccount created with minimal RBAC
- [ ] NetworkPolicy defined

### Post-Deployment
- [ ] Pod running as UID 65532 (verify with `kubectl exec -- id`)
- [ ] Filesystem read-only except /tmp and /.cache (verify with `touch /test`)
- [ ] Health probes responding (200 status)
- [ ] Metrics endpoint accessible on localhost only
- [ ] No unexpected network connections (use `kubectl exec -- netstat`)
- [ ] RBAC permissions validated (see RBAC-MATRIX.md validation commands)

### Runtime Verification
```bash
# Verify non-root user
kubectl exec -n kyklos-system deploy/kyklos-controller -- id
# Expected: uid=65532(nonroot) gid=65532(nonroot)

# Verify read-only filesystem
kubectl exec -n kyklos-system deploy/kyklos-controller -- touch /test
# Expected: touch: cannot touch '/test': Read-only file system

# Verify no capabilities
kubectl exec -n kyklos-system deploy/kyklos-controller -- grep Cap /proc/1/status
# Expected: CapEff: 0000000000000000 (all zeros)

# Verify seccomp profile
kubectl get pod -n kyklos-system -l app=kyklos-controller -o json | jq '.items[0].spec.securityContext.seccompProfile'
# Expected: {"type":"RuntimeDefault"}
```

---

## Incident Response

### If Controller Pod is Compromised
1. **Isolate**: Delete pod immediately (StatefulSet/Deployment will recreate)
2. **Audit**: Review audit logs for API calls made by controller ServiceAccount
   ```bash
   kubectl logs -n kube-system kube-apiserver-* | grep "system:serviceaccount:kyklos-system:kyklos-controller"
   ```
3. **Check Targets**: Verify all Deployment replicas match expected TimeWindowScaler configs
4. **Revoke**: Rotate ServiceAccount token (delete and recreate ServiceAccount)
5. **Image Scan**: Re-scan controller image for newly disclosed vulnerabilities
6. **Network Analysis**: Check if pod made unexpected outbound connections

### If RBAC Permissions Are Too Broad
1. **Snapshot**: Backup current Role/ClusterRole YAML
2. **Reduce**: Remove unnecessary permissions incrementally
3. **Test**: Verify controller still functions (run full test suite)
4. **Monitor**: Watch for RBAC denied errors in controller logs
5. **Document**: Update RBAC-MATRIX.md with changes

---

## Compliance Matrix

### Pod Security Standards
| Requirement | Setting | Compliance Level |
|-------------|---------|------------------|
| Non-root user | `runAsNonRoot: true` | Restricted |
| No privilege escalation | `allowPrivilegeEscalation: false` | Restricted |
| No capabilities | `capabilities.drop: ALL` | Restricted |
| Read-only filesystem | `readOnlyRootFilesystem: true` | Restricted |
| Seccomp | `RuntimeDefault` | Restricted |
| No host namespaces | (not used) | Restricted |
| No hostPath volumes | (not used) | Restricted |

**Result**: Fully compliant with Kubernetes Restricted Pod Security Standard.

### NIST SP 800-190 (Container Security)
- **4.1.1**: Image from trusted registry (gcr.io/distroless)
- **4.2.1**: Container runs with minimal privileges (no capabilities)
- **4.3.1**: Isolated via NetworkPolicy (no unnecessary network access)
- **4.4.1**: Resource limits prevent DoS

### CIS Docker Benchmark
- **4.1**: Run as non-root (PASS)
- **5.3**: Do not mount /var/run/docker.sock (PASS: not mounted)
- **5.7**: Do not map privileged ports (PASS: metrics on 8080)
- **5.25**: Restrict container from acquiring additional privileges (PASS: allowPrivilegeEscalation=false)

---

## Future Enhancements

### v0.2 Security Features
- [ ] Workload Identity integration (GKE, AKS, EKS)
- [ ] mTLS for metrics endpoint with sidecar
- [ ] Signed admission webhook certificates
- [ ] Runtime security monitoring (Falco rules)

### v0.3 Security Features
- [ ] Pod identity webhook for AWS IRSA
- [ ] SPIFFE/SPIRE integration
- [ ] Encrypted etcd secrets support (if controller stores sensitive data)

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-20 | Initial pod security baseline for v0.1 |

---

## References
- Kubernetes Pod Security Standards: https://kubernetes.io/docs/concepts/security/pod-security-standards/
- Distroless Images: https://github.com/GoogleContainerTools/distroless
- NIST SP 800-190: https://csrc.nist.gov/publications/detail/sp/800-190/final
- CIS Docker Benchmark: https://www.cisecurity.org/benchmark/docker
- Seccomp Profiles: https://kubernetes.io/docs/tutorials/security/seccomp/
