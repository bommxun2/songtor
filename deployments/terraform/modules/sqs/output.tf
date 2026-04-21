output "request_queue_url" {
  value = aws_sqs_queue.request_queue.url
}

output "request_queue_arn" {
  value = aws_sqs_queue.request_queue.arn
}