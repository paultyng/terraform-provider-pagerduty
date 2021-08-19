package pagerduty

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				Type:     schema.TypeMap,
				// Using the `Elem` block to define specific keys for the map is currently not possible.
				// The workaround described in SDK documentation is to confirm the required keys are set when expanding the Map object inside the resource code.
				// See https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ValidateDiagFunc: validation.MapKeyMatch(regexp.MustCompile("(id|type)"), "`contact_method` must only have `id` and `types` attributes"),
			},
		},
	}
}

func buildUserNotificationRuleStruct(d *schema.ResourceData) (*pagerduty.NotificationRule, error) {
	contactMethod, err := expandContactMethod(d.Get("contact_method"))
	if err != nil {
		return nil, err
	}
	notificationRule := &pagerduty.NotificationRule{
		Type:                "assignment_notification_rule",
		StartDelayInMinutes: d.Get("start_delay_in_minutes").(int),
		Urgency:             d.Get("urgency").(string),
		ContactMethod:       contactMethod,
	}

	return notificationRule, nil
}

func resourcePagerDutyUserNotificationRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	userID := d.Get("user_id").(string)

	notificationRule, err := buildUserNotificationRuleStruct(d)
	if err != nil {
		return err
	}

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

		d.Set("urgency", resp.Urgency)
		d.Set("start_delay_in_minutes", resp.StartDelayInMinutes)
		d.Set("contact_method", flattenContactMethod(resp.ContactMethod))

		return nil
	})
}

func resourcePagerDutyUserNotificationRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	notificationRule, err := buildUserNotificationRuleStruct(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Updating PagerDuty user notification rule %s", d.Id())

	userID := d.Get("user_id").(string)

	if _, _, err := client.Users.UpdateNotificationRule(userID, d.Id(), notificationRule); err != nil {
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

func expandContactMethod(v interface{}) (*pagerduty.ContactMethodReference, error) {
	cm := v.(map[string]interface{})

	if _, ok := cm["id"]; !ok {
		return nil, fmt.Errorf("the `id` attribute of `contact_method` is required")
	}

	if t, ok := cm["type"]; !ok {
		return nil, fmt.Errorf("the `type` attribute of `contact_method` is required")
	} else {
		switch t {
		case "email_contact_method":
		case "phone_contact_method":
		case "push_notification_contact_method":
		case "sms_contact_method":
			// Valid
		default:
			return nil, fmt.Errorf("the `type` attribute of `contact_method` must be one of `email_contact_method`, `phone_contact_method`, `push_notification_contact_method` or `sms_contact_method`")
		}
	}

	var contactMethod = &pagerduty.ContactMethodReference{
		ID:   cm["id"].(string),
		Type: cm["type"].(string),
	}

	return contactMethod, nil
}

func flattenContactMethod(v *pagerduty.ContactMethodReference) map[string]interface{} {

	var contactMethod = map[string]interface{}{
		"id":   v.ID,
		"type": v.Type,
	}

	return contactMethod
}
