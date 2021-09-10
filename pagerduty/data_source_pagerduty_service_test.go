package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourcePagerDutyService_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyServiceConfig(username, email, service, escalationPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourcePagerDutyService("pagerduty_service.test", "data.pagerduty_service.by_name"),
				),
			},
		},
	})
}

func testAccDataSourcePagerDutyService(src, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		srcR := s.RootModule().Resources[src]
		srcA := srcR.Primary.Attributes

		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["id"] == "" {
			return fmt.Errorf("Expected to get a service ID from PagerDuty")
		}

		testAtts := []string{"id", "name"}

		for _, att := range testAtts {
			if a[att] != srcA[att] {
				return fmt.Errorf("Expected the service %s to be: %s, but got: %s", att, srcA[att], a[att])
			}
		}

		return nil
	}
}

func testAccDataSourcePagerDutyServiceConfig(username, email, service, escalationPolicy string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "test" {
  name  = "%s"
  email = "%s"
}

resource "pagerduty_escalation_policy" "test" {
  name        = "%s"
  num_loops   = 2
  rule {
    escalation_delay_in_minutes = 10
    target {
      type = "user_reference"
      id   = pagerduty_user.test.id
    }
  }
}

resource "pagerduty_service" "test" {
  name                    = "%s"
  auto_resolve_timeout    = 14400
  acknowledgement_timeout = 600
  escalation_policy       = pagerduty_escalation_policy.test.id
  alert_creation          = "create_incidents"
}

data "pagerduty_service" "by_name" {
  name = pagerduty_service.test.name
}
`, username, email, service, escalationPolicy)
}
