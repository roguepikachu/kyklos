# README Badges

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** ci-release-engineer

This document provides README badge specifications for Kyklos, including Markdown snippets for build status, coverage, release version, license, and security metrics.

---

## Table of Contents

1. [Badge Overview](#badge-overview)
2. [Recommended Badge Layout](#recommended-badge-layout)
3. [Build Status Badges](#build-status-badges)
4. [Code Quality Badges](#code-quality-badges)
5. [Release and Version Badges](#release-and-version-badges)
6. [Security Badges](#security-badges)
7. [Community Badges](#community-badges)
8. [Deployment Badges](#deployment-badges)
9. [Complete Examples](#complete-examples)

---

## Badge Overview

Badges provide at-a-glance project health indicators in the README. They should be:

**Characteristics:**
- **Visible:** Prominent placement in README header
- **Accurate:** Automated from CI/CD, not manually maintained
- **Relevant:** Show critical project health metrics
- **Actionable:** Link to detailed information

**Badge Service:** [shields.io](https://shields.io) (standard for GitHub projects)

---

## Recommended Badge Layout

### Primary Badges (Always Show)

Place immediately after project title and description:

```markdown
# Kyklos Time Window Scaler

[![CI Status](https://img.shields.io/github/actions/workflow/status/kyklos/kyklos/ci.yml?branch=main&label=CI&logo=github)](https://github.com/kyklos/kyklos/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kyklos/kyklos?label=Release&logo=github)](https://github.com/kyklos/kyklos/releases/latest)
[![License](https://img.shields.io/github/license/kyklos/kyklos?label=License)](https://github.com/kyklos/kyklos/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyklos/kyklos)](https://goreportcard.com/report/github.com/kyklos/kyklos)
```

**Rendered:**

[![CI Status](https://img.shields.io/github/actions/workflow/status/kyklos/kyklos/ci.yml?branch=main&label=CI&logo=github)](https://github.com/kyklos/kyklos/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kyklos/kyklos?label=Release&logo=github)](https://github.com/kyklos/kyklos/releases/latest)
[![License](https://img.shields.io/github/license/kyklos/kyklos?label=License)](https://github.com/kyklos/kyklos/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyklos/kyklos)](https://goreportcard.com/report/github.com/kyklos/kyklos)

---

### Secondary Badges (Optional, Based on Context)

Place in a second row or "Badges" section:

```markdown
[![Coverage](https://img.shields.io/codecov/c/github/kyklos/kyklos?label=Coverage&logo=codecov)](https://codecov.io/gh/kyklos/kyklos)
[![Security](https://img.shields.io/badge/Security-Trivy%20Scan-blue?logo=aquasecuritytrivy)](https://github.com/kyklos/kyklos/security)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/kyklos/kyklos/badge)](https://securityscorecards.dev/viewer/?uri=github.com/kyklos/kyklos)
[![Docker Pulls](https://img.shields.io/docker/pulls/kyklos/kyklos?label=Docker%20Pulls&logo=docker)](https://hub.docker.com/r/kyklos/kyklos)
```

---

## Build Status Badges

### GitHub Actions CI Badge

**Badge:**
```markdown
[![CI Status](https://img.shields.io/github/actions/workflow/status/kyklos/kyklos/ci.yml?branch=main&label=CI&logo=github)](https://github.com/kyklos/kyklos/actions/workflows/ci.yml)
```

**Rendered:**
![CI Status](https://img.shields.io/github/actions/workflow/status/kyklos/kyklos/ci.yml?branch=main&label=CI&logo=github)

**Customization:**
```markdown
# Specific branch
?branch=release/v0.1

# Custom label
&label=Build%20Status

# Different logo
&logo=githubactions
```

**States:**
- ![passing](https://img.shields.io/badge/build-passing-brightgreen) - All jobs passed
- ![failing](https://img.shields.io/badge/build-failing-red) - One or more jobs failed
- ![no status](https://img.shields.io/badge/build-no%20status-lightgrey) - No recent runs

---

### Release Workflow Badge

**Badge:**
```markdown
[![Release Status](https://img.shields.io/github/actions/workflow/status/kyklos/kyklos/release.yml?label=Release&logo=github&event=release)](https://github.com/kyklos/kyklos/actions/workflows/release.yml)
```

**Rendered:**
![Release Status](https://img.shields.io/github/actions/workflow/status/kyklos/kyklos/release.yml?label=Release&logo=github&event=release)

**Use Case:** Show release pipeline health separately from CI

---

## Code Quality Badges

### Go Report Card

**Badge:**
```markdown
[![Go Report Card](https://goreportcard.com/badge/github.com/kyklos/kyklos)](https://goreportcard.com/report/github.com/kyklos/kyklos)
```

**Rendered:**
![Go Report Card](https://goreportcard.com/badge/github.com/kyklos/kyklos)

**What It Checks:**
- gofmt formatting
- go vet issues
- gocyclo complexity
- golint warnings
- ineffassign
- misspell

**Grades:**
- A+ (96-100%)
- A (90-95%)
- B (80-89%)
- C (70-79%)
- D (60-69%)
- F (< 60%)

**Update Frequency:** On-demand via goreportcard.com

---

### Code Coverage (Codecov)

**Badge:**
```markdown
[![Coverage](https://img.shields.io/codecov/c/github/kyklos/kyklos?label=Coverage&logo=codecov)](https://codecov.io/gh/kyklos/kyklos)
```

**Rendered:**
![Coverage](https://img.shields.io/codecov/c/github/kyklos/kyklos?label=Coverage&logo=codecov)

**Setup:**
1. Sign up at [codecov.io](https://codecov.io)
2. Link GitHub repository
3. Upload coverage in CI:
   ```yaml
   - uses: codecov/codecov-action@v3
     with:
       files: ./coverage.out
   ```

**Color Coding:**
- ![>80%](https://img.shields.io/badge/coverage-85%25-brightgreen) - Excellent (>80%)
- ![60-80%](https://img.shields.io/badge/coverage-75%25-yellow) - Good (60-80%)
- ![<60%](https://img.shields.io/badge/coverage-45%25-red) - Needs Improvement (<60%)

---

### Code Coverage (Alternative: Simple Badge)

**Badge:**
```markdown
[![Coverage](https://img.shields.io/badge/coverage-85%25-brightgreen?logo=go)](https://github.com/kyklos/kyklos/actions)
```

**Rendered:**
![Coverage](https://img.shields.io/badge/coverage-85%25-brightgreen?logo=go)

**Manual Update:** Update percentage in badge URL after each coverage run

**Dynamic Update (GitHub Actions):**
```yaml
- name: Update coverage badge
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo "![Coverage](https://img.shields.io/badge/coverage-${COVERAGE}-brightgreen)" > coverage-badge.md
```

---

## Release and Version Badges

### Latest Release Version

**Badge:**
```markdown
[![Release](https://img.shields.io/github/v/release/kyklos/kyklos?label=Release&logo=github)](https://github.com/kyklos/kyklos/releases/latest)
```

**Rendered:**
![Release](https://img.shields.io/github/v/release/kyklos/kyklos?label=Release&logo=github)

**Options:**
```markdown
# Include pre-releases
?include_prereleases

# Sort by semantic version
&sort=semver

# Filter by prefix
&filter=v*
```

---

### Container Image Version

**Badge (GitHub Container Registry):**
```markdown
[![Image](https://img.shields.io/badge/ghcr.io-v0.1.0-blue?logo=docker)](https://ghcr.io/kyklos/kyklos)
```

**Rendered:**
![Image](https://img.shields.io/badge/ghcr.io-v0.1.0-blue?logo=docker)

**Badge (Docker Hub, if used):**
```markdown
[![Docker Image](https://img.shields.io/docker/v/kyklos/kyklos?label=Docker&logo=docker)](https://hub.docker.com/r/kyklos/kyklos)
```

---

### Release Date

**Badge:**
```markdown
[![Release Date](https://img.shields.io/github/release-date/kyklos/kyklos?label=Released&logo=github)](https://github.com/kyklos/kyklos/releases)
```

**Rendered:**
![Release Date](https://img.shields.io/github/release-date/kyklos/kyklos?label=Released&logo=github)

---

### Download Count

**Badge:**
```markdown
[![Downloads](https://img.shields.io/github/downloads/kyklos/kyklos/total?label=Downloads&logo=github)](https://github.com/kyklos/kyklos/releases)
```

**Rendered:**
![Downloads](https://img.shields.io/github/downloads/kyklos/kyklos/total?label=Downloads&logo=github)

**Specific Release:**
```markdown
[![v0.1.0 Downloads](https://img.shields.io/github/downloads/kyklos/kyklos/v0.1.0/total?label=v0.1.0%20Downloads)](https://github.com/kyklos/kyklos/releases/tag/v0.1.0)
```

---

## Security Badges

### Security Policy

**Badge:**
```markdown
[![Security Policy](https://img.shields.io/badge/Security-Policy-blue?logo=shield)](https://github.com/kyklos/kyklos/security/policy)
```

**Rendered:**
![Security Policy](https://img.shields.io/badge/Security-Policy-blue?logo=shield)

**Links to:** SECURITY.md file or GitHub security policy

---

### Vulnerability Scanning (Trivy)

**Badge:**
```markdown
[![Trivy Scan](https://img.shields.io/badge/Trivy-Scan%20Clean-brightgreen?logo=aquasecuritytrivy)](https://github.com/kyklos/kyklos/security)
```

**Rendered:**
![Trivy Scan](https://img.shields.io/badge/Trivy-Scan%20Clean-brightgreen?logo=aquasecuritytrivy)

**Manual Update:** Update status after each scan

**States:**
- ![clean](https://img.shields.io/badge/Trivy-Scan%20Clean-brightgreen?logo=aquasecuritytrivy) - Zero HIGH/CRITICAL
- ![warnings](https://img.shields.io/badge/Trivy-Warnings-yellow?logo=aquasecuritytrivy) - Some MEDIUM
- ![critical](https://img.shields.io/badge/Trivy-Critical%20Issues-red?logo=aquasecuritytrivy) - HIGH/CRITICAL found

---

### OpenSSF Best Practices Badge

**Badge:**
```markdown
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/XXXX/badge)](https://www.bestpractices.dev/projects/XXXX)
```

**Setup:**
1. Visit [bestpractices.dev](https://www.bestpractices.dev)
2. Add project and answer checklist
3. Receive project ID (replace XXXX)

**Levels:**
- ![passing](https://img.shields.io/badge/OpenSSF-Passing-brightgreen) - Meets all criteria
- ![silver](https://img.shields.io/badge/OpenSSF-Silver-silver) - Advanced practices
- ![gold](https://img.shields.io/badge/OpenSSF-Gold-gold) - Exemplary practices

---

### OpenSSF Scorecard

**Badge:**
```markdown
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/kyklos/kyklos/badge)](https://securityscorecards.dev/viewer/?uri=github.com/kyklos/kyklos)
```

**What It Checks:**
- Branch protection
- Code review
- Signed commits
- Dependency updates
- CI tests
- SAST tools
- Pinned dependencies

**Score Range:** 0-10 (higher is better)

---

## Community Badges

### License

**Badge:**
```markdown
[![License](https://img.shields.io/github/license/kyklos/kyklos?label=License)](https://github.com/kyklos/kyklos/blob/main/LICENSE)
```

**Rendered:**
![License](https://img.shields.io/github/license/kyklos/kyklos?label=License)

**Common Licenses:**
- ![Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue)
- ![MIT](https://img.shields.io/badge/License-MIT-yellow)
- ![GPL v3](https://img.shields.io/badge/License-GPLv3-blue)

---

### Contributors

**Badge:**
```markdown
[![Contributors](https://img.shields.io/github/contributors/kyklos/kyklos?label=Contributors)](https://github.com/kyklos/kyklos/graphs/contributors)
```

**Rendered:**
![Contributors](https://img.shields.io/github/contributors/kyklos/kyklos?label=Contributors)

---

### GitHub Stars

**Badge:**
```markdown
[![Stars](https://img.shields.io/github/stars/kyklos/kyklos?style=social)](https://github.com/kyklos/kyklos/stargazers)
```

**Rendered:**
![Stars](https://img.shields.io/github/stars/kyklos/kyklos?style=social)

**Note:** Use `?style=social` for star count with icon

---

### Slack/Discord Community

**Badge:**
```markdown
[![Slack](https://img.shields.io/badge/Slack-Join%20Community-4A154B?logo=slack)](https://kyklos-community.slack.com)
```

**Rendered:**
![Slack](https://img.shields.io/badge/Slack-Join%20Community-4A154B?logo=slack)

**Discord Alternative:**
```markdown
[![Discord](https://img.shields.io/discord/123456789?label=Discord&logo=discord)](https://discord.gg/kyklos)
```

---

## Deployment Badges

### Kubernetes Version Compatibility

**Badge:**
```markdown
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.26+-326CE5?logo=kubernetes)](https://kubernetes.io)
```

**Rendered:**
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.26+-326CE5?logo=kubernetes)

**Specific Versions:**
```markdown
[![K8s 1.28](https://img.shields.io/badge/K8s-1.28-326CE5?logo=kubernetes)](https://kubernetes.io)
[![K8s 1.27](https://img.shields.io/badge/K8s-1.27-326CE5?logo=kubernetes)](https://kubernetes.io)
[![K8s 1.26](https://img.shields.io/badge/K8s-1.26-326CE5?logo=kubernetes)](https://kubernetes.io)
```

---

### Go Version

**Badge:**
```markdown
[![Go Version](https://img.shields.io/github/go-mod/go-version/kyklos/kyklos?label=Go&logo=go)](https://go.dev)
```

**Rendered:**
![Go Version](https://img.shields.io/github/go-mod/go-version/kyklos/kyklos?label=Go&logo=go)

**Reads from:** `go.mod` file automatically

---

### Docker Pulls (if using Docker Hub)

**Badge:**
```markdown
[![Docker Pulls](https://img.shields.io/docker/pulls/kyklos/kyklos?label=Docker%20Pulls&logo=docker)](https://hub.docker.com/r/kyklos/kyklos)
```

**Rendered:**
![Docker Pulls](https://img.shields.io/docker/pulls/kyklos/kyklos?label=Docker%20Pulls&logo=docker)

---

### Image Size

**Badge:**
```markdown
[![Image Size](https://img.shields.io/docker/image-size/kyklos/kyklos/latest?label=Image%20Size&logo=docker)](https://hub.docker.com/r/kyklos/kyklos)
```

**Rendered:**
![Image Size](https://img.shields.io/docker/image-size/kyklos/kyklos/latest?label=Image%20Size&logo=docker)

---

## Complete Examples

### Minimal Badge Set (3-4 badges)

```markdown
# Kyklos Time Window Scaler

[![CI](https://img.shields.io/github/actions/workflow/status/kyklos/kyklos/ci.yml?branch=main&label=CI&logo=github)](https://github.com/kyklos/kyklos/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kyklos/kyklos?label=Release&logo=github)](https://github.com/kyklos/kyklos/releases/latest)
[![License](https://img.shields.io/github/license/kyklos/kyklos?label=License)](https://github.com/kyklos/kyklos/blob/main/LICENSE)
[![Go Report](https://goreportcard.com/badge/github.com/kyklos/kyklos)](https://goreportcard.com/report/github.com/kyklos/kyklos)

Scale your Kubernetes workloads based on time windows and calendars.
```

---

### Standard Badge Set (6-8 badges)

```markdown
# Kyklos Time Window Scaler

[![CI Status](https://img.shields.io/github/actions/workflow/status/kyklos/kyklos/ci.yml?branch=main&label=CI&logo=github)](https://github.com/kyklos/kyklos/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kyklos/kyklos?label=Release&logo=github)](https://github.com/kyklos/kyklos/releases/latest)
[![License](https://img.shields.io/github/license/kyklos/kyklos?label=License)](https://github.com/kyklos/kyklos/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyklos/kyklos)](https://goreportcard.com/report/github.com/kyklos/kyklos)

[![Coverage](https://img.shields.io/codecov/c/github/kyklos/kyklos?label=Coverage&logo=codecov)](https://codecov.io/gh/kyklos/kyklos)
[![Security](https://img.shields.io/badge/Security-Trivy%20Scan-brightgreen?logo=aquasecuritytrivy)](https://github.com/kyklos/kyklos/security)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.26+-326CE5?logo=kubernetes)](https://kubernetes.io)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kyklos/kyklos?label=Go&logo=go)](https://go.dev)

Scale your Kubernetes workloads based on time windows and calendars.
```

---

### Comprehensive Badge Set (10+ badges)

```markdown
# Kyklos Time Window Scaler

<!-- Primary Badges -->
[![CI Status](https://img.shields.io/github/actions/workflow/status/kyklos/kyklos/ci.yml?branch=main&label=CI&logo=github)](https://github.com/kyklos/kyklos/actions/workflows/ci.yml)
[![Release Status](https://img.shields.io/github/actions/workflow/status/kyklos/kyklos/release.yml?label=Release&logo=github&event=release)](https://github.com/kyklos/kyklos/actions/workflows/release.yml)
[![Release Version](https://img.shields.io/github/v/release/kyklos/kyklos?label=Release&logo=github)](https://github.com/kyklos/kyklos/releases/latest)
[![License](https://img.shields.io/github/license/kyklos/kyklos?label=License)](https://github.com/kyklos/kyklos/blob/main/LICENSE)

<!-- Code Quality -->
[![Go Report Card](https://goreportcard.com/badge/github.com/kyklos/kyklos)](https://goreportcard.com/report/github.com/kyklos/kyklos)
[![Coverage](https://img.shields.io/codecov/c/github/kyklos/kyklos?label=Coverage&logo=codecov)](https://codecov.io/gh/kyklos/kyklos)
[![Go Version](https://img.shields.io/github/go-mod/go-version/kyklos/kyklos?label=Go&logo=go)](https://go.dev)

<!-- Security -->
[![Security Policy](https://img.shields.io/badge/Security-Policy-blue?logo=shield)](https://github.com/kyklos/kyklos/security/policy)
[![Trivy Scan](https://img.shields.io/badge/Trivy-Scan%20Clean-brightgreen?logo=aquasecuritytrivy)](https://github.com/kyklos/kyklos/security)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/kyklos/kyklos/badge)](https://securityscorecards.dev/viewer/?uri=github.com/kyklos/kyklos)

<!-- Deployment -->
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.26%20|%201.27%20|%201.28-326CE5?logo=kubernetes)](https://kubernetes.io)
[![Image Size](https://img.shields.io/badge/Image%20Size-24MB-blue?logo=docker)](https://ghcr.io/kyklos/kyklos)

<!-- Community -->
[![Contributors](https://img.shields.io/github/contributors/kyklos/kyklos?label=Contributors)](https://github.com/kyklos/kyklos/graphs/contributors)
[![Stars](https://img.shields.io/github/stars/kyklos/kyklos?style=social)](https://github.com/kyklos/kyklos/stargazers)
[![Slack](https://img.shields.io/badge/Slack-Join%20Community-4A154B?logo=slack)](https://kyklos-community.slack.com)

Scale your Kubernetes workloads based on time windows and calendars.
```

---

## Badge Customization

### Color Schemes

```markdown
# Success (green)
?color=brightgreen

# Warning (yellow)
?color=yellow

# Error (red)
?color=red

# Info (blue)
?color=blue

# Custom hex color
?color=ff69b4
```

### Logo Options

```markdown
# Built-in logos (shields.io)
?logo=github
?logo=docker
?logo=kubernetes
?logo=go
?logo=codecov

# Custom logo (base64 encoded)
?logo=data:image/svg+xml;base64,...
```

### Label Customization

```markdown
# Custom label
?label=Custom%20Label

# No label
?label=

# Label and message
?label=Build&message=Passing
```

---

## Dynamic Badge Updates

### GitHub Actions Badge Auto-Update

Badges linked to GitHub Actions update automatically:

```markdown
[![CI](https://img.shields.io/github/actions/workflow/status/kyklos/kyklos/ci.yml?branch=main)](...)
```

**Update Trigger:** Every workflow run completion

---

### Manual Badge Update Workflow

For badges requiring manual updates (e.g., image size):

```yaml
- name: Update README badges
  run: |
    IMAGE_SIZE=$(docker inspect ghcr.io/kyklos/kyklos:latest --format='{{.Size}}' | numfmt --to=iec-i --suffix=B)
    sed -i "s/Image%20Size-.*MB/Image%20Size-${IMAGE_SIZE}/" README.md
    git commit -am "docs: update image size badge"
    git push
```

---

## Badge Testing

**Preview badges before committing:**

1. Visit [shields.io](https://shields.io)
2. Use badge builder
3. Test URL in browser
4. Verify image renders correctly

**Example Test:**
```
https://img.shields.io/badge/Test-Badge-brightgreen?logo=github
```

---

## Document Version

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-28 | ci-release-engineer | Initial badge documentation |

## Related Documents

- [../release/REGISTRY-MAP.md](/Users/aykumar/personal/kyklos/docs/release/REGISTRY-MAP.md) - Container registry strategy
- [PIPELINE.md](/Users/aykumar/personal/kyklos/docs/ci/PIPELINE.md) - CI/CD pipeline design
- [../README.md](/Users/aykumar/personal/kyklos/README.md) - Project README
