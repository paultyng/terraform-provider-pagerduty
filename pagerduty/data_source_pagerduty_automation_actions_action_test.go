package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourcePagerDutyAutomationActionsAction_Basic(t *testing.T) {
	name := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyAutomationActionsActionConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourcePagerdutyAutomationActionsAction("pagerduty_automation_actions_action.test", "data.pagerduty_automation_actions_action.foo"),
				),
			},
		},
	})
}

func testAccDataSourcePagerdutyAutomationActionsAction(rName, dsName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		srcR := s.RootModule().Resources[rName]
		srcA := srcR.Primary.Attributes

		ds, ok := s.RootModule().Resources[dsName]
		if !ok {
			return fmt.Errorf("Not found: %s", dsName)
		}
		dsA := ds.Primary.Attributes

		if dsA["id"] == "" {
			return fmt.Errorf("No Action ID is set")
		}

		testAtts := []string{"id", "name", "description", "action_type", "runner_id", "action_data_reference", "type", "action_classification", "runner_type", "creation_time", "modify_time"}

		for _, att := range testAtts {
			if dsA[att] != srcA[att] {
				return fmt.Errorf("Expected the action %s to be: %s, but got: %s", att, srcA[att], dsA[att])
			}
		}

		return nil
	}
}

func testAccDataSourcePagerDutyAutomationActionsActionConfig(actionName string) string {
	return fmt.Sprintf(`
resource "pagerduty_automation_actions_runner" "foo_runner" {
	name = "%s runner"
	description = "Runner created by TF"
	runner_type = "runbook"
	runbook_base_uri = "cat-cat"
	runbook_api_key = "cat-secret"
}

resource "pagerduty_automation_actions_action" "test" {
	name = "%s"
	description = "PA Action created by TF"
	action_type = "process_automation"
	action_classification = "diagnostic"
	runner_id = pagerduty_automation_actions_runner.foo_runner.id
	action_data_reference {
		process_automation_job_id = "pa_job_id_123"
		process_automation_job_arguments = "-arg 1"
		process_automation_node_filter = "tags: production"
	  }
	only_invocable_on_unresolved_incidents = true
	map_to_all_services = true
}

data "pagerduty_automation_actions_action" "foo" {
  id = pagerduty_automation_actions_action.test.id
}
`, actionName, actionName)
}
