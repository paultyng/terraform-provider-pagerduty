package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPagerDutySlackConnection_import(t *testing.T) {
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
			},
			{
				Config: testAccCheckPagerDutySlackConnectionConfigUpdated(username, email, escalationPolicy, service, workspaceID, channelID),
			},
			{
				ResourceName:      "pagerduty_slack_connection.foo",
				ImportStateIdFunc: testAccCheckPagerDutySlackConnectionID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPagerDutySlackConnectionTeam_import(t *testing.T) {
	team := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutySlackConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutySlackConnectionConfigTeam(team, workspaceID, channelID),
			},
			{
				ResourceName:      "pagerduty_slack_connection.foo",
				ImportStateIdFunc: testAccCheckPagerDutySlackConnectionID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPagerDutySlackConnectionID(s *terraform.State) (string, error) {
	scatts := s.RootModule().Resources["pagerduty_slack_connection.foo"].Primary.Attributes

	return fmt.Sprintf("%v.%v", scatts["workspace_id"], s.RootModule().Resources["pagerduty_slack_connection.foo"].Primary.ID), nil
}
