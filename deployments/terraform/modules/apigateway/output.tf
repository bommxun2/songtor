output "api_endpoint" {
  value = aws_apigatewayv2_api.this.api_endpoint
}

output "target_group_arn" {
  value = aws_lb_target_group.this.arn
}
