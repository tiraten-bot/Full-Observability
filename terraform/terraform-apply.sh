#!/bin/bash

# ==============================================================================
# TERRAFORM APPLY SCRIPT
# ==============================================================================
# Purpose: Initialize and apply Terraform configuration
# Usage: ./terraform-apply.sh [environment]
# Environments: dev, staging, prod
# ==============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

ENV=${1:-dev}

if [ "$ENV" != "dev" ] && [ "$ENV" != "staging" ] && [ "$ENV" != "prod" ]; then
    echo -e "${RED}Invalid environment: ${ENV}${NC}"
    echo "Valid: dev, staging, prod"
    exit 1
fi

echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║              TERRAFORM AWS INFRASTRUCTURE                          ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo -e "${YELLOW}Environment: ${ENV}${NC}\n"

# ==============================================================================
# 1. Check AWS Credentials
# ==============================================================================
echo -e "${BLUE}[1/6] Checking AWS credentials...${NC}"

if ! aws sts get-caller-identity &>/dev/null; then
    echo -e "${RED}❌ AWS credentials not configured${NC}"
    echo "Run: aws configure"
    exit 1
fi

ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
echo -e "${GREEN}✅ AWS Account: ${ACCOUNT_ID}${NC}\n"

# ==============================================================================
# 2. Terraform Init
# ==============================================================================
echo -e "${BLUE}[2/6] Initializing Terraform...${NC}"

terraform init \
    -backend-config="key=${ENV}/terraform.tfstate" \
    -reconfigure

echo -e "${GREEN}✅ Terraform initialized${NC}\n"

# ==============================================================================
# 3. Terraform Validate
# ==============================================================================
echo -e "${BLUE}[3/6] Validating configuration...${NC}"

terraform validate

echo -e "${GREEN}✅ Configuration valid${NC}\n"

# ==============================================================================
# 4. Terraform Plan
# ==============================================================================
echo -e "${BLUE}[4/6] Creating execution plan...${NC}"

terraform plan \
    -var-file="environments/${ENV}/terraform.tfvars" \
    -out="${ENV}.tfplan"

echo -e "${GREEN}✅ Plan created: ${ENV}.tfplan${NC}\n"

# ==============================================================================
# 5. Review and Confirm
# ==============================================================================
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}REVIEW THE PLAN ABOVE${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════════${NC}\n"

if [ "$ENV" == "prod" ]; then
    echo -e "${RED}⚠️  PRODUCTION ENVIRONMENT${NC}"
    echo -e "${RED}This will create REAL AWS resources and incur COSTS${NC}\n"
fi

read -p "Apply this plan? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    echo -e "${YELLOW}Cancelled${NC}"
    exit 0
fi

# ==============================================================================
# 6. Terraform Apply
# ==============================================================================
echo -e "\n${BLUE}[6/6] Applying infrastructure changes...${NC}\n"

terraform apply "${ENV}.tfplan"

echo -e "\n${GREEN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║          INFRASTRUCTURE CREATED SUCCESSFULLY! 🎉                   ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}\n"

# ==============================================================================
# Display Outputs
# ==============================================================================
echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}INFRASTRUCTURE DETAILS:${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════════════${NC}\n"

terraform output -json | jq '.'

echo -e "\n${YELLOW}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}NEXT STEPS:${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════════${NC}\n"

echo -e "${GREEN}1. Configure kubectl:${NC}"
terraform output -raw configure_kubectl
echo ""

echo -e "\n${GREEN}2. Verify cluster:${NC}"
echo "   kubectl get nodes"

echo -e "\n${GREEN}3. Deploy application:${NC}"
echo "   cd ../helm && ./install.sh ${ENV}"

echo -e "\n${GREEN}4. Access services:${NC}"
echo "   API: $(terraform output -raw api_gateway_url)"
echo "   Grafana: $(terraform output -raw grafana_url)"
echo "   Jaeger: $(terraform output -raw jaeger_url)"

echo ""

