package pagerduty

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyAutomationActionsAction() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyAutomationActionsActionCreate,
		Read:   resourcePagerDutyAutomationActionsActionRead,
		Update: resourcePagerDutyAutomationActionsActionUpdate,
		Delete: resourcePagerDutyAutomationActionsActionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"action_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: validateValueDiagFunc([]string{
					"script",
					"process_automation",
				}),
				ForceNew: true, // Requires creation of new action
			},
			"runner_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"action_data_reference": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"process_automation_job_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"process_automation_job_arguments": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"process_automation_node_filter": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"script": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"invocation_command": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"action_classification": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateDiagFunc: validateValueDiagFunc([]string{
					"diagnostic",
					"remediation",
				}),
			},
			"runner_type": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"modify_time": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"only_invocable_on_unresolved_incidents": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"map_to_all_services": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func buildAutomationActionsActionStruct(d *schema.ResourceData) (*pagerduty.AutomationActionsAction, error) {
	automationActionsAction := pagerduty.AutomationActionsAction{
		Name:       d.Get("name").(string),
		ActionType: d.Get("action_type").(string),
	}

	// The API does not allow new actions without a description, but legacy actions without a description exist
	if attr, ok := d.GetOk("description"); ok {
		val := attr.(string)
		automationActionsAction.Description = &val
	} else {
		return nil, errors.New("action description must be specified when creating an action")
	}

	if attr, ok := d.GetOk("runner_id"); ok {
		val := attr.(string)
		automationActionsAction.RunnerID = &val
	}

	if attr, ok := d.GetOk("action_data_reference"); ok {
		automationActionsAction.ActionDataReference = expandActionDataReference(attr)
	} else {
		return nil, errors.New("action_data_reference must be specified when creating an action")
	}

	if attr, ok := d.GetOk("type"); ok {
		val := attr.(string)
		automationActionsAction.Type = &val
	}

	if attr, ok := d.GetOk("action_classification"); ok {
		val := attr.(string)
		automationActionsAction.ActionClassification = &val
	}

	if attr, ok := d.GetOk("runner_type"); ok {
		val := attr.(string)
		automationActionsAction.RunnerType = &val
	}

	if attr, ok := d.GetOk("creation_time"); ok {
		val := attr.(string)
		automationActionsAction.CreationTime = &val
	}

	if attr, ok := d.GetOk("modify_time"); ok {
		val := attr.(string)
		automationActionsAction.ModifyTime = &val
	}

	if attr, ok := d.GetOk("only_invocable_on_unresolved_incidents"); ok {
		val := attr.(bool)
		automationActionsAction.OnlyInvocableOnUnresolvedIncidents = &val
	}

	attr, _ := d.Get("only_invocable_on_unresolved_incidents").(bool)
	automationActionsAction.OnlyInvocableOnUnresolvedIncidents = &attr

	if attr, ok := d.GetOk("map_to_all_services"); ok {
		val := attr.(bool)
		automationActionsAction.MapToAllServices = &val
	}

	attr2, _ := d.Get("map_to_all_services").(bool)
	automationActionsAction.MapToAllServices = &attr2

	return &automationActionsAction, nil
}

func expandActionDataReference(v interface{}) pagerduty.AutomationActionsActionDataReference {
	attr_map := v.([]interface{})[0].(map[string]interface{})
	adr := pagerduty.AutomationActionsActionDataReference{}

	if v, ok := attr_map["process_automation_job_id"]; ok {
		v_str := v.(string)
		if v_str != "" {
			adr.ProcessAutomationJobId = &v_str
		}
	}

	if v, ok := attr_map["process_automation_job_arguments"]; ok {
		v_str := v.(string)
		if v_str != "" {
			adr.ProcessAutomationJobArguments = &v_str
		}
	}

	if v, ok := attr_map["process_automation_node_filter"]; ok {
		v_str := v.(string)
		if v_str != "" {
			adr.ProcessAutomationNodeFilter = &v_str
		}
	}

	if v, ok := attr_map["script"]; ok {
		v_str := v.(string)
		if v_str != "" {
			adr.Script = &v_str
		}
	}

	if v, ok := attr_map["invocation_command"]; ok {
		v_str := v.(string)
		if v_str != "" {
			adr.InvocationCommand = &v_str
		}
	}

	return adr
}

func resourcePagerDutyAutomationActionsActionCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	automationActionsAction, err := buildAutomationActionsActionStruct(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating PagerDuty AutomationActionsAction %s", automationActionsAction.Name)

	retryErr := retry.Retry(2*time.Minute, func() *retry.RetryError {
		if automationActionsAction, _, err := client.AutomationActionsAction.Create(automationActionsAction); err != nil {
			if isErrCode(err, 400) || isErrCode(err, 429) {
				time.Sleep(2 * time.Second)
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		} else if automationActionsAction != nil {
			d.SetId(automationActionsAction.ID)
		}
		return nil
	})

	if retryErr != nil {
		return retryErr
	}

	return resourcePagerDutyAutomationActionsActionRead(d, meta)
}

func resourcePagerDutyAutomationActionsActionUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	automationActionsAction, err := buildAutomationActionsActionStruct(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Updating PagerDuty AutomationActionsAction %s", d.Id())

	if _, _, err := client.AutomationActionsAction.Update(d.Id(), automationActionsAction); err != nil {
		return err
	}

	return nil
}

func resourcePagerDutyAutomationActionsActionRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading PagerDuty AutomationActionsAction %s", d.Id())

	return retry.Retry(2*time.Minute, func() *retry.RetryError {
		if automationActionsAction, _, err := client.AutomationActionsAction.Get(d.Id()); err != nil {
			if isErrCode(err, http.StatusBadRequest) {
				return retry.NonRetryableError(err)
			}

			time.Sleep(2 * time.Second)
			return retry.RetryableError(err)
		} else if automationActionsAction != nil {
			d.Set("name", automationActionsAction.Name)
			d.Set("type", automationActionsAction.Type)
			d.Set("action_type", automationActionsAction.ActionType)
			d.Set("creation_time", automationActionsAction.CreationTime)

			if automationActionsAction.Description != nil {
				d.Set("description", &automationActionsAction.Description)
			}

			f_adr := flattenActionDataReference(automationActionsAction.ActionDataReference)
			if err := d.Set("action_data_reference", f_adr); err != nil {
				return retry.NonRetryableError(err)
			}

			if automationActionsAction.ModifyTime != nil {
				d.Set("modify_time", &automationActionsAction.ModifyTime)
			}

			if automationActionsAction.RunnerID != nil {
				d.Set("runner_id", &automationActionsAction.RunnerID)
			}

			if automationActionsAction.RunnerType != nil {
				d.Set("runner_type", &automationActionsAction.RunnerType)
			}

			if automationActionsAction.ActionClassification != nil {
				d.Set("action_classification", &automationActionsAction.ActionClassification)
			}

			if automationActionsAction.OnlyInvocableOnUnresolvedIncidents != nil {
				d.Set("only_invocable_on_unresolved_incidents", *automationActionsAction.OnlyInvocableOnUnresolvedIncidents)
			}

			if automationActionsAction.MapToAllServices != nil {
				d.Set("map_to_all_services", *automationActionsAction.MapToAllServices)
			}
		}
		return nil
	})
}

func flattenActionDataReference(adr pagerduty.AutomationActionsActionDataReference) []interface{} {
	adr_map := map[string]interface{}{}

	if adr.ProcessAutomationJobId != nil {
		adr_map["process_automation_job_id"] = *adr.ProcessAutomationJobId
	}

	if adr.ProcessAutomationJobArguments != nil {
		adr_map["process_automation_job_arguments"] = *adr.ProcessAutomationJobArguments
	}

	if adr.ProcessAutomationNodeFilter != nil {
		adr_map["process_automation_node_filter"] = *adr.ProcessAutomationNodeFilter
	}

	if adr.Script != nil {
		adr_map["script"] = *adr.Script
	}

	if adr.InvocationCommand != nil {
		adr_map["invocation_command"] = *adr.InvocationCommand
	}

	return []interface{}{adr_map}
}

func resourcePagerDutyAutomationActionsActionDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Deleting PagerDuty AutomationActionsAction %s", d.Id())

	retryErr := retry.Retry(2*time.Minute, func() *retry.RetryError {
		if _, err := client.AutomationActionsAction.Delete(d.Id()); err != nil {
			if isErrCode(err, http.StatusBadRequest) {
				return retry.NonRetryableError(err)
			}

			return retry.RetryableError(err)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}
	d.SetId("")

	// giving the API time to catchup
	time.Sleep(time.Second)
	return nil
}
