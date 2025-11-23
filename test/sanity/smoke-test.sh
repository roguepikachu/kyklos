#!/bin/bash
set -e

# Kyklos Smoke Test - Ultra-fast basic functionality check
# Verifies Kyklos can create resources and compute scaling decisions
# Takes ~30 seconds

NAMESPACE="kyklos-smoke"
DEPLOYMENT="smoke-app"
TWS_NAME="smoke-scaler"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Kyklos Smoke Test (30 seconds)${NC}"
echo -e "${GREEN}========================================${NC}"

# Get current time and create a window that's active now
CURRENT_HOUR=$(date -u +"%H")
CURRENT_MINUTE=$(date -u +"%M")
WINDOW_START=$(printf "%02d:%02d" $((10#$CURRENT_HOUR)) $((10#$CURRENT_MINUTE - 1)))
WINDOW_END=$(printf "%02d:%02d" $((10#$CURRENT_HOUR)) $((10#$CURRENT_MINUTE + 10)))

echo -e "\n${YELLOW}Creating active window: ${WINDOW_START} - ${WINDOW_END}${NC}"
echo -e "${YELLOW}Expected: Deployment should scale to 5 replicas immediately${NC}\n"

# Cleanup any existing test
kubectl delete namespace $NAMESPACE --wait=false 2>/dev/null || true
sleep 2

# Create namespace
echo -e "${YELLOW}[1/4] Creating test namespace...${NC}"
kubectl create namespace $NAMESPACE

# Create deployment
echo -e "${YELLOW}[2/4] Creating deployment (1 replica)...${NC}"
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $DEPLOYMENT
  namespace: $NAMESPACE
spec:
  replicas: 1
  selector:
    matchLabels:
      app: smoke
  template:
    metadata:
      labels:
        app: smoke
    spec:
      containers:
      - name: pause
        image: registry.k8s.io/pause:3.9
        resources:
          requests:
            cpu: 1m
            memory: 8Mi
EOF

# Create TWS with active window
echo -e "${YELLOW}[3/4] Creating TimeWindowScaler...${NC}"
kubectl apply -f - <<EOF
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
  windows:
  - name: "active-now"
    start: "$WINDOW_START"
    end: "$WINDOW_END"
    replicas: 5
EOF

# Wait and verify
echo -e "${YELLOW}[4/4] Waiting for scaling (max 20s)...${NC}"
for i in {1..20}; do
    REPLICAS=$(kubectl get deployment $DEPLOYMENT -n $NAMESPACE -o jsonpath='{.spec.replicas}' 2>/dev/null || echo "0")
    if [ "$REPLICAS" = "5" ]; then
        echo -e "${GREEN}✓ SUCCESS: Deployment scaled to 5 replicas in ${i}s${NC}"
        break
    fi
    echo -n "."
    sleep 1
done
echo ""

# Verify status
FINAL_REPLICAS=$(kubectl get deployment $DEPLOYMENT -n $NAMESPACE -o jsonpath='{.spec.replicas}')
TWS_WINDOW=$(kubectl get tws $TWS_NAME -n $NAMESPACE -o jsonpath='{.status.currentWindow}')
TWS_EFFECTIVE=$(kubectl get tws $TWS_NAME -n $NAMESPACE -o jsonpath='{.status.effectiveReplicas}')
TWS_READY=$(kubectl get tws $TWS_NAME -n $NAMESPACE -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}Results:${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "Deployment Replicas: ${FINAL_REPLICAS}/5"
echo -e "TWS Current Window: ${TWS_WINDOW}"
echo -e "TWS Effective Replicas: ${TWS_EFFECTIVE}"
echo -e "TWS Ready Status: ${TWS_READY}"

# Check for scaling event
SCALED_EVENT=$(kubectl get events -n $NAMESPACE --field-selector involvedObject.name=$TWS_NAME 2>/dev/null | grep -c "ScaledUp" || echo "0")

if [ "$FINAL_REPLICAS" = "5" ] && [ "$TWS_READY" = "True" ] && [ "$SCALED_EVENT" -gt "0" ]; then
    echo -e "\n${GREEN}✓✓✓ SMOKE TEST PASSED ✓✓✓${NC}"
    echo -e "${GREEN}Kyklos is functioning correctly!${NC}"
    RESULT=0
else
    echo -e "\n${RED}✗✗✗ SMOKE TEST FAILED ✗✗✗${NC}"
    echo -e "${RED}Expected: 5 replicas, Ready=True, ScaledUp event${NC}"
    echo -e "${RED}Got: ${FINAL_REPLICAS} replicas, Ready=${TWS_READY}, Events=${SCALED_EVENT}${NC}"
    RESULT=1
fi

# Cleanup
echo -e "\n${YELLOW}Cleaning up...${NC}"
kubectl delete namespace $NAMESPACE --wait=false
echo -e "${GREEN}✓ Test namespace deleted${NC}\n"

exit $RESULT
