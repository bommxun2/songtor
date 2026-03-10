# Create HTTP API Gateway
resource "aws_api_gateway_rest_api" "go_app_api" {
  name          = "songtor-http-api"
  description   = "API Gateway routing traffic to Go App on EC2"
}

# Create Resource for /v1
resource "aws_api_gateway_resource" "v1" {
  rest_api_id = aws_api_gateway_rest_api.go_app_api.id
  parent_id   = aws_api_gateway_rest_api.go_app_api.root_resource_id
  path_part   = "v1"
}

# Create Route
resource "aws_api_gateway_resource" "proxy" {
  rest_api_id = aws_api_gateway_rest_api.go_app_api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "{proxy+}" 
}

resource "aws_api_gateway_method" "proxy_method" {
  rest_api_id   = aws_api_gateway_rest_api.go_app_api.id
  resource_id   = aws_api_gateway_resource.proxy.id
  http_method   = "ANY"
  authorization = "NONE"

  request_parameters = {
    "method.request.path.proxy" = true
  }
}

resource "aws_api_gateway_integration" "ec2_proxy" {
  rest_api_id             = aws_api_gateway_rest_api.go_app_api.id
  resource_id             = aws_api_gateway_resource.proxy.id
  http_method             = aws_api_gateway_method.proxy_method.http_method
  
  type                    = "HTTP_PROXY"
  integration_http_method = "ANY"
  
  uri = "http://${aws_instance.app_server.public_ip}:8080/v1/{proxy}"

  request_parameters = {
    "integration.request.path.proxy" = "method.request.path.proxy"
  }
}

resource "aws_api_gateway_deployment" "deployment" {
  depends_on  = [aws_api_gateway_integration.ec2_proxy]
  rest_api_id = aws_api_gateway_rest_api.go_app_api.id
}

resource "aws_api_gateway_stage" "prod" {
  deployment_id = aws_api_gateway_deployment.deployment.id
  rest_api_id   = aws_api_gateway_rest_api.go_app_api.id
  stage_name    = "prod"
}

output "rest_api_url" {
  value = "${aws_api_gateway_stage.prod.invoke_url}"
}