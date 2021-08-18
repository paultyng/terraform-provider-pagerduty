package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func TestAccPagerDutyUserNotificationRuleContactMethod_Basic(t *testing.T) {
	contactMethodType1 := "email_contact_method"
	contactMethodType2 := "phone_contact_method"
	contactMethodType3 := "sms_contact_method"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyUserNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserNotificationRuleContactMethodConfig(contactMethodType1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserNotificationRuleExists("pagerduty_user_notification_rule.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyUserNotificationRuleContactMethodConfig(contactMethodType2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserNotificationRuleExists("pagerduty_user_notification_rule.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyUserNotificationRuleContactMethodConfig(contactMethodType3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserNotificationRuleExists("pagerduty_user_notification_rule.foo"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyUserNotificationRuleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_user_notification_rule" {
			continue
		}

		if _, _, err := client.Users.GetNotificationRule(r.Primary.Attributes["user_id"], r.Primary.ID); err == nil {
			return fmt.Errorf("User notification rule still exists")
		}

	}
	return nil
}

func testAccCheckPagerDutyUserNotificationRuleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No user notification rule ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, _, err := client.Users.GetNotificationRule(rs.Primary.Attributes["user_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Notification rule not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyUserNotificationRuleContactMethodConfig(methodType string) string {
	return fmt.Sprintf(`
resource "pagerduty_user_notification_rule" "foo" {
  user_id                = pagerduty_user.foo.id
  start_delay_in_minutes = 1
  urgency                = "high"

  contact_method = {
    type = "%[1]v"
    id   = pagerduty_user_contact_method.%[1]v.id
  }
}

resource "pagerduty_user" "foo" {
  name        = "foo bar"
  email       = "foo@bar.com"
  color       = "red"
  role        = "user"
  job_title   = "bar"
  description = "bar"
}

resource "pagerduty_user_contact_method" "email_contact_method" {
  user_id = pagerduty_user.foo.id
  type    = "email_contact_method"
  address = "foo-1@bar.com"
  label   = "Work"
}

resource "pagerduty_user_contact_method" "sms_contact_method" {
  user_id      = pagerduty_user.foo.id
  type         = "sms_contact_method"
  address      = "8015541234"
  country_code = "+1"
  label        = "Work"
}

resource "pagerduty_user_contact_method" "phone_contact_method" {
  user_id      = pagerduty_user.foo.id
  type         = "phone_contact_method"
  country_code = "+1"
  address      = "8015541234"
  label        = "Work"
}

`, methodType)

}
