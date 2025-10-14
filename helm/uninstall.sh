#!/bin/bash

# ==============================================================================
# HELM UNINSTALLATION SCRIPT
# ==============================================================================
# Purpose: Uninstall Full Observability Microservices
# Usage: ./uninstall.sh
# ==============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

RELEASE_NAME="full-observability"
NAMESPACE="observability"

echo -e "${RED}========================================${NC}"
echo -e "${RED}Full Observability - Uninstallation${NC}"
echo -e "${RED}========================================${NC}"
echo -e "${YELLOW}Release: ${RELEASE_NAME}${NC}"
echo -e "${YELLOW}Namespace: ${NAMESPACE}${NC}"
echo ""

# Confirmation
read -p "Are you sure you want to uninstall? This will delete all resources! (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo -e "${YELLOW}Uninstallation cancelled${NC}"
    exit 0
fi

# Uninstall Helm release
echo -e "\n${YELLOW}Uninstalling Helm release...${NC}"
if helm list -n ${NAMESPACE} | grep -q ${RELEASE_NAME}; then
    helm uninstall ${RELEASE_NAME} -n ${NAMESPACE}
    echo -e "${GREEN}✅ Release uninstalled${NC}"
else
    echo -e "${YELLOW}⚠️  Release not found${NC}"
fi

# Optionally delete namespace
read -p "Delete namespace '${NAMESPACE}'? (yes/no): " DELETE_NS
if [ "$DELETE_NS" == "yes" ]; then
    kubectl delete namespace ${NAMESPACE}
    echo -e "${GREEN}✅ Namespace deleted${NC}"
fi

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✅ Uninstallation completed${NC}"
echo -e "${GREEN}========================================${NC}"

