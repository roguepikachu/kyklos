# RACI Matrix - Kyklos v0.1

**Legend:**
- **R** = Responsible (does the work)
- **A** = Accountable (final approval, one per item)
- **C** = Consulted (provides input)
- **I** = Informed (kept in loop)

## API and CRD Design

| Task | api-crd-designer | api-validation-defaults-designer | controller-reconcile-designer | kyklos-orchestrator |
|------|------------------|----------------------------------|-------------------------------|---------------------|
| CRD schema definition | R,A | C | C | I |
| Field naming and types | R,A | C | C | I |
| OpenAPI validation rules | R | A | C | I |
| Status subresource design | R,A | I | C | I |
| API versioning strategy | R | I | I | A |

## Validation and Defaults

| Task | api-validation-defaults-designer | api-crd-designer | security-rbac-designer | kyklos-orchestrator |
|------|----------------------------------|------------------|------------------------|---------------------|
| Default value logic | R,A | C | I | I |
| Time format validation | R,A | C | I | I |
| Timezone validation | R,A | C | I | I |
| Cross-field validation | R,A | C | I | I |
| Admission webhook design | R,A | I | C | I |

## Reconcile Loop Design

| Task | controller-reconcile-designer | api-crd-designer | observability-metrics-designer | kyklos-orchestrator |
|------|-------------------------------|------------------|--------------------------------|---------------------|
| State machine logic | R,A | C | C | I |
| Requeue timing calculation | R,A | I | C | I |
| Error handling strategy | R,A | I | C | I |
| Target scaling logic | R,A | I | I | I |
| Status updates | R,A | C | C | I |

## Observability and Metrics

| Task | observability-metrics-designer | controller-reconcile-designer | api-crd-designer | kyklos-orchestrator |
|------|--------------------------------|-------------------------------|------------------|---------------------|
| Prometheus metrics definition | R,A | C | I | I |
| Status condition types | R,A | C | C | I |
| Event emission strategy | R,A | C | I | I |
| Logging standards | R,A | C | I | I |
| Debug mode design | R,A | C | I | I |

## Security and RBAC

| Task | security-rbac-designer | api-crd-designer | controller-reconcile-designer | kyklos-orchestrator |
|------|------------------------|------------------|-------------------------------|---------------------|
| Controller RBAC roles | R,A | I | C | I |
| Cross-namespace policy | R,A | C | C | I |
| ServiceAccount setup | R,A | I | I | I |
| Admission webhook auth | R,A | I | C | I |
| Least privilege audit | R,A | I | C | I |

## Local Developer Workflow

| Task | local-workflow-designer | testing-strategy-designer | ci-release-designer | kyklos-orchestrator |
|------|-------------------------|---------------------------|---------------------|---------------------|
| kind/minikube setup | R,A | C | C | I |
| Quick start script | R,A | I | C | I |
| Sample manifests | R,A | C | I | I |
| Troubleshooting guide | R,A | C | I | I |
| Time-warp testing method | R,A | A | I | I |

## Testing Strategy

| Task | testing-strategy-designer | controller-reconcile-designer | local-workflow-designer | kyklos-orchestrator |
|------|---------------------------|-------------------------------|-------------------------|---------------------|
| Unit test plan | R,A | C | I | I |
| Envtest scenarios | R,A | C | C | I |
| E2E test design | R,A | C | C | I |
| DST edge case tests | R,A | C | I | I |
| Mock time strategy | R,A | C | C | I |

## CI and Release

| Task | ci-release-designer | testing-strategy-designer | security-rbac-designer | kyklos-orchestrator |
|------|---------------------|---------------------------|------------------------|---------------------|
| GitHub Actions setup | R,A | C | I | I |
| Container image build | R,A | I | C | I |
| Release versioning | R,A | I | I | I |
| Artifact publishing | R,A | I | C | I |
| Pre-merge checks | R,A | C | I | I |

## Documentation and Developer Experience

| Task | docs-dx-designer | api-crd-designer | local-workflow-designer | kyklos-orchestrator |
|------|------------------|------------------|-------------------------|---------------------|
| README quick start | R,A | C | C | I |
| Concepts guide | R,A | C | C | I |
| API reference | R | A | I | I |
| Examples directory | R,A | C | C | I |
| Troubleshooting guide | R,A | C | C | I |

## Community and Launch Preparation

| Task | community-launch-designer | docs-dx-designer | kyklos-orchestrator | all-agents |
|------|---------------------------|------------------|---------------------|------------|
| Announcement blog post | R,A | C | C | I |
| Demo video script | R,A | C | C | I |
| FAQ preparation | R,A | C | C | I |
| Issue templates | R,A | I | C | I |
| Contributing guide | R,A | C | C | I |

## Demos and Screenshots

| Task | demo-screenshot-designer | local-workflow-designer | docs-dx-designer | kyklos-orchestrator |
|------|--------------------------|-------------------------|------------------|---------------------|
| Demo scenario design | R,A | C | C | I |
| Screenshot capture | R,A | C | C | I |
| Terminal recording | R,A | C | C | I |
| Diagram creation | R,A | I | C | I |
| Demo script writing | R,A | C | C | I |

## Overall Project Governance

| Task | kyklos-orchestrator | all-agents |
|------|---------------------|------------|
| Scope lock decisions | A | C |
| Timeline adjustments | A | C |
| Conflict resolution | A | I |
| Quality gate approval | A | R |
| ADR creation | A | R |
| Brief updates | A | C |

## Notes on RACI Usage

1. **Single Accountable:** Each task has exactly one A. This person/agent approves completion.
2. **Consulted Loop:** C agents must be given opportunity to review before A approves.
3. **Handoff Protocol:** When R completes work, they notify A. A validates against quality gate, then notifies I.
4. **Conflict Resolution:** If R and A disagree, escalate to kyklos-orchestrator per COMMUNICATION.md.
5. **Agent Availability:** If an agent is blocked, they must notify A and kyklos-orchestrator within 4 hours.
