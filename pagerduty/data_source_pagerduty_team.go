package pagerduty

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func dataSourcePagerDutyTeam() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyTeamRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the team to find in the PagerDuty API",
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"members": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePagerDutyTeamRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty team")

	searchTeam := d.Get("name").(string)

	o := &pagerduty.ListTeamsOptions{
		Query: searchTeam,
	}

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, _, err := client.Teams.List(o)
		if err != nil {
			if isErrCode(err, 429) {
				// Delaying retry by 30s as recommended by PagerDuty
				// https://developer.pagerduty.com/docs/rest-api-v2/rate-limiting/#what-are-possible-workarounds-to-the-events-api-rate-limit
				time.Sleep(30 * time.Second)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		var found *pagerduty.Team

		for _, team := range resp.Teams {
			if team.Name == searchTeam {
				found = team
				break
			}
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("Unable to locate any team with name: %s", searchTeam),
			)
		}

		if found != nil {
			mo := &pagerduty.GetMembersOptions{}
			resp, _, err := client.Teams.GetMembers(searchTeam, mo)
			if err != nil {
				if isErrCode(err, 429) {
					// Delaying retry by 30s as recommended by PagerDuty
					// https://developer.pagerduty.com/docs/rest-api-v2/rate-limiting/#what-are-possible-workarounds-to-the-events-api-rate-limit
					time.Sleep(30 * time.Second)
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}

			if err := d.Set("members", dataSourceMembersRead(resp.Members)); err != nil {
				return fmt.Errorf("error setting users: %w", err)
			}
		}

		d.SetId(found.ID)
		d.Set("name", found.Name)
		d.Set("description", found.Description)
		d.Set("parent", found.Parent)

		return nil
	})
}

func dataSourceMembersRead(m []*pagerduty.Member) []map[string]interface{} {
	members := make([]map[string]interface{}, 0, len(m))
	for _, i := range m {
		m := make(map[string]interface{})
		m["id"] = i.User.ID
		members = append(members, m)
	}
	return members
}
