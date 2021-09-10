package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccPagerDutyServiceIntegration_import(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))
	serviceIntegration := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyServiceIntegrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyServiceIntegrationConfig(username, email, escalationPolicy, service, serviceIntegration),
			},

			{
				ResourceName:      "pagerduty_service_integration.foo",
				ImportStateIdFunc: testAccCheckPagerDutyServiceIntegrationId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPagerDutyServiceIntegrationId(s *terraform.State) (string, error) {
	return fmt.Sprintf("%v.%v", s.RootModule().Resources["pagerduty_service.foo"].Primary.ID, s.RootModule().Resources["pagerduty_service_integration.foo"].Primary.ID), nil
}
