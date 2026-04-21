data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

resource "aws_lambda_function" "this" {
  function_name    = var.function_name
  handler          = var.handler
  runtime          = var.runtime
  filename         = var.filename
  source_code_hash = filebase64sha256(var.filename)
  role             = data.aws_iam_role.lab_role.arn
  timeout          = 30

  environment {
    variables = var.environment_variables
  }
}

resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  count            = var.enable_sqs_trigger ? 1 : 0
  event_source_arn = var.sqs_trigger_arn
  function_name    = aws_lambda_function.this.arn
  batch_size       = 10
}
