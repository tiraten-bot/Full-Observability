# ==============================================================================
# PRODUCTION ENVIRONMENT VARIABLES
# ==============================================================================

project_name = "full-observability"
environment  = "prod"
aws_region   = "us-east-1"

# VPC
vpc_cidr = "10.0.0.0/16"

# EKS
eks_cluster_version         = "1.28"
eks_node_group_desired_size = 6
eks_node_group_min_size     = 3
eks_node_group_max_size     = 20

# RDS - Multi-AZ for high availability
rds_instance_class        = "db.r6g.xlarge"
rds_allocated_storage     = 500
rds_max_allocated_storage = 2000

# Redis - Cluster mode for high availability
redis_node_type = "cache.r6g.large"

# Kafka - 3+ nodes for high availability
kafka_instance_type   = "kafka.m5.large"
kafka_number_of_nodes = 3
kafka_ebs_volume_size = 500

# Domain
domain_name = "example.com"
subject_alternative_names = [
  "*.example.com",
  "api.example.com",
  "grafana.example.com",
  "jaeger.example.com",
  "prometheus.example.com"
]

# Alerts
alert_email_endpoints = [
  "devops@example.com",
  "oncall@example.com"
]

# Production settings
enable_spot_instances = false
enable_auto_shutdown  = false

