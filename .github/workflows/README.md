# GitHub Actions Workflows

## Overview

This directory contains CI/CD workflows for the Kyklos project.

## Workflows

### 1. `test.yml` - Unit and Controller Tests
**Triggers:** Push, Pull Request
**Duration:** ~30 seconds

Runs:
- `go mod tidy` - Clean dependencies
- `make test` - Unit tests (engine + controller)

**Important:** Cleans `bin/` directory to prevent architecture mismatch errors.

---

### 2. `test-e2e.yml` - End-to-End Tests
**Triggers:** Push, Pull Request
**Duration:** ~5 minutes

Runs:
- Creates Kind cluster
- Deploys Kyklos
- Runs E2E tests

**Important:** Cleans `bin/` directory to prevent architecture mismatch errors.

---

### 3. `lint.yml` - Code Quality
**Triggers:** Push, Pull Request
**Duration:** ~1 minute

Runs:
- golangci-lint via GitHub Action
- Uses version v2.1.6

---

### 4. `ci.yml` - Comprehensive CI
**Triggers:** Push to main, Pull Request to main
**Duration:** ~2 minutes

Jobs:
- **lint**: Code quality checks
- **test-unit**: Unit tests
- **verify**: Documentation checks

**Important:** Cleans `bin/` directory in all jobs to prevent architecture mismatch errors.

---

## Common Issues

### Architecture Mismatch
**Problem:** `cannot execute binary file: Exec format error`

**Cause:** Binaries in `bin/` directory were built for a different architecture (e.g., macOS arm64) but CI runs on Linux amd64.

**Solution:** All workflows now clean the `bin/` directory before running tests:
```yaml
- name: Clean local binaries
  run: rm -rf bin/
```

This ensures binaries are rebuilt for the correct architecture in CI.

---

### Go Version
All workflows use `go-version-file: go.mod` to automatically use the Go version specified in `go.mod`.

Current version: **Go 1.24.0**

---

## Local Testing

To test workflows locally, you can run the same commands:

```bash
# Clean binaries (important!)
rm -rf bin/

# Run what the workflows run
go mod tidy
make test           # For test.yml
make test-e2e       # For test-e2e.yml
make lint           # For ci.yml
```

---

## Workflow Best Practices

1. **Always clean bin/ directory** in CI to prevent architecture issues
2. **Use go-version-file** instead of hardcoded Go version
3. **Run go mod tidy** before tests to ensure clean dependencies
4. **Use concurrency groups** to cancel outdated runs
5. **Fail fast** - don't continue if setup fails

---

## Adding New Workflows

When adding new workflows that use make targets:

```yaml
steps:
  - name: Clone the code
    uses: actions/checkout@v4

  - name: Setup Go
    uses: actions/setup-go@v5
    with:
      go-version-file: go.mod

  # IMPORTANT: Clean binaries
  - name: Clean local binaries
    run: rm -rf bin/

  - name: Your workflow step
    run: |
      go mod tidy
      make your-target
```

---

## Debugging Failed Workflows

### Check Architecture
```bash
# In GitHub Actions logs, look for:
bash: line 1: /home/runner/work/kyklos/kyklos/bin/controller-gen: cannot execute binary file
```

**Fix:** Add `rm -rf bin/` step before running make commands.

### Check Go Version
```bash
# Ensure workflow uses:
go-version-file: go.mod

# NOT:
go-version: '1.21'  # Outdated/incorrect
```

### Check Dependencies
```bash
# Ensure workflow runs:
go mod tidy

# Before:
make test
```

---

## Performance Optimization

### Caching
Currently, workflows don't cache Go modules. To add caching:

```yaml
- name: Cache Go modules
  uses: actions/cache@v3
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

### Parallelization
The `ci.yml` workflow runs jobs in parallel:
- lint (1 min)
- test-unit (30s)
- verify (10s)

Total time: ~1 minute (instead of 1m 40s sequential)

---

## Status Badges

Add to README.md:

```markdown
[![Tests](https://github.com/your-org/kyklos/actions/workflows/test.yml/badge.svg)](https://github.com/your-org/kyklos/actions/workflows/test.yml)
[![Lint](https://github.com/your-org/kyklos/actions/workflows/lint.yml/badge.svg)](https://github.com/your-org/kyklos/actions/workflows/lint.yml)
[![E2E](https://github.com/your-org/kyklos/actions/workflows/test-e2e.yml/badge.svg)](https://github.com/your-org/kyklos/actions/workflows/test-e2e.yml)
```
