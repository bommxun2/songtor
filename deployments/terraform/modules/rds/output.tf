output "rds_endpoint" {
  value = aws_db_instance.this.endpoint
}

output "rds_address" {
  value = aws_db_instance.this.address
}