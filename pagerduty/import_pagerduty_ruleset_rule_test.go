package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPagerDutyRulesetRule_import(t *testing.T) {
	ruleset := fmt.Sprintf("tf-%s", acctest.RandString(5))
	teamName := fmt.Sprintf("tf-%s", acctest.RandString(5))
	rule := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyRulesetRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyRulesetRuleConfig(ruleset, teamName, rule),
			},

			{
				ResourceName:      "pagerduty_ruleset_rule.foo",
				ImportStateIdFunc: testAccCheckPagerDutyRulesetRuleID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPagerDutyRulesetRuleID(s *terraform.State) (string, error) {
	return fmt.Sprintf("%v.%v", s.RootModule().Resources["pagerduty_ruleset.foo"].Primary.ID, s.RootModule().Resources["pagerduty_ruleset_rule.foo"].Primary.ID), nil
}
