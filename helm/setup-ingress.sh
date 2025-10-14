#!/bin/bash

# ==============================================================================
# INGRESS CONTROLLER SETUP
# ==============================================================================
# Purpose: Install and configure NGINX Ingress Controller + Cert-Manager
# Usage: ./setup-ingress.sh
# ==============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Ingress Controller Setup${NC}"
echo -e "${GREEN}========================================${NC}"

# ==============================================================================
# 1. Install NGINX Ingress Controller
# ==============================================================================
echo -e "\n${BLUE}[1/4] Installing NGINX Ingress Controller...${NC}"

if kubectl get namespace ingress-nginx &> /dev/null; then
    echo -e "${YELLOW}⚠️  Ingress controller namespace already exists${NC}"
else
    echo -e "${YELLOW}Installing NGINX Ingress Controller...${NC}"
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml
    
    echo -e "${YELLOW}Waiting for ingress controller to be ready...${NC}"
    kubectl wait --namespace ingress-nginx \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/component=controller \
      --timeout=120s
    
    echo -e "${GREEN}✅ NGINX Ingress Controller installed${NC}"
fi

# ==============================================================================
# 2. Install Cert-Manager (for TLS certificates)
# ==============================================================================
echo -e "\n${BLUE}[2/4] Installing Cert-Manager...${NC}"

if kubectl get namespace cert-manager &> /dev/null; then
    echo -e "${YELLOW}⚠️  Cert-Manager namespace already exists${NC}"
else
    echo -e "${YELLOW}Installing Cert-Manager...${NC}"
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.2/cert-manager.yaml
    
    echo -e "${YELLOW}Waiting for cert-manager to be ready...${NC}"
    kubectl wait --namespace cert-manager \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/instance=cert-manager \
      --timeout=120s
    
    echo -e "${GREEN}✅ Cert-Manager installed${NC}"
fi

# ==============================================================================
# 3. Create ClusterIssuer for Let's Encrypt
# ==============================================================================
echo -e "\n${BLUE}[3/4] Creating Let's Encrypt ClusterIssuers...${NC}"

# Staging ClusterIssuer
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: admin@example.com  # CHANGE THIS!
    privateKeySecretRef:
      name: letsencrypt-staging
    solvers:
      - http01:
          ingress:
            class: nginx
EOF

# Production ClusterIssuer
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com  # CHANGE THIS!
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: nginx
EOF

echo -e "${GREEN}✅ ClusterIssuers created${NC}"

# ==============================================================================
# 4. Verify Installation
# ==============================================================================
echo -e "\n${BLUE}[4/4] Verifying installation...${NC}"

echo -e "\n${YELLOW}NGINX Ingress Controller:${NC}"
kubectl get pods -n ingress-nginx

echo -e "\n${YELLOW}Cert-Manager:${NC}"
kubectl get pods -n cert-manager

echo -e "\n${YELLOW}ClusterIssuers:${NC}"
kubectl get clusterissuers

# ==============================================================================
# Get Ingress Controller External IP
# ==============================================================================
echo -e "\n${BLUE}Getting Ingress Controller External IP...${NC}"
EXTERNAL_IP=""
while [ -z "$EXTERNAL_IP" ]; do
    echo "Waiting for external IP..."
    EXTERNAL_IP=$(kubectl get svc ingress-nginx-controller -n ingress-nginx -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
    if [ -z "$EXTERNAL_IP" ]; then
        EXTERNAL_IP=$(kubectl get svc ingress-nginx-controller -n ingress-nginx -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null)
    fi
    [ -z "$EXTERNAL_IP" ] && sleep 5
done

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✅ Ingress Setup Completed!${NC}"
echo -e "${GREEN}========================================${NC}"

echo -e "\n${YELLOW}Ingress Controller External IP/Hostname:${NC}"
echo -e "  ${EXTERNAL_IP}"

echo -e "\n${YELLOW}Next Steps:${NC}"
echo -e "  1. Update DNS records to point to: ${EXTERNAL_IP}"
echo -e "  2. Update email in ClusterIssuers (admin@example.com)"
echo -e "  3. Configure ingress hosts in values.yaml"
echo -e "  4. Deploy application with ingress enabled"

echo -e "\n${BLUE}DNS Configuration Example:${NC}"
echo -e "  api.example.com        A    ${EXTERNAL_IP}"
echo -e "  grafana.example.com    A    ${EXTERNAL_IP}"
echo -e "  jaeger.example.com     A    ${EXTERNAL_IP}"

echo -e "\n${BLUE}Test Ingress (after DNS setup):${NC}"
echo -e "  curl https://api.example.com/health"
echo -e "  curl https://grafana.example.com"
echo ""

