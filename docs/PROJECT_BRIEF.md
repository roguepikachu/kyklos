# Kyklos v0.1 Project Brief

**Version:** v0.1.0 | **Last Updated:** 2025-11-23 | **Status:** Implementation Phase

## Project Purpose
Kyklos is a Kubernetes operator that automatically scales Deployments based on time windows, enabling proactive resource management aligned with predictable usage patterns.

## Goals (v0.1)
- Scale Deployments based on daily recurring time windows with timezone awareness
- Handle holidays via ConfigMap with three modes (ignore, closed, open)
- Support cross-midnight windows and DST transitions automatically
- Provide graceful downscaling with configurable grace periods
- Offer manual drift correction and pause mode for operational control
- Export Prometheus metrics for observability

## Success Criteria (Verified Working)
1. Controller deploys successfully and reconciles TimeWindowScaler resources
2. Target deployments scale up/down based on time window matches
3. Status conditions reflect current state (Ready, Reconciling, Degraded)
4. Holiday ConfigMap integration works correctly
5. Grace period delays downscaling appropriately
6. Prometheus metrics expose scaling operations and window state
7. E2E test suite validates all core scenarios

## Current Status (83.8% Complete)

**COMPLETED:**
- Core time window engine with 83.8% test coverage
- CRD and API types fully defined
- Basic controller reconciliation loop
- Status updates and event emission
- Pause mode functionality
- Cross-midnight window support
- Manual drift detection (structure exists)
- Comprehensive documentation (83+ files)
- Example manifests and deployment configs

**IN PROGRESS:**
- Grace period timing logic (structure exists, timing implementation needed)
- Holiday ConfigMap reading (TODO in controller line 198)
- Metrics implementation (Prometheus integration)
- E2E test scenarios (suite exists, needs custom tests)

## Implementation Scope (v0.1)

**IN SCOPE:**
- Deployment targets only
- Daily recurring windows with day-of-week filtering
- IANA timezone support with full DST handling
- Grace periods for downscaling (0-3600 seconds)
- Holiday modes via ConfigMap
- Status conditions (Ready, Reconciling, Degraded)
- Prometheus metrics for scaling operations
- Single controller replica (no HA)

**OUT OF SCOPE:**
- StatefulSet/ReplicaSet targets (future)
- Admission webhooks for validation (future)
- High availability controller deployment (future)
- Web UI or dashboard
- Cost estimation features
- Multi-cluster support

## Non-Goals for v0.1
- CronJob-style cron expressions
- Advanced calendar integrations beyond ConfigMap
- HPA/VPA integration
- Webhook validation (CRD validation only)
- Multiple controller replicas

## Key Constraints
- Kubernetes 1.25+ required
- Single controller replica (no leader election)
- Grace period maximum: 3600 seconds (1 hour)
- System clock assumed NTP-synchronized
- Controller must have RBAC for target namespace(s)

## Dependencies
- Go 1.23.0+
- Kubernetes 1.25-1.31
- Docker 17.03+
- kubectl 1.25.0+
- controller-runtime v0.18+
- Prometheus (optional, for metrics)

## Timeline to v0.1 Release
**Target:** 2025-11-27 23:59 IST (4 days from now)

## Quality Gates
1. All unit tests pass (including engine 83.8% coverage maintained)
2. E2E test suite validates core scenarios
3. Manual testing confirms holiday ConfigMap integration
4. Metrics endpoint accessible and exposes expected metrics
5. Documentation reflects actual implementation
6. No critical bugs or security issues

## Open Questions
- Should grace period be cancellable if window becomes active again?
- Do we emit warning events on DST transition dates?
- Should we support ReadOnly mode for testing without ConfigMap?

## Glossary Reference
See /Users/aykumar/personal/kyklos/docs/user/GLOSSARY.md for complete terminology definitions.
