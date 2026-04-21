output "sns_topic_name" {
  value = aws_sns_topic.user_updates.name
}

output "sns_topic_arn" {
  value = aws_sns_topic.user_updates.arn
}