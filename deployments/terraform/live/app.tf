terraform {
  required_providers {
    aws = {
      source = "opentofu/aws"
      version = "6.41.0"
    }
  }
  backend "s3" {
    bucket = "songtor-terraform-state"
    key    = "terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = "us-east-1"
}

module "security_group" {
  source = "../modules/security"

  vpc_cidr = "10.0.0.0/16"
  public_subnet_1_cidr = "10.0.1.0/24"
  public_subnet_2_cidr = "10.0.2.0/24"
  ecs_security_group_name = "ecs-sg"
  rds_security_group_name = "rds-sg"
  container_port = 8080
  rds_port = 5432
  aws_region = "us-east-1"
}

module "ecr" {
  source = "../modules/ecr"

  ecr_repository_name = "go-app"
}

module "sns" {
  source = "../modules/sns"

  topic_name = "patient-topic"
}

module "ecs" {
  source = "../modules/ecs"

  ecs_cluster_name = "go-app-cluster"
  ecs_service_name = "go-app-service"
  container_image = "${module.ecr.ecr_repository_url}:latest"
  ecs_subnet_id = module.security_group.subnet_id
  ecs_security_group_id = module.security_group.ecs_security_group_id

  // In a production environment, you should use AWS Secrets Manager or AWS Systems Manager Parameter Store to manage sensitive information like database credentials.
  environment_variables = [
    { name = "DB_USER",       value = "notification_admin" },
    { name = "DB_PASSWORD",   value = "password" },
    { name = "DB_HOST",       value = module.rds.rds_address },
    { name = "DB_PORT",       value = "5432" },
    { name = "DB_NAME",       value = "goappdb" },
    { name = "SNS_TOPIC_ARN", value = module.sns.sns_topic_arn }
  ]
}

module "rds" {
  source = "../modules/rds"

  db_name = "goappdb"
  engine = "postgres"
  engine_version = "17.9"
  instance_class = "db.t3.micro"
  username = "notification_admin"
  password = "password"
  parameter_group_name = "default.postgres17"
  db_subnet_group_name = module.security_group.db_subnet_group_name
  db_security_group_ids = module.security_group.db_security_group_id
}

module "sqs" {
  source = "../modules/sqs"

  request_queue_name = "inbound-load-req"
  reply_queue_name   = "patient-reported-reply"
}

module "inbound_load_lambda" {
  source = "../modules/lambda"

  function_name = "inbound-load-function"
  handler       = "main"
  runtime       = "provided.al2"
  filename      = "inbound_load.zip"

  enable_sqs_trigger = true
  sqs_trigger_arn    = module.sqs.request_queue_arn

  environment_variables = {
    // In a production environment, you should use AWS Secrets Manager or AWS Systems Manager Parameter Store to manage sensitive information like database credentials.
    DB_DSN = "host=${module.rds.rds_address} user=notification_admin password=password dbname=goappdb port=5432 sslmode=require"
  }
}