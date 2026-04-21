variable "function_name" {
  type = string
}

variable "handler" {
  type = string
}

variable "runtime" {
  type = string
}

variable "filename" {
  type = string
}

variable "environment_variables" {
  type    = map(string)
  default = {}
}

variable "sqs_trigger_arn" {
  type    = string
  default = ""
}

variable "enable_sqs_trigger" {
  type    = bool
  default = false
}
