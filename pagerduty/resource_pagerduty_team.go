package pagerduty

import (
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyTeam() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyTeamCreate,
		Read:   resourcePagerDutyTeamRead,
		Update: resourcePagerDutyTeamUpdate,
		Delete: resourcePagerDutyTeamDelete,
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
			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func buildTeamStruct(d *schema.ResourceData) *pagerduty.Team {
	team := &pagerduty.Team{
		Name: d.Get("name").(string),
	}

	if attr, ok := d.GetOk("description"); ok {
		team.Description = attr.(string)
	}
	if attr, ok := d.GetOk("parent"); ok {
		team.Parent = &pagerduty.TeamReference{
			ID:   attr.(string),
			Type: "team_reference",
		}
	}
	return team
}

func resourcePagerDutyTeamCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	team := buildTeamStruct(d)

	log.Printf("[INFO] Creating PagerDuty team %s", team.Name)

	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if team, _, err := client.Teams.Create(team); err != nil {
			return resource.RetryableError(err)
		} else if team != nil {
			d.SetId(team.ID)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}

	return resourcePagerDutyTeamRead(d, meta)

}

func resourcePagerDutyTeamRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty team %s", d.Id())

	return resource.Retry(30*time.Second, func() *resource.RetryError {
		if team, _, err := client.Teams.Get(d.Id()); err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
		} else if team != nil {
			d.Set("name", team.Name)
			d.Set("description", team.Description)
			d.Set("html_url", team.HTMLURL)
		}
		return nil
	})
}

func resourcePagerDutyTeamUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	team := buildTeamStruct(d)

	log.Printf("[INFO] Updating PagerDuty team %s", d.Id())

	retryErr := resource.Retry(30*time.Second, func() *resource.RetryError {
		if _, _, err := client.Teams.Update(d.Id(), team); err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}
	return resourcePagerDutyTeamRead(d, meta)
}

func resourcePagerDutyTeamDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty team %s", d.Id())

	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if _, err := client.Teams.Delete(d.Id()); err != nil {
			return resource.RetryableError(err)
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
