package pagerduty

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func dataSourcePagerDutyBusinessService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyBusinessServiceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourcePagerDutyBusinessServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty business service")

	searchName := d.Get("name").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, _, err := client.BusinessServices.List()
		if err != nil {
			if isErrCode(err, 429) {
				// Delaying retry by 30s as recommended by PagerDuty
				// https://developer.pagerduty.com/docs/rest-api-v2/rate-limiting/#what-are-possible-workarounds-to-the-events-api-rate-limit
				time.Sleep(30 * time.Second)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		var found *pagerduty.BusinessService

		for _, businessService := range resp.BusinessServices {
			if businessService.Name == searchName {
				found = businessService
				break
			}
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("Unable to locate any business service with the name: %s", searchName),
			)
		}

		d.SetId(found.ID)
		d.Set("name", found.Name)

		return nil
	})

}
