# Create a Dead Letter Queue (DLQ) for the request queue
resource "aws_sqs_queue" "request_dlq" {
  name                      = "ambulance-divert-request-dlq"
  message_retention_seconds = 1209600
}

# Main queue configuration with redrive policy to send messages to DLQ after 3 failed processing attempts
resource "aws_sqs_queue" "request_queue" {
  name                      = "ambulance-divert-request-queue"
  message_retention_seconds = 86400
  receive_wait_time_seconds = 20
  visibility_timeout_seconds = 60
  
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.request_dlq.arn
    maxReceiveCount     = 3
  })
}

resource "aws_sqs_queue" "reply_queue" {
  name = "ambulance-divert-reply-queue"
}

# Allow the request queue to send messages to the DLQ
resource "aws_sqs_queue_redrive_allow_policy" "request_dlq_allow" {
  queue_url = aws_sqs_queue.request_dlq.id

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue",
    sourceQueueArns   = [aws_sqs_queue.request_queue.arn]
  })
}