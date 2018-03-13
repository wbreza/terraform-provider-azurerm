package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureRMCdnCustomDomain_basic(t *testing.T) {
	resourceName := "azurerm_cdn_custom_domain.test"
	ri := acctest.RandInt()
	config := testAccAzureRMCdnCustomDomain_basic(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMCdnCustomDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMCdnCustomDomainExists(resourceName),
				),
			},
		},
	})
}

func testCheckAzureRMCdnCustomDomainExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["domain_name"]
		profileName := rs.Primary.Attributes["profile_name"]
		endpointName := rs.Primary.Attributes["endpoint_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for cdn custom domain: %s", name)
		}

		conn := testAccProvider.Meta().(*ArmClient).cdnCustomDomainsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := conn.Get(ctx, resourceGroup, profileName, endpointName, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on cdnCustomDomainsClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: CDN Custom Domain %q (resource group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureRMCdnCustomDomainDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["domain_name"]
		profileName := rs.Primary.Attributes["profile_name"]
		endpointName := rs.Primary.Attributes["endpoint_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for cdn custom domain: %s", name)
		}

		client := testAccProvider.Meta().(*ArmClient).cdnCustomDomainsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		future, err := client.Delete(ctx, resourceGroup, profileName, endpointName, name)
		if err != nil {
			return fmt.Errorf("Bad: Delete on cdnCustomDomainsClient: %+v", err)
		}

		err = future.WaitForCompletion(ctx, client.Client)
		if err != nil {
			return fmt.Errorf("Bad: Delete on cdnCustomDomainsClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureRMCdnCustomDomainDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).cdnCustomDomainsClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_cdn_custom_domain" {
			continue
		}

		name := rs.Primary.Attributes["domain_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]
		profileName := rs.Primary.Attributes["profile_name"]
		endpointName := rs.Primary.Attributes["endpoint_name"]

		resp, err := conn.Get(ctx, resourceGroup, profileName, endpointName, name)
		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("CDN Custom Domain still exists:\n%#v", resp)
		}
	}

	return nil
}

func testAccAzureRMCdnCustomDomain_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_cdn_profile" "test" {
  name                = "acctestcdnprof%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard_Verizon"
}

resource "azurerm_cdn_endpoint" "test" {
  name                = "acctestcdnend%d"
  profile_name        = "${azurerm_cdn_profile.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  origin {
    name       = "acceptanceTestCdnOrigin1"
    host_name  = "www.example.com"
    https_port = 443
    http_port  = 80
  }
}

resource "azurerm_cdn_custom_domain" "test" {
  domain_name         = "hashicorptest.com"
  profile_name        = "${azurerm_cdn_profile.test.name}"
  endpoint_name       = "${azurerm_cdn_endpoint.test.name}"
  resource_group_name = "${azurerm_resource_group.test.name}"
}
`, rInt, location, rInt, rInt)
}
