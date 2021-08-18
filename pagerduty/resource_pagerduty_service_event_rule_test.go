package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func TestAccPagerDutyServiceEventRule_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))
	rule := fmt.Sprintf("tf-%s", acctest.RandString(5))
	ruleUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyServiceEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyServiceEventRuleConfig(username, email, escalationPolicy, service, rule),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceEventRuleExists("pagerduty_service_event_rule.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "position", "0"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "disabled", "true"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.operator", "and"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.subconditions.0.operator", "contains"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.subconditions.0.parameter.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.subconditions.0.parameter.0.value", "disk space"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "actions.0.annotate.0.value", rule),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "actions.0.extractions.1.template", "Overriding Summary"),
				),
			},
			{
				Config: testAccCheckPagerDutyServiceEventRuleConfigUpdated(username, email, escalationPolicy, service, ruleUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceEventRuleExists("pagerduty_service_event_rule.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "position", "0"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "disabled", "true"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.operator", "and"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.subconditions.0.operator", "contains"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.subconditions.0.parameter.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.subconditions.0.parameter.0.path", "summary"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "actions.0.annotate.0.value", ruleUpdated),
				),
			},
		},
	})
}

func TestAccPagerDutyServiceEventRule_MultipleRules(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))
	rule1 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	rule2 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	rule3 := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyServiceEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyServiceEventRuleConfigMultipleRules(username, email, escalationPolicy, service, rule1, rule2, rule3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceEventRuleExists("pagerduty_service_event_rule.foo"),
					testAccCheckPagerDutyServiceEventRuleExists("pagerduty_service_event_rule.bar"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "position", "0"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.bar", "position", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.baz", "position", "2"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "disabled", "true"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.operator", "and"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.subconditions.0.operator", "contains"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.subconditions.0.parameter.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "conditions.0.subconditions.0.parameter.0.value", "disk space"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.foo", "actions.0.annotate.0.value", rule1),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.bar", "actions.0.annotate.0.value", rule2),
					resource.TestCheckResourceAttr(
						"pagerduty_service_event_rule.baz", "actions.0.annotate.0.value", rule3),
				),
			},
		},
	})
}

func testAccCheckPagerDutyServiceEventRuleDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_service_event_rule" {
			continue
		}

		service, _ := s.RootModule().Resources["pagerduty_service.foo"]

		if _, _, err := client.Services.GetEventRule(service.Primary.ID, r.Primary.ID); err == nil {
			return fmt.Errorf("ServiceEvent Rule still exists")
		}
	}
	return nil
}

func testAccCheckPagerDutyServiceEventRuleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ServiceEvent Rule ID is set")
		}

		service, _ := s.RootModule().Resources["pagerduty_service.foo"]

		client := testAccProvider.Meta().(*pagerduty.Client)
		found, _, err := client.Services.GetEventRule(service.Primary.ID, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("ServiceEvent Rule not found: %v", rs.Primary.ID)
		}
		if found.ID != rs.Primary.ID {
			return fmt.Errorf("ServiceEvent Rule not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyServiceEventRuleConfig(username, email, escalationPolicy, service, rule string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
	name        = "%s"
	email       = "%s"
	color       = "green"
	role        = "user"
	job_title   = "foo"
	description = "foo"
}

resource "pagerduty_escalation_policy" "foo" {
	name        = "%s"
	description = "bar"
	num_loops   = 2
	rule {
		escalation_delay_in_minutes = 10
		target {
			type = "user_reference"
			id   = pagerduty_user.foo.id
		}
	}
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id
	alert_creation          = "create_alerts_and_incidents"
}

resource "pagerduty_service_event_rule" "foo" {
	service = pagerduty_service.foo.id
	position = 0
	disabled = true
	conditions {
		operator = "and"
		subconditions {
			operator = "contains"
			parameter {
				value = "disk space"
				path = "summary"
			}
		}
	}
	actions {
		annotate {
			value = "%s"
		}
		extractions {
			target = "dedup_key"
			source = "source"
			regex = "(.*)"
		}
		extractions {
			target   = "summary"
			template = "Overriding Summary"
		}
	}
}
`, username, email, escalationPolicy, service, rule)
}

func testAccCheckPagerDutyServiceEventRuleConfigUpdated(username, email, escalationPolicy, service, ruleUpdated string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
	name        = "%s"
	email       = "%s"
	color       = "green"
	role        = "user"
	job_title   = "foo"
	description = "foo"
}

resource "pagerduty_escalation_policy" "foo" {
	name        = "%s"
	description = "bar"
	num_loops   = 2
	rule {
		escalation_delay_in_minutes = 10
		target {
			type = "user_reference"
			id   = pagerduty_user.foo.id
		}
	}
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id
	alert_creation          = "create_alerts_and_incidents"
}

resource "pagerduty_service_event_rule" "foo" {
	service = pagerduty_service.foo.id
	position = 0
	disabled = true
	time_frame {
		scheduled_weekly {
			weekdays = [3,7]
			timezone = "America/Los_Angeles"
			start_time = "1000000"
			duration = "3600000"

		}
	}
	conditions {
		operator = "and"
		subconditions {
			operator = "contains"
			parameter {
				value = "disk space"
				path = "summary"
			}
		}
	}
	actions {
		severity  {
			value = "warning"
		}
		annotate {
			value = "%s"
		}
		extractions {
			target = "dedup_key"
			source = "source"
			regex = "(.*)"
		}
	}
}
`, username, email, escalationPolicy, service, ruleUpdated)
}

func testAccCheckPagerDutyServiceEventRuleConfigMultipleRules(username, email, escalationPolicy, service, rule1, rule2, rule3 string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
	name        = "%s"
	email       = "%s"
	color       = "green"
	role        = "user"
	job_title   = "foo"
	description = "foo"
}

resource "pagerduty_escalation_policy" "foo" {
	name        = "%s"
	description = "bar"
	num_loops   = 2
	rule {
		escalation_delay_in_minutes = 10
		target {
			type = "user_reference"
			id   = pagerduty_user.foo.id
		}
	}
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id
	alert_creation          = "create_alerts_and_incidents"
}

resource "pagerduty_service_event_rule" "foo" {
	service = pagerduty_service.foo.id
	disabled = true
	conditions {
		operator = "and"
		subconditions {
			operator = "contains"
			parameter {
				value = "disk space"
				path = "summary"
			}
		}
	}
	actions {
		annotate {
			value = "%s"
		}
		extractions {
			target = "dedup_key"
			source = "source"
			regex = "(.*)"
		}
	}
}

resource "pagerduty_service_event_rule" "bar" {
	service = pagerduty_service.foo.id
	position = 1
	disabled = true
	conditions {
		operator = "and"
		subconditions {
			operator = "contains"
			parameter {
				value = "cpu spike"
				path = "summary"
			}
		}
	}
	actions {
		annotate {
			value = "%s"
		}
	}
}
resource "pagerduty_service_event_rule" "baz" {
	service = pagerduty_service.foo.id
	position = 2
	disabled = true
	conditions {
		operator = "and"
		subconditions {
			operator = "contains"
			parameter {
				value = "slow ping"
				path = "summary"
			}
		}
	}
	actions {
		annotate {
			value = "%s"
		}
	}
}
`, username, email, escalationPolicy, service, rule1, rule2, rule3)
}
