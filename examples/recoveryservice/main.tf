resource "azurerm_resource_group" "rg" {
  name     = "${var.resource_group_name}"
  location = "${var.resource_group_location}"
}

resource "random_integer" "ri" {
  min = 100
  max = 999
}

module "vm" {
  source = "modules/vm"

  resource_group_name = "${azurerm_resource_group.rg.name}"
  prefix          = "tfexrecove${random_integer.ri.result}"
  hostname        = "tfexrecove${random_integer.ri.result}"
  dns_name        = "tfexrecove${random_integer.ri.result}"
  admin_username  = "vmadmin"
  admin_password  = "Password123!@#"
}

resource "azurerm_recovery_services_vault" "vault" {
  name                = "example-recovery-vault"
  location            = "${azurerm_resource_group.rg.location}"
  resource_group_name = "${azurerm_resource_group.rg.name}"
  sku                 = "Standard"
}

