# Kyklos Sanity Tests - Quick Reference

## TL;DR - Run This Now! ⚡

```bash
# 30-second smoke test (fastest)
make smoke-test

# 3-minute sanity test (comprehensive)
make sanity-test
```

## What Gets Tested

### Smoke Test (30s)
✓ Controller is running
✓ CRDs are installed
✓ Basic scaling works
✓ Status updates correctly
✓ Events are emitted

### Sanity Test (3min)
✓ All of above, plus:
✓ Time-based window transitions
✓ Multiple window support
✓ Default replica handling
✓ Window boundary calculations

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | ✅ Test passed |
| 1 | ❌ Test failed |

## Prerequisites

```bash
# Verify Kyklos is running
kubectl get pods -n kyklos-system

# Should show:
# kyklos-controller-manager-xxxxx   1/1   Running
```

## Common Issues

### "Controller not found"
```bash
# Deploy Kyklos first
make deploy IMG=kyklos:local
```

### "CRD not found"
```bash
# Install CRDs
make install
```

### "Permission denied"
```bash
# Make scripts executable
chmod +x test/sanity/*.sh
```

## Full Test Suite Comparison

| Test | Duration | Use Case | Auto-Cleanup |
|------|----------|----------|--------------|
| **Unit Tests** | ~1s | Development | N/A |
| **Engine Tests** | ~1s | Engine logic | N/A |
| **Controller Tests** | ~20s | Controller logic | N/A |
| **Smoke Test** | ~30s | Quick validation | ✅ |
| **Sanity Test** | ~3min | Full validation | Optional |
| **E2E Tests** | ~5min | End-to-end | ✅ |

## CI/CD Integration

### Minimal (Fast Feedback)
```yaml
test:
  script:
    - make test-engine    # 1s
    - make smoke-test     # 30s
```

### Standard (Good Coverage)
```yaml
test:
  script:
    - make test           # 20s
    - make smoke-test     # 30s
```

### Comprehensive (Full Validation)
```yaml
test:
  script:
    - make test           # 20s
    - make smoke-test     # 30s
    - make sanity-test    # 3min
```

## Watch Mode

Want to see scaling in real-time?

```bash
# Terminal 1: Run sanity test
make sanity-test

# Terminal 2: Watch deployment
watch -n 1 'kubectl get deployment -n kyklos-sanity'

# Terminal 3: Watch TWS
watch -n 1 'kubectl get tws -n kyklos-sanity -o wide'
```

## Manual Testing

For debugging or demonstration:

```bash
# Keep resources after test
./test/sanity/run-sanity-test.sh
# Choose 'N' when asked to delete

# Then inspect:
kubectl get all -n kyklos-sanity
kubectl describe tws -n kyklos-sanity
kubectl get events -n kyklos-sanity

# Cleanup when done:
kubectl delete namespace kyklos-sanity
```

## Troubleshooting Commands

```bash
# Check controller status
kubectl get pods -n kyklos-system
kubectl logs -n kyklos-system deployment/kyklos-controller-manager

# Check test resources (if not auto-cleaned)
kubectl get all -n kyklos-smoke
kubectl get all -n kyklos-sanity

# Check TWS status
kubectl get tws -A
kubectl describe tws -n kyklos-smoke smoke-scaler

# Force cleanup
kubectl delete namespace kyklos-smoke kyklos-sanity
```

## Success Metrics

After smoke test passes, you should see:
- Deployment scaled to 5 replicas
- TWS status showing "Ready: True"
- At least one "ScaledUp" event
- Controller logs showing "Computed scaling decision"

After sanity test passes, you should see:
- Multiple scaling operations (1→3→5→1)
- TWS current window updating
- Multiple window transitions in events
- No error conditions in status
