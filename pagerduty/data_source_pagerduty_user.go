package pagerduty

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func dataSourcePagerDutyUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyUserRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"job_title": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"time_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePagerDutyUserRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading PagerDuty user")

	searchEmail := d.Get("email").(string)

	o := &pagerduty.ListUsersOptions{
		Query: searchEmail,
	}

	return retry.Retry(5*time.Minute, func() *retry.RetryError {
		resp, err := client.Users.ListAll(o)
		if err != nil {
			if isErrCode(err, http.StatusBadRequest) {
				return retry.NonRetryableError(err)
			}

			// Delaying retry by 30s as recommended by PagerDuty
			// https://developer.pagerduty.com/docs/rest-api-v2/rate-limiting/#what-are-possible-workarounds-to-the-events-api-rate-limit
			time.Sleep(30 * time.Second)
			return retry.RetryableError(err)
		}

		var found *pagerduty.FullUser

		for _, user := range resp {
			if strings.EqualFold(user.Email, searchEmail) {
				found = user
				break
			}
		}

		if found == nil {
			return retry.NonRetryableError(
				fmt.Errorf("Unable to locate any user with the email: %s", searchEmail),
			)
		}

		d.SetId(found.ID)
		d.Set("name", found.Name)
		d.Set("email", found.Email)
		d.Set("role", found.Role)
		d.Set("job_title", found.JobTitle)
		d.Set("time_zone", found.TimeZone)
		d.Set("description", found.Description)

		return nil
	})
}
