#!/bin/bash

# ==============================================================================
# HELM INSTALLATION SCRIPT
# ==============================================================================
# Purpose: Install Full Observability Microservices using Helm
# Usage: ./install.sh [environment]
#   environment: dev, staging, prod (default: dev)
# ==============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Default values
ENVIRONMENT=${1:-dev}
RELEASE_NAME="full-observability"
CHART_PATH="./full-observability"
NAMESPACE="observability"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Full Observability - Helm Installation${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "${YELLOW}Environment: ${ENVIRONMENT}${NC}"
echo -e "${YELLOW}Release: ${RELEASE_NAME}${NC}"
echo -e "${YELLOW}Namespace: ${NAMESPACE}${NC}"
echo ""

# ==============================================================================
# Check Helm installation
# ==============================================================================
echo -e "${BLUE}[1/6] Checking Helm installation...${NC}"
if ! command -v helm &> /dev/null; then
    echo -e "${RED}❌ Helm is not installed${NC}"
    echo "Install Helm: https://helm.sh/docs/intro/install/"
    exit 1
fi
echo -e "${GREEN}✅ Helm installed: $(helm version --short)${NC}"

# ==============================================================================
# Check kubectl connection
# ==============================================================================
echo -e "\n${BLUE}[2/6] Checking Kubernetes connection...${NC}"
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ Cannot connect to Kubernetes cluster${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Connected to cluster${NC}"

# ==============================================================================
# Create namespace if not exists
# ==============================================================================
echo -e "\n${BLUE}[3/6] Creating namespace...${NC}"
if kubectl get namespace ${NAMESPACE} &> /dev/null; then
    echo -e "${YELLOW}⚠️  Namespace '${NAMESPACE}' already exists${NC}"
else
    kubectl create namespace ${NAMESPACE}
    echo -e "${GREEN}✅ Namespace '${NAMESPACE}' created${NC}"
fi

# ==============================================================================
# Validate Helm chart
# ==============================================================================
echo -e "\n${BLUE}[4/6] Validating Helm chart...${NC}"
if ! helm lint ${CHART_PATH}; then
    echo -e "${RED}❌ Helm chart validation failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Helm chart is valid${NC}"

# ==============================================================================
# Install/Upgrade Helm release
# ==============================================================================
echo -e "\n${BLUE}[5/6] Installing Helm release...${NC}"

# Check if release exists
if helm list -n ${NAMESPACE} | grep -q ${RELEASE_NAME}; then
    echo -e "${YELLOW}⚠️  Release exists. Upgrading...${NC}"
    helm upgrade ${RELEASE_NAME} ${CHART_PATH} \
        --namespace ${NAMESPACE} \
        --values ${CHART_PATH}/values.yaml \
        --wait \
        --timeout 10m
    echo -e "${GREEN}✅ Release upgraded successfully${NC}"
else
    echo -e "${YELLOW}Installing new release...${NC}"
    helm install ${RELEASE_NAME} ${CHART_PATH} \
        --namespace ${NAMESPACE} \
        --values ${CHART_PATH}/values.yaml \
        --create-namespace \
        --wait \
        --timeout 10m
    echo -e "${GREEN}✅ Release installed successfully${NC}"
fi

# ==============================================================================
# Verify deployment
# ==============================================================================
echo -e "\n${BLUE}[6/6] Verifying deployment...${NC}"

echo -e "\n${YELLOW}Pods:${NC}"
kubectl get pods -n ${NAMESPACE}

echo -e "\n${YELLOW}Services:${NC}"
kubectl get svc -n ${NAMESPACE}

echo -e "\n${YELLOW}Deployments:${NC}"
kubectl get deployments -n ${NAMESPACE}

# ==============================================================================
# Success message
# ==============================================================================
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✅ Installation completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"

echo -e "\n${BLUE}Get release information:${NC}"
echo -e "  helm list -n ${NAMESPACE}"
echo -e "  helm status ${RELEASE_NAME} -n ${NAMESPACE}"

echo -e "\n${BLUE}View application:${NC}"
echo -e "  kubectl get all -n ${NAMESPACE}"

echo -e "\n${BLUE}Uninstall:${NC}"
echo -e "  helm uninstall ${RELEASE_NAME} -n ${NAMESPACE}"
echo ""

