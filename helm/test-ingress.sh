#!/bin/bash

# ==============================================================================
# INGRESS TESTING SCRIPT
# ==============================================================================
# Purpose: Test ingress endpoints after deployment
# Usage: ./test-ingress.sh [hostname]
# ==============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get ingress hostname
INGRESS_IP=${1:-$(kubectl get svc ingress-nginx-controller -n ingress-nginx -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)}

if [ -z "$INGRESS_IP" ]; then
    INGRESS_IP=$(kubectl get svc ingress-nginx-controller -n ingress-nginx -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null)
fi

if [ -z "$INGRESS_IP" ]; then
    echo -e "${RED}❌ Could not get ingress controller IP${NC}"
    echo "Run: kubectl get svc -n ingress-nginx"
    exit 1
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Ingress Testing${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "${YELLOW}Ingress IP: ${INGRESS_IP}${NC}"
echo ""

# ==============================================================================
# Test API Gateway
# ==============================================================================
echo -e "${BLUE}[1/5] Testing API Gateway...${NC}"

API_HOST="api.example.com"
echo -e "Testing: http://${INGRESS_IP} (Host: ${API_HOST})"

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: ${API_HOST}" http://${INGRESS_IP}/health 2>/dev/null || echo "000")

if [ "$RESPONSE" = "200" ]; then
    echo -e "${GREEN}✅ API Gateway is responding (HTTP ${RESPONSE})${NC}"
else
    echo -e "${RED}❌ API Gateway failed (HTTP ${RESPONSE})${NC}"
fi

# ==============================================================================
# Test Grafana
# ==============================================================================
echo -e "\n${BLUE}[2/5] Testing Grafana...${NC}"

GRAFANA_HOST="grafana.example.com"
echo -e "Testing: http://${INGRESS_IP} (Host: ${GRAFANA_HOST})"

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: ${GRAFANA_HOST}" http://${INGRESS_IP}/api/health 2>/dev/null || echo "000")

if [ "$RESPONSE" = "200" ] || [ "$RESPONSE" = "401" ]; then
    echo -e "${GREEN}✅ Grafana is responding (HTTP ${RESPONSE})${NC}"
else
    echo -e "${RED}❌ Grafana failed (HTTP ${RESPONSE})${NC}"
fi

# ==============================================================================
# Test Jaeger
# ==============================================================================
echo -e "\n${BLUE}[3/5] Testing Jaeger...${NC}"

JAEGER_HOST="jaeger.example.com"
echo -e "Testing: http://${INGRESS_IP} (Host: ${JAEGER_HOST})"

RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -H "Host: ${JAEGER_HOST}" http://${INGRESS_IP}/ 2>/dev/null || echo "000")

if [ "$RESPONSE" = "200" ] || [ "$RESPONSE" = "401" ]; then
    echo -e "${GREEN}✅ Jaeger is responding (HTTP ${RESPONSE})${NC}"
else
    echo -e "${RED}❌ Jaeger failed (HTTP ${RESPONSE})${NC}"
fi

# ==============================================================================
# Test SSL/TLS (if configured)
# ==============================================================================
echo -e "\n${BLUE}[4/5] Testing SSL/TLS...${NC}"

if curl -s -k https://${INGRESS_IP} -H "Host: ${API_HOST}" &> /dev/null; then
    echo -e "${GREEN}✅ SSL/TLS is configured${NC}"
    
    # Check certificate
    CERT_INFO=$(echo | openssl s_client -servername ${API_HOST} -connect ${INGRESS_IP}:443 2>/dev/null | openssl x509 -noout -subject -dates 2>/dev/null || echo "")
    if [ ! -z "$CERT_INFO" ]; then
        echo -e "${YELLOW}Certificate Info:${NC}"
        echo "$CERT_INFO" | sed 's/^/  /'
    fi
else
    echo -e "${YELLOW}⚠️  SSL/TLS not configured or not responding${NC}"
fi

# ==============================================================================
# Test Ingress Rules
# ==============================================================================
echo -e "\n${BLUE}[5/5] Checking Ingress Rules...${NC}"

echo -e "${YELLOW}Ingress Resources:${NC}"
kubectl get ingress -n observability

echo -e "\n${YELLOW}Ingress Details:${NC}"
kubectl describe ingress -n observability | grep -E "Host|Path|Backend" | sed 's/^/  /'

# ==============================================================================
# Summary
# ==============================================================================
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}Test Summary${NC}"
echo -e "${GREEN}========================================${NC}"

echo -e "\n${YELLOW}Access URLs (update /etc/hosts or DNS):${NC}"
echo -e "  ${INGRESS_IP}  ${API_HOST}"
echo -e "  ${INGRESS_IP}  ${GRAFANA_HOST}"
echo -e "  ${INGRESS_IP}  ${JAEGER_HOST}"

echo -e "\n${YELLOW}Test Commands:${NC}"
echo -e "  ${BLUE}curl -H 'Host: ${API_HOST}' http://${INGRESS_IP}/health${NC}"
echo -e "  ${BLUE}curl -H 'Host: ${GRAFANA_HOST}' http://${INGRESS_IP}/api/health${NC}"
echo -e "  ${BLUE}curl -H 'Host: ${JAEGER_HOST}' http://${INGRESS_IP}/${NC}"

echo -e "\n${YELLOW}Browser Access (add to /etc/hosts):${NC}"
echo -e "  http://${API_HOST}"
echo -e "  http://${GRAFANA_HOST}"
echo -e "  http://${JAEGER_HOST}"

echo -e "\n${BLUE}For local testing, add to /etc/hosts:${NC}"
echo -e "  ${INGRESS_IP} ${API_HOST} ${GRAFANA_HOST} ${JAEGER_HOST}"
echo ""

