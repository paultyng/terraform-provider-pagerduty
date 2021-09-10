package pagerduty

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func init() {
	resource.AddTestSweepers("pagerduty_ruleset", &resource.Sweeper{
		Name: "pagerduty_ruleset",
		F:    testSweepRuleset,
	})
}

func testSweepRuleset(region string) error {
	config, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client, err := config.Client()
	if err != nil {
		return err
	}

	resp, _, err := client.Rulesets.List()
	if err != nil {
		return err
	}

	for _, ruleset := range resp.Rulesets {
		if strings.HasPrefix(ruleset.Name, "test") || strings.HasPrefix(ruleset.Name, "tf-") {
			log.Printf("Destroying ruleset %s (%s)", ruleset.Name, ruleset.ID)
			if _, err := client.Rulesets.Delete(ruleset.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccPagerDutyRuleset_Basic(t *testing.T) {
	ruleset := fmt.Sprintf("tf-%s", acctest.RandString(5))
	rulesetUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	teamName := fmt.Sprintf("tf-%s", acctest.RandString(5))
	teamNameUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyRulesetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyRulesetConfig(ruleset, teamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyRulesetExists("pagerduty_ruleset.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_ruleset.foo", "name", ruleset),
				),
			},
			{
				Config: testAccCheckPagerDutyRulesetConfigUpdated(rulesetUpdated, teamNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyRulesetExists("pagerduty_ruleset.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_ruleset.foo", "name", rulesetUpdated),
				),
			},
			{
				Config: testAccCheckPagerDutyRulesetConfigNoTeam(ruleset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyRulesetExists("pagerduty_ruleset.noteam"),
					resource.TestCheckResourceAttr(
						"pagerduty_ruleset.noteam", "name", ruleset),
				),
			},
		},
	})
}

func testAccCheckPagerDutyRulesetDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_ruleset" {
			continue
		}
		if _, _, err := client.Rulesets.Get(r.Primary.ID); err == nil {
			return fmt.Errorf("Ruleset still exists")
		}
	}
	return nil
}

func testAccCheckPagerDutyRulesetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Ruleset ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)
		found, _, err := client.Rulesets.Get(rs.Primary.ID)
		if err != nil {
			return err
		}
		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Ruleset not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyRulesetConfig(rulesetName, team string) string {
	return fmt.Sprintf(`
resource "pagerduty_team" "foo" {
	name = "%s"
}

resource "pagerduty_ruleset" "foo" {
	name = "%s"
	team { 
		id = pagerduty_team.foo.id
	}
}
`, team, rulesetName)
}

func testAccCheckPagerDutyRulesetConfigUpdated(rulesetName, team string) string {
	return fmt.Sprintf(`

resource "pagerduty_team" "foo" {
	name = "%s"
}
resource "pagerduty_ruleset" "foo" {
	name = "%s"
	team {
		id = pagerduty_team.foo.id
	}
}
`, team, rulesetName)
}
func testAccCheckPagerDutyRulesetConfigNoTeam(rulesetName string) string {
	return fmt.Sprintf(`

resource "pagerduty_ruleset" "noteam" {
	name = "%s"
}
`, rulesetName)
}
