package pagerduty

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

// Deprecated: Migrated to pagerdutyplugin.dataSourceEscalationPolicy. Kept for testing purposes.
func dataSourcePagerDutyEscalationPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyEscalationPolicyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourcePagerDutyEscalationPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading PagerDuty escalation policy")

	searchName := d.Get("name").(string)
	var offset int = 0
	var found *pagerduty.EscalationPolicy
	more := true

	for more {
		err := retry.Retry(5*time.Minute, func() *retry.RetryError {
			o := &pagerduty.ListEscalationPoliciesOptions{
				Query:  searchName,
				Limit:  100,
				Offset: offset,
			}

			resp, _, err := client.EscalationPolicies.List(o)
			if err != nil {
				if isErrCode(err, http.StatusBadRequest) {
					return retry.NonRetryableError(err)
				}

				// Delaying retry by 30s as recommended by PagerDuty
				// https://developer.pagerduty.com/docs/rest-api-v2/rate-limiting/#what-are-possible-workarounds-to-the-events-api-rate-limit
				time.Sleep(30 * time.Second)
				return retry.RetryableError(err)
			}

			offset += 100
			more = resp.More

			for _, policy := range resp.EscalationPolicies {
				if policy.Name == searchName {
					found = policy
					more = false
					return nil
				}
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	if found == nil {
		return fmt.Errorf("Unable to locate any escalation policy with the name: %s", searchName)
	}

	d.SetId(found.ID)
	d.Set("name", found.Name)

	return nil
}
