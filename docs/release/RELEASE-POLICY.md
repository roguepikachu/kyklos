# Release Management Policy

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** ci-release-engineer

This document defines the release process, versioning strategy, supported versions, and deprecation policy for Kyklos. It provides clear expectations for users and guidelines for maintainers.

---

## Table of Contents

1. [Versioning Strategy](#versioning-strategy)
2. [Release Cadence](#release-cadence)
3. [Release Process](#release-process)
4. [Supported Versions](#supported-versions)
5. [Kubernetes Compatibility Matrix](#kubernetes-compatibility-matrix)
6. [Changelog Management](#changelog-management)
7. [Security and CVE Policy](#security-and-cve-policy)
8. [Deprecation Policy](#deprecation-policy)
9. [Rollback Procedures](#rollback-procedures)

---

## Versioning Strategy

### Semantic Versioning 2.0.0

Kyklos follows [Semantic Versioning](https://semver.org/) with the format: **MAJOR.MINOR.PATCH**

```
v0.1.0
│ │ │
│ │ └─ PATCH: Backward-compatible bug fixes
│ └─── MINOR: Backward-compatible features
└───── MAJOR: Breaking changes
```

### Version Components

**MAJOR Version (X.0.0)**
- Breaking API changes (CRD schema breaking changes)
- Removal of deprecated features
- Incompatible behavior changes
- Supported Kubernetes version changes (major drops)

**Example:** v0.1.0 → v1.0.0
- Introduce webhook validation (admission control changes)
- Remove deprecated status fields
- Change default behavior for grace periods

**MINOR Version (0.X.0)**
- New backward-compatible features
- New CRD fields with defaults
- Performance improvements
- Dependency updates

**Example:** v0.1.0 → v0.2.0
- Add support for StatefulSet targets (new feature)
- Add `.spec.maxScaleDownRate` field (optional)
- Improve reconcile performance (no API changes)

**PATCH Version (0.0.X)**
- Backward-compatible bug fixes
- Security patches
- Documentation corrections
- Minor performance improvements

**Example:** v0.1.0 → v0.1.1
- Fix: Cross-midnight window calculation error
- Fix: Memory leak in reconcile loop
- Security: Update vulnerable dependency

### Pre-Release Versions

**Alpha (v0.1.0-alpha.1)**
- Early development, incomplete features
- May have breaking changes between alphas
- Not recommended for production
- Example: v0.1.0-alpha.1, v0.1.0-alpha.2

**Beta (v0.1.0-beta.1)**
- Feature-complete for that minor version
- Stabilization phase, API locked
- Suitable for testing in non-production
- Example: v0.1.0-beta.1, v0.1.0-beta.2

**Release Candidate (v0.1.0-rc.1)**
- Final testing before stable release
- No new features, only critical bug fixes
- Production-ready but not battle-tested
- Example: v0.1.0-rc.1, v0.1.0-rc.2

### Version 0.x Special Rules

**During 0.x series (pre-1.0):**
- MINOR version increments MAY include breaking changes
- Users should read release notes carefully
- API stability not guaranteed until 1.0
- Backward compatibility is best-effort

**Current status:** v0.1.x (alpha stage)

---

## Release Cadence

### Planned Release Schedule

| Release Type | Frequency | Example Dates |
|-------------|-----------|---------------|
| MAJOR | Annually | Q1 each year |
| MINOR | Quarterly | Q1, Q2, Q3, Q4 |
| PATCH | As needed | Within 1-2 weeks of issue |
| SECURITY | Immediately | Within 24-48 hours of CVE |

### v0.1.x Timeline (Example)

| Version | Target Date | Status | Notes |
|---------|------------|--------|-------|
| v0.1.0-alpha.1 | 2025-11-01 | Planned | Initial alpha release |
| v0.1.0-beta.1 | 2025-11-15 | Planned | Feature freeze, testing |
| v0.1.0 | 2025-12-01 | Planned | First stable release |
| v0.1.1 | As needed | Planned | Bug fix patch |
| v0.2.0 | 2026-03-01 | Planned | StatefulSet support |

### Hotfix Releases

**When to release hotfix:**
- Critical security vulnerability (CVE with CVSS >= 7.0)
- Data loss bug
- Cluster disruption issue
- Panic/crash loop in controller

**Hotfix process:**
- Branch from affected release tag (e.g., v0.1.0)
- Apply minimal fix
- Release as patch version (e.g., v0.1.1)
- Backport to supported versions
- Expedited release (skip beta/RC phases)

---

## Release Process

### Phase 1: Planning (T-4 weeks)

**Responsibilities:**
- Product Owner: Define feature scope
- Engineering Lead: Review roadmap
- CI Engineer: Ensure CI/CD ready

**Activities:**
1. Create release milestone in GitHub
2. Assign issues to milestone
3. Review and update ROADMAP.md
4. Communicate planned features to users

**Success Criteria:**
- Milestone created with target date
- All features have implementation issues
- Roadmap updated

---

### Phase 2: Development (T-4 weeks to T-1 week)

**Responsibilities:**
- Developers: Implement features
- Reviewers: Review and approve PRs
- CI Engineer: Monitor build health

**Activities:**
1. Implement features per milestone
2. Merge PRs to main branch
3. Continuous integration validation
4. Update documentation

**Quality Gates:**
- All tests pass (unit, envtest, e2e)
- Code coverage >= 80%
- No HIGH/CRITICAL CVEs
- All features have docs

---

### Phase 3: Feature Freeze (T-1 week)

**Announcement:**
```
Feature Freeze for v0.2.0

Effective: 2026-02-15
Stable Release: 2026-03-01

Status:
- New features: BLOCKED
- Bug fixes: ALLOWED
- Documentation: ALLOWED
- Dependency updates: ALLOWED (non-breaking only)

Branch: release/v0.2
```

**Activities:**
1. Create release branch: `release/v0.2`
2. Tag beta: `v0.2.0-beta.1`
3. Deploy to staging environment
4. Run extended test suite
5. Solicit community testing

**Allowed Changes:**
- Bug fixes only
- Documentation improvements
- Test stability improvements
- Translation updates

**Blocked Changes:**
- New features
- Refactoring
- Breaking changes
- Major dependency updates

---

### Phase 4: Release Candidate (T-3 days)

**Activities:**
1. Address all P0/P1 bugs from beta testing
2. Tag release candidate: `v0.2.0-rc.1`
3. Deploy RC to production-like environment
4. Final smoke tests and E2E validation
5. Generate preliminary changelog

**Go/No-Go Checklist:**
- [ ] All P0 bugs fixed
- [ ] All P1 bugs fixed or documented
- [ ] E2E tests pass (10 consecutive runs)
- [ ] Security scan clean (zero HIGH/CRITICAL)
- [ ] Documentation complete
- [ ] CHANGELOG.md updated
- [ ] Release notes drafted
- [ ] Installation tested on supported K8s versions
- [ ] Rollback procedure documented

**Decision:** Release Engineer approves go-live

---

### Phase 5: Release (T-0)

**Activities:**

**1. Tag Release**
```bash
git checkout release/v0.2
git tag -a v0.2.0 -m "Release v0.2.0: StatefulSet Support"
git push origin v0.2.0
```

**2. Automated Release Workflow Triggers**
- Multi-arch image build and push
- Security scanning
- E2E test suite execution
- Release artifact publication
- Changelog generation

**3. Manual Verification**
- Verify images pushed to ghcr.io
- Verify GitHub release created
- Download and verify install.yaml
- Test installation: `kubectl apply -f install.yaml`

**4. Documentation Updates**
- Update README.md version badge
- Update installation docs with new version
- Publish release notes to website
- Update API reference docs

**5. Announcement**
- Post release announcement to GitHub Discussions
- Update Slack/Discord channels
- Tweet release summary
- Email mailing list

**6. Merge Back to Main**
```bash
git checkout main
git merge release/v0.2
git push origin main
```

---

### Phase 6: Post-Release Monitoring (T+1 week)

**Activities:**
1. Monitor GitHub issues for new bug reports
2. Watch for security advisories
3. Collect user feedback
4. Track adoption metrics

**Metrics to Track:**
- Image pull counts
- GitHub release downloads
- Issue reports (new vs. closed)
- Community feedback sentiment

**Success Criteria:**
- No critical bugs reported within 72 hours
- Security scan remains clean
- Installation documentation confirmed accurate
- Positive community feedback

---

## Supported Versions

### Support Matrix

| Kyklos Version | Release Date | End of Support | Support Type |
|---------------|--------------|----------------|--------------|
| v0.1.x | 2025-12-01 | 2026-06-01 | Full support |
| v0.2.x | 2026-03-01 | 2026-09-01 | Full support |
| v1.0.x | TBD | TBD | Full support |

### Support Types

**Full Support:**
- Security patches backported
- Critical bug fixes backported
- Dependency updates (security only)
- Community support via GitHub

**Maintenance Support:**
- Security patches only (CVSS >= 7.0)
- No bug fixes unless critical
- Limited community support

**End of Life:**
- No patches or updates
- Repository archived or deprecated
- Users encouraged to upgrade

### Support Duration

**General Policy:**
- **Minor versions:** Supported for 6 months after release
- **Major versions:** Supported for 12 months after next major
- **Pre-1.0 versions:** Support until 1.0 released + 3 months

**Example:**
- v0.1.0 released: 2025-12-01
- v0.2.0 released: 2026-03-01
- v0.1.x end of support: 2026-06-01 (6 months after v0.1.0)

---

## Kubernetes Compatibility Matrix

### Tested and Supported Versions

| Kyklos Version | Kubernetes Versions | Notes |
|---------------|--------------------|------------------------------------|
| v0.1.x | 1.25, 1.26, 1.27, 1.28 | Tested on Kind, EKS, GKE |
| v0.2.x | 1.26, 1.27, 1.28, 1.29 | Drops 1.25 (EOL) |
| v1.0.x | 1.27, 1.28, 1.29, 1.30 | Stable API, minimum 1.27 |

### Support Policy for Kubernetes Versions

**Supported Kubernetes Versions:**
- Current Kubernetes version
- Two previous minor versions
- Aligned with Kubernetes version skew policy

**Example (November 2025):**
- Latest Kubernetes: 1.28
- Supported: 1.26, 1.27, 1.28
- Tested: 1.25, 1.26, 1.27, 1.28

**When Kubernetes version drops out of support:**
- Announce deprecation in previous Kyklos minor release
- Remove from test matrix in next minor release
- Document in release notes

**Untested Versions:**
- May work but not officially supported
- Community contributions welcome
- Report issues with `kubernetes-version` label

---

## Changelog Management

### Commit Message Convention

Kyklos uses [Conventional Commits](https://www.conventionalcommits.org/) for automated changelog generation.

**Format:**
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring (no behavior change)
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks (dependencies, build)
- `ci`: CI/CD changes
- `revert`: Revert previous commit

**Scopes (optional):**
- `api`: CRD or API changes
- `controller`: Reconcile loop changes
- `webhook`: Validation webhook
- `metrics`: Prometheus metrics
- `docs`: Documentation
- `deps`: Dependencies

**Examples:**
```
feat(api): add maxScaleDownRate field to TimeWindowScaler

Introduces a new optional field to limit the rate of scale-down
operations, preventing sudden traffic drops.

Closes #123
```

```
fix(controller): correct cross-midnight window calculation

Fixed an edge case where windows spanning midnight were incorrectly
evaluated due to timezone offset errors.

Fixes #456
```

```
docs: update installation guide with Helm instructions

Added Helm chart installation method as alternative to raw YAML.

Co-authored-by: John Doe <john@example.com>
```

### Breaking Changes

**Marking breaking changes:**
```
feat(api)!: change defaultReplicas behavior to required field

BREAKING CHANGE: The defaultReplicas field is now required in the
TimeWindowScaler spec. Previously, it defaulted to 0 if omitted.

Migration: Add "defaultReplicas: 0" to all existing TimeWindowScaler
resources before upgrading.

Closes #789
```

**Note:** `!` after type indicates breaking change

### Changelog Sections

**Generated CHANGELOG.md structure:**
```markdown
## [0.2.0] - 2026-03-01

### Added
- feat(api): StatefulSet target support
- feat(controller): maxScaleDownRate field for gradual downscaling

### Changed
- perf(controller): optimize reconcile loop with caching (30% faster)
- chore(deps): update controller-runtime to v0.17.0

### Fixed
- fix(controller): cross-midnight window calculation edge case
- fix(webhook): validation for overlapping windows

### Security
- fix(deps): update Go to 1.21.5 (CVE-2024-XXXXX)

### Deprecated
- api: `.status.lastUpdateTime` field (use `.status.lastTransitionTime`)

### Removed
- None

### Breaking Changes
- None
```

### Manual Changelog Curation

**Before release, maintainers should:**
1. Review auto-generated changelog
2. Combine duplicate entries
3. Add context where needed
4. Highlight user-facing changes
5. Link to documentation for new features

---

## Security and CVE Policy

### Security Release Process

**1. Vulnerability Reported**
- Security team receives report via security@kyklos.io
- Acknowledge within 24 hours
- Assess severity (use CVSS calculator)

**2. Assessment**
- CVSS >= 9.0: CRITICAL
- CVSS >= 7.0: HIGH
- CVSS >= 4.0: MEDIUM
- CVSS < 4.0: LOW

**3. Remediation Timeline**

| Severity | Patch Release | Public Disclosure |
|----------|---------------|-------------------|
| CRITICAL | 24-48 hours | After patch available |
| HIGH | 1 week | After patch available |
| MEDIUM | 2-4 weeks | After patch available |
| LOW | Next minor release | After patch available |

**4. Coordinated Disclosure**
- Notify affected users privately (if possible)
- Prepare patch
- Embargo period: 7 days for HIGH/CRITICAL
- Public disclosure with CVE ID

**5. Security Advisory**
```markdown
# Security Advisory: CVE-2026-XXXXX

**Severity:** HIGH (CVSS 8.5)

**Affected Versions:** v0.1.0 - v0.1.5

**Fixed Versions:** v0.1.6, v0.2.1

**Summary:**
The controller allows unauthorized modification of target deployments
due to insufficient RBAC validation.

**Impact:**
An attacker with access to create TimeWindowScaler resources can
scale arbitrary deployments in the cluster.

**Remediation:**
Upgrade to v0.1.6 or v0.2.1 immediately. Alternatively, restrict
TimeWindowScaler creation via RBAC.

**Credit:** Reported by John Doe (Company XYZ)
```

### CVE Assignment

**When to request CVE:**
- Security vulnerability in Kyklos code
- CVSS >= 4.0 (MEDIUM or higher)
- Affects released versions

**Process:**
1. Request CVE from GitHub Security Advisory
2. Receive CVE ID within 72 hours
3. Include CVE in release notes and advisory

---

## Deprecation Policy

### Deprecation Timeline

**Standard deprecation lifecycle:**
1. **N release:** Feature marked deprecated, deprecation warning added
2. **N+1 release:** Deprecation warning continues, migration guide published
3. **N+2 release:** Feature removed

**Example:**
- v0.1.0: Feature X working normally
- v0.2.0: Feature X marked deprecated, warning in logs
- v0.3.0: Deprecation warning continues
- v0.4.0: Feature X removed

### Deprecation Notice Format

**In Code (Controller Logs):**
```go
log.Info("DEPRECATED: .spec.oldField is deprecated and will be removed in v0.4.0. Use .spec.newField instead.")
```

**In CRD (Validation Warning):**
```yaml
properties:
  oldField:
    type: string
    description: |
      DEPRECATED: This field is deprecated and will be removed in v0.4.0.
      Use newField instead. See migration guide: https://kyklos.io/migrate
```

**In Documentation:**
```markdown
## Deprecated Features

| Feature | Deprecated In | Removed In | Replacement |
|---------|--------------|-----------|------------|
| .spec.oldField | v0.2.0 | v0.4.0 | .spec.newField |
| Holiday source: ConfigMap | v0.3.0 | v0.5.0 | Holiday source: API |
```

### Migration Guides

**Provide clear migration path:**
1. **Why:** Explain reason for deprecation
2. **What:** Describe replacement feature
3. **How:** Step-by-step migration instructions
4. **When:** Timeline for removal
5. **Help:** Link to support channels

**Example:**
```markdown
# Migrating from .spec.oldField to .spec.newField

## Why?
The oldField format is ambiguous and error-prone. newField provides
stronger type safety and better validation.

## What's Changing?
- oldField: string format (e.g., "2 hours")
- newField: duration format (e.g., "2h")

## Migration Steps

1. Update your TimeWindowScaler YAML:

**Before:**
```yaml
spec:
  oldField: "30 minutes"
```

**After:**
```yaml
spec:
  newField: "30m"
```

2. Apply updated resource:
```bash
kubectl apply -f updated-tws.yaml
```

3. Verify behavior unchanged:
```bash
kubectl get tws <name> -o yaml | grep effectiveGracePeriod
```

## Timeline
- v0.2.0: oldField deprecated (still works)
- v0.4.0: oldField removed (breaking change)

## Need Help?
- GitHub Discussions: https://github.com/kyklos/kyklos/discussions
- Slack: #kyklos-users
```

---

## Rollback Procedures

### When to Rollback

**Rollback triggers:**
- Critical bug discovered post-release
- Regression in core functionality
- Performance degradation > 50%
- Security vulnerability with no immediate patch

### Rollback Decision Matrix

| Severity | Time Since Release | Action |
|----------|-------------------|--------|
| Critical | < 24 hours | Immediate rollback + hotfix |
| Critical | > 24 hours | Hotfix patch release |
| High | < 7 days | Hotfix patch release |
| High | > 7 days | Include in next patch |
| Medium/Low | Any | Include in next minor |

### Rollback Process

**1. Assess Impact**
```bash
# Check adoption
docker pull ghcr.io/kyklos/controller:v0.2.0
# Review image pull metrics

# Check issue reports
gh issue list --label bug --label v0.2.0
```

**2. Prepare Rollback**
```bash
# Tag previous stable version as rollback target
git tag rollback/v0.2.0-to-v0.1.5 v0.1.5
git push origin rollback/v0.2.0-to-v0.1.5
```

**3. Communicate**
```markdown
# Urgent: Rollback Recommendation for v0.2.0

**Issue:** Critical regression in cross-midnight window handling

**Impact:** Deployments may not scale correctly between 23:00-01:00

**Recommendation:** Rollback to v0.1.5

**Rollback Steps:**
1. Update installation YAML:
   ```bash
   kubectl set image deployment/kyklos-controller \
     kyklos-controller=ghcr.io/kyklos/controller:v0.1.5 \
     -n kyklos-system
   ```

2. Verify controller version:
   ```bash
   kubectl get deployment kyklos-controller -n kyklos-system \
     -o jsonpath='{.spec.template.spec.containers[0].image}'
   ```

**Fix:** Hotfix release v0.2.1 planned for 2026-03-05

**Status:** https://github.com/kyklos/kyklos/issues/999
```

**4. Hotfix Release**
- Follow expedited release process
- Skip beta/RC phases for critical fixes
- Target release within 24-48 hours

**5. Post-Mortem**
- Document what went wrong
- Identify gaps in testing
- Update test suite to catch regression
- Review release process for improvements

---

## Release Checklist Template

### Pre-Release Checklist

**Code Quality:**
- [ ] All milestone issues closed or deferred
- [ ] All PRs merged to release branch
- [ ] Code coverage >= 80%
- [ ] Zero HIGH/CRITICAL vulnerabilities
- [ ] No flaky tests in last 100 CI runs

**Testing:**
- [ ] Unit tests pass (100% pass rate)
- [ ] Envtest passes (100% pass rate)
- [ ] E2E smoke tests pass (10 consecutive runs)
- [ ] E2E full suite passes (all scenarios)
- [ ] Manual smoke test on Kind cluster
- [ ] Manual smoke test on cloud provider (EKS/GKE/AKS)

**Documentation:**
- [ ] CHANGELOG.md updated
- [ ] Release notes drafted (RELEASE-NOTES-v0.X.0.md)
- [ ] API reference updated (if API changes)
- [ ] Installation docs updated with new version
- [ ] Migration guide written (if breaking changes)
- [ ] README.md version badge updated

**Build Artifacts:**
- [ ] Multi-arch image built (amd64, arm64)
- [ ] Images pushed to ghcr.io
- [ ] SBOM generated and attached
- [ ] install.yaml generated
- [ ] Checksums generated

**Governance:**
- [ ] Release notes reviewed by Product Owner
- [ ] Security scan reviewed by Security Lead
- [ ] Go/No-Go meeting held
- [ ] Release approved by Release Engineer

### Post-Release Checklist

**Verification:**
- [ ] GitHub release created with assets
- [ ] Images pullable from ghcr.io
- [ ] install.yaml tested on clean cluster
- [ ] Helm chart updated (if applicable)
- [ ] Documentation site updated

**Communication:**
- [ ] Release announcement posted (GitHub Discussions)
- [ ] Social media updated (Twitter, Mastodon)
- [ ] Mailing list notified
- [ ] Slack/Discord announcement
- [ ] CNCF Slack announcement (if applicable)

**Monitoring:**
- [ ] Image pull metrics tracked
- [ ] Issue reports monitored (next 72 hours)
- [ ] Security advisories monitored
- [ ] Community feedback collected

**Cleanup:**
- [ ] Release branch merged back to main
- [ ] Milestone closed in GitHub
- [ ] Next milestone created
- [ ] Post-release retrospective scheduled

---

## Document Version

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-28 | ci-release-engineer | Initial release policy for v0.1 |

## Related Documents

- [RELEASE-NOTES-TEMPLATE.md](/Users/aykumar/personal/kyklos/docs/release/RELEASE-NOTES-TEMPLATE.md) - Release notes format
- [REGISTRY-MAP.md](/Users/aykumar/personal/kyklos/docs/release/REGISTRY-MAP.md) - Container registry strategy
- [../ci/PIPELINE.md](/Users/aykumar/personal/kyklos/docs/ci/PIPELINE.md) - CI/CD pipeline design
- [../ROADMAP.md](/Users/aykumar/personal/kyklos/docs/ROADMAP.md) - Product roadmap
