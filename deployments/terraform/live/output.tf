output "rds_endpoint" {
  value = module.rds.rds_endpoint
}

output "sns_patient_reported_topic_arn" {
  value = module.patient_reported_sns.sns_topic_arn
}

output "sns_critical_case_topic_arn" {
  value = module.critical_case_sns.sns_topic_arn
}

output "api_gateway_endpoint" {
  value = module.apigateway.api_endpoint
}