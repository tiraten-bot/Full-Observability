#!/bin/bash

# ==============================================================================
# CREATE SECRETS SCRIPT
# ==============================================================================
# Purpose: Apply all secret configurations to Kubernetes cluster
# Usage: ./create-secrets.sh
# ==============================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Creating Kubernetes Secrets${NC}"
echo -e "${GREEN}========================================${NC}"

# ==============================================================================
# Check if namespace exists
# ==============================================================================
echo -e "\n${YELLOW}[1/6] Checking namespace...${NC}"
if ! kubectl get namespace observability &> /dev/null; then
    echo -e "${RED}❌ Namespace 'observability' does not exist${NC}"
    echo "Please create the namespace first:"
    echo "  cd ../00-namespace && kubectl apply -f namespace.yaml"
    exit 1
fi
echo -e "${GREEN}✅ Namespace 'observability' exists${NC}"

# ==============================================================================
# Create Database Secrets
# ==============================================================================
echo -e "\n${YELLOW}[2/6] Creating database secrets...${NC}"
kubectl apply -f database-secret.yaml
echo -e "${GREEN}✅ Database secrets created${NC}"

# ==============================================================================
# Create JWT Secrets
# ==============================================================================
echo -e "\n${YELLOW}[3/6] Creating JWT and API secrets...${NC}"
kubectl apply -f jwt-secret.yaml
echo -e "${GREEN}✅ JWT secrets created${NC}"

# ==============================================================================
# Create Kafka Secrets
# ==============================================================================
echo -e "\n${YELLOW}[4/6] Creating Kafka secrets...${NC}"
kubectl apply -f kafka-secret.yaml
echo -e "${GREEN}✅ Kafka secrets created${NC}"

# ==============================================================================
# Create Observability Secrets
# ==============================================================================
echo -e "\n${YELLOW}[5/6] Creating observability secrets...${NC}"
kubectl apply -f observability-secret.yaml
echo -e "${GREEN}✅ Observability secrets created${NC}"

# ==============================================================================
# Verify Secrets
# ==============================================================================
echo -e "\n${YELLOW}[6/6] Verifying secrets...${NC}"
echo ""
kubectl get secrets -n observability

# ==============================================================================
# Display Secret Details
# ==============================================================================
echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Secret Details:${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

echo -e "\n${YELLOW}Database Secrets:${NC}"
kubectl get secret postgres-secret -n observability -o jsonpath='{.data}' | jq 'keys'

echo -e "\n${YELLOW}JWT Secrets:${NC}"
kubectl get secret jwt-secret -n observability -o jsonpath='{.data}' | jq 'keys'

echo -e "\n${YELLOW}Kafka Secrets:${NC}"
kubectl get secret kafka-secret -n observability -o jsonpath='{.data}' | jq 'keys'

echo -e "\n${YELLOW}Grafana Secrets:${NC}"
kubectl get secret grafana-secret -n observability -o jsonpath='{.data}' | jq 'keys'

# ==============================================================================
# Security Warning
# ==============================================================================
echo -e "\n${RED}⚠️  SECURITY WARNING:${NC}"
echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "The secrets use default/example values."
echo -e "For PRODUCTION environments:"
echo -e "  1. Generate strong, random passwords"
echo -e "  2. Use a secret management system (Vault, AWS Secrets Manager, etc.)"
echo -e "  3. Enable encryption at rest"
echo -e "  4. Rotate secrets regularly"
echo -e "  5. Never commit secrets to version control"
echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# ==============================================================================
# Completion
# ==============================================================================
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✅ All secrets created successfully!${NC}"
echo -e "${GREEN}========================================${NC}"

echo -e "\n${BLUE}Useful commands:${NC}"
echo -e "  ${YELLOW}kubectl get secrets -n observability${NC}"
echo -e "  ${YELLOW}kubectl describe secret postgres-secret -n observability${NC}"
echo -e "  ${YELLOW}kubectl get secret postgres-secret -n observability -o yaml${NC}"
echo -e "  ${YELLOW}kubectl delete secret <secret-name> -n observability${NC}"
echo ""

