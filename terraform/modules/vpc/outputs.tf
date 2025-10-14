output "vpc_id" {
  value = aws_vpc.main.id
}

output "vpc_cidr" {
  value = aws_vpc.main.cidr_block
}

output "private_subnets" {
  value = aws_subnet.private[*].id
}

output "public_subnets" {
  value = aws_subnet.public[*].id
}

output "database_subnets" {
  value = aws_subnet.database[*].id
}

output "nat_gateway_ids" {
  value = aws_nat_gateway.main[*].id
}

output "internet_gateway_id" {
  value = aws_internet_gateway.main.id
}

output "db_subnet_group_name" {
  value = aws_db_subnet_group.database.name
}

output "elasticache_subnet_group_name" {
  value = aws_elasticache_subnet_group.redis.name
}

