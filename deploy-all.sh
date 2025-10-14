#!/bin/bash

# ==============================================================================
# DEPLOY EVERYTHING - ONE COMMAND DEPLOYMENT
# ==============================================================================
# Purpose: Deploy entire Full Observability stack with one command
# Usage: ./deploy-all.sh [environment]
# Environment: dev, prod (default: dev)
# ==============================================================================

set -e

ENV=${1:-dev}
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║           FULL OBSERVABILITY - COMPLETE DEPLOYMENT                ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo -e "${YELLOW}Environment: ${ENV}${NC}\n"

# ==============================================================================
# 1. Build Docker Images
# ==============================================================================
echo -e "${BLUE}[1/6] Building Docker images...${NC}"
docker-compose build
echo -e "${GREEN}✅ Docker images built${NC}\n"

# ==============================================================================
# 2. Install Istio (optional)
# ==============================================================================
read -p "Install Istio Service Mesh? (y/n): " INSTALL_ISTIO
if [ "$INSTALL_ISTIO" == "y" ]; then
    echo -e "${BLUE}[2/6] Installing Istio...${NC}"
    cd helm/istio
    ./install-istio.sh
    ./enable-injection.sh
    cd ../..
    echo -e "${GREEN}✅ Istio installed${NC}\n"
else
    echo -e "${YELLOW}⊘ Skipping Istio installation${NC}\n"
fi

# ==============================================================================
# 3. Install Ingress Controller
# ==============================================================================
read -p "Install NGINX Ingress Controller? (y/n): " INSTALL_INGRESS
if [ "$INSTALL_INGRESS" == "y" ]; then
    echo -e "${BLUE}[3/6] Installing Ingress Controller...${NC}"
    cd helm
    ./setup-ingress.sh
    cd ..
    echo -e "${GREEN}✅ Ingress Controller installed${NC}\n"
else
    echo -e "${YELLOW}⊘ Skipping Ingress installation${NC}\n"
fi

# ==============================================================================
# 4. Deploy with Helm
# ==============================================================================
echo -e "${BLUE}[4/6] Deploying with Helm...${NC}"
cd helm

if [ "$ENV" == "prod" ]; then
    helm install full-observability ./full-observability \
        -f values-prod.yaml \
        -n observability \
        --create-namespace \
        --wait \
        --timeout 15m
else
    helm install full-observability ./full-observability \
        -f values-dev.yaml \
        -n observability \
        --create-namespace \
        --wait \
        --timeout 15m
fi

cd ..
echo -e "${GREEN}✅ Helm deployment complete${NC}\n"

# ==============================================================================
# 5. Verify Deployment
# ==============================================================================
echo -e "${BLUE}[5/6] Verifying deployment...${NC}"
kubectl get all -n observability
echo -e "${GREEN}✅ Verification complete${NC}\n"

# ==============================================================================
# 6. Display Access Information
# ==============================================================================
echo -e "${BLUE}[6/6] Getting access information...${NC}\n"

# Get service URLs
echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}ACCESS INFORMATION:${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}\n"

# API Gateway
GATEWAY_PORT=$(kubectl get svc api-gateway -n observability -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "8000")
echo -e "${GREEN}API Gateway:${NC}"
echo -e "  kubectl port-forward -n observability svc/api-gateway 8000:8000"
echo -e "  Then: http://localhost:8000\n"

# Grafana
GRAFANA_PORT=$(kubectl get svc grafana -n observability -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "3000")
echo -e "${GREEN}Grafana:${NC}"
echo -e "  kubectl port-forward -n observability svc/grafana 3000:3000"
echo -e "  Then: http://localhost:3000"
echo -e "  Login: admin/admin\n"

# Prometheus
echo -e "${GREEN}Prometheus:${NC}"
echo -e "  kubectl port-forward -n observability svc/prometheus 9090:9090"
echo -e "  Then: http://localhost:9090\n"

# Jaeger
echo -e "${GREEN}Jaeger:${NC}"
echo -e "  kubectl port-forward -n observability svc/jaeger 16686:16686"
echo -e "  Then: http://localhost:16686\n"

# ==============================================================================
# Success
# ==============================================================================
echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║              DEPLOYMENT COMPLETED SUCCESSFULLY! 🎉                 ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}\n"

echo -e "${YELLOW}Quick commands:${NC}"
echo -e "  ${BLUE}kubectl get pods -n observability${NC}          # View all pods"
echo -e "  ${BLUE}kubectl logs -f -l app=user-service -n observability${NC}  # View logs"
echo -e "  ${BLUE}helm list -n observability${NC}                 # View Helm releases"
echo -e "  ${BLUE}helm uninstall full-observability -n observability${NC}  # Uninstall"
echo ""

