# ==============================================================================
# RDS POSTGRESQL MODULE
# ==============================================================================

resource "aws_db_instance" "main" {
  identifier = var.identifier
  
  # Engine
  engine         = var.engine
  engine_version = var.engine_version
  instance_class = var.instance_class
  
  # Storage
  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  storage_type          = "gp3"
  storage_encrypted     = var.storage_encrypted
  kms_key_id            = var.kms_key_id
  
  # Database
  db_name  = var.db_name
  username = var.username
  password = var.password
  port     = var.port
  
  # Network
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = var.vpc_security_group_ids
  publicly_accessible    = false
  
  # High Availability
  multi_az = var.multi_az
  
  # Backup
  backup_retention_period = var.backup_retention_period
  backup_window           = var.backup_window
  skip_final_snapshot     = var.environment != "prod"
  final_snapshot_identifier = var.environment == "prod" ? "${var.identifier}-final-snapshot-${formatdate("YYYY-MM-DD-hhmm", timestamp())}" : null
  
  # Maintenance
  maintenance_window              = var.maintenance_window
  auto_minor_version_upgrade      = true
  enabled_cloudwatch_logs_exports = var.enabled_cloudwatch_logs_exports
  
  # Performance Insights
  performance_insights_enabled    = var.environment == "prod"
  performance_insights_kms_key_id = var.kms_key_id
  performance_insights_retention_period = 7
  
  # Deletion protection
  deletion_protection = var.environment == "prod"
  
  tags = var.tags
}

resource "aws_db_subnet_group" "main" {
  name       = "${var.identifier}-subnet-group"
  subnet_ids = var.subnet_ids
  
  tags = var.tags
}

