#!/bin/bash

# ==============================================================================
# CONFIGURE MTLS - Interactive Setup
# ==============================================================================
# Purpose: Configure mTLS for the service mesh
# Usage: ./configure-mtls.sh [mode]
# Modes: permissive, strict, disable
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

MODE=${1:-strict}
NAMESPACE="observability"

echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                  CONFIGURE mTLS MODE                               ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# ==============================================================================
# Mode Selection
# ==============================================================================
echo -e "${YELLOW}Select mTLS mode:${NC}"
echo ""
echo -e "${BLUE}1. PERMISSIVE (Migration mode)${NC}"
echo "   - Accepts both mTLS and plain text"
echo "   - Use during rollout when not all pods have sidecars"
echo "   - Safe for initial deployment"
echo ""
echo -e "${GREEN}2. STRICT (Production mode) ⭐ RECOMMENDED${NC}"
echo "   - Only mTLS accepted"
echo "   - Plain text connections rejected"
echo "   - Maximum security"
echo ""
echo -e "${RED}3. DISABLE (No encryption)${NC}"
echo "   - No mTLS enforcement"
echo "   - Plain text allowed"
echo "   - NOT recommended for production"
echo ""

if [ "$MODE" != "permissive" ] && [ "$MODE" != "strict" ] && [ "$MODE" != "disable" ]; then
    read -p "Enter mode (permissive/strict/disable): " MODE
fi

MODE_UPPER=$(echo "$MODE" | tr '[:lower:]' '[:upper:]')

echo -e "\n${YELLOW}Selected mode: ${MODE_UPPER}${NC}"
echo ""

# ==============================================================================
# Apply PeerAuthentication
# ==============================================================================
echo -e "${BLUE}[1/4] Applying PeerAuthentication...${NC}"

cat <<EOF | kubectl apply -f -
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default-mtls
  namespace: ${NAMESPACE}
  labels:
    app: mtls-policy
spec:
  mtls:
    mode: ${MODE_UPPER}
EOF

echo -e "${GREEN}✅ PeerAuthentication applied (mode: ${MODE_UPPER})${NC}"

# ==============================================================================
# Apply AuthorizationPolicies
# ==============================================================================
echo -e "\n${BLUE}[2/4] Applying AuthorizationPolicies...${NC}"

kubectl apply -f ../full-observability/templates/istio/15-authorization-policy.yaml

echo -e "${GREEN}✅ AuthorizationPolicies applied${NC}"

# ==============================================================================
# Verify Configuration
# ==============================================================================
echo -e "\n${BLUE}[3/4] Verifying mTLS configuration...${NC}"

# Get all PeerAuthentication
echo -e "\n${YELLOW}PeerAuthentication policies:${NC}"
kubectl get peerauthentication -n ${NAMESPACE}

# Get all AuthorizationPolicy
echo -e "\n${YELLOW}Authorization policies:${NC}"
kubectl get authorizationpolicy -n ${NAMESPACE}

# ==============================================================================
# Test mTLS
# ==============================================================================
echo -e "\n${BLUE}[4/4] Testing mTLS...${NC}"

# Get a pod to test
TEST_POD=$(kubectl get pod -n ${NAMESPACE} -l app=user-service -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

if [ ! -z "$TEST_POD" ]; then
    echo -e "\n${CYAN}Testing mTLS from pod: ${TEST_POD}${NC}"
    
    if command -v istioctl &> /dev/null; then
        echo -e "\n${YELLOW}TLS Check:${NC}"
        istioctl authn tls-check ${TEST_POD}.${NAMESPACE} -n ${NAMESPACE} || true
    fi
fi

# ==============================================================================
# Summary
# ==============================================================================
echo -e "\n${CYAN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                  mTLS CONFIGURATION COMPLETE! 🔒                   ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

echo -e "${GREEN}✅ mTLS Mode: ${MODE_UPPER}${NC}"
echo -e "${GREEN}✅ Namespace: ${NAMESPACE}${NC}"

if [ "$MODE" == "strict" ]; then
    echo -e "\n${MAGENTA}═══════════════════════════════════════════════════════════════════${NC}"
    echo -e "${MAGENTA}STRICT MODE ACTIVE:${NC}"
    echo -e "${MAGENTA}═══════════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}✅ All service-to-service traffic is encrypted${NC}"
    echo -e "${GREEN}✅ Plain text connections will be rejected${NC}"
    echo -e "${GREEN}✅ Certificates automatically rotated every 24h${NC}"
    echo ""
    echo -e "${YELLOW}IMPORTANT:${NC}"
    echo -e "  - All pods MUST have Istio sidecar"
    echo -e "  - External clients must use Ingress Gateway"
    echo -e "  - Direct pod access will fail"
    
elif [ "$MODE" == "permissive" ]; then
    echo -e "\n${YELLOW}═══════════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}PERMISSIVE MODE ACTIVE:${NC}"
    echo -e "${YELLOW}═══════════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}⚠️  Both mTLS and plain text accepted${NC}"
    echo -e "${CYAN}⚠️  This is for MIGRATION only${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo -e "  1. Verify all pods have sidecars"
    echo -e "  2. Monitor traffic (should all be mTLS)"
    echo -e "  3. Switch to STRICT mode:"
    echo -e "     ./configure-mtls.sh strict"

elif [ "$MODE" == "disable" ]; then
    echo -e "\n${RED}═══════════════════════════════════════════════════════════════════${NC}"
    echo -e "${RED}mTLS DISABLED:${NC}"
    echo -e "${RED}═══════════════════════════════════════════════════════════════════${NC}"
    echo -e "${RED}⚠️  No encryption enforced${NC}"
    echo -e "${RED}⚠️  NOT recommended for production${NC}"
fi

echo -e "\n${CYAN}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}VERIFY COMMANDS:${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}"

echo -e "\n${BLUE}Check mTLS status for all services:${NC}"
echo -e "  istioctl authn tls-check -n ${NAMESPACE}"

echo -e "\n${BLUE}View certificates:${NC}"
echo -e "  istioctl proxy-config secret <pod-name> -n ${NAMESPACE}"

echo -e "\n${BLUE}Test connection:${NC}"
echo -e "  kubectl exec -it <pod> -c istio-proxy -n ${NAMESPACE} -- curl http://user-service:8080/health"

echo -e "\n${BLUE}View authorization policies:${NC}"
echo -e "  kubectl get authorizationpolicy -n ${NAMESPACE}"

echo -e "\n${BLUE}Test authorization (should fail if not allowed):${NC}"
echo -e "  kubectl run test --rm -it --image=curlimages/curl -- curl http://user-service:8080"
echo ""

