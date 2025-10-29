# Local Development Guide

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** local-workflow-designer

This guide walks you through setting up a complete local development environment for Kyklos, from clean machine to observing your first time-based scale event in under 15 minutes.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start (15-Minute Path)](#quick-start-15-minute-path)
3. [Detailed Setup Steps](#detailed-setup-steps)
4. [Development Workflow](#development-workflow)
5. [Verification Steps](#verification-steps)
6. [Common Tasks](#common-tasks)

---

## Prerequisites

### Required Tools

The following tools must be installed before starting. Time estimates assume fresh installations.

| Tool | Version | Installation Time | Purpose |
|------|---------|------------------|---------|
| Go | 1.21+ | 2-3 minutes | Build controller |
| Docker | 24.0+ | 3-5 minutes | Build container images |
| Kind or k3d | Latest | 1-2 minutes | Local Kubernetes cluster |
| kubectl | 1.28+ | 1-2 minutes | Interact with cluster |
| make | Any | Pre-installed (macOS/Linux) | Run build targets |
| git | Any | Pre-installed | Version control |

### Optional Tools (Recommended)

| Tool | Purpose | Installation |
|------|---------|--------------|
| kustomize | Manifest customization | `brew install kustomize` |
| controller-gen | CRD generation | Installed via `make tools` |
| golangci-lint | Code linting | Installed via `make tools` |
| watch | Monitor resources | `brew install watch` |
| jq | JSON parsing | `brew install jq` |

### Installation Commands

**macOS (Homebrew):**
```bash
# Core tools
brew install go docker kubectl kind

# Start Docker
open -a Docker

# Optional tools
brew install kustomize watch jq
```

**Linux (Ubuntu/Debian):**
```bash
# Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

### Verification

Run this command to verify all prerequisites:

```bash
make verify-tools
```

Expected output:
```
Checking prerequisites...
✓ Go 1.21.5 installed
✓ Docker 24.0.6 running
✓ kubectl 1.28.3 installed
✓ Kind 0.20.0 installed
✓ All prerequisites satisfied
```

---

## Quick Start (15-Minute Path)

This is the fastest path from zero to observing scale events. Total time: 12-15 minutes.

### Phase 1: Environment Setup (3-4 minutes)

```bash
# Clone repository
git clone https://github.com/your-org/kyklos.git
cd kyklos

# Install development tools
make tools
# Expected: 1-2 minutes, downloads controller-gen, golangci-lint

# Create local Kubernetes cluster
make cluster-up
# Expected: 60-90 seconds, creates kind cluster "kyklos-dev"
```

Success indicators:
- `tools/bin/controller-gen` exists
- `kubectl cluster-info` shows cluster running
- `kubectl get nodes` shows 1 node Ready

### Phase 2: Build and Deploy (4-5 minutes)

```bash
# Generate CRDs and build controller
make build
# Expected: 30-60 seconds, produces bin/controller

# Build Docker image
make docker-build
# Expected: 60-120 seconds, creates kyklos/controller:dev

# Load image into cluster
make kind-load
# Expected: 20-30 seconds, loads image into kind nodes

# Install CRDs
make install-crds
# Expected: 5-10 seconds, applies CRD manifests

# Deploy controller
make deploy
# Expected: 20-30 seconds, applies controller manifests
```

Success indicators:
- `bin/controller` binary exists
- `docker images | grep kyklos` shows image
- `kubectl get crd timewindowscalers.kyklos.io` shows CRD
- `kubectl get pods -n kyklos-system` shows controller pod Running

### Phase 3: Test with Minute-Scale Demo (5-6 minutes)

```bash
# Create demo namespace and target deployment
make demo-setup
# Expected: 10-15 seconds, creates demo-app with 1 replica

# Apply minute-scale TimeWindowScaler
make demo-apply-minute
# Expected: 5 seconds, creates TWS with 1-minute windows

# Watch scaling in action
make demo-watch
# Expected: Watch for 3-4 minutes to see scale events

# Observe controller logs
make logs-controller
# Shows reconciliation and scaling decisions
```

Success indicators:
- See `demo-app` scale from 1 to 5 replicas within first minute
- See scale-up event: `kubectl get events -n demo`
- Controller logs show: "Scaled deployment demo-app from 1 to 5 replicas"

### Phase 4: Verify and Explore (2-3 minutes)

```bash
# Check TimeWindowScaler status
kubectl get tws -n demo demo-minute-scaler -o yaml

# Verify current state
make verify-demo
# Shows TWS status, deployment replicas, recent events

# Clean up demo
make demo-cleanup
# Removes demo namespace
```

**Total Time:** 12-15 minutes from zero to observed scale event

---

## Detailed Setup Steps

### Step 1: Repository Setup

```bash
# Clone repository
git clone https://github.com/your-org/kyklos.git
cd kyklos

# Verify repository structure
ls -la
# Expected: api/, cmd/, config/, controllers/, docs/, examples/, Makefile
```

### Step 2: Install Development Tools

Kyklos uses code generation tools that must be installed locally:

```bash
make tools
```

This installs:
- `controller-gen` v0.13.0 - Generates CRDs and RBAC from Go markers
- `golangci-lint` v1.54.2 - Lints Go code
- `setup-envtest` - Downloads envtest binaries for integration tests

Installation location: `tools/bin/` (gitignored)

Time: 1-2 minutes

Verification:
```bash
ls -la tools/bin/
# Expected: controller-gen, golangci-lint, setup-envtest
```

### Step 3: Create Local Cluster

Kyklos supports both Kind and k3d. Choose one:

**Option A: Kind (Recommended)**

```bash
make cluster-up
```

This creates a Kind cluster named `kyklos-dev` with:
- Single control-plane node
- Docker port mappings for NodePort services
- CNI: kindnet
- Kubernetes version: 1.28

Time: 60-90 seconds

Configuration: See `config/kind-cluster.yaml` for customization

**Option B: k3d**

```bash
make cluster-up-k3d
```

This creates a k3d cluster named `kyklos-dev` with:
- Single server node
- Local registry enabled
- Traefik disabled (reduces resource usage)

Time: 30-60 seconds

**Verification:**

```bash
kubectl cluster-info
kubectl get nodes
# Expected: 1 node in Ready state
```

**Switching Contexts:**

If you have multiple clusters:
```bash
kubectl config use-context kind-kyklos-dev
# or
kubectl config use-context k3d-kyklos-dev
```

### Step 4: Build Controller

```bash
# Generate CRD manifests and Go code
make manifests generate
# Time: 10-20 seconds
# Outputs: config/crd/bases/*.yaml, api/v1alpha1/zz_generated.deepcopy.go

# Build controller binary
make build
# Time: 30-60 seconds
# Output: bin/controller
```

The build process:
1. Runs `go mod download` to fetch dependencies
2. Compiles `cmd/controller/main.go`
3. Produces static binary in `bin/controller`

**Build Options:**

Fast iteration (skip tests):
```bash
make build-fast
```

With race detector:
```bash
make build-race
```

Cross-compile for Linux (from macOS):
```bash
make build-linux
```

**Verification:**

```bash
./bin/controller --version
# Expected: kyklos-controller v0.1.0-dev
```

### Step 5: Build Container Image

```bash
make docker-build
```

This builds a multi-stage Docker image:
- Stage 1: Build binary in golang:1.21 builder
- Stage 2: Copy binary to distroless/static:nonroot
- Final image: ~20MB

Time: 60-120 seconds (longer on first build)

Tags created:
- `kyklos/controller:dev` - Development tag
- `kyklos/controller:latest` - Latest stable

**Custom Tags:**

```bash
make docker-build IMG=kyklos/controller:v0.1.0
```

**Verification:**

```bash
docker images | grep kyklos
# Expected: kyklos/controller with dev and latest tags
```

### Step 6: Load Image to Cluster

**For Kind:**
```bash
make kind-load
```

This loads the image into Kind cluster nodes without pushing to registry.

Time: 20-30 seconds

**For k3d:**
```bash
make k3d-load
```

**Verification:**

```bash
docker exec -it kyklos-dev-control-plane crictl images | grep kyklos
# Expected: kyklos/controller:dev listed
```

### Step 7: Install CRDs

```bash
make install-crds
```

This applies CRD manifests from `config/crd/` to cluster.

Time: 5-10 seconds

**What Gets Installed:**
- `timewindowscalers.kyklos.io` CRD with v1alpha1 version
- OpenAPI validation schema
- Printer columns for `kubectl get tws`

**Verification:**

```bash
kubectl get crd timewindowscalers.kyklos.io
kubectl describe crd timewindowscalers.kyklos.io
```

Expected output shows:
- Stored version: v1alpha1
- Accepted names: timewindowscaler, tws
- Status: Established

### Step 8: Deploy Controller

```bash
make deploy
```

This applies all manifests from `config/default/` via kustomize:
- Namespace: `kyklos-system`
- ServiceAccount: `kyklos-controller`
- ClusterRole + ClusterRoleBinding: Permissions for deployments, events
- Deployment: Controller pod with resource limits

Time: 20-30 seconds

**Deployment Configuration:**
- Replicas: 1
- Resources: 100m CPU, 128Mi memory (requests/limits)
- Image pull policy: IfNotPresent (uses locally loaded image)
- Restart policy: Always

**Verification:**

```bash
kubectl get pods -n kyklos-system
# Expected: kyklos-controller-manager-xxx-xxx Running

kubectl logs -n kyklos-system -l app=kyklos-controller --tail=50
# Expected: "Starting controller" log messages
```

**Health Check:**

```bash
make verify-controller
```

Expected output:
```
Checking controller health...
✓ Pod kyklos-controller-manager-xxx Running
✓ Controller logs show successful startup
✓ Metrics endpoint responding
Controller ready to reconcile TimeWindowScalers
```

### Step 9: Verify Complete Setup

Run the complete verification checklist:

```bash
make verify-all
```

This runs:
1. Tool verification
2. Cluster connectivity check
3. CRD installation check
4. Controller health check
5. RBAC permissions validation

Time: 10-15 seconds

---

## Development Workflow

### Typical Development Cycle

**1. Make Code Changes**

Edit files in `api/`, `controllers/`, or `internal/`

**2. Run Tests**

```bash
# Unit tests only (fast)
make test

# With coverage report
make test-coverage

# Integration tests (slower, uses envtest)
make test-integration

# All tests
make test-all
```

**3. Rebuild and Redeploy**

```bash
# Full rebuild: manifests, code, image, deploy
make redeploy

# Time: 2-3 minutes
```

This target chains:
- `make manifests` - Regenerate CRDs from Go types
- `make generate` - Regenerate deepcopy code
- `make build` - Compile controller
- `make docker-build` - Build container image
- `make kind-load` - Load image to cluster
- `make deploy-rollout` - Restart controller pod

**4. Verify Changes**

```bash
# Watch controller logs
make logs-controller-follow

# Watch events
kubectl get events -n kyklos-system --watch

# Test with sample
kubectl apply -f config/samples/basic.yaml
kubectl get tws basic -o yaml
```

**5. Debug Issues**

```bash
# Increase log verbosity
make deploy-debug

# Port-forward to metrics endpoint
make port-forward-metrics
curl localhost:8080/metrics | grep kyklos

# Describe controller pod
kubectl describe pod -n kyklos-system -l app=kyklos-controller

# Get controller events
kubectl get events -n kyklos-system --field-selector involvedObject.name=kyklos-controller-manager
```

### Fast Iteration Tips

**Skip Docker Build (Run Locally):**

For testing reconcile logic without cluster:
```bash
# Run controller locally against remote cluster
make run-local
```

This runs the controller as a local process with kubeconfig authentication.

Benefits:
- No Docker build (saves 60-90 seconds)
- Easy debugging with delve
- Faster code-test-fix cycle

**Use envtest for Testing:**

Instead of full cluster deployment:
```bash
make test-integration
```

This uses controller-runtime's envtest for fast integration tests.

**Incremental Builds:**

Go's build cache makes subsequent builds fast. To ensure clean builds:
```bash
make clean build
```

### Making API Changes

When modifying `api/v1alpha1/timewindowscaler_types.go`:

```bash
# 1. Edit Go types and kubebuilder markers
vim api/v1alpha1/timewindowscaler_types.go

# 2. Regenerate manifests and code
make manifests generate

# 3. Verify CRD changes
git diff config/crd/bases/

# 4. Update sample manifests if needed
vim config/samples/basic.yaml

# 5. Rebuild and deploy
make redeploy

# 6. Test changes
kubectl apply -f config/samples/basic.yaml
kubectl describe tws basic
```

---

## Verification Steps

### Quick Health Check

```bash
make verify-health
```

Checks:
- Cluster reachable
- CRDs installed
- Controller running
- No crash loops

Time: 5-10 seconds

### Full System Verification

```bash
make verify-all
```

Checks everything including:
- Tool versions
- Cluster resources
- RBAC permissions
- Controller logs for errors

Time: 15-20 seconds

### Manual Verification

**Check Cluster:**
```bash
kubectl cluster-info
kubectl get nodes
```

**Check CRDs:**
```bash
kubectl get crd | grep kyklos
kubectl api-resources | grep timewindowscaler
```

**Check Controller:**
```bash
kubectl get pods -n kyklos-system
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=100
```

**Check RBAC:**
```bash
kubectl auth can-i list deployments --as=system:serviceaccount:kyklos-system:kyklos-controller
# Expected: yes

kubectl auth can-i create events --as=system:serviceaccount:kyklos-system:kyklos-controller
# Expected: yes
```

---

## Common Tasks

### Restart Controller

```bash
make restart-controller
```

Deletes controller pod, triggering recreation with latest image.

### View Logs

```bash
# Last 100 lines
make logs-controller

# Follow logs
make logs-controller-follow

# With grep filter
make logs-controller | grep "Reconciling"
```

### Clean Up Resources

```bash
# Remove all TimeWindowScaler instances
make cleanup-tws

# Remove controller (keeps CRDs)
make undeploy

# Remove CRDs (deletes all TWS instances!)
make uninstall-crds

# Delete cluster
make cluster-down

# Complete cleanup (cluster + local binaries)
make clean-all
```

### Reset Environment

Start completely fresh:

```bash
make reset-env
```

This runs:
1. `make cluster-down` - Delete cluster
2. `make clean` - Remove build artifacts
3. `make cluster-up` - Create new cluster
4. Full rebuild and deploy

Time: 3-4 minutes

### Update Dependencies

```bash
# Update Go modules
go get -u ./...
go mod tidy

# Rebuild with new dependencies
make build
```

### Run Linter

```bash
make lint
```

Uses golangci-lint with project configuration (`.golangci.yml`).

Auto-fix issues:
```bash
make lint-fix
```

### Generate Documentation

```bash
# Generate CRD reference docs
make docs-api

# Generate controller metrics docs
make docs-metrics
```

---

## Next Steps

- Follow [MINUTE-DEMO.md](./user/MINUTE-DEMO.md) for a 10-minute walkthrough
- See [VERIFY-CHECKLIST.md](./VERIFY-CHECKLIST.md) for detailed health checks
- Consult [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) if issues arise
- Review [MAKE-TARGETS.md](./MAKE-TARGETS.md) for complete target reference

---

## Troubleshooting Quick Links

**Controller Not Starting:**
- Check logs: `make logs-controller`
- Verify image: `make verify-image-loaded`
- Check RBAC: `make verify-rbac`

**CRD Issues:**
- Reinstall: `make uninstall-crds install-crds`
- Verify: `kubectl get crd timewindowscalers.kyklos.io -o yaml`

**Cluster Issues:**
- Reset cluster: `make cluster-down cluster-up`
- Check Docker: `docker ps` (ensure Docker running)

**Build Issues:**
- Clean rebuild: `make clean build`
- Verify tools: `make verify-tools`

For detailed troubleshooting, see [TROUBLESHOOTING.md](./TROUBLESHOOTING.md).
