# Kyklos Project Brief

**Version:** 0.1.0 | **Last Updated:** 2025-10-19 | **Status:** Day 0 Planning

## Version Requirements
- **Project Version:** 0.1.0 (alpha)
- **API Version:** kyklos.io/v1alpha1
- **Kubernetes:** 1.25+ (tested on 1.28)
- **Go:** 1.21+ for building controller
- **Docker:** 24.0+ for building images

## Purpose
Kyklos is a Kubernetes operator that scales workloads based on time windows with timezone-aware scheduling.

## Goals
- Scale Deployments/StatefulSets up/down based on daily recurring time windows
- Support IANA timezone specification with DST transition handling
- Provide grace periods for scale-down operations
- Observable state transitions via metrics and status conditions
- Quick local verification (cluster up to visible scale change in under 15 minutes)

## Non-Goals for v0.1
- CronJob-style syntax or arbitrary cron expressions
- Advanced calendar features (recurring patterns, external calendar sync beyond ConfigMap)
- Multi-day or weekly patterns beyond daily recurrence
- Autoscaling integration (HPA/VPA)
- Cost estimation or recommendations
- Multi-cluster support

## Success Criteria (Verifiable in 15 Minutes Locally)
1. Create TimeWindowScaler CR with morning window (09:00-17:00 IST)
2. Target Deployment scales to activeReplicas=3 at window start
3. Target Deployment scales to inactiveReplicas=0 at window end
4. Status shows current state (Active/Inactive/GracePeriod)
5. Prometheus metrics expose current window state and scale events

## Out of Scope for v0.1
- Web UI or dashboard
- Slack/email notifications
- Multiple time windows per day (single window only)
- Permanent deletion or cascading cleanup
- Backup/restore integration
- GitOps workflow automation

## Assumptions
- Kubernetes 1.25+ cluster available
- Target workloads exist in same or different namespace
- Controller has RBAC to read/update target workloads
- System clock synchronized (NTP assumed)
- Targets support replica scaling (Deployment, StatefulSet, ReplicaSet)

## Constraints
- Single time window per TimeWindowScaler resource
- Time windows are daily recurring only
- DST transitions use Go time.Location standard library
- Controller runs as single replica (no HA in v0.1)
- Grace period maximum 60 minutes

## Glossary

**Active Window**: Time period when workload should be scaled to activeReplicas

**Inactive Window**: Time period outside active window when workload scales to inactiveReplicas

**Grace Period**: Delay before scale-down occurs, allowing workload to complete tasks

**DST Transition**: Daylight Saving Time clock changes that affect window timing

**Target Workload**: Kubernetes resource being scaled (Deployment, StatefulSet, ReplicaSet)

**IANA Timezone**: Standard timezone identifier (e.g., Asia/Kolkata, America/New_York)

**Requeue**: Controller pattern to re-check resource after calculated delay

**TimeWindowScaler (TWS)**: Custom Resource that defines scaling behavior

**Scale Subresource**: Kubernetes API for reading/writing replica counts

**windows[].replicas**: Desired replica count when this window is active (configured in spec)

**defaultReplicas**: Replica count when no windows match (often 2 for availability, not 0)

**effectiveReplicas**: The computed replica count right now (shown in status)

**pause**: When true, controller computes state but doesn't modify target workload

**crossMidnight**: Window spanning two calendar days (e.g., 22:00-02:00)

**windowStart**: HH:MM time when active window begins

**windowEnd**: HH:MM time when active window ends
