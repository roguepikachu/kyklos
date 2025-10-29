# Kyklos Time Window Scaler

Kyklos is a Kubernetes operator that automatically scales Deployments based on time windows, enabling you to align resource consumption with actual usage patterns—scaling up during business hours and down during off-peak times.

Unlike cron-based scaling (manual triggers) or HPA (metric-based), Kyklos scales proactively based on predictable time patterns. Perfect for business hours traffic, batch processing windows, or any workload with time-based demand.

## Why Kyklos?

- **Time-aware scaling** - Scale up before peak hours, down after
- **Calendar integration** - Handle holidays and special dates automatically
- **Timezone support** - Full DST handling with IANA timezones
- **Graceful downscaling** - Optional delay when reducing replicas
- **Drift correction** - Automatically reverts manual scaling changes
- **Declarative configuration** - Single resource instead of multiple cron jobs
- **Overlapping windows** - Handle complex schedules with automatic precedence
- **Production-ready** - Status conditions, events, and metrics built-in

## Quick Start (5 minutes)

Get Kyklos running on a fresh Kind cluster in under 5 minutes.

### Prerequisites

- Go 1.23.0+
- Docker 17.03+
- kubectl 1.25.0+
- Kind or k3d installed
- Access to a Kubernetes 1.25.0+ cluster

Verify prerequisites:
```bash
make verify-tools
```

### Installation

1. **Create local cluster** (if using Kind)
```bash
make cluster-up
```

2. **Build and push your image**
```bash
make docker-build docker-push IMG=<some-registry>/kyklos:tag
```

Or for local development with Kind:
```bash
make build docker-build kind-load
```

3. **Install CRDs and deploy controller**
```bash
make install
make deploy IMG=<some-registry>/kyklos:tag
```

Or for local development:
```bash
make install-crds deploy
```

4. **Verify controller is running**
```bash
kubectl get pods -n kyklos-system
```

Expected output:
```
NAME                                      READY   STATUS    RESTARTS   AGE
kyklos-controller-manager-abc123-xyz      1/1     Running   0          30s
```

### Your First TimeWindowScaler

Create a deployment and scale it based on office hours:

```bash
# Create demo namespace and target deployment (optional)
make demo-setup
```

Apply this TimeWindowScaler:

```yaml
apiVersion: kyklos.kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: business-hours-scaler
spec:
  targetRef:
    kind: Deployment
    name: my-app
  timezone: "America/New_York"
  defaultReplicas: 2
  windows:
  # Scale to 10 replicas during business hours
  - name: BusinessHours
    start: "09:00"
    end: "17:00"
    replicas: 10
    days: ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
  # Scale to 20 replicas during lunch peak
  - name: LunchPeak
    start: "12:00"
    end: "14:00"
    replicas: 20
    days: ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
```

Save as `office-hours.yaml` and apply:
```bash
kubectl apply -f office-hours.yaml
```

Watch it work:
```bash
kubectl get tws -w
# Or with more details
kubectl get tws,deploy,pods -n demo
```

For a fast 10-minute demo with minute-scale windows:
```bash
make demo-apply-minute
make demo-watch
```

### Next Steps

- **[Concepts](docs/user/CONCEPTS.md)** - Understand windows, boundaries, and holidays
- **[Operations Guide](docs/user/OPERATIONS.md)** - Run Kyklos in production
- **[FAQ](docs/user/FAQ.md)** - Common questions answered
- **[Troubleshooting](docs/user/TROUBLESHOOTING.md)** - Fix common issues
- **[API Reference](docs/api/CRD-SPEC.md)** - Complete CRD specification

## Features

- **Time-based scaling**: Define windows with start/end times and replica counts
- **Timezone aware**: Handles any IANA timezone with automatic DST adjustments
- **Holiday support**: Scale differently on holidays via ConfigMap (closed, open, ignore modes)
- **Overlapping windows**: Last window wins for precedence control
- **Grace periods**: Delayed scale-down to avoid flapping
- **Pause mode**: Temporarily disable scaling while keeping configuration
- **Cross-midnight windows**: Seamlessly handle windows that span days
- **Manual drift correction**: Automatically reverts manual scaling changes
- **Status tracking**: Comprehensive conditions and events for observability

## Examples

See the [examples/](examples/) directory:
- `tws-office-hours.yaml` - Business hours scaling pattern
- `tws-night-shift.yaml` - Cross-midnight window example
- `tws-holidays-closed.yaml` - Holiday handling example

Example configurations:

```yaml
# Cross-midnight window
apiVersion: kyklos.kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: night-shift-scaler
spec:
  targetRef:
    kind: Deployment
    name: batch-processor
  timezone: "UTC"
  defaultReplicas: 2
  windows:
  - name: NightShift
    start: "22:00"
    end: "06:00"
    replicas: 20
    days: ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"]
```

## Documentation

### User Documentation
- **[Concepts](docs/user/CONCEPTS.md)** - Core concepts and how they work
- **[Operations Guide](docs/user/OPERATIONS.md)** - Running in production
- **[FAQ](docs/user/FAQ.md)** - Common questions answered
- **[Troubleshooting](docs/user/TROUBLESHOOTING.md)** - Fixing common issues
- **[Minute Demo](docs/user/MINUTE-DEMO.md)** - Complete walkthrough with examples

### API Documentation
- **[API Reference](docs/api/CRD-SPEC.md)** - Complete API specification

### Design Documentation
- **[Design Docs](docs/design/)** - Reconcile logic, status, events

### Development Documentation
- **[Local Development Guide](docs/dev/LOCAL-DEV-GUIDE.md)** - Local development workflow
- **[Testing Strategy](docs/testing/TEST-STRATEGY.md)** - Test plans and coverage

## Project Status

**Version**: v0.1.0-alpha

**Supported Kubernetes**: 1.25-1.31

This is an alpha release focused on Deployment scaling with time windows. See [ROADMAP.md](docs/ROADMAP.md) for future features.

**Current Implementation Status**:
- ✅ Core time window engine (83.8% test coverage)
- ✅ Basic controller with scaling logic
- ✅ Status updates and events
- ✅ Pause mode
- ✅ Holiday modes (via ConfigMap)
- ✅ Cross-midnight windows
- ✅ DST handling
- ✅ Manual drift correction
- 🚧 Grace periods (structure in place, timing logic pending)
- 🚧 Metrics and observability
- 🚧 Admission webhooks

**Current Features**:
- Time window matching with inclusive start, exclusive end
- IANA timezone support with full DST handling
- Holiday modes (closed, open, ignore)
- Grace periods for downscaling
- Manual drift correction
- Pause functionality for incidents
- Cross-midnight window support

**Limitations (v0.1)**:
- Only Deployment targets supported
- Single controller replica (no HA)
- No webhook validation (CRD-level validation only)

## Development

### Running Tests

```bash
# Run pure engine tests (83.8% coverage)
make test-engine

# Run all tests
make test

# Run controller tests with envtest
make test-controller
```

### Local Development

See [Local Development Guide](docs/dev/LOCAL-DEV-GUIDE.md) for detailed setup instructions.

Quick start for development:
```bash
# Verify tools
make verify-tools

# Create local cluster
make cluster-up

# Build and deploy locally
make build docker-build kind-load
make install-crds deploy

# Run tests
make test
```

### Deployment Management

**To Deploy on the Cluster**:

```sh
# Build and push your image
make docker-build docker-push IMG=<some-registry>/kyklos:tag

# Install the CRDs
make install

# Deploy the controller
make deploy IMG=<some-registry>/kyklos:tag

# Apply sample configurations
kubectl apply -f examples/tws-office-hours.yaml
```

**To Uninstall**:

```sh
# Delete the instances (CRs)
kubectl delete -f examples/

# Delete the CRDs
make uninstall

# Remove the controller
make undeploy
```

## Repository Structure

```
kyklos/
├── api/v1alpha1/         # API types and CRD definitions
├── internal/
│   ├── engine/           # Pure time calculation logic (no K8s dependencies)
│   └── controller/       # Kubernetes controller implementation
├── config/               # Kustomize manifests
├── docs/                 # Comprehensive documentation
│   ├── api/             # API specifications
│   ├── design/          # Design documents
│   ├── testing/         # Test plans and strategies
│   ├── user/            # User documentation
│   ├── dev/             # Development guides
│   └── implementation/  # Implementation guides
├── examples/            # Sample TimeWindowScaler resources
└── test/               # Test fixtures and e2e tests
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/implementation/CONTRIBUTING-IMPL.md) for details.

See [docs/BRIEF.md](docs/BRIEF.md) for project organization and [docs/RACI.md](docs/RACI.md) for responsibilities.

This project is in active development. Key areas where we need help:
- Grace period timing implementation
- Metrics and Prometheus integration
- Additional timezone testing
- Documentation improvements
- Example configurations for common use cases

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025 roguepikachu.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
