#!/bin/bash

# ==============================================================================
# BLUE-GREEN DEPLOYMENT
# ==============================================================================
# Purpose: Instant version switch with rollback capability
# Usage: ./blue-green-deploy.sh <service-name> <new-version> <action>
# Actions: deploy, switch, rollback, cleanup
# Example: 
#   ./blue-green-deploy.sh user-service v2.0.0 deploy
#   ./blue-green-deploy.sh user-service v2.0.0 switch
#   ./blue-green-deploy.sh user-service v2.0.0 rollback
# ==============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

SERVICE_NAME=${1}
NEW_VERSION=${2}
ACTION=${3:-deploy}
NAMESPACE="observability"

if [ -z "$SERVICE_NAME" ] || [ -z "$NEW_VERSION" ]; then
    echo "Usage: $0 <service-name> <new-version> [action]"
    echo "Actions: deploy, switch, rollback, cleanup"
    echo "Example: $0 user-service v2.0.0 deploy"
    exit 1
fi

echo -e "${CYAN}╔════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                  BLUE-GREEN DEPLOYMENT                             ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Service: ${SERVICE_NAME}${NC}"
echo -e "${YELLOW}Version: ${NEW_VERSION}${NC}"
echo -e "${YELLOW}Action: ${ACTION}${NC}"
echo ""

# ==============================================================================
# ACTION: DEPLOY GREEN VERSION
# ==============================================================================
if [ "$ACTION" == "deploy" ]; then
    echo -e "${BLUE}[1/3] Deploying GREEN version (${NEW_VERSION})...${NC}"
    
    cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${SERVICE_NAME}-green
  namespace: ${NAMESPACE}
  labels:
    app: ${SERVICE_NAME}
    version: green
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ${SERVICE_NAME}
      color: green
  template:
    metadata:
      labels:
        app: ${SERVICE_NAME}
        version: v2
        color: green
    spec:
      containers:
        - name: ${SERVICE_NAME}
          image: your-registry/${SERVICE_NAME}:${NEW_VERSION}
          # ... same spec as blue version
EOF
    
    echo -e "${YELLOW}Waiting for green deployment to be ready...${NC}"
    kubectl rollout status deployment/${SERVICE_NAME}-green -n ${NAMESPACE}
    
    echo -e "\n${BLUE}[2/3] Creating DestinationRule with subsets...${NC}"
    
    cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: ${SERVICE_NAME}-blue-green-dr
  namespace: ${NAMESPACE}
spec:
  host: ${SERVICE_NAME}
  subsets:
    - name: blue
      labels:
        color: blue
    - name: green
      labels:
        color: green
EOF
    
    echo -e "\n${BLUE}[3/3] Creating VirtualService (100% BLUE)...${NC}"
    
    cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: ${SERVICE_NAME}-blue-green-vs
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
            subset: blue
          weight: 100
EOF
    
    echo -e "\n${GREEN}╔════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║              GREEN VERSION DEPLOYED! ✅                            ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Current state:${NC}"
    echo -e "  🔵 BLUE (v1): 100% traffic (active)"
    echo -e "  🟢 GREEN (${NEW_VERSION}): 0% traffic (standby)"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo -e "  1. Test green version: kubectl port-forward deployment/${SERVICE_NAME}-green 8080:8080 -n ${NAMESPACE}"
    echo -e "  2. When ready to switch: ./blue-green-deploy.sh ${SERVICE_NAME} ${NEW_VERSION} switch"
    echo -e "  3. If problems: ./blue-green-deploy.sh ${SERVICE_NAME} ${NEW_VERSION} rollback"
    echo ""

# ==============================================================================
# ACTION: SWITCH TO GREEN
# ==============================================================================
elif [ "$ACTION" == "switch" ]; then
    echo -e "${BLUE}Switching to GREEN version (100% traffic)...${NC}"
    
    kubectl patch virtualservice ${SERVICE_NAME}-blue-green-vs -n ${NAMESPACE} --type merge -p '
{
  "spec": {
    "http": [{
      "route": [{
        "destination": {
          "host": "'${SERVICE_NAME}'",
          "subset": "green"
        },
        "weight": 100
      }]
    }]
  }
}'
    
    echo -e "\n${GREEN}╔════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║              SWITCHED TO GREEN! ✅                                 ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Current state:${NC}"
    echo -e "  🔵 BLUE (v1): 0% traffic (standby for rollback)"
    echo -e "  🟢 GREEN (${NEW_VERSION}): 100% traffic (active)"
    echo ""
    echo -e "${YELLOW}Monitor metrics in Grafana/Prometheus${NC}"
    echo -e "${YELLOW}If problems: ./blue-green-deploy.sh ${SERVICE_NAME} ${NEW_VERSION} rollback${NC}"
    echo ""

# ==============================================================================
# ACTION: ROLLBACK TO BLUE
# ==============================================================================
elif [ "$ACTION" == "rollback" ]; then
    echo -e "${RED}Rolling back to BLUE version (stable)...${NC}"
    
    kubectl patch virtualservice ${SERVICE_NAME}-blue-green-vs -n ${NAMESPACE} --type merge -p '
{
  "spec": {
    "http": [{
      "route": [{
        "destination": {
          "host": "'${SERVICE_NAME}'",
          "subset": "blue"
        },
        "weight": 100
      }]
    }]
  }
}'
    
    echo -e "\n${YELLOW}╔════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║              ROLLED BACK TO BLUE! ⏮️                               ║${NC}"
    echo -e "${YELLOW}╚════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Current state:${NC}"
    echo -e "  🔵 BLUE (v1): 100% traffic (active)"
    echo -e "  🟢 GREEN (${NEW_VERSION}): 0% traffic (failed)"
    echo ""

# ==============================================================================
# ACTION: CLEANUP - Remove old version
# ==============================================================================
elif [ "$ACTION" == "cleanup" ]; then
    echo -e "${BLUE}Cleaning up BLUE version (removing old)...${NC}"
    
    # Delete blue deployment
    kubectl delete deployment ${SERVICE_NAME} -n ${NAMESPACE} || true
    
    # Rename green to blue (make it the new stable)
    kubectl patch deployment ${SERVICE_NAME}-green -n ${NAMESPACE} -p '
{
  "metadata": {
    "name": "'${SERVICE_NAME}'"
  },
  "spec": {
    "template": {
      "metadata": {
        "labels": {
          "color": "blue",
          "version": "v1"
        }
      }
    }
  }
}'
    
    # Update VirtualService to use blue only
    kubectl patch virtualservice ${SERVICE_NAME}-blue-green-vs -n ${NAMESPACE} --type merge -p '
{
  "spec": {
    "http": [{
      "route": [{
        "destination": {
          "host": "'${SERVICE_NAME}'",
          "subset": "blue"
        },
        "weight": 100
      }]
    }]
  }
}'
    
    echo -e "\n${GREEN}╔════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║              CLEANUP COMPLETED! 🧹                                 ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}GREEN promoted to BLUE (new stable)${NC}"
    echo -e "${CYAN}Old BLUE version removed${NC}"
    echo ""

else
    echo -e "${RED}Unknown action: ${ACTION}${NC}"
    echo "Valid actions: deploy, switch, rollback, cleanup"
    exit 1
fi

