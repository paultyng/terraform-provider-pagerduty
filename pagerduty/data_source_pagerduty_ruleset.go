package pagerduty

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func dataSourcePagerDutyRuleset() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyRulesetRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"routing_keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourcePagerDutyRulesetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty ruleset")

	searchName := d.Get("name").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, _, err := client.Rulesets.List()
		if err != nil {
			if isErrCode(err, 429) {
				// Delaying retry by 30s as recommended by PagerDuty
				// https://developer.pagerduty.com/docs/rest-api-v2/rate-limiting/#what-are-possible-workarounds-to-the-events-api-rate-limit
				time.Sleep(30 * time.Second)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		var found *pagerduty.Ruleset

		for _, ruleset := range resp.Rulesets {
			if ruleset.Name == searchName {
				found = ruleset
				break
			}
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("Unable to locate any ruleset with the name: %s", searchName),
			)
		}

		d.SetId(found.ID)
		d.Set("name", found.Name)
		d.Set("routing_keys", found.RoutingKeys)

		return nil
	})
}
