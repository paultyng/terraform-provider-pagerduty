package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPagerDutyUserContactMethod_import(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserContactMethodEmailConfig(username, email),
			},
			{
				ResourceName:      "pagerduty_user_contact_method.foo",
				ImportStateIdFunc: testAccCheckPagerDutyUserContactMethodId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPagerDutyUserContactMethodId(s *terraform.State) (string, error) {
	return fmt.Sprintf("%v:%v", s.RootModule().Resources["pagerduty_user.foo"].Primary.ID, s.RootModule().Resources["pagerduty_user_contact_method.foo"].Primary.ID), nil
}
