# Kyklos Time Window Scaler

Kyklos automatically scales your Kubernetes workloads based on time windows and calendars. Define when your application needs more capacity and Kyklos handles the rest.

Unlike cron-based scaling (manual triggers) or HPA (metric-based), Kyklos scales proactively based on predictable time patterns. Perfect for business hours traffic, batch processing windows, or any workload with time-based demand.

## Why Kyklos?

- **Time-aware scaling** - Scale up before peak hours, down after
- **Calendar integration** - Handle holidays and special dates
- **Timezone support** - Full DST handling with IANA timezones
- **Graceful downscaling** - Optional delay when reducing replicas
- **Drift correction** - Automatically reverts manual scaling changes
- **Production-ready** - Status conditions, events, and metrics built-in

## Quick Start (5 minutes)

Get Kyklos running on a fresh Kind cluster in under 5 minutes.

### Prerequisites

- Go 1.21+
- Docker running
- kubectl configured
- Kind or k3d installed

Verify prerequisites:
```bash
make verify-tools
```

### Installation

1. **Create local cluster**
```bash
make cluster-up
```

2. **Build and deploy Kyklos**
```bash
make build docker-build kind-load
make install-crds deploy
```

3. **Verify controller is running**
```bash
kubectl get pods -n kyklos-system
```

Expected output:
```
NAME                                      READY   STATUS    RESTARTS   AGE
kyklos-controller-manager-abc123-xyz      1/1     Running   0          30s
```

4. **Run smoke test** (optional but recommended)
   ```bash
   make test || echo "Tests will be available in implementation phase"
   ```

### Your First TimeWindowScaler

Create a deployment and scale it based on office hours:

```bash
# Create demo namespace and target deployment
make demo-setup
```

Apply this TimeWindowScaler:

```yaml
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: office-hours-scaler
  namespace: demo
spec:
  targetRef:
    kind: Deployment
    name: demo-app
  timezone: America/New_York
  defaultReplicas: 2
  windows:
  # Scale to 10 replicas during business hours
  - days: [Mon, Tue, Wed, Thu, Fri]
    start: "09:00"
    end: "17:00"
    replicas: 10
  # Reduced capacity evenings
  - days: [Mon, Tue, Wed, Thu, Fri]
    start: "17:00"
    end: "22:00"
    replicas: 5
```

Save as `office-hours.yaml` and apply:
```bash
kubectl apply -f office-hours.yaml
```

Watch it work:
```bash
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
- **[Minute Demo](docs/user/MINUTE-DEMO.md)** - Complete walkthrough with examples
- **[Troubleshooting](docs/user/TROUBLESHOOTING.md)** - Fix common issues
- **[API Reference](docs/api/CRD-SPEC.md)** - Complete CRD specification

## Project Status

**Version:** v0.1.0-alpha

This is an alpha release focused on Deployment scaling with time windows. See [ROADMAP.md](docs/ROADMAP.md) for future features.

**Current Features:**
- Time window matching with inclusive start, exclusive end
- IANA timezone support with full DST handling
- Holiday modes (closed, open, ignore)
- Grace periods for downscaling
- Manual drift correction
- Pause functionality for incidents
- Cross-midnight window support

**Limitations (v0.1):**
- Only Deployment targets supported
- Single controller replica (no HA)
- No webhook validation (CRD-level validation only)

## Examples

See the [examples/](examples/) directory:
- `tws-office-hours.yaml` - Business hours scaling pattern
- `tws-night-shift.yaml` - Cross-midnight window example
- `tws-holidays-closed.yaml` - Holiday handling example

## Documentation

- **[User Documentation](docs/user/)** - Getting started, concepts, operations
- **[API Documentation](docs/api/)** - CRD specification and validation
- **[Design Documentation](docs/design/)** - Reconcile logic, status, events
- **[Development Guide](docs/LOCAL-DEV-GUIDE.md)** - Local development workflow
- **[Testing Strategy](docs/testing/TEST-STRATEGY.md)** - Test plans and coverage

## Contributing

See [docs/BRIEF.md](docs/BRIEF.md) for project organization and [docs/RACI.md](docs/RACI.md) for responsibilities.

## License

Apache 2.0
