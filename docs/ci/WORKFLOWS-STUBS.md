# GitHub Actions Workflow Structure

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** ci-release-engineer

This document provides complete GitHub Actions workflow structures for Kyklos CI/CD pipelines. These are structural stubs showing job organization, dependencies, caching strategies, and time budgets without implementation-specific commands.

---

## Table of Contents

1. [ci.yml - Pull Request and Main Branch CI](#ciyml---pull-request-and-main-branch-ci)
2. [release.yml - Tag-Based Release Workflow](#releaseyml---tag-based-release-workflow)
3. [cache-warm.yml - Daily Cache Pre-Population](#cache-warmyml---daily-cache-pre-population)
4. [Common Patterns](#common-patterns)
5. [Runner Matrix Configurations](#runner-matrix-configurations)

---

## ci.yml - Pull Request and Main Branch CI

### Purpose

Validate code quality, run tests, and perform smoke testing for every pull request and commit to main branch. Target completion time: under 10 minutes.

### Triggers

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:

concurrency:
  group: ci-${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true  # Cancel previous runs for same PR
```

### Job Structure

```yaml
jobs:
  # ============================================================
  # JOB 1: LINT (1 minute budget)
  # ============================================================
  lint:
    name: Code Quality Checks
    runs-on: ubuntu-latest
    timeout-minutes: 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history for accurate linting

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true

      - name: Restore golangci-lint cache
        uses: actions/cache@v3
        with:
          path: |
            ~/.cache/golangci-lint
            tools/bin/golangci-lint
          key: lint-${{ runner.os }}-${{ hashFiles('.golangci.yml', 'hack/tools.go') }}
          restore-keys: |
            lint-${{ runner.os }}-

      - name: Run linters
        # Implementation: golangci-lint run --config .golangci.yml
        run: echo "Linting Go code"

      - name: Validate YAML manifests
        # Implementation: validate CRD examples, kustomize configs
        run: echo "Validating YAML"

      - name: Check license headers
        # Implementation: addlicense -check
        run: echo "Checking license headers"

      - name: Markdown lint
        # Implementation: markdownlint docs/
        run: echo "Linting documentation"

  # ============================================================
  # JOB 2: UNIT TESTS (2 minute budget)
  # ============================================================
  unit-test:
    name: Unit Tests
    runs-on: ubuntu-latest
    timeout-minutes: 10

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true

      - name: Restore Go module cache
        uses: actions/cache@v3
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: unit-${{ runner.os }}-go-${{ hashFiles('go.sum') }}
          restore-keys: |
            unit-${{ runner.os }}-go-

      - name: Download dependencies
        # Implementation: go mod download
        run: echo "Downloading Go modules"

      - name: Run unit tests with race detector
        # Implementation: go test -v -race -coverprofile=coverage.out ./...
        run: echo "Running unit tests"

      - name: Generate coverage report
        # Implementation: go tool cover -html=coverage.out -o coverage.html
        run: echo "Generating coverage HTML"

      - name: Check coverage threshold
        # Implementation: check coverage >= 80%
        run: echo "Verifying coverage threshold"

      - name: Upload coverage artifacts
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: unit-test-coverage
          path: |
            coverage.out
            coverage.html
          retention-days: 30

      - name: Comment PR with coverage
        if: github.event_name == 'pull_request'
        # Implementation: post coverage delta to PR comments
        run: echo "Posting coverage to PR"

  # ============================================================
  # JOB 3: ENVTEST (3 minute budget)
  # ============================================================
  envtest:
    name: Controller Integration Tests
    runs-on: ubuntu-latest
    timeout-minutes: 15

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true

      - name: Restore envtest binaries cache
        uses: actions/cache@v3
        with:
          path: |
            ~/envtest-binaries
            tools/bin/setup-envtest
          key: envtest-${{ runner.os }}-k8s-1.28-${{ hashFiles('go.sum') }}
          restore-keys: |
            envtest-${{ runner.os }}-k8s-1.28-

      - name: Install envtest binaries
        # Implementation: setup-envtest use 1.28.x
        run: echo "Installing etcd and kube-apiserver"

      - name: Run envtest suite
        # Implementation: make test-envtest
        run: echo "Running controller tests against API server"

      - name: Capture API server logs on failure
        if: failure()
        run: echo "Saving API server logs"

      - name: Upload envtest artifacts
        uses: actions/upload-artifact@v3
        if: failure()
        with:
          name: envtest-logs
          path: |
            /tmp/envtest-*
            test-namespaces.yaml
          retention-days: 90

  # ============================================================
  # JOB 4: BUILD BINARY (3 minute budget)
  # ============================================================
  build:
    name: Build Controller Binary
    runs-on: ubuntu-latest
    timeout-minutes: 10

    strategy:
      matrix:
        platform:
          - os: linux
            arch: amd64
          - os: linux
            arch: arm64

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true

      - name: Restore build cache
        uses: actions/cache@v3
        with:
          path: ~/.cache/go-build
          key: build-${{ runner.os }}-${{ matrix.platform.os }}-${{ matrix.platform.arch }}-${{ hashFiles('**/*.go') }}
          restore-keys: |
            build-${{ runner.os }}-${{ matrix.platform.os }}-${{ matrix.platform.arch }}-

      - name: Generate code
        # Implementation: make manifests generate
        run: echo "Running code generation"

      - name: Build controller binary
        # Implementation: GOOS=${{ matrix.platform.os }} GOARCH=${{ matrix.platform.arch }} make build
        run: echo "Compiling controller"

      - name: Verify binary
        # Implementation: ./bin/controller --version
        run: echo "Checking binary metadata"

      - name: Upload binary artifact
        uses: actions/upload-artifact@v3
        with:
          name: controller-${{ matrix.platform.os }}-${{ matrix.platform.arch }}
          path: bin/controller
          retention-days: 7

  # ============================================================
  # JOB 5: VERIFY CODE QUALITY (1 minute budget)
  # ============================================================
  verify:
    name: Additional Quality Checks
    runs-on: ubuntu-latest
    timeout-minutes: 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Check for TODOs without issue links
        # Implementation: grep -r "TODO" --include="*.go" | check for issue refs
        run: echo "Verifying TODOs have issue links"

      - name: Check for hardcoded secrets
        # Implementation: gitleaks detect
        run: echo "Scanning for leaked secrets"

      - name: Verify godoc coverage
        # Implementation: check exported functions have comments
        run: echo "Checking documentation coverage"

      - name: Validate CRD examples
        # Implementation: kubectl validate against generated CRD
        run: echo "Validating example YAMLs"

  # ============================================================
  # JOB 6: KIND SMOKE TEST (5 minute budget)
  # Depends on: lint, unit-test, envtest, build, verify
  # ============================================================
  smoke-test:
    name: Kind Cluster Smoke Test
    runs-on: ubuntu-latest
    timeout-minutes: 15
    needs: [lint, unit-test, envtest, build, verify]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true

      - name: Restore Kind cache
        uses: actions/cache@v3
        with:
          path: |
            ~/.cache/kind
            tools/bin/kind
          key: kind-${{ runner.os }}-${{ hashFiles('.kind-config.yaml') }}
          restore-keys: |
            kind-${{ runner.os }}-

      - name: Install Kind
        # Implementation: install Kind if not cached
        run: echo "Installing Kind"

      - name: Create Kind cluster
        # Implementation: kind create cluster --config .kind-config.yaml
        run: echo "Creating local Kubernetes cluster"

      - name: Build controller image
        # Implementation: make docker-build IMG=kyklos/controller:pr-${{ github.run_number }}
        run: echo "Building container image"

      - name: Load image into Kind
        # Implementation: kind load docker-image kyklos/controller:pr-${{ github.run_number }}
        run: echo "Loading image into cluster"

      - name: Install CRDs
        # Implementation: kubectl apply -f config/crd/bases/
        run: echo "Installing CustomResourceDefinitions"

      - name: Deploy controller
        # Implementation: kustomize build config/default | kubectl apply -f -
        run: echo "Deploying controller"

      - name: Wait for controller ready
        # Implementation: kubectl wait --timeout=60s --for=condition=Ready pod -n kyklos-system
        run: echo "Waiting for controller pod"

      - name: Create test deployment
        # Implementation: kubectl apply -f test/fixtures/demo-deployment.yaml
        run: echo "Creating test target"

      - name: Apply minute-scale TimeWindowScaler
        # Implementation: kubectl apply -f test/fixtures/tws-minute-demo.yaml
        run: echo "Applying TimeWindowScaler"

      - name: Observe scaling behavior
        # Implementation: watch for 3 minutes, verify scale up and down
        run: echo "Verifying time-based scaling"

      - name: Verify metrics endpoint
        # Implementation: kubectl port-forward and curl /metrics
        run: echo "Checking Prometheus metrics"

      - name: Verify events emitted
        # Implementation: kubectl get events -n demo
        run: echo "Checking scale events"

      - name: Export cluster diagnostics on failure
        if: failure()
        run: |
          echo "Capturing cluster state"
          # Implementation: kubectl get all --all-namespaces -o yaml
          # kubectl logs -n kyklos-system -l app=kyklos-controller
          # kind export logs

      - name: Upload smoke test artifacts
        uses: actions/upload-artifact@v3
        if: failure()
        with:
          name: smoke-test-diagnostics
          path: |
            /tmp/kind-logs/
            cluster-state.yaml
            controller-logs.txt
          retention-days: 90

      - name: Cleanup Kind cluster
        if: always()
        # Implementation: kind delete cluster
        run: echo "Deleting test cluster"

  # ============================================================
  # JOB 7: REPORT (30 seconds)
  # Depends on: smoke-test
  # ============================================================
  report:
    name: Generate Test Report
    runs-on: ubuntu-latest
    timeout-minutes: 5
    needs: [smoke-test]
    if: always()

    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v3

      - name: Aggregate test results
        # Implementation: combine coverage, test results, smoke test outcomes
        run: echo "Aggregating results"

      - name: Post summary to PR
        if: github.event_name == 'pull_request'
        # Implementation: post summary comment with pass/fail breakdown
        run: echo "Posting test summary to PR"

      - name: Update commit status
        # Implementation: set commit status check
        run: echo "Setting GitHub status check"
```

---

## release.yml - Tag-Based Release Workflow

### Purpose

Build multi-arch images, perform security scanning, run comprehensive E2E tests, and publish release artifacts when a version tag is pushed.

### Triggers

```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'
  workflow_dispatch:
    inputs:
      tag:
        description: 'Release tag (e.g., v0.1.0)'
        required: true

concurrency:
  group: release-${{ github.ref }}
  cancel-in-progress: false  # Never cancel release builds
```

### Job Structure

```yaml
jobs:
  # ============================================================
  # JOB 1-5: Reuse from ci.yml
  # ============================================================
  lint:
    # Same as ci.yml lint job
    name: Code Quality Checks
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      # ... (same steps as ci.yml)

  unit-test:
    # Same as ci.yml unit-test job
    name: Unit Tests
    runs-on: ubuntu-latest
    timeout-minutes: 10
    steps:
      # ... (same steps as ci.yml)

  envtest:
    # Same as ci.yml envtest job
    name: Controller Integration Tests
    runs-on: ubuntu-latest
    timeout-minutes: 15
    steps:
      # ... (same steps as ci.yml)

  build:
    # Same as ci.yml build job
    name: Build Controller Binary
    runs-on: ubuntu-latest
    timeout-minutes: 10
    strategy:
      matrix:
        platform:
          - os: linux
            arch: amd64
          - os: linux
            arch: arm64
    steps:
      # ... (same steps as ci.yml)

  # ============================================================
  # JOB 6: MULTI-ARCH IMAGE BUILD (4 minute budget)
  # Depends on: lint, unit-test, envtest, build
  # ============================================================
  build-image:
    name: Build Multi-Arch Container Image
    runs-on: ubuntu-latest
    timeout-minutes: 15
    needs: [lint, unit-test, envtest, build]

    permissions:
      contents: read
      packages: write
      id-token: write  # For OIDC token

    outputs:
      image-digest: ${{ steps.build.outputs.digest }}
      image-tag: ${{ steps.meta.outputs.tags }}

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up QEMU for multi-arch
        uses: docker/setup-qemu-action@v2

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
        with:
          buildkitd-flags: --debug

      - name: Restore Docker layer cache
        uses: actions/cache@v3
        with:
          path: /tmp/.buildx-cache
          key: docker-${{ runner.os }}-${{ hashFiles('**/Dockerfile') }}
          restore-keys: |
            docker-${{ runner.os }}-

      - name: Log in to GitHub Container Registry
        uses: docker/login-action@v2
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract version from tag
        id: version
        # Implementation: parse GITHUB_REF to get v0.1.0 format
        run: echo "Extracting version"

      - name: Generate image metadata
        id: meta
        uses: docker/metadata-action@v4
        with:
          images: ghcr.io/${{ github.repository }}
          tags: |
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix={{branch}}-

      - name: Build and push multi-arch image
        id: build
        uses: docker/build-push-action@v4
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=local,src=/tmp/.buildx-cache
          cache-to: type=local,dest=/tmp/.buildx-cache-new,mode=max
          build-args: |
            VERSION=${{ steps.version.outputs.version }}
            COMMIT=${{ github.sha }}
            BUILD_DATE=${{ github.event.head_commit.timestamp }}

      - name: Generate SBOM
        # Implementation: syft packages ghcr.io/.../controller:$version -o cyclonedx-json
        run: echo "Generating Software Bill of Materials"

      - name: Upload SBOM artifact
        uses: actions/upload-artifact@v3
        with:
          name: sbom
          path: sbom.json
          retention-days: 365

      - name: Move cache (workaround for cache size)
        # Implementation: cleanup old cache
        run: |
          rm -rf /tmp/.buildx-cache
          mv /tmp/.buildx-cache-new /tmp/.buildx-cache

  # ============================================================
  # JOB 7: SECURITY SCAN (2 minute budget)
  # Depends on: build-image
  # ============================================================
  security-scan:
    name: Vulnerability Scanning
    runs-on: ubuntu-latest
    timeout-minutes: 10
    needs: [build-image]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Restore Trivy cache
        uses: actions/cache@v3
        with:
          path: ~/.cache/trivy
          key: trivy-${{ runner.os }}-${{ github.run_id }}
          restore-keys: |
            trivy-${{ runner.os }}-

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ needs.build-image.outputs.image-tag }}
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'
          exit-code: 1  # Fail on vulnerabilities

      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        if: always()
        with:
          sarif_file: 'trivy-results.sarif'

      - name: Generate human-readable report
        if: always()
        # Implementation: trivy image --format table
        run: echo "Generating security summary"

      - name: Upload security artifacts
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: security-scan
          path: |
            trivy-results.sarif
            security-summary.txt
          retention-days: 365

  # ============================================================
  # JOB 8: E2E FULL TEST SUITE (8 minute budget)
  # Depends on: build-image, security-scan
  # ============================================================
  e2e-full:
    name: End-to-End Test Suite
    runs-on: ubuntu-latest
    timeout-minutes: 30
    needs: [build-image, security-scan]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true

      - name: Restore E2E cache
        uses: actions/cache@v3
        with:
          path: |
            ~/.cache/kind
            ~/.cache/e2e-binaries
          key: e2e-${{ runner.os }}-${{ hashFiles('test/e2e/**') }}
          restore-keys: |
            e2e-${{ runner.os }}-

      - name: Create Kind cluster
        # Implementation: kind create cluster
        run: echo "Creating E2E cluster"

      - name: Load controller image
        # Implementation: docker pull and kind load
        run: echo "Loading release image"

      - name: Deploy controller
        # Implementation: kubectl apply
        run: echo "Deploying controller"

      - name: Run demo scenario A (office hours)
        # Implementation: test/e2e/scenario-a.sh
        run: echo "Testing office hours pattern"

      - name: Run demo scenario B (cross-midnight)
        # Implementation: test/e2e/scenario-b.sh
        run: echo "Testing cross-midnight windows"

      - name: Test DST transition handling
        # Implementation: test with timezone changes
        run: echo "Testing daylight saving time"

      - name: Test grace period behavior
        # Implementation: verify delayed scale-down
        run: echo "Testing grace periods"

      - name: Test pause/resume functionality
        # Implementation: verify pause flag
        run: echo "Testing pause mode"

      - name: Validate metrics accuracy
        # Implementation: scrape /metrics and verify counts
        run: echo "Checking Prometheus metrics"

      - name: Export E2E diagnostics on failure
        if: failure()
        # Implementation: export logs, events, resources
        run: echo "Capturing E2E cluster state"

      - name: Upload E2E artifacts
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: e2e-results
          path: |
            test/e2e/results/
            e2e-logs.tar.gz
            scenario-traces/
          retention-days: 90

      - name: Cleanup E2E cluster
        if: always()
        run: echo "Deleting E2E cluster"

  # ============================================================
  # JOB 9: PUBLISH RELEASE ARTIFACTS (2 minute budget)
  # Depends on: e2e-full
  # ============================================================
  publish:
    name: Publish Release Artifacts
    runs-on: ubuntu-latest
    timeout-minutes: 10
    needs: [build-image, e2e-full]

    permissions:
      contents: write
      packages: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history for changelog

      - name: Download SBOM artifact
        uses: actions/download-artifact@v3
        with:
          name: sbom

      - name: Generate installation bundle
        # Implementation: kustomize build config/default > install.yaml
        run: echo "Creating installation bundle"

      - name: Create release archive
        # Implementation: tar czf kyklos-$version.tar.gz config/ examples/
        run: echo "Archiving source and configs"

      - name: Generate checksums
        # Implementation: sha256sum install.yaml sbom.json kyklos-*.tar.gz
        run: echo "Computing SHA256 checksums"

      - name: Tag images as 'latest' (optional)
        if: "!contains(github.ref, '-alpha') && !contains(github.ref, '-beta')"
        # Implementation: docker tag and push :latest
        run: echo "Tagging stable release as latest"

      - name: Create GitHub Release
        uses: actions/create-release@v1
        id: create_release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ github.ref_name }}
          release_name: Kyklos ${{ github.ref_name }}
          body_path: RELEASE-NOTES-${{ github.ref_name }}.md
          draft: false
          prerelease: ${{ contains(github.ref, '-alpha') || contains(github.ref, '-beta') }}

      - name: Upload install.yaml
        uses: actions/upload-release-asset@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          upload_url: ${{ steps.create_release.outputs.upload_url }}
          asset_path: ./install.yaml
          asset_name: install.yaml
          asset_content_type: application/x-yaml

      - name: Upload SBOM
        uses: actions/upload-release-asset@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          upload_url: ${{ steps.create_release.outputs.upload_url }}
          asset_path: ./sbom.json
          asset_name: sbom.json
          asset_content_type: application/json

      - name: Upload checksums
        uses: actions/upload-release-asset@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          upload_url: ${{ steps.create_release.outputs.upload_url }}
          asset_path: ./checksums.txt
          asset_name: checksums.txt
          asset_content_type: text/plain

  # ============================================================
  # JOB 10: GENERATE RELEASE NOTES (1 minute budget)
  # Depends on: publish
  # ============================================================
  release-notes:
    name: Generate Release Notes
    runs-on: ubuntu-latest
    timeout-minutes: 5
    needs: [publish]

    permissions:
      contents: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Get previous tag
        id: previous_tag
        # Implementation: git describe --tags --abbrev=0 HEAD^
        run: echo "Finding previous release"

      - name: Generate changelog
        # Implementation: parse commits between tags, group by conventional commit type
        run: echo "Generating changelog from commits"

      - name: Extract breaking changes
        # Implementation: grep for BREAKING CHANGE in commit messages
        run: echo "Identifying breaking changes"

      - name: Link issues and PRs
        # Implementation: convert #123 to GitHub issue links
        run: echo "Adding hyperlinks to issues"

      - name: Update release notes
        uses: actions/github-script@v6
        with:
          script: |
            // Implementation: update release description with generated notes
            console.log('Updating release notes')

      - name: Upload changelog artifact
        uses: actions/upload-artifact@v3
        with:
          name: changelog
          path: |
            CHANGELOG-${{ github.ref_name }}.md
            RELEASE-NOTES-${{ github.ref_name }}.md
          retention-days: 365
```

---

## cache-warm.yml - Daily Cache Pre-Population

### Purpose

Pre-populate caches daily to ensure fast first builds of the day. Reduces cold start time by 80%.

### Triggers

```yaml
name: Cache Warm-Up

on:
  schedule:
    - cron: '0 2 * * *'  # 2 AM UTC daily
  workflow_dispatch:

concurrency:
  group: cache-warm
  cancel-in-progress: true
```

### Job Structure

```yaml
jobs:
  warm-go-cache:
    name: Pre-populate Go Dependency Cache
    runs-on: ubuntu-latest
    timeout-minutes: 10

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true

      - name: Download all dependencies
        # Implementation: go mod download
        run: echo "Downloading Go modules"

      - name: Build to populate build cache
        # Implementation: go build ./...
        run: echo "Populating build cache"

      - name: Save Go cache
        uses: actions/cache/save@v3
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: unit-${{ runner.os }}-go-${{ hashFiles('go.sum') }}

  warm-tool-cache:
    name: Pre-populate Development Tool Cache
    runs-on: ubuntu-latest
    timeout-minutes: 15

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true

      - name: Install controller-gen
        # Implementation: go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
        run: echo "Installing controller-gen"

      - name: Install golangci-lint
        # Implementation: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        run: echo "Installing golangci-lint"

      - name: Install setup-envtest
        # Implementation: go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
        run: echo "Installing setup-envtest"

      - name: Download envtest binaries
        # Implementation: setup-envtest use 1.28.x
        run: echo "Downloading Kubernetes test binaries"

      - name: Save tool cache
        uses: actions/cache/save@v3
        with:
          path: |
            ~/go/bin/
            ~/envtest-binaries/
          key: tools-${{ runner.os }}-${{ hashFiles('hack/tools.go') }}

  warm-docker-cache:
    name: Pre-populate Docker Layer Cache
    runs-on: ubuntu-latest
    timeout-minutes: 20

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Pull base images
        # Implementation: docker pull golang:1.21, distroless/static:nonroot
        run: echo "Pulling base images"

      - name: Build with layer cache
        uses: docker/build-push-action@v4
        with:
          context: .
          push: false
          cache-from: type=local,src=/tmp/.buildx-cache
          cache-to: type=local,dest=/tmp/.buildx-cache-new,mode=max

      - name: Save Docker cache
        uses: actions/cache/save@v3
        with:
          path: /tmp/.buildx-cache
          key: docker-${{ runner.os }}-${{ hashFiles('**/Dockerfile') }}

  warm-kind-cache:
    name: Pre-populate Kind Node Image
    runs-on: ubuntu-latest
    timeout-minutes: 10

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Install Kind
        # Implementation: curl -Lo kind https://github.com/kubernetes-sigs/kind/releases/download/...
        run: echo "Installing Kind"

      - name: Pull Kind node image
        # Implementation: docker pull kindest/node:v1.28.0
        run: echo "Pulling Kind node image"

      - name: Save Kind cache
        uses: actions/cache/save@v3
        with:
          path: |
            ~/.cache/kind
            tools/bin/kind
          key: kind-${{ runner.os }}-${{ hashFiles('.kind-config.yaml') }}
```

---

## Common Patterns

### Cache Key Strategy

**Pattern:**
```yaml
key: <job>-${{ runner.os }}-<content-hash>
restore-keys: |
  <job>-${{ runner.os }}-
```

**Examples:**
- `unit-${{ runner.os }}-go-${{ hashFiles('go.sum') }}`
- `lint-${{ runner.os }}-${{ hashFiles('.golangci.yml') }}`
- `docker-${{ runner.os }}-${{ hashFiles('**/Dockerfile') }}`

### Conditional Execution

**Skip on Draft PRs:**
```yaml
if: "!github.event.pull_request.draft"
```

**Run Only on Tags:**
```yaml
if: startsWith(github.ref, 'refs/tags/v')
```

**Failure Artifact Upload:**
```yaml
if: failure()
```

**Always Run (Even on Failure):**
```yaml
if: always()
```

### Job Dependencies

**Serial Dependency:**
```yaml
jobs:
  job-a:
    # runs first
  job-b:
    needs: [job-a]  # waits for job-a
```

**Parallel with Barrier:**
```yaml
jobs:
  lint:
    # runs in parallel
  test:
    # runs in parallel
  build:
    # runs in parallel
  deploy:
    needs: [lint, test, build]  # waits for all three
```

### Artifact Management

**Upload with Conditional Retention:**
```yaml
- name: Upload artifacts
  uses: actions/upload-artifact@v3
  if: failure()  # Only on failure
  with:
    name: debug-logs
    path: /tmp/logs/
    retention-days: 90  # Keep failures longer
```

**Download for Aggregation:**
```yaml
- name: Download all artifacts
  uses: actions/download-artifact@v3
  # Downloads to current directory
```

---

## Runner Matrix Configurations

### Operating System Matrix

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
runs-on: ${{ matrix.os }}
```

### Go Version Matrix

```yaml
strategy:
  matrix:
    go-version: ['1.21', '1.22']
runs-on: ubuntu-latest
steps:
  - uses: actions/setup-go@v4
    with:
      go-version: ${{ matrix.go-version }}
```

### Kubernetes Version Matrix

```yaml
strategy:
  matrix:
    k8s-version: ['1.25', '1.26', '1.27', '1.28']
runs-on: ubuntu-latest
steps:
  - name: Create Kind cluster
    run: |
      kind create cluster --image kindest/node:v${{ matrix.k8s-version }}.0
```

### Platform Build Matrix

```yaml
strategy:
  matrix:
    platform:
      - os: linux
        arch: amd64
      - os: linux
        arch: arm64
      - os: darwin
        arch: amd64
      - os: darwin
        arch: arm64
runs-on: ubuntu-latest
steps:
  - name: Build
    run: |
      GOOS=${{ matrix.platform.os }} GOARCH=${{ matrix.platform.arch }} make build
```

---

## Time Budget Enforcement

### Job-Level Timeouts

```yaml
jobs:
  lint:
    timeout-minutes: 5  # Hard stop after 5 minutes
  unit-test:
    timeout-minutes: 10
  smoke-test:
    timeout-minutes: 15
```

### Step-Level Timeouts

```yaml
steps:
  - name: Run tests
    timeout-minutes: 8
    run: make test
```

### Test-Level Timeouts

```go
// In Go test code
func TestLongRunning(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    // ... test logic
}
```

---

## Security Considerations

### Secrets Management

```yaml
env:
  GHCR_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  COSIGN_KEY: ${{ secrets.COSIGN_PRIVATE_KEY }}

steps:
  - name: Mask secrets in logs
    run: |
      echo "::add-mask::$GHCR_TOKEN"
```

### OIDC Token Authentication

```yaml
permissions:
  id-token: write  # For OIDC
  contents: read
  packages: write

steps:
  - name: Log in with OIDC
    uses: docker/login-action@v2
    with:
      registry: ghcr.io
      username: ${{ github.actor }}
      password: ${{ secrets.GITHUB_TOKEN }}
```

### Signed Commits Verification

```yaml
steps:
  - name: Verify commit signature
    run: |
      git verify-commit HEAD || exit 1
```

---

## Optimization Tips

### 1. Aggressive Parallelization

Run independent jobs in parallel:
```yaml
jobs:
  lint:     # Start immediately
  test:     # Start immediately
  build:    # Start immediately
  verify:   # Start immediately
  deploy:
    needs: [lint, test, build, verify]  # Wait for all
```

### 2. Selective Job Execution

Skip jobs based on file changes:
```yaml
jobs:
  test-go:
    if: contains(github.event.head_commit.modified, '.go')
  test-docs:
    if: contains(github.event.head_commit.modified, '.md')
```

### 3. Cache Layering

Use restore-keys for progressive cache hits:
```yaml
restore-keys: |
  build-${{ runner.os }}-${{ matrix.arch }}-
  build-${{ runner.os }}-
  build-
```

### 4. Fail-Fast Matrix

Abort remaining matrix jobs on first failure:
```yaml
strategy:
  fail-fast: true
  matrix:
    go-version: ['1.21', '1.22']
```

### 5. Artifact Size Optimization

Compress large artifacts:
```yaml
- name: Compress logs
  run: tar czf logs.tar.gz /tmp/logs/
- name: Upload compressed
  uses: actions/upload-artifact@v3
  with:
    path: logs.tar.gz
```

---

## Debugging Workflows

### Enable Debug Logging

Set repository secrets:
- `ACTIONS_STEP_DEBUG`: `true`
- `ACTIONS_RUNNER_DEBUG`: `true`

### Inspect Cache Keys

```yaml
- name: Show cache key
  run: |
    echo "Cache key: unit-${{ runner.os }}-go-${{ hashFiles('go.sum') }}"
    echo "Go sum hash: ${{ hashFiles('go.sum') }}"
```

### Manual Workflow Dispatch

```yaml
on:
  workflow_dispatch:
    inputs:
      debug_enabled:
        description: 'Enable SSH debugging'
        required: false
        default: 'false'

jobs:
  test:
    steps:
      - name: Setup tmate session
        if: ${{ github.event.inputs.debug_enabled == 'true' }}
        uses: mxschmitt/action-tmate@v3
```

---

## Document Version

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-28 | ci-release-engineer | Initial workflow structure for v0.1 |

## Related Documents

- [PIPELINE.md](/Users/aykumar/personal/kyklos/docs/ci/PIPELINE.md) - Overall CI pipeline design
- [ARTIFACTS.md](/Users/aykumar/personal/kyklos/docs/ci/ARTIFACTS.md) - Build artifacts strategy
- [../release/RELEASE-POLICY.md](/Users/aykumar/personal/kyklos/docs/release/RELEASE-POLICY.md) - Release management
- [../testing/TEST-STRATEGY.md](/Users/aykumar/personal/kyklos/docs/testing/TEST-STRATEGY.md) - Test strategy
