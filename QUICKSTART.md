# Kyklos Quick Start Guide

## 🚀 Get Started in 5 Minutes

### Option 1: Quick Test (Fastest)
```bash
# Prerequisites: Docker, kubectl, k3d installed
make verify-tools      # Verify prerequisites
make k3d-test          # Create cluster, deploy, test (3 minutes)
```

### Option 2: Manual Setup
```bash
# 1. Install dependencies
make install-tools

# 2. Create k3d cluster
make k3d-create

# 3. Build and deploy
make k3d-deploy IMG=kyklos:local

# 4. Run smoke test
make smoke-test
```

---

## 📋 Common Commands

### Development
```bash
make tidy              # Clean up go.mod dependencies
make build             # Build manager binary
make test              # Run all tests (unit + controller)
make test-engine       # Run engine tests only (fastest)
make lint              # Run golangci-lint
```

### Testing
```bash
make smoke-test        # 30-second quick validation
make sanity-test       # 3-minute comprehensive test
make test-controller   # Controller integration tests
make test-e2e          # End-to-end tests (requires Kind)
```

### K3d Cluster
```bash
make k3d-create        # Create test cluster
make k3d-deploy        # Build & deploy to k3d
make k3d-test          # Full test pipeline
make k3d-demo          # Deploy & run demo
make k3d-status        # Show cluster status
make k3d-delete        # Remove cluster
```

### Deployment
```bash
make install           # Install CRDs
make deploy IMG=...    # Deploy controller
make undeploy          # Remove controller
make uninstall         # Remove CRDs
```

---

## 🎯 Common Workflows

### Quick Validation
```bash
make verify-tools && make k3d-test
```

### Development Loop
```bash
# Make changes to code
make test-engine       # Fast unit tests
make build             # Build binary
make docker-build      # Build image
make k3d-deploy        # Deploy to test cluster
make smoke-test        # Quick validation
```

### Demo for Stakeholders
```bash
make k3d-demo          # Creates cluster, deploys, runs live demo
```

### Before Committing
```bash
make tidy              # Clean dependencies
make test              # Run all tests
make lint              # Check code quality
```

---

## 🛠️ Tool Installation

### One Command
```bash
make install-tools     # Installs: kustomize, controller-gen, envtest, golangci-lint
```

### Individual Tools
```bash
make kustomize         # Install kustomize
make controller-gen    # Install controller-gen
make envtest           # Install envtest
make golangci-lint     # Install golangci-lint
```

### Verify Installation
```bash
make verify-tools      # Check all prerequisites
```

---

## 📊 Test Coverage

| Test Type | Duration | Coverage | Command |
|-----------|----------|----------|---------|
| Unit (engine) | 1s | 88.4% | `make test-engine` |
| Controller tests | 20s | 56.4% | `make test-controller` |
| Smoke test | 30s | Basic | `make smoke-test` |
| Sanity test | 3min | Comprehensive | `make sanity-test` |
| E2E tests | 5min | Full stack | `make test-e2e` |

---

## 🐛 Troubleshooting

### Controller not starting
```bash
make k3d-status                    # Check status
kubectl logs -n kyklos-system deployment/kyklos-controller-manager
make k3d-delete && make k3d-test  # Fresh start
```

### Tests failing
```bash
make verify-tools                  # Check prerequisites
make install-tools                 # Reinstall tools
make tidy                          # Clean dependencies
```

### Image not loading
```bash
make docker-build IMG=kyklos:local # Rebuild image
k3d image import kyklos:local -c kyklos-test  # Re-import
```

### CRD issues
```bash
make uninstall                     # Remove old CRDs
make install                       # Install fresh CRDs
```

---

## 📚 More Information

- **Full Documentation**: [docs/](docs/)
- **API Reference**: [docs/api/CRD-SPEC.md](docs/api/CRD-SPEC.md)
- **Examples**: [examples/](examples/)
- **Testing Guide**: [test/sanity/README.md](test/sanity/README.md)
- **Contributing**: [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 🎓 Learning Path

### Beginner
1. Run `make verify-tools`
2. Run `make k3d-test`
3. Check [examples/](examples/)
4. Read [docs/user/CONCEPTS.md](docs/user/CONCEPTS.md)

### Intermediate
1. Create custom TimeWindowScaler
2. Run `make sanity-test` and watch behavior
3. Read [docs/design/](docs/design/)
4. Explore [test/sanity/](test/sanity/)

### Advanced
1. Read controller code
2. Run `make test-controller`
3. Add new features
4. Read [docs/dev/](docs/dev/)

---

## 💡 Pro Tips

```bash
# Watch resources in real-time
watch -n 1 'kubectl get tws,deploy,pods -A'

# Follow controller logs
kubectl logs -n kyklos-system deployment/kyklos-controller-manager -f

# Quick cluster reset
make k3d-delete && make k3d-test

# Build without cache
docker build --no-cache -t kyklos:local .

# Check all resources
kubectl get all,tws -A
```

---

## 🆘 Getting Help

```bash
make help              # Show all make targets
make verify-tools      # Check your environment
make k3d-status        # Show current status
```

For issues: [GitHub Issues](https://github.com/your-org/kyklos/issues)
