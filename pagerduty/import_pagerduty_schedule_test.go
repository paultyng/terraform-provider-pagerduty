package pagerduty

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPagerDutySchedule_import(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	schedule := fmt.Sprintf("tf-%s", acctest.RandString(5))
	location := "Europe/Berlin"
	start := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)
	rotationVirtualStart := timeNowInLoc(location).Add(24 * time.Hour).Round(1 * time.Hour).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyScheduleConfig(username, email, schedule, location, start, rotationVirtualStart),
			},

			{
				ResourceName:      "pagerduty_schedule.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
