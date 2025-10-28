# Video Shotlist and Recording Plan

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** demo-scenario-designer

## Purpose

This document provides a complete shot-by-shot plan for recording promotional and educational videos for Kyklos. Optimized for 2-3 minute silent recordings with text overlays added in post-production.

---

## Video Specifications

### Technical Requirements

**Video Format:**
- Resolution: 1920x1080 (1080p)
- Frame rate: 30fps
- Format: MP4 (H.264 codec)
- Audio: None (silent video, text overlays only)
- Duration: 2-3 minutes per video

**Terminal Configuration:**
- Font: Monaco or Menlo, 16pt
- Size: 120 columns x 35 rows
- Color scheme: Dark background with high-contrast colors
- Prompt: Simple `$ ` or `>` without hostname

**Recording Setup:**
- Screen recording: Full screen or terminal-only
- Cursor: Visible and highlighted
- Typing speed: Real-time (with post-edit speedup if needed)
- Transitions: Natural command flow, no jump cuts

---

## Video 1: "Kyklos in 2 Minutes" (Overview)

**Target Duration:** 2:00 minutes
**Audience:** First-time viewers, decision makers
**Goal:** Show Kyklos scaling from 0→2→0 with minimal explanation

### Shot Breakdown

**Shot 1: Title Card (0:00-0:05) - 5 seconds**
```
TEXT OVERLAY:
Kyklos Time Window Scaler
Kubernetes deployments that scale on schedule

[Show static terminal with clean prompt]
```

**NO TERMINAL ACTIVITY** - Just show clean state with overlay text.

---

**Shot 2: Environment Setup (0:05-0:15) - 10 seconds**
```
TERMINAL COMMANDS:
$ kubectl get deploy,pods -n demo

OUTPUT:
NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   0/0     0            0           10s

No pods in demo namespace.

TEXT OVERLAY:
Starting state: 0 replicas
```

**Recording Notes:**
- Show command typed in real-time
- Let output display fully
- 2-second pause before next shot

---

**Shot 3: Show TimeWindowScaler (0:15-0:30) - 15 seconds**
```
TERMINAL COMMANDS:
$ cat demo-scaler.yaml

OUTPUT:
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: webapp-minute-scaler
  namespace: demo
spec:
  targetRef:
    kind: Deployment
    name: webapp-demo
  timezone: UTC
  defaultReplicas: 0
  windows:
  - days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
    start: "14:38"
    end: "14:41"
    replicas: 2

TEXT OVERLAY (at 0:18):
Define time windows in YAML
Window active: 14:38-14:41 UTC → 2 replicas
```

**Recording Notes:**
- Scroll slowly through YAML
- Pause on `windows:` section
- Highlight (via overlay in post) the start/end/replicas lines

---

**Shot 4: Apply TimeWindowScaler (0:30-0:40) - 10 seconds**
```
TERMINAL COMMANDS:
$ kubectl apply -f demo-scaler.yaml

OUTPUT:
timewindowscaler.kyklos.io/webapp-minute-scaler created

$ kubectl get tws -n demo

OUTPUT:
NAME                    WINDOW     REPLICAS   TARGET        AGE
webapp-minute-scaler    OffHours   0          webapp-demo   2s

TEXT OVERLAY:
Kyklos controller is now watching
```

**Recording Notes:**
- Show immediate kubectl get tws output
- Emphasize OffHours and 0 replicas

---

**Shot 5: Watch Command Setup (0:40-0:45) - 5 seconds**
```
TERMINAL COMMANDS:
$ watch -n 1 'kubectl get tws,deploy,pods -n demo'

TEXT OVERLAY:
Waiting for window to open...
```

**Recording Notes:**
- Show watch command starting
- Display initial OffHours state
- Clear indication we're now monitoring

---

**Shot 6: Scale-Up Transition (0:45-1:05) - 20 seconds**
```
WATCH OUTPUT (showing changes):

[At 0:45 - Before window]
NAME                                    WINDOW     REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-...  OffHours   0          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   0/0     0            0           50s

No pods in demo namespace.

[At 0:50 - Window opens]
NAME                                    WINDOW          REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-...  BusinessHours   2          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   0/2     2            0           55s

NAME                              READY   STATUS              RESTARTS   AGE
pod/webapp-demo-7d8f9c5b4-abc12   0/1     ContainerCreating   0          2s
pod/webapp-demo-7d8f9c5b4-def34   0/1     ContainerCreating   0          2s

[At 1:00 - Pods running]
NAME                                    WINDOW          REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-...  BusinessHours   2          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   2/2     2            2           1m5s

NAME                              READY   STATUS    RESTARTS   AGE
pod/webapp-demo-7d8f9c5b4-abc12   1/1     Running   0          12s
pod/webapp-demo-7d8f9c5b4-def34   1/1     Running   0          12s

TEXT OVERLAY (at 0:50):
Window opened → Scaling to 2 replicas

TEXT OVERLAY (at 1:00):
Pods running → Ready to serve traffic
```

**Recording Notes:**
- THIS IS THE KEY MOMENT
- Show natural timing of scale-up
- Don't speed up this section in post
- Let viewers see the progression naturally

---

**Shot 7: Steady State (1:05-1:15) - 10 seconds**
```
WATCH OUTPUT (stable):
[Same as final state from Shot 6]

TEXT OVERLAY:
Window active: 2 replicas serving traffic
```

**Recording Notes:**
- Show stable state for a few refreshes
- No changes during this period
- Emphasize stability

---

**Shot 8: Scale-Down Transition (1:15-1:30) - 15 seconds**
```
WATCH OUTPUT (showing changes):

[At 1:15 - Window closes]
NAME                                    WINDOW     REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-...  OffHours   0          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   2/2     0            2           4m5s

NAME                              READY   STATUS        RESTARTS   AGE
pod/webapp-demo-7d8f9c5b4-abc12   1/1     Terminating   0          3m2s
pod/webapp-demo-7d8f9c5b4-def34   1/1     Terminating   0          3m2s

[At 1:25 - Pods removed]
NAME                                    WINDOW     REPLICAS   TARGET
timewindowscaler.kyklos.io/webapp-...  OffHours   0          webapp-demo

NAME                          READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/webapp-demo   0/0     0            0           4m15s

No pods in demo namespace.

TEXT OVERLAY (at 1:15):
Window closed → Scaling to 0 replicas

TEXT OVERLAY (at 1:25):
Back to initial state → No resources used
```

**Recording Notes:**
- Show natural scale-down timing
- Highlight return to 0/0 state
- Emphasize resource reclamation

---

**Shot 9: Events Timeline (1:30-1:45) - 15 seconds**
```
TERMINAL COMMANDS:
[Press Ctrl+C to stop watch]

$ kubectl get events -n demo --sort-by='.lastTimestamp' | tail -10

OUTPUT:
LAST SEEN   TYPE     REASON              MESSAGE
3m45s       Normal   WindowTransition    Entered window: BusinessHours (14:38-14:41)
3m45s       Normal   ScalingTarget       Scaling webapp-demo from 0 to 2 replicas
3m44s       Normal   ScaledUp            Scaled up replica set to 2
45s         Normal   WindowTransition    Exited window: BusinessHours
45s         Normal   ScalingTarget       Scaling webapp-demo from 2 to 0 replicas
44s         Normal   ScaledDown          Scaled down replica set to 0

TEXT OVERLAY:
Events show complete scaling lifecycle
```

**Recording Notes:**
- Scroll through events slowly
- Let viewers read the key events
- Highlight WindowTransition and ScalingTarget events in post

---

**Shot 10: Controller Decision Logs (1:45-2:00) - 15 seconds**
```
TERMINAL COMMANDS:
$ kubectl logs -n kyklos-system -l app=kyklos-controller --tail=20 | grep webapp

OUTPUT:
INFO  Current time in UTC: 2025-10-28T14:38:05Z
INFO  Matched window: BusinessHours (14:38-14:41) -> 2 replicas
INFO  Scaling deployment webapp-demo from 0 to 2 replicas
INFO  Requeue scheduled at next window boundary: 2025-10-28T14:41:00Z
...
INFO  Current time in UTC: 2025-10-28T14:41:05Z
INFO  No matching windows, using defaultReplicas: 0
INFO  Scaling deployment webapp-demo from 2 to 0 replicas

TEXT OVERLAY:
Controller makes intelligent scaling decisions
Predictable requeue scheduling
```

**Recording Notes:**
- Show decision logic clearly
- Emphasize deterministic behavior
- Highlight requeue scheduling

---

**Shot 11: Closing Card (2:00-2:05) - 5 seconds**
```
TEXT OVERLAY:
Kyklos Time Window Scaler
github.com/your-org/kyklos
Time-based autoscaling for Kubernetes

[Show static terminal or fade to black]
```

**NO TERMINAL ACTIVITY** - Closing credits.

---

### Post-Production Notes (Video 1)

**Editing Checklist:**
- [ ] Speed up typing where appropriate (1.5-2x)
- [ ] Keep scale transitions at real-time speed
- [ ] Add text overlays at specified timestamps
- [ ] Highlight key YAML lines (windows section) with box or arrow
- [ ] Highlight key events in event listing
- [ ] Add subtle zoom-in on critical moments (window opening, scaling)
- [ ] Ensure all text is readable at 1080p
- [ ] Add fade-in at start, fade-out at end
- [ ] Export as MP4, H.264, 30fps, 1080p

**Text Overlay Style:**
- Font: Sans-serif, clean (Helvetica, Arial, or Roboto)
- Size: 48pt for main text, 36pt for details
- Color: White with 80% black background bar (for readability)
- Position: Bottom third of screen
- Duration: 3-5 seconds per overlay
- Transition: Fade in/out over 0.5 seconds

---

## Video 2: "Cross-Midnight Windows" (Advanced Feature)

**Target Duration:** 2:30 minutes
**Audience:** Advanced users, technical evaluators
**Goal:** Demonstrate cross-midnight window behavior convincingly

### Shot Breakdown

**Shot 1: Title Card (0:00-0:05) - 5 seconds**
```
TEXT OVERLAY:
Kyklos Cross-Midnight Windows
Windows that span calendar day boundaries

[Show static terminal]
```

---

**Shot 2: Window Specification (0:05-0:25) - 20 seconds**
```
TERMINAL COMMANDS:
$ cat nightshift-scaler.yaml

OUTPUT:
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: nightshift-scaler
spec:
  timezone: Europe/Berlin
  defaultReplicas: 0
  windows:
  - days: [Fri]
    start: "22:00"  # Friday 10:00 PM
    end: "02:00"    # Saturday 2:00 AM
    replicas: 3

TEXT OVERLAY (at 0:10):
Window from 22:00 to 02:00
end < start → crosses midnight

TEXT OVERLAY (at 0:18):
Spans Friday night into Saturday morning
```

**Recording Notes:**
- Slow scroll to show full spec
- Emphasize start/end times
- Show diagram overlay of timeline (Friday 22:00 → Saturday 02:00)

---

**Shot 3: Pre-Midnight State (0:25-0:40) - 15 seconds**
```
TERMINAL COMMANDS:
$ TZ=Europe/Berlin date

OUTPUT:
Fri Oct 28 23:30:00 CET 2025

$ kubectl get tws,deploy,pods -n demo

OUTPUT:
NAME                                  WINDOW       REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift NightShift   3          nightshift-demo

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nightshift-demo   3/3     3            3           30m

NAME                                  READY   STATUS    RESTARTS   AGE
pod/nightshift-demo-7f8c9b5d-abc12    1/1     Running   0          30m
pod/nightshift-demo-7f8c9b5d-def34    1/1     Running   0          30m
pod/nightshift-demo-7f8c9b5d-ghi56    1/1     Running   0          30m

TEXT OVERLAY:
Friday 23:30 CET → Window active, 3 replicas
```

**Recording Notes:**
- Clearly show Friday date
- Show stable window state
- Emphasize it's late evening

---

**Shot 4: Midnight Transition (0:40-1:00) - 20 seconds**
```
TERMINAL COMMANDS:
$ watch -n 2 'TZ=Europe/Berlin date && echo && kubectl get tws,deploy,pods -n demo'

WATCH OUTPUT:

[At 0:40 - Before midnight]
Fri Oct 28 23:58:00 CET 2025

NAME                                  WINDOW       REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift NightShift   3          nightshift-demo
[... deployment and pods as before ...]

[At 0:50 - After midnight]
Sat Oct 29 00:01:00 CET 2025

NAME                                  WINDOW       REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift NightShift   3          nightshift-demo

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nightshift-demo   3/3     3            3           31m

NAME                                  READY   STATUS    RESTARTS   AGE
pod/nightshift-demo-7f8c9b5d-abc12    1/1     Running   0          31m
pod/nightshift-demo-7f8c9b5d-def34    1/1     Running   0          31m
pod/nightshift-demo-7f8c9b5d-ghi56    1/1     Running   0          31m

TEXT OVERLAY (at 0:50):
Saturday 00:01 CET → Window STILL active
No scaling at midnight → Day transition handled
```

**Recording Notes:**
- THIS IS THE CRITICAL MOMENT
- Show clear day change (Fri → Sat)
- Show window remains NightShift
- Show replicas remain 3
- Emphasize NO CHANGE at midnight
- Use visual highlight (circle or box) around date and window status in post

---

**Shot 5: Post-Midnight Explanation (1:00-1:20) - 20 seconds**
```
TERMINAL COMMANDS:
[Press Ctrl+C to stop watch]

$ kubectl get tws nightshift-scaler -n demo -o jsonpath='{.status.windowMetadata}' | jq

OUTPUT:
{
  "crossesMidnight": true,
  "currentDay": "Saturday",
  "windowStartDay": "Friday",
  "windowEndDay": "Saturday",
  "activeFor": "2h 1m",
  "remainingTime": "59m"
}

TEXT OVERLAY:
Window metadata shows cross-day state
Started Friday, ends Saturday
```

**Recording Notes:**
- Show JSON formatting clearly
- Highlight crossesMidnight: true
- Emphasize day tracking

---

**Shot 6: Window End (1:20-1:40) - 20 seconds**
```
TERMINAL COMMANDS:
$ watch -n 2 'TZ=Europe/Berlin date && echo && kubectl get tws,deploy,pods -n demo'

WATCH OUTPUT:

[At 1:20 - Before window end]
Sat Oct 29 01:58:00 CET 2025

NAME                                  WINDOW       REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift NightShift   3          nightshift-demo
[... deployment and pods as before ...]

[At 1:30 - After window end]
Sat Oct 29 02:01:00 CET 2025

NAME                                  WINDOW     REPLICAS   TARGET
timewindowscaler.kyklos.io/nightshift OffHours   0          nightshift-demo

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nightshift-demo   0/0     0            0           3h1m

No pods in demo namespace.

TEXT OVERLAY (at 1:30):
Saturday 02:01 CET → Window ended on schedule
Scaled to 0 replicas → Cross-midnight handled correctly
```

**Recording Notes:**
- Show window closing at 02:00 Saturday
- Emphasize correct end day (Saturday)
- Show return to OffHours and 0 replicas

---

**Shot 7: Controller Cross-Midnight Logic (1:40-2:00) - 20 seconds**
```
TERMINAL COMMANDS:
$ kubectl logs -n kyklos-system -l app=kyklos-controller --tail=30 | grep -A 5 "cross-midnight"

OUTPUT:
INFO  Evaluating cross-midnight window  {"start": "22:00", "end": "02:00", "currentDay": "Friday"}
INFO  Cross-midnight calculation: window extends into next day
INFO  Matched window: NightShift (22:00 Fri → 02:00 Sat) -> 3 replicas
...
INFO  Cross-midnight window ended  {"start": "22:00 Fri", "end": "02:00 Sat", "duration": "4h"}

TEXT OVERLAY:
Controller explicitly handles cross-midnight logic
Calculates boundaries spanning days
```

**Recording Notes:**
- Show controller awareness of cross-midnight
- Highlight calculation mentions
- Emphasize deterministic behavior

---

**Shot 8: Use Cases (2:00-2:20) - 20 seconds**
```
TEXT OVERLAY SEQUENCE (no terminal, just text):

Cross-Midnight Use Cases:

• Night shifts: 22:00 → 06:00
• Batch processing: 23:00 → 04:00
• Backup windows: 01:00 → 05:00
• Off-peak hours: 20:00 → 08:00

Kyklos handles DST transitions automatically
```

**Recording Notes:**
- Show text overlay on dark background or blurred terminal
- Use bullet points with icons
- 5 seconds per use case section

---

**Shot 9: Closing Card (2:20-2:30) - 10 seconds**
```
TEXT OVERLAY:
Kyklos Time Window Scaler
Cross-midnight windows work seamlessly
github.com/your-org/kyklos

[Fade to black]
```

---

### Post-Production Notes (Video 2)

**Editing Checklist:**
- [ ] Add visual highlight around date change at midnight
- [ ] Add visual highlight around window status (remains active)
- [ ] Include timeline diagram showing Friday → Saturday transition
- [ ] Add boxed highlight for crossesMidnight: true in JSON
- [ ] Speed up watch transitions (2x) except midnight crossing (keep real-time)
- [ ] Add zoom-in on midnight transition moment
- [ ] Use split-screen to show before/after midnight side-by-side (optional)
- [ ] Add "No Change" graphic at midnight moment
- [ ] Ensure all text overlays readable at 1080p

**Special Visual Effects:**
- At midnight transition: Add subtle flash or color change to emphasize moment
- Add timeline graphic showing 22:00 Fri → 00:00 → 02:00 Sat
- Use contrasting colors for Friday vs Saturday in timeline

---

## Video 3: "Quick Start" (Fast-Paced Tutorial)

**Target Duration:** 3:00 minutes
**Audience:** Developers wanting to try Kyklos
**Goal:** Show complete setup to first scale event

### Shot Breakdown

**Shot 1: Title (0:00-0:05) - 5 seconds**
```
TEXT OVERLAY:
Kyklos Quick Start
From zero to scaling in 3 minutes
```

---

**Shot 2: Prerequisites (0:05-0:20) - 15 seconds**
```
TERMINAL COMMANDS:
$ kubectl cluster-info
$ kubectl get nodes

OUTPUT:
[Show cluster is running]

TEXT OVERLAY:
Prerequisites:
✓ Kubernetes cluster (Kind, k3d, or any K8s)
✓ kubectl configured
✓ 5 minutes of your time
```

---

**Shot 3: Install Kyklos (0:20-0:40) - 20 seconds**
```
TERMINAL COMMANDS:
$ kubectl apply -f https://raw.githubusercontent.com/your-org/kyklos/main/config/crd/bases/kyklos.io_timewindowscalers.yaml

OUTPUT:
customresourcedefinition.apiextensions.k8s.io/timewindowscalers.kyklos.io created

$ kubectl apply -f https://raw.githubusercontent.com/your-org/kyklos/main/config/default/deployment.yaml

OUTPUT:
namespace/kyklos-system created
serviceaccount/kyklos-controller created
deployment.apps/kyklos-controller-manager created

TEXT OVERLAY:
Install CRD and controller
```

**Recording Notes:**
- Show real URLs (replace with actual when available)
- Keep installation output concise

---

**Shot 4: Verify Installation (0:40-0:50) - 10 seconds**
```
TERMINAL COMMANDS:
$ kubectl get pods -n kyklos-system

OUTPUT:
NAME                                         READY   STATUS    RESTARTS   AGE
kyklos-controller-manager-7d9c8b5f4-abc12   1/1     Running   0          15s

TEXT OVERLAY:
Controller running → Ready to use
```

---

**Shot 5: Create Example Deployment (0:50-1:05) - 15 seconds**
```
TERMINAL COMMANDS:
$ kubectl create namespace demo

$ kubectl create deployment webapp-demo --image=nginx:alpine --replicas=0 -n demo

OUTPUT:
deployment.apps/webapp-demo created

TEXT OVERLAY:
Create target deployment with 0 replicas
```

---

**Shot 6: Create TimeWindowScaler (1:05-1:35) - 30 seconds**
```
TERMINAL COMMANDS:
$ kubectl apply -f - <<EOF
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: webapp-scaler
  namespace: demo
spec:
  targetRef:
    kind: Deployment
    name: webapp-demo
  timezone: UTC
  defaultReplicas: 0
  windows:
  - days: [Mon, Tue, Wed, Thu, Fri]
    start: "09:00"
    end: "17:00"
    replicas: 3
EOF

OUTPUT:
timewindowscaler.kyklos.io/webapp-scaler created

TEXT OVERLAY:
Define business hours: 9 AM - 5 PM weekdays
3 replicas during business hours, 0 otherwise
```

**Recording Notes:**
- Show YAML creation in real-time (don't speed this up)
- Let viewers see the structure
- Pause briefly after creation

---

**Shot 7: Check Status (1:35-1:50) - 15 seconds**
```
TERMINAL COMMANDS:
$ kubectl get tws webapp-scaler -n demo

OUTPUT:
NAME             WINDOW     REPLICAS   TARGET        AGE
webapp-scaler    OffHours   0          webapp-demo   10s

$ kubectl get deploy webapp-demo -n demo

OUTPUT:
NAME          READY   UP-TO-DATE   AVAILABLE   AGE
webapp-demo   0/0     0            0           1m

TEXT OVERLAY:
Outside business hours → 0 replicas
Window will open tomorrow at 9 AM
```

---

**Shot 8: Simulate Business Hours (1:50-2:30) - 40 seconds**
```
TERMINAL COMMANDS:
$ kubectl patch tws webapp-scaler -n demo --type=merge -p '{"spec":{"windows":[{"days":["Mon","Tue","Wed","Thu","Fri","Sat","Sun"],"start":"'$(date -u +%H:%M --date='1 minute')'","end":"'$(date -u +%H:%M --date='4 minutes')'","replicas":3}]}}'

OUTPUT:
timewindowscaler.kyklos.io/webapp-scaler patched

$ watch -n 1 'kubectl get tws,deploy,pods -n demo'

WATCH OUTPUT:
[Show transition from OffHours to BusinessHours]
[Show scaling from 0 to 3 replicas]
[Show pods reaching Running state]

TEXT OVERLAY (at 2:00):
Window opens → Scaling begins

TEXT OVERLAY (at 2:20):
3 replicas running → Ready for traffic
```

**Recording Notes:**
- Speed up watch output (1.5-2x)
- Keep scale transition visible
- Show full lifecycle

---

**Shot 9: Verify Events (2:30-2:45) - 15 seconds**
```
TERMINAL COMMANDS:
[Press Ctrl+C]

$ kubectl get events -n demo --sort-by='.lastTimestamp' | tail -5

OUTPUT:
LAST SEEN   TYPE     REASON              MESSAGE
45s         Normal   WindowTransition    Entered window: BusinessHours
45s         Normal   ScalingTarget       Scaling webapp-demo from 0 to 3
44s         Normal   ScaledUp            Scaled up to 3

TEXT OVERLAY:
Events confirm scaling decisions
```

---

**Shot 10: Next Steps (2:45-3:00) - 15 seconds**
```
TEXT OVERLAY:
What's Next?

✓ Customize windows for your schedule
✓ Add grace periods for smooth scale-downs
✓ Configure holidays with ConfigMaps
✓ Monitor with Prometheus metrics

Docs: kyklos.io/docs
Examples: github.com/your-org/kyklos/examples
```

---

### Post-Production Notes (Video 3)

**Editing Checklist:**
- [ ] Speed up typing (2x) for commands
- [ ] Keep YAML display readable (don't speed up)
- [ ] Add "fast forward" indicator during watch
- [ ] Include prerequisite icons (checkmarks)
- [ ] Add "Next Steps" as separate title cards
- [ ] Ensure URLs are clearly visible and correct
- [ ] Add QR code overlay for docs/GitHub (optional)

---

## Recording Environment Setup

### Terminal Configuration Script

```bash
# Save as setup-recording-terminal.sh

#!/bin/bash

# Terminal colors (dark theme)
export CLICOLOR=1
export LSCOLORS=GxFxCxDxBxegedabagaced

# PS1 prompt (simple)
export PS1='$ '

# Clear scrollback
clear

# Set terminal title
echo -ne "\033]0;Kyklos Demo\007"

# Verify terminal size
tput cols  # Should be 120
tput lines # Should be 35-40

echo "Terminal configured for recording"
echo "Font: 16pt Monaco"
echo "Size: $(tput cols)x$(tput lines)"
echo ""
echo "Ready to record!"
```

---

### Pre-Recording Checklist

**Environment:**
- [ ] Notifications disabled (Do Not Disturb)
- [ ] Desktop clean (hide icons, close apps)
- [ ] Terminal configured (font, size, colors)
- [ ] Kubectl context correct
- [ ] Demo namespace clean (deleted if exists)
- [ ] Controller running and healthy
- [ ] Window times calculated (for real-time demos)

**Recording Software:**
- [ ] Screen recording app open and tested
- [ ] Recording settings verified (1080p, 30fps)
- [ ] Audio disabled (silent recording)
- [ ] Cursor visibility enabled
- [ ] Test recording saved successfully
- [ ] Adequate disk space (10GB+ recommended)

**Demo Materials:**
- [ ] Manifests prepared with correct times
- [ ] Commands listed in order (script or notes)
- [ ] Shotlist printed or on second screen
- [ ] Timing notes handy (when to pause, when to emphasize)

---

## Post-Production Workflow

### Step 1: Raw Recording Review
- [ ] Watch full recording
- [ ] Note any mistakes or retakes needed
- [ ] Identify sections to speed up (typing, waiting)
- [ ] Mark capture points for screenshots

### Step 2: Editing
- [ ] Import to video editor (iMovie, Final Cut, Premiere, DaVinci Resolve)
- [ ] Cut out dead time (before commands, long waits)
- [ ] Speed up typing sections (1.5-2x)
- [ ] Keep key moments at real-time (scale transitions)
- [ ] Add fade-in at start, fade-out at end

### Step 3: Text Overlays
- [ ] Add title card at beginning
- [ ] Add text overlays at specified timestamps
- [ ] Use consistent font and styling
- [ ] Ensure text is readable at 1080p (test on smaller screen)
- [ ] Add closing card with links

### Step 4: Visual Enhancements
- [ ] Add subtle zoom-ins on key moments
- [ ] Highlight important lines (YAML, events, logs) with boxes/arrows
- [ ] Add timeline diagrams where helpful (cross-midnight)
- [ ] Add "fast forward" indicators during sped-up sections
- [ ] Ensure cursor is visible throughout

### Step 5: Export and Test
- [ ] Export as MP4, H.264, 1080p, 30fps
- [ ] Test playback on multiple devices (desktop, mobile)
- [ ] Verify text is readable on small screens
- [ ] Check file size (should be < 50MB for 3-minute video)
- [ ] Upload to YouTube as unlisted for review

### Step 6: Final Review
- [ ] Watch full video on different screens
- [ ] Verify all text overlays appear at correct times
- [ ] Check for any audio (should be silent)
- [ ] Ensure video length matches target duration
- [ ] Get feedback from team member

---

## Video Hosting and Distribution

### YouTube Upload
- Title: "Kyklos Time Window Scaler - [Video Name]"
- Description: Include GitHub link, docs link, key features
- Tags: kubernetes, autoscaling, time-based, cron, scheduling
- Thumbnail: Screenshot of key moment with text overlay
- Playlist: Create "Kyklos Demos" playlist

### Embed in Documentation
```markdown
## Video Demo

<video width="800" controls>
  <source src="kyklos-2-minute-demo.mp4" type="video/mp4">
  Your browser does not support the video tag.
</video>

Or watch on [YouTube](https://youtube.com/watch?v=...)
```

### Social Media Versions
- Twitter/X: 2:20 max - Use Video 1 trimmed
- LinkedIn: Full 3-minute Video 1 or Video 3
- Reddit: Upload to Reddit video (2-minute limit)

---

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2025-10-28 | 1.0 | Initial video shotlist for three demo videos |
