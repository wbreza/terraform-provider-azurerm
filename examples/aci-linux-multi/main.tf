resource "azurerm_resource_group" "aci-rg" {
    name     = "${var.resource_group_name}"
    location = "${var.resource_group_location}"
}

#an attempt to keep the aci container group name (and dns label) somewhat unique
resource "random_integer" "random_int" {
    min = 100
    max = 999
}

resource "azurerm_container_group" "aci-example" {
    name                = "my-aci-cg-${random_integer.random_int.result}"
    location            = "${azurerm_resource_group.aci-rg.location}"
    resource_group_name = "${azurerm_resource_group.aci-rg.name}"
    ip_address_type     = "public"
    dns_name_label      = "my-aci-cg-${random_integer.random_int.result}"
    os_type             = "linux"
    protocol            = "tcp"

    container {
        name     = "hw"
        image    = "microsoft/aci-helloworld:latest"
        cpu      = "0.5"
        memory   = "1.5"
        ports    = ["80", "81"]
    }

    container {
        name    = "sidecar"
        image   = "microsoft/aci-tutorial-sidecar"
        cpu     = "0.5"
        memory  = "1.5"
        ports   = ["8080"]
    }

    tags {
        environment = "testing"
    }
}
