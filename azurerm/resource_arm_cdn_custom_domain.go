package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/cdn/mgmt/2017-04-02/cdn"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmCdnCustomDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmCdnCustomDomainCreate,
		Read:   resourceArmCdnCustomDomainRead,
		Delete: resourceArmCdnCustomDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": resourceGroupNameSchema(),

			"profile_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"endpoint_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceArmCdnCustomDomainCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).cdnCustomDomainsClient
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for AzureRM CDN Custom Domains creation.")

	name := d.Get("domain_name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	profileName := d.Get("profile_name").(string)
	endpointName := d.Get("endpoint_name").(string)

	domain := cdn.CustomDomainParameters{
		CustomDomainPropertiesParameters: &cdn.CustomDomainPropertiesParameters{
			HostName: utils.String(name),
		},
	}

	future, err := client.Create(ctx, resourceGroup, profileName, endpointName, name, domain)
	if err != nil {
		return fmt.Errorf("Error creating CDN Custom Domain %q (Endpoint %q / Profile %q / Resource Group %q): %+v", name, endpointName, profileName, resourceGroup, err)
	}

	err = future.WaitForCompletion(ctx, client.Client)
	if err != nil {
		return fmt.Errorf("Error waiting for CDN Custom Domain %q (Endpoint %q / Profile %q / Resource Group %q) to finish creating: %+v", name, endpointName, profileName, resourceGroup, err)
	}

	read, err := client.Get(ctx, resourceGroup, profileName, endpointName, name)
	if err != nil {
		return fmt.Errorf("Error retrieving CDN Custom Domain %q (Endpoint %q / Profile %q / Resource Group %q): %+v", name, endpointName, profileName, resourceGroup, err)
	}

	d.SetId(*read.ID)

	return resourceArmCdnCustomDomainRead(d, meta)
}

func resourceArmCdnCustomDomainRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).cdnCustomDomainsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	name := id.Path["customDomains"]
	endpointName := id.Path["endpoints"]
	profileName := id.Path["profiles"]
	if profileName == "" {
		profileName = id.Path["Profiles"]
	}
	log.Printf("[INFO] Retrieving CDN Custom Domain %q (Endpoint %q / Profile %q / Resource Group %q)", name, endpointName, profileName, resourceGroup)
	resp, err := client.Get(ctx, resourceGroup, profileName, endpointName, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on Azure CDN Custom Domain %q (Endpoint %q / Profile %q / Resource Group %q): %+v", name, endpointName, profileName, resourceGroup, err)
	}

	d.Set("domain_name", resp.Name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("profile_name", profileName)
	d.Set("endpoint_name", endpointName)

	return nil
}

func resourceArmCdnCustomDomainDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).cdnCustomDomainsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	name := id.Path["customDomains"]
	endpointName := id.Path["endpoints"]
	profileName := id.Path["profiles"]
	if profileName == "" {
		profileName = id.Path["Profiles"]
	}

	future, err := client.Delete(ctx, resourceGroup, profileName, endpointName, name)
	if err != nil {
		if response.WasNotFound(future.Response()) {
			return nil
		}
		return fmt.Errorf("Error deleting CDN Custom Domain %q (Endpoint %q / Profile %q / Resource Group %q): %+v", name, endpointName, profileName, resourceGroup, err)
	}

	err = future.WaitForCompletion(ctx, client.Client)
	if err != nil {
		if response.WasNotFound(future.Response()) {
			return nil
		}
		return fmt.Errorf("Error waiting for CDN Custom Domain %q (Endpoint %q / Profile %q / Resource Group %q) to be deleted: %+v", name, endpointName, profileName, resourceGroup, err)
	}

	return nil
}
