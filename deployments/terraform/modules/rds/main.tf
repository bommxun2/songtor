resource "aws_db_instance" "this" {
  identifier           = "go-app-db"
  allocated_storage    = 10
  db_name              = var.db_name
  engine               = var.engine
  engine_version       = var.engine_version
  instance_class       = var.instance_class
  username             = var.username
  password             = var.password
  parameter_group_name = var.parameter_group_name

  db_subnet_group_name   = var.db_subnet_group_name
  vpc_security_group_ids = [var.db_security_group_ids]
  
  publicly_accessible    = true
  skip_final_snapshot    = true
}
