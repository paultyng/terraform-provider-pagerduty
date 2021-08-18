package pagerduty

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

var (
	/* workspaceID and channelID must be valid IDs from a Slack workspace connected
	to the PagerDuty account running these tests. For these tests value for workspaceID
	is taken from the SLACK_CONNECTION_WORKSPACE_ID environment variable */
	channelID   string = "C02CLUSDAC9"
	workspaceID string = "T02ADG9LV1A"
)

func TestAccPagerDutySlackConnection_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutySlackConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutySlackConnectionConfig(username, email, escalationPolicy, service, workspaceID, channelID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutySlackConnectionExists("pagerduty_slack_connection.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_slack_connection.foo", "source_name", service),
					resource.TestCheckResourceAttr(
						"pagerduty_slack_connection.foo", "config.0.events.#", "13"),
				),
			},
			{
				Config: testAccCheckPagerDutySlackConnectionConfigUpdated(username, email, escalationPolicy, service, workspaceID, channelID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutySlackConnectionExists("pagerduty_slack_connection.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_slack_connection.foo", "config.0.urgency", ""),
				),
			},
		},
	})
}

func TestAccPagerDutySlackConnection_Team(t *testing.T) {
	team := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutySlackConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutySlackConnectionConfigTeam(team, workspaceID, channelID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutySlackConnectionExists("pagerduty_slack_connection.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_slack_connection.foo", "source_name", team),
					resource.TestCheckResourceAttr(
						"pagerduty_slack_connection.foo", "config.0.events.#", "13"),
				),
			},
			{
				Config: testAccCheckPagerDutySlackConnectionConfigTeamUpdated(team, workspaceID, channelID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutySlackConnectionExists("pagerduty_slack_connection.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_slack_connection.foo", "config.0.urgency", "low"),
				),
			},
		},
	})
}

func TestAccPagerDutySlackConnection_Envar(t *testing.T) {
	team := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutySlackConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutySlackConnectionConfigEnvar(team, channelID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutySlackConnectionExists("pagerduty_slack_connection.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_slack_connection.foo", "workspace_id", os.Getenv("SLACK_CONNECTION_WORKSPACE_ID")),
				),
			},
		},
	})
}

func testAccCheckPagerDutySlackConnectionDestroy(s *terraform.State) error {
	config := &pagerduty.Config{
		Token:   os.Getenv("PAGERDUTY_USER_TOKEN"),
		BaseURL: "https://app.pagerduty.com",
	}
	client, err := pagerduty.NewClient(config)
	if err != nil {
		return err
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_slack_connection" {
			continue
		}

		scatts := r.Primary.Attributes
		if _, _, err := client.SlackConnections.Get(scatts["workspace_id"], r.Primary.ID); err == nil {
			return fmt.Errorf("response play still exists")
		}

	}
	return nil
}

func testAccCheckPagerDutySlackConnectionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		sc, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if sc.Primary.ID == "" {
			return fmt.Errorf("No slack connection ID is set")
		}
		log.Printf("[DEBUG] Slack Connection EXISTS: %v", sc.Primary.ID)

		config := &pagerduty.Config{
			Token:   os.Getenv("PAGERDUTY_USER_TOKEN"),
			BaseURL: "https://app.pagerduty.com",
		}
		client, err := pagerduty.NewClient(config)
		if err != nil {
			return err
		}

		scatts := sc.Primary.Attributes
		found, _, err := client.SlackConnections.Get(scatts["workspace_id"], sc.Primary.ID)
		if err != nil {
			return err
		}

		if found.ID != sc.Primary.ID {
			return fmt.Errorf("slack connection not found: %v - %v", sc.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutySlackConnectionConfig(username, useremail, escalationPolicy, service, workspaceID, channelID string) string {
	return fmt.Sprintf(`
	resource "pagerduty_user" "foo" {
		name        = "%s"  
		email       = "%s"
	}
	
	resource "pagerduty_escalation_policy" "foo" {
		name        = "%s"
		description = "foo"
		num_loops   = 1
		
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
		
		incident_urgency_rule {
			type = "constant"
			urgency = "high"
		}
	}
	data "pagerduty_priority" "p1" {
		name = "P1"
	}
	resource "pagerduty_slack_connection" "foo" {
		source_id = pagerduty_service.foo.id
		source_type = "service_reference"
		workspace_id = "%s"
		channel_id = "%s"
		notification_type = "responder"
		config {
			events = [
				"incident.triggered",
				"incident.acknowledged",
				"incident.escalated",
				"incident.resolved",
				"incident.reassigned",
				"incident.annotated",
				"incident.unacknowledged",
				"incident.delegated",
				"incident.priority_updated",
				"incident.responder.added",
				"incident.responder.replied",
				"incident.status_update_published",
				"incident.reopened"		  
			]
			priorities = [data.pagerduty_priority.p1.id]
			urgency = "high"
		}
	}
	`, username, useremail, escalationPolicy, service, workspaceID, channelID)
}

func testAccCheckPagerDutySlackConnectionConfigUpdated(username, email, escalationPolicy, service, workspaceID, channelID string) string {
	return fmt.Sprintf(`
	resource "pagerduty_user" "foo" {
		name        = "%s"  
		email       = "%s"
	}
	
	resource "pagerduty_escalation_policy" "foo" {
		name        = "%s"
		description = "foo"
		num_loops   = 1
		
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
		
		incident_urgency_rule {
			type = "constant"
			urgency = "high"
		}
	}
	data "pagerduty_priority" "p1" {
		name = "P1"
	}
	resource "pagerduty_slack_connection" "foo" {
		source_id = pagerduty_service.foo.id
		source_type = "service_reference"
		workspace_id = "%s"
		channel_id = "%s"
		notification_type = "stakeholder"
		config {
			events = [
				"incident.triggered",
				"incident.acknowledged",
				"incident.escalated",
				"incident.resolved",
				"incident.reassigned",
				"incident.annotated",
				"incident.unacknowledged",
				"incident.delegated",
				"incident.priority_updated",
				"incident.responder.added",
				"incident.responder.replied",
				"incident.status_update_published",
				"incident.reopened"		  
			]
			priorities = [data.pagerduty_priority.p1.id]
		}
	}
	`, username, email, escalationPolicy, service, workspaceID, channelID)
}

func testAccCheckPagerDutySlackConnectionConfigTeam(team, workspaceID, channelID string) string {
	return fmt.Sprintf(`
		resource "pagerduty_team" "foo" {
			name = "%s"
		}
		resource "pagerduty_slack_connection" "foo" {
			source_id = pagerduty_team.foo.id
			source_type = "team_reference"
			workspace_id = "%s"
			channel_id = "%s"
			notification_type = "responder"
			config {
				events = [
					"incident.triggered",
					"incident.acknowledged",
					"incident.escalated",
					"incident.resolved",
					"incident.reassigned",
					"incident.annotated",
					"incident.unacknowledged",
					"incident.delegated",
					"incident.priority_updated",
					"incident.responder.added",
					"incident.responder.replied",
					"incident.status_update_published",
					"incident.reopened"		  
				]
			}
		}
		`, team, workspaceID, channelID)
}
func testAccCheckPagerDutySlackConnectionConfigTeamUpdated(team, workspaceID, channelID string) string {
	return fmt.Sprintf(`
		resource "pagerduty_team" "foo" {
			name = "%s"
		}
		resource "pagerduty_slack_connection" "foo" {
			source_id = pagerduty_team.foo.id
			source_type = "team_reference"
			workspace_id = "%s"
			channel_id = "%s"
			notification_type = "responder"
			config {
				events = [
					"incident.triggered",
					"incident.acknowledged",
					"incident.escalated",
					"incident.resolved",
					"incident.reassigned",
					"incident.annotated",
					"incident.unacknowledged",
					"incident.delegated",
					"incident.priority_updated",
					"incident.responder.added",
					"incident.responder.replied",
					"incident.status_update_published",
					"incident.reopened"		  
				]
				urgency = "low"
			}
		}
		`, team, workspaceID, channelID)
}

func testAccCheckPagerDutySlackConnectionConfigEnvar(team, channelID string) string {
	return fmt.Sprintf(`
		resource "pagerduty_team" "foo" {
			name = "%s"
		}
		resource "pagerduty_slack_connection" "foo" {
			source_id = pagerduty_team.foo.id
			source_type = "team_reference"
			channel_id = "%s"
			notification_type = "responder"
			config {
				events = [
					"incident.triggered",
					"incident.acknowledged",
					"incident.escalated",
					"incident.resolved",
					"incident.reassigned",
					"incident.annotated",
					"incident.unacknowledged",
					"incident.delegated",
					"incident.priority_updated",
					"incident.responder.added",
					"incident.responder.replied",
					"incident.status_update_published",
					"incident.reopened"		  
				]
				urgency = "low"
			}
		}
		`, team, channelID)
}
