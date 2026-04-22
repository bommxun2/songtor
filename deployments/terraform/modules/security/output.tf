output "vpc_id" {
  value = aws_vpc.vpc.id
}

output "subnet_id" {
  value = aws_subnet.public_subnet_1.id
}

output "subnet_ids" {
  value = [aws_subnet.public_subnet_1.id, aws_subnet.public_subnet_2.id]
}

output "db_subnet_group_name" {
  value = aws_db_subnet_group.db_group.name
}

output "db_security_group_id" {
  value = aws_security_group.rds.id
}

output "ecs_security_group_id" {
  value = aws_security_group.ecs.id
}