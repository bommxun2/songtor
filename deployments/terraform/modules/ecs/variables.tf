variable "ecs_cluster_name" { type = string }
variable "ecs_service_name" { type = string }
variable "container_image" { type = string }
variable "ecs_subnet_id" { type = string }
variable "ecs_security_group_id" { type = string }

variable "target_group_arn" {
  type    = string
}

variable "environment_variables" {
  type = list(object({
    name  = string
    value = string
  }))
  default = []
}
