# Security Checklist

## Purpose
This operational checklist ensures the Kyklos controller is deployed with proper security controls and provides runbook snippets for common security-related failures.

---

## Pre-Release Checks

### Code and Dependencies
- [ ] All dependencies updated to latest patch versions
- [ ] No HIGH or CRITICAL CVEs in dependency scan (Trivy/Grype/Snyk)
- [ ] Go version is supported and patched (https://go.dev/dl/)
- [ ] SBOM generated and published with release artifacts
- [ ] Image signed with cosign/Notary (if registry supports)

**Validation Commands**:
```bash
# Scan dependencies
go list -json -m all | docker run --rm -i aquasec/trivy:latest fs --scanners vuln --severity HIGH,CRITICAL -

# Generate SBOM
syft packages . -o cyclonedx-json > sbom.json

# Sign image (if using cosign)
cosign sign --key cosign.key gcr.io/kyklos/controller:v0.1.0
```

### RBAC Configuration
- [ ] Role/ClusterRole uses explicit verbs (no wildcards `*`)
- [ ] No `delete` permission on Deployments or StatefulSets
- [ ] No access to `secrets` resource
- [ ] No access to RBAC resources (roles, rolebindings, etc.)
- [ ] ServiceAccount dedicated to controller (not `default`)
- [ ] RoleBinding/ClusterRoleBinding references correct ServiceAccount

**Validation Commands**:
```bash
# Check for wildcard permissions
kubectl get role kyklos-controller -n kyklos-system -o yaml | grep -E "verbs:.*\*|resources:.*\*|apiGroups:.*\*"
# Should return no results

# Verify no Secrets access
kubectl get role kyklos-controller -n kyklos-system -o jsonpath='{.rules[*].resources}' | grep secrets
# Should return no results

# Verify correct ServiceAccount binding
kubectl get rolebinding kyklos-controller -n kyklos-system -o jsonpath='{.subjects[0].name}'
# Should return: kyklos-controller
```

### Pod Security
- [ ] `securityContext.runAsNonRoot: true` set on pod
- [ ] `securityContext.runAsUser: 65532` (or other non-root UID)
- [ ] `securityContext.readOnlyRootFilesystem: true` on container
- [ ] `securityContext.allowPrivilegeEscalation: false` on container
- [ ] `securityContext.capabilities.drop: [ALL]` on container
- [ ] `seccomp` profile set to `RuntimeDefault`
- [ ] No `hostPath`, `hostNetwork`, or `hostPID` in pod spec
- [ ] Resource requests and limits defined

**Validation Commands**:
```bash
# Get pod security context
kubectl get pod -n kyklos-system -l app=kyklos-controller -o jsonpath='{.items[0].spec.securityContext}' | jq .

# Expected output includes:
# "runAsNonRoot": true
# "runAsUser": 65532
# "seccompProfile": {"type": "RuntimeDefault"}

# Get container security context
kubectl get pod -n kyklos-system -l app=kyklos-controller -o jsonpath='{.items[0].spec.containers[0].securityContext}' | jq .

# Expected output includes:
# "allowPrivilegeEscalation": false
# "readOnlyRootFilesystem": true
# "capabilities": {"drop": ["ALL"]}
```

### Image Security
- [ ] Base image is distroless or minimal (< 20 MB)
- [ ] Image reference uses digest (sha256), not just tag
- [ ] Image scanned in CI/CD pipeline (zero HIGH/CRITICAL CVEs)
- [ ] No shell (`/bin/sh`, `/bin/bash`) in image
- [ ] Image FROM statement uses trusted registry (gcr.io, docker.io official)

**Validation Commands**:
```bash
# Check image size
docker images gcr.io/kyklos/controller:v0.1.0 --format "{{.Size}}"
# Should be < 50 MB for distroless

# Scan image
trivy image gcr.io/kyklos/controller:v0.1.0 --severity HIGH,CRITICAL
# Should report 0 vulnerabilities

# Verify no shell
docker run --rm gcr.io/kyklos/controller:v0.1.0 /bin/sh -c "echo test"
# Should fail with: container_linux.go:367: starting container process caused: exec: "/bin/sh": stat /bin/sh: no such file or directory
```

### Health Probes
- [ ] Liveness probe configured (path: `/healthz`)
- [ ] Readiness probe configured (path: `/readyz`)
- [ ] Probe ports bound to localhost if metrics are internal
- [ ] `initialDelaySeconds` allows cache sync (recommend 15s)
- [ ] `failureThreshold` prevents premature restarts (recommend 3+)

**Validation Commands**:
```bash
# Check liveness probe
kubectl get deployment kyklos-controller -n kyklos-system -o jsonpath='{.spec.template.spec.containers[0].livenessProbe}' | jq .

# Verify probe responds
kubectl port-forward -n kyklos-system deploy/kyklos-controller 8081:8081
curl http://localhost:8081/healthz
# Should return 200 OK
```

### Network Security
- [ ] NetworkPolicy defined for controller pod
- [ ] Egress restricted to API server and DNS only
- [ ] No ingress rules (or only from monitoring namespace)
- [ ] Metrics endpoint bound to localhost (127.0.0.1:8080)
- [ ] No Service exposing controller externally

**Validation Commands**:
```bash
# Check NetworkPolicy exists
kubectl get networkpolicy -n kyklos-system kyklos-controller-netpol
# Should exist

# Verify metrics not externally accessible
kubectl run -it --rm test-curl --image=curlimages/curl --restart=Never -- curl -m 5 http://kyklos-controller.kyklos-system:8080/metrics
# Should timeout (no Service or ingress allowed)
```

### Namespace Hardening
- [ ] Pod Security Admission labels set on namespace
- [ ] `pod-security.kubernetes.io/enforce: restricted`
- [ ] ResourceQuota defined (optional but recommended)
- [ ] LimitRange defined for default resource limits (optional)
- [ ] Default-deny NetworkPolicy in place (optional)

**Validation Commands**:
```bash
# Check Pod Security Admission labels
kubectl get namespace kyklos-system -o jsonpath='{.metadata.labels}' | jq .

# Expected:
# "pod-security.kubernetes.io/enforce": "restricted"

# Verify Pod Security Admission blocks non-compliant pods
kubectl run -n kyklos-system test-root --image=nginx --restart=Never --overrides='{"spec":{"securityContext":{"runAsUser":0}}}'
# Should fail with: pods "test-root" is forbidden: violates PodSecurity "restricted:latest": runAsNonRoot != true
```

### Metrics and Monitoring
- [ ] Prometheus ServiceMonitor created (if using Prometheus Operator)
- [ ] Alert rules defined for controller health
- [ ] Dashboard created showing scaling operations
- [ ] Logs forwarded to centralized logging (Loki, Splunk, etc.)
- [ ] Audit logging enabled on API server for TimeWindowScaler operations

**Example Alert Rules**:
```yaml
# High reconcile error rate
- alert: KyklosHighErrorRate
  expr: rate(controller_runtime_reconcile_errors_total{controller="timewindowscaler"}[5m]) > 0.1
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: Kyklos controller error rate above 10%

# Controller pod not ready
- alert: KyklosControllerDown
  expr: kube_deployment_status_replicas_available{deployment="kyklos-controller"} == 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: Kyklos controller has no ready replicas
```

---

## Installation Verification

### Post-Install Checks (Namespaced Mode)

#### Step 1: Verify Deployment Running
```bash
kubectl get deployment kyklos-controller -n kyklos-system
# Expected: 1/1 READY

kubectl get pod -n kyklos-system -l app=kyklos-controller
# Expected: STATUS Running
```

#### Step 2: Verify RBAC Bindings
```bash
# Check Role exists
kubectl get role kyklos-controller -n kyklos-system

# Check RoleBinding exists
kubectl get rolebinding kyklos-controller -n kyklos-system

# Verify ServiceAccount
kubectl get serviceaccount kyklos-controller -n kyklos-system

# Test permissions (should succeed)
kubectl auth can-i get deployments --namespace=kyklos-system --as=system:serviceaccount:kyklos-system:kyklos-controller
kubectl auth can-i patch deployments --namespace=kyklos-system --as=system:serviceaccount:kyklos-system:kyklos-controller
kubectl auth can-i get timewindowscalers --namespace=kyklos-system --as=system:serviceaccount:kyklos-system:kyklos-controller

# Test forbidden permissions (should fail)
kubectl auth can-i delete deployments --namespace=kyklos-system --as=system:serviceaccount:kyklos-system:kyklos-controller
kubectl auth can-i get secrets --namespace=kyklos-system --as=system:serviceaccount:kyklos-system:kyklos-controller
```

#### Step 3: Verify Pod Security
```bash
# Check container is non-root
kubectl exec -n kyklos-system deploy/kyklos-controller -- id
# Expected: uid=65532(nonroot) gid=65532(nonroot)

# Check read-only filesystem
kubectl exec -n kyklos-system deploy/kyklos-controller -- touch /test
# Expected: touch: cannot touch '/test': Read-only file system

# Check no capabilities
kubectl exec -n kyklos-system deploy/kyklos-controller -- grep CapEff /proc/1/status
# Expected: CapEff: 0000000000000000
```

#### Step 4: Verify Health Endpoints
```bash
# Port-forward to pod
kubectl port-forward -n kyklos-system deploy/kyklos-controller 8081:8081 &

# Test healthz
curl http://localhost:8081/healthz
# Expected: ok

# Test readyz
curl http://localhost:8081/readyz
# Expected: ok

# Stop port-forward
kill %1
```

#### Step 5: Verify Logging
```bash
# Check logs for errors
kubectl logs -n kyklos-system deploy/kyklos-controller --tail=50

# Expected patterns:
# "Starting controller" (startup message)
# "Becoming leader" (if leader election enabled)
# "Cache synced" (informer ready)

# Should NOT see:
# "permission denied" (RBAC issues)
# "panic" (crashes)
# "connection refused" (API server issues)
```

### Post-Install Checks (Cluster Mode)

#### Step 1: Verify ClusterRole and ClusterRoleBinding
```bash
# Check ClusterRole
kubectl get clusterrole kyklos-controller

# Check ClusterRoleBinding
kubectl get clusterrolebinding kyklos-controller

# Verify cross-namespace permissions
kubectl auth can-i get deployments --all-namespaces --as=system:serviceaccount:kyklos-system:kyklos-controller
# Expected: yes

kubectl auth can-i patch deployments --namespace=production --as=system:serviceaccount:kyklos-system:kyklos-controller
# Expected: yes
```

#### Step 2: Verify Namespace Watch Scope
```bash
# Check controller logs for watch scope
kubectl logs -n kyklos-system deploy/kyklos-controller | grep "watch"

# Expected in cluster mode:
# "Watching all namespaces"

# Expected in namespaced mode:
# "Watching namespace: kyklos-system"
```

---

## Runtime Verification

### Functional Security Tests

#### Test 1: Verify RBAC Enforcement
```bash
# Create test namespace
kubectl create namespace kyklos-test

# Try to create TimeWindowScaler as unauthorized user
kubectl create -f - <<EOF
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: test-unauthorized
  namespace: kyklos-test
spec:
  targetRef:
    kind: Deployment
    name: test-app
  timezone: America/New_York
  windows:
    - days: [Mon, Tue, Wed, Thu, Fri]
      start: "09:00"
      end: "17:00"
      replicas: 3
EOF

# Should succeed if you have create permissions
# Controller should handle it if it has Role in kyklos-test namespace (namespaced mode)
# Controller should handle it cluster-wide (cluster mode)
```

#### Test 2: Verify Read-Only Filesystem
```bash
# Attempt to write to root filesystem
kubectl exec -n kyklos-system deploy/kyklos-controller -- sh -c "echo test > /testfile"
# Expected error: sh: can't create /testfile: Read-only file system

# Verify /tmp is writable
kubectl exec -n kyklos-system deploy/kyklos-controller -- sh -c "echo test > /tmp/testfile && cat /tmp/testfile"
# Expected: test
```

#### Test 3: Verify NetworkPolicy Enforcement
```bash
# Try to access external website from controller pod
kubectl exec -n kyklos-system deploy/kyklos-controller -- curl -m 5 https://example.com
# Expected: timeout or connection refused (NetworkPolicy blocks egress)

# Verify API server access still works (check logs for successful reconciles)
kubectl logs -n kyklos-system deploy/kyklos-controller --tail=20 | grep "Reconciled"
# Should see successful reconciliation logs
```

#### Test 4: Verify No Secrets Access
```bash
# Create test Secret
kubectl create secret generic test-secret -n kyklos-system --from-literal=password=secret123

# Check controller logs for Secret access attempts
kubectl logs -n kyklos-system deploy/kyklos-controller | grep -i secret
# Should NOT see any Secret read attempts

# Try to read Secret as controller ServiceAccount (should fail)
kubectl auth can-i get secrets --namespace=kyklos-system --as=system:serviceaccount:kyklos-system:kyklos-controller
# Expected: no

# Cleanup
kubectl delete secret test-secret -n kyklos-system
```

---

## Common Failure Runbooks

### Failure: RBAC Permission Denied

**Symptoms**:
- Controller logs show: `deployments.apps is forbidden: User "system:serviceaccount:kyklos-system:kyklos-controller" cannot patch resource "deployments"`
- TimeWindowScaler status shows Degraded=True

**Diagnosis**:
```bash
# Check if Role/ClusterRole exists
kubectl get role kyklos-controller -n kyklos-system
kubectl get clusterrole kyklos-controller

# Check if RoleBinding/ClusterRoleBinding exists and references correct ServiceAccount
kubectl get rolebinding kyklos-controller -n kyklos-system -o yaml
kubectl get clusterrolebinding kyklos-controller -o yaml

# Manually test permissions
kubectl auth can-i patch deployments --namespace=kyklos-system --as=system:serviceaccount:kyklos-system:kyklos-controller
```

**Resolution**:
```bash
# If Role is missing, create it (use RBAC-MATRIX.md as reference)
kubectl apply -f deploy/rbac/role.yaml

# If RoleBinding is incorrect, fix subject reference
kubectl edit rolebinding kyklos-controller -n kyklos-system
# Ensure subjects[0].name = kyklos-controller
# Ensure subjects[0].namespace = kyklos-system

# Restart controller to retry
kubectl rollout restart deployment kyklos-controller -n kyklos-system
```

---

### Failure: Target Deployment Not Found

**Symptoms**:
- TimeWindowScaler status shows Ready=False, Reason=TargetNotFound
- Events show: `Target Deployment not found`

**Diagnosis**:
```bash
# Check TimeWindowScaler configuration
kubectl get timewindowscaler <name> -n <namespace> -o yaml

# Verify targetRef.name and targetRef.namespace are correct
# Check if Deployment exists
kubectl get deployment <targetRef.name> -n <namespace>
```

**Resolution**:
```bash
# If Deployment name is wrong, fix TimeWindowScaler
kubectl edit timewindowscaler <name> -n <namespace>
# Update spec.targetRef.name to correct value

# If Deployment is missing, create it
kubectl create deployment <name> --image=<image> -n <namespace>

# Wait for controller to reconcile (check status)
kubectl get timewindowscaler <name> -n <namespace> -o jsonpath='{.status.conditions}'
```

---

### Failure: Events Blocked

**Symptoms**:
- Controller logs show: `Failed to create event: events is forbidden`
- No events visible with `kubectl get events -n <namespace>`

**Diagnosis**:
```bash
# Check if controller has events permission
kubectl auth can-i create events --namespace=kyklos-system --as=system:serviceaccount:kyklos-system:kyklos-controller
```

**Resolution (Option 1: Add Permission)**:
```bash
# Add events permission to Role
kubectl patch role kyklos-controller -n kyklos-system --type=json -p='[
  {"op": "add", "path": "/rules/-", "value": {
    "apiGroups": [""],
    "resources": ["events"],
    "verbs": ["create", "patch"]
  }}
]'

# Restart controller
kubectl rollout restart deployment kyklos-controller -n kyklos-system
```

**Resolution (Option 2: Disable Events)**:
```bash
# If events are not needed, disable in controller
kubectl set env deployment/kyklos-controller -n kyklos-system ENABLE_EVENTS=false

# Or set flag in Helm values
# controller:
#   enableEvents: false
```

---

### Failure: ConfigMap Not Found (Holidays)

**Symptoms**:
- TimeWindowScaler status shows Degraded=True, Reason=HolidaySourceMissing
- Controller logs show: `ConfigMap <name> not found`

**Diagnosis**:
```bash
# Check TimeWindowScaler holiday configuration
kubectl get timewindowscaler <name> -n <namespace> -o jsonpath='{.spec.holidays}'

# Check if ConfigMap exists
kubectl get configmap <sourceRef.name> -n <namespace>
```

**Resolution**:
```bash
# Create missing ConfigMap with holiday dates
kubectl create configmap company-holidays -n <namespace> --from-literal=2025-12-25="" --from-literal=2025-01-01=""

# Or disable holidays if not needed
kubectl patch timewindowscaler <name> -n <namespace> --type=json -p='[
  {"op": "remove", "path": "/spec/holidays"}
]'

# Wait for controller to reconcile
kubectl get timewindowscaler <name> -n <namespace> -o jsonpath='{.status.conditions}'
```

---

### Failure: Pod Fails Pod Security Admission

**Symptoms**:
- Pod creation fails with: `violates PodSecurity "restricted:latest"`
- Deployment shows 0/1 READY

**Diagnosis**:
```bash
# Check namespace Pod Security labels
kubectl get namespace kyklos-system -o jsonpath='{.metadata.labels}' | jq .

# Check pod spec for violations
kubectl get deployment kyklos-controller -n kyklos-system -o yaml | grep -A 10 securityContext
```

**Resolution**:
```bash
# Update Deployment with correct SecurityContext (see POD-SECURITY-BASELINE.md)
kubectl patch deployment kyklos-controller -n kyklos-system --type=json -p='[
  {"op": "add", "path": "/spec/template/spec/securityContext", "value": {
    "runAsNonRoot": true,
    "runAsUser": 65532,
    "seccompProfile": {"type": "RuntimeDefault"}
  }},
  {"op": "add", "path": "/spec/template/spec/containers/0/securityContext", "value": {
    "allowPrivilegeEscalation": false,
    "readOnlyRootFilesystem": true,
    "capabilities": {"drop": ["ALL"]}
  }}
]'

# Or relax namespace policy temporarily (NOT recommended for production)
kubectl label namespace kyklos-system pod-security.kubernetes.io/enforce=baseline --overwrite
```

---

### Failure: High Memory Usage

**Symptoms**:
- Pod approaching memory limit (kubectl top pod shows > 80% usage)
- OOMKilled events in pod status

**Diagnosis**:
```bash
# Check current memory usage
kubectl top pod -n kyklos-system -l app=kyklos-controller

# Check memory limit
kubectl get deployment kyklos-controller -n kyklos-system -o jsonpath='{.spec.template.spec.containers[0].resources.limits.memory}'

# Count watched resources
kubectl get deployments --all-namespaces | wc -l
kubectl get timewindowscalers --all-namespaces | wc -l
```

**Resolution**:
```bash
# Increase memory limit (see POD-SECURITY-BASELINE.md sizing guidance)
kubectl patch deployment kyklos-controller -n kyklos-system --type=json -p='[
  {"op": "replace", "path": "/spec/template/spec/containers/0/resources/limits/memory", "value": "512Mi"},
  {"op": "replace", "path": "/spec/template/spec/containers/0/resources/requests/memory", "value": "256Mi"}
]'

# Or switch from cluster mode to namespaced mode to reduce cache size
kubectl set env deployment/kyklos-controller -n kyklos-system WATCH_NAMESPACE=kyklos-system
```

---

### Failure: Clock Skew / Wrong Timezone Scaling

**Symptoms**:
- Scaling happens at wrong times
- Controller logs show times that don't match expected timezone

**Diagnosis**:
```bash
# Check controller's view of current time
kubectl logs -n kyklos-system deploy/kyklos-controller | grep "Current local time"

# Check node time
kubectl get node -o wide
kubectl debug node/<node-name> -it --image=busybox -- date

# Verify TimeWindowScaler timezone
kubectl get timewindowscaler <name> -n <namespace> -o jsonpath='{.spec.timezone}'
```

**Resolution**:
```bash
# Fix timezone in TimeWindowScaler if incorrect
kubectl patch timewindowscaler <name> -n <namespace> --type=merge -p '{"spec":{"timezone":"America/New_York"}}'

# If node time is wrong, fix NTP on nodes (cluster admin responsibility)
# Ensure nodes have chrony or ntpd running

# Restart controller to reload timezone data
kubectl rollout restart deployment kyklos-controller -n kyklos-system
```

---

## Security Incident Response

### If Controller Pod is Compromised

#### Immediate Actions (Within 5 Minutes)
1. **Isolate**: Delete the pod immediately
   ```bash
   kubectl delete pod -n kyklos-system -l app=kyklos-controller
   ```

2. **Disable Auto-Restart**: Scale Deployment to 0 while investigating
   ```bash
   kubectl scale deployment kyklos-controller -n kyklos-system --replicas=0
   ```

3. **Preserve Evidence**: Save pod logs and events
   ```bash
   kubectl logs -n kyklos-system -l app=kyklos-controller --previous > /tmp/kyklos-compromise-logs.txt
   kubectl get events -n kyklos-system --sort-by='.lastTimestamp' > /tmp/kyklos-events.txt
   ```

#### Investigation (Within 30 Minutes)
4. **Check API Audit Logs**: Review all actions by controller ServiceAccount
   ```bash
   # On control plane node or via API server audit log
   grep "system:serviceaccount:kyklos-system:kyklos-controller" /var/log/kubernetes/audit.log | tail -1000 > /tmp/kyklos-audit.txt
   ```

5. **Verify Deployment Replicas**: Check if any unexpected scaling occurred
   ```bash
   kubectl get deployments --all-namespaces -o json | jq -r '.items[] | "\(.metadata.namespace)/\(.metadata.name): \(.spec.replicas)"'
   # Compare with expected values from TimeWindowScaler configs
   ```

6. **Check for Unauthorized Resource Access**: Look for 403 Forbidden attempts
   ```bash
   grep "Forbidden" /tmp/kyklos-compromise-logs.txt
   ```

7. **Network Analysis**: Check if pod made unexpected connections
   ```bash
   # Review NetworkPolicy events
   kubectl get events --all-namespaces --field-selector involvedObject.kind=NetworkPolicy
   ```

#### Remediation (Within 1 Hour)
8. **Rotate ServiceAccount Token**:
   ```bash
   kubectl delete serviceaccount kyklos-controller -n kyklos-system
   kubectl create serviceaccount kyklos-controller -n kyklos-system
   # Recreate RoleBinding
   kubectl apply -f deploy/rbac/rolebinding.yaml
   ```

9. **Re-scan Image**: Verify controller image has no vulnerabilities
   ```bash
   trivy image gcr.io/kyklos/controller:v0.1.0 --severity HIGH,CRITICAL
   ```

10. **Restore Controller**: Re-enable Deployment with clean image
    ```bash
    kubectl set image deployment/kyklos-controller -n kyklos-system kyklos-controller=gcr.io/kyklos/controller@sha256:<new-digest>
    kubectl scale deployment kyklos-controller -n kyklos-system --replicas=1
    ```

11. **Enhanced Monitoring**: Add alerts for suspicious activity
    ```yaml
    # Alert on unexpected API calls
    - alert: KyklosUnexpectedAPICall
      expr: sum(rate(apiserver_audit_event_total{user=~"system:serviceaccount:kyklos-system:kyklos-controller", verb!~"get|list|watch|patch|update"}[5m])) > 0
    ```

---

## Periodic Security Audits

### Monthly Checks
- [ ] Review audit logs for unusual scaling operations
- [ ] Check controller logs for RBAC denied errors (indicates permission changes)
- [ ] Verify no new CVEs in controller dependencies
- [ ] Review TimeWindowScaler configurations for risky patterns (e.g., replicas=0 in production)

### Quarterly Checks
- [ ] Full RBAC review against RBAC-MATRIX.md
- [ ] Image re-scan for all deployed versions
- [ ] NetworkPolicy effectiveness test (attempt unauthorized egress)
- [ ] Pod Security Admission compliance test
- [ ] Threat model review and update

### Annual Checks
- [ ] Penetration test including controller as target
- [ ] Full security documentation review
- [ ] Compliance audit (CIS Kubernetes Benchmark, SOC2, etc.)
- [ ] Incident response plan tabletop exercise

---

## Security Contacts

**Security Issues**: Report vulnerabilities via email to security@kyklos.io (or your organization's security team)

**Disclosure Policy**: Follow coordinated disclosure (90-day embargo)

**Security Advisories**: Monitor https://github.com/kyklos/kyklos/security/advisories

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025-10-20 | Initial security checklist for v0.1 |

---

## References
- Kyklos RBAC Matrix: `/Users/aykumar/personal/kyklos/docs/security/RBAC-MATRIX.md`
- Kyklos Pod Security Baseline: `/Users/aykumar/personal/kyklos/docs/security/POD-SECURITY-BASELINE.md`
- Kyklos Threat Model: `/Users/aykumar/personal/kyklos/docs/security/THREAT-MODEL.md`
- Kubernetes Security Checklist: https://kubernetes.io/docs/concepts/security/security-checklist/
- Pod Security Standards: https://kubernetes.io/docs/concepts/security/pod-security-standards/
