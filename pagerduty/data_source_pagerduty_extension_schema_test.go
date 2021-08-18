package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourcePagerDutyExtensionSchema_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyExtensionSchemaConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourcePagerDutyExtensionSchema("data.pagerduty_extension_schema.foo"),
				),
			},
		},
	})
}

func testAccDataSourcePagerDutyExtensionSchema(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["id"] == "" {
			return fmt.Errorf("Expected to get an Extension Schema  ID from PagerDuty")
		}

		if a["id"] != "PD8SURB" {
			return fmt.Errorf("Expected the Slack Extension Schema ID to be: PD8SURB, but got: %s", a["id"])
		}

		if a["name"] != "Slack" {
			return fmt.Errorf("Expected the Slack Extension Schema Name to be: Slack, but got: %s", a["name"])
		}

		if a["type"] != "extension_schema" {
			return fmt.Errorf("Expected the Slack Extension Schema Type to be: extension_schema, but got: %s", a["type"])
		}

		return nil
	}
}

const testAccDataSourcePagerDutyExtensionSchemaConfig = `
data "pagerduty_extension_schema" "foo" {
  name = "slack"
}
`
