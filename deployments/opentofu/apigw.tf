# Create HTTP API Gateway
resource "aws_apigatewayv2_api" "go_app_api" {
  name          = "songtor-http-api"
  protocol_type = "HTTP"
  description   = "API Gateway routing traffic to Go App on EC2"
}

# Create Integration
resource "aws_apigatewayv2_integration" "ec2_integration" {
  api_id             = aws_apigatewayv2_api.go_app_api.id
  integration_type   = "HTTP_PROXY"
  
  # Pointing to EC2 instance's public IP and port 8080 where Go App is running
  integration_uri    = "http://${aws_instance.app_server.public_ip}:8080"
  
  # HTTp method
  integration_method = "POST" 
}

# Create Route
resource "aws_apigatewayv2_route" "post_notification_route" {
  api_id    = aws_apigatewayv2_api.go_app_api.id
  
  # Define the route key (HTTP method + path) and link it to the integration
  route_key = "POST /v1/notifications"
  target    = "integrations/${aws_apigatewayv2_integration.ec2_integration.id}"
}

# Create Stage
resource "aws_apigatewayv2_stage" "default_stage" {
  api_id      = aws_apigatewayv2_api.go_app_api.id
  name        = "$default"
  auto_deploy = true
}

# Output the API endpoint URL
output "api_endpoint_post_notification" {
  description = "URL for POST /v1/notifications"
  value       = "${aws_apigatewayv2_stage.default_stage.invoke_url}v1/notifications"
}