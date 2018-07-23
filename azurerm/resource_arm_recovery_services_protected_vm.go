package azurerm

import (
	"fmt"
	"log"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/services/recoveryservices/mgmt/2016-06-01/backup"

	"time"

	"strings"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmRecoveryServicesProtectedVm() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmRecoveryServicesProtectedVMCreateUpdate,
		Read:   resourceArmRecoveryServicesProtectedVMRead,
		Update: resourceArmRecoveryServicesProtectedVMCreateUpdate,
		Delete: resourceArmRecoveryServicesProtectedVMDelete,

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

			"source_vm_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"source_vm_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"backup_policy_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "DefaultPolicy", //every vault comes with a 'DefaultPolicy' by default
				ValidateFunc: validation.NoZeroValues,
			},

			"tags": tagsSchema(),
		},
	}
}

//Subscriptions/c0a607b2-6372-4ef3-abdb-dbe52a7b56ba/resourceGroups/tfex-recovery_services/providers/Microsoft.RecoveryServices/vaults/example-recovery-vault/backupFabrics/Azure
func resourceArmRecoveryServicesProtectedVMCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).recoveryServicesProtectedItemsClient
	ctx := meta.(*ArmClient).StopContext

	name := d.Get("name").(string)
	location := d.Get("location").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	tags := d.Get("tags").(map[string]interface{})

	vaultName := d.Get("recovery_vault_name").(string)
	vmId := d.Get("source_vm_id").(string)
	vmName := d.Get("source_vm_name").(string)

	policyName := d.Get("backup_policy_name").(string)

	protectedItemName := fmt.Sprintf("VM;iaasvmcontainerv2;%s;%s", resourceGroup, vmName)
	containerName := fmt.Sprintf("iaasvmcontainer;iaasvmcontainerv2;%s;%s", resourceGroup, vmName)

	log.Printf("[DEBUG] Creating/updating Recovery Service Protected Item %q (resource group %q)", name, resourceGroup)

	// TODO: support for the Backup Policy Resource - available in the `backup` SDK
	backupPolicyId := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.RecoveryServices/vaults/%s/backupPolicies/%s", client.SubscriptionID, resourceGroup, vaultName, policyName)

	item := backup.ProtectedItemResource{
		Location: utils.String(location),
		Tags:     expandTags(tags),
		Properties: &backup.AzureIaaSComputeVMProtectedItem{
			PolicyID:          &backupPolicyId,
			ProtectedItemType: backup.ProtectedItemType(backup.ProtectedItemTypeMicrosoftClassicComputevirtualMachines),
			WorkloadType:      backup.VM,
			SourceResourceID:  utils.String(vmId),
			FriendlyName:      utils.String(vmName),
			VirtualMachineID:  utils.String(vmId),
		},
	}

	if _, err := client.CreateOrUpdate(ctx, vaultName, resourceGroup, "Azure", containerName, protectedItemName, item); err != nil {
		return fmt.Errorf("Error creating/updating Recovery Service Protected Item %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"NotFound"},
		Target:     []string{"Found"},
		Timeout:    60 * time.Minute,
		MinTimeout: 30 * time.Second,
		Delay:      10 * time.Second,
		Refresh: func() (interface{}, string, error) {

			resp, err := client.Get(ctx, vaultName, resourceGroup, "Azure", containerName, protectedItemName, "")
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

	resp, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for the Recovery Service Protected Item %q (Resource Group %q) to provision: %+v", name, resourceGroup, err)
	}

	id := strings.Replace(*resp.(backup.ProtectedItemResource).ID, "Subscriptions", "subscriptions", 1)
	d.SetId(id)

	return resourceArmRecoveryServicesVaultRead(d, meta)
}

//"id": "/subscriptions/c0a607b2-6372-4ef3-abdb-dbe52a7b56ba/resourceGroups
// /tfex-recovery_services/providers/Microsoft.RecoveryServices/vaults/example-recovery-vault/backupFabrics/Azure/protectionContainers/iaasvmcontainer;
// iaasvmcontainerv2;tfex-recovery_services;tfexrecove407vm/protectedItems/vm;iaasvmcontainerv2;tfex-recovery_services;tfexrecove407vm",

//"protectionContainer": "[concat(']",
//"protectedItem": "[concat('vm;iaasvmcontainerv2;', resourceGroup().name, ';', variables('vmName'))]"

func resourceArmRecoveryServicesProtectedVMRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).recoveryServicesProtectedItemsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["protectedItems"]
	vaultName := id.Path["vaults"]
	resourceGroup := id.ResourceGroup
	containerName := id.Path["protectionContainers"]

	log.Printf("[DEBUG] Reading Recovery Service Protected Item %q (resource group %q)", name, resourceGroup)

	resp, err := client.Get(ctx, vaultName, resourceGroup, "Azure", containerName, name, "")
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
	d.Set("recovery_vault_name", vaultName)

	if properties := resp.Properties; properties != nil {
		if vm, ok := properties.AsAzureIaaSComputeVMProtectedItem(); ok {
			d.Set("source_vm_id", vm.SourceResourceID)
			d.Set("source_vm_name", vm.FriendlyName)
			//d.Set("backup_policy_name", vm.PolicyID) //parse out?
		}
	}

	flattenAndSetTags(d, resp.Tags)

	return nil
}

func resourceArmRecoveryServicesProtectedVMDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).recoveryServicesProtectedItemsClient
	ctx := meta.(*ArmClient).StopContext

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	name := id.Path["protectedItems"]
	resourceGroup := id.ResourceGroup
	vaultName := id.Path["vaults"]
	containerName := id.Path["protectionContainers"]

	log.Printf("[DEBUG] Deleting Recovery Service Protected Item %q (resource group %q)", name, resourceGroup)

	resp, err := client.Delete(ctx, vaultName, resourceGroup, "Azure", containerName, name)
	if err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("Error issuing delete request for Recovery Service Protected Item %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	return nil
}
