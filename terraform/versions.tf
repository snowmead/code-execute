terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.15.0"
    }
  }

  required_version = ">= 0.14"
}

provider "google" {
  project     = var.project_id
  region      = var.region
}