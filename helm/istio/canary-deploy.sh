#!/bin/bash

# ==============================================================================
# CANARY DEPLOYMENT AUTOMATION
# ==============================================================================
# Purpose: Automate canary deployment with traffic shifting
# Usage: ./canary-deploy.sh <service-name> <new-version-tag>
# Example: ./canary-deploy.sh user-service v2.0.0
# ==============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Configuration
SERVICE_NAME=${1}
NEW_VERSION=${2}
NAMESPACE="observability"

if [ -z "$SERVICE_NAME" ] || [ -z "$NEW_VERSION" ]; then
    echo "Usage: $0 <service-name> <new-version-tag>"
    echo "Example: $0 user-service v2.0.0"
    exit 1
fi

echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                    CANARY DEPLOYMENT                               ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

echo -e "${YELLOW}Service: ${SERVICE_NAME}${NC}"
echo -e "${YELLOW}New Version: ${NEW_VERSION}${NC}"
echo -e "${YELLOW}Namespace: ${NAMESPACE}${NC}"
echo ""

# ==============================================================================
# STAGE 1: Deploy Canary Version (0% traffic)
# ==============================================================================
echo -e "${BLUE}[Stage 1/6] Deploying canary version (0% traffic)...${NC}"

# Create canary deployment (example for user-service)
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${SERVICE_NAME}-canary
  namespace: ${NAMESPACE}
  labels:
    app: ${SERVICE_NAME}
    version: canary
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${SERVICE_NAME}
      version: canary
  template:
    metadata:
      labels:
        app: ${SERVICE_NAME}
        version: canary
      annotations:
        prometheus.io/scrape: "true"
    spec:
      containers:
        - name: ${SERVICE_NAME}
          image: your-registry/${SERVICE_NAME}:${NEW_VERSION}
          # ... rest of container spec same as stable
EOF

echo -e "${YELLOW}Waiting for canary pods to be ready...${NC}"
kubectl rollout status deployment/${SERVICE_NAME}-canary -n ${NAMESPACE}

echo -e "${GREEN}✅ Canary version deployed (0% traffic)${NC}"

# ==============================================================================
# STAGE 2: Route 10% traffic to canary
# ==============================================================================
echo -e "\n${BLUE}[Stage 2/6] Routing 10% traffic to canary...${NC}"

cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: ${SERVICE_NAME}-canary-vs
  namespace: ${NAMESPACE}
spec:
  hosts:
    - ${SERVICE_NAME}
  gateways:
    - mesh
  http:
    - route:
        - destination:
            host: ${SERVICE_NAME}
            subset: v1
          weight: 90
        - destination:
            host: ${SERVICE_NAME}
            subset: canary
          weight: 10
EOF

echo -e "${GREEN}✅ 10% traffic routed to canary${NC}"
echo -e "${YELLOW}Monitoring for 2 minutes...${NC}"
sleep 120

# Check error rate
ERROR_RATE=$(kubectl exec -n ${NAMESPACE} deployment/prometheus -- \
  wget -qO- "http://localhost:9090/api/v1/query?query=rate(istio_requests_total{destination_app=\"${SERVICE_NAME}\",destination_version=\"canary\",response_code=~\"5..\"}[2m])" 2>/dev/null | grep -o '"value":\[[^]]*\]' | grep -o '[0-9.]*$' || echo "0")

echo -e "${CYAN}Error rate (canary): ${ERROR_RATE}${NC}"

if (( $(echo "$ERROR_RATE > 0.05" | bc -l) )); then
    echo -e "${RED}❌ Error rate too high! Rolling back...${NC}"
    # Rollback to 100% v1
    kubectl delete virtualservice ${SERVICE_NAME}-canary-vs -n ${NAMESPACE}
    kubectl delete deployment ${SERVICE_NAME}-canary -n ${NAMESPACE}
    exit 1
fi

# ==============================================================================
# STAGE 3: Increase to 30%
# ==============================================================================
echo -e "\n${BLUE}[Stage 3/6] Increasing to 30% traffic...${NC}"

kubectl patch virtualservice ${SERVICE_NAME}-canary-vs -n ${NAMESPACE} --type merge -p '
{
  "spec": {
    "http": [{
      "route": [
        {"destination": {"host": "'${SERVICE_NAME}'", "subset": "v1"}, "weight": 70},
        {"destination": {"host": "'${SERVICE_NAME}'", "subset": "canary"}, "weight": 30}
      ]
    }]
  }
}'

echo -e "${GREEN}✅ 30% traffic to canary${NC}"
sleep 120

# ==============================================================================
# STAGE 4: Increase to 50%
# ==============================================================================
echo -e "\n${BLUE}[Stage 4/6] Increasing to 50% traffic...${NC}"

kubectl patch virtualservice ${SERVICE_NAME}-canary-vs -n ${NAMESPACE} --type merge -p '
{
  "spec": {
    "http": [{
      "route": [
        {"destination": {"host": "'${SERVICE_NAME}'", "subset": "v1"}, "weight": 50},
        {"destination": {"host": "'${SERVICE_NAME}'", "subset": "canary"}, "weight": 50}
      ]
    }]
  }
}'

echo -e "${GREEN}✅ 50% traffic to canary${NC}"
sleep 120

# ==============================================================================
# STAGE 5: Increase to 100% (full rollout)
# ==============================================================================
echo -e "\n${BLUE}[Stage 5/6] Full rollout (100% traffic)...${NC}"

kubectl patch virtualservice ${SERVICE_NAME}-canary-vs -n ${NAMESPACE} --type merge -p '
{
  "spec": {
    "http": [{
      "route": [
        {"destination": {"host": "'${SERVICE_NAME}'", "subset": "canary"}, "weight": 100}
      ]
    }]
  }
}'

echo -e "${GREEN}✅ 100% traffic to canary${NC}"
sleep 60

# ==============================================================================
# STAGE 6: Cleanup - Make canary the new stable
# ==============================================================================
echo -e "\n${BLUE}[Stage 6/6] Promoting canary to stable...${NC}"

# Scale down old version
kubectl scale deployment ${SERVICE_NAME} -n ${NAMESPACE} --replicas=0

# Rename canary to stable (or update labels)
kubectl label deployment ${SERVICE_NAME}-canary -n ${NAMESPACE} version=v1 --overwrite
kubectl label deployment ${SERVICE_NAME}-canary -n ${NAMESPACE} version-

# Update stable deployment with new image
kubectl set image deployment/${SERVICE_NAME} -n ${NAMESPACE} \
  ${SERVICE_NAME}=your-registry/${SERVICE_NAME}:${NEW_VERSION}

# Scale up stable
kubectl scale deployment ${SERVICE_NAME} -n ${NAMESPACE} --replicas=3

# Remove canary
kubectl delete deployment ${SERVICE_NAME}-canary -n ${NAMESPACE}
kubectl delete virtualservice ${SERVICE_NAME}-canary-vs -n ${NAMESPACE}

echo -e "\n${CYAN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║              CANARY DEPLOYMENT COMPLETED! 🎉                       ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

echo -e "${GREEN}✅ ${SERVICE_NAME} successfully upgraded to ${NEW_VERSION}${NC}"
echo ""

