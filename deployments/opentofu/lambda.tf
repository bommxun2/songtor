# IAM Role and Policies
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# Package Lambda Function
data "archive_file" "lambda_zip" {
  type        = "zip"
  source_file = "${path.module}/bootstrap" 
  output_path = "${path.module}/lambda_deployment.zip"
}

# Lambda Function
resource "aws_lambda_function" "divert_lambda" {
  filename         = data.archive_file.lambda_zip.output_path
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256
  
  function_name = "ambulance-divert-processor"
  role          = data.aws_iam_role.lab_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  timeout = 15
  
  vpc_config {
    subnet_ids         = [aws_subnet.private_subnet.id]
    security_group_ids = [aws_security_group.lambda_sg.id]
  }

  environment {
    variables = {
      DB_USER     = "db_user"
      DB_PASSWORD = "db_password"
      DB_HOST     = aws_instance.app_server.private_ip
      DB_PORT     = "3306"
      DB_NAME     = "songtor_db"
    }
  }

}

# SQS Trigger
resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  event_source_arn = aws_sqs_queue.request_queue.arn
  function_name    = aws_lambda_function.divert_lambda.arn
  batch_size       = 10
}