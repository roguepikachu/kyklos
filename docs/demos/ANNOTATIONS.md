# Demo Annotations and Captions

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** demo-scenario-designer

## Purpose

This document provides consistent text annotations, captions, and overlay language for screenshots, videos, presentations, and documentation. Ensures clear, accurate communication of Kyklos concepts.

---

## Annotation Principles

### Consistency
- Use same terminology across all materials
- Follow Kubernetes conventions (e.g., "replica" not "instance")
- Maintain consistent capitalization (e.g., "TimeWindowScaler" not "Time Window Scaler")

### Clarity
- Keep captions short (< 10 words when possible)
- Use active voice ("Scales to 2 replicas" not "2 replicas are scaled")
- Avoid jargon unless essential (prefer "window" over "time-bounded scheduling interval")

### Accuracy
- All technical terms must match CRD field names
- Timestamps must show timezone (UTC or Europe/Berlin)
- Replica counts must be exact (don't round or approximate)

---

## Standard Terminology

### Core Concepts

**TimeWindowScaler (TWS)**
- Full name: TimeWindowScaler
- Short name: TWS (in CLI contexts)
- NOT: "Time Window Scaler", "time-window-scaler", "Scaler"

**Window States**
- BusinessHours: Active work hours window
- OffHours: Outside any defined windows
- NightShift: Evening/overnight window
- NOT: "business-hours", "off-hours", "night-shift"

**Replica Terminology**
- Current replicas: "2 replicas"
- Scaling action: "Scaling from 0 to 2 replicas"
- Target state: "Desired: 2 replicas"
- NOT: "instances", "pods" (unless specifically referring to Pod resources)

**Time References**
- Always include timezone: "14:38 UTC" or "22:00 CET"
- Use 24-hour format: "14:00" not "2:00 PM"
- Window boundaries: "14:38-14:41 UTC"
- NOT: AM/PM format in technical contexts

---

## Screenshot Annotations

### Annotation Placement

**Do NOT annotate directly on screenshots.** Keep originals clean. Create annotated versions separately.

**Annotation Layers:**
1. Background: Original screenshot
2. Highlight layer: Boxes/circles around important elements
3. Annotation layer: Text labels with arrows
4. Overlay layer: Title/explanation at top or bottom

---

### Visual Annotation Styles

**Highlight Boxes:**
```
Style: 3px solid line
Color: Yellow (#FFD700) or Red (#FF4444) for emphasis
Corner radius: 4px
No fill (transparent background)
```

**Arrows:**
```
Style: 2px solid line with arrowhead
Color: Yellow (#FFD700)
Curve: Slight curve (for aesthetics, not sharp angles)
```

**Text Labels:**
```
Font: Sans-serif (Helvetica, Arial, Roboto)
Size: 24pt for main labels, 18pt for details
Color: White text
Background: Semi-transparent black (#000000 80% opacity)
Padding: 8px around text
Border radius: 4px
```

**Example Annotation Workflow:**
1. Take clean screenshot
2. Open in image editor (Photoshop, Figma, GIMP, or online tool)
3. Add highlight box around element of interest (e.g., WINDOW column)
4. Add arrow pointing from box to empty area
5. Add text label at arrow endpoint
6. Export as PNG with `-annotated` suffix

---

### Common Screenshot Annotations

#### Initial State (0/0 Replicas)
```
HIGHLIGHT: Deployment line showing "0/0" in READY column
LABEL: "Starting state: No replicas running"
```

#### TWS Created (OffHours)
```
HIGHLIGHT: WINDOW column showing "OffHours"
LABEL: "Outside defined windows → defaultReplicas (0)"
```

#### Window Opens (BusinessHours)
```
HIGHLIGHT: WINDOW column showing "BusinessHours"
LABEL: "Window opened at 14:38 UTC → Scaling to 2 replicas"
```

#### Pods Running
```
HIGHLIGHT: STATUS column showing "Running"
LABEL: "Pods ready to serve traffic"
```

#### Events Timeline
```
HIGHLIGHT: WindowTransition and ScalingTarget events
LABEL: "Kyklos events show scaling decisions"
```

#### Controller Logs
```
HIGHLIGHT: "Matched window: BusinessHours (14:38-14:41) -> 2 replicas" line
LABEL: "Controller explains every scaling decision"
```

#### Cross-Midnight (Key Moment)
```
HIGHLIGHT: Date showing "Wed Oct 29 00:01 CET" AND WINDOW showing "NightShift"
LABEL: "Day changed but window remains active → Cross-midnight works!"
```

---

## Video Captions/Overlays

### Title Cards

**Opening Title (Video 1):**
```
Kyklos Time Window Scaler
Kubernetes deployments that scale on schedule
```

**Opening Title (Video 2):**
```
Kyklos Cross-Midnight Windows
Windows that span calendar day boundaries
```

**Opening Title (Video 3):**
```
Kyklos Quick Start
From zero to scaling in 3 minutes
```

---

### Concept Explanations

**What is Kyklos?**
```
Kyklos scales Kubernetes deployments based on time windows

Define business hours → Replicas scale automatically
Outside hours → Scale to zero → Save resources
```

**Time Windows Explained:**
```
Time Window = When + How Many

start: "09:00"  ← Window opens
end: "17:00"    ← Window closes
replicas: 5     ← Desired replica count

timezone: UTC  ← All times in specified timezone
```

**Default Replicas:**
```
defaultReplicas: 0

Replica count when NO windows match
Typically set to 0 for maximum cost savings
```

**Cross-Midnight Concept:**
```
Cross-Midnight Window:
start > end  →  Window spans midnight

Example:
start: "22:00"  (Friday evening)
end: "02:00"    (Saturday morning)

Window active from Fri 22:00 to Sat 02:00
```

---

### Action Captions

**Scale-Up:**
```
Window opened → Scaling to 2 replicas
```

```
Scaling from 0 to 2 replicas
```

```
Creating pods...
```

```
Pods running → Ready to serve traffic
```

**Scale-Down:**
```
Window closed → Scaling to 0 replicas
```

```
Scaling from 2 to 0 replicas
```

```
Terminating pods...
```

```
Back to 0 replicas → No resources used
```

**Midnight Transition:**
```
Friday 23:59 → Window active, 3 replicas
```

```
Midnight crosses...
```

```
Saturday 00:01 → Window STILL active, 3 replicas
Day changed but window continues!
```

```
No scaling at midnight → Cross-midnight works correctly
```

---

### Status Captions

**Before Window:**
```
Current: OffHours
Replicas: 0
Status: Waiting for window to open
```

**During Window:**
```
Current: BusinessHours
Replicas: 2
Status: Window active, serving traffic
```

**After Window:**
```
Current: OffHours
Replicas: 0
Status: Window closed, resources reclaimed
```

---

### Educational Captions

**How It Works:**
```
1. Controller watches TimeWindowScalers
2. Evaluates current time vs windows
3. Computes desired replica count
4. Updates target Deployment
5. Schedules next evaluation
```

**Use Cases:**
```
✓ Business hours scaling (9-5 weekdays)
✓ Night batch processing (22:00-06:00)
✓ Weekend capacity reduction
✓ Timezone-aware global deployments
✓ Cost optimization via zero-scaling
```

**Benefits:**
```
Predictable scaling based on time
No metrics or thresholds needed
Timezone and DST aware
Works with any Kubernetes cluster
Simple YAML configuration
```

---

## Presentation Slides Annotations

### Slide 1: Title Slide
```
TITLE: Kyklos Time Window Scaler
SUBTITLE: Time-based autoscaling for Kubernetes
FOOTER: github.com/your-org/kyklos
```

### Slide 2: The Problem
```
TITLE: The Cost of Always-On Services

❌ Dev environments running 24/7
❌ Batch jobs with idle periods
❌ Regional services outside business hours
❌ Weekend over-provisioning

RESULT: Wasted compute resources
```

### Slide 3: The Solution
```
TITLE: Kyklos Time Window Scaler

✓ Define time windows in YAML
✓ Automatic scaling at boundaries
✓ Timezone-aware (DST support)
✓ Zero replicas outside windows

RESULT: 60-80% cost reduction for scheduled workloads
```

### Slide 4: How It Works
```
DIAGRAM:
[TimeWindowScaler] → [Kyklos Controller] → [Deployment]
                      ↓
                  [Time Evaluation]
                      ↓
                  [Scale Action]

CAPTION: Controller reconciles at window boundaries
```

### Slide 5: Example Configuration
```
CODE SNIPPET:
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
spec:
  windows:
  - days: [Mon, Tue, Wed, Thu, Fri]
    start: "09:00"
    end: "17:00"
    replicas: 5
  defaultReplicas: 0

CAPTION: Business hours: 5 replicas | Nights/weekends: 0 replicas
```

### Slide 6: Live Demo
```
TITLE: See It In Action

[Embed video or link to recording]

CAPTION: Watch a deployment scale from 0→2→0 in 2 minutes
```

### Slide 7: Advanced Features
```
TITLE: Beyond Basic Scheduling

✓ Cross-midnight windows (22:00-02:00)
✓ Grace periods for smooth scale-downs
✓ Holiday handling via ConfigMaps
✓ Pause/resume without deleting resources
✓ Prometheus metrics integration

CAPTION: Production-ready features for complex scenarios
```

### Slide 8: Getting Started
```
TITLE: Quick Start

1. Install CRD and controller
   kubectl apply -f https://...

2. Create TimeWindowScaler
   kubectl apply -f my-scaler.yaml

3. Watch it work
   kubectl get tws --watch

FOOTER: Full docs at kyklos.io/docs
```

---

## Blog Post / README Captions

### Hero Image Caption
```
Kyklos automatically scales Kubernetes deployments based on time windows,
reducing costs by up to 80% for scheduled workloads.
```

### Architecture Diagram Caption
```
Figure 1: Kyklos Controller watches TimeWindowScaler resources and updates
target Deployments at window boundaries. All time calculations are timezone-aware.
```

### Screenshot: Scale-Up Event
```
Figure 2: When a time window opens, Kyklos scales the deployment from 0 to 2
replicas. Events show the complete scaling decision chain.
```

### Screenshot: Cross-Midnight
```
Figure 3: Cross-midnight windows remain active after the calendar day changes.
This example shows a window from Friday 22:00 to Saturday 02:00 with no
disruption at midnight.
```

### Code Example Caption
```
Example 1: A simple TimeWindowScaler that scales an nginx deployment to 5
replicas during business hours (9 AM - 5 PM weekdays) and 0 replicas otherwise.
```

---

## Social Media Annotations

### Twitter/X Post
```
🕐 Kyklos Time Window Scaler for Kubernetes

📅 Define time windows in YAML
⚙️ Automatic scaling at boundaries
🌍 Timezone + DST aware
💰 Save 60-80% on scheduled workloads

Watch it work: [video link]
Docs: [link]
GitHub: [link]

#kubernetes #cloudnative #costoptimization
```

### LinkedIn Post
```
Introducing Kyklos Time Window Scaler

Kyklos brings time-based autoscaling to Kubernetes, enabling:

✅ Business hours scaling (9-5 weekdays)
✅ Zero-replica cost savings (nights, weekends)
✅ Cross-midnight window support
✅ Timezone-aware scheduling (DST handled automatically)

Perfect for:
• Development environments
• Batch processing jobs
• Regional services
• Weekend capacity reduction

See it in action in our 2-minute demo: [link]

Built on proven Kubernetes patterns, Kyklos integrates seamlessly with your
existing infrastructure. Open source and production-ready.

Learn more: [docs link]
GitHub: [repo link]

#Kubernetes #CloudNative #CostOptimization #OpenSource #DevOps
```

### Reddit r/kubernetes Post
```
Title: [Project] Kyklos - Time-based autoscaling for Kubernetes

We built Kyklos to solve a simple problem: dev environments running 24/7 when
they're only used 9-5 weekdays.

What it does:
- Scales deployments based on time windows (not metrics)
- Handles cross-midnight windows (e.g., 22:00-02:00)
- Timezone-aware with automatic DST handling
- Can scale to zero outside windows for maximum cost savings

Example use cases:
- Business hours scaling (9 AM - 5 PM → N replicas, else 0)
- Night batch jobs (22:00-06:00 → N replicas)
- Regional services (scale up during local business hours)

2-minute demo: [link]
GitHub: [link]
Docs: [link]

Built using Kubebuilder, follows Kubernetes best practices. Looking for
feedback on the approach and any use cases we might have missed!
```

---

## Error Message Annotations

When showing error scenarios in demos or documentation:

### Invalid Timezone Error
```
SCREENSHOT: Controller logs showing timezone error

CAPTION:
Error: Invalid timezone "Mars/OlympusMons"
Fix: Use valid IANA timezone (e.g., "America/New_York")
Verify: TZ=America/New_York date
```

### Target Not Found Error
```
SCREENSHOT: TWS status showing TargetNotFound condition

CAPTION:
Error: Target deployment "webapp-demo" not found in namespace "demo"
Fix: Create deployment first, or correct targetRef.name
Verify: kubectl get deploy -n demo
```

### Permission Error
```
SCREENSHOT: Controller logs showing RBAC error

CAPTION:
Error: Controller cannot update deployment (forbidden)
Fix: Verify RBAC permissions with make verify-rbac
Check: kubectl auth can-i update deployments --as=system:serviceaccount:kyklos-system:kyklos-controller
```

---

## Glossary Annotations

For documentation and tooltips:

**TimeWindowScaler**
```
A Kubernetes custom resource that defines when and how to scale a Deployment
based on time windows.

Example: Scale to 5 replicas during business hours, 0 replicas otherwise.
```

**Window**
```
A time range (start to end) during which a specific replica count is desired.

Windows can cross midnight (e.g., 22:00-02:00) and are timezone-aware.
```

**defaultReplicas**
```
The replica count used when no windows match the current time.

Typically set to 0 for maximum cost savings, or to a minimum viable count for
always-on services.
```

**gracePeriodSeconds**
```
Delay (in seconds) before applying scale-down actions.

Useful for gradual traffic reduction or connection draining.
Scale-up actions are immediate; grace period applies only to scale-down.
```

**Timezone**
```
IANA timezone identifier (e.g., "America/New_York", "Europe/Berlin").

All window times are interpreted in this timezone. Controller automatically
handles Daylight Saving Time (DST) transitions.
```

**Cross-Midnight Window**
```
A window where the end time is earlier than the start time, indicating the
window spans the midnight boundary.

Example: start: "22:00", end: "02:00" means 22:00 today to 02:00 tomorrow.
```

---

## Accessibility Annotations

For screen readers and accessibility compliance:

### Alt Text for Screenshots

**Initial State Screenshot:**
```
Alt text: Terminal screenshot showing kubectl get deploy output.
Deployment named webapp-demo has zero replicas, zero pods available.
READY column shows 0/0.
```

**Active Window Screenshot:**
```
Alt text: Terminal screenshot showing kubectl get tws output.
TimeWindowScaler named webapp-minute-scaler shows WINDOW column as BusinessHours,
REPLICAS column as 2, TARGET column as webapp-demo. Below it, deployment shows
READY 2/2 with two pods in Running status.
```

**Cross-Midnight Screenshot:**
```
Alt text: Terminal screenshot showing date command output "Wed Oct 29 00:01 CET"
followed by kubectl get tws output. TimeWindowScaler shows WINDOW as NightShift
and REPLICAS as 3, demonstrating window remains active after midnight on Wednesday.
```

### ARIA Labels for Interactive Elements

**Demo video player:**
```
aria-label="Kyklos time window scaler demonstration video showing deployment
scaling from zero to two replicas and back to zero"
```

**Code example:**
```
aria-label="YAML configuration example for TimeWindowScaler with business hours
window from 9 AM to 5 PM, Monday through Friday"
```

---

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2025-10-28 | 1.0 | Initial annotations and caption guidelines |
