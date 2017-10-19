provider "azurerm" {
  subscription_id = "${var.subscription_id}"
  client_id       = "${var.client_id}"
  client_secret   = "${var.client_secret}"
  tenant_id       = "${var.tenant_id}"
}

resource "random_id" "namer" {
  keepers = {
    resource_group    = "res"
    container_service = "cs"
    dns_prefix        = "k8s"
  }

  byte_length = 8
}

variable "username" {
  type    = "string"
  default = "amis"
}

resource "azurerm_resource_group" "test" {
  name     = "istanbul-test-${random_id.namer.hex}"
  location = "Southeast Asia"
}

resource "azurerm_container_service" "test" {
  name                   = "istanbul-test-${random_id.namer.hex}"
  location               = "${azurerm_resource_group.test.location}"
  resource_group_name    = "${azurerm_resource_group.test.name}"
  orchestration_platform = "Kubernetes"

  master_profile {
    count      = 1
    dns_prefix = "istanbul-${random_id.namer.hex}"
  }

  linux_profile {
    admin_username = "${var.username}"

    ssh_key {
      key_data = "${file("~/.ssh/id_rsa.pub")}"
    }
  }

  agent_pool_profile {
    name       = "default"
    count      = "${length(var.svcs)}"
    dns_prefix = "agent-${random_id.namer.hex}"
    vm_size    = "Standard_DS2_v2"
  }

  service_principal {
    client_id     = "${var.client_id}"
    client_secret = "${var.client_secret}"
  }

  diagnostics_profile {
    enabled = false
  }

  tags {
    Environment = "Testing"
  }
}

resource "null_resource" "kubeconfig" {
  provisioner "local-exec" {
    command     = "scp -o StrictHostKeyChecking=no ${var.username}@istanbul-${random_id.namer.hex}.${azurerm_resource_group.test.location}.cloudapp.azure.com:~/.kube/config ~/.kube/config"
    interpreter = ["bash", "-c"]
  }

  depends_on = ["azurerm_container_service.test"]
}
