# Demo Capture Checklist

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** demo-scenario-designer

## Purpose

This checklist ensures comprehensive capture of demo materials for documentation, README images, blog posts, and video production. Use this during demo execution to ensure no critical moments are missed.

---

## Pre-Demo Preparation

### Equipment Setup

**Terminal Configuration:**
- [ ] Terminal font: Monaco 14pt or larger (for screenshots)
- [ ] Terminal size: 120 columns x 40 rows minimum
- [ ] Color scheme: High contrast (light background or dark with good contrast)
- [ ] Prompt: Short and clean (e.g., `$ ` without hostname)
- [ ] Clear scrollback buffer before starting

**Screen Recording:**
- [ ] Screen recording software ready (QuickTime, OBS, or asciinema)
- [ ] Audio disabled (we want silent recording)
- [ ] Resolution: 1920x1080 minimum
- [ ] Frame rate: 30fps minimum
- [ ] Cursor highlighting enabled (if available)

**Screenshot Tool:**
- [ ] Native screenshot tool ready (macOS: Cmd+Shift+4, Linux: scrot/flameshot)
- [ ] Screenshot directory prepared: `/tmp/kyklos-demo-captures/`
- [ ] Naming convention: `YYYY-MM-DD-HHmmss-description.png`

**Environment:**
- [ ] Notifications disabled (macOS: Do Not Disturb, Linux: disable notifications)
- [ ] Desktop clean (hide desktop icons, close unrelated applications)
- [ ] Browser tabs prepared (for opening docs if needed in recording)
- [ ] Second terminal window ready for controller logs

---

## Scenario A: Minute Demo Captures

### Critical Screenshots (10 required)

#### Capture Point 1: Initial State
**When:** T+0:10 (Step 1.2)
**Command:** `kubectl get deploy,pods -n demo`
**What to show:**
- [ ] Deployment with 0/0 replicas
- [ ] "No pods found in demo namespace" message
- [ ] Clean namespace state

**File naming:** `01-initial-state-zero-replicas.png`

**Framing:**
- Include full command and output
- Show prompt with date command for timestamp
- Crop to relevant terminal area only

---

#### Capture Point 2: TWS OffHours State
**When:** T+0:20 (Step 2.2)
**Command:** `kubectl get tws -n demo`
**What to show:**
- [ ] TWS created with AGE ~5s
- [ ] WINDOW shows "OffHours"
- [ ] REPLICAS shows 0
- [ ] TARGET shows webapp-demo

**File naming:** `02-tws-offhours-before-window.png`

**Framing:**
- Include kubectl command
- Show clean output with header row
- Highlight WINDOW and REPLICAS columns

---

#### Capture Point 3: BusinessHours Active
**When:** T+1:15 (Step 4.1)
**Command:** `kubectl get tws,deploy,pods -n demo`
**What to show:**
- [ ] TWS WINDOW changed to "BusinessHours"
- [ ] TWS REPLICAS shows 2
- [ ] Deployment 2/2 Ready
- [ ] Two pods with STATUS: Running

**File naming:** `03-businesshours-active-two-replicas.png`

**Framing:**
- Show all three resource types in one shot
- Verify all pods show "Running" not "ContainerCreating"
- Clean, stable state (wait for full readiness)

**This is the PRIMARY HERO SHOT for README.**

---

#### Capture Point 4: Scale-Up Events
**When:** T+1:20 (Step 4.2)
**Command:** `kubectl get events -n demo --sort-by='.lastTimestamp' | tail -10`
**What to show:**
- [ ] WindowTransition event: "Entered window: BusinessHours"
- [ ] ScalingTarget event: "Scaling webapp-demo from 0 to 2"
- [ ] ScaledUp event from deployment
- [ ] SuccessfulCreate events for pods

**File naming:** `04-scaleup-events-timeline.png`

**Framing:**
- Show events in chronological order (oldest to newest)
- Include LAST SEEN, REASON, and MESSAGE columns
- Highlight Kyklos-specific events (WindowTransition, ScalingTarget)

---

#### Capture Point 5: Controller Scale-Up Decision Logs
**When:** T+1:25 (Step 4.3)
**Command:** `kubectl logs -n kyklos-system -l app=kyklos-controller --tail=20 | grep webapp`
**What to show:**
- [ ] "Current time in UTC" log line
- [ ] "Matched window: BusinessHours" log line
- [ ] "Scaling deployment from 0 to 2" log line
- [ ] "Requeue scheduled at next boundary" log line

**File naming:** `05-controller-logs-scaleup-decision.png`

**Framing:**
- Show 10-15 lines of context
- Include timestamps if present in logs
- Highlight decision rationale lines

---

#### Capture Point 6: TWS Detailed Status
**When:** T+2:00 (Step 5.2)
**Command:** `kubectl get tws webapp-minute-scaler -n demo -o yaml | grep -A 20 status:`
**What to show:**
- [ ] status.currentWindow: BusinessHours
- [ ] status.effectiveReplicas: 2
- [ ] status.targetObservedReplicas: 2
- [ ] All three conditions (Ready, Reconciling, Degraded)
- [ ] Ready condition shows status: "True"

**File naming:** `06-tws-status-all-conditions.png`

**Framing:**
- Show full status block
- Ensure YAML formatting is preserved
- Highlight condition types and statuses

---

#### Capture Point 7: OffHours After Scale-Down
**When:** T+4:15 (Step 6.1)
**Command:** `kubectl get tws,deploy,pods -n demo`
**What to show:**
- [ ] TWS WINDOW shows "OffHours"
- [ ] TWS REPLICAS shows 0
- [ ] Deployment 0/0
- [ ] "No pods in demo namespace" message

**File naming:** `07-offhours-zero-replicas-after-window.png`

**Framing:**
- Mirror framing of Capture Point 3 (for before/after comparison)
- Clean final state
- Show return to initial condition

---

#### Capture Point 8: Scale-Down Events
**When:** T+4:20 (Step 6.2)
**Command:** `kubectl get events -n demo --sort-by='.lastTimestamp' | tail -10`
**What to show:**
- [ ] WindowTransition event: "Exited window: BusinessHours"
- [ ] ScalingTarget event: "Scaling webapp-demo from 2 to 0"
- [ ] ScaledDown event
- [ ] Killing events for pods

**File naming:** `08-scaledown-events-timeline.png`

**Framing:**
- Show events in chronological order
- Parallel to Capture Point 4 (for comparison)
- Highlight window exit event

---

#### Capture Point 9: Controller Scale-Down Decision Logs
**When:** T+4:25 (Step 6.3)
**Command:** `kubectl logs -n kyklos-system -l app=kyklos-controller --tail=20 | grep webapp`
**What to show:**
- [ ] "Current time in UTC" log line
- [ ] "No matching windows, using defaultReplicas" log line
- [ ] "Scaling deployment from 2 to 0" log line
- [ ] Requeue scheduled for next day

**File naming:** `09-controller-logs-scaledown-decision.png`

**Framing:**
- Parallel to Capture Point 5
- Highlight default replicas logic
- Show requeue time (next occurrence)

---

#### Capture Point 10: Complete Event Timeline
**When:** T+5:00 (Step 7.2)
**Command:** `kubectl get events -n demo --sort-by='.lastTimestamp'`
**What to show:**
- [ ] Full chronological event list
- [ ] Scale-up events (oldest)
- [ ] Scale-down events (newest)
- [ ] Complete lifecycle in one view

**File naming:** `10-complete-event-timeline.png`

**Framing:**
- Show 20-30 events (full demo lifecycle)
- Ensure oldest events visible at top
- Newest events at bottom
- May need to scroll or use `| head -30`

---

### Additional Captures (Optional but Recommended)

#### Supplementary Screenshot 1: kubectl get tws Wide Output
**Command:** `kubectl get tws -n demo -o wide`
**File naming:** `S1-tws-wide-output.png`
**Use case:** Shows additional columns not in default view

#### Supplementary Screenshot 2: Deployment Describe
**Command:** `kubectl describe deploy webapp-demo -n demo | grep -A 10 Replicas`
**File naming:** `S2-deployment-describe-replicas.png`
**Use case:** Shows deployment-level replica details

#### Supplementary Screenshot 3: Watch Output Animation
**Command:** Multiple frames from watch command
**File naming:** `S3-watch-transition-[00-10].png`
**Use case:** Create animated GIF of scale transition

---

## Scenario B: Cross-Midnight Demo Captures

### Critical Screenshots (11 required)

#### Capture Point 1: Berlin Time Initial State
**When:** T+0:10 (Step 1.2)
**Command:** `TZ=Europe/Berlin date && kubectl get deploy,pods -n demo`
**What to show:**
- [ ] Berlin timezone timestamp (CET or CEST)
- [ ] Tuesday evening time (~22:47)
- [ ] Deployment 0/0 replicas
- [ ] No pods

**File naming:** `B01-berlin-time-initial-state.png`

**Framing:**
- Date output must show timezone (CET/CEST)
- Show day of week (Tuesday)
- Full deploy and pods output

---

#### Capture Point 2: Cross-Midnight Window Spec
**When:** T+0:30 (Step 2.3)
**Command:** `kubectl get tws nightshift-scaler -n demo -o yaml | grep -A 10 windows:`
**What to show:**
- [ ] start: "22:48" (or your calculated time)
- [ ] end: "00:50" (or your calculated time)
- [ ] **end value is LESS than start value** (key indicator)
- [ ] days: all seven days
- [ ] replicas: 3

**File naming:** `B02-cross-midnight-window-spec.png`

**Framing:**
- Highlight start and end times
- Show that end < start (cross-midnight indicator)
- Include full window definition

---

#### Capture Point 3: NightShift Active on Tuesday
**When:** T+1:20 (Step 4.1)
**Command:** `TZ=Europe/Berlin date && kubectl get tws,deploy,pods -n demo`
**What to show:**
- [ ] Tuesday date (~22:48)
- [ ] TWS WINDOW: "NightShift"
- [ ] TWS REPLICAS: 3
- [ ] Deployment 3/3
- [ ] Three pods Running

**File naming:** `B03-nightshift-active-tuesday.png`

**Framing:**
- Date must clearly show Tuesday
- All resources in stable state
- Clean, complete output

---

#### Capture Point 4: Controller Cross-Midnight Detection
**When:** T+1:25 (Step 4.2)
**Command:** `kubectl logs -n kyklos-system -l app=kyklos-controller --tail=30 | grep -A 5 nightshift`
**What to show:**
- [ ] "Evaluating cross-midnight window" log line
- [ ] "window extends into next day" log line
- [ ] Window notation showing day transition (Tue → Wed)
- [ ] Requeue scheduled for 00:50 next day

**File naming:** `B04-controller-detects-cross-midnight.png`

**Framing:**
- Show explicit cross-midnight detection logic
- Highlight day transition mention
- Include requeue time calculation

---

#### Capture Point 5: Status Metadata Before Midnight
**When:** T+15:00 (Step 5.2, ~23:00)
**Command:** `kubectl get tws nightshift-scaler -n demo -o yaml | grep -A 15 status:`
**What to show:**
- [ ] currentWindow: NightShift
- [ ] windowMetadata.crossesMidnight: true
- [ ] windowEndDay: "Wednesday"
- [ ] Current date still Tuesday

**File naming:** `B05-status-metadata-pre-midnight.png`

**Framing:**
- Show crossesMidnight flag
- Show window end day (next day)
- Emphasize metadata tracking cross-day state

---

#### Capture Point 6: Wednesday Date, Window Still Active (HERO SHOT)
**When:** T+73:05 (Step 6.1, ~00:00)
**Command:** `TZ=Europe/Berlin date && kubectl get tws,deploy,pods -n demo`
**What to show:**
- [ ] **Wednesday date** (~00:00)
- [ ] **TWS WINDOW still "NightShift"** (not OffHours!)
- [ ] **TWS REPLICAS still 3** (no scale change)
- [ ] **Three pods still Running**
- [ ] Clean, stable state

**File naming:** `B06-wednesday-window-still-active-HERO.png`

**Framing:**
- This is THE critical screenshot for cross-midnight proof
- Date must be very prominent
- Show complete stability across midnight
- Consider terminal title showing "WEDNESDAY 00:00 - WINDOW STILL ACTIVE"

**THIS IS THE PRIMARY HERO SHOT FOR CROSS-MIDNIGHT DEMONSTRATION.**

---

#### Capture Point 7: Controller Logs Confirming Midnight Crossed
**When:** T+73:10 (Step 6.2)
**Command:** `kubectl logs -n kyklos-system -l app=kyklos-controller --since=5m | grep -A 3 "00:00"`
**What to show:**
- [ ] Reconciliation around 00:00 timestamp
- [ ] "Cross-midnight window still active" or equivalent
- [ ] Remaining time calculation
- [ ] No scaling action taken

**File naming:** `B07-controller-confirms-midnight-crossed.png`

**Framing:**
- Show 00:00 timestamp
- Highlight "still active" decision
- Show no scale action

---

#### Capture Point 8: Cross-Day Window Metadata
**When:** T+73:15 (Step 6.3)
**Command:** `kubectl get tws nightshift-scaler -n demo -o jsonpath='{.status.windowMetadata}' | jq`
**What to show:**
- [ ] crossesMidnight: true
- [ ] currentDay: "Wednesday"
- [ ] windowStartDay: "Tuesday"
- [ ] windowEndDay: "Wednesday"
- [ ] activeFor and remainingTime fields

**File naming:** `B08-cross-day-metadata-post-midnight.png`

**Framing:**
- Show JSON formatted output
- Highlight day fields showing cross-day state
- Include time calculations

---

#### Capture Point 9: OffHours on Wednesday After Window End
**When:** T+123:15 (Step 7.1, ~00:50)
**Command:** `TZ=Europe/Berlin date && kubectl get tws,deploy,pods -n demo`
**What to show:**
- [ ] Wednesday date (~00:50)
- [ ] TWS WINDOW: "OffHours"
- [ ] TWS REPLICAS: 0
- [ ] Deployment 0/0
- [ ] No pods

**File naming:** `B09-wednesday-offhours-after-window-end.png`

**Framing:**
- Show Wednesday date (window end day)
- Return to OffHours state
- Clean final state

---

#### Capture Point 10: Controller Window End Calculation
**When:** T+123:20 (Step 7.2)
**Command:** `kubectl logs -n kyklos-system -l app=kyklos-controller --tail=20 | grep nightshift`
**What to show:**
- [ ] "Cross-midnight window ended" log line
- [ ] Total duration calculation (e.g., "2h 2m")
- [ ] Scale-down decision
- [ ] Next window scheduled for tonight (Wed 22:48)

**File naming:** `B10-controller-calculates-window-end.png`

**Framing:**
- Show explicit window end detection
- Highlight duration calculation
- Show next occurrence scheduling

---

#### Capture Point 11: Complete Event Timeline with MidnightCrossed
**When:** T+125:00 (Step 8.1)
**Command:** `kubectl get events -n demo --sort-by='.lastTimestamp'`
**What to show:**
- [ ] Entered window event (Tuesday timestamp)
- [ ] **MidnightCrossed event** (if implemented)
- [ ] Exited window event (Wednesday timestamp)
- [ ] Complete lifecycle spanning midnight

**File naming:** `B11-complete-event-timeline-cross-midnight.png`

**Framing:**
- Show full timeline (20-30 events)
- Highlight MidnightCrossed event if present
- Show timestamp day changes (Tue → Wed)

---

### Additional Captures (Scenario B Specific)

#### Supplementary Screenshot B1: Timezone Conversion
**Command:** `TZ=UTC date && TZ=Europe/Berlin date`
**File naming:** `SB1-timezone-comparison-utc-berlin.png`
**Use case:** Show UTC vs Berlin time difference

#### Supplementary Screenshot B2: Window Duration Calculation
**Command:** Terminal showing manual calculation of cross-midnight duration
**File naming:** `SB2-duration-calculation.png`
**Use case:** Educational diagram for docs

---

## Terminal Recording Guidelines

### Full Demo Recording

**Scenario A: Minute Demo**
- **Duration:** 10 minutes total (T+0 to T+10)
- **What to record:**
  - Start: Just before creating namespace
  - End: After cleanup verification
  - Include: All watch output, showing real-time transitions
  - Focus: Primary terminal with watch running

**Scenario B: Cross-Midnight**
- **Duration:** 2+ hours (not practical for continuous recording)
- **Alternative:** Record three segments:
  1. Setup and window opening (T+0 to T+2) - 2 minutes
  2. Pre-midnight to post-midnight (T+72 to T+74) - 2 minutes
  3. Window closing (T+122 to T+124) - 2 minutes
  - **Total recorded:** 6 minutes (edited together)

---

### Short Clip Recordings (For Video)

**Clip 1: Scale-Up Animation (30 seconds)**
```bash
# Start recording
kubectl get tws,deploy,pods -n demo
# Wait 5 seconds
watch -n 1 'kubectl get tws,deploy,pods -n demo'
# Record through scale-up transition
# Stop after pods reach Running
```

**Clip 2: Events Stream (15 seconds)**
```bash
# Start recording
kubectl get events -n demo --watch
# Record 3-4 events appearing
# Stop recording
```

**Clip 3: Controller Logs Live (20 seconds)**
```bash
# Start recording
kubectl logs -n kyklos-system -l app=kyklos-controller --follow
# Record through one reconciliation cycle
# Stop recording
```

---

## asciinema Recordings (CLI-friendly)

### Why asciinema?

- Produces copyable terminal recordings (users can select text)
- Small file size
- Can be played in browser
- Can export to animated GIF

### Commands to Record

**Full Minute Demo:**
```bash
asciinema rec /tmp/kyklos-minute-demo.cast
# Run full demo
# Press Ctrl+D to stop
```

**Editing:**
```bash
# Trim recording
asciinema play /tmp/kyklos-minute-demo.cast
# Note times to keep
asciinema play -s 2 /tmp/kyklos-minute-demo.cast  # 2x speed preview

# Upload (optional)
asciinema upload /tmp/kyklos-minute-demo.cast
```

---

## Screenshot Post-Processing

### Consistency Checklist

For all screenshots:
- [ ] Terminal size consistent (120x40 minimum)
- [ ] Font size consistent (14pt+)
- [ ] Color scheme consistent
- [ ] Crop to show only relevant terminal area (no desktop chrome)
- [ ] No personal information visible (usernames, hostnames)
- [ ] Timestamps preserved (where relevant)

### Annotation Guidelines

**Do NOT annotate directly on screenshots.** Keep originals clean.

**For presentation decks:**
- Create annotated versions separately
- Use consistent arrow/box style
- Annotations should be removable layer

### File Organization

```
/tmp/kyklos-demo-captures/
├── scenario-a/
│   ├── 01-initial-state-zero-replicas.png
│   ├── 02-tws-offhours-before-window.png
│   ├── ...
│   └── 10-complete-event-timeline.png
├── scenario-b/
│   ├── B01-berlin-time-initial-state.png
│   ├── B02-cross-midnight-window-spec.png
│   ├── ...
│   └── B11-complete-event-timeline-cross-midnight.png
├── supplementary/
│   ├── S1-tws-wide-output.png
│   ├── SB1-timezone-comparison-utc-berlin.png
│   └── ...
├── recordings/
│   ├── scenario-a-full-demo.mov
│   ├── scenario-b-clip1-setup.mov
│   ├── scenario-b-clip2-midnight.mov
│   └── scenario-b-clip3-end.mov
└── asciinema/
    ├── minute-demo.cast
    └── cross-midnight-segments.cast
```

---

## kubectl Output Captures

### Key Outputs to Save as Text

Save these to files for easy reference in documentation:

**Scenario A:**
```bash
# TWS spec
kubectl get tws webapp-minute-scaler -n demo -o yaml > /tmp/captures/tws-spec-minute-demo.yaml

# TWS status
kubectl get tws webapp-minute-scaler -n demo -o jsonpath='{.status}' | jq > /tmp/captures/tws-status-minute-demo.json

# Events
kubectl get events -n demo --sort-by='.lastTimestamp' > /tmp/captures/events-minute-demo.txt

# Controller logs
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=100 > /tmp/captures/controller-logs-minute-demo.txt
```

**Scenario B:**
```bash
# Same pattern for nightshift-scaler
kubectl get tws nightshift-scaler -n demo -o yaml > /tmp/captures/tws-spec-cross-midnight.yaml
kubectl get tws nightshift-scaler -n demo -o jsonpath='{.status}' | jq > /tmp/captures/tws-status-cross-midnight.json
kubectl get events -n demo --sort-by='.lastTimestamp' > /tmp/captures/events-cross-midnight.txt
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=200 > /tmp/captures/controller-logs-cross-midnight.txt
```

---

## Quality Assurance Checklist

### Before Finalizing Captures

**Review each screenshot:**
- [ ] Is the command visible and correct?
- [ ] Is the output complete (not truncated)?
- [ ] Are timestamps/dates clearly visible?
- [ ] Is text readable at 50% zoom?
- [ ] Are all critical fields highlighted or visible?
- [ ] Does filename accurately describe content?

**Review recordings:**
- [ ] Audio is disabled (silent recording)
- [ ] Cursor is visible
- [ ] No lag or frame drops during transitions
- [ ] Resolution is high enough for readability
- [ ] Length is appropriate (not too long)

**Review text outputs:**
- [ ] Complete (not truncated by terminal buffer)
- [ ] Properly formatted (YAML/JSON valid)
- [ ] Sensitive data redacted if present
- [ ] File encoding is UTF-8

---

## Handoff Package Contents

When passing captures to Docs Writer, include:

### README Bundle
- [ ] Scenario A Capture Points 1, 3, 4 (essential trio)
- [ ] Scenario B Capture Point 6 (hero shot)
- [ ] Both full event timelines (10, B11)
- [ ] One-page summary of what each shows

### Complete Documentation Bundle
- [ ] All 21 screenshots (10 from A, 11 from B)
- [ ] All supplementary screenshots
- [ ] All text output files
- [ ] Terminal recordings (full and clips)
- [ ] asciinema casts
- [ ] File manifest listing all materials

### Video Production Bundle
- [ ] All short clips (30s or less)
- [ ] asciinema recordings
- [ ] Text outputs for overlay captions
- [ ] Timing notes (when each event occurred)
- [ ] Shotlist cross-reference

---

## Capture Point Quick Reference

### Scenario A: 10 Critical Screenshots
1. Initial state (0/0)
2. TWS OffHours
3. BusinessHours active (HERO)
4. Scale-up events
5. Controller scale-up logs
6. TWS detailed status
7. OffHours after scale-down
8. Scale-down events
9. Controller scale-down logs
10. Complete event timeline

### Scenario B: 11 Critical Screenshots
1. Berlin time initial
2. Cross-midnight spec (end < start)
3. NightShift active Tuesday
4. Controller detects cross-midnight
5. Status metadata pre-midnight
6. Wednesday window still active (HERO)
7. Controller confirms midnight crossed
8. Cross-day metadata post-midnight
9. OffHours Wednesday after end
10. Controller window end calculation
11. Complete event timeline with midnight crossing

---

## Troubleshooting Captures

### Issue: Screenshot text too small

**Fix:**
- Increase terminal font size to 16pt or 18pt
- Reduce terminal columns to 100 (forces line wrapping at reasonable width)
- Use high-DPI display or increase scaling

### Issue: Watch output changing too fast to capture

**Fix:**
- Increase watch interval: `watch -n 5` instead of `-n 2`
- Pause watch with Ctrl+Z, capture, resume with `fg`
- Use `kubectl get ... -o yaml` for static output instead

### Issue: Events showing wrong order

**Fix:**
- Always use `--sort-by='.lastTimestamp'` for chronological order
- If timestamps are same, events may appear in arbitrary order (expected)
- Capture multiple times to get clean ordering

### Issue: Controller logs missing key lines

**Fix:**
- Increase tail lines: `--tail=50` or `--tail=100`
- Use `--since=5m` to get last 5 minutes
- If reconciliation happened earlier, use `--since=10m`
- Combine: `kubectl logs --since=10m --tail=200`

---

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2025-10-28 | 1.0 | Initial capture checklist for both demo scenarios |
