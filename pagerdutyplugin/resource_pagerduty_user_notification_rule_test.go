package pagerduty

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPagerDutyUserNotificationRuleContactMethod_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.test", username)
	contactMethodType1 := "email_contact_method"
	contactMethodType2 := "phone_contact_method"
	contactMethodType3 := "sms_contact_method"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyUserNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserNotificationRuleContactMethodConfig(contactMethodType1, username, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserNotificationRuleExists("pagerduty_user_notification_rule.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyUserNotificationRuleContactMethodConfig(contactMethodType2, username, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserNotificationRuleExists("pagerduty_user_notification_rule.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyUserNotificationRuleContactMethodConfig(contactMethodType3, username, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserNotificationRuleExists("pagerduty_user_notification_rule.foo"),
				),
			},
		},
	})
}

func TestAccPagerDutyUserNotificationRuleContactMethod_Invalid(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.test", username)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyUserNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserNotificationRuleContactMethodConfig_Invalid(username, email),
				ExpectError: regexp.MustCompile(
					`Attribute contact_method.type value must be one of: \["email_contact_method"` +
						`\s+"phone_contact_method" "push_notification_contact_method"` +
						`\s+"sms_contact_method"\], got: "invalid_contact_method`,
				),
			},
		},
	})
}

func TestAccPagerDutyUserNotificationRuleContactMethod_Missing_id(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.test", username)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyUserNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckPagerDutyUserNotificationRuleContactMethodConfig_Missing_id(username, email),
				ExpectError: regexp.MustCompile(`The argument "id" is required, but no definition was found.`),
			},
		},
	})
}

func TestAccPagerDutyUserNotificationRuleContactMethod_Missing_type(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.test", username)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyUserNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckPagerDutyUserNotificationRuleContactMethodConfig_Missing_type(username, email),
				ExpectError: regexp.MustCompile(`The argument "type" is required, but no definition was found.`),
			},
		},
	})
}

func TestAccPagerDutyUserNotificationRuleContactMethod_Unknown_key(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.test", username)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyUserNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckPagerDutyUserNotificationRuleContactMethodConfig_Unknown_key(username, email),
				ExpectError: regexp.MustCompile(`An argument named "foo" is not expected here.`),
			},
		},
	})
}

func testAccCheckPagerDutyUserNotificationRuleDestroy(s *terraform.State) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_user_notification_rule" {
			continue
		}

		client := testAccProvider.client
		ctx := context.Background()
		if _, err := client.GetUserNotificationRuleWithContext(ctx, r.Primary.Attributes["user_id"], r.Primary.ID); err == nil {
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

		client := testAccProvider.client
		ctx := context.Background()
		found, err := client.GetUserNotificationRuleWithContext(ctx, rs.Primary.Attributes["user_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Notification rule not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyUserNotificationRuleContactMethodConfig(methodType, username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_user_notification_rule" "foo" {
  user_id                = pagerduty_user.foo.id
  start_delay_in_minutes = 1
  urgency                = "high"

  contact_method {
    type = "%[1]v"
    id   = pagerduty_user_contact_method.%[1]v.id
  }
}

resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
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
`, methodType, username, email)
}

func testAccCheckPagerDutyUserNotificationRuleContactMethodConfig_Invalid(username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_user_notification_rule" "foo" {
  user_id                = pagerduty_user.foo.id
  start_delay_in_minutes = 1
  urgency                = "high"

  contact_method {
    type = "invalid_contact_method"
    id   = pagerduty_user_contact_method.email_contact_method.id
  }
}

resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
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
`, username, email)
}

func testAccCheckPagerDutyUserNotificationRuleContactMethodConfig_Missing_id(username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_user_notification_rule" "foo" {
  user_id                = pagerduty_user.foo.id
  start_delay_in_minutes = 1
  urgency                = "high"

  contact_method {
    type = "invalid_contact_method"
  }
}

resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
  color       = "red"
  role        = "user"
  job_title   = "bar"
  description = "bar"
}
`, username, email)
}

func testAccCheckPagerDutyUserNotificationRuleContactMethodConfig_Missing_type(username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_user_notification_rule" "foo" {
  user_id                = pagerduty_user.foo.id
  start_delay_in_minutes = 1
  urgency                = "high"

  contact_method {
    id   = pagerduty_user_contact_method.email_contact_method.id
  }
}

resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
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
`, username, email)
}

func testAccCheckPagerDutyUserNotificationRuleContactMethodConfig_Unknown_key(username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_user_notification_rule" "foo" {
  user_id                = pagerduty_user.foo.id
  start_delay_in_minutes = 1
  urgency                = "high"

  contact_method {
    type = pagerduty_user_contact_method.email_contact_method.type
    id   = pagerduty_user_contact_method.email_contact_method.id
    foo  = "bar"
  }
}

resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
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
`, username, email)
}
