# ==============================================================================
# TERRAFORM OUTPUTS
# ==============================================================================
# Purpose: Export important values after infrastructure creation
# ==============================================================================

# ==============================================================================
# VPC Outputs
# ==============================================================================

output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "vpc_cidr" {
  description = "VPC CIDR block"
  value       = module.vpc.vpc_cidr
}

output "private_subnets" {
  description = "Private subnet IDs"
  value       = module.vpc.private_subnets
}

output "public_subnets" {
  description = "Public subnet IDs"
  value       = module.vpc.public_subnets
}

# ==============================================================================
# EKS Outputs
# ==============================================================================

output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "eks_cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = module.eks.cluster_endpoint
  sensitive   = true
}

output "eks_cluster_security_group_id" {
  description = "EKS cluster security group ID"
  value       = module.eks.cluster_security_group_id
}

output "eks_cluster_arn" {
  description = "EKS cluster ARN"
  value       = module.eks.cluster_arn
}

output "configure_kubectl" {
  description = "Command to configure kubectl"
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${module.eks.cluster_name}"
}

# ==============================================================================
# RDS Outputs
# ==============================================================================

output "rds_endpoint" {
  description = "RDS instance endpoint"
  value       = module.rds.db_instance_endpoint
  sensitive   = true
}

output "rds_instance_id" {
  description = "RDS instance ID"
  value       = module.rds.db_instance_id
}

output "rds_arn" {
  description = "RDS instance ARN"
  value       = module.rds.db_instance_arn
}

# ==============================================================================
# ElastiCache Outputs
# ==============================================================================

output "redis_endpoint" {
  description = "Redis cluster endpoint"
  value       = module.elasticache.cache_cluster_address
  sensitive   = true
}

output "redis_port" {
  description = "Redis port"
  value       = module.elasticache.cache_cluster_port
}

# ==============================================================================
# MSK (Kafka) Outputs
# ==============================================================================

output "kafka_bootstrap_brokers" {
  description = "Kafka bootstrap brokers"
  value       = module.msk.bootstrap_brokers
  sensitive   = true
}

output "kafka_bootstrap_brokers_tls" {
  description = "Kafka bootstrap brokers (TLS)"
  value       = module.msk.bootstrap_brokers_tls
  sensitive   = true
}

output "kafka_zookeeper_connect" {
  description = "Kafka Zookeeper connection string"
  value       = module.msk.zookeeper_connect_string
  sensitive   = true
}

# ==============================================================================
# Load Balancer Outputs
# ==============================================================================

output "alb_dns_name" {
  description = "Application Load Balancer DNS name"
  value       = module.alb.alb_dns_name
}

output "alb_arn" {
  description = "Application Load Balancer ARN"
  value       = module.alb.alb_arn
}

output "alb_zone_id" {
  description = "Application Load Balancer Zone ID"
  value       = module.alb.alb_zone_id
}

# ==============================================================================
# Route53 Outputs
# ==============================================================================

output "route53_zone_id" {
  description = "Route53 hosted zone ID"
  value       = module.route53.zone_id
}

output "route53_name_servers" {
  description = "Route53 name servers"
  value       = module.route53.name_servers
}

# ==============================================================================
# Access URLs
# ==============================================================================

output "api_gateway_url" {
  description = "API Gateway URL"
  value       = "https://api.${var.domain_name}"
}

output "grafana_url" {
  description = "Grafana dashboard URL"
  value       = "https://grafana.${var.domain_name}"
}

output "jaeger_url" {
  description = "Jaeger tracing URL"
  value       = "https://jaeger.${var.domain_name}"
}

# ==============================================================================
# Deployment Information
# ==============================================================================

output "deployment_info" {
  description = "Complete deployment information"
  value = {
    region           = var.aws_region
    environment      = var.environment
    cluster_name     = module.eks.cluster_name
    vpc_id           = module.vpc.vpc_id
    database_endpoint = module.rds.db_instance_endpoint
    redis_endpoint   = module.elasticache.cache_cluster_address
    kafka_brokers    = module.msk.bootstrap_brokers_tls
    load_balancer    = module.alb.alb_dns_name
  }
  sensitive = true
}

# ==============================================================================
# Next Steps Instructions
# ==============================================================================

output "next_steps" {
  description = "Instructions for next steps"
  value = <<-EOT
    
    ═══════════════════════════════════════════════════════════════
    INFRASTRUCTURE CREATED SUCCESSFULLY! 🎉
    ═══════════════════════════════════════════════════════════════
    
    1. Configure kubectl:
       ${module.eks.configure_kubectl_command}
    
    2. Deploy Istio:
       cd helm/istio && ./install-istio.sh
    
    3. Deploy application:
       cd helm && ./install.sh ${var.environment}
    
    4. Access services:
       API Gateway: https://api.${var.domain_name}
       Grafana:     https://grafana.${var.domain_name}
       Jaeger:      https://jaeger.${var.domain_name}
    
    5. Update DNS nameservers:
       ${join("\n       ", module.route53.name_servers)}
    
    ═══════════════════════════════════════════════════════════════
    
  EOT
}

