locals {
  ami = "ami-830c94e3"
}

resource "aws_instance" "app_server_a" {
  ami           = local.ami
  instance_type = "t2.micro"
}

resource "aws_instance" "app_server_b" {
  ami           = local.ami
  instance_type = "t2.small"
}
