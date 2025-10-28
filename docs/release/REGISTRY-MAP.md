# Container Registry Strategy

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** ci-release-engineer

This document defines the container image naming, tagging, and registry strategy for Kyklos. It provides clear guidelines for image publication, multi-architecture support, and branch/tag mappings.

---

## Table of Contents

1. [Registry Selection](#registry-selection)
2. [Image Naming Convention](#image-naming-convention)
3. [Tagging Strategy](#tagging-strategy)
4. [Multi-Architecture Support](#multi-architecture-support)
5. [Branch and Tag Mapping](#branch-and-tag-mapping)
6. [Retention and Cleanup](#retention-and-cleanup)
7. [Image Verification](#image-verification)

---

## Registry Selection

### Primary Registry: GitHub Container Registry (ghcr.io)

**Registry URL:** `ghcr.io/kyklos/kyklos`

**Reasons for Selection:**
- **Tight GitHub Integration:** Automatic OIDC authentication, no separate credentials
- **Unlimited Public Images:** Free for open-source projects
- **High Availability:** 99.9% SLA, global CDN
- **Built-in Security:** Vulnerability scanning, SBOM support
- **Access Control:** Granular permissions via GitHub teams

**Authentication:**
```bash
# Using GitHub Personal Access Token
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Using OIDC in GitHub Actions (automatic)
- uses: docker/login-action@v2
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}
```

### Secondary Registry: Docker Hub (docker.io)

**Registry URL:** `docker.io/kyklos/kyklos` (future consideration)

**Purpose:**
- Broader discoverability
- Fallback if ghcr.io unavailable
- Support for users with Docker Hub workflows

**Status:** Not implemented in v0.1, planned for v1.0

---

## Image Naming Convention

### Full Image Reference Format

```
<registry>/<namespace>/<repository>:<tag>@<digest>
```

**Example:**
```
ghcr.io/kyklos/kyklos:v0.1.0@sha256:abc123...
│       │      │      │       │
│       │      │      │       └─ Digest (SHA256 hash)
│       │      │      └───────── Tag (version)
│       │      └──────────────── Repository (project name)
│       └─────────────────────── Namespace (organization)
└─────────────────────────────── Registry (ghcr.io)
```

### Repository Name

**Standard:** `ghcr.io/kyklos/kyklos`

**Components:**
- **Namespace:** `kyklos` (GitHub organization name)
- **Repository:** `kyklos` (project/controller name)

**Rationale:**
- Simple, memorable
- Matches GitHub repository structure
- Consistent with Kubernetes naming (kyklos-controller)

### Repository Variants

Kyklos uses a **single repository** with multi-arch manifests:

**Pattern:** `ghcr.io/kyklos/kyklos:<tag>`

**NOT using separate repositories per arch:**
```
❌ ghcr.io/kyklos/kyklos-amd64:v0.1.0
❌ ghcr.io/kyklos/kyklos-arm64:v0.1.0
```

**Benefits:**
- Single image reference works on all platforms
- Docker/Kubernetes automatically pulls correct architecture
- Simplified documentation and deployment manifests

---

## Tagging Strategy

### Tag Categories

Kyklos uses multiple concurrent tags for different use cases:

```
ghcr.io/kyklos/kyklos:v0.1.0          # Semantic version (immutable)
ghcr.io/kyklos/kyklos:v0.1            # Minor version (mutable)
ghcr.io/kyklos/kyklos:v0              # Major version (mutable)
ghcr.io/kyklos/kyklos:latest          # Latest stable (mutable)
ghcr.io/kyklos/kyklos:main-abc123     # Branch + commit SHA (immutable)
ghcr.io/kyklos/kyklos:pr-456          # Pull request number (mutable)
```

---

### Semantic Version Tags

**Format:** `v<MAJOR>.<MINOR>.<PATCH>[-<PRERELEASE>]`

**Examples:**
- `v0.1.0` - First stable release
- `v0.1.1` - Patch release
- `v0.2.0` - Minor release with new features
- `v1.0.0` - Major release (breaking changes)
- `v0.1.0-alpha.1` - Pre-release alpha
- `v0.1.0-beta.2` - Pre-release beta
- `v0.1.0-rc.1` - Release candidate

**Characteristics:**
- **Immutable:** Once published, never overwritten
- **Production Use:** Safe for production deployments
- **Pinned:** Users can pin to exact version

**Recommendation:** Use semantic version tags in production manifests

```yaml
spec:
  containers:
  - name: kyklos-controller
    image: ghcr.io/kyklos/kyklos:v0.1.0  # ✅ Recommended
```

---

### Rolling Version Tags

**Minor Version Tag:** `v<MAJOR>.<MINOR>`

**Example:** `v0.1`

**Behavior:**
- Points to latest patch in that minor series
- Updated when new patches released

**Timeline:**
- v0.1.0 released → `v0.1` points to v0.1.0
- v0.1.1 released → `v0.1` updated to point to v0.1.1
- v0.1.2 released → `v0.1` updated to point to v0.1.2

**Use Case:** Automatic patch updates

```yaml
spec:
  containers:
  - name: kyklos-controller
    image: ghcr.io/kyklos/kyklos:v0.1  # Receives patches automatically
```

**Warning:** May update unexpectedly. Test before deploying.

---

**Major Version Tag:** `v<MAJOR>`

**Example:** `v0`

**Behavior:**
- Points to latest minor/patch in that major series
- Updated when new minors or patches released

**Timeline:**
- v0.1.0 released → `v0` points to v0.1.0
- v0.2.0 released → `v0` updated to point to v0.2.0
- v0.2.1 released → `v0` updated to point to v0.2.1

**Use Case:** Development environments, CI/CD testing

```yaml
spec:
  containers:
  - name: kyklos-controller
    image: ghcr.io/kyklos/kyklos:v0  # Always latest v0.x.x
```

**Warning:** May introduce breaking changes (in v0.x series). Not recommended for production.

---

### Latest Tag

**Format:** `latest`

**Behavior:**
- Points to most recent stable release (no pre-releases)
- Updated on every stable release

**Timeline:**
- v0.1.0 released → `latest` points to v0.1.0
- v0.2.0-beta.1 released → `latest` unchanged (still v0.1.0)
- v0.2.0 released → `latest` updated to v0.2.0

**Use Case:** Quick starts, demos, local development

```yaml
spec:
  containers:
  - name: kyklos-controller
    image: ghcr.io/kyklos/kyklos:latest  # ⚠️ Not recommended for production
```

**Warning:** Unpredictable updates, may break deployments. Avoid in production.

---

### Branch Tags

**Format:** `<branch>-<commit-sha>`

**Example:** `main-a1b2c3d`

**Characteristics:**
- Built from every commit to tracked branches (main, release/*)
- Immutable (unique SHA)
- Useful for CI/CD testing

**Use Case:** Integration testing, nightly builds

```bash
# Pull specific commit from main branch
docker pull ghcr.io/kyklos/kyklos:main-a1b2c3d
```

**Retention:** 30 days (see [Retention and Cleanup](#retention-and-cleanup))

---

### Pull Request Tags

**Format:** `pr-<number>`

**Example:** `pr-456`

**Characteristics:**
- Built from every PR commit
- Mutable (updated on new PR commits)
- Enables PR reviewers to test changes

**Use Case:** PR testing and validation

```bash
# Test PR #456 changes
docker pull ghcr.io/kyklos/kyklos:pr-456
```

**Retention:** Deleted when PR closed

---

## Multi-Architecture Support

### Supported Architectures

Kyklos provides multi-arch images for:

| Architecture | Platform String | Status | Notes |
|-------------|----------------|--------|-------|
| x86-64 | `linux/amd64` | ✅ Supported | Primary development platform |
| ARM64 | `linux/arm64` | ✅ Supported | AWS Graviton, Raspberry Pi |

**Planned (future):**
- `linux/arm/v7` - 32-bit ARM (Raspberry Pi 3)
- `linux/ppc64le` - IBM Power
- `linux/s390x` - IBM Z mainframe

### Manifest Lists

Kyklos uses Docker manifest lists (OCI image index) to bundle multiple architectures under a single tag.

**Single Tag, Multiple Architectures:**
```
ghcr.io/kyklos/kyklos:v0.1.0
├── linux/amd64 @ sha256:abc123...
└── linux/arm64 @ sha256:def456...
```

**How It Works:**
1. User pulls `ghcr.io/kyklos/kyklos:v0.1.0`
2. Docker/Kubernetes detects node architecture
3. Automatically pulls correct image for that platform
4. No user configuration needed

**Verification:**
```bash
# Inspect manifest list
docker manifest inspect ghcr.io/kyklos/kyklos:v0.1.0

# Output shows supported platforms:
{
  "manifests": [
    {
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      },
      "digest": "sha256:abc123..."
    },
    {
      "platform": {
        "architecture": "arm64",
        "os": "linux"
      },
      "digest": "sha256:def456..."
    }
  ]
}
```

### Build Process

Multi-arch images are built using Docker Buildx with QEMU emulation:

```yaml
- name: Set up QEMU
  uses: docker/setup-qemu-action@v2

- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v2

- name: Build and push multi-arch image
  uses: docker/build-push-action@v4
  with:
    platforms: linux/amd64,linux/arm64
    push: true
    tags: ghcr.io/kyklos/kyklos:v0.1.0
```

**Build Time:**
- amd64: ~2 minutes (native)
- arm64: ~4 minutes (QEMU emulation)
- Total: ~6 minutes for multi-arch build

---

## Branch and Tag Mapping

### Git → Image Tag Mapping

| Git Event | Image Tags Created | Retention | Use Case |
|-----------|-------------------|-----------|----------|
| Push to `main` | `main-<sha>` | 30 days | CI testing |
| Push to `release/v0.1` | `v0.1-<sha>` | 90 days | Pre-release testing |
| Push tag `v0.1.0` | `v0.1.0`, `v0.1`, `v0`, `latest` | Indefinite | Production release |
| Push tag `v0.1.0-beta.1` | `v0.1.0-beta.1` | 90 days | Beta testing |
| Push tag `v0.1.0-rc.1` | `v0.1.0-rc.1` | 90 days | Release candidate |
| Push to PR #456 | `pr-456` | Until PR closed | PR validation |

### Detailed Examples

#### Scenario 1: Feature Development on Main

```bash
# Commit to main branch
git commit -m "feat: add StatefulSet support"
git push origin main

# GitHub Actions builds and tags:
ghcr.io/kyklos/kyklos:main-a1b2c3d  # Commit SHA
```

**Verification:**
```bash
docker pull ghcr.io/kyklos/kyklos:main-a1b2c3d
```

---

#### Scenario 2: Stable Release

```bash
# Tag stable release
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0

# GitHub Actions builds and tags:
ghcr.io/kyklos/kyklos:v0.2.0   # Immutable semantic version
ghcr.io/kyklos/kyklos:v0.2     # Rolling minor version
ghcr.io/kyklos/kyklos:v0       # Rolling major version
ghcr.io/kyklos/kyklos:latest   # Latest stable
```

**Verification:**
```bash
# All point to same image digest
docker pull ghcr.io/kyklos/kyklos:v0.2.0
docker pull ghcr.io/kyklos/kyklos:v0.2
docker pull ghcr.io/kyklos/kyklos:latest

# Verify digest matches
docker inspect ghcr.io/kyklos/kyklos:v0.2.0 | jq -r '.[0].RepoDigests'
docker inspect ghcr.io/kyklos/kyklos:latest | jq -r '.[0].RepoDigests'
```

---

#### Scenario 3: Beta Release

```bash
# Tag beta release
git tag -a v0.2.0-beta.1 -m "Beta release v0.2.0-beta.1"
git push origin v0.2.0-beta.1

# GitHub Actions builds and tags:
ghcr.io/kyklos/kyklos:v0.2.0-beta.1  # Pre-release tag only

# Does NOT update:
ghcr.io/kyklos/kyklos:latest   # Still points to v0.1.0 (last stable)
ghcr.io/kyklos/kyklos:v0.2     # Does not exist yet (no stable v0.2.x)
```

**Rationale:** Pre-releases are isolated from stable tags to prevent accidental deployment.

---

#### Scenario 4: Hotfix Patch

```bash
# Branch from v0.1.0 tag
git checkout -b hotfix/v0.1.1 v0.1.0
git commit -m "fix: critical bug"
git tag -a v0.1.1 -m "Hotfix v0.1.1"
git push origin v0.1.1

# GitHub Actions builds and tags:
ghcr.io/kyklos/kyklos:v0.1.1   # New patch version
ghcr.io/kyklos/kyklos:v0.1     # Updated to v0.1.1 (was v0.1.0)
ghcr.io/kyklos/kyklos:v0       # Updated to v0.1.1 (if v0.1.1 is latest v0.x)

# If v0.2.0 already exists:
ghcr.io/kyklos/kyklos:latest   # Unchanged (still v0.2.0)
ghcr.io/kyklos/kyklos:v0       # Unchanged (still v0.2.0)
```

---

## Retention and Cleanup

### Retention Policies

| Tag Type | Retention | Rationale |
|----------|-----------|-----------|
| Semantic versions (v0.1.0) | Indefinite | Production releases |
| Pre-releases (v0.1.0-beta.1) | 90 days | Beta testing window |
| Branch tags (main-abc123) | 30 days | CI testing history |
| PR tags (pr-456) | Until PR closed | Active PR validation |
| Rolling tags (v0.1, latest) | Indefinite | Always point to current |

### Automated Cleanup

**Cleanup Workflow:**

```yaml
name: Image Cleanup

on:
  schedule:
    - cron: '0 3 * * 0'  # Weekly on Sunday 3 AM UTC

jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - name: Delete old branch tags
        # Delete main-* and pr-* tags older than 30 days
        run: |
          gh api repos/kyklos/kyklos/packages/container/kyklos/versions \
            --jq '.[] | select(.metadata.container.tags[] | startswith("main-") or startswith("pr-")) | select(.created_at < (now - 2592000)) | .id' | \
            xargs -I {} gh api -X DELETE repos/kyklos/kyklos/packages/container/kyklos/versions/{}

      - name: Delete old pre-release tags
        # Delete beta/rc tags older than 90 days
        run: |
          # Similar cleanup for pre-release tags
```

### Manual Cleanup

**Delete specific tag:**
```bash
# Using GitHub CLI
gh api -X DELETE repos/kyklos/kyklos/packages/container/kyklos/versions/<version-id>

# Using Docker (removes locally only)
docker rmi ghcr.io/kyklos/kyklos:old-tag
```

**Untagged Images:**

Untagged images (no tags pointing to them) are automatically removed by GitHub after 14 days.

---

## Image Verification

### Digest Verification

**Always verify image digest in production:**

```yaml
spec:
  containers:
  - name: kyklos-controller
    image: ghcr.io/kyklos/kyklos:v0.1.0@sha256:abc123...  # Pin to specific digest
```

**Why?**
- Tags are mutable (can be repointed)
- Digests are immutable (cryptographic hash)
- Guarantees exact image content

### Getting Image Digest

**From release notes:**
```markdown
## Image Digests

- linux/amd64: sha256:abc123def456...
- linux/arm64: sha256:789ghi012jkl...
```

**From command line:**
```bash
# Pull image and get digest
docker pull ghcr.io/kyklos/kyklos:v0.1.0
docker inspect ghcr.io/kyklos/kyklos:v0.1.0 --format='{{index .RepoDigests 0}}'

# Output: ghcr.io/kyklos/kyklos@sha256:abc123...
```

**From registry:**
```bash
# Query registry directly (no pull)
docker manifest inspect ghcr.io/kyklos/kyklos:v0.1.0 | jq -r '.config.digest'
```

### SBOM Verification

**Verify image provenance:**
```bash
# Download SBOM from release
curl -LO https://github.com/kyklos/kyklos/releases/download/v0.1.0/sbom.json

# Verify image matches SBOM
syft packages ghcr.io/kyklos/kyklos:v0.1.0 -o cyclonedx-json | \
  diff - sbom.json
```

### Signature Verification (Optional)

**If images are signed with Cosign:**

```bash
# Verify signature
cosign verify ghcr.io/kyklos/kyklos:v0.1.0 \
  --certificate-identity https://github.com/kyklos/kyklos/.github/workflows/release.yml \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com

# Output: Verified OK
```

**Signature in admission controller:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cosign-policy
data:
  policy: |
    apiVersion: v1
    images:
    - pattern: "ghcr.io/kyklos/kyklos:*"
      authorities:
      - keyless:
          url: https://fulcio.sigstore.dev
```

---

## Image Size Expectations

### Size Budgets

| Component | Target Size | Max Size | Actual (v0.1.0) |
|-----------|------------|----------|-----------------|
| Go Binary | 15 MB | 20 MB | 16 MB |
| Base Image (distroless) | 5 MB | 10 MB | 8 MB |
| Total Image | 20 MB | 30 MB | 24 MB |

**Size Breakdown:**
```
24 MB  Total image
├── 16 MB  Controller binary (Go, statically linked)
└──  8 MB  Distroless base image (gcr.io/distroless/static:nonroot)
```

### Optimization Techniques

**Applied:**
- Static linking (no libc dependencies)
- Stripped symbols (`-ldflags "-s -w"`)
- Distroless base image (no shell, no package manager)
- Multi-stage Dockerfile

**Future Optimizations:**
- UPX compression (may reduce 30-40%)
- Pruned dependencies
- Binary optimization flags

---

## Troubleshooting

### Issue: Image Pull Fails

**Symptoms:**
```
Failed to pull image "ghcr.io/kyklos/kyklos:v0.1.0": rpc error: code = Unknown desc = Error response from daemon: unauthorized
```

**Diagnosis:**
```bash
# Test authentication
docker login ghcr.io -u USERNAME

# Verify image exists
docker manifest inspect ghcr.io/kyklos/kyklos:v0.1.0
```

**Solutions:**
1. Check image visibility (public vs private)
2. Verify authentication credentials
3. Check ghcr.io service status
4. Try alternate registry (Docker Hub) if available

---

### Issue: Wrong Architecture Pulled

**Symptoms:**
```
exec /controller: exec format error
```

**Diagnosis:**
```bash
# Check node architecture
kubectl get nodes -o jsonpath='{.items[0].status.nodeInfo.architecture}'

# Check image architectures
docker manifest inspect ghcr.io/kyklos/kyklos:v0.1.0
```

**Solutions:**
1. Ensure multi-arch manifest exists
2. Verify platform string matches node
3. Pull specific architecture: `ghcr.io/kyklos/kyklos:v0.1.0-arm64` (if using arch-specific tags)

---

### Issue: Image Size Too Large

**Symptoms:**
- Pod creation timeout
- Image pull takes > 5 minutes

**Diagnosis:**
```bash
# Check actual image size
docker images ghcr.io/kyklos/kyklos:v0.1.0

# Check layers
docker history ghcr.io/kyklos/kyklos:v0.1.0
```

**Solutions:**
1. Verify using distroless base
2. Check for accidental inclusion of build tools
3. Review Dockerfile for unnecessary COPY commands
4. Use registry mirror closer to cluster

---

## Document Version

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-28 | ci-release-engineer | Initial registry strategy for v0.1 |

## Related Documents

- [RELEASE-POLICY.md](/Users/aykumar/personal/kyklos/docs/release/RELEASE-POLICY.md) - Release management
- [../ci/PIPELINE.md](/Users/aykumar/personal/kyklos/docs/ci/PIPELINE.md) - CI/CD pipeline design
- [../ci/WORKFLOWS-STUBS.md](/Users/aykumar/personal/kyklos/docs/ci/WORKFLOWS-STUBS.md) - GitHub Actions workflows
