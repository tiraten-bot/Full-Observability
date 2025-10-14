#!/bin/bash

# ==============================================================================
# CREATE CONFIGMAPS SCRIPT
# ==============================================================================
# Purpose: Apply all ConfigMap configurations to Kubernetes cluster
# Usage: ./create-configmaps.sh
# ==============================================================================

set -e  # Exit on error

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Creating Kubernetes ConfigMaps${NC}"
echo -e "${GREEN}========================================${NC}"

# ==============================================================================
# Check namespace
# ==============================================================================
echo -e "\n${YELLOW}[1/3] Checking namespace...${NC}"
if ! kubectl get namespace observability &> /dev/null; then
    echo -e "${RED}❌ Namespace 'observability' does not exist${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Namespace exists${NC}"

# ==============================================================================
# Apply ConfigMaps
# ==============================================================================
echo -e "\n${YELLOW}[2/3] Creating ConfigMaps...${NC}"
kubectl apply -f init-db-scripts.yaml
echo -e "${GREEN}✅ ConfigMaps created${NC}"

# ==============================================================================
# Verify
# ==============================================================================
echo -e "\n${YELLOW}[3/3] Verifying ConfigMaps...${NC}"
echo ""
kubectl get configmaps -n observability

echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}ConfigMap Details:${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

echo -e "\n${YELLOW}Init Scripts ConfigMap:${NC}"
kubectl get configmap postgres-init-scripts -n observability -o jsonpath='{.data}' | jq 'keys'

echo -e "\n${YELLOW}App Config:${NC}"
kubectl get configmap app-config -n observability -o jsonpath='{.data}' | jq 'keys'

echo -e "\n${YELLOW}Prometheus Config:${NC}"
kubectl get configmap prometheus-config -n observability -o jsonpath='{.data}' | jq 'keys'

echo -e "\n${YELLOW}Grafana Datasources:${NC}"
kubectl get configmap grafana-datasources -n observability -o jsonpath='{.data}' | jq 'keys'

# ==============================================================================
# Completion
# ==============================================================================
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✅ All ConfigMaps created successfully!${NC}"
echo -e "${GREEN}========================================${NC}"

echo -e "\n${BLUE}Useful commands:${NC}"
echo -e "  ${YELLOW}kubectl get configmaps -n observability${NC}"
echo -e "  ${YELLOW}kubectl describe configmap app-config -n observability${NC}"
echo -e "  ${YELLOW}kubectl get configmap prometheus-config -n observability -o yaml${NC}"
echo ""

