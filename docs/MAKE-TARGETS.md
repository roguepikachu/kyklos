# Makefile Targets Reference

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** local-workflow-designer

This document provides a complete reference of all Make targets in the Kyklos project, organized by workflow phase. Each target includes intent, expected duration, dependencies, and success indicators.

---

## Table of Contents

1. [Target Overview](#target-overview)
2. [Setup and Installation](#setup-and-installation)
3. [Build Targets](#build-targets)
4. [Cluster Management](#cluster-management)
5. [Deploy and Install](#deploy-and-install)
6. [Testing Targets](#testing-targets)
7. [Development Workflow](#development-workflow)
8. [Verification and Validation](#verification-and-validation)
9. [Demo and Examples](#demo-and-examples)
10. [Cleanup Targets](#cleanup-targets)
11. [Utility Targets](#utility-targets)
12. [Target Dependency Graph](#target-dependency-graph)

---

## Target Overview

### Quick Reference Table

| Target | Time | Phase | Common Use |
|--------|------|-------|------------|
| `verify-tools` | 5s | Setup | Check prerequisites |
| `tools` | 60-120s | Setup | Install dev tools |
| `cluster-up` | 60-90s | Setup | Create local cluster |
| `build` | 30-60s | Build | Compile controller |
| `docker-build` | 60-120s | Build | Create container image |
| `install-crds` | 5-10s | Deploy | Install CRDs |
| `deploy` | 20-30s | Deploy | Deploy controller |
| `test` | 20-30s | Test | Run unit tests |
| `demo-minute` | 300s | Demo | Quick scale demo |
| `clean-all` | 30-60s | Cleanup | Full cleanup |

### Target Categories

- **Prerequisites:** Tool and environment verification
- **Build:** Code generation, compilation, images
- **Cluster:** Local Kubernetes cluster lifecycle
- **Deploy:** Install CRDs and controller to cluster
- **Test:** Unit, integration, and e2e tests
- **Demo:** Example scenarios and walkthroughs
- **Cleanup:** Resource removal and environment reset
- **Utility:** Helper targets for common tasks

---

## Setup and Installation

### verify-tools

**Intent:** Verify all prerequisite tools are installed and correct versions

**Duration:** 5 seconds

**Dependencies:** None

**Usage:**
```bash
make verify-tools
```

**What It Checks:**
- Go version 1.21+
- Docker installed and daemon running
- kubectl installed and functional
- Kind or k3d installed
- make available

**Success Indicator:**
```
Checking prerequisites...
✓ Go 1.21.5 installed
✓ Docker 24.0.6 running
✓ kubectl 1.28.3 installed
✓ Kind 0.20.0 installed
✓ All prerequisites satisfied
```

**Failure Actions:**
- Prints missing tools with installation instructions
- Exits with code 1
- Provides links to installation guides

**Environment Variables:**
- `MIN_GO_VERSION` - Override minimum Go version (default: 1.21)
- `SKIP_DOCKER_CHECK` - Skip Docker verification (default: false)

---

### tools

**Intent:** Install development tools needed for code generation and linting

**Duration:** 60-120 seconds (first run), 5 seconds (subsequent, cached)

**Dependencies:** Go installed

**Usage:**
```bash
make tools
```

**What It Installs:**
- `controller-gen` v0.13.0 - Generates CRDs and RBAC from kubebuilder markers
- `golangci-lint` v1.54.2 - Comprehensive Go linter
- `setup-envtest` - Downloads kubebuilder test binaries

**Installation Location:** `tools/bin/` (gitignored)

**Success Indicator:**
```
Installing development tools...
→ controller-gen v0.13.0
→ golangci-lint v1.54.2
→ setup-envtest latest
✓ Tools installed to tools/bin/
```

**Idempotency:** Safe to run multiple times, skips if already installed

**Environment Variables:**
- `CONTROLLER_GEN_VERSION` - Override controller-gen version
- `GOLANGCI_LINT_VERSION` - Override golangci-lint version

**Troubleshooting:**
```bash
# Force reinstall
rm -rf tools/bin && make tools

# Verify installation
ls -la tools/bin/
```

---

## Build Targets

### manifests

**Intent:** Generate CRD and RBAC manifests from Go type definitions and kubebuilder markers

**Duration:** 10-20 seconds

**Dependencies:** `tools` (requires controller-gen)

**Usage:**
```bash
make manifests
```

**What It Generates:**
- `config/crd/bases/kyklos.io_timewindowscalers.yaml` - CRD with OpenAPI schema
- `config/rbac/role.yaml` - RBAC permissions from markers

**Source Files:**
- `api/v1alpha1/timewindowscaler_types.go` - API definitions
- Controller files with `+kubebuilder:rbac` markers

**Success Indicator:**
```
Generating manifests...
→ CRDs: config/crd/bases/
→ RBAC: config/rbac/
✓ Manifests generated
```

**When to Run:**
- After modifying API types in `api/v1alpha1/`
- After adding/changing `+kubebuilder:rbac` markers
- After adding/changing `+kubebuilder:validation` markers

**Verification:**
```bash
# Check generated files
git status
git diff config/crd/bases/
```

---

### generate

**Intent:** Generate deepcopy methods for Kubernetes API types

**Duration:** 5-10 seconds

**Dependencies:** `tools` (requires controller-gen)

**Usage:**
```bash
make generate
```

**What It Generates:**
- `api/v1alpha1/zz_generated.deepcopy.go` - DeepCopy, DeepCopyInto, DeepCopyObject methods

**Success Indicator:**
```
Generating code...
→ api/v1alpha1/zz_generated.deepcopy.go
✓ Code generation complete
```

**When to Run:**
- After modifying struct definitions in API types
- After adding new fields to Spec or Status

**Note:** This file should never be manually edited - it's fully generated.

---

### build

**Intent:** Compile controller binary for local architecture

**Duration:** 30-60 seconds (first build), 10-20 seconds (incremental)

**Dependencies:** `manifests`, `generate`

**Usage:**
```bash
make build
```

**What It Does:**
1. Runs `go mod download` - Fetches dependencies
2. Runs `go build` - Compiles cmd/controller/main.go
3. Places binary in `bin/controller`

**Build Flags:**
- `-ldflags "-s -w"` - Strip debug info, reduce binary size
- CGO_ENABLED=0 - Static binary, no C dependencies

**Success Indicator:**
```
Building controller...
→ Fetching dependencies
→ Compiling cmd/controller/main.go
✓ Binary created: bin/controller (12.4 MB)
```

**Verification:**
```bash
./bin/controller --version
# Expected: kyklos-controller v0.1.0-dev
```

**Environment Variables:**
- `GOOS` - Target OS (default: host OS)
- `GOARCH` - Target architecture (default: host arch)

**Related Targets:**
- `build-fast` - Skip tests and linting
- `build-race` - Enable race detector
- `build-linux` - Cross-compile for Linux

---

### build-fast

**Intent:** Fast build for rapid iteration, skipping tests and checks

**Duration:** 15-30 seconds

**Dependencies:** None (skips manifests/generate)

**Usage:**
```bash
make build-fast
```

**Use Case:** When iterating on controller logic with no API changes

**What It Skips:**
- Code generation
- Manifest generation
- Dependency checks

**Warning:** May produce stale artifacts if API types changed

---

### build-race

**Intent:** Build controller with race detector enabled

**Duration:** 45-75 seconds

**Dependencies:** `manifests`, `generate`

**Usage:**
```bash
make build-race
```

**What It Does:**
- Adds `-race` flag to go build
- Instruments code to detect data races
- Produces larger binary (~3x size)

**Use Case:**
- Debugging concurrency issues
- Running integration tests with race detection

**Verification:**
```bash
./bin/controller --version
# Expected: includes "race detector enabled"
```

---

### build-linux

**Intent:** Cross-compile controller for Linux (from macOS/Windows)

**Duration:** 30-60 seconds

**Dependencies:** `manifests`, `generate`

**Usage:**
```bash
make build-linux
```

**Environment Variables Set:**
- `GOOS=linux`
- `GOARCH=amd64`

**Output:** `bin/controller-linux-amd64`

**Use Case:** Building for container image on non-Linux host

---

### docker-build

**Intent:** Build container image with controller binary

**Duration:** 60-120 seconds (first build), 30-60 seconds (cached layers)

**Dependencies:** None (builds in container)

**Usage:**
```bash
make docker-build

# With custom tag
make docker-build IMG=kyklos/controller:v0.1.0
```

**What It Does:**
1. Multi-stage build:
   - Stage 1: golang:1.21 - Build binary
   - Stage 2: distroless/static:nonroot - Runtime
2. Tags image as `kyklos/controller:dev` and `kyklos/controller:latest`
3. Uses BuildKit for caching

**Image Details:**
- Base: distroless/static:nonroot
- Size: ~20 MB
- User: nonroot (UID 65532)
- Entrypoint: /controller

**Success Indicator:**
```
Building container image...
[+] Building 65.3s (15/15) FINISHED
→ Image: kyklos/controller:dev
→ Size: 19.8 MB
✓ Image built successfully
```

**Verification:**
```bash
docker images | grep kyklos
docker run --rm kyklos/controller:dev --version
```

**Environment Variables:**
- `IMG` - Image name and tag (default: kyklos/controller:dev)
- `DOCKER_BUILDKIT` - Enable BuildKit (default: 1)

---

### docker-build-nocache

**Intent:** Build container image without using Docker layer cache

**Duration:** 120-180 seconds

**Usage:**
```bash
make docker-build-nocache
```

**Use Case:**
- Forcing fresh build after upstream image updates
- Troubleshooting cache-related build issues

---

## Cluster Management

### cluster-up

**Intent:** Create local Kind cluster for development

**Duration:** 60-90 seconds

**Dependencies:** Kind installed

**Usage:**
```bash
make cluster-up
```

**What It Creates:**
- Cluster name: `kyklos-dev`
- Nodes: 1 control-plane
- Kubernetes version: 1.28
- CNI: kindnet
- Port mappings: 30000-32767 (NodePort range)

**Configuration:** Uses `config/kind-cluster.yaml`

**Success Indicator:**
```
Creating Kind cluster...
Creating cluster "kyklos-dev" ...
✓ Ensuring node image (kindest/node:v1.28.0)
✓ Preparing nodes
✓ Writing configuration
✓ Starting control-plane
✓ Installing CNI
✓ Installing StorageClass
Cluster kyklos-dev created successfully
```

**Verification:**
```bash
kubectl cluster-info --context kind-kyklos-dev
kubectl get nodes
```

**Idempotency:** Safe to run multiple times - checks if cluster exists first

**Automatic Actions:**
- Sets kubectl context to new cluster
- Creates local kubeconfig entry

---

### cluster-up-k3d

**Intent:** Create local k3d cluster as alternative to Kind

**Duration:** 30-60 seconds

**Dependencies:** k3d installed

**Usage:**
```bash
make cluster-up-k3d
```

**What It Creates:**
- Cluster name: `kyklos-dev`
- Nodes: 1 server
- Port mappings: 8080:80, 8443:443
- Traefik: Disabled (saves resources)

**Advantages Over Kind:**
- Faster startup (~30% faster)
- Lower memory footprint
- Built-in local registry support

**Success Indicator:**
```
Creating k3d cluster...
INFO[0000] Creating cluster 'kyklos-dev'
INFO[0015] Cluster 'kyklos-dev' created successfully
✓ Cluster ready
```

---

### cluster-down

**Intent:** Delete local cluster and cleanup resources

**Duration:** 10-20 seconds

**Dependencies:** None

**Usage:**
```bash
make cluster-down
```

**What It Deletes:**
- Kind cluster `kyklos-dev` (if exists)
- k3d cluster `kyklos-dev` (if exists)
- kubectl context entries

**Success Indicator:**
```
Deleting cluster...
Deleting cluster "kyklos-dev" ...
✓ Cluster deleted
```

**Warning:** This permanently deletes all resources in the cluster

**Verification:**
```bash
kind get clusters
# Expected: kyklos-dev not listed
```

---

### cluster-status

**Intent:** Show current cluster state and resource usage

**Duration:** 2-5 seconds

**Dependencies:** Cluster running

**Usage:**
```bash
make cluster-status
```

**What It Shows:**
- Cluster info and endpoint
- Node status and resource usage
- Pod count by namespace
- CRD installation status

**Success Indicator:**
```
Cluster Status:
→ Name: kind-kyklos-dev
→ Kubernetes: v1.28.0
→ Nodes: 1/1 Ready
→ Pods: 12/15 Running
→ CRDs: timewindowscalers.kyklos.io installed
✓ Cluster healthy
```

---

## Deploy and Install

### install-crds

**Intent:** Install TimeWindowScaler CRD to cluster

**Duration:** 5-10 seconds

**Dependencies:** Cluster running, `manifests` generated

**Usage:**
```bash
make install-crds
```

**What It Installs:**
```bash
kubectl apply -f config/crd/bases/kyklos.io_timewindowscalers.yaml
```

**Success Indicator:**
```
Installing CRDs...
customresourcedefinition.apiextensions.k8s.io/timewindowscalers.kyklos.io created
✓ CRDs installed
```

**Verification:**
```bash
kubectl get crd timewindowscalers.kyklos.io
kubectl explain timewindowscaler.spec
```

**Idempotency:** Safe to run multiple times (applies with server-side apply)

---

### uninstall-crds

**Intent:** Remove CRDs from cluster

**Duration:** 5-10 seconds

**Dependencies:** None

**Usage:**
```bash
make uninstall-crds
```

**Warning:** Deletes all TimeWindowScaler instances in cluster

**What It Does:**
```bash
kubectl delete crd timewindowscalers.kyklos.io
```

**Success Indicator:**
```
Uninstalling CRDs...
customresourcedefinition.apiextensions.k8s.io "timewindowscalers.kyklos.io" deleted
✓ CRDs removed
```

---

### deploy

**Intent:** Deploy controller to cluster with full configuration

**Duration:** 20-30 seconds

**Dependencies:** Cluster running, CRDs installed, image loaded

**Usage:**
```bash
make deploy
```

**What It Deploys:**
```bash
kubectl apply -k config/default/
```

This creates:
- Namespace: `kyklos-system`
- ServiceAccount: `kyklos-controller`
- ClusterRole: `kyklos-controller-role` (permissions for deployments, events, TWS)
- ClusterRoleBinding: Binds role to service account
- Deployment: Controller pod with 1 replica

**Resource Limits:**
- CPU: 100m request, 200m limit
- Memory: 128Mi request, 256Mi limit

**Success Indicator:**
```
Deploying controller...
namespace/kyklos-system created
serviceaccount/kyklos-controller created
clusterrole.rbac.authorization.k8s.io/kyklos-controller-role created
clusterrolebinding.rbac.authorization.k8s.io/kyklos-controller-rolebinding created
deployment.apps/kyklos-controller-manager created
✓ Controller deployed
```

**Verification:**
```bash
kubectl get pods -n kyklos-system
kubectl logs -n kyklos-system -l app=kyklos-controller
```

**Image Pull Policy:** IfNotPresent (uses locally loaded image)

---

### deploy-debug

**Intent:** Deploy controller with verbose logging and debug configuration

**Duration:** 20-30 seconds

**Dependencies:** Same as `deploy`

**Usage:**
```bash
make deploy-debug
```

**Differences from `deploy`:**
- Log level: debug (vs info)
- Args: `--zap-log-level=debug --zap-development=true`
- Additional flags: `--leader-elect=false` (for single instance)

**Use Case:**
- Troubleshooting reconcile issues
- Observing detailed state transitions
- Development with verbose output

---

### undeploy

**Intent:** Remove controller from cluster (keeps CRDs)

**Duration:** 10-15 seconds

**Dependencies:** None

**Usage:**
```bash
make undeploy
```

**What It Removes:**
```bash
kubectl delete -k config/default/
```

**What It Keeps:**
- CRDs (TimeWindowScaler definitions)
- Existing TimeWindowScaler instances
- Custom resources in other namespaces

**Success Indicator:**
```
Undeploying controller...
deployment.apps "kyklos-controller-manager" deleted
clusterrolebinding.rbac.authorization.k8s.io "kyklos-controller-rolebinding" deleted
clusterrole.rbac.authorization.k8s.io "kyklos-controller-role" deleted
serviceaccount "kyklos-controller" deleted
namespace "kyklos-system" deleted
✓ Controller removed
```

---

### redeploy

**Intent:** Full rebuild and redeploy cycle for rapid iteration

**Duration:** 2-3 minutes

**Dependencies:** All build and deploy dependencies

**Usage:**
```bash
make redeploy
```

**What It Does (in order):**
1. `make manifests` - Regenerate CRDs
2. `make generate` - Regenerate code
3. `make build` - Compile controller
4. `make docker-build` - Build image
5. `make kind-load` or `make k3d-load` - Load image
6. `make deploy-rollout` - Rolling restart

**Success Indicator:**
```
Full redeploy cycle...
→ Manifests generated
→ Code generated
→ Binary built
→ Image built
→ Image loaded
→ Controller restarted
✓ Redeploy complete (elapsed: 2m 15s)
```

**Use Case:** After making code changes during development

---

### deploy-rollout

**Intent:** Restart controller pods to pick up new image

**Duration:** 10-15 seconds

**Dependencies:** Controller deployed

**Usage:**
```bash
make deploy-rollout
```

**What It Does:**
```bash
kubectl rollout restart deployment -n kyklos-system kyklos-controller-manager
kubectl rollout status deployment -n kyklos-system kyklos-controller-manager
```

**Success Indicator:**
```
Restarting controller...
deployment.apps/kyklos-controller-manager restarted
Waiting for rollout to finish...
✓ Deployment successfully rolled out
```

---

### kind-load

**Intent:** Load Docker image into Kind cluster nodes

**Duration:** 20-30 seconds

**Dependencies:** Kind cluster running, image built

**Usage:**
```bash
make kind-load

# Load specific image
make kind-load IMG=kyklos/controller:v0.1.0
```

**What It Does:**
```bash
kind load docker-image kyklos/controller:dev --name kyklos-dev
```

**Success Indicator:**
```
Loading image into Kind cluster...
Image: "kyklos/controller:dev" with ID "sha256:abc123..." not yet present on node "kyklos-dev-control-plane", loading...
✓ Image loaded
```

**Verification:**
```bash
docker exec -it kyklos-dev-control-plane crictl images | grep kyklos
```

---

### k3d-load

**Intent:** Load Docker image into k3d cluster

**Duration:** 15-25 seconds

**Dependencies:** k3d cluster running, image built

**Usage:**
```bash
make k3d-load
```

**What It Does:**
```bash
k3d image import kyklos/controller:dev -c kyklos-dev
```

---

## Testing Targets

### test

**Intent:** Run unit tests for all packages

**Duration:** 20-30 seconds

**Dependencies:** None (pure Go tests)

**Usage:**
```bash
make test
```

**What It Runs:**
```bash
go test ./api/... ./controllers/... ./internal/... -v
```

**Success Indicator:**
```
Running unit tests...
ok   github.com/your-org/kyklos/api/v1alpha1   0.523s
ok   github.com/your-org/kyklos/controllers    2.847s
ok   github.com/your-org/kyklos/internal/webhook   0.412s
✓ All tests passed
```

**Coverage:** Tests in `*_test.go` files colocated with source

---

### test-coverage

**Intent:** Run tests and generate coverage report

**Duration:** 25-35 seconds

**Dependencies:** None

**Usage:**
```bash
make test-coverage
```

**What It Generates:**
- `coverage.out` - Coverage profile
- `coverage.html` - HTML coverage report

**Opens browser with coverage report automatically**

**Success Indicator:**
```
Running tests with coverage...
→ Writing coverage profile: coverage.out
→ Generating HTML report: coverage.html
→ Opening coverage.html in browser
✓ Coverage: 84.3%
```

**View Coverage:**
```bash
go tool cover -html=coverage.out
```

---

### test-integration

**Intent:** Run integration tests using envtest

**Duration:** 60-90 seconds

**Dependencies:** `setup-envtest` tool installed

**Usage:**
```bash
make test-integration
```

**What It Does:**
- Downloads kubebuilder test binaries (etcd, kube-apiserver)
- Starts local control plane
- Runs tests in `controllers/*_test.go` with `// +build integration` tag
- Tears down test environment

**Success Indicator:**
```
Running integration tests...
→ Setting up envtest
→ Starting test control plane
Running Suite: Controller Suite
✓ All integration tests passed (45 tests, 0 failures)
```

---

### test-e2e

**Intent:** Run end-to-end tests against real cluster

**Duration:** 5-10 minutes

**Dependencies:** Cluster running, controller deployed

**Usage:**
```bash
make test-e2e
```

**What It Tests:**
- Full controller lifecycle
- Time window transitions
- Grace period behavior
- Manual drift correction
- Holiday handling

**Location:** `test/e2e/timewindowscaler_test.go`

**Success Indicator:**
```
Running e2e tests...
→ Verifying cluster and controller
→ Running test scenarios
✓ 15 test scenarios passed
✓ E2E tests complete
```

---

### test-all

**Intent:** Run all test suites (unit, integration, e2e)

**Duration:** 6-12 minutes

**Dependencies:** All test dependencies

**Usage:**
```bash
make test-all
```

**What It Runs:**
1. `make test` - Unit tests
2. `make test-integration` - Envtest integration tests
3. `make test-e2e` - Full cluster e2e tests

---

### lint

**Intent:** Run golangci-lint on all Go code

**Duration:** 15-30 seconds

**Dependencies:** `tools` (requires golangci-lint)

**Usage:**
```bash
make lint
```

**What It Checks:**
- Code style and formatting
- Common bugs and pitfalls
- Security issues
- Performance issues
- Unused code

**Configuration:** `.golangci.yml`

**Success Indicator:**
```
Running linter...
✓ No issues found
```

**With Issues:**
```
controllers/timewindowscaler_controller.go:45:2: error: undefined: foo
✗ Found 1 issue
```

---

### lint-fix

**Intent:** Auto-fix linting issues where possible

**Duration:** 20-40 seconds

**Usage:**
```bash
make lint-fix
```

**What It Fixes:**
- Import formatting
- Code style issues
- Unused imports

**Does Not Fix:**
- Logic errors
- Undefined references
- Complex refactorings

---

## Verification and Validation

### verify-all

**Intent:** Run all verification checks to ensure healthy system

**Duration:** 15-20 seconds

**Dependencies:** All components should be running

**Usage:**
```bash
make verify-all
```

**What It Checks:**
1. Tool versions
2. Cluster connectivity
3. CRD installation
4. Controller health
5. RBAC permissions
6. Recent logs for errors

**Success Indicator:**
```
Running all verifications...
✓ Tools: All present and correct versions
✓ Cluster: Reachable and healthy
✓ CRDs: Installed and established
✓ Controller: Running and ready
✓ RBAC: Permissions correctly configured
✓ All verifications passed
```

---

### verify-controller

**Intent:** Check controller pod health and logs

**Duration:** 5-10 seconds

**Dependencies:** Controller deployed

**Usage:**
```bash
make verify-controller
```

**What It Checks:**
- Pod exists and is Running
- No crash loops (restarts < 3)
- Recent logs show successful startup
- Metrics endpoint responding

**Success Indicator:**
```
Verifying controller...
✓ Pod: kyklos-controller-manager-abc123-xyz Running
✓ Ready: 1/1
✓ Restarts: 0
✓ Logs: No errors in last 100 lines
✓ Metrics: Endpoint responding
Controller healthy
```

---

### verify-rbac

**Intent:** Validate RBAC permissions are correctly configured

**Duration:** 5 seconds

**Dependencies:** Controller deployed

**Usage:**
```bash
make verify-rbac
```

**What It Tests:**
```bash
kubectl auth can-i list deployments --as=system:serviceaccount:kyklos-system:kyklos-controller
kubectl auth can-i update deployments --as=system:serviceaccount:kyklos-system:kyklos-controller
kubectl auth can-i create events --as=system:serviceaccount:kyklos-system:kyklos-controller
kubectl auth can-i list timewindowscalers --as=system:serviceaccount:kyklos-system:kyklos-controller
```

**Success Indicator:**
```
Verifying RBAC permissions...
✓ Can list deployments: yes
✓ Can update deployments: yes
✓ Can create events: yes
✓ Can list timewindowscalers: yes
✓ All required permissions present
```

---

### verify-image-loaded

**Intent:** Confirm controller image is available in cluster

**Duration:** 3-5 seconds

**Dependencies:** Image should be loaded

**Usage:**
```bash
make verify-image-loaded
```

**What It Checks:**
```bash
docker exec -it kyklos-dev-control-plane crictl images | grep kyklos/controller
```

**Success Indicator:**
```
Verifying image in cluster...
✓ Image kyklos/controller:dev present on nodes
```

---

## Demo and Examples

### demo-setup

**Intent:** Create demo namespace and target deployment

**Duration:** 10-15 seconds

**Dependencies:** Cluster running

**Usage:**
```bash
make demo-setup
```

**What It Creates:**
- Namespace: `demo`
- Deployment: `demo-app` (nginx) with 1 replica

**Success Indicator:**
```
Setting up demo environment...
namespace/demo created
deployment.apps/demo-app created
✓ Demo namespace ready
✓ Target deployment created with 1 replica
```

---

### demo-apply-minute

**Intent:** Apply TimeWindowScaler with minute-scale windows for quick testing

**Duration:** 5 seconds

**Dependencies:** `demo-setup` completed, CRDs installed

**Usage:**
```bash
make demo-apply-minute
```

**What It Applies:**
```yaml
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: demo-minute-scaler
  namespace: demo
spec:
  targetRef:
    kind: Deployment
    name: demo-app
  timezone: UTC
  defaultReplicas: 1
  windows:
  # Scale to 5 replicas on even-numbered minutes
  - days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
    start: "00:00"
    end: "00:01"
    replicas: 5
  - days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
    start: "00:02"
    end: "00:03"
    replicas: 5
  # Repeats every 2 minutes...
```

**Success Indicator:**
```
Applying minute-scale demo...
timewindowscaler.kyklos.io/demo-minute-scaler created
✓ Demo TWS applied
✓ Watch for scale changes every 2 minutes
```

**Expected Behavior:**
- At :00 seconds: Scales to 5 replicas
- At :01 seconds: Scales to 1 replica
- Repeats every 2 minutes

---

### demo-watch

**Intent:** Watch demo resources to observe scaling

**Duration:** Runs continuously until Ctrl+C

**Dependencies:** `demo-apply-minute` completed

**Usage:**
```bash
make demo-watch
```

**What It Shows:**
```bash
watch -n 2 'kubectl get tws,deploy,pods -n demo'
```

**Output:**
```
Every 2.0s: kubectl get tws,deploy,pods -n demo

NAME                                        WINDOW         REPLICAS   TARGET
timewindowscaler.kyklos.io/demo-minute-scaler   BusinessHours  5          demo-app

NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/demo-app   5/5     5            5           3m

NAME                           READY   STATUS    RESTARTS   AGE
pod/demo-app-abc123-xyz        1/1     Running   0          45s
pod/demo-app-abc123-abc        1/1     Running   0          45s
pod/demo-app-abc123-def        1/1     Running   0          45s
pod/demo-app-abc123-ghi        1/1     Running   0          45s
pod/demo-app-abc123-jkl        1/1     Running   0          45s
```

**Stop Watching:** Press Ctrl+C

---

### demo-verify

**Intent:** Quick snapshot of demo state and recent events

**Duration:** 3-5 seconds

**Dependencies:** Demo running

**Usage:**
```bash
make demo-verify
```

**What It Shows:**
1. Current TWS status
2. Deployment replica count
3. Pod list
4. Recent events

**Success Indicator:**
```
Demo Status:
→ TWS: demo-minute-scaler
→ Current Window: BusinessHours
→ Effective Replicas: 5
→ Target Replicas: 5
→ Alignment: ✓ Matched

Recent Events:
47s  Normal  ScaledUp  Scaled deployment demo-app from 1 to 5 replicas
```

---

### demo-cleanup

**Intent:** Remove all demo resources

**Duration:** 5-10 seconds

**Dependencies:** None

**Usage:**
```bash
make demo-cleanup
```

**What It Deletes:**
```bash
kubectl delete namespace demo
```

**Success Indicator:**
```
Cleaning up demo...
namespace "demo" deleted
✓ Demo resources removed
```

---

## Cleanup Targets

### clean

**Intent:** Remove local build artifacts

**Duration:** 1-2 seconds

**Dependencies:** None

**Usage:**
```bash
make clean
```

**What It Removes:**
- `bin/` - Compiled binaries
- `tools/bin/` - Development tools
- `coverage.out` - Test coverage files

**Does Not Remove:**
- Docker images
- Cluster resources
- Generated manifests

---

### clean-all

**Intent:** Complete cleanup including cluster

**Duration:** 30-60 seconds

**Dependencies:** None

**Usage:**
```bash
make clean-all
```

**What It Removes:**
1. Local build artifacts (`make clean`)
2. Docker images (`docker rmi kyklos/controller:*`)
3. Local cluster (`make cluster-down`)

**Warning:** This is destructive and cannot be undone

---

### cleanup-tws

**Intent:** Delete all TimeWindowScaler instances

**Duration:** 5-10 seconds

**Dependencies:** Cluster running

**Usage:**
```bash
make cleanup-tws
```

**What It Deletes:**
```bash
kubectl delete tws --all --all-namespaces
```

**Warning:** Removes all TWS instances across all namespaces

---

## Utility Targets

### logs-controller

**Intent:** View controller logs

**Duration:** Instant

**Dependencies:** Controller deployed

**Usage:**
```bash
make logs-controller

# Last 100 lines
make logs-controller LINES=100

# Tail logs
make logs-controller-follow
```

**What It Shows:**
```bash
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=50
```

**Environment Variables:**
- `LINES` - Number of lines to show (default: 50)

---

### logs-controller-follow

**Intent:** Follow controller logs in real-time

**Duration:** Runs continuously

**Usage:**
```bash
make logs-controller-follow
```

**Stop Following:** Press Ctrl+C

---

### port-forward-metrics

**Intent:** Port-forward to controller metrics endpoint

**Duration:** Runs until stopped

**Dependencies:** Controller running

**Usage:**
```bash
make port-forward-metrics
```

**Access:**
```bash
# In another terminal
curl http://localhost:8080/metrics | grep kyklos
```

**Exposed Endpoints:**
- `localhost:8080/metrics` - Prometheus metrics
- `localhost:8081/healthz` - Health check

---

### restart-controller

**Intent:** Quick restart of controller pod

**Duration:** 10-15 seconds

**Dependencies:** Controller deployed

**Usage:**
```bash
make restart-controller
```

**Equivalent to:** `make deploy-rollout`

---

### reset-env

**Intent:** Complete environment reset (cluster + build)

**Duration:** 3-4 minutes

**Dependencies:** None

**Usage:**
```bash
make reset-env
```

**What It Does:**
1. `make cluster-down`
2. `make clean`
3. `make cluster-up`
4. `make tools`
5. Full rebuild and deploy

**Use Case:** Starting fresh after significant changes or issues

---

### help

**Intent:** Show all Make targets with descriptions

**Duration:** Instant

**Usage:**
```bash
make help
# or just
make
```

**Output:**
```
Kyklos Time Window Scaler - Makefile Targets

Setup:
  verify-tools          Verify all prerequisites installed
  tools                 Install development tools (controller-gen, linters)

Build:
  manifests             Generate CRDs and RBAC from Go types
  generate              Generate deepcopy code
  build                 Compile controller binary
  docker-build          Build container image

...
```

---

## Target Dependency Graph

### Visual Dependency Flow

```
verify-tools
    └─→ tools
        ├─→ manifests
        │   └─→ build
        │       └─→ docker-build
        │           └─→ kind-load
        │               └─→ deploy
        └─→ generate
            └─→ build
```

### Common Workflows

**Fresh Setup:**
```
verify-tools → tools → cluster-up → build → docker-build → kind-load → install-crds → deploy
```

**Code Change:**
```
build → docker-build → kind-load → deploy-rollout
```

**API Change:**
```
manifests → generate → build → docker-build → kind-load → uninstall-crds → install-crds → deploy-rollout
```

**Quick Test:**
```
test → lint
```

**Full Validation:**
```
test-all → lint → verify-all
```

### Parallel-Safe Targets

These can run concurrently:
- `test` and `lint`
- `verify-tools` and `cluster-status`
- `logs-controller` and `demo-watch`

### Sequential-Only Targets

These must run in order:
1. `install-crds` before `deploy`
2. `docker-build` before `kind-load`
3. `cluster-up` before any kubectl operations
4. `tools` before `manifests` or `generate`

---

## Environment Variables

### Commonly Used

| Variable | Default | Purpose |
|----------|---------|---------|
| `IMG` | `kyklos/controller:dev` | Container image name |
| `CLUSTER_NAME` | `kyklos-dev` | Local cluster name |
| `NAMESPACE` | `kyklos-system` | Controller namespace |
| `TIMEOUT` | `2m` | kubectl wait timeout |

### Override Examples

```bash
# Use custom image
make docker-build IMG=myregistry/kyklos:test

# Deploy to custom namespace
make deploy NAMESPACE=testing

# Longer timeout for slow clusters
make deploy TIMEOUT=5m

# Use different cluster name
make cluster-up CLUSTER_NAME=kyklos-test
```

---

## Tips and Best Practices

### Efficient Workflow

**For Controller Logic Changes:**
```bash
make build docker-build kind-load deploy-rollout
# Or use shortcut
make redeploy
```

**For API Changes:**
```bash
make manifests generate uninstall-crds install-crds redeploy
```

**For Quick Testing:**
```bash
# Test locally without cluster
make test test-coverage

# Then integration test
make test-integration
```

### Avoiding Common Mistakes

**Always verify after deploy:**
```bash
make deploy && make verify-controller
```

**Check image loaded before deploy:**
```bash
make kind-load && make verify-image-loaded && make deploy
```

**Clean state for troubleshooting:**
```bash
make clean-all && make verify-tools
# Then start fresh
```

### Makefile Customization

Add custom targets to local `Makefile.local` (gitignored):

```makefile
# Makefile.local
.PHONY: my-workflow
my-workflow:
    @echo "Running custom workflow"
    make test
    make build
    make deploy

-include Makefile.local
```

---

## Reference: All Targets Alphabetical

| Target | Phase | Duration | Description |
|--------|-------|----------|-------------|
| `build` | Build | 30-60s | Compile controller binary |
| `build-fast` | Build | 15-30s | Fast build skipping checks |
| `build-linux` | Build | 30-60s | Cross-compile for Linux |
| `build-race` | Build | 45-75s | Build with race detector |
| `clean` | Cleanup | 1-2s | Remove build artifacts |
| `clean-all` | Cleanup | 30-60s | Complete cleanup |
| `cleanup-tws` | Cleanup | 5-10s | Delete all TWS instances |
| `cluster-down` | Cluster | 10-20s | Delete local cluster |
| `cluster-status` | Utility | 2-5s | Show cluster state |
| `cluster-up` | Cluster | 60-90s | Create Kind cluster |
| `cluster-up-k3d` | Cluster | 30-60s | Create k3d cluster |
| `demo-apply-minute` | Demo | 5s | Apply minute-scale demo |
| `demo-cleanup` | Demo | 5-10s | Remove demo resources |
| `demo-setup` | Demo | 10-15s | Create demo environment |
| `demo-verify` | Demo | 3-5s | Show demo status |
| `demo-watch` | Demo | continuous | Watch demo resources |
| `deploy` | Deploy | 20-30s | Deploy controller |
| `deploy-debug` | Deploy | 20-30s | Deploy with debug logging |
| `deploy-rollout` | Deploy | 10-15s | Restart controller pods |
| `docker-build` | Build | 60-120s | Build container image |
| `docker-build-nocache` | Build | 120-180s | Build without cache |
| `generate` | Build | 5-10s | Generate deepcopy code |
| `help` | Utility | instant | Show all targets |
| `install-crds` | Deploy | 5-10s | Install CRDs |
| `k3d-load` | Deploy | 15-25s | Load image to k3d |
| `kind-load` | Deploy | 20-30s | Load image to Kind |
| `lint` | Test | 15-30s | Run linter |
| `lint-fix` | Test | 20-40s | Auto-fix lint issues |
| `logs-controller` | Utility | instant | View controller logs |
| `logs-controller-follow` | Utility | continuous | Follow controller logs |
| `manifests` | Build | 10-20s | Generate CRDs and RBAC |
| `port-forward-metrics` | Utility | continuous | Forward metrics port |
| `redeploy` | Workflow | 2-3m | Full rebuild and deploy |
| `reset-env` | Cleanup | 3-4m | Reset environment |
| `restart-controller` | Utility | 10-15s | Restart controller |
| `test` | Test | 20-30s | Run unit tests |
| `test-all` | Test | 6-12m | Run all tests |
| `test-coverage` | Test | 25-35s | Test with coverage |
| `test-e2e` | Test | 5-10m | End-to-end tests |
| `test-integration` | Test | 60-90s | Integration tests |
| `tools` | Setup | 60-120s | Install dev tools |
| `undeploy` | Cleanup | 10-15s | Remove controller |
| `uninstall-crds` | Cleanup | 5-10s | Remove CRDs |
| `verify-all` | Verify | 15-20s | All verifications |
| `verify-controller` | Verify | 5-10s | Check controller health |
| `verify-image-loaded` | Verify | 3-5s | Check image in cluster |
| `verify-rbac` | Verify | 5s | Validate RBAC |
| `verify-tools` | Setup | 5s | Check prerequisites |

---

For usage examples and complete workflows, see [LOCAL-DEV-GUIDE.md](./LOCAL-DEV-GUIDE.md).
