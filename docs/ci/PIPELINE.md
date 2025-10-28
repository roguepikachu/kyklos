# Kyklos CI/CD Pipeline Design

## Executive Summary

The Kyklos CI/CD pipeline is designed to provide fast feedback, deterministic builds, and reliable releases. The pipeline architecture prioritizes developer experience with smoke tests completing in under 10 minutes while maintaining comprehensive quality gates.

### Key Design Decisions

- **Parallelization First**: Independent stages run concurrently for maximum speed
- **Intelligent Caching**: Multi-layer cache strategy for dependencies, build artifacts, and test data
- **Fail Fast**: Critical failures block immediately; diagnostic artifacts always preserved
- **Deterministic Builds**: Reproducible outcomes across all runs via controlled environments
- **Security by Default**: Image scanning, SBOM generation, and signature verification integrated

### Pipeline Goals

| Goal | Target | Measurement |
|------|--------|-------------|
| Smoke Test Speed | < 10 minutes | Time from commit to smoke test completion |
| Full Test Suite | < 15 minutes | Time from commit to all tests passing |
| Build Reliability | > 99.5% | Success rate excluding external failures |
| Cache Hit Rate | > 85% | Percentage of builds using cached dependencies |
| Flake Rate | < 0.1% | Test failures that pass on retry |

---

## Pipeline Architecture

### Stage Graph Overview

```
PR and Main Branch Flow:
┌──────────┐
│ Checkout │
└────┬─────┘
     │
     ├─────────────┬──────────────┬──────────────┬──────────────┐
     ▼             ▼              ▼              ▼              ▼
┌─────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│  Lint   │  │   Unit   │  │ Envtest  │  │  Build   │  │  Verify  │
│ (1 min) │  │ (2 min)  │  │ (3 min)  │  │ (3 min)  │  │ (1 min)  │
└────┬────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘
     │            │              │              │              │
     └────────────┴──────────────┴──────────────┴──────────────┘
                                 │
                                 ▼
                          ┌─────────────┐
                          │ Kind Smoke  │
                          │  (5 min)    │
                          └──────┬──────┘
                                 │
                                 ▼
                          ┌─────────────┐
                          │   Report    │
                          │  (30 sec)   │
                          └─────────────┘

Tag Flow (Release):
┌──────────┐
│ Checkout │
└────┬─────┘
     │
     ├─────────────┬──────────────┬──────────────┐
     ▼             ▼              ▼              ▼
┌─────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
│  Lint   │  │   Unit   │  │ Envtest  │  │  Build   │
│         │  │          │  │          │  │  Binary  │
└────┬────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘
     │            │              │              │
     └────────────┴──────────────┴──────────────┘
                                 │
                                 ▼
                          ┌─────────────┐
                          │  Multi-arch │
                          │   Image     │
                          │  (4 min)    │
                          └──────┬──────┘
                                 │
                                 ▼
                          ┌─────────────┐
                          │    Scan     │
                          │  Security   │
                          │  (2 min)    │
                          └──────┬──────┘
                                 │
                                 ▼
                          ┌─────────────┐
                          │   E2E Full  │
                          │  (8 min)    │
                          └──────┬──────┘
                                 │
                                 ▼
                          ┌─────────────┐
                          │   Publish   │
                          │  Artifacts  │
                          │  (2 min)    │
                          └──────┬──────┘
                                 │
                                 ▼
                          ┌─────────────┐
                          │   Release   │
                          │    Notes    │
                          └─────────────┘
```

---

## Job Specifications

### Job 1: Lint

**Purpose**: Code quality validation and formatting checks

**Trigger**: Every commit (PR and main)

**Runner**: `ubuntu-latest` (2 CPU, 7GB RAM)

**Time Budget**: 1 minute

**Responsibilities**:
- Go code formatting check (`gofmt -s`)
- Go linting (`golangci-lint run`)
- YAML validation for CRDs and examples
- Markdown linting for documentation
- License header verification
- Dependency vulnerability scan (go mod check)

**Success Criteria**:
- All linters pass with zero errors
- Code is properly formatted
- All YAML files parse successfully
- Documentation follows style guide

**Failure Actions**:
- Block PR merge
- Post detailed lint errors as PR comment
- Suggest auto-fixes where applicable

**Cache Strategy**:
- Key: `lint-${{ runner.os }}-${{ hashFiles('.golangci.yml') }}`
- Cached: golangci-lint binary and cache directory
- Invalidation: Configuration file changes

---

### Job 2: Unit Tests

**Purpose**: Fast validation of business logic without Kubernetes dependencies

**Trigger**: Every commit (PR and main)

**Runner**: `ubuntu-latest` (4 CPU, 14GB RAM)

**Time Budget**: 2 minutes

**Responsibilities**:
- Run all unit tests with `go test -v -race -coverprofile=coverage.out ./...`
- Generate coverage report
- Verify minimum 80% coverage threshold
- Run tests with race detector enabled
- Test with controlled time using FakeClock

**Success Criteria**:
- All unit tests pass
- Coverage >= 80% overall
- Coverage >= 95% for time calculation and state machine logic
- No data races detected
- Tests complete in under 2 minutes

**Failure Actions**:
- Block PR merge
- Upload coverage report as artifact
- Comment on PR with coverage delta
- Highlight uncovered critical paths

**Cache Strategy**:
- Key: `unit-${{ runner.os }}-go-${{ hashFiles('go.sum') }}`
- Cached: Go module cache, build cache
- Invalidation: Dependency changes

**Artifacts**:
- `coverage.out` - Raw coverage data
- `coverage.html` - HTML coverage report
- `test-results.json` - Structured test results

---

### Job 3: Envtest Tests

**Purpose**: Controller integration tests with embedded API server

**Trigger**: Every commit (PR and main)

**Runner**: `ubuntu-latest` (4 CPU, 14GB RAM)

**Time Budget**: 3 minutes

**Responsibilities**:
- Install envtest binaries (etcd, kube-apiserver)
- Run controller tests against real API server
- Test reconciliation loops
- Validate status updates and conditions
- Test webhook validation (if implemented)
- Verify event emission

**Success Criteria**:
- All envtest scenarios pass
- No API server errors
- Status conditions correctly set
- Events emitted for all state transitions
- Tests complete in under 3 minutes

**Failure Actions**:
- Block PR merge
- Upload envtest logs as artifact
- Capture API server logs on failure
- Preserve test namespace YAML for debugging

**Cache Strategy**:
- Key: `envtest-${{ runner.os }}-${{ hashFiles('go.sum') }}-${{ hashFiles('**/testenv/**') }}`
- Cached: envtest binaries (1.28.x), Go module cache
- Invalidation: Dependency changes or test environment changes

**Artifacts**:
- `envtest-logs.txt` - Test output
- `api-server.log` - API server logs
- `test-namespaces.yaml` - Resource dumps on failure

---

### Job 4: Build Controller Binary

**Purpose**: Compile controller binary for verification

**Trigger**: Every commit (PR and main)

**Runner**: `ubuntu-latest` (2 CPU, 7GB RAM)

**Time Budget**: 3 minutes

**Responsibilities**:
- Build controller binary with `go build`
- Set version from commit SHA or tag
- Build for linux/amd64 and linux/arm64
- Verify binary size (< 100MB)
- Run basic smoke check (--version, --help)

**Success Criteria**:
- Binary compiles successfully for all platforms
- No linker errors
- Version information embedded correctly
- Binary size within expected range

**Failure Actions**:
- Block PR merge
- Report build errors
- Upload partial build artifacts for debugging

**Cache Strategy**:
- Key: `build-${{ runner.os }}-go-${{ hashFiles('go.sum') }}`
- Cached: Go module cache, build cache
- Invalidation: Dependency changes

**Artifacts**:
- `kyklos-controller-linux-amd64` - Controller binary
- `kyklos-controller-linux-arm64` - Controller binary (ARM)
- `build-info.json` - Build metadata

---

### Job 5: Verify Code Quality

**Purpose**: Additional verification checks beyond lint

**Trigger**: Every commit (PR and main)

**Runner**: `ubuntu-latest` (2 CPU, 7GB RAM)

**Time Budget**: 1 minute

**Responsibilities**:
- Check for TODO/FIXME comments with missing issue links
- Verify all exported functions have godoc comments
- Check for hardcoded sensitive values (credentials, tokens)
- Verify CRD examples are up to date with schema
- Check for deprecated API usage

**Success Criteria**:
- No hardcoded secrets detected
- All TODOs have associated issues
- Exported functions documented
- Examples match current CRD schema

**Failure Actions**:
- Warn on PR (non-blocking for minor issues)
- Block on security issues (hardcoded secrets)

**Cache Strategy**:
- Key: `verify-${{ runner.os }}-${{ github.sha }}`
- Cached: Verification tool binaries
- Invalidation: Per-commit

---

### Job 6: Kind Smoke Test

**Purpose**: Fast end-to-end validation in real cluster

**Trigger**: Every commit (PR and main)

**Runner**: `ubuntu-latest` (4 CPU, 14GB RAM)

**Time Budget**: 5 minutes

**Responsibilities**:
- Create kind cluster (single node)
- Build controller image
- Load image into kind
- Install CRDs
- Deploy controller
- Apply minute-scale demo TimeWindowScaler
- Verify scaling behavior over 3 minutes
- Check metrics endpoint
- Verify events emitted

**Success Criteria**:
- Cluster creation succeeds
- Controller pod reaches Ready
- TimeWindowScaler transitions through states
- Deployment scales according to windows
- Metrics available at /metrics endpoint
- All expected events present

**Failure Actions**:
- Block PR merge
- Upload full diagnostic bundle
- Preserve kind cluster state
- Capture all pod logs

**Cache Strategy**:
- Key: `kind-${{ runner.os }}-kind-${{ hashFiles('.kind-config.yaml') }}`
- Cached: Kind binary, node image
- Invalidation: Kind configuration changes

**Artifacts**:
- `kind-logs.tar.gz` - All cluster logs
- `smoke-test-results.json` - Test outcomes
- `cluster-state.yaml` - All resources at failure time
- `metrics-snapshot.txt` - Metrics output

**Smoke Test Scenarios**:
1. **Scale Up**: Window starts, replicas increase
2. **Scale Down**: Window ends, replicas decrease
3. **Pause**: Set pause=true, verify no scaling
4. **Resume**: Set pause=false, verify scaling resumes

---

### Job 7: Multi-arch Image Build (Release Only)

**Purpose**: Build and push production container images

**Trigger**: Tag push matching `v*.*.*`

**Runner**: `ubuntu-latest` (4 CPU, 14GB RAM)

**Time Budget**: 4 minutes

**Responsibilities**:
- Build multi-arch image (amd64, arm64)
- Use BuildKit for layer caching
- Tag with semantic version and commit SHA
- Push to container registry
- Generate image manifest
- Create SBOM (Software Bill of Materials)

**Success Criteria**:
- Images built for both architectures
- Manifest list created
- Images pushed to registry
- SBOM generated and attached
- Image size < 50MB (distroless base)

**Failure Actions**:
- Abort release process
- Alert release engineer
- Preserve build logs

**Cache Strategy**:
- Key: `docker-${{ runner.os }}-${{ hashFiles('**/Dockerfile') }}`
- Cached: Docker layer cache, BuildKit cache
- Invalidation: Dockerfile changes

**Artifacts**:
- `sbom.json` - Software Bill of Materials
- `image-manifest.json` - Multi-arch manifest
- `image-digests.txt` - SHA256 digests

---

### Job 8: Security Scan (Release Only)

**Purpose**: Vulnerability scanning and security validation

**Trigger**: Tag push matching `v*.*.*`

**Runner**: `ubuntu-latest` (2 CPU, 7GB RAM)

**Time Budget**: 2 minutes

**Responsibilities**:
- Scan image with Trivy for CVEs
- Check for HIGH and CRITICAL vulnerabilities
- Verify SBOM completeness
- Validate image signature (if signing enabled)
- Check base image for known vulnerabilities

**Success Criteria**:
- Zero HIGH or CRITICAL CVEs
- SBOM contains all dependencies
- Image signature valid (if applicable)
- Base image is current

**Failure Actions**:
- Block release
- Create security issue
- Notify security team
- Generate vulnerability report

**Cache Strategy**:
- Key: `trivy-${{ runner.os }}-db-${{ github.run_id }}`
- Cached: Trivy vulnerability database
- Invalidation: Daily

**Artifacts**:
- `trivy-report.json` - Vulnerability scan results
- `security-summary.txt` - Human-readable summary

---

### Job 9: E2E Full Test Suite (Release Only)

**Purpose**: Comprehensive end-to-end validation

**Trigger**: Tag push matching `v*.*.*`

**Runner**: `ubuntu-latest` (4 CPU, 14GB RAM)

**Time Budget**: 8 minutes

**Responsibilities**:
- Run all demo scenarios (office hours, night shift, holidays)
- Test DST transitions
- Test cross-midnight windows
- Verify grace period behavior
- Test pause/resume functionality
- Validate metrics accuracy
- Test upgrade path (if applicable)

**Success Criteria**:
- All scenarios pass
- DST transitions handled correctly
- Cross-midnight windows work
- Grace periods respected
- Metrics match expected values

**Failure Actions**:
- Block release
- Upload comprehensive diagnostics
- Preserve cluster for investigation

**Cache Strategy**:
- Key: `e2e-${{ runner.os }}-${{ hashFiles('test/e2e/**') }}`
- Cached: Test framework binaries, cluster images
- Invalidation: E2E test code changes

**Artifacts**:
- `e2e-results.xml` - JUnit format results
- `e2e-logs.tar.gz` - All test logs
- `scenario-traces/` - Per-scenario execution traces

---

### Job 10: Publish Release Artifacts (Release Only)

**Purpose**: Publish release assets and update registries

**Trigger**: Tag push matching `v*.*.*`

**Runner**: `ubuntu-latest` (2 CPU, 7GB RAM)

**Time Budget**: 2 minutes

**Responsibilities**:
- Create GitHub Release
- Upload installation YAML bundle
- Upload SBOM and signatures
- Update Helm chart (if applicable)
- Generate and upload checksums
- Tag images as `latest` (optional)

**Success Criteria**:
- Release created with all assets
- Installation bundle verified
- Checksums match artifacts
- Container images tagged correctly

**Failure Actions**:
- Alert release engineer
- Preserve partial release state
- Document manual completion steps

**Artifacts** (Published):
- `install.yaml` - Complete installation bundle
- `kyklos-v0.1.0.tar.gz` - Source archive
- `sbom.json` - Software Bill of Materials
- `checksums.txt` - SHA256 checksums

---

### Job 11: Generate Release Notes (Release Only)

**Purpose**: Automated changelog and release notes

**Trigger**: Tag push matching `v*.*.*`

**Runner**: `ubuntu-latest` (2 CPU, 7GB RAM)

**Time Budget**: 1 minute

**Responsibilities**:
- Parse conventional commits since last tag
- Group commits by type (feat, fix, chore, etc.)
- Generate structured changelog
- Include breaking changes section
- Link to relevant issues and PRs
- Update GitHub Release description

**Success Criteria**:
- Changelog generated successfully
- All commits categorized
- Breaking changes highlighted
- Links to issues work

**Failure Actions**:
- Continue release (non-critical)
- Generate manual release notes template
- Alert documentation team

**Artifacts**:
- `CHANGELOG-v0.1.0.md` - Version changelog
- `RELEASE-NOTES-v0.1.0.md` - User-facing notes

---

## Concurrency and Parallelization

### Parallel Execution Strategy

**PR and Main Branch**:
```
Parallel Group 1 (Independent):
├── Lint (1 min)
├── Unit Tests (2 min)
├── Envtest (3 min)
├── Build (3 min)
└── Verify (1 min)

Sequential Group 2 (Depends on Group 1):
└── Kind Smoke Test (5 min)
    └── Report (30 sec)

Total Wall Time: ~9 minutes
Total CPU Time: ~15 minutes
```

**Release Flow**:
```
Parallel Group 1:
├── Lint (1 min)
├── Unit Tests (2 min)
├── Envtest (3 min)
└── Build (3 min)

Sequential Group 2:
└── Multi-arch Image Build (4 min)
    ├── Security Scan (2 min)
    └── E2E Full (8 min)
        ├── Publish Artifacts (2 min)
        └── Release Notes (1 min)

Total Wall Time: ~21 minutes
Total CPU Time: ~27 minutes
```

### Concurrency Controls

**PR Builds**:
- Cancel previous runs for same PR on new push
- Maximum 3 concurrent builds per repository
- Queue overflow: fail oldest, keep newest

**Release Builds**:
- Never cancel (must complete or fail)
- Single-threaded (no concurrent releases)
- Require manual approval for major versions

---

## Caching Strategy

### Cache Hierarchy

#### Level 1: Dependency Cache (Highest Hit Rate)
```
Key Pattern: ${{ runner.os }}-go-mod-${{ hashFiles('go.sum') }}
Contents:
  - $GOPATH/pkg/mod/
  - ~/.cache/go-build/
Retention: 7 days
Expected Hit Rate: 95%
Space: ~500 MB
```

**Invalidation**: Go module changes (go.sum modified)

**Optimization**: Separate cache for test dependencies

#### Level 2: Build Artifact Cache
```
Key Pattern: ${{ runner.os }}-build-${{ hashFiles('**/*.go') }}-${{ github.sha }}
Contents:
  - Compiled binaries
  - Intermediate object files
Retention: 3 days
Expected Hit Rate: 60% (for re-runs)
Space: ~100 MB
```

**Invalidation**: Source code changes

**Optimization**: Use commit SHA for exact match

#### Level 3: Tool Binary Cache
```
Key Pattern: ${{ runner.os }}-tools-${{ hashFiles('hack/tools.go') }}
Contents:
  - golangci-lint
  - controller-gen
  - envtest binaries
  - kind binary
Retention: 30 days
Expected Hit Rate: 99%
Space: ~300 MB
```

**Invalidation**: Tool version changes

**Optimization**: Long retention, rarely changes

#### Level 4: Docker Layer Cache
```
Key Pattern: docker-layers-${{ hashFiles('**/Dockerfile') }}
Contents:
  - Docker BuildKit cache
  - Base image layers
Retention: 7 days
Expected Hit Rate: 80%
Space: ~1 GB
```

**Invalidation**: Dockerfile changes or base image updates

**Optimization**: Use registry cache for cross-runner sharing

#### Level 5: Test Data Cache
```
Key Pattern: test-data-${{ hashFiles('test/fixtures/**') }}
Contents:
  - Test fixtures
  - Sample CRD YAML
  - Mock API responses
Retention: 30 days
Expected Hit Rate: 99%
Space: ~10 MB
```

**Invalidation**: Test data changes

**Optimization**: Rarely changes, very stable

### Cache Warming

**Strategy**: Pre-populate caches on schedule

```yaml
# Daily cache warm-up (scheduled workflow)
schedule:
  - cron: '0 2 * * *'  # 2 AM UTC

jobs:
  warm-cache:
    - Download and cache Go modules
    - Build and cache tool binaries
    - Pull and cache base images
    - Generate and cache test fixtures
```

**Benefits**:
- First build of day has warm cache
- Reduces cold start time by 80%
- Predictable build times

### Cache Monitoring

**Metrics to Track**:
- Cache hit rate per job
- Cache restore time
- Cache save time
- Cache size growth

**Alerts**:
- Cache hit rate drops below 70%
- Cache size exceeds 2GB
- Cache restore time > 1 minute

---

## Expected Run Times and Budgets

### Time Budget Allocation

| Stage | Target | Maximum | Timeout |
|-------|--------|---------|---------|
| Checkout | 30s | 1m | 2m |
| Cache Restore | 30s | 2m | 5m |
| Lint | 1m | 2m | 5m |
| Unit Tests | 2m | 3m | 10m |
| Envtest | 3m | 5m | 15m |
| Build Binary | 3m | 5m | 10m |
| Verify | 1m | 2m | 5m |
| Kind Smoke | 5m | 8m | 15m |
| Multi-arch Build | 4m | 6m | 15m |
| Security Scan | 2m | 3m | 10m |
| E2E Full | 8m | 12m | 30m |
| Publish | 2m | 3m | 10m |
| Release Notes | 1m | 2m | 5m |

### Resource Budgets

**Compute**:
- PR builds: ~15 CPU-minutes per run
- Release builds: ~27 CPU-minutes per run
- Monthly estimate: ~500 hours (assuming 100 PR builds, 10 releases)

**Storage**:
- Cache: ~2 GB per runner
- Artifacts: ~500 MB per build (retained 90 days)
- Container images: ~100 MB per release (retained indefinitely)

**Network**:
- Image pulls: ~500 MB per build
- Image pushes: ~100 MB per release
- Artifact uploads: ~50 MB per build

---

## Retry and Flake Handling

### Retry Policy

**Infrastructure Failures** (Always Retry):
- Network timeouts during dependency download
- Registry rate limits
- Transient API server errors
- Runner disk space issues

**Retry Strategy**:
```
Max Attempts: 3
Backoff: Exponential (2s, 4s, 8s)
Jitter: +/- 20%
```

**Test Failures** (No Automatic Retry):
- Unit test failures
- Integration test failures
- E2E test failures

**Rationale**: Automatic retry masks flaky tests. Manual retry after investigation only.

### Flake Detection

**Definition**: A test is flaky if it fails and passes on retry without code changes.

**Detection Method**:
```
1. Capture test failures with environment snapshot
2. Mark build as "suspected flake"
3. Require manual confirmation before retry
4. Track flake rate in dashboard
5. Quarantine tests exceeding flake budget
```

**Flake Budget**:
- Unit tests: 0% tolerance (must be deterministic)
- Envtest: 0.05% tolerance
- E2E tests: 0.5% tolerance
- Overall: 0.1% tolerance

**Quarantine Process**:
1. Test fails 2+ times in 100 runs
2. Automatically mark with `[Flaky]` tag
3. Exclude from required checks
4. Create GitHub issue
5. Fix within 2 sprints or remove

### Failure Categorization

**Category 1: Fast Fail** (Block Immediately)
- Compilation errors
- Lint failures
- Unit test failures
- Coverage drops below threshold

**Category 2: Slow Fail** (Complete Diagnostics First)
- Integration test failures
- E2E test failures
- Smoke test failures

**Category 3: Soft Fail** (Warn but Don't Block)
- Documentation linting
- TODO comment checks
- Non-critical security advisories

### Diagnostic Preservation

**On Any Failure**:
1. Upload all logs as artifacts
2. Capture environment variables (sanitized)
3. Save full git state (commit, branch, PR info)
4. Preserve test fixtures and inputs
5. Generate failure report with repro steps

**Artifact Retention**:
- Successful builds: 7 days
- Failed builds: 90 days
- Release builds: Indefinitely

---

## Quality Gates

### Pre-Merge Gates (Required for PR Approval)

1. **Code Quality**:
   - [ ] All linters pass
   - [ ] Code formatted correctly
   - [ ] No hardcoded secrets

2. **Test Coverage**:
   - [ ] Unit tests pass
   - [ ] Envtest passes
   - [ ] Coverage >= 80%
   - [ ] No new uncovered critical paths

3. **Build Success**:
   - [ ] Binary compiles for all platforms
   - [ ] Smoke test passes in kind cluster

4. **Review**:
   - [ ] At least 1 approving review
   - [ ] All comments resolved
   - [ ] Conventional commit format

### Pre-Release Gates (Required for Tag Push)

1. **All Pre-Merge Gates** (above)

2. **Extended Testing**:
   - [ ] Full E2E test suite passes
   - [ ] All demo scenarios validated
   - [ ] DST transition tests pass

3. **Security**:
   - [ ] Image scan shows zero HIGH/CRITICAL CVEs
   - [ ] SBOM generated and complete
   - [ ] Dependencies updated to latest patches

4. **Documentation**:
   - [ ] Changelog generated
   - [ ] Release notes drafted
   - [ ] Installation docs updated

5. **Artifacts**:
   - [ ] Multi-arch images built and pushed
   - [ ] Installation bundle generated
   - [ ] Checksums verified

---

## Monitoring and Alerting

### Pipeline Health Metrics

**Build Metrics**:
- Build success rate (target: > 99.5%)
- Average build time (target: < 10 min for smoke)
- Cache hit rate (target: > 85%)
- Flake rate (target: < 0.1%)

**Resource Metrics**:
- CPU usage per job
- Memory peak per job
- Disk usage per job
- Network data transfer

**Test Metrics**:
- Test execution time per suite
- Test failure rate by category
- Coverage percentage by component
- Flaky test count

### Alerts

**Critical Alerts** (Page On-Call):
- Build infrastructure down > 5 minutes
- Release build fails
- Security scan blocks release with CRITICAL CVE
- Cache corruption detected

**Warning Alerts** (Slack Notification):
- Build time exceeds target by 50%
- Cache hit rate drops below 70%
- Flaky test detected
- Coverage drops by > 5%

**Info Alerts** (Dashboard Only):
- Build queue depth > 5
- Artifact storage > 80% capacity
- Scheduled cache warm-up fails

### Dashboards

**Pipeline Overview Dashboard**:
- Build success rate (7-day, 30-day)
- Average build time trend
- Flake rate by job
- Cache efficiency

**Resource Dashboard**:
- Runner utilization
- Queue wait times
- Artifact storage usage
- Network bandwidth usage

**Test Dashboard**:
- Test execution time by suite
- Failure rate by test category
- Coverage trends
- Flaky test list

---

## Security Considerations

### Secrets Management

**Required Secrets**:
- `GHCR_TOKEN` - GitHub Container Registry push token
- `COSIGN_KEY` - Image signing key (optional)
- `SLACK_WEBHOOK` - Build notification webhook (optional)

**Security Practices**:
- Secrets never logged or exposed in artifacts
- Secrets scoped to minimum required jobs
- Secrets rotated quarterly
- Use OIDC tokens where possible (GitHub Actions)

### Image Security

**Build-Time**:
- Multi-stage builds to minimize attack surface
- Distroless base image (no shell)
- Run as non-root user (UID 65532)
- Read-only root filesystem

**Scan-Time**:
- Trivy scan for CVEs
- SBOM generation for supply chain transparency
- Base image verification

**Runtime** (Not CI, but documented here):
- Pod Security Admission (restricted)
- NetworkPolicy enforcement
- RBAC least privilege

### Supply Chain Security

**Dependencies**:
- Go modules verified with checksums
- Dependabot for automated updates
- Regular dependency audits

**Provenance**:
- SLSA Level 3 compliance (goal)
- Signed commits required for releases
- Signed container images (optional)
- Verifiable build provenance

---

## Disaster Recovery

### Build Infrastructure Failure

**Scenario**: GitHub Actions unavailable

**Mitigation**:
1. Document local build process
2. Maintain local build scripts
3. Have backup CI option (CircleCI, GitLab CI)

**Recovery Time Objective**: < 4 hours

### Cache Corruption

**Scenario**: Cache causes persistent build failures

**Mitigation**:
1. Clear cache via API or workflow
2. Rebuild cache from scratch
3. Use cache version suffix for invalidation

**Recovery Time Objective**: < 30 minutes

### Registry Unavailable

**Scenario**: Container registry unreachable

**Mitigation**:
1. Use fallback registry (Docker Hub, Quay)
2. Local builds continue with local registry
3. Queue releases until registry recovers

**Recovery Time Objective**: < 2 hours

### Test Cluster Failure

**Scenario**: Kind cluster creation fails consistently

**Mitigation**:
1. Fall back to minikube or k3d
2. Use pre-created test cluster (static)
3. Skip E2E temporarily with manual gate

**Recovery Time Objective**: < 1 hour

---

## Optimization Opportunities

### Current Bottlenecks

1. **Kind Cluster Creation** (2 min):
   - Opportunity: Use pre-warmed clusters
   - Potential Savings: 1.5 min (75% reduction)

2. **Multi-arch Image Build** (4 min):
   - Opportunity: Use remote BuildKit caching
   - Potential Savings: 2 min (50% reduction)

3. **E2E Test Suite** (8 min):
   - Opportunity: Parallelize independent scenarios
   - Potential Savings: 4 min (50% reduction)

### Future Enhancements

**Phase 2** (Post v0.1):
- Matrix testing for multiple Kubernetes versions
- Performance benchmarking job
- Helm chart testing
- Upgrade path testing

**Phase 3** (Post v0.2):
- Chaos engineering tests
- Multi-cluster testing
- Load testing
- Soak testing (24-hour runs)

**Phase 4** (Post v1.0):
- Automated security patching
- Auto-generated API docs
- Preview environments per PR
- Canary releases

---

## Runbook: Common Issues

### Issue: Build Times Exceed Budget

**Symptoms**:
- Smoke test takes > 10 minutes
- Developers complaining about slow feedback

**Diagnosis**:
```bash
# Check recent build times
gh run list --workflow=ci.yml --limit=20 --json durationMs,conclusion

# Identify slow job
gh run view <run-id> --log | grep "Duration:"
```

**Resolution**:
1. Check cache hit rates
2. Review runner capacity
3. Consider parallelization improvements
4. Profile slow tests

### Issue: High Flake Rate

**Symptoms**:
- Tests failing intermittently
- Build queue backing up due to retries

**Diagnosis**:
```bash
# List recent failures
gh run list --workflow=ci.yml --status=failure --limit=50

# Check for patterns
gh run view <run-id> --log | grep -A 5 "FAIL:"
```

**Resolution**:
1. Identify flaky test
2. Add [Flaky] tag
3. Investigate root cause (timing, resources, race conditions)
4. Fix or quarantine

### Issue: Cache Miss Rate High

**Symptoms**:
- Builds slower than expected
- Cache restore shows "no cache found"

**Diagnosis**:
```bash
# Check cache keys in workflow logs
gh run view <run-id> --log | grep "Cache key:"
gh run view <run-id> --log | grep "Cache restored from key:"
```

**Resolution**:
1. Verify cache key patterns
2. Check if cache size exceeds limits
3. Review cache retention settings
4. Consider cache warming

### Issue: Security Scan Blocks Release

**Symptoms**:
- Trivy scan fails with HIGH/CRITICAL CVEs
- Release cannot proceed

**Diagnosis**:
```bash
# View scan results
gh run view <run-id> --log-failed | grep -A 20 "Trivy scan"

# Download full report
gh run download <run-id> -n trivy-report.json
```

**Resolution**:
1. Review CVE details and severity
2. Update affected dependencies
3. If no fix available, document exception
4. Consider alternative dependency
5. Update base image if CVE is in OS packages

---

## Appendix A: Tool Versions

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.21+ | Language runtime |
| golangci-lint | 1.54+ | Code linting |
| controller-gen | 0.13+ | CRD generation |
| envtest | 0.16+ | Integration testing |
| kind | 0.20+ | Local Kubernetes |
| docker | 24.0+ | Image building |
| trivy | 0.45+ | Security scanning |
| cosign | 2.2+ | Image signing |
| syft | 0.90+ | SBOM generation |

## Appendix B: Workflow Triggers

| Trigger | Workflows | Rationale |
|---------|-----------|-----------|
| `push` to `main` | ci.yml | Validate main branch |
| `pull_request` | ci.yml | Validate before merge |
| `push` tag `v*` | release.yml | Create release |
| `schedule` (daily) | cache-warm.yml | Pre-populate caches |
| `workflow_dispatch` | ci.yml, release.yml | Manual trigger |

## Appendix C: GitHub Actions Syntax Reference

**Concurrency Control**:
```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true  # For PRs only
```

**Cache Usage**:
```yaml
- uses: actions/cache@v3
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

**Artifact Upload**:
```yaml
- uses: actions/upload-artifact@v3
  if: always()  # Upload even on failure
  with:
    name: test-results
    path: |
      coverage.out
      test-results.json
    retention-days: 90
```

---

## Document Version

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-28 | ci-release-engineer | Initial pipeline design for v0.1 |

## Related Documents

- [WORKFLOWS-STUBS.md](WORKFLOWS-STUBS.md) - GitHub Actions workflow structure
- [ARTIFACTS.md](ARTIFACTS.md) - Build artifacts strategy
- [../release/RELEASE-POLICY.md](../release/RELEASE-POLICY.md) - Release management
- [../testing/TEST-STRATEGY.md](../testing/TEST-STRATEGY.md) - Test plans and coverage
- [../security/SECURITY-CHECKLIST.md](../security/SECURITY-CHECKLIST.md) - Security requirements
