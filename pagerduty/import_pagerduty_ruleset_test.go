package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPagerDutyRuleset_import(t *testing.T) {
	ruleset := fmt.Sprintf("tf-%s", acctest.RandString(5))
	teamName := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyRulesetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyRulesetConfig(ruleset, teamName),
			},

			{
				ResourceName:      "pagerduty_ruleset.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPagerDutyRulesetWithNoTeam_import(t *testing.T) {
	ruleset := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyRulesetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyRulesetConfigNoTeam(ruleset),
			},

			{
				ResourceName:      "pagerduty_ruleset.noteam",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
