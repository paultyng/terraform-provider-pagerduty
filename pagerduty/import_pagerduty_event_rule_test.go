package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPagerDutyEventRule_import(t *testing.T) {
	eventRule := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyEventRuleConfig(eventRule),
			},

			{
				ResourceName:      "pagerduty_event_rule.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
