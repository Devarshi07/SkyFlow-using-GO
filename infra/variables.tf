variable "location" {
  description = "Azure region (allowed: austriaeast, koreacentral, southeastasia, uaenorth, eastasia)"
  type        = string
  default     = "Southeast Asia"
}

variable "github_username" {
  description = "GitHub username for container registry"
  type        = string
  default     = "SomyaPadhy4501"
}

variable "database_url" {
  description = "Neon PostgreSQL connection string"
  type        = string
  sensitive   = true
}

variable "redis_url" {
  description = "Upstash Redis URL"
  type        = string
  sensitive   = true
}

variable "rabbitmq_url" {
  description = "CloudAMQP RabbitMQ URL"
  type        = string
  sensitive   = true
}

variable "frontend_url" {
  description = "Vercel frontend URL"
  type        = string
  default     = "https://sky-flow-using-go-nsn8.vercel.app"
}

variable "stripe_secret_key" {
  description = "Stripe secret key"
  type        = string
  sensitive   = true
}

variable "google_client_id" {
  description = "Google OAuth client ID"
  type        = string
}

variable "google_client_secret" {
  description = "Google OAuth client secret"
  type        = string
  sensitive   = true
}

variable "smtp_from" {
  description = "SMTP sender email"
  type        = string
  default     = "aryanpadhy40501@gmail.com"
}

variable "smtp_password" {
  description = "Gmail app password"
  type        = string
  sensitive   = true
}
