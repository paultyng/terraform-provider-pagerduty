package pagerduty

import (
	"context"
	"fmt"
	"testing"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourcePagerDutyExtensionSchema_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		CheckDestroy:             testAccCheckPagerDutyScheduleDestroy,
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

		if a["id"] != "PAKM60Z" {
			return fmt.Errorf("Expected Schema ID to be: PAKM60Z, but got: %s", a["id"])
		}

		if a["name"] != "ServiceNow (v7)" {
			return fmt.Errorf("Expected Schema Name to be: ServiceNow (v7), but got: %s", a["name"])
		}

		if a["type"] != "extension_schema" {
			return fmt.Errorf("Expected the Schema Type to be: extension_schema, but got: %s", a["type"])
		}

		return nil
	}
}

const testAccDataSourcePagerDutyExtensionSchemaConfig = `
data "pagerduty_extension_schema" "foo" {
  name = "ServiceNow (v7)"
}
`

func testAccCheckPagerDutyScheduleDestroy(s *terraform.State) error {
	ctx := context.Background()

	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_schedule" {
			continue
		}

		_, err := testAccProvider.client.GetScheduleWithContext(ctx, r.Primary.ID, pagerduty.GetScheduleOptions{})
		if err == nil {
			return fmt.Errorf("Schedule still exists")
		}
	}

	return nil
}
