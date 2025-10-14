#!/bin/bash

# ==============================================================================
# TEST ISTIO SERVICE MESH
# ==============================================================================
# Purpose: Verify Istio installation and functionality
# Usage: ./test-istio.sh
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

NAMESPACE="observability"

echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                    ISTIO SERVICE MESH TEST                         ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# ==============================================================================
# TEST 1: Istio Installation
# ==============================================================================
echo -e "${BLUE}[1/10] Testing Istio installation...${NC}"

if ! kubectl get namespace istio-system &> /dev/null; then
    echo -e "${RED}❌ istio-system namespace not found${NC}"
    exit 1
fi
echo -e "${GREEN}✅ istio-system namespace exists${NC}"

# Check Istiod
if kubectl get deployment istiod -n istio-system &> /dev/null; then
    READY=$(kubectl get deployment istiod -n istio-system -o jsonpath='{.status.readyReplicas}')
    DESIRED=$(kubectl get deployment istiod -n istio-system -o jsonpath='{.status.replicas}')
    if [ "$READY" == "$DESIRED" ]; then
        echo -e "${GREEN}✅ Istiod: ${READY}/${DESIRED} ready${NC}"
    else
        echo -e "${YELLOW}⚠️  Istiod: ${READY}/${DESIRED} ready${NC}"
    fi
else
    echo -e "${RED}❌ Istiod not found${NC}"
    exit 1
fi

# Check Ingress Gateway
if kubectl get deployment istio-ingressgateway -n istio-system &> /dev/null; then
    READY=$(kubectl get deployment istio-ingressgateway -n istio-system -o jsonpath='{.status.readyReplicas}')
    DESIRED=$(kubectl get deployment istio-ingressgateway -n istio-system -o jsonpath='{.status.replicas}')
    if [ "$READY" == "$DESIRED" ]; then
        echo -e "${GREEN}✅ Ingress Gateway: ${READY}/${DESIRED} ready${NC}"
    else
        echo -e "${YELLOW}⚠️  Ingress Gateway: ${READY}/${DESIRED} ready${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Ingress Gateway not found${NC}"
fi

# ==============================================================================
# TEST 2: Istio Version
# ==============================================================================
echo -e "\n${BLUE}[2/10] Checking Istio version...${NC}"

if command -v istioctl &> /dev/null; then
    istioctl version --short
    echo -e "${GREEN}✅ istioctl available${NC}"
else
    echo -e "${YELLOW}⚠️  istioctl not in PATH${NC}"
fi

# ==============================================================================
# TEST 3: Namespace Injection
# ==============================================================================
echo -e "\n${BLUE}[3/10] Checking sidecar injection...${NC}"

INJECTION_LABEL=$(kubectl get namespace ${NAMESPACE} -o jsonpath='{.metadata.labels.istio-injection}' 2>/dev/null || echo "")

if [ "$INJECTION_LABEL" == "enabled" ]; then
    echo -e "${GREEN}✅ Sidecar injection enabled in '${NAMESPACE}'${NC}"
else
    echo -e "${YELLOW}⚠️  Sidecar injection not enabled in '${NAMESPACE}'${NC}"
    echo -e "${YELLOW}   Run: ./enable-injection.sh${NC}"
fi

# ==============================================================================
# TEST 4: Pods with Sidecars
# ==============================================================================
echo -e "\n${BLUE}[4/10] Checking pods with sidecars...${NC}"

TOTAL_PODS=$(kubectl get pods -n ${NAMESPACE} --no-headers 2>/dev/null | wc -l)
PODS_WITH_SIDECAR=$(kubectl get pods -n ${NAMESPACE} -o jsonpath='{.items[?(@.spec.containers[*].name=="istio-proxy")].metadata.name}' 2>/dev/null | wc -w)

echo -e "${CYAN}Total pods: ${TOTAL_PODS}${NC}"
echo -e "${CYAN}Pods with sidecar: ${PODS_WITH_SIDECAR}${NC}"

if [ "$PODS_WITH_SIDECAR" -gt 0 ]; then
    echo -e "${GREEN}✅ Sidecars injected${NC}"
    kubectl get pods -n ${NAMESPACE} -o jsonpath='{.items[?(@.spec.containers[*].name=="istio-proxy")].metadata.name}' | tr ' ' '\n' | sed 's/^/  ✓ /'
else
    echo -e "${YELLOW}⚠️  No pods with sidecars found${NC}"
fi

# ==============================================================================
# TEST 5: Proxy Status
# ==============================================================================
echo -e "\n${BLUE}[5/10] Checking proxy status...${NC}"

if command -v istioctl &> /dev/null; then
    istioctl proxy-status | head -20
    
    SYNCED=$(istioctl proxy-status 2>/dev/null | grep -c "SYNCED" || echo 0)
    echo -e "\n${GREEN}✅ Synced proxies: ${SYNCED}${NC}"
else
    echo -e "${YELLOW}⚠️  istioctl not available${NC}"
fi

# ==============================================================================
# TEST 6: Configuration Analysis
# ==============================================================================
echo -e "\n${BLUE}[6/10] Analyzing configuration...${NC}"

if command -v istioctl &> /dev/null; then
    ANALYSIS=$(istioctl analyze -n ${NAMESPACE} 2>&1)
    
    if echo "$ANALYSIS" | grep -q "No validation issues found"; then
        echo -e "${GREEN}✅ No configuration issues${NC}"
    else
        echo -e "${YELLOW}⚠️  Configuration warnings:${NC}"
        echo "$ANALYSIS"
    fi
else
    echo -e "${YELLOW}⚠️  Skipping (istioctl not available)${NC}"
fi

# ==============================================================================
# TEST 7: Istio Resources
# ==============================================================================
echo -e "\n${BLUE}[7/10] Checking Istio resources...${NC}"

# Gateway
GATEWAYS=$(kubectl get gateway -n ${NAMESPACE} 2>/dev/null | tail -n +2 | wc -l)
echo -e "${CYAN}Gateways: ${GATEWAYS}${NC}"

# VirtualService
VIRTUALSERVICES=$(kubectl get virtualservice -n ${NAMESPACE} 2>/dev/null | tail -n +2 | wc -l)
echo -e "${CYAN}VirtualServices: ${VIRTUALSERVICES}${NC}"

# DestinationRule
DESTINATIONRULES=$(kubectl get destinationrule -n ${NAMESPACE} 2>/dev/null | tail -n +2 | wc -l)
echo -e "${CYAN}DestinationRules: ${DESTINATIONRULES}${NC}"

# PeerAuthentication
PEERAUTH=$(kubectl get peerauthentication -n ${NAMESPACE} 2>/dev/null | tail -n +2 | wc -l)
echo -e "${CYAN}PeerAuthentication: ${PEERAUTH}${NC}"

# AuthorizationPolicy
AUTHZ=$(kubectl get authorizationpolicy -n ${NAMESPACE} 2>/dev/null | tail -n +2 | wc -l)
echo -e "${CYAN}AuthorizationPolicy: ${AUTHZ}${NC}"

if [ $((GATEWAYS + VIRTUALSERVICES + DESTINATIONRULES)) -gt 0 ]; then
    echo -e "${GREEN}✅ Istio resources configured${NC}"
else
    echo -e "${YELLOW}⚠️  No Istio traffic management resources found${NC}"
    echo -e "${YELLOW}   Deploy Istio resources with: kubectl apply -f ../full-observability/templates/istio/${NC}"
fi

# ==============================================================================
# TEST 8: Metrics Integration
# ==============================================================================
echo -e "\n${BLUE}[8/10] Checking metrics integration...${NC}"

# Check if Prometheus can scrape Istio
if kubectl get servicemonitor -n istio-system &> /dev/null; then
    SM_COUNT=$(kubectl get servicemonitor -n istio-system 2>/dev/null | tail -n +2 | wc -l)
    echo -e "${GREEN}✅ ServiceMonitors: ${SM_COUNT}${NC}"
else
    echo -e "${YELLOW}⚠️  ServiceMonitor CRD not found${NC}"
fi

# Check Istiod metrics endpoint
ISTIOD_POD=$(kubectl get pod -n istio-system -l app=istiod -o jsonpath='{.items[0].metadata.name}')
if [ ! -z "$ISTIOD_POD" ]; then
    if kubectl exec -n istio-system ${ISTIOD_POD} -- curl -s http://localhost:15014/metrics > /dev/null; then
        echo -e "${GREEN}✅ Istiod metrics endpoint accessible${NC}"
    fi
fi

# ==============================================================================
# TEST 9: Tracing Integration
# ==============================================================================
echo -e "\n${BLUE}[9/10] Checking tracing integration...${NC}"

# Check if Jaeger is accessible from Istio
if kubectl get svc jaeger -n ${NAMESPACE} &> /dev/null; then
    echo -e "${GREEN}✅ Jaeger service found in '${NAMESPACE}'${NC}"
    
    # Check Istio tracing configuration
    TRACING_CONFIG=$(kubectl get configmap istio -n istio-system -o yaml 2>/dev/null | grep -A5 "tracing:" || echo "")
    if [ ! -z "$TRACING_CONFIG" ]; then
        echo -e "${GREEN}✅ Tracing configured in Istio${NC}"
    else
        echo -e "${YELLOW}⚠️  Tracing not configured${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Jaeger service not found${NC}"
fi

# ==============================================================================
# TEST 10: Test Traffic
# ==============================================================================
echo -e "\n${BLUE}[10/10] Testing traffic through mesh...${NC}"

# Get a pod to test from
TEST_POD=$(kubectl get pod -n ${NAMESPACE} -l component=microservice -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

if [ ! -z "$TEST_POD" ]; then
    echo -e "${CYAN}Test pod: ${TEST_POD}${NC}"
    
    # Test internal service call
    echo -e "${YELLOW}Testing internal service call...${NC}"
    RESULT=$(kubectl exec ${TEST_POD} -n ${NAMESPACE} -c istio-proxy -- curl -s -o /dev/null -w "%{http_code}" http://user-service:8080/health 2>/dev/null || echo "000")
    
    if [ "$RESULT" == "200" ]; then
        echo -e "${GREEN}✅ Service mesh traffic working (HTTP ${RESULT})${NC}"
    else
        echo -e "${YELLOW}⚠️  Service call returned HTTP ${RESULT}${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  No test pod found${NC}"
fi

# ==============================================================================
# Summary
# ==============================================================================
echo -e "\n${CYAN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                       TEST SUMMARY                                 ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

echo -e "${YELLOW}Istio Components:${NC}"
kubectl get pods -n istio-system

echo -e "\n${YELLOW}Application Pods (with sidecars):${NC}"
kubectl get pods -n ${NAMESPACE} | head -10

echo -e "\n${MAGENTA}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${MAGENTA}USEFUL DASHBOARDS:${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════════════════════════════${NC}"

if command -v istioctl &> /dev/null; then
    echo -e "\n${CYAN}Istio Control Plane:${NC}"
    echo -e "   istioctl dashboard controlz deployment/istiod -n istio-system"
    
    echo -e "\n${CYAN}Envoy Admin (for a pod):${NC}"
    echo -e "   istioctl dashboard envoy <pod-name> -n ${NAMESPACE}"
    
    echo -e "\n${CYAN}Prometheus:${NC}"
    echo -e "   kubectl port-forward -n ${NAMESPACE} svc/prometheus 9090:9090"
    
    echo -e "\n${CYAN}Grafana:${NC}"
    echo -e "   kubectl port-forward -n ${NAMESPACE} svc/grafana 3000:3000"
    
    echo -e "\n${CYAN}Jaeger:${NC}"
    echo -e "   kubectl port-forward -n ${NAMESPACE} svc/jaeger 16686:16686"
    
    echo -e "\n${CYAN}Kiali (if installed):${NC}"
    echo -e "   istioctl dashboard kiali"
fi

echo -e "\n${MAGENTA}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${MAGENTA}TROUBLESHOOTING:${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════════════════════════════${NC}"

echo -e "\n${BLUE}View Envoy logs:${NC}"
echo -e "   kubectl logs <pod-name> -c istio-proxy -n ${NAMESPACE}"

echo -e "\n${BLUE}Check Envoy configuration:${NC}"
echo -e "   istioctl proxy-config routes <pod-name> -n ${NAMESPACE}"
echo -e "   istioctl proxy-config clusters <pod-name> -n ${NAMESPACE}"

echo -e "\n${BLUE}Debug connectivity:${NC}"
echo -e "   istioctl experimental describe pod <pod-name> -n ${NAMESPACE}"

echo ""

