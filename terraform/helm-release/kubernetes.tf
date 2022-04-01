data "terraform_remote_state" "gke" {
  backend = "gcs"
  config = {
    bucket = "codeexecute-terraform-state-gke"
    prefix = "terraform/state"
  }
}

# Retrieve GKE cluster configuration
data "google_container_cluster" "cluster" {
  name     = data.terraform_remote_state.gke.outputs.kubernetes_cluster_name
  location = data.terraform_remote_state.gke.outputs.kubernetes_cluster_location
}

module "gke_auth" {
  source       = "terraform-google-modules/kubernetes-engine/google//modules/auth"
  project_id   = data.terraform_remote_state.gke.outputs.project_id
  location     = data.google_container_cluster.cluster.location
  cluster_name = data.google_container_cluster.cluster.name
}
