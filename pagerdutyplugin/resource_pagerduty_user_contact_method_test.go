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

func TestAccPagerDutyUserContactMethodEmail_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	usernameUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.test", username)
	emailUpdated := fmt.Sprintf("%s@foo.test", usernameUpdated)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyUserContactMethodDestroy,
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
	email := fmt.Sprintf("%s@foo.test", username)
	emailUpdated := fmt.Sprintf("%s@foo.test", usernameUpdated)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyUserContactMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserContactMethodPhoneConfig(username, email, "4153013250"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserContactMethodExists("pagerduty_user_contact_method.foo"),
				),
			},
			{
				Config:      testAccCheckPagerDutyUserContactMethodPhoneConfig(username, email, "04153013250"),
				ExpectError: regexp.MustCompile("Phone number can't start with a zero"),
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

func TestAccPagerDutyUserContactMethodPhone_FormatValidation(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.test", username)
	tooLongNumber := "4153013250415301325041530132504153013250,415301325041530132504,530132504153013250"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyUserContactMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckPagerDutyUserContactMethodPhoneFormatValidationConfig(username, email, "phone_contact_method", "1", tooLongNumber),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("phone numbers may not exceed 40 characters"),
			},
			{
				Config:      testAccCheckPagerDutyUserContactMethodPhoneFormatValidationConfig(username, email, "phone_contact_method", "1", "+4153013250"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("phone numbers may only include digits from 0-9 and the symbols"),
			},
			{
				Config:      testAccCheckPagerDutyUserContactMethodPhoneFormatValidationConfig(username, email, "phone_contact_method", "44", "01332412251"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Trunk prefixes are not supported for following countries and regions: France, Romania, UK, Denmark, Germany, Australia, Thailand, India and North America, so must be formatted for international use without the leading 0"),
			},
			{
				Config:      testAccCheckPagerDutyUserContactMethodPhoneFormatValidationConfig(username, email, "sms_contact_method", "52", "15558889999"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Mexico-based SMS numbers should be free of area code prefixes, so please remove the leading 1 in the number"),
			},
		},
	})
}

func TestAccPagerDutyUserContactMethodPhone_EnforceUpdateIfAlreadyExist(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.test", username)
	phoneNumber := "4153013250"
	newPhoneNumber := "4153013251"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyUserContactMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserContactMethodPhoneConfig(username, email, phoneNumber),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserContactMethodExists("pagerduty_user_contact_method.foo"),
					testAccAddPhoneContactOutsideTerraform("pagerduty_user_contact_method.foo", newPhoneNumber),
				),
			},
			{
				Config: testAccCheckPagerDutyUserContactMethodPhoneConfig(username, email, newPhoneNumber),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"pagerduty_user_contact_method.foo", "label", username),
					resource.TestCheckResourceAttr(
						"pagerduty_user_contact_method.foo", "address", newPhoneNumber),
				),
			},
		},
	})
}

func TestAccPagerDutyUserContactMethodSMS_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	usernameUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.test", username)
	emailUpdated := fmt.Sprintf("%s@foo.test", usernameUpdated)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyUserContactMethodDestroy,
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

func TestAccPagerDutyUserContactMethodPhone_NoPermaDiffWhenOmittingCountryCode(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.test", username)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyUserContactMethodDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserContactMethodPhoneNoPermaDiffWhenOmittingCountryCodeConfig(username, email, "4153013250"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserContactMethodExists("pagerduty_user_contact_method.foo"),
				),
			},
			{
				Config:   testAccCheckPagerDutyUserContactMethodPhoneNoPermaDiffWhenOmittingCountryCodeConfig(username, email, "4153013250"),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckPagerDutyUserContactMethodDestroy(s *terraform.State) error {
	ctx := context.Background()
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_user_contact_method" {
			continue
		}

		_, err := testAccProvider.client.GetUserContactMethodWithContext(ctx, r.Primary.Attributes["user_id"], r.Primary.ID)
		if err == nil {
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

		ctx := context.Background()
		found, err := testAccProvider.client.GetUserContactMethodWithContext(ctx, rs.Primary.Attributes["user_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Contact method not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccAddPhoneContactOutsideTerraform(n, p string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		resID := rs.Primary.ID

		if resID == "" {
			return fmt.Errorf("No User Contact Method ID is set")
		}
		userID := rs.Primary.Attributes["user_id"]

		ctx := context.Background()
		found, err := testAccProvider.client.GetUserContactMethodWithContext(ctx, rs.Primary.Attributes["user_id"], rs.Primary.ID)
		if err != nil {
			return err
		}

		found.Address = p
		_, err = testAccProvider.client.CreateUserContactMethodWithContext(ctx, userID, *found)
		if err != nil {
			return fmt.Errorf("was not possible to set phone %s contact number outside Terraform state: %v", p, err)
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

func testAccCheckPagerDutyUserContactMethodPhoneFormatValidationConfig(username, email, method_type, countryCode, phone string) string {
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
  type         = "%[3]s"
  country_code = "+%[4]s"
  address      = "%[5]s"
  label        = "%[1]v"
}
`, username, email, method_type, countryCode, phone)
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

func testAccCheckPagerDutyUserContactMethodPhoneNoPermaDiffWhenOmittingCountryCodeConfig(username, email, phone string) string {
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
  address      = "%[3]s"
  label        = "%[1]v"
}
`, username, email, phone)
}
