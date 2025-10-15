# ==============================================================================
# ANSIBLE MAKEFILE
# ==============================================================================
# Purpose: Shortcuts for Ansible playbook execution
# ==============================================================================

.PHONY: help install requirements deploy rollback scale health-check backup clean

# Default target
.DEFAULT_GOAL := help

# Variables
ENV ?= dev
SERVICE ?= all
REPLICAS ?= 3

# ==============================================================================
# HELP
# ==============================================================================
help:
	@echo "╔════════════════════════════════════════════════════════════════════╗"
	@echo "║              Ansible Automation Commands                           ║"
	@echo "╚════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "Setup:"
	@echo "  make install          Install Ansible and dependencies"
	@echo "  make requirements     Install Ansible collections"
	@echo ""
	@echo "Deployment:"
	@echo "  make deploy           Deploy full stack (ENV=dev|staging|prod)"
	@echo "  make rollback         Rollback service (SERVICE=user-service)"
	@echo "  make scale            Scale service (SERVICE=user-service REPLICAS=5)"
	@echo ""
	@echo "Operations:"
	@echo "  make health-check     Run health checks on all services"
	@echo "  make backup           Backup databases and configurations"
	@echo "  make clean            Remove local artifacts"
	@echo ""
	@echo "Examples:"
	@echo "  make deploy ENV=prod"
	@echo "  make scale SERVICE=user-service REPLICAS=10"
	@echo "  make rollback SERVICE=payment-service"
	@echo ""

# ==============================================================================
# INSTALLATION
# ==============================================================================
install:
	@echo "Installing Ansible..."
	pip3 install ansible ansible-core kubernetes
	@echo "✅ Ansible installed"

requirements:
	@echo "Installing Ansible collections..."
	ansible-galaxy collection install -r requirements.yml
	@echo "✅ Collections installed"

# ==============================================================================
# DEPLOYMENT
# ==============================================================================
deploy:
	@echo "Deploying Full Observability Stack (Environment: $(ENV))..."
	ENVIRONMENT=$(ENV) ansible-playbook playbooks/deploy-all.yml
	@echo "✅ Deployment complete"

rollback:
	@echo "Rolling back service: $(SERVICE)..."
	ansible-playbook playbooks/rollback.yml -e service=$(SERVICE)
	@echo "✅ Rollback complete"

scale:
	@echo "Scaling $(SERVICE) to $(REPLICAS) replicas..."
	ansible-playbook playbooks/scale.yml -e "service=$(SERVICE) replicas=$(REPLICAS)"
	@echo "✅ Scaling complete"

# ==============================================================================
# OPERATIONS
# ==============================================================================
health-check:
	@echo "Running health checks..."
	ansible-playbook playbooks/health-check.yml
	@echo "✅ Health check complete"

backup:
	@echo "Starting backup..."
	ansible-playbook playbooks/backup.yml
	@echo "✅ Backup complete"

# ==============================================================================
# CLEANUP
# ==============================================================================
clean:
	@echo "Cleaning up..."
	rm -f ansible.log
	rm -rf retry/
	rm -rf /tmp/ansible_facts/
	@echo "✅ Cleanup complete"

# ==============================================================================
# ADVANCED OPERATIONS
# ==============================================================================
canary-deploy:
	@echo "Canary deployment for $(SERVICE)..."
	ansible-playbook playbooks/canary-deploy.yml -e "service=$(SERVICE) version=$(VERSION)"

update-config:
	@echo "Updating configuration..."
	ansible-playbook playbooks/update-config.yml -e "config_file=$(CONFIG)"

restart-service:
	@echo "Restarting $(SERVICE)..."
	kubectl rollout restart deployment/$(SERVICE) -n observability

logs:
	@echo "Viewing logs for $(SERVICE)..."
	kubectl logs -f -l app=$(SERVICE) -n observability

describe:
	@echo "Describing $(SERVICE)..."
	kubectl describe deployment/$(SERVICE) -n observability

