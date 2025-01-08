package pagerduty

import (
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourcePagerDutyAutomationActionsAction() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyAutomationActionsActionRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"action_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"runner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"action_data_reference": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"process_automation_job_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"process_automation_job_arguments": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"process_automation_node_filter": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"script": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"invocation_command": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
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
				Computed: true,
			},
			"runner_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

func dataSourcePagerDutyAutomationActionsActionRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading PagerDuty AutomationActionsAction")

	return retry.Retry(2*time.Minute, func() *retry.RetryError {
		automationActionsAction, _, err := client.AutomationActionsAction.Get(d.Get("id").(string))
		if err != nil {
			if isErrCode(err, http.StatusBadRequest) {
				return retry.NonRetryableError(err)
			}

			// Delaying retry by 30s as recommended by PagerDuty
			// https://developer.pagerduty.com/docs/rest-api-v2/rate-limiting/#what-are-possible-workarounds-to-the-events-api-rate-limit
			time.Sleep(30 * time.Second)
			return retry.RetryableError(err)
		}

		d.SetId(automationActionsAction.ID)
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

		return nil
	})
}
