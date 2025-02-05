package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func init() {
	resource.AddTestSweepers("automation_actions_action", &resource.Sweeper{
		Name: "automation_actions_action",
		F:    testSweepAutomationActionsAction,
	})
}

func testSweepAutomationActionsAction(region string) error {
	return nil
}

func TestAccPagerDutyAutomationActionsActionTypeProcessAutomation_Basic(t *testing.T) {
	actionName := fmt.Sprintf("tf-%s", acctest.RandString(5))

	nameUpdated := fmt.Sprintf("tf-update-%s", acctest.RandString(5))
	descriptionUpdated := fmt.Sprintf("Description updated tf-%s", acctest.RandString(5))
	actionClassificationUpdated := "remediation"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyAutomationActionsActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyAutomationActionsActionTypeProcessAutomationConfig(actionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyAutomationActionsActionExists("pagerduty_automation_actions_action.foo"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "name", actionName),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "action_type", "process_automation"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "description", "PA Action created by TF"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "type", "action"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "action_classification", "diagnostic"),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_job_id", "pa_job_id_123"),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_job_arguments", "-arg 1"),
					// Known defect with inconsistent handling of nested aggregates: https://github.com/hashicorp/terraform-plugin-sdk/issues/413
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_node_filter", "tags: production"),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.script", ""),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.invocation_command", ""),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "id"),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "creation_time"),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "modify_time"),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "runner_id"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "runner_type", "runbook"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "only_invocable_on_unresolved_incidents", "true"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "map_to_all_services", "true"),
				),
			},
			{
				Config: testAccCheckPagerDutyAutomationActionsActionTypeProcessAutomationConfigUpdated(actionName, nameUpdated, descriptionUpdated, actionClassificationUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyAutomationActionsActionExists("pagerduty_automation_actions_action.foo"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "name", nameUpdated),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "action_type", "process_automation"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "description", descriptionUpdated),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "type", "action"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "action_classification", actionClassificationUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_job_id", "updated_pa_job_id_123"),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_job_arguments", ""),
					// Known defect with inconsistent handling of nested aggregates: https://github.com/hashicorp/terraform-plugin-sdk/issues/413
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_node_filter", ""),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.script", ""),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.invocation_command", ""),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "id"),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "creation_time"),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "modify_time"),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "runner_id"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "runner_type", "runbook"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "only_invocable_on_unresolved_incidents", "false"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "map_to_all_services", "false"),
				),
			},
		},
	})
}

func TestAccPagerDutyAutomationActionsActionTypeScript_Basic(t *testing.T) {
	actionName := fmt.Sprintf("tf-%s", acctest.RandString(5))

	nameUpdated := fmt.Sprintf("tf-update-%s", acctest.RandString(5))
	descriptionUpdated := fmt.Sprintf("Description updated tf-%s", acctest.RandString(5))
	actionClassificationUpdated := "remediation"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyAutomationActionsActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyAutomationActionsActionTypeScriptConfig(actionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyAutomationActionsActionExists("pagerduty_automation_actions_action.foo"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "name", actionName),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "action_type", "script"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "description", "PA Action created by TF"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "type", "action"),
					resource.TestCheckNoResourceAttr("pagerduty_automation_actions_action.foo", "action_classification"),
					// Known defect with inconsistent handling of nested aggregates: https://github.com/hashicorp/terraform-plugin-sdk/issues/413
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_job_id", ""),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_job_arguments", ""),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_node_filter", ""),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.script", "java --version"),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.invocation_command", "/bin/bash"),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "id"),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "creation_time"),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "modify_time"),
					resource.TestCheckNoResourceAttr("pagerduty_automation_actions_action.foo", "runner_type"),
					resource.TestCheckNoResourceAttr("pagerduty_automation_actions_action.foo", "runner_id"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "only_invocable_on_unresolved_incidents", "false"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "map_to_all_services", "false"),
				),
			},
			{
				Config: testAccCheckPagerDutyAutomationActionsActionTypeScriptConfigUpdated(nameUpdated, descriptionUpdated, actionClassificationUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyAutomationActionsActionExists("pagerduty_automation_actions_action.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "name", nameUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "description", descriptionUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_classification", actionClassificationUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_job_id", ""),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_job_arguments", ""),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.process_automation_node_filter", ""),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.script", "echo 777"),
					resource.TestCheckResourceAttr(
						"pagerduty_automation_actions_action.foo", "action_data_reference.0.invocation_command", ""),
					resource.TestCheckResourceAttrSet("pagerduty_automation_actions_action.foo", "modify_time"),
					resource.TestCheckNoResourceAttr("pagerduty_automation_actions_action.foo", "runner_type"),
					resource.TestCheckNoResourceAttr("pagerduty_automation_actions_action.foo", "runner_id"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "only_invocable_on_unresolved_incidents", "false"),
					resource.TestCheckResourceAttr("pagerduty_automation_actions_action.foo", "map_to_all_services", "false"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyAutomationActionsActionDestroy(s *terraform.State) error {
	client, _ := testAccProvider.Meta().(*Config).Client()
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_automation_actions_action" {
			continue
		}
		if _, _, err := client.AutomationActionsAction.Get(r.Primary.ID); err == nil {
			return fmt.Errorf("Automation Actions Action still exists")
		}
	}
	return nil
}

func testAccCheckPagerDutyAutomationActionsActionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Automation Actions Action ID is set")
		}

		client, _ := testAccProvider.Meta().(*Config).Client()
		found, _, err := client.AutomationActionsAction.Get(rs.Primary.ID)
		if err != nil {
			return err
		}
		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Automation Actions Action not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyAutomationActionsActionTypeProcessAutomationConfig(actionName string) string {
	return fmt.Sprintf(`
resource "pagerduty_automation_actions_runner" "foo_runner" {
	name = "%s runner"
	description = "Runner created by TF"
	runner_type = "runbook"
	runbook_base_uri = "cat-cat"
	runbook_api_key = "cat-secret"
}

resource "pagerduty_automation_actions_action" "foo" {
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
	only_invocable_on_unresolved_incidents = "true"
	map_to_all_services = "true"
}
`, actionName, actionName)
}

func testAccCheckPagerDutyAutomationActionsActionTypeProcessAutomationConfigUpdated(previousActionName, actionName, actionDescription, actionClassification string) string {
	return fmt.Sprintf(`

resource "pagerduty_automation_actions_runner" "foo_runner" {
	name = "%s runner"
	description = "Runner created by TF"
	runner_type = "runbook"
	runbook_base_uri = "cat-cat"
	runbook_api_key = "cat-secret"
}

resource "pagerduty_automation_actions_action" "foo" {
	name = "%s"
	description = "%s"
	action_type = "process_automation"
	action_classification = "%s"
	runner_id = pagerduty_automation_actions_runner.foo_runner.id
	action_data_reference {
		process_automation_job_id = "updated_pa_job_id_123"
	}
	only_invocable_on_unresolved_incidents = "false"
	map_to_all_services = "false"
}
`, previousActionName, actionName, actionDescription, actionClassification)
}

func testAccCheckPagerDutyAutomationActionsActionTypeScriptConfig(actionName string) string {
	return fmt.Sprintf(`
resource "pagerduty_automation_actions_action" "foo" {
	name = "%s"
	description = "PA Action created by TF"
	action_type = "script"
	action_data_reference {
		script = "java --version"
		invocation_command = "/bin/bash"
	  }
}
`, actionName)
}

func testAccCheckPagerDutyAutomationActionsActionTypeScriptConfigUpdated(actionName, actionDescription, actionClassification string) string {
	return fmt.Sprintf(`
	resource "pagerduty_automation_actions_action" "foo" {
		name = "%s"
		description = "%s"
		action_type = "script"
		action_classification = "%s"
		action_data_reference {
			script = "echo 777"
		  }
	}
`, actionName, actionDescription, actionClassification)
}
