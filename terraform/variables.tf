# ==============================================================================
# TERRAFORM VARIABLES
# ==============================================================================

# ==============================================================================
# General Configuration
# ==============================================================================

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "full-observability"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod"
  }
}

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "domain_name" {
  description = "Primary domain name for the application"
  type        = string
  default     = "example.com"
}

variable "subject_alternative_names" {
  description = "Additional domain names for SSL certificate"
  type        = list(string)
  default = [
    "*.example.com",
    "api.example.com",
    "grafana.example.com",
    "jaeger.example.com"
  ]
}

# ==============================================================================
# VPC Configuration
# ==============================================================================

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

# ==============================================================================
# EKS Configuration
# ==============================================================================

variable "eks_cluster_version" {
  description = "Kubernetes version for EKS cluster"
  type        = string
  default     = "1.28"
}

variable "eks_node_group_desired_size" {
  description = "Desired number of worker nodes"
  type        = number
  default     = 3
}

variable "eks_node_group_min_size" {
  description = "Minimum number of worker nodes"
  type        = number
  default     = 2
}

variable "eks_node_group_max_size" {
  description = "Maximum number of worker nodes"
  type        = number
  default     = 10
}

variable "eks_node_instance_types" {
  description = "Instance types for EKS nodes"
  type        = list(string)
  default     = ["t3.medium", "t3.large"]
}

# ==============================================================================
# RDS Configuration
# ==============================================================================

variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.medium"
}

variable "rds_allocated_storage" {
  description = "Initial storage size in GB"
  type        = number
  default     = 100
}

variable "rds_max_allocated_storage" {
  description = "Maximum storage size for autoscaling in GB"
  type        = number
  default     = 500
}

variable "db_master_username" {
  description = "Master username for RDS"
  type        = string
  default     = "postgres"
  sensitive   = true
}

variable "db_master_password" {
  description = "Master password for RDS"
  type        = string
  sensitive   = true
}

# ==============================================================================
# ElastiCache Redis Configuration
# ==============================================================================

variable "redis_node_type" {
  description = "ElastiCache Redis node type"
  type        = string
  default     = "cache.t3.medium"
}

variable "redis_num_cache_nodes" {
  description = "Number of cache nodes"
  type        = number
  default     = 1
}

# ==============================================================================
# MSK (Kafka) Configuration
# ==============================================================================

variable "kafka_instance_type" {
  description = "MSK broker instance type"
  type        = string
  default     = "kafka.t3.small"
}

variable "kafka_number_of_nodes" {
  description = "Number of Kafka broker nodes"
  type        = number
  default     = 3
  
  validation {
    condition     = var.kafka_number_of_nodes >= 3
    error_message = "Kafka requires minimum 3 nodes for high availability"
  }
}

variable "kafka_ebs_volume_size" {
  description = "EBS volume size for Kafka brokers in GB"
  type        = number
  default     = 100
}

# ==============================================================================
# Monitoring & Alerting
# ==============================================================================

variable "alert_email_endpoints" {
  description = "Email addresses for CloudWatch alerts"
  type        = list(string)
  default     = ["devops@example.com"]
}

variable "enable_container_insights" {
  description = "Enable CloudWatch Container Insights for EKS"
  type        = bool
  default     = true
}

# ==============================================================================
# Cost Optimization
# ==============================================================================

variable "enable_spot_instances" {
  description = "Use spot instances for cost savings"
  type        = bool
  default     = false
}

variable "enable_auto_shutdown" {
  description = "Auto shutdown non-prod resources at night"
  type        = bool
  default     = true
}

