terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.15.0"
    }
  }

  backend "gcs" {
    bucket = "codeexecute-terraform-state-gke"
    prefix = "terraform/state"
  }

  required_version = ">= 0.14"
}

provider "google" {
  project = var.project_id
  region  = var.region
}
