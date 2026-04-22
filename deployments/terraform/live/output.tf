output "rds_endpoint" {
  value = module.rds.rds_endpoint
}

output "sns_topic_arn" {
  value = module.sns.sns_topic_arn
}

output "api_gateway_endpoint" {
  value = module.apigateway.api_endpoint
}