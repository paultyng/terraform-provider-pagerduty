package pagerduty

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func init() {
	resource.AddTestSweepers("pagerduty_schedule", &resource.Sweeper{
		Name: "pagerduty_schedule",
		F:    testSweepSchedule,
	})
}

func testSweepSchedule(region string) error {
	config, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client, err := config.Client()
	if err != nil {
		return err
	}

	resp, _, err := client.Schedules.List(&pagerduty.ListSchedulesOptions{})
	if err != nil {
		return err
	}

	for _, schedule := range resp.Schedules {
		if strings.HasPrefix(schedule.Name, "test") || strings.HasPrefix(schedule.Name, "tf-") {
			log.Printf("Destroying schedule %s (%s)", schedule.Name, schedule.ID)
			if _, err := client.Schedules.Delete(schedule.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccPagerDutySchedule_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	schedule := fmt.Sprintf("tf-%s", acctest.RandString(5))
	scheduleUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	location := "America/New_York"
	start := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)
	rotationVirtualStart := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyScheduleConfig(username, email, schedule, location, start, rotationVirtualStart),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyScheduleExists("pagerduty_schedule.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "name", schedule),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "time_zone", location),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.name", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.start", start),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.rotation_virtual_start", rotationVirtualStart),
				),
			},
			{
				Config: testAccCheckPagerDutyScheduleConfigUpdated(username, email, scheduleUpdated, location, start, rotationVirtualStart),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyScheduleExists("pagerduty_schedule.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "name", scheduleUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "time_zone", location),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.name", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.start", start),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.rotation_virtual_start", rotationVirtualStart),
				),
			},
			{
				Config:      testAccCheckPagerDutyScheduleConfigRestrictionType(username, email, schedule, location, start, rotationVirtualStart),
				ExpectError: regexp.MustCompile("start_day_of_week must only be set for a weekly_restriction schedule restriction type"),
			},
		},
	})
}

func TestAccPagerDutyScheduleWithTeams_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	schedule := fmt.Sprintf("tf-%s", acctest.RandString(5))
	scheduleUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	location := "America/New_York"
	start := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)
	rotationVirtualStart := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)
	team := fmt.Sprintf("tf-%s", acctest.RandString(5))
	teamUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyScheduleWithTeamsConfig(username, email, schedule, location, start, rotationVirtualStart, team),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyScheduleExists("pagerduty_schedule.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "name", schedule),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "time_zone", location),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.name", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.start", start),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.rotation_virtual_start", rotationVirtualStart),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "teams.#", "1"),
				),
			},
			{
				Config: testAccCheckPagerDutyScheduleWithTeamsConfigUpdated(username, email, scheduleUpdated, location, start, rotationVirtualStart, teamUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyScheduleExists("pagerduty_schedule.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "name", scheduleUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "time_zone", location),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.name", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.start", start),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.rotation_virtual_start", rotationVirtualStart),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "teams.#", "1"),
				),
			},
		},
	})
}
func TestAccPagerDutyScheduleOverflow_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	schedule := fmt.Sprintf("tf-%s", acctest.RandString(5))
	scheduleUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	location := "America/New_York"
	start := timeNowInLoc(location).Add(30 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)
	rotationVirtualStart := timeNowInLoc(location).Add(30 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyScheduleOverflowConfig(username, email, schedule, location, start, rotationVirtualStart),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyScheduleExists("pagerduty_schedule.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyScheduleOverflowConfigUpdated(username, email, scheduleUpdated, location, start, rotationVirtualStart),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyScheduleExists("pagerduty_schedule.foo"),
				),
			},
		},
	})
}

func TestAccPagerDutySchedule_BasicWeek(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	schedule := fmt.Sprintf("tf-%s", acctest.RandString(5))
	scheduleUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	location := "Australia/Melbourne"
	start := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)
	rotationVirtualStart := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyScheduleConfigWeek(username, email, schedule, location, start, rotationVirtualStart),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyScheduleExists("pagerduty_schedule.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "name", schedule),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "time_zone", location),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.name", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.restriction.0.start_day_of_week", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.start", start),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.rotation_virtual_start", rotationVirtualStart),
				),
			},
			{
				Config: testAccCheckPagerDutyScheduleConfigWeekUpdated(username, email, scheduleUpdated, location, start, rotationVirtualStart),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyScheduleExists("pagerduty_schedule.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "name", scheduleUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "time_zone", location),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.name", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.restriction.0.start_day_of_week", "5"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.start", start),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.rotation_virtual_start", rotationVirtualStart),
				),
			},
		},
	})
}

func TestAccPagerDutySchedule_Multi(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	schedule := fmt.Sprintf("tf-%s", acctest.RandString(5))
	location := "Europe/Berlin"
	start := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)
	end := timeNowInLoc(location).Add(72 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)
	rotationVirtualStart := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyScheduleConfigMulti(username, email, schedule, location, start, rotationVirtualStart, end),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyScheduleExists("pagerduty_schedule.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "name", schedule),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "time_zone", location),

					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.#", "3"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.name", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.restriction.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.restriction.0.duration_seconds", "32101"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.restriction.0.start_time_of_day", "08:00:00"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.rotation_turn_length_seconds", "86400"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.users.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.start", start),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.0.rotation_virtual_start", rotationVirtualStart),

					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.1.name", "bar"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.1.restriction.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.1.restriction.0.duration_seconds", "32101"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.1.restriction.0.start_time_of_day", "08:00:00"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.1.restriction.0.start_day_of_week", "5"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.1.rotation_turn_length_seconds", "86400"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.1.users.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.1.start", start),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.1.end", end),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.1.rotation_virtual_start", rotationVirtualStart),

					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.2.name", "foobar"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.2.restriction.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.2.restriction.0.duration_seconds", "32101"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.2.restriction.0.start_time_of_day", "08:00:00"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.2.restriction.0.start_day_of_week", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.2.rotation_turn_length_seconds", "86400"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.2.users.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.2.start", start),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.2.rotation_virtual_start", rotationVirtualStart),
				),
			},
			{
				Config: testAccCheckPagerDutyScheduleConfigMultiUpdated(username, email, schedule, location, start, rotationVirtualStart, end),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyScheduleExists("pagerduty_schedule.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "name", schedule),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "time_zone", location),
					resource.TestCheckResourceAttr(
						"pagerduty_schedule.foo", "layer.#", "2"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyScheduleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_schedule" {
			continue
		}

		if _, _, err := client.Schedules.Get(r.Primary.ID, &pagerduty.GetScheduleOptions{}); err == nil {
			return fmt.Errorf("Schedule still exists")
		}

	}
	return nil
}

func testAccCheckPagerDutyScheduleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Schedule ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, _, err := client.Schedules.Get(rs.Primary.ID, &pagerduty.GetScheduleOptions{})
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Schedule not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyScheduleConfig(username, email, schedule, location, start, rotationVirtualStart string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name  = "%s"
  email = "%s"
}

resource "pagerduty_schedule" "foo" {
  name = "%s"

  time_zone   = "%s"
  description = "foo"

  layer {
    name                         = "foo"
    start                        = "%s"
    rotation_virtual_start       = "%s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "daily_restriction"
      start_time_of_day = "08:00:00"
      duration_seconds  = 32101
    }
  }
}
`, username, email, schedule, location, start, rotationVirtualStart)
}

func testAccCheckPagerDutyScheduleConfigRestrictionType(username, email, schedule, location, start, rotationVirtualStart string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name  = "%s"
  email = "%s"
}

resource "pagerduty_schedule" "foo" {
  name = "%s"

  time_zone   = "%s"
  description = "foo"

  layer {
    name                         = "foo"
    start                        = "%s"
    rotation_virtual_start       = "%s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "daily_restriction"
      start_time_of_day = "08:00:00"
      duration_seconds  = 32101
	  start_day_of_week = 5
    }
  }
}
`, username, email, schedule, location, start, rotationVirtualStart)
}

func testAccCheckPagerDutyScheduleConfigUpdated(username, email, schedule, location, start, rotationVirtualStart string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
}

resource "pagerduty_schedule" "foo" {
  name = "%s"

  time_zone = "%s"

  layer {
    name                         = "foo"
    start                        = "%s"
    rotation_virtual_start       = "%s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "daily_restriction"
      start_time_of_day = "08:00:00"
      duration_seconds  = 32101
    }
  }
}
`, username, email, schedule, location, start, rotationVirtualStart)
}

func testAccCheckPagerDutyScheduleOverflowConfig(username, email, schedule, location, start, rotationVirtualStart string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name  = "%s"
  email = "%s"
}

resource "pagerduty_schedule" "foo" {
  name      = "%s"
  overflow  = true
  time_zone = "%s"

  layer {
    name                         = "foo"
    start                        = "%s"
    rotation_virtual_start       = "%s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "daily_restriction"
      start_time_of_day = "08:00:00"
      duration_seconds  = 32101
    }
  }
}
`, username, email, schedule, location, start, rotationVirtualStart)
}

func testAccCheckPagerDutyScheduleOverflowConfigUpdated(username, email, schedule, location, start, rotationVirtualStart string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name  = "%s"
  email = "%s"
}

resource "pagerduty_schedule" "foo" {
  name      = "%s"
  overflow  = false
  time_zone = "%s"

  layer {
    name                         = "foo"
    start                        = "%s"
    rotation_virtual_start       = "%s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "daily_restriction"
      start_time_of_day = "08:00:00"
      duration_seconds  = 32101
    }
  }
}
`, username, email, schedule, location, start, rotationVirtualStart)
}

func testAccCheckPagerDutyScheduleConfigWeek(username, email, schedule, location, start, rotationVirtualStart string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name  = "%s"
  email = "%s"
}

resource "pagerduty_schedule" "foo" {
  name = "%s"

  time_zone   = "%s"
  description = "foo"

  layer {
    name                         = "foo"
    start                        = "%s"
    rotation_virtual_start       = "%s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "weekly_restriction"
      start_time_of_day = "08:00:00"
			start_day_of_week = 1
      duration_seconds  = 32101
    }
  }
}
`, username, email, schedule, location, start, rotationVirtualStart)
}

func testAccCheckPagerDutyScheduleConfigWeekUpdated(username, email, schedule, location, start, rotationVirtualStart string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
}

resource "pagerduty_schedule" "foo" {
  name = "%s"

  time_zone = "%s"

  layer {
    name                         = "foo"
    start                        = "%s"
    rotation_virtual_start       = "%s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

		restriction {
      type              = "weekly_restriction"
      start_time_of_day = "08:00:00"
			start_day_of_week = 5
      duration_seconds  = 32101
    }
  }
}
`, username, email, schedule, location, start, rotationVirtualStart)
}

func testAccCheckPagerDutyScheduleConfigMulti(username, email, schedule, location, start, rotationVirtualStart, end string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
}

resource "pagerduty_schedule" "foo" {
  name = "%s"

  time_zone   = "%s"
  description = "foo"

  layer {
    name                         = "foo"
	start                        = "%[5]v"
	end = null
    rotation_virtual_start       = "%[6]v"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "daily_restriction"
      start_time_of_day = "08:00:00"
      duration_seconds  = 32101
    }
  }

  layer {
    name                         = "bar"
	start                        = "%[5]v"
	end							 = "%[7]v"
    rotation_virtual_start       = "%[6]v"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "weekly_restriction"
      start_time_of_day = "08:00:00"
			start_day_of_week = 5
      duration_seconds  = 32101
    }
  }

  layer {
    name                         = "foobar"
	start                        = "%[5]v"
	end = null
    rotation_virtual_start       = "%[6]v"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "weekly_restriction"
      start_time_of_day = "08:00:00"
			start_day_of_week = 1
      duration_seconds  = 32101
    }
  }
}
`, username, email, schedule, location, start, rotationVirtualStart, end)
}

func testAccCheckPagerDutyScheduleConfigMultiUpdated(username, email, schedule, location, start, rotationVirtualStart, end string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
}

resource "pagerduty_schedule" "foo" {
  name = "%s"

  time_zone   = "%s"
  description = "foo"

  layer {
    name                         = "foo"
	start                        = "%[5]v"
	end = null
    rotation_virtual_start       = "%[6]v"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "daily_restriction"
      start_time_of_day = "08:00:00"
      duration_seconds  = 32101
    }
  }

  layer {
    name                         = "bar"
	start                        = "%[5]v"
	end							 = "%[7]v"
    rotation_virtual_start       = "%[6]v"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "weekly_restriction"
      start_time_of_day = "08:00:00"
			start_day_of_week = 5
      duration_seconds  = 32101
    }
  }
}
`, username, email, schedule, location, start, rotationVirtualStart, end)
}

func testAccCheckPagerDutyScheduleWithTeamsConfig(username, email, schedule, location, start, rotationVirtualStart, team string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name  = "%s"
  email = "%s"
}

resource "pagerduty_team" "foo" {
	name = "%s"
	description = "fighters"
}

resource "pagerduty_schedule" "foo" {
  name = "%s"

  time_zone   = "%s"
  description = "foo"
  
  teams = [pagerduty_team.foo.id]

  layer {
    name                         = "foo"
    start                        = "%s"
    rotation_virtual_start       = "%s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "daily_restriction"
      start_time_of_day = "08:00:00"
      duration_seconds  = 32101
    }
  }
}
`, username, email, team, schedule, location, start, rotationVirtualStart)
}
func testAccCheckPagerDutyScheduleWithTeamsConfigUpdated(username, email, schedule, location, start, rotationVirtualStart, team string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name  = "%s"
  email = "%s"
}

resource "pagerduty_team" "foo" {
	name = "%s"
	description = "bar"
}

resource "pagerduty_schedule" "foo" {
  name = "%s"

  time_zone   = "%s"
  description = "Managed by Terraform"
  
  teams = [pagerduty_team.foo.id]

  layer {
    name                         = "foo"
    start                        = "%s"
    rotation_virtual_start       = "%s"
    rotation_turn_length_seconds = 86400
    users                        = [pagerduty_user.foo.id]

    restriction {
      type              = "daily_restriction"
      start_time_of_day = "08:00:00"
      duration_seconds  = 32101
    }
  }
}
`, username, email, team, schedule, location, start, rotationVirtualStart)
}
