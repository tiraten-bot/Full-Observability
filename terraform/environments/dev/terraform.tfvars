# ==============================================================================
# DEVELOPMENT ENVIRONMENT VARIABLES
# ==============================================================================

project_name = "full-observability"
environment  = "dev"
aws_region   = "us-east-1"

# VPC
vpc_cidr = "10.0.0.0/16"

# EKS
eks_cluster_version        = "1.28"
eks_node_group_desired_size = 2
eks_node_group_min_size     = 1
eks_node_group_max_size     = 4

# RDS
rds_instance_class        = "db.t3.small"
rds_allocated_storage     = 20
rds_max_allocated_storage = 100

# Redis
redis_node_type = "cache.t3.micro"

# Kafka
kafka_instance_type   = "kafka.t3.small"
kafka_number_of_nodes = 1  # Single node for dev
kafka_ebs_volume_size = 50

# Domain
domain_name = "dev.example.com"

# Alerts
alert_email_endpoints = ["dev-team@example.com"]

# Cost optimization
enable_spot_instances = true
enable_auto_shutdown  = true

