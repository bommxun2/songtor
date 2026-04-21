variable "vpc_cidr" {
  type = string
}

variable "public_subnet_1_cidr" {
  type = string
}

variable "public_subnet_2_cidr" {
  type = string
}

variable "ecs_security_group_name" {
  type = string
}

variable "container_port" {
  type = number
}

variable "rds_security_group_name" {
  type = string
}

variable "rds_port" {
  type = number
}

variable "aws_region" {
  type = string
}