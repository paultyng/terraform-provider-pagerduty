package pagerduty

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyEscalationPolicyCreate,
		Read:   resourcePagerDutyEscalationPolicyRead,
		Update: resourcePagerDutyEscalationPolicyUpdate,
		Delete: resourcePagerDutyEscalationPolicyDelete,
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
				Default:  "Managed by Terraform",
			},
			"num_loops": {
				Type:     schema.TypeInt,
				Optional: true,
				ValidateFunc: validation.IntBetween(0, 9),
			},
			"teams": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"escalation_delay_in_minutes": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"target": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "user_reference",
									},
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func buildEscalationPolicyStruct(d *schema.ResourceData) *pagerduty.EscalationPolicy {
	escalationPolicy := &pagerduty.EscalationPolicy{
		Name:            d.Get("name").(string),
		EscalationRules: expandEscalationRules(d.Get("rule").([]interface{})),
	}

	if attr, ok := d.GetOk("description"); ok {
		escalationPolicy.Description = attr.(string)
	}

	loops := d.Get("num_loops").(int)
	escalationPolicy.NumLoops = &loops

	if attr, ok := d.GetOk("teams"); ok {
		escalationPolicy.Teams = expandTeams(attr.([]interface{}))
	}

	return escalationPolicy
}

func resourcePagerDutyEscalationPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	escalationPolicy := buildEscalationPolicyStruct(d)

	log.Printf("[INFO] Creating PagerDuty escalation policy: %s", escalationPolicy.Name)

	escalationPolicy, _, err := client.EscalationPolicies.Create(escalationPolicy)
	if err != nil {
		return err
	}

	d.SetId(escalationPolicy.ID)

	return resourcePagerDutyEscalationPolicyRead(d, meta)
}

func resourcePagerDutyEscalationPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty escalation policy: %s", d.Id())

	o := &pagerduty.GetEscalationPolicyOptions{}

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		escalationPolicy, _, err := client.EscalationPolicies.Get(d.Id(), o)
		if err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
		}

		d.Set("name", escalationPolicy.Name)
		d.Set("description", escalationPolicy.Description)
		d.Set("num_loops", escalationPolicy.NumLoops)

		if err := d.Set("teams", flattenTeams(escalationPolicy.Teams)); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error setting teams: %s", err))
		}

		if err := d.Set("rule", flattenEscalationRules(escalationPolicy.EscalationRules)); err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
}

func resourcePagerDutyEscalationPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	escalationPolicy := buildEscalationPolicyStruct(d)

	log.Printf("[INFO] Updating PagerDuty escalation policy: %s", d.Id())

	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if _, _, err := client.EscalationPolicies.Update(d.Id(), escalationPolicy); err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}

	return nil
}

func resourcePagerDutyEscalationPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty escalation policy: %s", d.Id())

	// Retrying to give other resources (such as services) to delete
	retryErr := resource.Retry(30*time.Second, func() *resource.RetryError {
		if _, err := client.EscalationPolicies.Delete(d.Id()); err != nil {
			if isErrCode(err, 400) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
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

func expandEscalationRules(v interface{}) []*pagerduty.EscalationRule {
	var escalationRules []*pagerduty.EscalationRule

	for _, er := range v.([]interface{}) {
		rer := er.(map[string]interface{})
		escalationRule := &pagerduty.EscalationRule{
			EscalationDelayInMinutes: rer["escalation_delay_in_minutes"].(int),
		}

		for _, ert := range rer["target"].([]interface{}) {
			rert := ert.(map[string]interface{})
			escalationRuleTarget := &pagerduty.EscalationTargetReference{
				ID:   rert["id"].(string),
				Type: rert["type"].(string),
			}

			escalationRule.Targets = append(escalationRule.Targets, escalationRuleTarget)
		}

		escalationRules = append(escalationRules, escalationRule)
	}

	return escalationRules
}

func flattenEscalationRules(v []*pagerduty.EscalationRule) []map[string]interface{} {
	var escalationRules []map[string]interface{}

	for _, er := range v {
		escalationRule := map[string]interface{}{
			"id":                          er.ID,
			"escalation_delay_in_minutes": er.EscalationDelayInMinutes,
		}

		var targets []map[string]interface{}

		for _, ert := range er.Targets {
			escalationRuleTarget := map[string]interface{}{"id": ert.ID, "type": ert.Type}
			targets = append(targets, escalationRuleTarget)
		}

		escalationRule["target"] = targets

		escalationRules = append(escalationRules, escalationRule)
	}

	return escalationRules
}

func expandTeams(v interface{}) []*pagerduty.TeamReference {
	var teams []*pagerduty.TeamReference

	for _, t := range v.([]interface{}) {
		team := &pagerduty.TeamReference{
			ID:   t.(string),
			Type: "team_reference",
		}
		teams = append(teams, team)
	}

	return teams
}

func flattenTeams(teams []*pagerduty.TeamReference) []string {
	res := make([]string, len(teams))
	for i, t := range teams {
		res[i] = t.ID
	}

	return res
}
