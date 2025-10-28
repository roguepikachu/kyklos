# Release Notes Template

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** ci-release-engineer

This document provides templates for creating user-facing release notes. Release notes should be clear, concise, and actionable for end users.

---

## Table of Contents

1. [Template for Stable Releases](#template-for-stable-releases)
2. [Template for Pre-Release Versions](#template-for-pre-release-versions)
3. [Template for Patch Releases](#template-for-patch-releases)
4. [Template for Security Releases](#template-for-security-releases)
5. [Writing Guidelines](#writing-guidelines)
6. [Examples](#examples)

---

## Template for Stable Releases

Use this template for major (X.0.0) and minor (0.X.0) releases.

```markdown
# Kyklos v0.2.0 Release Notes

**Release Date:** 2026-03-01
**Release Type:** Minor Release
**Previous Version:** v0.1.5

---

## Overview

<!-- 2-3 sentence summary of this release -->
Kyklos v0.2.0 introduces support for StatefulSet targets, adds granular control over scale-down rates, and improves reconciliation performance by 30%. This release maintains backward compatibility with v0.1.x configurations.

---

## What's New

### StatefulSet Target Support

Kyklos now scales StatefulSets in addition to Deployments. StatefulSets are scaled using the same time window logic, with special handling for ordered pod creation and deletion.

**Configuration Example:**
```yaml
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: db-scaler
spec:
  targetRef:
    kind: StatefulSet  # NEW: StatefulSet support
    name: postgres
  timezone: America/New_York
  defaultReplicas: 3
  windows:
    - days: [Mon, Tue, Wed, Thu, Fri]
      start: "09:00"
      end: "17:00"
      replicas: 5
```

**Documentation:** [StatefulSet Scaling Guide](https://kyklos.io/docs/statefulset)

---

### Gradual Scale-Down with maxScaleDownRate

Control how quickly replicas are removed during scale-down operations to prevent abrupt traffic drops.

**New Field:**
```yaml
spec:
  maxScaleDownRate: 5  # Remove max 5 replicas per reconcile cycle
```

**Behavior:**
- If current replicas = 20, target = 5, maxScaleDownRate = 5
- Cycle 1: Scale to 15 (remove 5)
- Cycle 2: Scale to 10 (remove 5)
- Cycle 3: Scale to 5 (remove 5)

**Documentation:** [Gradual Scale-Down](https://kyklos.io/docs/scale-down-rate)

---

### Performance Improvements

Reconciliation loop optimized with intelligent caching, reducing API server load and improving response times.

**Metrics:**
- 30% faster reconcile cycles (avg 150ms → 105ms)
- 50% reduction in API server requests
- Lower memory footprint (200 MB → 150 MB per 1000 resources)

**Benchmark Details:** [Performance Report](https://kyklos.io/perf/v0.2.0)

---

## Improvements

- **Controller:** Improved error messages with actionable remediation steps
- **Metrics:** Added `kyklos_scale_operations_rate` metric for observability
- **Logging:** Structured logging with contextual fields for better debugging
- **Docs:** New troubleshooting guide with failure diagnostics

---

## Bug Fixes

- **Controller:** Fixed cross-midnight window calculation edge case ([#456](https://github.com/kyklos/kyklos/issues/456))
- **Controller:** Resolved memory leak in watch cache ([#478](https://github.com/kyklos/kyklos/issues/478))
- **Webhook:** Corrected validation for overlapping time windows ([#491](https://github.com/kyklos/kyklos/issues/491))
- **API:** Fixed status condition race during rapid reconciles ([#502](https://github.com/kyklos/kyklos/issues/502))

---

## Security

- **Dependencies:** Updated Go to 1.21.5 (addresses CVE-2024-XXXXX)
- **Dependencies:** Updated controller-runtime to v0.17.2 (no known vulnerabilities)
- **Image:** Trivy scan clean (zero HIGH/CRITICAL CVEs)

**SBOM:** [Download](https://github.com/kyklos/kyklos/releases/download/v0.2.0/sbom.json)

---

## Deprecations

### `.status.lastUpdateTime` Field

**Deprecated:** v0.2.0
**Removed:** v0.4.0
**Replacement:** `.status.lastTransitionTime`

**Migration:**
Update your monitoring queries and dashboards to use `.status.lastTransitionTime` instead of `.status.lastUpdateTime`. Both fields contain the same data in v0.2.0 and v0.3.0.

```bash
# Old query
kubectl get tws -o jsonpath='{.status.lastUpdateTime}'

# New query
kubectl get tws -o jsonpath='{.status.lastTransitionTime}'
```

---

## Breaking Changes

None. This release is fully backward compatible with v0.1.x.

---

## Upgrade Instructions

### Prerequisites

- Kubernetes 1.26, 1.27, or 1.28
- kubectl configured with cluster admin access
- Existing Kyklos v0.1.x installation (if upgrading)

### Upgrade Steps

**1. Backup Existing Resources**
```bash
kubectl get timewindowscalers --all-namespaces -o yaml > kyklos-backup.yaml
```

**2. Update CRDs**
```bash
kubectl apply -f https://github.com/kyklos/kyklos/releases/download/v0.2.0/install.yaml
```

**3. Restart Controller**
```bash
kubectl rollout restart deployment kyklos-controller -n kyklos-system
```

**4. Verify Upgrade**
```bash
kubectl get deployment kyklos-controller -n kyklos-system -o jsonpath='{.spec.template.spec.containers[0].image}'
# Expected: ghcr.io/kyklos/controller:v0.2.0
```

**5. Test Existing Resources**
```bash
kubectl get tws --all-namespaces
kubectl get events -n <your-namespace> --field-selector involvedObject.kind=TimeWindowScaler
```

**Rollback (if needed):**
```bash
kubectl set image deployment/kyklos-controller \
  kyklos-controller=ghcr.io/kyklos/controller:v0.1.5 \
  -n kyklos-system
```

---

## Known Issues

### Issue: StatefulSet Scale-Down May Delay with PVC

**Description:** When scaling down StatefulSets with persistent volumes, pod deletion may delay if PVCs are not cleaned up.

**Workaround:** Set `podManagementPolicy: Parallel` in StatefulSet spec for faster scale-down.

**Tracking:** [#521](https://github.com/kyklos/kyklos/issues/521)

---

## Kubernetes Compatibility

| Kubernetes Version | Supported | Tested |
|-------------------|-----------|--------|
| 1.29 | Yes | Yes |
| 1.28 | Yes | Yes |
| 1.27 | Yes | Yes |
| 1.26 | Yes | Yes |
| 1.25 | No | Dropped |

**Note:** Kubernetes 1.25 support dropped as it reached end-of-life in October 2025.

---

## Installation

### Fresh Installation

```bash
kubectl apply -f https://github.com/kyklos/kyklos/releases/download/v0.2.0/install.yaml
```

### Helm Chart

```bash
helm repo add kyklos https://kyklos.io/helm
helm repo update
helm install kyklos kyklos/kyklos --version v0.2.0 --namespace kyklos-system --create-namespace
```

### Verify Installation

```bash
kubectl get pods -n kyklos-system
# Expected: kyklos-controller-manager-xxx Running
```

---

## Resources

- **Documentation:** https://kyklos.io/docs/v0.2
- **API Reference:** https://kyklos.io/api/v0.2
- **Changelog:** [CHANGELOG-v0.2.0.md](https://github.com/kyklos/kyklos/blob/main/CHANGELOG-v0.2.0.md)
- **Examples:** https://github.com/kyklos/kyklos/tree/main/examples
- **GitHub Discussions:** https://github.com/kyklos/kyklos/discussions

---

## Contributors

Thank you to all contributors who made this release possible:

- @contributor1 - StatefulSet support implementation
- @contributor2 - Performance optimization
- @contributor3 - Documentation improvements
- @contributor4 - Bug fixes and testing

Full contributor list: https://github.com/kyklos/kyklos/graphs/contributors

---

## Feedback

We'd love to hear your feedback on this release:

- Report issues: https://github.com/kyklos/kyklos/issues
- Feature requests: https://github.com/kyklos/kyklos/discussions/categories/ideas
- Community chat: Join us on Slack (#kyklos)

---

**Happy Scaling!**

The Kyklos Team
```

---

## Template for Pre-Release Versions

Use this template for alpha, beta, and release candidate versions.

```markdown
# Kyklos v0.2.0-beta.1 Release Notes

**Release Date:** 2026-02-15
**Release Type:** Beta Release
**Stable Release:** v0.2.0 (planned for 2026-03-01)

---

## ⚠️ Pre-Release Notice

This is a **BETA** release for testing and feedback. Not recommended for production use.

**What to Expect:**
- Feature-complete for v0.2.0
- API is locked (no further changes before stable)
- May contain minor bugs
- Suitable for staging/testing environments

**Feedback:** Please report issues at https://github.com/kyklos/kyklos/issues with label `v0.2.0-beta`

---

## What's New in v0.2.0-beta.1

<!-- Same sections as stable release template -->

### StatefulSet Target Support (BETA)

StatefulSet scaling is feature-complete but undergoing final testing. Please test thoroughly in non-production environments.

**Known Limitations:**
- Large StatefulSets (>100 replicas) may take longer to scale
- PVC cleanup behavior differs by cloud provider

---

## Testing Instructions

We need your help testing the following scenarios:

**Test Case 1: StatefulSet Scaling**
```yaml
# Create a StatefulSet with Kyklos scaler
kubectl apply -f examples/tws-statefulset.yaml
# Observe scaling behavior across window boundaries
# Report: Scale-up time, scale-down time, any errors
```

**Test Case 2: Gradual Scale-Down**
```yaml
# Test maxScaleDownRate with large replica counts
# Start with replicas=50, scale to 10 with maxScaleDownRate=5
# Report: Number of reconcile cycles, total time
```

**Feedback Template:** [beta-testing-report.md](https://github.com/kyklos/kyklos/issues/new?template=beta-testing.md)

---

## Upgrade from v0.2.0-alpha.2

```bash
# Update CRDs
kubectl apply -f https://github.com/kyklos/kyklos/releases/download/v0.2.0-beta.1/install.yaml

# Restart controller
kubectl rollout restart deployment kyklos-controller -n kyklos-system
```

**Changes from alpha.2:**
- Fixed StatefulSet race condition
- Improved validation error messages
- Updated dependencies

---

## Known Issues

- [#521](https://github.com/kyklos/kyklos/issues/521) - StatefulSet PVC cleanup delay
- [#528](https://github.com/kyklos/kyklos/issues/528) - Metrics lag during high load

---

## Next Steps

- **Final testing:** 2026-02-15 to 2026-02-28
- **Release candidate:** v0.2.0-rc.1 on 2026-02-28
- **Stable release:** v0.2.0 on 2026-03-01

---

**Thank you for testing Kyklos v0.2.0-beta.1!**
```

---

## Template for Patch Releases

Use this template for patch versions (0.0.X) with bug fixes.

```markdown
# Kyklos v0.1.3 Release Notes

**Release Date:** 2026-01-15
**Release Type:** Patch Release
**Previous Version:** v0.1.2

---

## Overview

Kyklos v0.1.3 is a maintenance release addressing two critical bugs and updating security dependencies. All v0.1.x users are encouraged to upgrade.

---

## Bug Fixes

### Critical: Cross-Midnight Window Calculation Error

**Issue:** TimeWindowScalers with windows spanning midnight (e.g., 22:00-02:00) incorrectly calculated active periods in certain timezones with DST.

**Impact:** Deployments may not scale at expected times if window crosses midnight during DST transition.

**Fix:** Corrected timezone offset handling in window boundary calculation.

**Affected Versions:** v0.1.0 - v0.1.2

**Issue:** [#456](https://github.com/kyklos/kyklos/issues/456)

---

### High: Memory Leak in Watch Cache

**Issue:** Controller memory usage grew unbounded over 7+ days of operation due to watch cache not releasing old entries.

**Impact:** Controller pod may be OOMKilled after extended runtime (7-14 days).

**Fix:** Implemented periodic cache cleanup with configurable retention.

**Affected Versions:** v0.1.0 - v0.1.2

**Issue:** [#478](https://github.com/kyklos/kyklos/issues/478)

---

## Security

- **Dependencies:** Updated Go to 1.21.5 (addresses CVE-2024-XXXXX)
- **Image:** Trivy scan clean (zero HIGH/CRITICAL CVEs)

---

## Upgrade Instructions

**Recommended for:** All v0.1.x users

```bash
# Update controller image
kubectl set image deployment/kyklos-controller \
  kyklos-controller=ghcr.io/kyklos/controller:v0.1.3 \
  -n kyklos-system

# Verify upgrade
kubectl get deployment kyklos-controller -n kyklos-system -o jsonpath='{.spec.template.spec.containers[0].image}'
```

**No CRD changes.** No resource migration required.

---

## Full Changelog

See [CHANGELOG-v0.1.3.md](https://github.com/kyklos/kyklos/blob/main/CHANGELOG-v0.1.3.md)

---

**Questions?** https://github.com/kyklos/kyklos/discussions
```

---

## Template for Security Releases

Use this template for releases addressing security vulnerabilities.

```markdown
# Kyklos v0.1.4 Security Release

**Release Date:** 2026-01-20
**Release Type:** Security Patch
**Previous Version:** v0.1.3
**CVE:** CVE-2026-XXXXX

---

## ⚠️ Security Advisory

This release addresses **CVE-2026-XXXXX**, a HIGH severity vulnerability in Kyklos controller RBAC handling.

**CVSS Score:** 8.5 (HIGH)

**Upgrade Immediately:** All users running v0.1.0 - v0.1.3 should upgrade to v0.1.4 or v0.2.1.

---

## Vulnerability Details

### CVE-2026-XXXXX: Insufficient RBAC Validation

**Severity:** HIGH (CVSS 8.5)

**Description:**
The Kyklos controller does not properly validate RBAC permissions when applying scale operations, allowing users with TimeWindowScaler create permissions to scale arbitrary deployments in namespaces where they lack direct access.

**Attack Scenario:**
1. Attacker has permission to create TimeWindowScaler in namespace A
2. Attacker creates TWS targeting Deployment in namespace B
3. Controller scales Deployment in namespace B on attacker's behalf
4. Result: Unauthorized scaling of workloads

**Affected Versions:** v0.1.0 - v0.1.3

**Fixed Versions:** v0.1.4, v0.2.1

---

## Impact Assessment

**Who is affected:**
- Multi-tenant clusters where users can create TimeWindowScaler resources
- Clusters with namespace-level RBAC isolation

**Who is NOT affected:**
- Single-tenant clusters
- Clusters where only cluster admins create TimeWindowScaler resources
- Clusters with admission controllers blocking cross-namespace references

---

## Remediation

### Option 1: Upgrade (Recommended)

**Upgrade to v0.1.4 or v0.2.1:**

```bash
# v0.1.x users
kubectl set image deployment/kyklos-controller \
  kyklos-controller=ghcr.io/kyklos/controller:v0.1.4 \
  -n kyklos-system

# v0.2.x users
kubectl set image deployment/kyklos-controller \
  kyklos-controller=ghcr.io/kyklos/controller:v0.2.1 \
  -n kyklos-system
```

### Option 2: Temporary Mitigation

If immediate upgrade is not possible, restrict TimeWindowScaler creation:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kyklos-create-deny
rules:
- apiGroups: ["kyklos.io"]
  resources: ["timewindowscalers"]
  verbs: ["create"]
  # Only allow trusted service accounts
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kyklos-create-deny
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kyklos-create-deny
subjects:
- kind: Group
  name: system:authenticated  # Block all users
  apiGroup: rbac.authorization.k8s.io
```

---

## Technical Details

**Root Cause:**
Controller assumed all TimeWindowScaler resources were created by authorized users and did not re-validate permissions against target workloads.

**Fix:**
Added RBAC SubjectAccessReview check before applying scale operations. Controller now verifies that the TWS creator has `update` permission on the target Deployment.

**Code Change:** [PR #510](https://github.com/kyklos/kyklos/pull/510)

---

## Verification

**Verify your cluster is protected:**

```bash
# Attempt to create TWS targeting namespace you don't have access to
kubectl apply -f - <<EOF
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: unauthorized-test
  namespace: my-namespace
spec:
  targetRef:
    kind: Deployment
    name: sensitive-app
    namespace: other-namespace  # Namespace you don't have access to
  timezone: UTC
  defaultReplicas: 1
  windows: []
EOF

# Expected result: Controller logs error and does not scale target
kubectl logs -n kyklos-system deploy/kyklos-controller | grep "permission denied"
```

---

## Credit

**Reported by:** John Doe (Company XYZ)

**Disclosure Timeline:**
- 2026-01-10: Vulnerability reported
- 2026-01-11: Confirmed by Kyklos team
- 2026-01-15: Fix developed and tested
- 2026-01-20: Public disclosure and patch release

---

## References

- **CVE:** https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2026-XXXXX
- **GitHub Advisory:** https://github.com/kyklos/kyklos/security/advisories/GHSA-XXXX
- **Fix PR:** https://github.com/kyklos/kyklos/pull/510

---

## Questions?

- Security team: security@kyklos.io
- GitHub Discussions: https://github.com/kyklos/kyklos/discussions

---

**Thank you for keeping your clusters secure.**
```

---

## Writing Guidelines

### Tone and Style

**Do:**
- Write for end users (operators, SREs, developers)
- Use clear, concise language
- Explain "why" not just "what"
- Include examples and code snippets
- Link to detailed documentation
- Highlight breaking changes prominently

**Don't:**
- Use internal jargon or acronyms without explanation
- Assume deep technical knowledge
- Bury important information
- Write overly technical implementation details
- Skip upgrade instructions

### Structure

**Order of Sections:**
1. Header (version, date, type)
2. Overview (2-3 sentences)
3. What's New (new features)
4. Improvements
5. Bug Fixes
6. Security
7. Deprecations
8. Breaking Changes
9. Upgrade Instructions
10. Known Issues
11. Resources/Links

### Formatting

**Headers:**
- Use H1 for release title
- Use H2 for major sections
- Use H3 for subsections

**Code Blocks:**
- Always use syntax highlighting (```yaml, ```bash)
- Include comments for clarity
- Show before/after examples

**Links:**
- Link to GitHub issues: [#123](https://github.com/kyklos/kyklos/issues/123)
- Link to PRs: [PR #456](https://github.com/kyklos/kyklos/pull/456)
- Link to docs: [Feature Guide](https://kyklos.io/docs/feature)

---

## Examples

### Example 1: Feature Introduction

**Good:**
```markdown
### StatefulSet Target Support

Kyklos now scales StatefulSets in addition to Deployments. StatefulSets are scaled using the same time window logic, with special handling for ordered pod creation and deletion.

**Configuration Example:**
```yaml
spec:
  targetRef:
    kind: StatefulSet
    name: postgres
```

**Documentation:** [StatefulSet Scaling Guide](https://kyklos.io/docs/statefulset)
```

**Bad:**
```markdown
### StatefulSet Support

Added StatefulSet scaling. Use `kind: StatefulSet` in targetRef.
```

**Why bad:** No context, no example, no link to docs.

---

### Example 2: Bug Fix

**Good:**
```markdown
### Fixed: Cross-Midnight Window Calculation Error

**Issue:** Windows spanning midnight (e.g., 22:00-02:00) incorrectly calculated active periods in timezones with DST.

**Impact:** Deployments may not scale at expected times.

**Fix:** Corrected timezone offset handling.

**Affected Versions:** v0.1.0 - v0.1.2

**Issue:** [#456](https://github.com/kyklos/kyklos/issues/456)
```

**Bad:**
```markdown
- Fixed window calculation bug (#456)
```

**Why bad:** No explanation of impact or affected versions.

---

### Example 3: Breaking Change

**Good:**
```markdown
### BREAKING CHANGE: defaultReplicas Now Required

**What Changed:**
The `defaultReplicas` field is now required in TimeWindowScaler spec. Previously, it defaulted to 0 if omitted.

**Why:**
Explicit declaration prevents accidental scale-to-zero scenarios.

**Migration:**
Add `defaultReplicas: 0` to all existing TimeWindowScaler resources:

```yaml
spec:
  defaultReplicas: 0  # Add this line
  timezone: UTC
  windows: []
```

**Automation:**
```bash
kubectl get tws --all-namespaces -o json | \
  jq '.items[] | select(.spec.defaultReplicas == null)' | \
  # Add defaultReplicas: 0 to each
```

**Timeline:**
- v0.3.0: Field required (breaking change)
- Migration window: None (must update before upgrading)
```

**Bad:**
```markdown
### Breaking Changes
- defaultReplicas is now required
```

**Why bad:** No migration instructions or context.

---

## Document Version

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-28 | ci-release-engineer | Initial release notes template |

## Related Documents

- [RELEASE-POLICY.md](/Users/aykumar/personal/kyklos/docs/release/RELEASE-POLICY.md) - Release management policy
- [../ci/PIPELINE.md](/Users/aykumar/personal/kyklos/docs/ci/PIPELINE.md) - CI/CD pipeline design
