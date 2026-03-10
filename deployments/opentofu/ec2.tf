data "aws_ami" "amazon_linux_2023" {
  most_recent = true
  owners      = ["amazon"]
  filter {
    name   = "name"
    values = ["al2023-ami-2023.*-x86_64"]
  }
}

# Generate a new SSH key pair for EC2 access
resource "tls_private_key" "ec2_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# Create an AWS Key Pair using the generated public key
resource "aws_key_pair" "deployer" {
  key_name   = "songtor-key"
  public_key = tls_private_key.ec2_key.public_key_openssh
}

# Save the private key to a local file with appropriate permissions
resource "local_file" "ssh_key" {
  filename        = "${path.module}/songtor-key.pem"
  content         = tls_private_key.ec2_key.private_key_pem
  file_permission = "0700"
}

resource "aws_instance" "app_server" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = "t3.micro"
  subnet_id              = aws_subnet.public_subnet.id
  vpc_security_group_ids = [aws_security_group.ec2_sg.id]

  key_name = aws_key_pair.deployer.key_name

  tags = { Name = "Songtor-App-Server" }
}

output "ec2_public_ip" {
  description = "Public IP of the EC2 Instance"
  value       = aws_instance.app_server.public_ip
}