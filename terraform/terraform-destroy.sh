#!/bin/bash

# ==============================================================================
# TERRAFORM DESTROY SCRIPT
# ==============================================================================
# Purpose: Destroy all AWS infrastructure
# Usage: ./terraform-destroy.sh [environment]
# WARNING: This will delete ALL resources!
# ==============================================================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ENV=${1:-dev}

echo -e "${RED}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                  DESTROY AWS INFRASTRUCTURE                        ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo -e "${YELLOW}Environment: ${ENV}${NC}\n"

echo -e "${RED}⚠️  WARNING: This will DESTROY all AWS resources!${NC}"
echo -e "${RED}⚠️  This action CANNOT be undone!${NC}\n"

if [ "$ENV" == "prod" ]; then
    echo -e "${RED}╔════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║              PRODUCTION ENVIRONMENT DESTRUCTION                    ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════════════════╝${NC}\n"
    
    read -p "Type 'destroy-production' to confirm: " CONFIRM
    
    if [ "$CONFIRM" != "destroy-production" ]; then
        echo -e "${YELLOW}Cancelled${NC}"
        exit 0
    fi
else
    read -p "Type 'destroy' to confirm: " CONFIRM
    
    if [ "$CONFIRM" != "destroy" ]; then
        echo -e "${YELLOW}Cancelled${NC}"
        exit 0
    fi
fi

echo -e "\n${YELLOW}Creating destruction plan...${NC}\n"

terraform plan \
    -destroy \
    -var-file="environments/${ENV}/terraform.tfvars" \
    -out="${ENV}-destroy.tfplan"

echo -e "\n${RED}Final confirmation required!${NC}"
read -p "Proceed with destruction? (yes/no): " FINAL_CONFIRM

if [ "$FINAL_CONFIRM" != "yes" ]; then
    echo -e "${YELLOW}Cancelled${NC}"
    rm -f "${ENV}-destroy.tfplan"
    exit 0
fi

terraform apply "${ENV}-destroy.tfplan"

echo -e "\n${GREEN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║          ALL RESOURCES DESTROYED                                   ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}\n"

