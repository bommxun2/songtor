output "rds_endpoint" {
  value = module.rds.rds_endpoint
}

output "sns_topic_arn" {
  value = module.sns.sns_topic_arn
}