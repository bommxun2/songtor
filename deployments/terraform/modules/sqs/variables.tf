variable "request_queue_name" {
  type        = string
}

variable "reply_queue_name" {
  type        = string
}

variable "sns_topic_arn" {
  type        = string
  default     = null
}
