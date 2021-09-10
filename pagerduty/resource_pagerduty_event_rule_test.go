package pagerduty

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func init() {
	resource.AddTestSweepers("pagerduty_event_rule", &resource.Sweeper{
		Name: "pagerduty_event_rule",
		F:    testSweepEventRule,
		Dependencies: []string{
			"pagerduty_service",
		},
	})
}

func testSweepEventRule(region string) error {
	config, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client, err := config.Client()
	if err != nil {
		return err
	}

	resp, _, err := client.EventRules.List()
	if err != nil {
		return err
	}

	for _, rule := range resp.EventRules {
		log.Printf("Destroying event rule %s", rule.ID)
		if _, err := client.EventRules.Delete(rule.ID); err != nil {
			log.Printf("[ERROR] Failed to delete event rule: %s", err)
		}
	}

	return nil
}

func TestAccPagerDutyEventRule_Basic(t *testing.T) {
	eventRule := fmt.Sprintf("tf-%s", acctest.RandString(5))
	eventRuleUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyEventRuleConfig(eventRule),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyEventRuleExists("pagerduty_event_rule.foo"),
				),
			},

			{
				Config: testAccCheckPagerDutyEventRuleConfigUpdated(eventRuleUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyEventRuleExists("pagerduty_event_rule.foo-update"),
					resource.TestCheckNoResourceAttr("pagerduty_event_rule.foo-update", "advanced_condition_json"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyEventRuleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_event_rule" {
			continue
		}
		// get list of event_rules and then check that list.
		resp, _, err := client.EventRules.List()
		if err != nil {
			return err
		}
		for _, er := range resp.EventRules {
			if er.ID == r.Primary.ID {
				return fmt.Errorf("Event Rule still exists")
			}
		}
	}
	return nil
}

func testAccCheckPagerDutyEventRuleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Event Rule ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)
		resp, _, err := client.EventRules.List()
		if err != nil {
			return err
		}
		var found *pagerduty.EventRule

		for _, rule := range resp.EventRules {
			if rule.ID == rs.Primary.ID {
				found = rule
			}
		}
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Event rule not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyEventRuleConfig(eventRule string) string {
	return fmt.Sprintf(`
resource "pagerduty_event_rule" "foo" {
	action_json = jsonencode([["route","P5DTL0K"],["severity","warning"],["annotate","%s"],["priority","PL451DT"]])
	condition_json = jsonencode(["and",["contains",["path","payload","source"],"website"]])
	advanced_condition_json = jsonencode([
		[
			"scheduled-weekly",
			1565392127032,
			3600000,
			"America/Los_Angeles",
			[
				1,
				3,
				5,
				7
			]
		]
	])
}
`, eventRule)
}

func testAccCheckPagerDutyEventRuleConfigUpdated(eventRule string) string {
	return fmt.Sprintf(`
resource "pagerduty_event_rule" "foo-update" {
	action_json = jsonencode([["route","P5DTL0K"],["severity","warning"],["annotate","%s"],["priority","PL451DT"]])
	condition_json = jsonencode(["and",["contains",["path","payload","source"],"website"]])
}
`, eventRule)
}
