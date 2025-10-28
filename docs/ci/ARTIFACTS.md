# CI/CD Build Artifacts Strategy

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** ci-release-engineer

This document defines the artifact generation, storage, retention, and usage strategy for Kyklos CI/CD pipelines. Clear artifact management is critical for debugging failures, maintaining audit trails, and enabling release processes.

---

## Table of Contents

1. [Artifact Categories](#artifact-categories)
2. [Upload Strategy by Job](#upload-strategy-by-job)
3. [Retention Policies](#retention-policies)
4. [Size Expectations and Budgets](#size-expectations-and-budgets)
5. [Failure Triage Guide](#failure-triage-guide)
6. [Artifact Naming Conventions](#artifact-naming-conventions)
7. [Download and Inspection](#download-and-inspection)

---

## Artifact Categories

### Category 1: Test Results

**Purpose:** Prove test execution, enable coverage analysis, support failure investigation

**Includes:**
- Test output logs (stdout/stderr)
- Coverage profiles (coverage.out)
- Coverage HTML reports (coverage.html)
- JUnit XML results (test-results.xml)
- Benchmark results (benchmarks.json)

**Primary Consumers:**
- Developers debugging test failures
- Code review tools (coverage bots)
- Test result dashboards

---

### Category 2: Build Artifacts

**Purpose:** Verify build success, enable binary distribution

**Includes:**
- Controller binaries (linux/amd64, linux/arm64)
- SBOM files (sbom.json)
- Build metadata (build-info.json)
- Checksums (checksums.txt)

**Primary Consumers:**
- Release process workflows
- Verification scripts
- Manual testing

---

### Category 3: Container Images

**Purpose:** Deployable artifacts for releases and testing

**Includes:**
- Multi-arch container images
- Image manifests
- Image digests
- Layer metadata

**Storage:** Container registry (ghcr.io), not GitHub artifacts

**Primary Consumers:**
- Release workflows
- E2E test jobs
- End users

---

### Category 4: Diagnostic Bundles

**Purpose:** Debug failures, understand system state at failure time

**Includes:**
- Controller logs (controller-logs.txt)
- Cluster resource dumps (cluster-state.yaml)
- Kind cluster logs export (kind-logs.tar.gz)
- Event timelines (events.txt)
- Pod describe outputs (pod-describe.txt)

**Primary Consumers:**
- CI engineers debugging flaky tests
- Developers reproducing failures locally

---

### Category 5: Security Artifacts

**Purpose:** Audit trail, compliance, vulnerability tracking

**Includes:**
- Trivy scan results (trivy-report.json, trivy-results.sarif)
- SBOM (Software Bill of Materials)
- Security summary reports (security-summary.txt)

**Primary Consumers:**
- Security team
- Compliance auditors
- Release approval process

---

### Category 6: Documentation Artifacts

**Purpose:** Generated docs, changelogs, release notes

**Includes:**
- Generated API documentation
- Changelog files (CHANGELOG-v0.1.0.md)
- Release notes (RELEASE-NOTES-v0.1.0.md)

**Primary Consumers:**
- Documentation site builds
- Release process
- Users

---

## Upload Strategy by Job

### Job: Lint

**Artifacts Generated:** None (fast feedback, no preservation needed)

**Exception:** On failure, upload lint output for analysis

```yaml
- name: Upload lint failures
  if: failure()
  uses: actions/upload-artifact@v3
  with:
    name: lint-failures-${{ github.run_id }}
    path: lint-output.txt
    retention-days: 30
```

**Rationale:** Lint failures are code issues, not infrastructure issues. Developers can reproduce locally.

---

### Job: Unit Tests

**Artifacts Generated:**

| Artifact Name | Contents | Condition | Retention |
|--------------|----------|-----------|-----------|
| `unit-test-coverage` | coverage.out, coverage.html | Always | 30 days |
| `unit-test-results` | test-results.json | Always | 30 days |
| `unit-test-failures` | Failed test logs | On failure | 90 days |

**Upload Strategy:**

```yaml
- name: Upload coverage reports
  uses: actions/upload-artifact@v3
  if: always()  # Upload even if tests fail
  with:
    name: unit-test-coverage-${{ github.run_id }}
    path: |
      coverage.out
      coverage.html
    retention-days: 30

- name: Upload failure diagnostics
  uses: actions/upload-artifact@v3
  if: failure()
  with:
    name: unit-test-failures-${{ github.run_id }}
    path: |
      /tmp/test-*.log
      test-failures.txt
    retention-days: 90
```

**Size Expectations:**
- coverage.out: ~500 KB
- coverage.html: ~2 MB
- test-results.json: ~100 KB
- Total: ~3 MB per run

---

### Job: Envtest

**Artifacts Generated:**

| Artifact Name | Contents | Condition | Retention |
|--------------|----------|-----------|-----------|
| `envtest-logs` | API server logs, test output | On failure | 90 days |
| `envtest-resources` | Dumped Kubernetes resources | On failure | 90 days |

**Upload Strategy:**

```yaml
- name: Export envtest diagnostics
  if: failure()
  run: |
    # Save API server logs
    cp /tmp/envtest-*/kube-apiserver.log ./
    # Dump all created resources
    kubectl get all --all-namespaces -o yaml > test-namespaces.yaml

- name: Upload envtest artifacts
  uses: actions/upload-artifact@v3
  if: failure()
  with:
    name: envtest-diagnostics-${{ github.run_id }}
    path: |
      kube-apiserver.log
      etcd.log
      test-namespaces.yaml
      envtest-output.txt
    retention-days: 90
```

**Size Expectations:**
- kube-apiserver.log: ~5 MB
- test-namespaces.yaml: ~2 MB
- Total: ~8 MB per failure

**Rationale:** Envtest failures are often timing or resource issues. Full diagnostics help distinguish flakes from real bugs.

---

### Job: Build Binary

**Artifacts Generated:**

| Artifact Name | Contents | Condition | Retention |
|--------------|----------|-----------|-----------|
| `controller-linux-amd64` | Binary for x86_64 Linux | Always | 7 days |
| `controller-linux-arm64` | Binary for ARM64 Linux | Always | 7 days |
| `build-info` | Build metadata JSON | Always | 7 days |

**Upload Strategy:**

```yaml
- name: Upload controller binary
  uses: actions/upload-artifact@v3
  with:
    name: controller-${{ matrix.platform.os }}-${{ matrix.platform.arch }}-${{ github.run_id }}
    path: |
      bin/controller
      build-info.json
    retention-days: 7
```

**Size Expectations:**
- controller binary: ~15 MB (static Go binary)
- build-info.json: ~1 KB
- Total per platform: ~15 MB
- Total both platforms: ~30 MB

**Rationale:** Short retention (7 days) since images are the primary artifacts. Binaries mainly for verification and local testing.

---

### Job: Verify

**Artifacts Generated:** None (lightweight checks)

**Exception:** On secret detection, upload scan results

```yaml
- name: Upload secret scan results
  if: failure()
  uses: actions/upload-artifact@v3
  with:
    name: secret-scan-${{ github.run_id }}
    path: gitleaks-report.json
    retention-days: 365  # Security issues kept long-term
```

---

### Job: Kind Smoke Test

**Artifacts Generated:**

| Artifact Name | Contents | Condition | Retention |
|--------------|----------|-----------|-----------|
| `smoke-test-diagnostics` | Full cluster dump | On failure | 90 days |
| `smoke-test-results` | Test outcomes JSON | Always | 30 days |
| `smoke-test-metrics` | Scraped Prometheus metrics | Always | 30 days |

**Upload Strategy:**

```yaml
- name: Export Kind cluster logs
  if: failure()
  run: |
    kind export logs /tmp/kind-logs --name kyklos-dev
    tar czf kind-logs.tar.gz /tmp/kind-logs/

- name: Capture cluster state
  if: failure()
  run: |
    kubectl get all --all-namespaces -o yaml > cluster-state.yaml
    kubectl get events --all-namespaces --sort-by='.lastTimestamp' > events.txt
    kubectl logs -n kyklos-system -l app=kyklos-controller --tail=500 > controller-logs.txt

- name: Scrape metrics
  run: |
    kubectl port-forward -n kyklos-system deploy/kyklos-controller 8080:8080 &
    sleep 2
    curl -s http://localhost:8080/metrics > metrics-snapshot.txt
    kill %1

- name: Upload smoke test diagnostics
  uses: actions/upload-artifact@v3
  if: failure()
  with:
    name: smoke-test-diagnostics-${{ github.run_id }}
    path: |
      kind-logs.tar.gz
      cluster-state.yaml
      events.txt
      controller-logs.txt
    retention-days: 90

- name: Upload smoke test results
  uses: actions/upload-artifact@v3
  if: always()
  with:
    name: smoke-test-results-${{ github.run_id }}
    path: |
      smoke-test-results.json
      metrics-snapshot.txt
    retention-days: 30
```

**Size Expectations:**
- kind-logs.tar.gz: ~20 MB (compressed)
- cluster-state.yaml: ~5 MB
- events.txt: ~500 KB
- controller-logs.txt: ~2 MB
- Total: ~28 MB per failure

**Rationale:** Smoke test is the most complex job. Comprehensive diagnostics are critical for debugging timing issues, race conditions, and flakes.

---

### Job: Multi-Arch Image Build

**Artifacts Generated:**

| Artifact Name | Contents | Condition | Retention |
|--------------|----------|-----------|-----------|
| `sbom` | Software Bill of Materials | Always | 365 days |
| `image-manifest` | Multi-arch manifest JSON | Always | 365 days |
| `image-digests` | SHA256 digests | Always | 365 days |

**Upload Strategy:**

```yaml
- name: Generate SBOM
  run: |
    syft packages ghcr.io/${{ github.repository }}:${{ steps.version.outputs.version }} \
      -o cyclonedx-json > sbom.json

- name: Extract image digests
  run: |
    docker buildx imagetools inspect ghcr.io/${{ github.repository }}:${{ steps.version.outputs.version }} \
      --format "{{json .Manifest}}" > image-manifest.json
    docker buildx imagetools inspect ghcr.io/${{ github.repository }}:${{ steps.version.outputs.version }} \
      --format "{{.Manifest.Digest}}" > image-digests.txt

- name: Upload SBOM and manifests
  uses: actions/upload-artifact@v3
  with:
    name: image-metadata-${{ github.run_id }}
    path: |
      sbom.json
      image-manifest.json
      image-digests.txt
    retention-days: 365  # Long-term for compliance
```

**Size Expectations:**
- sbom.json: ~100 KB
- image-manifest.json: ~5 KB
- image-digests.txt: ~200 bytes
- Total: ~105 KB

**Rationale:** Small artifacts, critical for supply chain security. Kept for full year to support audit trails.

---

### Job: Security Scan

**Artifacts Generated:**

| Artifact Name | Contents | Condition | Retention |
|--------------|----------|-----------|-----------|
| `security-scan` | Trivy SARIF and summary | Always | 365 days |

**Upload Strategy:**

```yaml
- name: Upload Trivy results to GitHub Security
  uses: github/codeql-action/upload-sarif@v2
  if: always()
  with:
    sarif_file: trivy-results.sarif

- name: Generate human-readable summary
  run: |
    trivy image ghcr.io/${{ github.repository }}:${{ steps.version.outputs.version }} \
      --format table --severity HIGH,CRITICAL > security-summary.txt

- name: Upload security artifacts
  uses: actions/upload-artifact@v3
  if: always()
  with:
    name: security-scan-${{ github.run_id }}
    path: |
      trivy-results.sarif
      security-summary.txt
    retention-days: 365
```

**Size Expectations:**
- trivy-results.sarif: ~50 KB
- security-summary.txt: ~10 KB
- Total: ~60 KB

**Rationale:** Security artifacts kept indefinitely (365 days) for compliance and vulnerability tracking over time.

---

### Job: E2E Full Test Suite

**Artifacts Generated:**

| Artifact Name | Contents | Condition | Retention |
|--------------|----------|-----------|-----------|
| `e2e-results` | JUnit XML, scenario traces | Always | 30 days |
| `e2e-diagnostics` | Full cluster state on failure | On failure | 90 days |

**Upload Strategy:**

```yaml
- name: Run E2E scenarios with trace capture
  run: |
    for scenario in office-hours night-shift dst-transition; do
      ./test/e2e/run-scenario.sh $scenario 2>&1 | tee scenario-traces/${scenario}.log
    done

- name: Export E2E cluster state on failure
  if: failure()
  run: |
    kubectl get all --all-namespaces -o yaml > e2e-cluster-state.yaml
    kubectl get events --all-namespaces --sort-by='.lastTimestamp' > e2e-events.txt
    kubectl logs -n kyklos-system -l app=kyklos-controller --tail=1000 > e2e-controller-logs.txt
    tar czf e2e-logs.tar.gz /tmp/e2e-logs/

- name: Upload E2E results
  uses: actions/upload-artifact@v3
  if: always()
  with:
    name: e2e-results-${{ github.run_id }}
    path: |
      test/e2e/results.xml
      scenario-traces/
    retention-days: 30

- name: Upload E2E diagnostics
  uses: actions/upload-artifact@v3
  if: failure()
  with:
    name: e2e-diagnostics-${{ github.run_id }}
    path: |
      e2e-cluster-state.yaml
      e2e-events.txt
      e2e-controller-logs.txt
      e2e-logs.tar.gz
    retention-days: 90
```

**Size Expectations:**
- results.xml: ~200 KB
- scenario-traces/: ~10 MB (all scenarios)
- e2e-cluster-state.yaml: ~10 MB (on failure)
- e2e-logs.tar.gz: ~50 MB (on failure)
- Total success: ~10 MB
- Total failure: ~70 MB

**Rationale:** E2E tests run complex multi-minute scenarios. Detailed traces and full cluster state essential for debugging.

---

### Job: Publish Release Artifacts

**Artifacts Generated:**

| Artifact Name | Contents | Condition | Retention |
|--------------|----------|-----------|-----------|
| `release-assets` | install.yaml, checksums | Always | 365 days |
| `source-archive` | kyklos-v0.1.0.tar.gz | Always | 365 days |

**Upload Strategy:**

```yaml
- name: Generate installation bundle
  run: |
    kustomize build config/default > install.yaml

- name: Create source archive
  run: |
    git archive --format=tar.gz --prefix=kyklos-${{ steps.version.outputs.version }}/ \
      HEAD > kyklos-${{ steps.version.outputs.version }}.tar.gz

- name: Generate checksums
  run: |
    sha256sum install.yaml sbom.json kyklos-*.tar.gz > checksums.txt

- name: Upload release artifacts
  uses: actions/upload-artifact@v3
  with:
    name: release-assets-${{ github.ref_name }}
    path: |
      install.yaml
      kyklos-*.tar.gz
      checksums.txt
      sbom.json
    retention-days: 365
```

**Size Expectations:**
- install.yaml: ~50 KB
- kyklos-v0.1.0.tar.gz: ~500 KB
- checksums.txt: ~500 bytes
- sbom.json: ~100 KB (referenced from previous job)
- Total: ~650 KB

**Rationale:** Release artifacts are tiny and critical. Kept indefinitely (365 days) as official release record.

---

### Job: Generate Release Notes

**Artifacts Generated:**

| Artifact Name | Contents | Condition | Retention |
|--------------|----------|-----------|-----------|
| `changelog` | CHANGELOG and RELEASE-NOTES | Always | 365 days |

**Upload Strategy:**

```yaml
- name: Generate changelog
  run: |
    ./scripts/generate-changelog.sh ${{ github.ref_name }} > CHANGELOG-${{ github.ref_name }}.md

- name: Generate release notes
  run: |
    ./scripts/generate-release-notes.sh ${{ github.ref_name }} > RELEASE-NOTES-${{ github.ref_name }}.md

- name: Upload changelog artifacts
  uses: actions/upload-artifact@v3
  with:
    name: changelog-${{ github.ref_name }}
    path: |
      CHANGELOG-${{ github.ref_name }}.md
      RELEASE-NOTES-${{ github.ref_name }}.md
    retention-days: 365
```

**Size Expectations:**
- CHANGELOG: ~20 KB
- RELEASE-NOTES: ~10 KB
- Total: ~30 KB

---

## Retention Policies

### Policy Matrix

| Artifact Category | Success Retention | Failure Retention | Rationale |
|------------------|-------------------|-------------------|-----------|
| Test Coverage | 30 days | 30 days | Coverage trends tracked elsewhere |
| Test Failures | N/A | 90 days | Failure investigation, flake analysis |
| Binaries | 7 days | 7 days | Images are primary artifacts |
| Container Images | Indefinite | N/A | Stored in registry, not artifacts |
| Smoke Test Results | 30 days | 90 days | Monitor smoke test health trends |
| Smoke Test Diagnostics | N/A | 90 days | Debug complex integration issues |
| E2E Results | 30 days | 90 days | Scenario execution trends |
| E2E Diagnostics | N/A | 90 days | Long-running test debugging |
| Security Scans | 365 days | 365 days | Compliance audit trail |
| SBOM | 365 days | 365 days | Supply chain security record |
| Release Assets | 365 days | N/A | Official release artifacts |
| Changelogs | 365 days | N/A | Documentation history |

### Retention Justifications

**7 Days (Short-Term):**
- Binaries: Superseded by container images
- Quick verification artifacts

**30 Days (Medium-Term):**
- Test results for trend analysis
- Coverage tracking
- Normal build outcomes

**90 Days (Long-Term Debugging):**
- Test failure diagnostics
- Flake investigation data
- Complex integration test artifacts

**365 Days (Compliance/Audit):**
- Security scan results
- SBOM files
- Release artifacts
- Official documentation

---

## Size Expectations and Budgets

### Per-Job Storage Budget

| Job | Avg Success Size | Avg Failure Size | Max Failure Size |
|-----|-----------------|------------------|------------------|
| Lint | 0 MB | 1 MB | 5 MB |
| Unit Test | 3 MB | 5 MB | 10 MB |
| Envtest | 0 MB | 8 MB | 20 MB |
| Build | 30 MB | 30 MB | 30 MB |
| Verify | 0 MB | 1 MB | 5 MB |
| Smoke Test | 1 MB | 28 MB | 50 MB |
| Image Build | 0.1 MB | 0.1 MB | 0.1 MB |
| Security Scan | 0.06 MB | 0.06 MB | 0.06 MB |
| E2E Full | 10 MB | 70 MB | 100 MB |
| Publish | 0.65 MB | N/A | N/A |
| Release Notes | 0.03 MB | N/A | N/A |

### Monthly Storage Estimates

**Assumptions:**
- 100 PR builds per month
- 10 release builds per month
- 10% failure rate for PR builds
- 0% failure rate for release builds (must pass)

**PR Builds:**
- Success: 90 builds × 44 MB = 3,960 MB
- Failure: 10 builds × 169 MB = 1,690 MB
- Total: 5,650 MB (~5.5 GB)

**Release Builds:**
- Success: 10 builds × 45 MB = 450 MB
- Total: 450 MB

**Monthly Total: ~6 GB**

**Annual Total: ~72 GB**

### Storage Budget Alerts

**Warning:** Monthly storage exceeds 10 GB
**Action:** Review failure rate and artifact retention

**Critical:** Monthly storage exceeds 20 GB
**Action:** Immediate cleanup of old artifacts

---

## Failure Triage Guide

### Quick Triage Matrix

| Symptom | Artifact to Check | Key Indicators | Likely Cause |
|---------|------------------|----------------|--------------|
| Unit test fails | unit-test-coverage | Specific test name, line number | Code bug |
| Unit test timeout | unit-test-failures | Hanging test, stuck goroutine | Deadlock, missing timeout |
| Envtest fails | envtest-logs | API server errors, etcd logs | Resource timing, API incompatibility |
| Build fails | Build logs (stdout) | Compilation error, linker error | Code syntax, missing dependency |
| Smoke test fails | smoke-test-diagnostics | Pod status, events, controller logs | Deployment issue, timing problem |
| E2E fails | e2e-diagnostics | Scenario traces, timestamps | Complex interaction, race condition |
| Security scan fails | security-scan | CVE IDs, severity | Vulnerable dependency |

### Triage Workflows

#### Workflow 1: Unit Test Failure

**Step 1: Download artifacts**
```bash
gh run download <run-id> -n unit-test-failures-<run-id>
```

**Step 2: Identify failed test**
```bash
grep "FAIL:" test-failures.txt
```

**Step 3: Check coverage impact**
```bash
# Compare coverage.out with previous run
go tool cover -func=coverage.out | grep -A 5 "failed_package"
```

**Step 4: Reproduce locally**
```bash
go test -v -run TestSpecificFailure ./pkg/...
```

#### Workflow 2: Smoke Test Failure

**Step 1: Download comprehensive diagnostics**
```bash
gh run download <run-id> -n smoke-test-diagnostics-<run-id>
```

**Step 2: Extract and review logs**
```bash
tar xzf kind-logs.tar.gz
less kind-logs/kyklos-dev-control-plane/kubelet.log
```

**Step 3: Examine controller behavior**
```bash
grep "ERROR\|WARN" controller-logs.txt
```

**Step 4: Check event timeline**
```bash
cat events.txt | grep demo  # Focus on demo namespace
```

**Step 5: Review cluster state at failure**
```bash
kubectl apply -f cluster-state.yaml --dry-run=client
# Look for resource status, error messages
```

**Step 6: Identify failure category**
- **Timing issue:** Events show correct sequence but took too long
- **Logic bug:** Events show incorrect scaling behavior
- **Infrastructure flake:** Node issues, image pull failures
- **Configuration error:** RBAC, CRD, or manifest issues

#### Workflow 3: E2E Scenario Failure

**Step 1: Download E2E artifacts**
```bash
gh run download <run-id> -n e2e-results-<run-id>
gh run download <run-id> -n e2e-diagnostics-<run-id>  # If failure
```

**Step 2: Review scenario trace**
```bash
less scenario-traces/dst-transition.log
```

**Step 3: Correlate with controller logs**
```bash
# Find timestamps in scenario log
grep "T+4:00" scenario-traces/dst-transition.log
# Cross-reference with controller decisions
grep "2025-10-28T14:41" e2e-controller-logs.txt
```

**Step 4: Check for timing assumptions**
```bash
# Look for hardcoded sleep statements
grep "sleep" scenario-traces/*.log
# Verify wait conditions
grep "wait\|timeout" scenario-traces/*.log
```

#### Workflow 4: Security Scan Failure

**Step 1: Download security scan**
```bash
gh run download <run-id> -n security-scan-<run-id>
```

**Step 2: Review human-readable summary**
```bash
cat security-summary.txt
```

**Step 3: Examine SARIF for details**
```bash
jq '.runs[0].results[] | select(.level == "error")' trivy-results.sarif
```

**Step 4: Check affected dependencies**
```bash
# Cross-reference with SBOM
jq '.components[] | select(.name == "vulnerable-dep")' ../sbom.json
```

**Step 5: Determine remediation**
- **Patch available:** Update dependency version
- **No patch:** Document exception, add suppression
- **False positive:** Report to Trivy, add exception

---

## Artifact Naming Conventions

### Pattern: `<job>-<category>-<identifier>`

**Components:**
- `<job>`: Job name (lint, unit-test, smoke-test, etc.)
- `<category>`: Artifact type (results, diagnostics, coverage, etc.)
- `<identifier>`: Unique ID (github.run_id or version tag)

**Examples:**
- `unit-test-coverage-1234567890`
- `smoke-test-diagnostics-1234567890`
- `e2e-results-1234567890`
- `release-assets-v0.1.0`
- `changelog-v0.1.0`

### Why Use run_id?

**Uniqueness:** Guarantees no collisions across builds
**Traceability:** Easy to map artifacts back to workflow runs
**Cleanup:** Enables automated cleanup by run ID

### When to Use Version Tags?

**Release artifacts:** Use semantic version (v0.1.0)
- `release-assets-v0.1.0`
- `changelog-v0.1.0`
- `sbom-v0.1.0`

**Rationale:** Human-readable, stable identifiers for releases

---

## Download and Inspection

### Using GitHub CLI

**List all artifacts for a run:**
```bash
gh run view <run-id> --log-failed
gh run view <run-id> --json artifacts --jq '.artifacts[] | {name, size, expired}'
```

**Download specific artifact:**
```bash
gh run download <run-id> -n smoke-test-diagnostics-<run-id>
```

**Download all artifacts:**
```bash
gh run download <run-id>
```

### Using GitHub Web UI

1. Navigate to Actions → <workflow run>
2. Scroll to "Artifacts" section
3. Click artifact name to download

### Automated Artifact Aggregation

**Example: Aggregate all coverage reports**
```bash
#!/bin/bash
for run in $(gh run list --workflow ci.yml --limit 30 --json databaseId -q '.[].databaseId'); do
  gh run download $run -n unit-test-coverage-$run || true
done

# Merge coverage reports
go tool covdata merge -i=./**/coverage.out -o=merged-coverage.out
```

---

## Artifact Compression Best Practices

### When to Compress

**Always compress:**
- Log directories (10+ files)
- Text files > 1 MB
- Kubernetes resource dumps

**Example:**
```yaml
- name: Compress logs before upload
  run: |
    tar czf logs.tar.gz /tmp/logs/

- name: Upload compressed logs
  uses: actions/upload-artifact@v3
  with:
    path: logs.tar.gz
```

**Space Savings:** 70-90% for text logs

### When NOT to Compress

**Don't compress:**
- Single small files (< 100 KB)
- Binary files (already compressed)
- JSON files used by subsequent jobs

**Example:**
```yaml
- name: Upload JSON results (uncompressed)
  uses: actions/upload-artifact@v3
  with:
    path: test-results.json  # Next job parses this
```

---

## Artifact Security

### Sensitive Data Handling

**Never upload:**
- Secrets or credentials
- Private keys
- API tokens
- User data

**Sanitization Example:**
```yaml
- name: Sanitize logs before upload
  run: |
    # Remove secret environment variables
    sed -i 's/GHCR_TOKEN=.*/GHCR_TOKEN=***REDACTED***/g' logs.txt
    # Remove base64 encoded secrets
    sed -i 's/data:.*base64.*/data: ***REDACTED***/g' cluster-state.yaml

- name: Upload sanitized artifacts
  uses: actions/upload-artifact@v3
  with:
    path: |
      logs.txt
      cluster-state.yaml
```

### Access Control

**GitHub Artifacts:**
- Accessible to repository collaborators
- Not public by default
- Require authentication via GitHub CLI or web UI

**Release Assets:**
- Public by default
- Ensure no sensitive data in release artifacts

---

## Monitoring Artifact Health

### Key Metrics

**Track monthly:**
- Total artifact size uploaded
- Average artifact size per job
- Failure artifact count (indicates flakiness)
- Expired artifacts not cleaned up

**Dashboard Queries:**
```graphql
# GraphQL query for artifact metrics
query ArtifactMetrics($owner: String!, $repo: String!) {
  repository(owner: $owner, name: $repo) {
    workflowRuns(last: 100) {
      nodes {
        artifacts {
          totalCount
          nodes {
            name
            sizeInBytes
          }
        }
      }
    }
  }
}
```

### Cleanup Automation

**Script: Clean up expired artifacts**
```bash
#!/bin/bash
# Delete artifacts older than retention period
gh api repos/:owner/:repo/actions/artifacts \
  --jq '.artifacts[] | select(.expired == true) | .id' | \
  while read artifact_id; do
    gh api -X DELETE repos/:owner/:repo/actions/artifacts/$artifact_id
  done
```

---

## Troubleshooting Common Issues

### Issue: Artifact Upload Fails

**Symptoms:**
- `Error: Unable to upload artifact`
- Workflow hangs at upload step

**Diagnosis:**
```bash
# Check artifact size
du -sh /path/to/artifact
# Check disk space on runner
df -h
```

**Causes:**
- Artifact exceeds GitHub limit (10 GB per workflow run)
- Runner out of disk space
- Network timeout

**Solutions:**
1. Compress large artifacts
2. Upload only essential files
3. Split into multiple smaller artifacts
4. Increase timeout for upload step

### Issue: Artifact Not Found

**Symptoms:**
- `Error: Unable to find artifact`
- Download fails in dependent job

**Diagnosis:**
```bash
# Check if artifact was uploaded
gh run view <run-id> --json artifacts
```

**Causes:**
- Conditional upload (if: failure()) didn't run
- Job failed before upload step
- Artifact expired
- Name mismatch

**Solutions:**
1. Use `if: always()` for critical artifacts
2. Upload earlier in job
3. Verify artifact name matches exactly

### Issue: Artifact Retention Not Working

**Symptoms:**
- Artifacts deleted sooner than expected
- Old artifacts still present beyond retention

**Diagnosis:**
```bash
# Check artifact expiration date
gh api repos/:owner/:repo/actions/artifacts/<id> | jq '.expires_at'
```

**Causes:**
- Organizational policy overrides retention-days
- Manual deletion
- Artifact retention calculation bug

**Solutions:**
1. Verify organization settings
2. Use longer retention periods
3. Download critical artifacts externally

---

## Document Version

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-28 | ci-release-engineer | Initial artifact strategy for v0.1 |

## Related Documents

- [PIPELINE.md](/Users/aykumar/personal/kyklos/docs/ci/PIPELINE.md) - Overall CI pipeline design
- [WORKFLOWS-STUBS.md](/Users/aykumar/personal/kyklos/docs/ci/WORKFLOWS-STUBS.md) - GitHub Actions workflow structure
- [../testing/TEST-STRATEGY.md](/Users/aykumar/personal/kyklos/docs/testing/TEST-STRATEGY.md) - Test strategy
- [../testing/FLAKE-POLICY.md](/Users/aykumar/personal/kyklos/docs/testing/FLAKE-POLICY.md) - Flake handling
