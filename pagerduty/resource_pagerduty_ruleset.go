package pagerduty

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyRuleset() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyRulesetCreate,
		Read:   resourcePagerDutyRulesetRead,
		Update: resourcePagerDutyRulesetUpdate,
		Delete: resourcePagerDutyRulesetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"team": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"routing_keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func buildRulesetStruct(d *schema.ResourceData) *pagerduty.Ruleset {
	ruleset := &pagerduty.Ruleset{
		Name: d.Get("name").(string),
	}

	if attr, ok := d.GetOk("team"); ok {
		ruleset.Team = expandTeam(attr)
	}

	if attr, ok := d.GetOk("routing_keys"); ok {
		ruleset.RoutingKeys = expandKeys(attr.([]interface{}))
	}

	if attr, ok := d.GetOk("type"); ok {
		ruleset.Type = attr.(string)
	}

	return ruleset
}

func expandKeys(v []interface{}) []string {
	keys := make([]string, len(v))

	for i, k := range v {
		keys[i] = fmt.Sprintf("%v", k)
	}

	return keys
}

func expandTeam(v interface{}) *pagerduty.RulesetObject {
	var team *pagerduty.RulesetObject
	t := v.([]interface{})[0].(map[string]interface{})
	team = &pagerduty.RulesetObject{
		ID: t["id"].(string),
	}

	return team
}

func flattenTeam(v *pagerduty.RulesetObject) []interface{} {
	team := map[string]interface{}{
		"id": v.ID,
	}

	return []interface{}{team}
}
func resourcePagerDutyRulesetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	ruleset := buildRulesetStruct(d)

	log.Printf("[INFO] Creating PagerDuty ruleset: %s", ruleset.Name)

	retryErr := resource.Retry(10*time.Second, func() *resource.RetryError {
		if ruleset, _, err := client.Rulesets.Create(ruleset); err != nil {
			if isErrCode(err, 400) || isErrCode(err, 429) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		} else if ruleset != nil {
			d.SetId(ruleset.ID)
		}
		return nil
	})

	if retryErr != nil {
		return retryErr
	}

	return resourcePagerDutyRulesetRead(d, meta)
}

func resourcePagerDutyRulesetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty ruleset: %s", d.Id())

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		ruleset, _, err := client.Rulesets.Get(d.Id())
		if err != nil {
			errResp := handleNotFoundError(err, d)
			if errResp != nil {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(errResp)
			}

			return nil
		}
		d.Set("name", ruleset.Name)
		d.Set("type", ruleset.Type)

		// if ruleset is found set to ResourceData
		if ruleset.Team != nil {
			d.Set("team", flattenTeam(ruleset.Team))
		}
		d.Set("routing_keys", ruleset.RoutingKeys)

		return nil
	})
}
func resourcePagerDutyRulesetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	ruleset := buildRulesetStruct(d)

	log.Printf("[INFO] Updating PagerDuty ruleset: %s", d.Id())

	if _, _, err := client.Rulesets.Update(d.Id(), ruleset); err != nil {
		return err
	}

	return nil
}

func resourcePagerDutyRulesetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty ruleset: %s", d.Id())

	if _, err := client.Rulesets.Delete(d.Id()); err != nil {
		return err
	}

	d.SetId("")

	return nil
}
