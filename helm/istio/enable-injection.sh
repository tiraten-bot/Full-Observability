#!/bin/bash

# ==============================================================================
# ENABLE ISTIO SIDECAR INJECTION
# ==============================================================================
# Purpose: Enable automatic Envoy sidecar injection in namespace
# Usage: ./enable-injection.sh [namespace]
# ==============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
NAMESPACE=${1:-observability}

echo -e "${CYAN}"
echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║           ENABLE ISTIO SIDECAR INJECTION                         ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# ==============================================================================
# STEP 1: Check Istio Installation
# ==============================================================================
echo -e "${BLUE}[1/5] Checking Istio installation...${NC}"

if ! command -v istioctl &> /dev/null; then
    echo -e "${RED}❌ istioctl not found${NC}"
    echo -e "${YELLOW}Run: ./install-istio.sh${NC}"
    exit 1
fi

if ! kubectl get namespace istio-system &> /dev/null; then
    echo -e "${RED}❌ Istio not installed${NC}"
    echo -e "${YELLOW}Run: ./install-istio.sh${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Istio is installed${NC}"

# ==============================================================================
# STEP 2: Check Namespace
# ==============================================================================
echo -e "\n${BLUE}[2/5] Checking namespace '${NAMESPACE}'...${NC}"

if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
    echo -e "${RED}❌ Namespace '${NAMESPACE}' not found${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Namespace '${NAMESPACE}' exists${NC}"

# ==============================================================================
# STEP 3: Label Namespace for Injection
# ==============================================================================
echo -e "\n${BLUE}[3/5] Enabling sidecar injection...${NC}"

# Check current label
CURRENT_LABEL=$(kubectl get namespace ${NAMESPACE} -o jsonpath='{.metadata.labels.istio-injection}' 2>/dev/null || echo "")

if [ "$CURRENT_LABEL" == "enabled" ]; then
    echo -e "${YELLOW}⚠️  Sidecar injection already enabled${NC}"
else
    kubectl label namespace ${NAMESPACE} istio-injection=enabled --overwrite
    echo -e "${GREEN}✅ Sidecar injection enabled${NC}"
fi

# Verify label
echo -e "\n${YELLOW}Namespace labels:${NC}"
kubectl get namespace ${NAMESPACE} --show-labels

# ==============================================================================
# STEP 4: Restart Existing Pods
# ==============================================================================
echo -e "\n${BLUE}[4/5] Restarting pods to inject sidecars...${NC}"

# Get deployments in namespace
DEPLOYMENTS=$(kubectl get deployments -n ${NAMESPACE} -o jsonpath='{.items[*].metadata.name}')

if [ -z "$DEPLOYMENTS" ]; then
    echo -e "${YELLOW}⚠️  No deployments found in namespace${NC}"
else
    echo -e "${YELLOW}Deployments to restart:${NC}"
    echo "${DEPLOYMENTS}" | tr ' ' '\n' | sed 's/^/  - /'
    
    echo -e "\n${YELLOW}Restarting deployments...${NC}"
    for DEPLOY in ${DEPLOYMENTS}; do
        echo -e "${CYAN}  Restarting: ${DEPLOY}${NC}"
        kubectl rollout restart deployment/${DEPLOY} -n ${NAMESPACE}
    done
    
    echo -e "\n${YELLOW}Waiting for rollout to complete...${NC}"
    for DEPLOY in ${DEPLOYMENTS}; do
        kubectl rollout status deployment/${DEPLOY} -n ${NAMESPACE} --timeout=300s
    done
    
    echo -e "${GREEN}✅ All deployments restarted${NC}"
fi

# ==============================================================================
# STEP 5: Verify Sidecar Injection
# ==============================================================================
echo -e "\n${BLUE}[5/5] Verifying sidecar injection...${NC}"

echo -e "\n${YELLOW}Pods with sidecars:${NC}"
kubectl get pods -n ${NAMESPACE} -o wide

echo -e "\n${YELLOW}Container count per pod:${NC}"
kubectl get pods -n ${NAMESPACE} -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].name}{"\n"}{end}' | column -t

# Check if pods have istio-proxy
PODS_WITH_SIDECAR=$(kubectl get pods -n ${NAMESPACE} -o jsonpath='{.items[?(@.spec.containers[*].name=="istio-proxy")].metadata.name}')

if [ -z "$PODS_WITH_SIDECAR" ]; then
    echo -e "\n${RED}❌ No pods with Istio sidecar found${NC}"
    echo -e "${YELLOW}Pods may still be starting. Check with:${NC}"
    echo -e "   kubectl get pods -n ${NAMESPACE}"
else
    echo -e "\n${GREEN}✅ Pods with Istio sidecar:${NC}"
    echo "${PODS_WITH_SIDECAR}" | tr ' ' '\n' | sed 's/^/  ✓ /'
fi

# ==============================================================================
# Summary
# ==============================================================================
echo -e "\n${CYAN}"
echo "╔═══════════════════════════════════════════════════════════════════╗"
echo "║                  SIDECAR INJECTION ENABLED! 🎉                    ║"
echo "╚═══════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

echo -e "${GREEN}✅ Namespace: ${NAMESPACE}${NC}"
echo -e "${GREEN}✅ Label: istio-injection=enabled${NC}"
echo -e "${GREEN}✅ Pods restarted${NC}"

echo -e "\n${YELLOW}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}WHAT HAPPENED:${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════════${NC}"

echo -e "\n${BLUE}1. Namespace labeled for injection${NC}"
echo -e "   All new pods will automatically get Envoy sidecar"

echo -e "\n${BLUE}2. Existing pods restarted${NC}"
echo -e "   Each pod now has 2 containers:"
echo -e "   • Your application container"
echo -e "   • istio-proxy (Envoy) container"

echo -e "\n${BLUE}3. Envoy proxy intercepts all traffic${NC}"
echo -e "   • Inbound traffic (to your service)"
echo -e "   • Outbound traffic (from your service)"

echo -e "\n${YELLOW}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}VERIFY COMMANDS:${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════════${NC}"

echo -e "\n${CYAN}Check pods:${NC}"
echo -e "   kubectl get pods -n ${NAMESPACE}"

echo -e "\n${CYAN}Check container count (should be 2+):${NC}"
echo -e "   kubectl get pods -n ${NAMESPACE} -o jsonpath='{.items[0].spec.containers[*].name}'"

echo -e "\n${CYAN}View Envoy proxy logs:${NC}"
echo -e "   kubectl logs <pod-name> -c istio-proxy -n ${NAMESPACE}"

echo -e "\n${CYAN}Check proxy status:${NC}"
echo -e "   istioctl proxy-status"

echo -e "\n${CYAN}Analyze configuration:${NC}"
echo -e "   istioctl analyze -n ${NAMESPACE}"

echo -e "\n${YELLOW}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}DISABLE INJECTION (if needed):${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════════${NC}"

echo -e "\n${RED}Remove label:${NC}"
echo -e "   kubectl label namespace ${NAMESPACE} istio-injection-"

echo -e "\n${RED}Restart pods:${NC}"
echo -e "   kubectl rollout restart deployment -n ${NAMESPACE}"

echo ""

