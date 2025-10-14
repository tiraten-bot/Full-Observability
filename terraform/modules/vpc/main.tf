# ==============================================================================
# VPC MODULE
# ==============================================================================
# Purpose: Create VPC with public/private subnets across multiple AZs
# 
# Architecture:
#   - 3 Availability Zones for high availability
#   - Private subnets: EKS worker nodes, application pods
#   - Public subnets: NAT Gateways, Load Balancers
#   - Database subnets: RDS, ElastiCache (isolated)
#   - Internet Gateway for public access
#   - NAT Gateways for private subnet internet access
# ==============================================================================

# ==============================================================================
# VPC
# ==============================================================================

resource "aws_vpc" "main" {
  cidr_block = var.vpc_cidr
  
  enable_dns_hostnames = var.enable_dns_hostnames
  enable_dns_support   = var.enable_dns_support
  
  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-vpc"
    }
  )
}

# ==============================================================================
# INTERNET GATEWAY
# ==============================================================================

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  
  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-igw"
    }
  )
}

# ==============================================================================
# PUBLIC SUBNETS
# ==============================================================================

resource "aws_subnet" "public" {
  count = length(var.public_subnets)
  
  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.public_subnets[count.index]
  availability_zone       = var.azs[count.index]
  map_public_ip_on_launch = true
  
  tags = merge(
    var.tags,
    {
      Name                                           = "${var.project_name}-${var.environment}-public-${var.azs[count.index]}"
      "kubernetes.io/role/elb"                       = "1"
      "kubernetes.io/cluster/${var.project_name}-${var.environment}-eks" = "shared"
    }
  )
}

# ==============================================================================
# PRIVATE SUBNETS
# ==============================================================================

resource "aws_subnet" "private" {
  count = length(var.private_subnets)
  
  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnets[count.index]
  availability_zone = var.azs[count.index]
  
  tags = merge(
    var.tags,
    {
      Name                                           = "${var.project_name}-${var.environment}-private-${var.azs[count.index]}"
      "kubernetes.io/role/internal-elb"              = "1"
      "kubernetes.io/cluster/${var.project_name}-${var.environment}-eks" = "shared"
    }
  )
}

# ==============================================================================
# DATABASE SUBNETS
# ==============================================================================

resource "aws_subnet" "database" {
  count = length(var.database_subnets)
  
  vpc_id            = aws_vpc.main.id
  cidr_block        = var.database_subnets[count.index]
  availability_zone = var.azs[count.index]
  
  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-database-${var.azs[count.index]}"
      Type = "database"
    }
  )
}

# ==============================================================================
# ELASTIC IPs FOR NAT GATEWAYS
# ==============================================================================

resource "aws_eip" "nat" {
  count  = var.enable_nat_gateway ? length(var.azs) : 0
  domain = "vpc"
  
  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-eip-${var.azs[count.index]}"
    }
  )
  
  depends_on = [aws_internet_gateway.main]
}

# ==============================================================================
# NAT GATEWAYS
# ==============================================================================

resource "aws_nat_gateway" "main" {
  count = var.enable_nat_gateway ? length(var.azs) : 0
  
  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id
  
  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-nat-${var.azs[count.index]}"
    }
  )
  
  depends_on = [aws_internet_gateway.main]
}

# ==============================================================================
# ROUTE TABLES - PUBLIC
# ==============================================================================

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
  
  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-public-rt"
    }
  )
}

resource "aws_route_table_association" "public" {
  count = length(var.public_subnets)
  
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# ==============================================================================
# ROUTE TABLES - PRIVATE
# ==============================================================================

resource "aws_route_table" "private" {
  count = length(var.private_subnets)
  
  vpc_id = aws_vpc.main.id
  
  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main[count.index].id
  }
  
  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-private-rt-${var.azs[count.index]}"
    }
  )
}

resource "aws_route_table_association" "private" {
  count = length(var.private_subnets)
  
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}

# ==============================================================================
# ROUTE TABLES - DATABASE
# ==============================================================================

resource "aws_route_table" "database" {
  vpc_id = aws_vpc.main.id
  
  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-database-rt"
    }
  )
}

resource "aws_route_table_association" "database" {
  count = length(var.database_subnets)
  
  subnet_id      = aws_subnet.database[count.index].id
  route_table_id = aws_route_table.database.id
}

# ==============================================================================
# DB SUBNET GROUP
# ==============================================================================

resource "aws_db_subnet_group" "database" {
  name       = "${var.project_name}-${var.environment}-db-subnet-group"
  subnet_ids = aws_subnet.database[*].id
  
  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-db-subnet-group"
    }
  )
}

# ==============================================================================
# ELASTICACHE SUBNET GROUP
# ==============================================================================

resource "aws_elasticache_subnet_group" "redis" {
  name       = "${var.project_name}-${var.environment}-redis-subnet-group"
  subnet_ids = aws_subnet.database[*].id
  
  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-redis-subnet-group"
    }
  )
}

# ==============================================================================
# VPC FLOW LOGS (for network monitoring)
# ==============================================================================

resource "aws_flow_log" "main" {
  count = var.enable_flow_logs ? 1 : 0
  
  vpc_id          = aws_vpc.main.id
  traffic_type    = "ALL"
  iam_role_arn    = aws_iam_role.flow_logs[0].arn
  log_destination = aws_cloudwatch_log_group.flow_logs[0].arn
  
  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-flow-logs"
    }
  )
}

resource "aws_cloudwatch_log_group" "flow_logs" {
  count = var.enable_flow_logs ? 1 : 0
  
  name              = "/aws/vpc/${var.project_name}-${var.environment}-flow-logs"
  retention_in_days = 30
  
  tags = var.tags
}

resource "aws_iam_role" "flow_logs" {
  count = var.enable_flow_logs ? 1 : 0
  
  name = "${var.project_name}-${var.environment}-flow-logs-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "vpc-flow-logs.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "flow_logs" {
  count = var.enable_flow_logs ? 1 : 0
  
  name = "${var.project_name}-${var.environment}-flow-logs-policy"
  role = aws_iam_role.flow_logs[0].id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams"
      ]
      Resource = "*"
    }]
  })
}

