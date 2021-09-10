package pagerduty

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourcePagerDutySchedule_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	schedule := fmt.Sprintf("tf-%s", acctest.RandString(5))
	location := "Europe/Berlin"
	start := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)
	rotationVirtualStart := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyScheduleConfig(username, email, schedule, location, start, rotationVirtualStart),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourcePagerDutySchedule("pagerduty_schedule.test", "data.pagerduty_schedule.by_name"),
				),
			},
		},
	})
}

func testAccDataSourcePagerDutySchedule(src, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		srcR := s.RootModule().Resources[src]
		srcA := srcR.Primary.Attributes

		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["id"] == "" {
			return fmt.Errorf("Expected to get a schedule ID from PagerDuty")
		}

		testAtts := []string{"id", "name"}

		for _, att := range testAtts {
			if a[att] != srcA[att] {
				return fmt.Errorf("Expected the schedule %s to be: %s, but got: %s", att, srcA[att], a[att])
			}
		}

		return nil
	}
}

func testAccDataSourcePagerDutyScheduleConfig(username, email, schedule, location, start, rotationVirtualStart string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "test" {
  name  = "%s"
  email = "%s"
}

resource "pagerduty_schedule" "test" {
  name = "%s"

  time_zone = "%s"

  layer {
    name                         = "foo"
    start                        = "%s"
    rotation_virtual_start       = "%s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.test.id]

    restriction {
      type              = "weekly_restriction"
      start_time_of_day = "08:00:00"
      start_day_of_week = 5
      duration_seconds  = 32101
    }
  }
}

data "pagerduty_schedule" "by_name" {
  name = pagerduty_schedule.test.name
}
`, username, email, schedule, location, start, rotationVirtualStart)
}
