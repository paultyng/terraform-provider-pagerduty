package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourcePagerDutyUser_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.test", username)
	jobTitle := fmt.Sprintf("%s-title", username)
	timeZone := "America/New_York"
	role := "user"
	description := fmt.Sprintf("%s-description", username)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyUserConfig(username, email, jobTitle, timeZone, role, description),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourcePagerDutyUser("pagerduty_user.test", "data.pagerduty_user.by_email"),
				),
			},
		},
	})
}

func testAccDataSourcePagerDutyUser(src, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		srcR := s.RootModule().Resources[src]
		srcA := srcR.Primary.Attributes

		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["id"] == "" {
			return fmt.Errorf("Expected to get a user ID from PagerDuty")
		}

		testAtts := []string{"id", "name", "email", "job_title", "time_zone", "role", "description"}

		for _, att := range testAtts {
			if a[att] != srcA[att] {
				return fmt.Errorf("Expected the user %s to be: %s, but got: %s", att, srcA[att], a[att])
			}
		}

		return nil
	}
}

func testAccDataSourcePagerDutyUserConfig(username, email, jobTitle, timeZone, role, description string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "test" {
  name = "%s"
  email = "%s"
  job_title = "%s"
  time_zone = "%s"
  role = "%s"
  description = "%s"
}

data "pagerduty_user" "by_email" {
	email = pagerduty_user.test.email
}
`, username, email, jobTitle, timeZone, role, description)
}
