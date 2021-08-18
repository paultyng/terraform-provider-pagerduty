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

func resourcePagerDutyUserNotificationRule() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyUserNotificationRuleCreate,
		Read:   resourcePagerDutyUserNotificationRuleRead,
		Update: resourcePagerDutyUserNotificationRuleUpdate,
		Delete: resourcePagerDutyUserNotificationRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePagerDutyUserNotificationRuleImport,
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"start_delay_in_minutes": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"urgency": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validateValueFunc([]string{
					"high",
					"low",
				}),
			},
			"contact_method": {
				Required: true,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validateValueFunc([]string{
								"email_contact_method",
								"phone_contact_method",
								"push_notification_contact_method",
								"sms_contact_method",
							}),
						},
					},
				},
			},
		},
	}
}

func buildUserNotificationRuleStruct(d *schema.ResourceData) *pagerduty.NotificationRule {
	notificationRule := &pagerduty.NotificationRule{
		Type:                "assignment_notification_rule",
		StartDelayInMinutes: d.Get("start_delay_in_minutes").(int),
		Urgency:             d.Get("urgency").(string),
		ContactMethod:       expandContactMethod(d.Get("contact_method")),
	}

	return notificationRule
}

func resourcePagerDutyUserNotificationRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	userID := d.Get("user_id").(string)

	notificationRule := buildUserNotificationRuleStruct(d)

	resp, _, err := client.Users.CreateNotificationRule(userID, notificationRule)
	if err != nil {
		return err
	}

	d.SetId(resp.ID)

	return resourcePagerDutyUserNotificationRuleRead(d, meta)
}

func resourcePagerDutyUserNotificationRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	userID := d.Get("user_id").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		resp, _, err := client.Users.GetNotificationRule(userID, d.Id())
		if err != nil {
			errResp := handleNotFoundError(err, d)
			if errResp != nil {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(errResp)
			}

			return nil
		}

		d.Set("type", resp.Type)
		d.Set("urgency", resp.Urgency)
		d.Set("start_delay_in_minutes", resp.StartDelayInMinutes)
		d.Set("contact_method", flattenContactMethod(resp.ContactMethod))

		return nil
	})
}

func resourcePagerDutyUserNotificationRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	contactMethod := buildUserNotificationRuleStruct(d)

	log.Printf("[INFO] Updating PagerDuty user notification rule %s", d.Id())

	userID := d.Get("user_id").(string)

	if _, _, err := client.Users.UpdateNotificationRule(userID, d.Id(), contactMethod); err != nil {
		return err
	}

	return resourcePagerDutyUserNotificationRuleRead(d, meta)
}

func resourcePagerDutyUserNotificationRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty user notification rule %s", d.Id())

	userID := d.Get("user_id").(string)

	if _, err := client.Users.DeleteNotificationRule(userID, d.Id()); err != nil {
		return handleNotFoundError(err, d)
	}

	d.SetId("")

	return nil
}

func resourcePagerDutyUserNotificationRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*pagerduty.Client)

	ids := strings.Split(d.Id(), ":")

	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing pagerduty_user_notification_rule. Expecting an ID formed as '<user_id>.<notification_rule_id>'")
	}
	uid, id := ids[0], ids[1]

	_, _, err := client.Users.GetNotificationRule(uid, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)
	d.Set("user_id", uid)

	return []*schema.ResourceData{d}, nil
}

func expandContactMethod(v interface{}) *pagerduty.ContactMethodReference {
	cm := v.(map[string]interface{})

	var contactMethod = &pagerduty.ContactMethodReference{
		ID:   cm["id"].(string),
		Type: cm["type"].(string),
	}

	return contactMethod
}

func flattenContactMethod(v *pagerduty.ContactMethodReference) map[string]interface{} {

	var contactMethod = map[string]interface{}{
		"id":   v.ID,
		"type": v.Type,
	}

	return contactMethod
}
