#!/bin/bash
set -e

# Kyklos Quick Sanity Test
# This test creates a deployment and TWS with minute-scale windows
# to verify Kyklos is functioning correctly in under 5 minutes

NAMESPACE="kyklos-sanity"
DEPLOYMENT="test-app"
TWS_NAME="test-scaler"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Kyklos Quick Sanity Test${NC}"
echo -e "${GREEN}========================================${NC}"

# Get current time in UTC
CURRENT_TIME=$(date -u +"%H:%M")
CURRENT_MINUTE=$(date -u +"%M")
CURRENT_HOUR=$(date -u +"%H")

echo -e "\n${YELLOW}Current UTC time: ${CURRENT_TIME}${NC}"

# Calculate window times (current minute + 1, + 2, + 3)
WINDOW1_START_MIN=$(printf "%02d" $(( (10#$CURRENT_MINUTE + 1) % 60 )))
WINDOW1_END_MIN=$(printf "%02d" $(( (10#$CURRENT_MINUTE + 2) % 60 )))
WINDOW2_START_MIN=$(printf "%02d" $(( (10#$CURRENT_MINUTE + 2) % 60 )))
WINDOW2_END_MIN=$(printf "%02d" $(( (10#$CURRENT_MINUTE + 3) % 60 )))

WINDOW1_START="${CURRENT_HOUR}:${WINDOW1_START_MIN}"
WINDOW1_END="${CURRENT_HOUR}:${WINDOW1_END_MIN}"
WINDOW2_START="${CURRENT_HOUR}:${WINDOW2_START_MIN}"
WINDOW2_END="${CURRENT_HOUR}:${WINDOW2_END_MIN}"

echo -e "${YELLOW}Test windows:${NC}"
echo -e "  Window 1 (scale to 3): ${WINDOW1_START} - ${WINDOW1_END}"
echo -e "  Window 2 (scale to 5): ${WINDOW2_START} - ${WINDOW2_END}"
echo -e "  Default replicas: 1"

# Step 1: Create namespace
echo -e "\n${YELLOW}[1/5] Creating namespace...${NC}"
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Step 2: Create test deployment
echo -e "${YELLOW}[2/5] Creating test deployment...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $DEPLOYMENT
  namespace: $NAMESPACE
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        resources:
          requests:
            cpu: 10m
            memory: 32Mi
EOF

# Step 3: Wait for deployment to be ready
echo -e "${YELLOW}[3/5] Waiting for deployment to be ready...${NC}"
kubectl wait --for=condition=available --timeout=60s deployment/$DEPLOYMENT -n $NAMESPACE

INITIAL_REPLICAS=$(kubectl get deployment $DEPLOYMENT -n $NAMESPACE -o jsonpath='{.spec.replicas}')
echo -e "${GREEN}✓ Deployment ready with ${INITIAL_REPLICAS} replica(s)${NC}"

# Step 4: Create TimeWindowScaler
echo -e "${YELLOW}[4/5] Creating TimeWindowScaler with minute-scale windows...${NC}"
cat <<EOF | kubectl apply -f -
apiVersion: kyklos.kyklos.io/v1alpha1
kind: TimeWindowScaler
metadata:
  name: $TWS_NAME
  namespace: $NAMESPACE
spec:
  targetRef:
    name: $DEPLOYMENT
  timezone: "UTC"
  defaultReplicas: 1
  gracePeriodSeconds: 10
  windows:
  - name: "window-1"
    start: "$WINDOW1_START"
    end: "$WINDOW1_END"
    replicas: 3
  - name: "window-2"
    start: "$WINDOW2_START"
    end: "$WINDOW2_END"
    replicas: 5
EOF

echo -e "${GREEN}✓ TimeWindowScaler created${NC}"

# Step 5: Monitor scaling for 3 minutes
echo -e "${YELLOW}[5/5] Monitoring scaling behavior for 3 minutes...${NC}"
echo -e "${YELLOW}Expected behavior:${NC}"
echo -e "  - Now: 1 replica (default)"
echo -e "  - At ${WINDOW1_START}: Scale to 3 replicas"
echo -e "  - At ${WINDOW2_START}: Scale to 5 replicas"
echo -e "  - After ${WINDOW2_END}: Return to 1 replica (default)"
echo ""

# Watch for 3 minutes and display status
END_TIME=$(($(date +%s) + 180))
LAST_REPLICAS=0
LAST_WINDOW=""

while [ $(date +%s) -lt $END_TIME ]; do
    CURRENT_REPLICAS=$(kubectl get deployment $DEPLOYMENT -n $NAMESPACE -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")
    TWS_STATUS=$(kubectl get tws $TWS_NAME -n $NAMESPACE -o jsonpath='{.status.currentWindow}' 2>/dev/null || echo "unknown")
    CURRENT_TIME=$(date -u +"%H:%M:%S")

    if [ "$CURRENT_REPLICAS" != "$LAST_REPLICAS" ] || [ "$TWS_STATUS" != "$LAST_WINDOW" ]; then
        echo -e "[${CURRENT_TIME}] Replicas: ${GREEN}${CURRENT_REPLICAS}${NC}, Window: ${YELLOW}${TWS_STATUS}${NC}"
        LAST_REPLICAS=$CURRENT_REPLICAS
        LAST_WINDOW=$TWS_STATUS
    fi

    sleep 5
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Sanity Test Complete!${NC}"
echo -e "${GREEN}========================================${NC}"

# Verify final state
FINAL_REPLICAS=$(kubectl get deployment $DEPLOYMENT -n $NAMESPACE -o jsonpath='{.spec.replicas}')
FINAL_WINDOW=$(kubectl get tws $TWS_NAME -n $NAMESPACE -o jsonpath='{.status.currentWindow}')
TWS_READY=$(kubectl get tws $TWS_NAME -n $NAMESPACE -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')

echo -e "\n${YELLOW}Final Status:${NC}"
echo -e "  Current Replicas: ${FINAL_REPLICAS}"
echo -e "  Current Window: ${FINAL_WINDOW}"
echo -e "  TWS Ready: ${TWS_READY}"

# Check events
echo -e "\n${YELLOW}Scaling Events:${NC}"
kubectl get events -n $NAMESPACE --field-selector involvedObject.name=$TWS_NAME --sort-by='.lastTimestamp' 2>/dev/null | grep -E "ScaledUp|ScaledDown" || echo "  No scaling events found"

echo -e "\n${YELLOW}Cleanup:${NC}"
echo -e "  To remove test resources: ${GREEN}kubectl delete namespace $NAMESPACE${NC}"
echo -e "  To keep for inspection: Resources remain in namespace '${NAMESPACE}'"
echo ""

# Ask user if they want to clean up
read -p "Delete test namespace now? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    kubectl delete namespace $NAMESPACE
    echo -e "${GREEN}✓ Test namespace deleted${NC}"
else
    echo -e "${YELLOW}Test resources preserved in namespace: ${NAMESPACE}${NC}"
fi
