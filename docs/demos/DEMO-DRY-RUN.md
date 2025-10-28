# Demo Dry Run Checklist

**Project:** Kyklos Time Window Scaler
**Last Updated:** 2025-10-28
**Owner:** demo-scenario-designer

## Purpose

This checklist ensures successful demo execution by verifying all prerequisites, testing the environment, and rehearsing the demonstration flow. Use this before live demos, video recordings, or any public presentation.

---

## Pre-Rehearsal Setup (Day Before Demo)

### Environment Verification

**Cluster Health:**
```bash
# Verify cluster is running
kubectl cluster-info

# Check node status
kubectl get nodes

# Verify available resources
kubectl top nodes

# Expected: 1 node Ready, CPU/Memory available
```

**Expected Output:**
```
Kubernetes control plane is running at https://127.0.0.1:XXXXX
CoreDNS is running at https://127.0.0.1:XXXXX/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

NAME                 STATUS   ROLES           AGE   VERSION
kyklos-dev-control-plane   Ready    control-plane   5d    v1.28.0

NAME                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
kyklos-dev-control-plane   250m         12%    1500Mi          37%
```

**Checklist:**
- [ ] Cluster reachable
- [ ] Node in Ready state
- [ ] At least 1 CPU core available
- [ ] At least 1GB memory available

---

**Controller Deployment:**
```bash
# Verify controller is running
kubectl get pods -n kyklos-system

# Check controller logs for health
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=50

# Verify no recent errors
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=100 | grep -i error

# Expected: No error lines or only historical errors
```

**Expected Output:**
```
NAME                                         READY   STATUS    RESTARTS   AGE
kyklos-controller-manager-7d9c8b5f4-abc12   1/1     Running   0          2d
```

**Checklist:**
- [ ] Controller pod Running
- [ ] 0 recent restarts (< 3 in last hour)
- [ ] No ERROR or FATAL logs in last 50 lines
- [ ] Controller started successfully (check startup logs)

---

**CRD Installation:**
```bash
# Verify CRD is installed
kubectl get crd timewindowscalers.kyklos.io

# Check CRD status
kubectl get crd timewindowscalers.kyklos.io -o jsonpath='{.status.conditions[?(@.type=="Established")].status}'

# Expected: True
```

**Expected Output:**
```
NAME                              CREATED AT
timewindowscalers.kyklos.io       2025-10-26T10:15:30Z

True
```

**Checklist:**
- [ ] CRD exists
- [ ] CRD status is Established
- [ ] No validation errors in CRD definition

---

**RBAC Permissions:**
```bash
# Verify controller can manage deployments
kubectl auth can-i list deployments --as=system:serviceaccount:kyklos-system:kyklos-controller
kubectl auth can-i update deployments --as=system:serviceaccount:kyklos-system:kyklos-controller

# Verify controller can create events
kubectl auth can-i create events --as=system:serviceaccount:kyklos-system:kyklos-controller

# Verify controller can manage TWS resources
kubectl auth can-i list timewindowscalers --as=system:serviceaccount:kyklos-system:kyklos-controller
kubectl auth can-i update timewindowscalers/status --as=system:serviceaccount:kyklos-system:kyklos-controller

# Expected: All should return "yes"
```

**Checklist:**
- [ ] Can list deployments: yes
- [ ] Can update deployments: yes
- [ ] Can create events: yes
- [ ] Can list TWS: yes
- [ ] Can update TWS status: yes

---

### Demo Namespace Cleanup

```bash
# Delete any existing demo namespace
kubectl delete namespace demo --ignore-not-found=true

# Verify deletion complete (may take 10-20 seconds)
kubectl get namespace demo

# Expected: Error from server (NotFound)
```

**Checklist:**
- [ ] Demo namespace does not exist
- [ ] No orphaned resources from previous demos

---

### Terminal Setup

**Font and Size:**
```bash
# Set terminal to demo-friendly configuration
# macOS Terminal: Preferences > Profiles > Font
# iTerm2: Preferences > Profiles > Text
# Gnome Terminal: Preferences > Profile > Text

# Recommended:
# Font: Monaco, Menlo, or Source Code Pro
# Size: 14pt minimum (16pt for recordings)
# Colors: High contrast (Dark with bright colors, or Light with dark text)
```

**Checklist:**
- [ ] Terminal font size >= 14pt
- [ ] High contrast color scheme enabled
- [ ] Terminal size 120 columns x 40 rows
- [ ] Scrollback buffer cleared
- [ ] Prompt is short (no long hostname/path)

---

**Terminal Tools:**
```bash
# Verify watch is installed
which watch

# Verify date command supports timezone
TZ=Europe/Berlin date

# Verify jq is installed (for JSON parsing)
which jq

# Verify grep supports required flags
echo "test" | grep --color=auto test

# Expected: All commands found and working
```

**Checklist:**
- [ ] watch command available
- [ ] date command supports TZ variable
- [ ] jq installed (optional but recommended)
- [ ] grep supports --color

---

### Recording Setup (If Recording)

**Screen Recording Tool:**
```bash
# macOS: QuickTime Player
open -a "QuickTime Player"

# Or OBS Studio
open -a "OBS"

# Or terminal recording
which asciinema

# Configure recording settings:
# - Resolution: 1920x1080 minimum
# - Frame rate: 30fps
# - Audio: Disabled
# - Show mouse cursor: Enabled
```

**Checklist:**
- [ ] Recording software installed and tested
- [ ] Recording settings configured
- [ ] Test recording saved successfully
- [ ] Adequate disk space (5GB+ free)

---

**Screenshot Tool:**
```bash
# macOS: Built-in (Cmd+Shift+4)
# Test by taking a screenshot

# Linux: flameshot or scrot
which flameshot || which scrot

# Create capture directory
mkdir -p /tmp/kyklos-demo-captures/scenario-a
mkdir -p /tmp/kyklos-demo-captures/scenario-b

# Verify write permissions
touch /tmp/kyklos-demo-captures/test.txt && rm /tmp/kyklos-demo-captures/test.txt
```

**Checklist:**
- [ ] Screenshot tool works
- [ ] Capture directories created
- [ ] Test screenshot saved successfully
- [ ] File naming convention decided

---

### Notification and Distraction Prevention

```bash
# macOS: Enable Do Not Disturb
# System Preferences > Notifications > Do Not Disturb

# Hide desktop icons (optional)
defaults write com.apple.finder CreateDesktop false && killall Finder

# Close unnecessary applications
# Keep only: Terminal, Browser (for docs reference if needed)

# Disable Slack, email, messaging apps
```

**Checklist:**
- [ ] Do Not Disturb enabled
- [ ] All notifications disabled
- [ ] Desktop clean (hide icons if preferred)
- [ ] Only essential apps open
- [ ] Browser tabs prepared (docs, issue tracker)

---

## Dry Run: Scenario A (Minute Demo)

### Pre-Demo Checklist

**Time Calculation:**
```bash
# Get current UTC time
date -u +"%H:%M:%S"

# Calculate T+1min (window start)
# Example: If current is 14:37:45
#   Round up to next minute: 14:38:00
#   Add 1 minute: 14:38

# Calculate T+4min (window end)
# From 14:38, add 3 minutes: 14:41

# Write down your times:
# Current UTC: ___________
# Window start: ___:___
# Window end: ___:___
```

**Checklist:**
- [ ] Current UTC time noted
- [ ] Window start time calculated (T+1min)
- [ ] Window end time calculated (T+4min)
- [ ] Times written in demo notes
- [ ] At least 15 minutes available before next obligation

---

### Dry Run Execution

**Phase 1: Setup (30 seconds)**

```bash
# T-1:00 - Start dry run
date -u

# T-0:00 - Create namespace
kubectl create namespace demo

# T-0:05 - Create deployment
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp-demo
  namespace: demo
spec:
  replicas: 0
  selector:
    matchLabels:
      app: webapp-demo
  template:
    metadata:
      labels:
        app: webapp-demo
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
EOF

# T-0:10 - Verify
kubectl get deploy,pods -n demo
```

**Expected:**
- Namespace created in < 2 seconds
- Deployment created in < 5 seconds
- Deployment shows 0/0 replicas
- No pods present

**Checklist:**
- [ ] Namespace created successfully
- [ ] Deployment created successfully
- [ ] Initial state is 0/0 replicas
- [ ] Commands completed within time budget

**If any failures, stop and fix before proceeding.**

---

**Phase 2: Apply TWS (15 seconds)**

```bash
# T-0:15 - Create TWS manifest with your calculated times
cat > /tmp/demo-minute-scaler.yaml <<EOF
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
    start: "14:38"  # REPLACE WITH YOUR T+1min
    end: "14:41"    # REPLACE WITH YOUR T+4min
    replicas: 2
EOF

# T-0:20 - Apply TWS
kubectl apply -f /tmp/demo-minute-scaler.yaml

# T-0:25 - Verify TWS created
kubectl get tws -n demo
```

**Expected:**
- TWS created in < 3 seconds
- TWS shows OffHours window (before start time)
- REPLICAS shows 0

**Checklist:**
- [ ] TWS manifest created with correct times
- [ ] TWS applied successfully
- [ ] TWS shows OffHours initially
- [ ] No errors in TWS status

---

**Phase 3: Watch for Scale-Up (1-2 minutes)**

```bash
# T-0:30 - Start watching
watch -n 2 'date -u && echo && kubectl get tws,deploy,pods -n demo'

# Wait for T+1min boundary
# Observe changes:
# 1. WINDOW changes to "BusinessHours"
# 2. REPLICAS changes to 2
# 3. Pods appear in ContainerCreating
# 4. Pods transition to Running (~15 seconds)
```

**Expected Timeline:**
- T+1:05: WINDOW changes, REPLICAS → 2
- T+1:08: Pods in ContainerCreating
- T+1:15: Pods Running, 2/2 Ready

**Checklist:**
- [ ] Window opened at calculated time (±10 seconds)
- [ ] Replicas scaled to 2
- [ ] Pods created successfully
- [ ] Pods reached Running state
- [ ] Transition completed in < 20 seconds

**If scale-up didn't happen, check:**
- [ ] Current UTC time is past window start time
- [ ] Controller logs for errors: `kubectl logs -n kyklos-system -l app=kyklos-controller --tail=50`
- [ ] TWS status for errors: `kubectl get tws webapp-minute-scaler -n demo -o yaml | grep -A 10 conditions`

---

**Phase 4: Observe Steady State (1-2 minutes)**

```bash
# Watch continues running
# Observe:
# - WINDOW remains "BusinessHours"
# - REPLICAS remains 2
# - Pods remain Running
# - No unexpected restarts
```

**Expected:**
- No changes during this period
- Stable state maintained

**Checklist:**
- [ ] Window remains BusinessHours
- [ ] Replicas remain 2
- [ ] Pods remain Running with 0 restarts
- [ ] No unexpected events

---

**Phase 5: Watch for Scale-Down (1-2 minutes)**

```bash
# Watch continues running
# At T+4min:
# 1. WINDOW changes to "OffHours"
# 2. REPLICAS changes to 0
# 3. Pods enter Terminating state
# 4. Pods removed (~10 seconds)
```

**Expected Timeline:**
- T+4:05: WINDOW changes, REPLICAS → 0
- T+4:08: Pods Terminating
- T+4:15: Pods removed, 0/0 replicas

**Checklist:**
- [ ] Window closed at calculated time (±10 seconds)
- [ ] Replicas scaled to 0
- [ ] Pods terminated gracefully
- [ ] Final state is 0/0 replicas
- [ ] Transition completed in < 15 seconds

**Press Ctrl+C to stop watching.**

---

**Phase 6: Verify Events and Logs (1 minute)**

```bash
# T+5:00 - Check events
kubectl get events -n demo --sort-by='.lastTimestamp' | tail -20

# Verify scale-up events present:
# - WindowTransition: Entered window
# - ScalingTarget: Scaling from 0 to 2
# - ScaledUp: Deployment scaled up

# Verify scale-down events present:
# - WindowTransition: Exited window
# - ScalingTarget: Scaling from 2 to 0
# - ScaledDown: Deployment scaled down

# Check controller logs
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=50 | grep webapp

# Verify logs show:
# - "Matched window: BusinessHours"
# - "Scaling from 0 to 2"
# - "No matching windows, using defaultReplicas: 0"
# - "Scaling from 2 to 0"
```

**Checklist:**
- [ ] Scale-up events present and correct
- [ ] Scale-down events present and correct
- [ ] Events in correct chronological order
- [ ] Controller logs show correct decisions
- [ ] No ERROR or WARN logs related to demo

---

**Phase 7: Cleanup (15 seconds)**

```bash
# T+6:00 - Clean up
kubectl delete namespace demo

# Verify cleanup
kubectl get namespace demo
# Expected: Error from server (NotFound)
```

**Checklist:**
- [ ] Namespace deleted successfully
- [ ] Namespace removed within 15 seconds
- [ ] No orphaned resources

---

### Scenario A Dry Run Results

**Timing Verification:**
- [ ] Total time from T+0 to cleanup: ≤ 10 minutes
- [ ] Scale-up occurred within 10 seconds of window start
- [ ] Scale-down occurred within 10 seconds of window end
- [ ] All transitions completed smoothly

**Technical Verification:**
- [ ] No controller restarts during demo
- [ ] No unexpected errors in logs
- [ ] All events generated correctly
- [ ] TWS status remained healthy throughout

**Presentation Verification:**
- [ ] Watch output was clear and readable
- [ ] Commands executed without typos
- [ ] Timing was comfortable (not rushed)
- [ ] All capture points identified

**Pass/Fail:**
- [ ] PASS - Ready for live demo
- [ ] CONDITIONAL PASS - Minor issues noted, can proceed
- [ ] FAIL - Must address issues before live demo

**Issues Encountered:**
```
(List any issues observed during dry run)

1.
2.
3.
```

---

## Dry Run: Scenario B (Cross-Midnight Demo)

**IMPORTANT:** This dry run can only be performed if current time is between 21:30 and 23:30 local time (Europe/Berlin). If outside this window, perform a simulated dry run using daytime hours.

### Pre-Demo Checklist

**Time Verification:**
```bash
# Check current Berlin time
TZ=Europe/Berlin date +"%H:%M:%S %Z"

# Expected: Between 21:30 and 23:30 (CET or CEST)
```

**Checklist:**
- [ ] Current Berlin time is between 21:30 and 23:30
- [ ] At least 2.5 hours available for full demo
- [ ] OR: Willing to perform shortened version (30-minute window)

**If time check fails:**
- Proceed to "Simulated Dry Run" section below
- Or schedule dry run for evening

---

**Time Calculation (Evening):**
```bash
# Get current Berlin time
TZ=Europe/Berlin date +"%H:%M"

# Calculate window times:
# Example: If current is 22:47
#   Window start: 22:48 (next minute)
#   Window end: 00:50 (2 hours 2 minutes later, next day)

# Write down your times:
# Current Berlin: ___________
# Window start: ___:___
# Window end: ___:___ (next day)
# Minutes until midnight: _____
```

**Checklist:**
- [ ] Window start calculated (T+1min)
- [ ] Window end calculated (crosses midnight)
- [ ] End time is LESS than start time (e.g., 00:50 < 22:48)
- [ ] Minutes until midnight calculated

---

### Dry Run Execution (Evening - Full Cross-Midnight)

**Due to 2+ hour duration, perform ABBREVIATED dry run:**

**Phase 1: Setup and Window Open (2 minutes)**

```bash
# Create namespace and deployment
kubectl create namespace demo

kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nightshift-demo
  namespace: demo
spec:
  replicas: 0
  selector:
    matchLabels:
      app: nightshift-demo
  template:
    metadata:
      labels:
        app: nightshift-demo
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
EOF

# Create and apply TWS with cross-midnight window
cat > /tmp/nightshift-scaler.yaml <<EOF
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: nightshift-scaler
  namespace: demo
spec:
  targetRef:
    kind: Deployment
    name: nightshift-demo
  timezone: Europe/Berlin
  defaultReplicas: 0
  windows:
  - days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
    start: "22:48"  # REPLACE WITH YOUR CALCULATED TIME
    end: "00:50"    # REPLACE WITH YOUR CALCULATED END
    replicas: 3
EOF

kubectl apply -f /tmp/nightshift-scaler.yaml

# Start watching with Berlin time
watch -n 2 'TZ=Europe/Berlin date +"%a %Y-%m-%d %H:%M:%S %Z" && echo && kubectl get tws,deploy,pods -n demo'

# Wait for window to open
# Verify scale-up occurs
```

**Expected:**
- Window opens at calculated start time
- Replicas scale to 3
- Three pods reach Running state

**Checklist:**
- [ ] Namespace and deployment created
- [ ] TWS applied with cross-midnight window (end < start)
- [ ] Window opened successfully
- [ ] Scaled to 3 replicas
- [ ] Pods running
- [ ] Current day displayed correctly (pre-midnight)

---

**Phase 2: Pre-Midnight Verification (5 minutes before midnight)**

```bash
# Stop watch (Ctrl+C)

# Verify TWS shows cross-midnight metadata
kubectl get tws nightshift-scaler -n demo -o yaml | grep -A 15 status:

# Expected status fields:
# - currentWindow: NightShift
# - windowMetadata.crossesMidnight: true
# - windowEndDay: (next day)

# Check controller logs
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=30 | grep -A 3 nightshift

# Look for: "cross-midnight window" mentions
```

**Checklist:**
- [ ] Status shows crossesMidnight: true
- [ ] Window end day is next day
- [ ] Controller logs mention cross-midnight handling
- [ ] Window is stable and active

**At this point, you have two options:**

**Option A: Continue to Midnight (for full verification)**
- Set an alarm for 5 minutes before midnight
- Resume watch at 23:55
- Observe midnight transition at 00:00
- Verify window remains active on new day
- Continue until window end (~00:50)

**Option B: Abort Dry Run (if time is limited)**
- You've verified setup and window opening
- Cross-midnight behavior is controller-level logic
- Trust controller implementation (verified by unit tests)
- Clean up now and proceed to live demo later

---

**Phase 3: Midnight Transition Observation (Optional)**

**If continuing to midnight:**

```bash
# Resume watch at 23:55
watch -n 2 'TZ=Europe/Berlin date +"%a %Y-%m-%d %H:%M:%S %Z" && echo && kubectl get tws,deploy,pods -n demo'

# At 00:00, observe:
# 1. Date changes to next day (e.g., Tuesday → Wednesday)
# 2. WINDOW remains "NightShift" (does NOT change)
# 3. REPLICAS remains 3 (no scale change)
# 4. Pods remain Running (no disruption)

# This is the KEY verification point!
```

**Expected at Midnight:**
- Day changes in date output
- Window stays active
- No scaling action
- Pods undisturbed

**Checklist:**
- [ ] Date changed to next day
- [ ] Window remained active (did not reset)
- [ ] Replicas remained 3 (no scale change)
- [ ] Pods remained Running
- [ ] No events generated at midnight

**If window closed at midnight:**
- **STOP - This is a critical bug**
- Check window end time is truly less than start time
- Check controller logs for timezone errors
- Do not proceed with live demo until fixed

---

**Phase 4: Post-Midnight and Window End (Optional)**

```bash
# Continue watching until window end (e.g., 00:50)

# At window end, observe:
# 1. WINDOW changes to "OffHours"
# 2. REPLICAS changes to 0
# 3. Pods terminate
# 4. Date is next day (window end day)

# Check events
kubectl get events -n demo --sort-by='.lastTimestamp'

# Look for:
# - MidnightCrossed event (if implemented)
# - WindowTransition: Exited window (on next day)
```

**Expected:**
- Window closes at calculated end time (next day)
- Scale-down completes successfully
- Events show complete lifecycle

**Checklist:**
- [ ] Window closed at correct time
- [ ] Scaled down to 0
- [ ] All pods removed
- [ ] Events show cross-midnight lifecycle
- [ ] Controller logs show correct day transition

---

**Phase 5: Cleanup**

```bash
kubectl delete namespace demo
```

**Checklist:**
- [ ] Cleanup successful

---

### Scenario B Dry Run Results (Evening)

**Timing Verification:**
- [ ] Window opened at calculated time
- [ ] Midnight transition occurred smoothly
- [ ] Window closed at calculated time (next day)
- [ ] Total duration as expected

**Critical Verification:**
- [ ] **Window remained active across midnight boundary**
- [ ] **No scale change at midnight (00:00)**
- [ ] **Day changed but window did not reset**
- [ ] Controller logs show cross-midnight awareness

**Technical Verification:**
- [ ] Status shows crossesMidnight: true
- [ ] Window end day is next day
- [ ] Events show complete cross-midnight lifecycle
- [ ] No errors during entire demo

**Pass/Fail:**
- [ ] PASS - Cross-midnight handling verified
- [ ] CONDITIONAL PASS - Minor issues, can proceed
- [ ] FAIL - Midnight transition failed, do not proceed

---

### Simulated Dry Run (Daytime Alternative)

**Use if outside evening hours or for quick verification.**

**Strategy:** Use current-hour windows to test mechanics (not true midnight cross).

```bash
# Get current UTC time
date -u +"%H:%M"

# Example: 14:37

# Create "simulated cross-midnight" window
# Start: 14:38 (T+1min)
# End: 14:41 (T+4min)
# This does NOT actually cross midnight, but tests window mechanics

cat > /tmp/nightshift-scaler-simulated.yaml <<EOF
apiVersion: kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: nightshift-scaler
  namespace: demo
spec:
  targetRef:
    kind: Deployment
    name: nightshift-demo
  timezone: Europe/Berlin
  defaultReplicas: 0
  windows:
  - days: [Mon, Tue, Wed, Thu, Fri, Sat, Sun]
    start: "14:38"  # Current hour + 1 minute
    end: "14:41"    # Current hour + 4 minutes (NOT cross-midnight!)
    replicas: 3
EOF

# Run same steps as Scenario A
# This verifies:
# - Setup works
# - Controller reconciles
# - Scaling works
# - Events generated

# This does NOT verify:
# - Actual midnight crossing
# - Calendar day transition
# - Cross-day window logic
```

**Simulated Dry Run Checklist:**
- [ ] Setup successful
- [ ] Window opened and closed
- [ ] Scaling worked correctly
- [ ] Events generated correctly
- [ ] Controller logs healthy

**Note:** This does NOT replace true evening dry run for verifying cross-midnight logic.

---

## Common Issues and Fixes

### Issue: Window didn't open at calculated time

**Symptoms:**
- Passed calculated start time by 30+ seconds
- WINDOW still shows "OffHours"
- REPLICAS still 0

**Diagnosis:**
```bash
# Check TWS spec
kubectl get tws -n demo -o yaml | grep -A 10 windows:

# Check controller logs
kubectl logs -n kyklos-system -l app=kyklos-controller --tail=50

# Check current time vs window times
date -u && kubectl get tws -n demo -o jsonpath='{.spec.windows[0]}'
```

**Common Causes:**
1. Window times in manifest are incorrect (typo, wrong time)
2. Timezone mismatch (using local time instead of UTC)
3. Controller not reconciling (check logs for errors)
4. Window days don't include today

**Fix:**
```bash
# Fix 1: Recalculate and reapply with correct times
kubectl delete tws -n demo --all
# Edit manifest with new times (T+1min from now)
kubectl apply -f /tmp/demo-minute-scaler.yaml

# Fix 2: Verify timezone in manifest matches demo scenario
# Scenario A: Must be "UTC"
# Scenario B: Must be "Europe/Berlin"

# Fix 3: Restart controller if not reconciling
kubectl rollout restart deployment -n kyklos-system kyklos-controller-manager
```

---

### Issue: Pods stuck in ContainerCreating

**Symptoms:**
- WINDOW changed correctly
- REPLICAS changed correctly
- Pods created but stuck in ContainerCreating for > 30 seconds

**Diagnosis:**
```bash
# Describe pods
kubectl describe pods -n demo | grep -A 10 Events

# Check node resources
kubectl top nodes
```

**Common Causes:**
1. Image pull errors (nginx:alpine not available)
2. Insufficient node resources

**Fix:**
```bash
# Fix 1: Pre-pull image
docker pull nginx:alpine
kind load docker-image nginx:alpine --name kyklos-dev

# Fix 2: Reduce replicas
# Edit TWS to use replicas: 1 instead of 2 or 3

# Fix 3: Check node has capacity
kubectl describe node | grep -A 5 "Allocated resources"
```

---

### Issue: Controller logs show timezone errors

**Symptoms:**
- Window didn't open despite correct times
- Controller logs show "unknown time zone" or similar

**Diagnosis:**
```bash
kubectl logs -n kyklos-system -l app=kyklos-controller | grep -i timezone
```

**Common Causes:**
1. Typo in timezone name (e.g., "europe/berlin" instead of "Europe/Berlin")
2. Controller container missing timezone database

**Fix:**
```bash
# Fix 1: Verify timezone name is correct (case-sensitive)
# Correct: Europe/Berlin
# Wrong: europe/berlin, EUROPE/BERLIN

# Fix 2: Check if timezone is in IANA database
TZ=Europe/Berlin date
# If this works on your machine but not in controller, rebuild controller with timezone data

# Temporary workaround: Use UTC for demos until fixed
```

---

### Issue: Events not appearing

**Symptoms:**
- Scaling happened but no events shown
- `kubectl get events` returns empty or missing key events

**Diagnosis:**
```bash
# Check controller has event creation permissions
kubectl auth can-i create events -n demo --as=system:serviceaccount:kyklos-system:kyklos-controller

# Check for RBAC errors in logs
kubectl logs -n kyklos-system -l app=kyklos-controller | grep -i "forbidden\|unauthorized"
```

**Fix:**
```bash
# If permission issue:
make verify-rbac
make deploy  # Reapply RBAC config

# If event age issue:
# Events older than 1 hour may not show in kubectl get events
# Use --all-namespaces and --sort-by to find older events
kubectl get events --all-namespaces --sort-by='.lastTimestamp' | grep demo
```

---

## Final Dry Run Checklist

Before marking dry run as complete:

### Scenario A (Minute Demo)
- [ ] Environment verified and healthy
- [ ] Full dry run executed successfully
- [ ] Scale-up occurred at correct time
- [ ] Scale-down occurred at correct time
- [ ] All events generated correctly
- [ ] Controller logs show correct decisions
- [ ] All capture points identified
- [ ] Total time < 10 minutes
- [ ] No critical issues encountered

### Scenario B (Cross-Midnight Demo)
- [ ] Evening timing confirmed (if doing full demo)
- [ ] Setup and window opening verified
- [ ] Cross-midnight window specification correct (end < start)
- [ ] Status shows crossesMidnight metadata
- [ ] Controller detects cross-midnight window
- [ ] **If did full demo:** Midnight transition verified successfully
- [ ] **If simulated:** Mechanics verified, understand true midnight test needed later
- [ ] All capture points identified

### Recording Setup (If Recording)
- [ ] Screen recording tested and working
- [ ] Screenshot tool tested and working
- [ ] Capture directory prepared
- [ ] File naming convention decided
- [ ] Adequate disk space available

### Presentation Readiness
- [ ] Terminal configured for demo
- [ ] Notifications disabled
- [ ] Desktop clean
- [ ] Second terminal ready for logs
- [ ] Demo notes prepared with calculated times
- [ ] Recovery procedures understood

### Documentation
- [ ] Issues log updated with any problems found
- [ ] Timing notes recorded (actual vs expected)
- [ ] Screenshots taken during dry run (optional)
- [ ] Command history saved (for reference)

---

## Post-Dry-Run Actions

### If Dry Run PASSED

**Confidence Level:**
- [ ] High confidence - Ready for live demo or recording
- [ ] Medium confidence - Minor issues noted, acceptable for demo
- [ ] Low confidence - Consider additional dry run

**Next Steps:**
1. Schedule live demo or recording session
2. Prepare demo materials (manifests with calculated times)
3. Set up recording equipment (if recording)
4. Review capture checklist
5. Perform final environment check 15 minutes before demo

---

### If Dry Run FAILED or Had Issues

**Issue Severity:**
- [ ] Critical - Demo cannot proceed (e.g., controller broken, midnight transition failed)
- [ ] Major - Demo can proceed but with degraded experience (e.g., timing off, missing events)
- [ ] Minor - Demo can proceed with acceptable workarounds (e.g., slow pod startup)

**Action Required:**
- **Critical:** Fix before scheduling live demo
- **Major:** Assess if acceptable for target audience, fix if time permits
- **Minor:** Document workaround, proceed with demo

**Remediation Checklist:**
- [ ] Issues documented in detail
- [ ] Root cause identified
- [ ] Fix implemented
- [ ] Fix tested (mini dry run)
- [ ] New dry run scheduled if needed

---

## Rehearsal Notes Template

```
Demo Dry Run Report
Date: ___________
Scenario: [ ] A - Minute Demo  [ ] B - Cross-Midnight Demo
Operator: ___________

ENVIRONMENT STATUS:
Cluster: _________ (kind/k3d)
Kubernetes version: _________
Controller version: _________
Node resources: CPU _____ / Memory _____

PRE-CHECKS:
[ ] Cluster healthy
[ ] Controller running
[ ] CRDs installed
[ ] RBAC verified
[ ] Terminal configured
[ ] Recording setup (if applicable)

TIMING DATA:
Calculated window start: _________
Calculated window end: _________
Actual scale-up time: _________
Actual scale-down time: _________
Total demo duration: _________

Scale-up delay: _____ seconds (expected: < 10s)
Scale-down delay: _____ seconds (expected: < 10s)

SCENARIO-SPECIFIC (Scenario B only):
Midnight crossing time: _________
Window remained active: [ ] Yes [ ] No
Day transition observed: [ ] Yes [ ] No

ISSUES ENCOUNTERED:
1. _________
2. _________
3. _________

OVERALL RESULT:
[ ] PASS - Ready for live demo
[ ] CONDITIONAL PASS - Can proceed with noted issues
[ ] FAIL - Must address issues before proceeding

NOTES:
_________
_________
_________

NEXT STEPS:
_________
_________
_________
```

---

## Revision History

| Date | Version | Changes |
|------|---------|---------|
| 2025-10-28 | 1.0 | Initial dry run checklist for both scenarios |
