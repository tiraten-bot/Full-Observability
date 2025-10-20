# ==============================================================================
# TERRAFORM MAIN CONFIGURATION
# ==============================================================================
# Purpose: AWS Infrastructure for Full Observability Microservices
# 
# Architecture:
#   - VPC with public/private subnets across 3 AZs
#   - EKS Cluster for Kubernetes workloads
#   - RDS PostgreSQL for databases
#   - ElastiCache Redis for caching
#   - MSK (Managed Streaming for Kafka)
#   - Application Load Balancer
#   - Route53 for DNS
#   - ACM for SSL certificates
#   - CloudWatch for logging
#
# Note: This is infrastructure as code demonstration
#       Not meant for actual deployment without review
# ==============================================================================

terraform {
  required_version = ">= 1.6.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.17"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.24"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.0"
    }
  }
  
  # Remote state storage (recommended for production)
  backend "s3" {
    bucket         = "full-observability-terraform-state"
    key            = "terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock"
  }
}

# ==============================================================================
# PROVIDER CONFIGURATION
# ==============================================================================

provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = {
      Project     = "Full-Observability"
      Environment = var.environment
      ManagedBy   = "Terraform"
      Owner       = "DevOps Team"
    }
  }
}

provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_ca_certificate)
  
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args = [
      "eks",
      "get-token",
      "--cluster-name",
      module.eks.cluster_name
    ]
  }
}

provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_ca_certificate)
    
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args = [
        "eks",
        "get-token",
        "--cluster-name",
        module.eks.cluster_name
      ]
    }
  }
}

# ==============================================================================
# DATA SOURCES
# ==============================================================================

data "aws_availability_zones" "available" {
  state = "available"
  
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_caller_identity" "current" {}

# ==============================================================================
# LOCAL VARIABLES
# ==============================================================================

locals {
  cluster_name = "${var.project_name}-${var.environment}-eks"
  
  vpc_cidr = var.vpc_cidr
  azs      = slice(data.aws_availability_zones.available.names, 0, 3)
  
  # Subnet CIDR calculations
  private_subnets = [
    cidrsubnet(local.vpc_cidr, 4, 0),  # 10.0.0.0/20
    cidrsubnet(local.vpc_cidr, 4, 1),  # 10.0.16.0/20
    cidrsubnet(local.vpc_cidr, 4, 2),  # 10.0.32.0/20
  ]
  
  public_subnets = [
    cidrsubnet(local.vpc_cidr, 4, 3),  # 10.0.48.0/20
    cidrsubnet(local.vpc_cidr, 4, 4),  # 10.0.64.0/20
    cidrsubnet(local.vpc_cidr, 4, 5),  # 10.0.80.0/20
  ]
  
  database_subnets = [
    cidrsubnet(local.vpc_cidr, 4, 6),  # 10.0.96.0/20
    cidrsubnet(local.vpc_cidr, 4, 7),  # 10.0.112.0/20
    cidrsubnet(local.vpc_cidr, 4, 8),  # 10.0.128.0/20
  ]
  
  tags = {
    Project     = var.project_name
    Environment = var.environment
    Terraform   = "true"
  }
}

# ==============================================================================
# VPC MODULE
# ==============================================================================

module "vpc" {
  source = "./modules/vpc"
  
  project_name = var.project_name
  environment  = var.environment
  
  vpc_cidr            = local.vpc_cidr
  azs                 = local.azs
  private_subnets     = local.private_subnets
  public_subnets      = local.public_subnets
  database_subnets    = local.database_subnets
  
  enable_nat_gateway   = true
  enable_dns_hostnames = true
  enable_dns_support   = true
  
  tags = local.tags
}

# ==============================================================================
# EKS CLUSTER MODULE
# ==============================================================================

module "eks" {
  source = "./modules/eks"
  
  cluster_name    = local.cluster_name
  cluster_version = var.eks_cluster_version
  
  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.private_subnets
  
  # Node groups configuration
  node_groups = {
    general = {
      desired_capacity = var.eks_node_group_desired_size
      max_capacity     = var.eks_node_group_max_size
      min_capacity     = var.eks_node_group_min_size
      instance_types   = ["t3.medium"]
      capacity_type    = "ON_DEMAND"
      
      labels = {
        role = "general"
      }
      
      tags = {
        Name = "${local.cluster_name}-general-node"
      }
    }
    
    observability = {
      desired_capacity = 2
      max_capacity     = 5
      min_capacity     = 2
      instance_types   = ["t3.large"]
      capacity_type    = "ON_DEMAND"
      
      labels = {
        role = "observability"
      }
      
      taints = [{
        key    = "observability"
        value  = "true"
        effect = "NoSchedule"
      }]
      
      tags = {
        Name = "${local.cluster_name}-observability-node"
      }
    }
  }
  
  tags = local.tags
}

# ==============================================================================
# RDS POSTGRESQL MODULE
# ==============================================================================

module "rds" {
  source = "./modules/rds"
  
  identifier = "${var.project_name}-${var.environment}-postgres"
  
  engine         = "postgres"
  engine_version = "15.4"
  instance_class = var.rds_instance_class
  
  allocated_storage     = var.rds_allocated_storage
  max_allocated_storage = var.rds_max_allocated_storage
  storage_encrypted     = true
  
  db_name  = "postgres"
  username = var.db_master_username
  password = var.db_master_password
  port     = 5432
  
  vpc_id                 = module.vpc.vpc_id
  subnet_ids             = module.vpc.database_subnets
  vpc_security_group_ids = [module.security_groups.rds_sg_id]
  
  multi_az               = var.environment == "prod" ? true : false
  backup_retention_period = var.environment == "prod" ? 30 : 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "mon:04:00-mon:05:00"
  
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  
  tags = local.tags
}

# ==============================================================================
# ELASTICACHE REDIS MODULE
# ==============================================================================

module "elasticache" {
  source = "./modules/elasticache"
  
  cluster_id = "${var.project_name}-${var.environment}-redis"
  
  engine         = "redis"
  engine_version = "7.0"
  node_type      = var.redis_node_type
  num_cache_nodes = var.environment == "prod" ? 3 : 1
  
  port                 = 6379
  parameter_group_name = "default.redis7"
  
  subnet_group_name      = module.vpc.elasticache_subnet_group_name
  security_group_ids     = [module.security_groups.redis_sg_id]
  
  automatic_failover_enabled = var.environment == "prod" ? true : false
  
  tags = local.tags
}

# ==============================================================================
# MSK (KAFKA) MODULE
# ==============================================================================

module "msk" {
  source = "./modules/msk"
  
  cluster_name = "${var.project_name}-${var.environment}-kafka"
  
  kafka_version   = "3.5.1"
  number_of_nodes = var.environment == "prod" ? 3 : 1
  instance_type   = var.kafka_instance_type
  
  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets
  
  security_group_ids = [module.security_groups.kafka_sg_id]
  
  encryption_in_transit_client_broker = "TLS"
  encryption_at_rest_kms_key_arn      = module.kms.kafka_key_arn
  
  # Storage
  ebs_volume_size = var.kafka_ebs_volume_size
  
  # Logging
  cloudwatch_logs_enabled = true
  s3_logs_enabled         = true
  s3_logs_bucket          = module.s3.logs_bucket_id
  
  tags = local.tags
}

# ==============================================================================
# SECURITY GROUPS MODULE
# ==============================================================================

module "security_groups" {
  source = "./modules/security-groups"
  
  project_name = var.project_name
  environment  = var.environment
  
  vpc_id = module.vpc.vpc_id
  
  # CIDR blocks
  vpc_cidr = local.vpc_cidr
  
  tags = local.tags
}

# ==============================================================================
# KMS MODULE (Encryption Keys)
# ==============================================================================

module "kms" {
  source = "./modules/kms"
  
  project_name = var.project_name
  environment  = var.environment
  
  tags = local.tags
}

# ==============================================================================
# S3 MODULE (Logs, Backups)
# ==============================================================================

module "s3" {
  source = "./modules/s3"
  
  project_name = var.project_name
  environment  = var.environment
  
  tags = local.tags
}

# ==============================================================================
# ALB (Application Load Balancer) MODULE
# ==============================================================================

module "alb" {
  source = "./modules/alb"
  
  name = "${var.project_name}-${var.environment}-alb"
  
  vpc_id          = module.vpc.vpc_id
  subnets         = module.vpc.public_subnets
  security_groups = [module.security_groups.alb_sg_id]
  
  # SSL certificate
  certificate_arn = module.acm.certificate_arn
  
  tags = local.tags
}

# ==============================================================================
# ACM (Certificate Manager) MODULE
# ==============================================================================

module "acm" {
  source = "./modules/acm"
  
  domain_name               = var.domain_name
  subject_alternative_names = var.subject_alternative_names
  
  tags = local.tags
}

# ==============================================================================
# ROUTE53 MODULE (DNS)
# ==============================================================================

module "route53" {
  source = "./modules/route53"
  
  domain_name = var.domain_name
  
  # A records for services
  records = {
    api      = module.alb.alb_dns_name
    grafana  = module.alb.alb_dns_name
    jaeger   = module.alb.alb_dns_name
  }
  
  tags = local.tags
}

# ==============================================================================
# IAM MODULE (Roles and Policies)
# ==============================================================================

module "iam" {
  source = "./modules/iam"
  
  project_name = var.project_name
  environment  = var.environment
  
  eks_cluster_name = module.eks.cluster_name
  
  tags = local.tags
}

# ==============================================================================
# CLOUDWATCH MODULE (Logging and Monitoring)
# ==============================================================================

module "cloudwatch" {
  source = "./modules/cloudwatch"
  
  project_name = var.project_name
  environment  = var.environment
  
  # Log retention
  log_retention_days = var.environment == "prod" ? 90 : 30
  
  # Alarms
  enable_alarms = var.environment == "prod" ? true : false
  sns_topic_arn = module.sns.topic_arn
  
  tags = local.tags
}

# ==============================================================================
# SNS MODULE (Notifications)
# ==============================================================================

module "sns" {
  source = "./modules/sns"
  
  project_name = var.project_name
  environment  = var.environment
  
  # Email subscriptions for alerts
  email_endpoints = var.alert_email_endpoints
  
  tags = local.tags
}

