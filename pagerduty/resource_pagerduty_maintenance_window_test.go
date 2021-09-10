package pagerduty

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func testSweepMaintenanceWindow(region string) error {
	config, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client, err := config.Client()
	if err != nil {
		return err
	}

	resp, _, err := client.MaintenanceWindows.List(&pagerduty.ListMaintenanceWindowsOptions{})
	if err != nil {
		return err
	}

	for _, window := range resp.MaintenanceWindows {
		if strings.HasPrefix(window.Description, "test") || strings.HasPrefix(window.Description, "tf-") {
			log.Printf("Destroying maintenance window %s (%s)", window.Description, window.ID)
			if _, err := client.MaintenanceWindows.Delete(window.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccPagerDutyMaintenanceWindow_Basic(t *testing.T) {
	window := fmt.Sprintf("tf-%s", acctest.RandString(5))
	windowStartTime := timeNowInAccLoc().Add(24 * time.Hour).Format(time.RFC3339)
	windowEndTime := timeNowInAccLoc().Add(48 * time.Hour).Format(time.RFC3339)
	windowUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	windowUpdatedStartTime := timeNowInAccLoc().Add(48 * time.Hour).Format(time.RFC3339)
	windowUpdatedEndTime := timeNowInAccLoc().Add(72 * time.Hour).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyAddonDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyMaintenanceWindowConfig(window, windowStartTime, windowEndTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyMaintenanceWindowExists("pagerduty_maintenance_window.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyMaintenanceWindowConfigUpdated(windowUpdated, windowUpdatedStartTime, windowUpdatedEndTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyMaintenanceWindowExists("pagerduty_maintenance_window.foo"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyMaintenanceWindowDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_maintenance_window" {
			continue
		}

		if _, _, err := client.MaintenanceWindows.Get(r.Primary.ID); err == nil {
			return fmt.Errorf("maintenance window still exists")
		}

	}
	return nil
}

func testAccCheckPagerDutyMaintenanceWindowExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No maintenance window ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, _, err := client.MaintenanceWindows.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("maintenance window not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyMaintenanceWindowConfig(desc, start, end string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%[1]v"
  email       = "%[1]v@foo.com"
  color       = "green"
  role        = "user"
  job_title   = "foo"
  description = "foo"
}

resource "pagerduty_escalation_policy" "foo" {
  name        = "%[1]v"
  description = "bar"
  num_loops   = 2

  rule {
    escalation_delay_in_minutes = 10

    target {
      type = "user_reference"
      id   = pagerduty_user.foo.id
    }
  }
}

resource "pagerduty_service" "foo" {
  name                    = "%[1]v"
  description             = "foo"
  auto_resolve_timeout    = 1800
  acknowledgement_timeout = 1800
  escalation_policy       = pagerduty_escalation_policy.foo.id

  incident_urgency_rule {
    type    = "constant"
    urgency = "high"
  }
}

resource "pagerduty_maintenance_window" "foo" {
  description = "%[1]v"
  start_time  = "%[2]v"
  end_time    = "%[3]v"
  services    = [pagerduty_service.foo.id]
}
`, desc, start, end)
}

func testAccCheckPagerDutyMaintenanceWindowConfigUpdated(desc, start, end string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%[1]v"
  email       = "%[1]v@foo.com"
  color       = "green"
  role        = "user"
  job_title   = "foo"
  description = "foo"
}

resource "pagerduty_escalation_policy" "foo" {
  name        = "%[1]v"
  description = "bar"
  num_loops   = 2

  rule {
    escalation_delay_in_minutes = 10

    target {
      type = "user_reference"
      id   = pagerduty_user.foo.id
    }
  }
}

resource "pagerduty_service" "foo" {
  name                    = "%[1]v"
  description             = "foo"
  auto_resolve_timeout    = 1800
  acknowledgement_timeout = 1800
  escalation_policy       = pagerduty_escalation_policy.foo.id

  incident_urgency_rule {
    type    = "constant"
    urgency = "high"
  }
}

resource "pagerduty_service" "foo2" {
  name                    = "%[1]v2"
  description             = "foo2"
  auto_resolve_timeout    = 1800
  acknowledgement_timeout = 1800
  escalation_policy       = pagerduty_escalation_policy.foo.id

  incident_urgency_rule {
    type    = "constant"
    urgency = "high"
  }
}

resource "pagerduty_maintenance_window" "foo" {
  description = "%[1]v"
  start_time  = "%[2]v"
  end_time    = "%[3]v"
  services    = [pagerduty_service.foo.id, pagerduty_service.foo2.id]
}
`, desc, start, end)
}
