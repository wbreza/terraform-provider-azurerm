---
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_cdn_custom_domain"
sidebar_current: "docs-azurerm-resource-cdn-custom-domain"
description: |-
  Manages a Custom Domain within a CDN Endpoint.

---

# azurerm_cdn_custom_domain

Manages a Custom Domain within a CDN Endpoint.

## Example Usage

```hcl
resource "random_id" "server" {
  keepers = {
    azi_id = 1
  }

  byte_length = 8
}

resource "azurerm_resource_group" "test" {
  name     = "acceptanceTestResourceGroup1"
  location = "West US"
}

resource "azurerm_cdn_profile" "test" {
  name                = "exampleCdnProfile"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard_Verizon"
}

resource "azurerm_cdn_endpoint" "test" {
  name                = "${random_id.server.hex}"
  profile_name        = "${azurerm_cdn_profile.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  origin {
    name      = "exampleCdnOrigin"
    host_name = "www.example.com"
  }
}

resource "azurerm_cdn_custom_domain" "test" {
  domain_name         = "mydomain.com"
  endpoint_name       = "${azurerm_cdn_endpoint.test.name}"
  profile_name        = "${azurerm_cdn_profile.test.name}"
  resource_group_name = "${azurerm_resource_group.test.name}"
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) Specifies the top-level Domain Name to be attached to the CDN Endpoint. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which the CDN Endpoint exists.

* `profile_name` - (Required) The name of the CDN Profile where the CDN Endpoint exists.

* `profile_name` - (Required) The name of the CDN Endpoint within the CDN Profile.

## Attributes Reference

The following attributes are exported:

* `id` - The CDN Custom Domain ID.

## Import

CDN Custom Domains can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_cdn_custom_domain.myCustomDomain /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.Cdn/profiles/myprofile1/endpoints/myendpoint1/customDomains/myCustomDomain.com
```
