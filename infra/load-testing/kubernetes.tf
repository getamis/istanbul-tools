resource "kubernetes_service" "validator-svc" {
  metadata {
    name = "validator-svc-${count.index}"
  }

  spec {
    selector {
      app = "validator-${count.index}"
    }
    type = "LoadBalancer"
    port {
      port = 8546
      target_port = 8546
    }

    type = "LoadBalancer"
  }

  count = "${length(var.svcs)}"
}