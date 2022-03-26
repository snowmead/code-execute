data "terraform_remote_state" "gke" {
  backend = "remote"
  config = {
    bucket = "codeexecute-terraform-state"
    prefix = "terraform/state"
  }
}

# Retrieve GKE cluster configuration
data "google_container_cluster" "cluster" {
  name = data.terraform_remote_state.gke.outputs.kubernetes_cluster_name
}

module "gke_auth" {
  source       = "terraform-google-modules/kubernetes-engine/google//modules/auth"
  project_id   = var.project_id
  location     = google_container_cluster.cluster.region
  cluster_name = google_container_cluster.cluster.name
}
