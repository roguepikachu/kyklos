# Kyklos Demo Scenarios

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** demo-scenario-designer

## Overview

This directory contains comprehensive demo scenarios, capture plans, and recording materials for showcasing Kyklos Time Window Scaler. All demos are designed to complete within 10 minutes and provide clear, reproducible demonstrations of Kyklos capabilities.

---

## Demo Scenarios

### Scenario A: Minute-Scale Demo (UTC)
**File:** [SCENARIO-A-MINUTE-DEMO.md](/Users/aykumar/personal/kyklos/docs/demos/SCENARIO-A-MINUTE-DEMO.md)

**Duration:** 10 minutes
**Complexity:** Beginner-friendly
**Key Demonstration:** Basic time-based scaling from 0→2→0 replicas

**Use this when:**
- Introducing Kyklos to new users
- Creating README hero shots
- Recording quick overview videos
- Live demonstrations with time constraints
- Testing basic controller functionality

**Highlights:**
- Uses UTC timezone (no DST complications)
- 3-minute active window (T+1min to T+4min)
- Clear scale-up and scale-down transitions
- Complete event timeline
- Controller decision logs
- 10 defined capture points

---

### Scenario B: Cross-Midnight Demo (Europe/Berlin)
**File:** [SCENARIO-B-CROSS-MIDNIGHT.md](/Users/aykumar/personal/kyklos/docs/demos/SCENARIO-B-CROSS-MIDNIGHT.md)

**Duration:** 2+ hours (with 30-minute abbreviated option)
**Complexity:** Advanced
**Key Demonstration:** Windows that span calendar day boundaries

**Use this when:**
- Demonstrating advanced Kyklos features
- Showing DST-aware timezone handling
- Explaining cross-midnight boundary logic
- Creating technical deep-dive content
- Validating controller correctness

**Highlights:**
- Uses Europe/Berlin timezone (DST-aware)
- Window crosses midnight (22:00 to 02:00)
- Demonstrates calendar day transition
- Proves window remains active past midnight
- Shows cross-day metadata tracking
- 11 defined capture points including midnight transition

**Critical Moment:** Window remains active when date changes from Friday to Saturday at 00:00, proving cross-midnight handling works.

---

## Supporting Documentation

### Capture Checklist
**File:** [CAPTURE-CHECKLIST.md](/Users/aykumar/personal/kyklos/docs/demos/CAPTURE-CHECKLIST.md)

**Purpose:** Comprehensive checklist for capturing screenshots, recordings, and outputs during demo execution.

**Contents:**
- Pre-demo equipment setup
- 10 critical screenshot capture points (Scenario A)
- 11 critical screenshot capture points (Scenario B)
- Additional supplementary captures
- Terminal recording guidelines
- asciinema recording instructions
- Screenshot post-processing guidelines
- File organization structure
- Quality assurance checklist

**Use this to:**
- Ensure no critical moments are missed
- Maintain consistency across captures
- Organize demo materials for docs team
- Prepare for video recordings

---

### Demo Dry Run Checklist
**File:** [DEMO-DRY-RUN.md](/Users/aykumar/personal/kyklos/docs/demos/DEMO-DRY-RUN.md)

**Purpose:** Pre-demo rehearsal checklist to verify environment, test timing, and identify issues before live demos.

**Contents:**
- Pre-rehearsal environment verification
- Scenario A dry run procedure (10 minutes)
- Scenario B dry run procedure (evening only or simulated)
- Common issues and fixes
- Recovery procedures
- Pass/fail criteria
- Rehearsal notes template

**Use this to:**
- Validate demo environment before live execution
- Test timing calculations
- Identify and fix issues in advance
- Build confidence for live demos
- Document rehearsal results

---

### Video Shotlist and Recording Plan
**File:** [VIDEO-SHOTLIST.md](/Users/aykumar/personal/kyklos/docs/demos/VIDEO-SHOTLIST.md)

**Purpose:** Shot-by-shot plans for recording promotional and educational videos.

**Contents:**
- Video 1: "Kyklos in 2 Minutes" (2:00) - Overview
- Video 2: "Cross-Midnight Windows" (2:30) - Advanced feature
- Video 3: "Quick Start" (3:00) - Tutorial
- Shot-by-shot breakdown with timestamps
- Terminal commands and expected outputs
- Text overlay specifications
- Post-production editing notes
- Recording environment setup
- Export and distribution guidelines

**Use this to:**
- Record professional demo videos
- Create consistent video content
- Plan text overlays and captions
- Guide video editing process

---

### Annotations and Captions
**File:** [ANNOTATIONS.md](/Users/aykumar/personal/kyklos/docs/demos/ANNOTATIONS.md)

**Purpose:** Consistent text annotations, captions, and terminology for all demo materials.

**Contents:**
- Standard terminology reference
- Screenshot annotation styles and examples
- Video caption library
- Presentation slide annotations
- Blog post / README captions
- Social media post templates
- Error message annotations
- Accessibility alt text

**Use this to:**
- Ensure consistent terminology across materials
- Create professional-looking annotations
- Write clear captions and overlays
- Maintain accessibility compliance
- Craft social media posts

---

## Quick Start Guide

### Running Your First Demo (Scenario A)

**Time Required:** 15 minutes (including setup)

**Steps:**

1. **Verify Prerequisites:**
   ```bash
   make verify-all
   ```

2. **Calculate Window Times:**
   ```bash
   # Get current UTC time
   date -u

   # Calculate T+1min (window start)
   # Calculate T+4min (window end)
   # Write these down!
   ```

3. **Follow Scenario A:**
   Open [SCENARIO-A-MINUTE-DEMO.md](/Users/aykumar/personal/kyklos/docs/demos/SCENARIO-A-MINUTE-DEMO.md) and execute step-by-step.

4. **Capture Screenshots:**
   Follow [CAPTURE-CHECKLIST.md](/Users/aykumar/personal/kyklos/docs/demos/CAPTURE-CHECKLIST.md) capture points 1-10.

5. **Success Criteria:**
   - Window opened at T+1min ± 10 seconds
   - Scaled from 0 to 2 replicas
   - Window closed at T+4min ± 10 seconds
   - Scaled from 2 to 0 replicas
   - No controller errors

---

### Recording Your First Video

**Time Required:** 45 minutes (recording + editing)

**Steps:**

1. **Dry Run First:**
   Complete [DEMO-DRY-RUN.md](/Users/aykumar/personal/kyklos/docs/demos/DEMO-DRY-RUN.md) Scenario A dry run to verify timing and identify issues.

2. **Setup Recording:**
   - Configure terminal (16pt font, 120x40)
   - Disable notifications
   - Start screen recording software
   - Test with 10-second trial recording

3. **Follow Video 1 Shotlist:**
   Open [VIDEO-SHOTLIST.md](/Users/aykumar/personal/kyklos/docs/demos/VIDEO-SHOTLIST.md) and follow Video 1: "Kyklos in 2 Minutes" shot-by-shot.

4. **Post-Production:**
   - Import to video editor
   - Add text overlays from ANNOTATIONS.md
   - Speed up typing (1.5-2x)
   - Keep scale transitions at real-time
   - Export as 1080p MP4

---

## Demo Execution Checklist

Use this quick checklist before any demo:

**Pre-Demo (5 minutes before):**
- [ ] Cluster running and healthy: `kubectl cluster-info`
- [ ] Controller running: `kubectl get pods -n kyklos-system`
- [ ] CRDs installed: `kubectl get crd timewindowscalers.kyklos.io`
- [ ] Demo namespace clean: `kubectl delete namespace demo --ignore-not-found`
- [ ] Window times calculated and written down
- [ ] Manifests prepared with correct times
- [ ] Terminal configured (font, size, colors)
- [ ] Notifications disabled

**During Demo:**
- [ ] Execute commands calmly (don't rush)
- [ ] Wait for output to complete before next command
- [ ] Narrate what you're doing (for live demos)
- [ ] Capture screenshots at designated points
- [ ] Monitor controller logs in second terminal

**Post-Demo:**
- [ ] Verify all capture points obtained
- [ ] Clean up namespace: `kubectl delete namespace demo`
- [ ] Review recording if applicable
- [ ] Note any issues for next time

---

## File Manifest

```
docs/demos/
├── README.md                      # This file - Overview and guide
├── SCENARIO-A-MINUTE-DEMO.md      # UTC minute-scale demo (10 min)
├── SCENARIO-B-CROSS-MIDNIGHT.md   # Europe/Berlin cross-midnight demo (2+ hours)
├── CAPTURE-CHECKLIST.md           # Screenshot and recording capture guide
├── DEMO-DRY-RUN.md                # Rehearsal checklist and procedures
├── VIDEO-SHOTLIST.md              # Video recording shot-by-shot plans
└── ANNOTATIONS.md                 # Consistent captions and terminology
```

---

## Best Practices

### Demo Execution

1. **Always Dry Run First:** Run through the demo at least once before any live presentation or recording. Use DEMO-DRY-RUN.md.

2. **Calculate Times in Advance:** Don't calculate window times during the demo. Do it beforehand and write them in your notes.

3. **Have a Second Terminal:** Keep controller logs visible in a second terminal window for reference and troubleshooting.

4. **Monitor Watch Output:** Don't stare at terminal during watch. Look at the output to explain what's happening.

5. **Pause at Key Moments:** After window opens or closes, pause for 3-5 seconds to let viewers absorb the change.

### Recording

1. **Silent Videos Are Better:** Don't record audio. Add text overlays in post-production. This allows for easier editing and internationalization.

2. **Keep It Short:** Target 2-3 minutes per video. Viewers have short attention spans.

3. **Show Real Timing:** Don't speed up scale transitions. Show them at natural speed to demonstrate real-world behavior.

4. **Use High Contrast:** Dark terminal with bright text, or light terminal with dark text. Avoid low-contrast color schemes.

### Capture

1. **Capture More Than Needed:** Take extra screenshots. You can always discard extras, but you can't go back to capture moments you missed.

2. **Use Consistent File Naming:** Follow the naming convention in CAPTURE-CHECKLIST.md for easy organization.

3. **Keep Originals Clean:** Don't annotate original screenshots. Create annotated versions separately for presentation.

4. **Verify Captures Immediately:** Review screenshots right after capture to ensure text is readable and framing is correct.

---

## Troubleshooting Common Issues

### Window Didn't Open at Calculated Time

**Symptom:** Passed T+1min but WINDOW still shows OffHours.

**Quick Fix:**
1. Check current UTC time: `date -u`
2. Verify TWS spec has correct times: `kubectl get tws -n demo -o yaml | grep -A 5 windows`
3. Check controller logs: `kubectl logs -n kyklos-system -l app=kyklos-controller --tail=20`
4. If times are wrong, recalculate and reapply TWS

**Prevention:** Double-check window times before applying TWS. Write them down.

---

### Pods Stuck in ContainerCreating

**Symptom:** WINDOW opened, REPLICAS changed, but pods won't start.

**Quick Fix:**
1. Pre-pull image: `docker pull nginx:alpine && kind load docker-image nginx:alpine --name kyklos-dev`
2. Reduce replica count in TWS to 1 instead of 2 or 3

**Prevention:** Pre-pull images before demo. Test with demo-setup first.

---

### Recording Has Audio When It Shouldn't

**Symptom:** Recorded video has microphone audio or system sounds.

**Quick Fix:**
1. Re-record with audio input muted
2. If already recorded, use video editor to remove audio track

**Prevention:** Configure recording software to disable audio before starting. Test with 10-second trial recording.

---

## Handoff to Docs Writer

After executing demos and capturing materials, provide the Docs Writer with:

### Essential Materials
1. **Screenshots** - All capture points from both scenarios (21 total)
2. **Terminal Recordings** - Full demo recordings (Scenario A: 10 min, Scenario B: 6 min edited)
3. **Text Outputs** - kubectl outputs saved to files (TWS spec, status, events, logs)
4. **Manifests** - Exact YAML files used with actual window times

### Organization
```
/tmp/kyklos-demo-captures/
├── scenario-a/
│   ├── 01-initial-state-zero-replicas.png
│   ├── 02-tws-offhours-before-window.png
│   ├── 03-businesshours-active-two-replicas.png  ← HERO SHOT
│   └── ... (10 total)
├── scenario-b/
│   ├── B01-berlin-time-initial-state.png
│   ├── B06-wednesday-window-still-active-HERO.png  ← HERO SHOT
│   └── ... (11 total)
├── recordings/
│   ├── scenario-a-full-demo.mov
│   └── scenario-b-clips-edited.mov
└── text-outputs/
    ├── tws-spec-minute-demo.yaml
    ├── tws-status-minute-demo.json
    ├── events-minute-demo.txt
    └── controller-logs-minute-demo.txt
```

### Hero Shots to Emphasize
- **Scenario A:** Capture Point 3 - BusinessHours with 2/2 Running pods
- **Scenario B:** Capture Point 6 - Wednesday date with window still active

### Documentation Priorities
1. README needs: Scenario A Capture Point 3 (primary hero shot)
2. Advanced docs needs: Scenario B Capture Point 6 (cross-midnight proof)
3. Tutorial needs: Complete Scenario A timeline (showing progression)
4. Reference docs needs: Event timelines and controller logs

---

## Future Enhancements

### Planned Additions

1. **Scenario C: Multi-Window Day**
   - Morning ramp-up, peak hours, evening wind-down
   - 3 windows in one day with different replica counts
   - Duration: 8 hours (not practical for live demo, but valuable for docs)

2. **Scenario D: Grace Period Demo**
   - Show gradual scale-down with 5-minute grace period
   - Compare with vs without grace period
   - Duration: 12 minutes

3. **Scenario E: Holiday Handling**
   - Demonstrate treat-as-closed and treat-as-open modes
   - Show ConfigMap integration
   - Duration: 10 minutes

4. **Animated GIF Sequence**
   - Create GIFs from watch output showing scale transitions
   - 5-10 second loops for embedding in docs

5. **Interactive Demo**
   - Web-based Katacoda or Killercoda scenario
   - Users run demo in browser without local setup

---

## Contributing

When adding new demo scenarios to this directory:

1. **Follow Naming Convention:** SCENARIO-[LETTER]-[DESCRIPTIVE-NAME].md
2. **Include Duration:** Specify expected duration in overview
3. **Define Capture Points:** Number all critical screenshot moments
4. **Provide Recovery:** Include troubleshooting for common issues
5. **Update This README:** Add new scenario to overview and manifest
6. **Test Thoroughly:** Execute dry run at least twice before committing
7. **Cross-Reference:** Update CAPTURE-CHECKLIST.md with new capture points

---

## Feedback and Issues

If you encounter issues with these demo scenarios:

1. **Check Troubleshooting:** Review troubleshooting sections in each scenario document
2. **Consult Dry Run:** Run through DEMO-DRY-RUN.md to diagnose issues
3. **Review Logs:** Check controller logs for errors: `kubectl logs -n kyklos-system -l app=kyklos-controller`
4. **File Issue:** Open GitHub issue with:
   - Which scenario you were running
   - What step failed
   - Error messages or unexpected output
   - Your environment details (cluster type, Kubernetes version)

---

## Acknowledgments

These demo scenarios were designed to showcase Kyklos effectively while remaining practical and reproducible. Special considerations were made for:

- Time constraints (10-minute maximum for Scenario A)
- Reliability (UTC timezone, simple patterns)
- Visual clarity (clear state transitions)
- Documentation readiness (defined capture points)
- Accessibility (alt text guidelines)

---

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2025-10-28 | 1.0 | Initial demo scenarios package with 6 comprehensive documents |
