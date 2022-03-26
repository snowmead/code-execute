# Retrieve an access token as the Terraform runner
data "google_client_config" "provider" {}

provider "helm" {
  kubernetes {
    host  = "https://${data.google_container_cluster.cluster.endpoint}"
    token = data.google_client_config.provider.access_token
    cluster_ca_certificate = base64decode(
      data.google_container_cluster.cluster.master_auth[0].cluster_ca_certificate,
    )
  }
}

resource "helm_release" "codeexecute" {
  name  = "ce-chart"
  chart = "../../chart/"

  set {
    name  = "image.tag"
    value = var.image_tag
  }

  set_sensitive {
    name  = "bot.token"
    value = base64encode(var.bot_token)
  }
}
