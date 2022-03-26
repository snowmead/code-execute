variable "project_id" {
  description = "Project ID"
}

variable "region" {
  default = "us-east1"
}

variable "application_name" {
  type    = string
  default = "code-execute"
}

variable "image_tag" {
  description = "Docker Image Tag"
}

variable "bot_token" {
  type        = string
  description = "Discord Bot Token"
  sensitive   = true
}
