#!/bin/bash

# ==============================================================================
# KUBECTL CONTEXT SETUP SCRIPT
# ==============================================================================
# Purpose: Configure kubectl to work with the observability namespace
# Usage: ./setup-context.sh
# ==============================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Kubectl Context Setup${NC}"
echo -e "${GREEN}========================================${NC}"

# ==============================================================================
# STEP 1: Check kubectl installation
# ==============================================================================
echo -e "\n${YELLOW}[1/7] Checking kubectl installation...${NC}"
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl is not installed${NC}"
    echo "Please install kubectl: https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi
echo -e "${GREEN}✅ kubectl is installed: $(kubectl version --client --short 2>/dev/null || kubectl version --client)${NC}"

# ==============================================================================
# STEP 2: Check cluster connection
# ==============================================================================
echo -e "\n${YELLOW}[2/7] Checking cluster connection...${NC}"
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ Cannot connect to Kubernetes cluster${NC}"
    echo "Please ensure:"
    echo "  - Minikube is running: minikube start"
    echo "  - Or cloud cluster is configured"
    echo "  - kubeconfig is properly set"
    exit 1
fi
echo -e "${GREEN}✅ Connected to cluster${NC}"
kubectl cluster-info | head -n 2

# ==============================================================================
# STEP 3: Display current context
# ==============================================================================
echo -e "\n${YELLOW}[3/7] Current context:${NC}"
CURRENT_CONTEXT=$(kubectl config current-context)
echo "  Context: ${CURRENT_CONTEXT}"
CURRENT_NAMESPACE=$(kubectl config view --minify --output 'jsonpath={..namespace}')
echo "  Namespace: ${CURRENT_NAMESPACE:-default}"

# ==============================================================================
# STEP 4: Create namespace if not exists
# ==============================================================================
echo -e "\n${YELLOW}[4/7] Creating observability namespace...${NC}"
if kubectl get namespace observability &> /dev/null; then
    echo -e "${GREEN}✅ Namespace 'observability' already exists${NC}"
else
    kubectl apply -f namespace.yaml
    echo -e "${GREEN}✅ Namespace 'observability' created${NC}"
fi

# ==============================================================================
# STEP 5: Set default namespace for current context
# ==============================================================================
echo -e "\n${YELLOW}[5/7] Setting default namespace to 'observability'...${NC}"
kubectl config set-context --current --namespace=observability
echo -e "${GREEN}✅ Default namespace set to 'observability'${NC}"

# ==============================================================================
# STEP 6: Verify namespace switch
# ==============================================================================
echo -e "\n${YELLOW}[6/7] Verifying namespace switch...${NC}"
CURRENT_NS=$(kubectl config view --minify --output 'jsonpath={..namespace}')
if [ "$CURRENT_NS" == "observability" ]; then
    echo -e "${GREEN}✅ Successfully switched to namespace: ${CURRENT_NS}${NC}"
else
    echo -e "${RED}❌ Failed to switch namespace${NC}"
    exit 1
fi

# ==============================================================================
# STEP 7: Display context information
# ==============================================================================
echo -e "\n${YELLOW}[7/7] Context Information:${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
kubectl config get-contexts $(kubectl config current-context)
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ==============================================================================
# STEP 8: Verify RBAC permissions (optional)
# ==============================================================================
echo -e "\n${YELLOW}Checking RBAC permissions...${NC}"
echo "Can create deployments: $(kubectl auth can-i create deployments -n observability)"
echo "Can create services: $(kubectl auth can-i create services -n observability)"
echo "Can create pods: $(kubectl auth can-i create pods -n observability)"
echo "Can create secrets: $(kubectl auth can-i create secrets -n observability)"
echo "Can create configmaps: $(kubectl auth can-i create configmaps -n observability)"

# ==============================================================================
# COMPLETION MESSAGE
# ==============================================================================
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✅ Context setup completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "\nYou are now working in the '${GREEN}observability${NC}' namespace."
echo -e "\nUseful commands:"
echo -e "  ${YELLOW}kubectl get all${NC}                    - List all resources"
echo -e "  ${YELLOW}kubectl get pods${NC}                   - List all pods"
echo -e "  ${YELLOW}kubectl get svc${NC}                    - List all services"
echo -e "  ${YELLOW}kubectl describe pod <name>${NC}        - Describe a pod"
echo -e "  ${YELLOW}kubectl logs <pod-name>${NC}            - View pod logs"
echo -e "  ${YELLOW}kubectl config get-contexts${NC}        - List all contexts"
echo -e "  ${YELLOW}kubectl config use-context <name>${NC}  - Switch context"
echo ""

