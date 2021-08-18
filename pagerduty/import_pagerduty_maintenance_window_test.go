package pagerduty

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPagerDutyMaintenanceWindow_import(t *testing.T) {
	window := fmt.Sprintf("tf-%s", acctest.RandString(5))
	windowStartTime := timeNowInAccLoc().Add(24 * time.Hour).Format(time.RFC3339)
	windowEndTime := timeNowInAccLoc().Add(48 * time.Hour).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyTeamMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyMaintenanceWindowConfig(window, windowStartTime, windowEndTime),
			},

			{
				ResourceName:      "pagerduty_maintenance_window.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
