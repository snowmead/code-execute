terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0.1"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0.1"
    }
    google = {
      source  = "hashicorp/google"
      version = "4.15.0"
    }
  }

  backend "gcs" {
    bucket = "ce-terraform-state-helm"
    prefix = "terraform/state"
  }

  required_version = "~> 0.14"
}

provider "google" {
  project = var.project_id
  region  = var.region
}