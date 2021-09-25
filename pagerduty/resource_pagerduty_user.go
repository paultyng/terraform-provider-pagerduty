package pagerduty

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyUser() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyUserCreate,
		Read:   resourcePagerDutyUserRead,
		Update: resourcePagerDutyUserUpdate,
		Delete: resourcePagerDutyUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				// Suppress the diff shown if there are leading or trailing spaces
				DiffSuppressFunc: suppressLeadTrailSpaceDiff,
			},

			"email": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressCaseDiff,
			},

			"color": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"role": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "user",
				ValidateFunc: validateValueFunc([]string{
					"admin",
					"limited_user",
					"observer",
					"owner",
					"read_only_user",
					"restricted_access",
					"read_only_limited_user",
					"user",
				}),
			},

			"job_title": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"avatar_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"teams": {
				Type:       schema.TypeSet,
				Deprecated: "Use the 'pagerduty_team_membership' resource instead.",
				Computed:   true,
				Optional:   true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},

			"time_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"invitation_sent": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
		},
	}
}

func buildUserStruct(d *schema.ResourceData) *pagerduty.User {
	user := &pagerduty.User{
		Name:  strings.TrimSpace(d.Get("name").(string)),
		Email: d.Get("email").(string),
	}

	if attr, ok := d.GetOk("color"); ok {
		user.Color = attr.(string)
	}

	if attr, ok := d.GetOk("time_zone"); ok {
		user.TimeZone = attr.(string)
	}

	if attr, ok := d.GetOk("role"); ok {
		user.Role = attr.(string)
	}

	if attr, ok := d.GetOk("job_title"); ok {
		user.JobTitle = attr.(string)
	}

	if attr, ok := d.GetOk("description"); ok {
		user.Description = attr.(string)
	}
	log.Printf("[DEBUG] buildUserStruct-- d: .%v. user:%v.", d.Get("name").(string), user.Name)
	return user
}

func resourcePagerDutyUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	user := buildUserStruct(d)

	log.Printf("[INFO] Creating PagerDuty user %s", user.Name)

	user, _, err := client.Users.Create(user)
	if err != nil {
		return err
	}

	d.SetId(user.ID)

	return resourcePagerDutyUserUpdate(d, meta)
}

func resourcePagerDutyUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] pooh Reading PagerDuty user %s", d.Id())

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		user, _, err := client.Users.Get(d.Id(), &pagerduty.GetUserOptions{})
		if err != nil {
			errResp := handleNotFoundError(err, d)
			if errResp != nil {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(errResp)
			}

			return nil
		}
		// Trimming whitespace on names in case of mistyped spaces
		d.Set("name", user.Name)
		d.Set("email", user.Email)
		d.Set("time_zone", user.TimeZone)
		d.Set("html_url", user.HTMLURL)
		d.Set("color", user.Color)
		d.Set("role", user.Role)
		d.Set("avatar_url", user.AvatarURL)
		d.Set("description", user.Description)
		d.Set("job_title", user.JobTitle)

		if err := d.Set("teams", flattenTeams(user.Teams)); err != nil {
			return resource.NonRetryableError(
				fmt.Errorf("error setting teams: %s", err),
			)
		}

		d.Set("invitation_sent", user.InvitationSent)

		return nil
	})
}

func resourcePagerDutyUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	user := buildUserStruct(d)

	log.Printf("[INFO] Updating PagerDuty user %s", d.Id())

	// Retrying to give other resources (such as escalation policies) to delete
	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if _, _, err := client.Users.Update(d.Id(), user); err != nil {
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

	if d.HasChange("teams") {
		o, n := d.GetChange("teams")

		if o == nil {
			o = new(schema.Set)
		}

		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		remove := expandStringList(os.Difference(ns).List())
		add := expandStringList(ns.Difference(os).List())

		for _, t := range remove {

			if _, _, err := client.Teams.Get(t); err != nil {
				log.Printf("[INFO] PagerDuty team: %s not found, removing dangling team reference for user %s", t, d.Id())
				continue
			}

			log.Printf("[INFO] Removing PagerDuty user %s from team: %s", d.Id(), t)

			if _, err := client.Teams.RemoveUser(t, d.Id()); err != nil {
				return err
			}
		}

		for _, t := range add {
			log.Printf("[INFO] Adding PagerDuty user %s to team: %s", d.Id(), t)

			if _, err := client.Teams.AddUser(t, d.Id()); err != nil {
				return err
			}
		}
	}

	return resourcePagerDutyUserRead(d, meta)
}

func resourcePagerDutyUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty user %s", d.Id())

	// Retrying to give other resources (such as escalation policies) to delete
	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if _, err := client.Users.Delete(d.Id()); err != nil {
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
