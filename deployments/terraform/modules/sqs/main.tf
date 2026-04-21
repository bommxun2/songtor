resource "aws_sqs_queue" "request_queue" {
  name = var.request_queue_name
}

resource "aws_sqs_queue" "reply_queue" {
  name = var.reply_queue_name
}
