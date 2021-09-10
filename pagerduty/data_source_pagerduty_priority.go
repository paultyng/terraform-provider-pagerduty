package pagerduty

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func dataSourcePagerDutyPriority() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyPriorityRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the priority to find in the PagerDuty API",
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePagerDutyPriorityRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty priority")

	searchTeam := d.Get("name").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, _, err := client.Priorities.List()
		if err != nil {
			if isErrCode(err, 429) {
				// Delaying retry by 30s as recommended by PagerDuty
				// https://developer.pagerduty.com/docs/rest-api-v2/rate-limiting/#what-are-possible-workarounds-to-the-events-api-rate-limit
				time.Sleep(30 * time.Second)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		var found *pagerduty.Priority

		for _, priority := range resp.Priorities {
			if strings.EqualFold(priority.Name, searchTeam) {
				found = priority
				break
			}
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("Unable to locate any priority with name: %s", searchTeam),
			)
		}

		d.SetId(found.ID)
		d.Set("name", found.Name)
		d.Set("description", found.Description)

		return nil
	})
}
