# Troubleshooting Guide

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** local-workflow-designer

This guide provides solutions to common issues encountered during Kyklos local development. Issues are organized by symptom with clear diagnosis steps and resolution procedures.

---

## Table of Contents

1. [Environment Setup Issues](#environment-setup-issues)
2. [Cluster Issues](#cluster-issues)
3. [Build and Image Issues](#build-and-image-issues)
4. [Deployment Issues](#deployment-issues)
5. [Controller Runtime Issues](#controller-runtime-issues)
6. [TimeWindowScaler Behavior Issues](#timewindowscaler-behavior-issues)
7. [RBAC and Permission Issues](#rbac-and-permission-issues)
8. [Performance Issues](#performance-issues)
9. [Testing Issues](#testing-issues)
10. [Reset Procedures](#reset-procedures)

---

## Environment Setup Issues

### Issue: `make verify-tools` Reports Missing Tools

**Symptom:**
```
✗ Go not found
✗ Docker not running
```

**Diagnosis:**
```bash
which go
which docker
docker info
```

**Resolution:**

**For macOS:**
```bash
# Install missing tools
brew install go docker kubectl kind

# Start Docker
open -a Docker
# Wait for Docker to start (30-60 seconds)

# Verify
make verify-tools
```

**For Linux (Ubuntu/Debian):**
```bash
# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
# Log out and back in for group membership

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install Kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Verify
make verify-tools
```

---

### Issue: Go Version Too Old

**Symptom:**
```
Error: Go 1.19.0 installed, but 1.21+ required
```

**Diagnosis:**
```bash
go version
```

**Resolution:**

**macOS:**
```bash
brew upgrade go
go version
```

**Linux:**
```bash
# Remove old version
sudo rm -rf /usr/local/go

# Install new version
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Verify
go version
```

**Verify Go Modules Work:**
```bash
cd /Users/aykumar/personal/kyklos
go mod download
go mod verify
```

---

### Issue: `make tools` Fails to Download

**Symptom:**
```
Error: Failed to install controller-gen
dial tcp: lookup github.com: no such host
```

**Diagnosis:**
```bash
# Check internet connectivity
curl -I https://github.com

# Check Go proxy
echo $GOPROXY
```

**Resolution:**

**Network Issue:**
```bash
# Test connectivity
ping -c 3 github.com

# If behind proxy, set proxy environment
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
export NO_PROXY=localhost,127.0.0.1

# Retry
make tools
```

**Go Proxy Issue:**
```bash
# Use direct mode (bypass proxy)
export GOPROXY=direct
make tools

# Or use alternative proxy
export GOPROXY=https://goproxy.cn,direct
make tools
```

**Permissions Issue:**
```bash
# Ensure tools directory is writable
mkdir -p tools/bin
chmod 755 tools/bin

# Retry
make tools
```

---

## Cluster Issues

### Issue: `make cluster-up` Fails with Port Already Allocated

**Symptom:**
```
ERROR: failed to create cluster: failed to create cluster: could not bind to port 6443
```

**Diagnosis:**
```bash
# Check if cluster already exists
kind get clusters

# Check what's using the port
lsof -i :6443
```

**Resolution:**

**Existing Cluster:**
```bash
# Delete existing cluster
make cluster-down

# Create fresh cluster
make cluster-up
```

**Port Conflict with Another Process:**
```bash
# Find process using port
lsof -i :6443

# Kill process (if safe)
kill <PID>

# Or use different port in Kind config
# Edit config/kind-cluster.yaml and change apiServerPort
```

**Multiple Kind Clusters:**
```bash
# List all Kind clusters
kind get clusters

# Delete all Kind clusters
kind delete clusters --all

# Create fresh Kyklos cluster
make cluster-up
```

---

### Issue: Cluster Created But kubectl Can't Connect

**Symptom:**
```
Unable to connect to the server: dial tcp 127.0.0.1:xxxxx: connect: connection refused
```

**Diagnosis:**
```bash
# Check cluster exists
kind get clusters
docker ps | grep kyklos-dev

# Check kubeconfig
kubectl config current-context
kubectl cluster-info
```

**Resolution:**

**Wrong Context:**
```bash
# List contexts
kubectl config get-contexts

# Switch to Kyklos cluster
kubectl config use-context kind-kyklos-dev

# Verify
kubectl cluster-info
```

**Kubeconfig Corrupted:**
```bash
# Recreate kubeconfig entry
kind export kubeconfig --name kyklos-dev

# Verify
kubectl get nodes
```

**Cluster Not Fully Started:**
```bash
# Wait for cluster to be ready
kubectl wait --for=condition=Ready nodes --all --timeout=120s

# Check control plane pods
kubectl get pods -n kube-system
```

---

### Issue: Cluster Nodes Not Ready

**Symptom:**
```
NAME                        STATUS     ROLES           AGE
kyklos-dev-control-plane    NotReady   control-plane   2m
```

**Diagnosis:**
```bash
kubectl describe node kyklos-dev-control-plane
kubectl get pods -n kube-system
```

**Resolution:**

**CNI Not Installed:**
```bash
# Check CNI pods
kubectl get pods -n kube-system -l k8s-app=kindnet

# If not running, recreate cluster
make cluster-down cluster-up
```

**Insufficient Resources:**
```bash
# Check Docker resources
docker info | grep -A 5 "CPUs\|Total Memory"

# Increase Docker resources in Docker Desktop:
# Settings -> Resources -> Advanced
# Set CPUs: 4+, Memory: 8GB+
```

**Disk Space Issue:**
```bash
# Check disk space
df -h

# Clean up Docker
docker system prune -a --volumes
```

---

### Issue: Cluster Works But Pods Stuck in Pending

**Symptom:**
```
NAME                       READY   STATUS    RESTARTS   AGE
demo-app-7d4c8bf5c9-abc    0/1     Pending   0          2m
```

**Diagnosis:**
```bash
kubectl describe pod <pod-name> -n <namespace>
```

**Resolution:**

**Insufficient CPU/Memory:**
```
Events:
  Warning  FailedScheduling  pod didn't fit: Insufficient cpu
```

**Fix:**
```bash
# Reduce resource requests in deployment
# Or increase Docker resources
# Or delete other pods to free resources
```

**Image Pull Error:**
```
Events:
  Warning  Failed  Failed to pull image "kyklos/controller:dev": not found
```

**Fix:**
```bash
# Build and load image
make docker-build kind-load
```

**No Nodes Available:**
```bash
# Check nodes
kubectl get nodes

# If no nodes, recreate cluster
make cluster-down cluster-up
```

---

## Build and Image Issues

### Issue: `make build` Fails with Compilation Errors

**Symptom:**
```
Error: undefined: someFunction
Error: cannot find package "github.com/..."
```

**Diagnosis:**
```bash
# Check Go modules
go mod tidy
go mod verify

# Check generated code
ls api/v1alpha1/zz_generated.deepcopy.go
```

**Resolution:**

**Missing Dependencies:**
```bash
# Update dependencies
go mod download
go mod tidy

# Verify
go mod verify

# Retry build
make build
```

**Outdated Generated Code:**
```bash
# Regenerate
make manifests generate

# Clean and rebuild
make clean build
```

**Import Path Issues:**
```bash
# Verify module name in go.mod
head -1 go.mod
# Expected: module github.com/your-org/kyklos

# If incorrect, fix all imports and go.mod
```

---

### Issue: `make docker-build` Fails

**Symptom:**
```
Error: Cannot connect to the Docker daemon
Error: failed to solve with frontend dockerfile.v0
```

**Diagnosis:**
```bash
docker info
docker ps
```

**Resolution:**

**Docker Not Running:**
```bash
# macOS: Start Docker Desktop
open -a Docker

# Linux: Start Docker daemon
sudo systemctl start docker

# Verify
docker info
```

**BuildKit Issues:**
```bash
# Disable BuildKit
export DOCKER_BUILDKIT=0
make docker-build

# Or update Docker to latest version
```

**Disk Space Issues:**
```bash
# Check space
df -h

# Clean Docker
docker system prune -a --volumes

# Retry
make docker-build
```

**Dockerfile Issues:**
```bash
# Test Dockerfile manually
docker build -t test -f Dockerfile .

# Check for syntax errors in Dockerfile
cat Dockerfile
```

---

### Issue: Image Built But Not Visible in Cluster

**Symptom:**
```
Error: Failed to pull image "kyklos/controller:dev": not found
```

**Diagnosis:**
```bash
# Check image exists locally
docker images | grep kyklos

# Check image loaded in Kind nodes
docker exec -it kyklos-dev-control-plane crictl images | grep kyklos
```

**Resolution:**

**Image Not Loaded:**
```bash
# Load image into Kind cluster
make kind-load

# Verify
make verify-image-loaded
```

**Wrong Image Tag:**
```bash
# Check deployment image
kubectl get deployment -n kyklos-system kyklos-controller-manager -o jsonpath='{.spec.template.spec.containers[0].image}'

# If mismatch, rebuild with correct tag
make docker-build IMG=kyklos/controller:dev kind-load
```

**Image Pull Policy Wrong:**
```bash
# Check pull policy
kubectl get deployment -n kyklos-system kyklos-controller-manager -o jsonpath='{.spec.template.spec.containers[0].imagePullPolicy}'

# Should be: IfNotPresent

# If "Always", patch deployment
kubectl patch deployment -n kyklos-system kyklos-controller-manager -p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","imagePullPolicy":"IfNotPresent"}]}}}}'
```

---

## Deployment Issues

### Issue: `make install-crds` Fails

**Symptom:**
```
Error: error validating "config/crd/bases/kyklos.io_timewindowscalers.yaml": error validating data
```

**Diagnosis:**
```bash
# Validate CRD manifest
kubectl apply --dry-run=server -f config/crd/bases/kyklos.io_timewindowscalers.yaml

# Check CRD file exists
ls -la config/crd/bases/
```

**Resolution:**

**CRD Not Generated:**
```bash
# Generate CRDs
make manifests

# Verify file exists
ls -la config/crd/bases/kyklos.io_timewindowscalers.yaml

# Apply
make install-crds
```

**Invalid CRD Schema:**
```bash
# Check for syntax errors
kubectl apply --dry-run=server -f config/crd/bases/kyklos.io_timewindowscalers.yaml

# If errors, check Go types and kubebuilder markers
vim api/v1alpha1/timewindowscaler_types.go

# Regenerate
make manifests
```

**Cluster API Server Issue:**
```bash
# Verify cluster
kubectl get nodes

# Check API server
kubectl get --raw /healthz
```

---

### Issue: `make deploy` Fails with Permission Denied

**Symptom:**
```
Error: serviceaccounts "kyklos-controller" is forbidden: User "system:anonymous" cannot create resource
```

**Diagnosis:**
```bash
# Check current context
kubectl config current-context

# Check user
kubectl config view --minify
```

**Resolution:**

**Wrong Cluster Context:**
```bash
# Switch to correct context
kubectl config use-context kind-kyklos-dev

# Verify
kubectl config current-context
```

**Insufficient Permissions:**
```bash
# For Kind cluster, you should have admin access
kubectl auth can-i create namespace
# Expected: yes

# If no, recreate cluster
make cluster-down cluster-up
```

---

### Issue: Controller Pod in CrashLoopBackOff

**Symptom:**
```
NAME                                       READY   STATUS             RESTARTS
kyklos-controller-manager-abc123-xyz       0/1     CrashLoopBackOff   5
```

**Diagnosis:**
```bash
# Check logs
make logs-controller

# Describe pod
kubectl describe pod -n kyklos-system -l app=kyklos-controller
```

**Resolution:**

**Missing CRDs:**
```
Error: no matches for kind "TimeWindowScaler" in version "kyklos.io/v1alpha1"
```

**Fix:**
```bash
make install-crds
make restart-controller
```

**RBAC Issues:**
```
Error: Failed to list *v1alpha1.TimeWindowScaler: timewindowscalers.kyklos.io is forbidden
```

**Fix:**
```bash
# Verify RBAC
make verify-rbac

# Reapply RBAC
kubectl apply -f config/rbac/

# Restart controller
make restart-controller
```

**Binary Panic:**
```
panic: runtime error: invalid memory address
```

**Fix:**
```bash
# Rebuild controller
make clean build docker-build kind-load

# Redeploy
make deploy-rollout
```

**Wrong Image Architecture:**
```
Error: exec format error
```

**Fix:**
```bash
# Rebuild for correct architecture
GOARCH=amd64 make build docker-build kind-load
```

---

## Controller Runtime Issues

### Issue: Controller Not Reconciling Resources

**Symptom:**
- TimeWindowScaler created but status never populated
- No events generated
- Controller logs show no reconciliation

**Diagnosis:**
```bash
# Check controller logs
make logs-controller | grep Reconciling

# Check controller pod is running
kubectl get pods -n kyklos-system

# Check TWS resource exists
kubectl get tws --all-namespaces
```

**Resolution:**

**Controller Not Watching TWS:**
```bash
# Check controller startup logs
make logs-controller | grep "Starting EventSource"

# Should see:
# "Starting EventSource" {"controller": "timewindowscaler", "source": "kind source: *v1alpha1.TimeWindowScaler"}

# If missing, rebuild and redeploy
make redeploy
```

**RBAC Blocking List/Watch:**
```bash
# Verify permissions
make verify-rbac

# Test specific permission
kubectl auth can-i list timewindowscalers --as=system:serviceaccount:kyklos-system:kyklos-controller
# Expected: yes
```

**Controller Manager Not Running:**
```bash
# Check pod status
kubectl get pods -n kyklos-system -l app=kyklos-controller

# If not running, check events
kubectl get events -n kyklos-system
```

---

### Issue: Reconcile Loop Errors in Logs

**Symptom:**
```
ERROR Reconciler error {"controller": "timewindowscaler", "error": "..."}
```

**Diagnosis:**
```bash
# Get full error context
make logs-controller | grep -A 10 "Reconciler error"
```

**Common Errors and Fixes:**

**"deployment not found"**
```bash
# Verify target deployment exists
kubectl get deployment <target-name> -n <namespace>

# Create if missing
kubectl create deployment <target-name> --image=nginx -n <namespace>
```

**"failed to update deployment"**
```bash
# Check RBAC permissions
kubectl auth can-i update deployments --as=system:serviceaccount:kyklos-system:kyklos-controller

# Verify deployment is not being managed by another controller
kubectl get deployment <name> -n <namespace> -o yaml | grep ownerReferences
```

**"context deadline exceeded"**
```bash
# Cluster may be overloaded or slow
# Check cluster resources
kubectl top nodes

# Increase timeout in controller code (if necessary)
```

---

### Issue: Controller Requeue Loop Too Fast

**Symptom:**
```
INFO Requeue scheduled in 0s
INFO Requeue scheduled in 0s
INFO Requeue scheduled in 0s
```

**Diagnosis:**
```bash
# Check requeue timing in logs
make logs-controller | grep "Requeue scheduled"
```

**Resolution:**

This indicates an error condition causing immediate requeue.

**Check for errors:**
```bash
make logs-controller | grep ERROR
```

**Common causes:**
- Invalid timezone in TWS spec
- Target deployment not found
- RBAC permission issues

**Fix:**
```bash
# Check TWS status for error conditions
kubectl get tws <name> -n <namespace> -o yaml | grep -A 20 conditions

# Address specific error condition
```

---

## TimeWindowScaler Behavior Issues

### Issue: Scaling Not Happening at Expected Times

**Symptom:**
- Deployment should scale at specific time but doesn't
- Scaling happens at wrong times

**Diagnosis:**
```bash
# Check TWS configuration
kubectl get tws <name> -n <namespace> -o yaml

# Check controller's time interpretation
make logs-controller | grep "Current time"

# Check system time
date
date -u  # UTC time
```

**Resolution:**

**Timezone Mismatch:**
```bash
# Verify timezone in TWS
kubectl get tws <name> -o jsonpath='{.spec.timezone}'

# Test timezone is valid
TZ=<timezone-from-spec> date

# If invalid, update TWS
kubectl patch tws <name> -n <namespace> --type=merge -p '{"spec":{"timezone":"UTC"}}'
```

**Window Configuration Error:**
```yaml
# Check window times
windows:
- days: [Mon, Tue, Wed]
  start: "09:00"
  end: "17:00"
  replicas: 5
```

**Common mistakes:**
- Start equals end (invalid)
- Wrong day abbreviations (use Mon, Tue, Wed, Thu, Fri, Sat, Sun)
- Time format wrong (use HH:MM)

**System Clock Skew:**
```bash
# Check if system time is accurate
date
# Compare with actual time

# If skewed, sync time (Linux)
sudo ntpdate -s time.nist.gov

# Or restart Docker (macOS)
```

---

### Issue: Cross-Midnight Windows Not Working

**Symptom:**
- Window spans midnight (e.g., 22:00-02:00) but doesn't scale correctly

**Diagnosis:**
```bash
# Check window configuration
kubectl get tws <name> -o yaml | grep -A 10 windows

# Check controller logs for window matching
make logs-controller | grep "Matched window"
```

**Resolution:**

**Expected Behavior:**
- Window from 22:00 to 02:00 on Friday means:
  - Friday 22:00 to Friday 23:59 (in window)
  - Saturday 00:00 to Saturday 01:59 (in window)

**Check Days Configuration:**
```yaml
# Must include the starting day
windows:
- days: [Fri]  # This is CORRECT for Fri 22:00 to Sat 02:00
  start: "22:00"
  end: "02:00"
  replicas: 5
```

**Don't include both days:**
```yaml
# WRONG: Don't do this
- days: [Fri, Sat]
  start: "22:00"
  end: "02:00"
```

---

### Issue: Manual Changes Keep Getting Reverted

**Symptom:**
- Manually scale deployment to X replicas
- Controller immediately scales back

**Diagnosis:**
```bash
# Check TWS effectiveReplicas
kubectl get tws <name> -o jsonpath='{.status.effectiveReplicas}'

# Check deployment replicas
kubectl get deployment <name> -o jsonpath='{.spec.replicas}'

# Check controller logs
make logs-controller | grep "Corrected manual drift"
```

**Resolution:**

**This is expected behavior** - Kyklos continuously enforces desired state.

**To allow manual changes:**

**Option 1: Pause the TWS**
```bash
kubectl patch tws <name> -n <namespace> --type=merge -p '{"spec":{"pause":true}}'

# Now manual changes won't be reverted
# Resume when ready
kubectl patch tws <name> -n <namespace> --type=merge -p '{"spec":{"pause":false}}'
```

**Option 2: Delete the TWS**
```bash
kubectl delete tws <name> -n <namespace>

# Deployment no longer managed
```

**Option 3: Update TWS to match desired state**
```bash
# If you want 7 replicas, update TWS
kubectl patch tws <name> -n <namespace> --type=merge -p '{"spec":{"defaultReplicas":7}}'
```

---

### Issue: Grace Period Not Working

**Symptom:**
- Grace period set but downscaling happens immediately

**Diagnosis:**
```bash
# Check grace period configuration
kubectl get tws <name> -o jsonpath='{.spec.gracePeriodSeconds}'

# Check controller logs
make logs-controller | grep "grace"
```

**Resolution:**

**Grace Period Only Applies to Downscaling:**
- Grace period delays scale-downs only
- Scale-ups are immediate

**Verify Correct Scenario:**
```bash
# Grace should apply when:
# 1. Leaving window that had higher replicas
# 2. New desired replicas < current replicas

# Example:
# In window: 10 replicas
# Leave window: defaultReplicas=2
# Grace period: 120 seconds
# Behavior: Maintain 10 replicas for 120s, then scale to 2
```

**Check Status During Grace:**
```bash
kubectl get tws <name> -o yaml | grep -A 5 status
# Should show:
#   effectiveReplicas: 10 (maintaining during grace)
#   conditions: may indicate grace period active
```

---

### Issue: Holidays Not Being Honored

**Symptom:**
- Holiday ConfigMap configured but windows still match on holidays

**Diagnosis:**
```bash
# Check holiday configuration
kubectl get tws <name> -o yaml | grep -A 5 holidays

# Check ConfigMap exists
kubectl get configmap <configmap-name> -n <namespace>

# Check ConfigMap data
kubectl get configmap <configmap-name> -o yaml | grep -A 10 data
```

**Resolution:**

**ConfigMap Not Found:**
```bash
# Create holiday ConfigMap
kubectl create configmap company-holidays -n <namespace> \
  --from-literal='2025-12-25'='' \
  --from-literal='2025-01-01'=''

# Verify
kubectl get configmap company-holidays -o yaml
```

**ConfigMap Format Wrong:**
```yaml
# Correct format:
data:
  "2025-12-25": ""
  "2025-01-01": ""
  "2025-07-04": ""

# Keys must be ISO dates: YYYY-MM-DD
# Values are ignored (can be empty or description)
```

**Wrong Namespace:**
```bash
# ConfigMap must be in same namespace as TWS
# Or controller needs RBAC to read from other namespace
```

**Check Controller Logs:**
```bash
make logs-controller | grep holiday
# Should show holiday detection if configured correctly
```

---

## RBAC and Permission Issues

### Issue: "Forbidden" Errors in Controller Logs

**Symptom:**
```
ERROR Failed to list *v1alpha1.TimeWindowScaler: timewindowscalers.kyklos.io is forbidden
ERROR Failed to update deployment: deployments.apps is forbidden
```

**Diagnosis:**
```bash
# Check RBAC configuration
make verify-rbac

# Check specific permissions
kubectl auth can-i list timewindowscalers --as=system:serviceaccount:kyklos-system:kyklos-controller
kubectl auth can-i update deployments --as=system:serviceaccount:kyklos-system:kyklos-controller
```

**Resolution:**

**Missing RBAC Resources:**
```bash
# Reapply all RBAC manifests
kubectl apply -f config/rbac/

# Verify ClusterRole exists
kubectl get clusterrole kyklos-controller-role

# Verify ClusterRoleBinding exists
kubectl get clusterrolebinding kyklos-controller-rolebinding
```

**Incorrect Permissions in ClusterRole:**
```bash
# Check current permissions
kubectl describe clusterrole kyklos-controller-role

# Should include:
# - timewindowscalers: get, list, watch, update, patch
# - timewindowscalers/status: get, update, patch
# - deployments: get, list, watch, update, patch
# - events: create, patch
# - configmaps: get, list, watch (for holidays)
```

**ServiceAccount Not Bound:**
```bash
# Check binding
kubectl describe clusterrolebinding kyklos-controller-rolebinding

# Should show:
# Subjects: ServiceAccount/kyklos-controller/kyklos-system

# If incorrect, reapply
kubectl apply -f config/rbac/role_binding.yaml
```

---

### Issue: Cannot Create Events

**Symptom:**
- Scaling happens but no events appear
- Controller logs show event creation errors

**Diagnosis:**
```bash
# Check event creation permission
kubectl auth can-i create events --as=system:serviceaccount:kyklos-system:kyklos-controller

# Check for permission errors in logs
make logs-controller | grep -i "failed to create event"
```

**Resolution:**

**Add Event Permissions:**
```bash
# Verify ClusterRole includes events
kubectl get clusterrole kyklos-controller-role -o yaml | grep -A 5 events

# Should show:
# - apiGroups: [""]
#   resources: [events]
#   verbs: [create, patch]

# If missing, add to config/rbac/role.yaml and reapply
```

---

## Performance Issues

### Issue: Controller Using Too Much CPU/Memory

**Symptom:**
```
NAME                                   CPU     MEMORY
kyklos-controller-manager-abc123-xyz   180m    450Mi
```

**Diagnosis:**
```bash
# Check resource usage
kubectl top pod -n kyklos-system

# Check resource limits
kubectl get deployment -n kyklos-system kyklos-controller-manager -o yaml | grep -A 10 resources
```

**Resolution:**

**High CPU:**

**Possible causes:**
- Too many TWS resources (high reconcile frequency)
- Reconcile loop errors causing tight loop
- Insufficient resource limits causing throttling

**Check:**
```bash
# Count TWS resources
kubectl get tws --all-namespaces | wc -l

# Check reconcile frequency in logs
make logs-controller | grep "Reconciling" | tail -50

# Look for errors
make logs-controller | grep ERROR
```

**Fix:**
```bash
# Increase CPU limits if justified
kubectl patch deployment -n kyklos-system kyklos-controller-manager -p '{"spec":{"template":{"spec":{"containers":[{"name":"manager","resources":{"limits":{"cpu":"500m"}}}]}}}}'

# Or reduce reconcile frequency (code change)
```

**High Memory:**

**Possible causes:**
- Memory leak in controller
- Too many cached objects
- Large number of TWS resources

**Check:**
```bash
# Check for memory leaks over time
watch kubectl top pod -n kyklos-system

# Monitor for 5-10 minutes, check if memory grows continuously
```

**Fix:**
```bash
# Restart controller to reclaim memory
make restart-controller

# If leak persists, investigate controller code
# Enable profiling and collect pprof data
```

---

### Issue: Slow Reconcile Times

**Symptom:**
- Scaling takes longer than expected
- Minutes between window boundary and actual scale

**Diagnosis:**
```bash
# Check reconcile duration metrics
make port-forward-metrics
curl http://localhost:8080/metrics | grep reconcile_duration

# Check controller logs for timing
make logs-controller | grep "Reconcile completed"
```

**Resolution:**

**Cluster API Server Slow:**
```bash
# Check API server response time
time kubectl get nodes

# If slow (>2s), cluster may be overloaded
# Check cluster resources
kubectl top nodes
```

**Too Many TWS Resources:**
```bash
# Count TWS resources
kubectl get tws --all-namespaces | wc -l

# If >100, consider:
# - Increasing controller replicas (not supported in v0.1)
# - Optimizing reconcile logic
# - Using webhooks for filtering
```

**Network Issues:**
```bash
# Check network latency to API server
kubectl cluster-info | grep "Kubernetes control plane"
# Note the IP and test latency
ping -c 10 <api-server-ip>
```

---

## Testing Issues

### Issue: `make test` Fails

**Symptom:**
```
FAIL github.com/your-org/kyklos/controllers
```

**Diagnosis:**
```bash
# Run tests with verbose output
make test | tee test-output.log

# Check specific test failure
go test -v ./controllers -run TestTimeWindowScalerController
```

**Resolution:**

**Import Errors:**
```bash
# Update dependencies
go mod tidy
go mod download

# Regenerate code
make generate

# Retry
make test
```

**Test Logic Errors:**
```bash
# Run specific failing test
go test -v ./path/to/package -run TestName

# Add debug output to test
# Fix test or code
```

**Race Conditions:**
```bash
# Run with race detector
make build-race
go test -race ./...
```

---

### Issue: `make test-integration` Fails to Start envtest

**Symptom:**
```
Error: failed to start test control plane
```

**Diagnosis:**
```bash
# Check if setup-envtest is installed
ls tools/bin/setup-envtest

# Check if test binaries are downloaded
setup-envtest list
```

**Resolution:**

**Missing setup-envtest:**
```bash
# Reinstall tools
rm -rf tools/bin
make tools
```

**Envtest Binaries Not Downloaded:**
```bash
# Download envtest binaries
tools/bin/setup-envtest use 1.28

# Retry tests
make test-integration
```

**Port Already in Use:**
```bash
# Envtest uses random ports, but may conflict
# Kill any hanging test processes
pkill -f envtest

# Retry
make test-integration
```

---

## Reset Procedures

### Complete Environment Reset

When all else fails, start fresh:

```bash
# 1. Delete cluster
make cluster-down

# 2. Clean local artifacts
make clean

# 3. Clean Docker images
docker rmi kyklos/controller:dev kyklos/controller:latest

# 4. Clean Go cache (nuclear option)
go clean -cache -modcache -testcache

# 5. Verify tools still work
make verify-tools

# 6. Recreate everything
make cluster-up
make tools
make manifests generate build docker-build kind-load
make install-crds deploy

# 7. Verify
make verify-all
```

**Time:** 4-5 minutes

---

### Partial Resets

**Reset Cluster Only:**
```bash
make cluster-down cluster-up
# Then redeploy: make install-crds deploy
```

**Reset Controller Only:**
```bash
make undeploy deploy
```

**Reset Images Only:**
```bash
docker rmi kyklos/controller:dev
make docker-build kind-load deploy-rollout
```

**Reset Generated Code:**
```bash
rm -rf api/v1alpha1/zz_generated.deepcopy.go
rm -rf config/crd/bases/*.yaml
make manifests generate
```

---

## Getting Help

### Collecting Debug Information

When reporting issues, collect this information:

```bash
# 1. Environment info
make verify-tools > debug-info.txt
kubectl version >> debug-info.txt
docker version >> debug-info.txt

# 2. Cluster state
kubectl get nodes -o yaml >> debug-info.txt
kubectl get pods --all-namespaces >> debug-info.txt

# 3. Controller logs
make logs-controller >> debug-info.txt

# 4. Resource status
kubectl get tws --all-namespaces -o yaml >> debug-info.txt
kubectl get events --all-namespaces >> debug-info.txt

# 5. RBAC info
kubectl get clusterrole kyklos-controller-role -o yaml >> debug-info.txt
kubectl get clusterrolebinding kyklos-controller-rolebinding -o yaml >> debug-info.txt

# Share debug-info.txt
```

---

### Common Diagnostic Commands

**Quick health check:**
```bash
make verify-all
```

**Controller diagnostics:**
```bash
kubectl describe pod -n kyklos-system -l app=kyklos-controller
make logs-controller
kubectl get events -n kyklos-system
```

**Resource diagnostics:**
```bash
kubectl get tws --all-namespaces -o yaml
kubectl get deployments --all-namespaces
kubectl get events --all-namespaces --sort-by='.lastTimestamp' | tail -50
```

**RBAC diagnostics:**
```bash
make verify-rbac
kubectl auth can-i --list --as=system:serviceaccount:kyklos-system:kyklos-controller
```

---

### Additional Resources

- [LOCAL-DEV-GUIDE.md](./LOCAL-DEV-GUIDE.md) - Setup instructions
- [VERIFY-CHECKLIST.md](./VERIFY-CHECKLIST.md) - Health checks
- [MINUTE-DEMO.md](./MINUTE-DEMO.md) - Working example
- [MAKE-TARGETS.md](./MAKE-TARGETS.md) - Make command reference
- [CRD-SPEC.md](./api/CRD-SPEC.md) - API reference

---

### Still Stuck?

1. Check if it's a known issue in project docs
2. Search controller logs for specific error messages
3. Try a complete environment reset
4. Collect debug information (see above)
5. Open an issue with:
   - Symptom description
   - Steps to reproduce
   - debug-info.txt contents
   - What you've tried already
