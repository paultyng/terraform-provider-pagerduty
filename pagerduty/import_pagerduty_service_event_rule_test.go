package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPagerDutyServiceEventRule_import(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))
	rule := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyServiceEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyServiceEventRuleConfig(username, email, escalationPolicy, service, rule),
			},

			{
				ResourceName:      "pagerduty_service_event_rule.foo",
				ImportStateIdFunc: testAccCheckPagerDutyServiceEventRuleID,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPagerDutyServiceEventRuleID(s *terraform.State) (string, error) {
	return fmt.Sprintf("%v.%v", s.RootModule().Resources["pagerduty_service.foo"].Primary.ID, s.RootModule().Resources["pagerduty_service_event_rule.foo"].Primary.ID), nil
}
