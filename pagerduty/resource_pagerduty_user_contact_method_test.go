package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func TestAccPagerDutyUserContactMethodEmail_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	usernameUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	emailUpdated := fmt.Sprintf("%s@foo.com", usernameUpdated)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyUserContactMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserContactMethodEmailConfig(username, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserContactMethodExists("pagerduty_user_contact_method.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyUserContactMethodEmailConfigUpdated(usernameUpdated, emailUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserContactMethodExists("pagerduty_user_contact_method.foo"),
				),
			},
		},
	})
}

func TestAccPagerDutyUserContactMethodPhone_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	usernameUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	emailUpdated := fmt.Sprintf("%s@foo.com", usernameUpdated)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyUserContactMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserContactMethodPhoneConfig(username, email, "4153013250"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserContactMethodExists("pagerduty_user_contact_method.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyUserContactMethodPhoneConfig(usernameUpdated, emailUpdated, "8019351337"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserContactMethodExists("pagerduty_user_contact_method.foo"),
				),
			},
		},
	})
}

func TestAccPagerDutyUserContactMethodSMS_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	usernameUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	emailUpdated := fmt.Sprintf("%s@foo.com", usernameUpdated)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyUserContactMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserContactMethodSMSConfig(username, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserContactMethodExists("pagerduty_user_contact_method.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyUserContactMethodSMSConfigUpdated(usernameUpdated, emailUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserContactMethodExists("pagerduty_user_contact_method.foo"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyUserContactMethodDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_user_contact_method" {
			continue
		}

		if _, _, err := client.Users.GetContactMethod(r.Primary.Attributes["user_id"], r.Primary.ID); err == nil {
			return fmt.Errorf("User contact method still exists")
		}

	}
	return nil
}

func testAccCheckPagerDutyUserContactMethodExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No user contact method ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, _, err := client.Users.GetContactMethod(rs.Primary.Attributes["user_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Contact method not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyUserContactMethodEmailConfig(username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%[1]v"
  email       = "%[2]v"
  color       = "red"
  role        = "user"
  job_title   = "bar"
  description = "bar"
}

resource "pagerduty_user_contact_method" "foo" {
  user_id = pagerduty_user.foo.id
  type    = "email_contact_method"
  address = "%[1]v%[2]v"
  label   = "%[1]v"
}
`, username, email)
}

func testAccCheckPagerDutyUserContactMethodEmailConfigUpdated(username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%[1]v"
  email       = "%[2]v"
  color       = "red"
  role        = "user"
  job_title   = "bar"
  description = "bar"
}

resource "pagerduty_user_contact_method" "foo" {
  user_id = pagerduty_user.foo.id
  type    = "email_contact_method"
  address = "%[1]v%[2]v"
  label   = "%[1]v"
}
`, username, email)
}

func testAccCheckPagerDutyUserContactMethodPhoneConfig(username, email, phone string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%[1]v"
  email       = "%[2]v"
  color       = "red"
  role        = "user"
  job_title   = "bar"
  description = "bar"
}

resource "pagerduty_user_contact_method" "foo" {
  user_id      = pagerduty_user.foo.id
  type         = "phone_contact_method"
  country_code = "+1"
  address      = "%[3]s"
  label        = "%[1]v"
}
`, username, email, phone)
}

func testAccCheckPagerDutyUserContactMethodSMSConfig(username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%[1]v"
  email       = "%[2]v"
  color       = "red"
  role        = "user"
  job_title   = "bar"
  description = "bar"
}

resource "pagerduty_user_contact_method" "foo" {
  user_id      = pagerduty_user.foo.id
  type         = "sms_contact_method"
  country_code = "+1"
  address      = "8458003889"
  label        = "%[1]v"
}
`, username, email)
}

func testAccCheckPagerDutyUserContactMethodSMSConfigUpdated(username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%[1]v"
  email       = "%[2]v"
  color       = "red"
  role        = "user"
  job_title   = "bar"
  description = "bar"
}

resource "pagerduty_user_contact_method" "foo" {
  user_id      = pagerduty_user.foo.id
  type         = "sms_contact_method"
  country_code = "+1"
  address      = "6509892965"
  label        = "%[1]v"
}
`, username, email)
}
