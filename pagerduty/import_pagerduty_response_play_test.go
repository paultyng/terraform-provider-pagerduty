package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPagerDutyResponsePlay_import(t *testing.T) {
	name := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyResponsePlayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyResponsePlayConfig(name),
			},

			{
				ResourceName:      "pagerduty_response_play.foo",
				ImportStateIdFunc: testAccCheckPagerDutyResponsePlayID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPagerDutyResponsePlayID(s *terraform.State) (string, error) {
	ua := s.RootModule().Resources["pagerduty_response_play.foo"].Primary.Attributes

	return fmt.Sprintf("%v.%v", s.RootModule().Resources["pagerduty_response_play.foo"].Primary.ID, ua["from"]), nil
}
