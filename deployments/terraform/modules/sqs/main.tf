resource "aws_sqs_queue" "request_queue" {
  name = var.request_queue_name
}

resource "aws_sqs_queue" "reply_queue" {
  name = var.reply_queue_name
}

resource "aws_sns_topic_subscription" "reply_queue_subscription" {
  count     = var.sns_topic_arn != null ? 1 : 0
  topic_arn = var.sns_topic_arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.reply_queue.arn
}

resource "aws_sqs_queue_policy" "reply_queue_policy" {
  count     = var.sns_topic_arn != null ? 1 : 0
  queue_url = aws_sqs_queue.reply_queue.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action   = "sqs:SendMessage"
        Resource = aws_sqs_queue.reply_queue.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = var.sns_topic_arn
          }
        }
      }
    ]
  })
}
