package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourcePagerDutyUserContactMethod_Basic(t *testing.T) {
	name := fmt.Sprintf("%s %s", acctest.RandString(8), acctest.RandString(10))
	method_type := "email_contact_method"
	address := fmt.Sprintf("%s@%s.com", acctest.RandString(6), acctest.RandString(7))
	second_address := fmt.Sprintf("%s@%s.com", acctest.RandString(6), acctest.RandString(7))
	label := "Work"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyUserContactMethodConfig(name, method_type, address, second_address, label),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourcePagerDutyUserContactMethod("pagerduty_user_contact_method.test", "data.pagerduty_user_contact_method.by_summary_type_and_user_id"),
				),
			},
		},
	})
}

func testAccDataSourcePagerDutyUserContactMethod(src, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		srcR := s.RootModule().Resources[src]
		srcA := srcR.Primary.Attributes

		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["id"] == "" {
			return fmt.Errorf("Expected to get a contact method ID from PagerDuty")
		}

		testAtts := []string{"id", "summary"}

		for _, att := range testAtts {
			if a[att] != srcA[att] {
				return fmt.Errorf("Expected the contact method %s to be: %s, but got: %s", att, srcA[att], a[att])
			}
		}

		return nil
	}
}

func testAccDataSourcePagerDutyUserContactMethodConfig(name string, method_type string, address string, second_address string, label string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%[1]v"
  email       = "%[3]v"
  color       = "red"
  role        = "user"
  job_title   = "bar"
  description = "bar"
}
resource "pagerduty_user_contact_method" "test" {
  user_id      = pagerduty_user.foo.id
  type         = "%[2]v"
  address      = "%[4]v"
  label        = "%[5]v"
}

data "pagerduty_user_contact_method" "by_summary_type_and_user_id" {
  label = pagerduty_user_contact_method.test.label
  user_id = pagerduty_user.foo.id
  type = "%[2]s"
}
`, name, method_type, address, second_address, label)
}
