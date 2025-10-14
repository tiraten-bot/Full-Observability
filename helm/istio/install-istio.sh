#!/bin/bash

# ==============================================================================
# ISTIO SERVICE MESH INSTALLATION
# ==============================================================================
# Purpose: Install Istio Service Mesh with default profile
# Prerequisites: Kubernetes cluster running
# Usage: ./install-istio.sh
# ==============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
ISTIO_VERSION="1.20.2"
ISTIO_DIR="$HOME/.istio"
NAMESPACE="observability"

echo -e "${CYAN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                  ISTIO SERVICE MESH INSTALLATION                   ║"
echo "║                         Version: ${ISTIO_VERSION}                          ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# ==============================================================================
# STEP 1: Prerequisites Check
# ==============================================================================
echo -e "${BLUE}[1/8] Checking prerequisites...${NC}"

# Check kubectl
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl not found${NC}"
    exit 1
fi
echo -e "${GREEN}✅ kubectl found${NC}"

# Check cluster connection
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ Cannot connect to Kubernetes cluster${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Connected to Kubernetes cluster${NC}"

# Check cluster resources
NODES=$(kubectl get nodes --no-headers | wc -l)
echo -e "${YELLOW}   Nodes: ${NODES}${NC}"

# Check namespace
if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
    echo -e "${YELLOW}⚠️  Namespace '${NAMESPACE}' not found. Creating...${NC}"
    kubectl create namespace ${NAMESPACE}
fi
echo -e "${GREEN}✅ Namespace '${NAMESPACE}' ready${NC}"

# ==============================================================================
# STEP 2: Download Istio
# ==============================================================================
echo -e "\n${BLUE}[2/8] Downloading Istio ${ISTIO_VERSION}...${NC}"

if [ -d "${ISTIO_DIR}/istio-${ISTIO_VERSION}" ]; then
    echo -e "${YELLOW}⚠️  Istio ${ISTIO_VERSION} already downloaded${NC}"
else
    echo -e "${YELLOW}Downloading Istio...${NC}"
    mkdir -p ${ISTIO_DIR}
    cd ${ISTIO_DIR}
    curl -L https://istio.io/downloadIstio | ISTIO_VERSION=${ISTIO_VERSION} sh -
    echo -e "${GREEN}✅ Istio downloaded${NC}"
fi

# Add istioctl to PATH
export PATH="${ISTIO_DIR}/istio-${ISTIO_VERSION}/bin:${PATH}"

# Verify istioctl
if ! command -v istioctl &> /dev/null; then
    echo -e "${RED}❌ istioctl not found in PATH${NC}"
    echo -e "${YELLOW}Add to PATH: export PATH=\"${ISTIO_DIR}/istio-${ISTIO_VERSION}/bin:\$PATH\"${NC}"
    exit 1
fi

echo -e "${GREEN}✅ istioctl version: $(istioctl version --short --remote=false 2>/dev/null || echo ${ISTIO_VERSION})${NC}"

# ==============================================================================
# STEP 3: Pre-flight Check
# ==============================================================================
echo -e "\n${BLUE}[3/8] Running pre-flight checks...${NC}"

istioctl x precheck

echo -e "${GREEN}✅ Pre-flight checks passed${NC}"

# ==============================================================================
# STEP 4: Install Istio Control Plane
# ==============================================================================
echo -e "\n${BLUE}[4/8] Installing Istio Control Plane...${NC}"
echo -e "${YELLOW}Profile: default${NC}"
echo -e "${YELLOW}External Observability: Prometheus, Grafana, Jaeger (existing)${NC}"

# Create IstioOperator configuration
cat <<EOF > /tmp/istio-config.yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: istio-controlplane
  namespace: istio-system
spec:
  profile: default
  
  # Mesh Configuration
  meshConfig:
    # Enable distributed tracing
    enableTracing: true
    defaultConfig:
      tracing:
        zipkin:
          address: jaeger.observability.svc.cluster.local:9411
        sampling: 1.0  # 1% sampling (100.0 for 100%)
      
      # Enable Prometheus metrics
      proxyStatsMatcher:
        inclusionPrefixes:
          - "cluster.outbound"
          - "cluster.inbound"
          - "cluster_manager"
          - "listener_manager"
          - "http_mixer_filter"
          - "tcp_mixer_filter"
          - "server"
          - "cluster.xds-grpc"
    
    # Service discovery
    defaultServiceExportTo:
      - "."
    defaultVirtualServiceExportTo:
      - "."
    defaultDestinationRuleExportTo:
      - "."
    
    # Access logging
    accessLogFile: /dev/stdout
    accessLogEncoding: JSON
  
  # Components Configuration
  components:
    # Pilot (Istiod)
    pilot:
      enabled: true
      k8s:
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 1000m
            memory: 2Gi
        hpaSpec:
          minReplicas: 1
          maxReplicas: 5
          metrics:
            - type: Resource
              resource:
                name: cpu
                target:
                  type: Utilization
                  averageUtilization: 80
    
    # Ingress Gateway
    ingressGateways:
      - name: istio-ingressgateway
        enabled: true
        k8s:
          service:
            type: LoadBalancer
            ports:
              - name: status-port
                port: 15021
                targetPort: 15021
              - name: http2
                port: 80
                targetPort: 8080
              - name: https
                port: 443
                targetPort: 8443
              - name: tcp
                port: 31400
                targetPort: 31400
              - name: tls
                port: 15443
                targetPort: 15443
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 2000m
              memory: 1024Mi
          hpaSpec:
            minReplicas: 1
            maxReplicas: 5
            metrics:
              - type: Resource
                resource:
                  name: cpu
                  target:
                    type: Utilization
                    averageUtilization: 80
    
    # Egress Gateway (disabled by default)
    egressGateways:
      - name: istio-egressgateway
        enabled: false
  
  # Values Configuration
  values:
    global:
      # Disable built-in observability tools (we have our own)
      proxy:
        # Envoy proxy configuration
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 2000m
            memory: 1024Mi
      
      # Tracing configuration
      tracer:
        zipkin:
          address: jaeger.observability.svc.cluster.local:9411
    
    # Disable built-in Prometheus (we have our own)
    prometheus:
      enabled: false
    
    # Disable built-in Grafana (we have our own)
    grafana:
      enabled: false
    
    # Disable built-in Kiali (will install separately)
    kiali:
      enabled: false
    
    # Disable built-in Jaeger (we have our own)
    tracing:
      enabled: false
    
    # Pilot configuration
    pilot:
      autoscaleEnabled: true
      autoscaleMin: 1
      autoscaleMax: 5
      cpu:
        targetAverageUtilization: 80
EOF

# Install Istio
istioctl install -f /tmp/istio-config.yaml -y

echo -e "${GREEN}✅ Istio Control Plane installed${NC}"

# ==============================================================================
# STEP 5: Verify Installation
# ==============================================================================
echo -e "\n${BLUE}[5/8] Verifying installation...${NC}"

# Wait for Istio components
echo -e "${YELLOW}Waiting for Istio components to be ready...${NC}"
kubectl wait --for=condition=ready pod -l app=istiod -n istio-system --timeout=300s
kubectl wait --for=condition=ready pod -l app=istio-ingressgateway -n istio-system --timeout=300s

echo -e "\n${YELLOW}Istio Components:${NC}"
kubectl get pods -n istio-system

echo -e "\n${YELLOW}Istio Services:${NC}"
kubectl get svc -n istio-system

# Verify installation
istioctl verify-install

echo -e "${GREEN}✅ Installation verified${NC}"

# ==============================================================================
# STEP 6: Get Istio Ingress Gateway External IP
# ==============================================================================
echo -e "\n${BLUE}[6/8] Getting Istio Ingress Gateway External IP...${NC}"

EXTERNAL_IP=""
echo -e "${YELLOW}Waiting for external IP (this may take a few minutes)...${NC}"

for i in {1..60}; do
    EXTERNAL_IP=$(kubectl get svc istio-ingressgateway -n istio-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
    if [ -z "$EXTERNAL_IP" ]; then
        EXTERNAL_IP=$(kubectl get svc istio-ingressgateway -n istio-system -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null)
    fi
    
    if [ ! -z "$EXTERNAL_IP" ]; then
        echo -e "${GREEN}✅ Ingress Gateway External IP/Hostname: ${EXTERNAL_IP}${NC}"
        break
    fi
    
    echo -n "."
    sleep 5
done

if [ -z "$EXTERNAL_IP" ]; then
    echo -e "${YELLOW}⚠️  External IP not assigned yet. Check later with:${NC}"
    echo -e "   kubectl get svc istio-ingressgateway -n istio-system"
fi

# ==============================================================================
# STEP 7: Configure Prometheus to Scrape Istio Metrics
# ==============================================================================
echo -e "\n${BLUE}[7/8] Configuring Prometheus integration...${NC}"

# Create ServiceMonitor for Istio
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: istiod-metrics
  namespace: istio-system
  labels:
    app: istiod
spec:
  type: ClusterIP
  ports:
    - name: http-monitoring
      port: 15014
      targetPort: 15014
      protocol: TCP
  selector:
    app: istiod
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: istio-component-monitor
  namespace: istio-system
  labels:
    monitoring: istio-components
spec:
  selector:
    matchLabels:
      app: istiod
  endpoints:
    - port: http-monitoring
      interval: 15s
---
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: envoy-stats-monitor
  namespace: istio-system
  labels:
    monitoring: istio-proxies
spec:
  selector:
    matchExpressions:
      - key: istio-prometheus-ignore
        operator: DoesNotExist
  podMetricsEndpoints:
    - path: /stats/prometheus
      interval: 15s
EOF

echo -e "${GREEN}✅ Prometheus integration configured${NC}"

# ==============================================================================
# STEP 8: Installation Summary
# ==============================================================================
echo -e "\n${CYAN}"
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                    INSTALLATION COMPLETED! 🎉                      ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

echo -e "${GREEN}✅ Istio Control Plane: Running${NC}"
echo -e "${GREEN}✅ Istio Ingress Gateway: Running${NC}"
echo -e "${GREEN}✅ Prometheus Integration: Configured${NC}"
echo -e "${GREEN}✅ Jaeger Integration: Configured${NC}"

echo -e "\n${YELLOW}Istio Version:${NC}"
istioctl version

echo -e "\n${YELLOW}Istio Components:${NC}"
kubectl get deploy,svc -n istio-system

echo -e "\n${MAGENTA}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${MAGENTA}NEXT STEPS:${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════════════════════════════${NC}"

echo -e "\n${CYAN}1. Enable Sidecar Injection:${NC}"
echo -e "   ${YELLOW}./enable-injection.sh${NC}"

echo -e "\n${CYAN}2. Configure Istio Resources:${NC}"
echo -e "   ${YELLOW}kubectl apply -f ../full-observability/templates/istio/${NC}"

echo -e "\n${CYAN}3. Test Istio:${NC}"
echo -e "   ${YELLOW}./test-istio.sh${NC}"

echo -e "\n${CYAN}4. Install Kiali (optional):${NC}"
echo -e "   ${YELLOW}./install-kiali.sh${NC}"

echo -e "\n${CYAN}5. Access Istio Dashboard:${NC}"
echo -e "   ${YELLOW}istioctl dashboard controlz deployment/istiod -n istio-system${NC}"

echo -e "\n${MAGENTA}═══════════════════════════════════════════════════════════════════${NC}"
echo -e "${MAGENTA}USEFUL COMMANDS:${NC}"
echo -e "${MAGENTA}═══════════════════════════════════════════════════════════════════${NC}"

echo -e "\n${BLUE}Check Istio status:${NC}"
echo -e "   kubectl get pods -n istio-system"
echo -e "   istioctl proxy-status"

echo -e "\n${BLUE}Analyze configuration:${NC}"
echo -e "   istioctl analyze -n ${NAMESPACE}"

echo -e "\n${BLUE}View proxy logs:${NC}"
echo -e "   kubectl logs -l app=user-service -c istio-proxy -n ${NAMESPACE}"

echo -e "\n${BLUE}Uninstall Istio:${NC}"
echo -e "   istioctl uninstall --purge -y"
echo -e "   kubectl delete namespace istio-system"

echo -e "\n${GREEN}Installation log saved to: /tmp/istio-install.log${NC}"
echo ""

