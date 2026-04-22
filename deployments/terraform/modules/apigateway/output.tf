output "api_endpoint" {
  value = aws_apigatewayv2_api.this.api_endpoint
}

output "service_discovery_arn" {
  value = aws_service_discovery_service.this.arn
}
