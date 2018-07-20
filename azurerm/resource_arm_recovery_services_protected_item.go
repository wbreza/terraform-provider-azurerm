package azurerm

import (
	"fmt"
	"log"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/services/recoveryservices/mgmt/2016-06-01/backup"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"time"
)

func resourceArmRecoveryServicesProtectedItem() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmRecoveryServicesProtectedItemCreateUpdate,
		Read:   resourceArmRecoveryServicesProtectedItemRead,
		Update: resourceArmRecoveryServicesProtectedItemCreateUpdate,
		Delete: resourceArmRecoveryServicesProtectedItemDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				/*ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[a-zA-Z][-a-zA-Z0-9]{1,49}$"),
					"Recovery Service Vault name must be 2 - 50 characters long, start with a letter, contain only letters, numbers and hyphens.",
				),*/
			},

			"location": locationSchema(),

			"resource_group_name": resourceGroupNameSchema(),

			"recovery_vault_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[a-zA-Z][-a-zA-Z0-9]{1,49}$"),
					"Recovery Service Vault name must be 2 - 50 characters long, start with a letter, contain only letters, numbers and hyphens.",
				),
			},

			"fabric_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"container_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(backup.ProtectableItemTypeIaaSVMProtectableItem),
					string(backup.ProtectableItemTypeMicrosoftClassicComputevirtualMachines),
					string(backup.ProtectableItemTypeMicrosoftComputevirtualMachines),
					string(backup.ProtectableItemTypeWorkloadProtectableItem),
				}, true),
				DiffSuppressFunc: ignoreCaseDiffSuppressFunc,
			},

			"source_resource_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceArmRecoveryServicesProtectedItemCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).recoveryServicesProtectedItemsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	location := d.Get("location").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	tags := d.Get("tags").(map[string]interface{})

	vaultName := d.Get("recovery_vault_name").(string)
	fabricName := d.Get("fabric_name").(string)
	containerName := d.Get("container_name").(string)
	itemType := d.Get("type").(string)
	sourceId := d.Get("source_resource_id").(string)

	log.Printf("[DEBUG] Creating/updating Recovery Service Protected Item %q (resource group %q)", name, resourceGroup)

	item := backup.ProtectedItemResource{
		Location: utils.String(location),
		Tags:     expandTags(tags),
		Properties: &backup.ProtectedItem{
			//PolicyID: &sourceId,
			ProtectedItemType: backup.ProtectedItemType(itemType),
			SourceResourceID:  &sourceId,
		},
	}

	//create recovery services vault

	if _, err := client.CreateOrUpdate(ctx, vaultName, resourceGroup, fabricName, containerName, name, item); err != nil {
		return fmt.Errorf("Error creating/updating Recovery Service Protected Item %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"NotFound"},
		Target:     []string{"Found"},
		Timeout:    60 * time.Minute,
		MinTimeout: 30 * time.Second,
		Delay:      30 * time.Second, // required because it takes some time before the 'creating' location shows up
		Refresh: func() (interface{}, string, error) {

			resp, err := client.Get(ctx, vaultName, resourceGroup, fabricName, containerName, name, "")
			if err != nil {
				if utils.ResponseWasNotFound(resp.Response) {
					d.SetId("")
					return resp, "NotFound", nil
				}

				return resp, "Error", fmt.Errorf("Error making Read request on Recovery Service Protected Item %q (Resource Group %q): %+v", name, resourceGroup, err)
			}

			return resp, "Found", nil
		},
	}

	resp, errr := stateConf.WaitForState()
	if errr != nil {
		return fmt.Errorf("Error waiting for the Recovery Service Protected Item %q (Resource Group %q) to provision: %+v", name, resourceGroup, errr)
	}

	d.SetId(*resp.(backup.ProtectedItemResource).ID)

	return resourceArmRecoveryServicesVaultRead(d, meta)
}

//"id": "/subscriptions/c0a607b2-6372-4ef3-abdb-dbe52a7b56ba/resourceGroups
// /tfex-recovery_services/providers/Microsoft.RecoveryServices/vaults/example-recovery-vault/backupFabrics/Azure/protectionContainers/iaasvmcontainer;
// iaasvmcontainerv2;tfex-recovery_services;tfexrecove407vm/protectedItems/vm;iaasvmcontainerv2;tfex-recovery_services;tfexrecove407vm",

func resourceArmRecoveryServicesProtectedItemRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).recoveryServicesProtectedItemsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["vaults"]
	vaultName := id.Path["vaults"]
	fabricName := id.Path["backupFabrics"]
	containerName := id.Path["protectionContainers"]
	resourceGroup := id.ResourceGroup

	log.Printf("[DEBUG] Reading Recovery Service Protected Item %q (resource group %q)", name, resourceGroup)

	resp, err := client.Get(ctx, vaultName, resourceGroup, fabricName, containerName, name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on Recovery Service Protected Item %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azureRMNormalizeLocation(*location))
	}

	flattenAndSetTags(d, resp.Tags)

	return nil
}

func resourceArmRecoveryServicesProtectedItemDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).recoveryServicesProtectedItemsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["vaults"]
	vaultName := id.Path["vaults"]
	fabricName := id.Path["backupFabrics"]
	containerName := id.Path["protectionContainers"]
	resourceGroup := id.ResourceGroup

	log.Printf("[DEBUG] Deleting Recovery Service Protected Item %q (resource group %q)", name, resourceGroup)

	resp, err := client.Delete(ctx, vaultName, resourceGroup, fabricName, containerName, name)
	if err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("Error issuing delete request for Recovery Service Protected Item %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	return nil
}
