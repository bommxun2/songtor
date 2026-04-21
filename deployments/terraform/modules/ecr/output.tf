output "ecr_repository_url" {
  value = aws_ecr_repository.app.repository_url
  description = "URL of the ECR repository for the application"
}