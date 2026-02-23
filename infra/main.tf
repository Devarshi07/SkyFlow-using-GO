terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
  required_version = ">= 1.0"
}

provider "azurerm" {
  features {}
}

# ── Resource Group ──
resource "azurerm_resource_group" "skyflow" {
  name     = "skyflow-rg"
  location = "East US"
}

# ── Log Analytics (required by Container Apps) ──
resource "azurerm_log_analytics_workspace" "skyflow" {
  name                = "skyflow-logs"
  resource_group_name = azurerm_resource_group.skyflow.name
  location            = azurerm_resource_group.skyflow.location
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

# ── Container App Environment ──
resource "azurerm_container_app_environment" "skyflow" {
  name                       = "skyflow-env"
  resource_group_name        = azurerm_resource_group.skyflow.name
  location                   = azurerm_resource_group.skyflow.location
  log_analytics_workspace_id = azurerm_log_analytics_workspace.skyflow.id
}

# ── Container App (Free: 180K vCPU-sec + 360K GiB-sec/month) ──
resource "azurerm_container_app" "skyflow_api" {
  name                         = "skyflow-api"
  container_app_environment_id = azurerm_container_app_environment.skyflow.id
  resource_group_name          = azurerm_resource_group.skyflow.name
  revision_mode                = "Single"

  ingress {
    external_enabled = true
    target_port      = 8080

    traffic_weight {
      latest_revision = true
      percentage      = 100
    }
  }

  template {
    min_replicas = 0
    max_replicas = 1

    container {
      name   = "skyflow-api"
      image  = "ghcr.io/${var.github_username}/skyflow-backend:latest"
      cpu    = 0.25
      memory = "0.5Gi"

      env {
        name  = "PORT"
        value = "8080"
      }
      env {
        name        = "DATABASE_URL"
        secret_name = "database-url"
      }
      env {
        name        = "REDIS_URL"
        secret_name = "redis-url"
      }
      env {
        name        = "RABBITMQ_URL"
        secret_name = "rabbitmq-url"
      }
      env {
        name  = "FRONTEND_URL"
        value = var.frontend_url
      }
      env {
        name        = "STRIPE_SECRET_KEY"
        secret_name = "stripe-secret-key"
      }
      env {
        name  = "GOOGLE_CLIENT_ID"
        value = var.google_client_id
      }
      env {
        name        = "GOOGLE_CLIENT_SECRET"
        secret_name = "google-client-secret"
      }
      env {
        name  = "SMTP_HOST"
        value = "smtp.gmail.com"
      }
      env {
        name  = "SMTP_PORT"
        value = "587"
      }
      env {
        name  = "SMTP_FROM"
        value = var.smtp_from
      }
      env {
        name        = "SMTP_PASSWORD"
        secret_name = "smtp-password"
      }
    }
  }

  secret {
    name  = "database-url"
    value = var.database_url
  }
  secret {
    name  = "redis-url"
    value = var.redis_url
  }
  secret {
    name  = "rabbitmq-url"
    value = var.rabbitmq_url
  }
  secret {
    name  = "stripe-secret-key"
    value = var.stripe_secret_key
  }
  secret {
    name  = "google-client-secret"
    value = var.google_client_secret
  }
  secret {
    name  = "smtp-password"
    value = var.smtp_password
  }
}

# ── Random suffix ──
resource "random_string" "suffix" {
  length  = 6
  special = false
  upper   = false
}

# ── Outputs ──
output "backend_url" {
  value = "https://${azurerm_container_app.skyflow_api.ingress[0].fqdn}"
}
