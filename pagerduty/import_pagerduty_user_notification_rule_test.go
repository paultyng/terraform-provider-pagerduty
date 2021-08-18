package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPagerDutyUserNotificationRule_import(t *testing.T) {
	contactMethodType := "phone_contact_method"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserNotificationRuleContactMethodConfig(contactMethodType),
			},
			{
				ResourceName:      "pagerduty_user_notification_rule.foo",
				ImportStateIdFunc: testAccCheckPagerDutyUserNotificationRuleId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPagerDutyUserNotificationRuleId(s *terraform.State) (string, error) {
	return fmt.Sprintf("%v:%v", s.RootModule().Resources["pagerduty_user.foo"].Primary.ID, s.RootModule().Resources["pagerduty_user_notification_rule.foo"].Primary.ID), nil
}
