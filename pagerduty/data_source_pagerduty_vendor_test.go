package pagerduty

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourcePagerDutyVendor_Basic(t *testing.T) {
	dataSourceName := "data.pagerduty_vendor.foo"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyVendorConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", "PZQ6AUS"),
					resource.TestCheckResourceAttr(dataSourceName, "name", "Amazon CloudWatch"),
				),
			},
		},
	})
}

func TestAccDataSourcePagerDutyVendor_ExactMatch(t *testing.T) {
	dataSourceName := "data.pagerduty_vendor.foo"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyExactMatchConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", "PKAPG94"),
					resource.TestCheckResourceAttr(dataSourceName, "name", "Sentry"),
				),
			},
		},
	})
}

func TestAccDataSourcePagerDutyVendor_SpecialChars(t *testing.T) {
	dataSourceName := "data.pagerduty_vendor.foo"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutySpecialCharsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", "PRYWPH4"),
					resource.TestCheckResourceAttr(dataSourceName, "name", "Slack to PagerDuty (Legacy)"),
				),
			},
		},
	})
}

const testAccDataSourcePagerDutyVendorConfig = `
data "pagerduty_vendor" "foo" {
  name = "cloudwatch"
}
`

const testAccDataSourcePagerDutyExactMatchConfig = `
data "pagerduty_vendor" "foo" {
  name = "sentry"
}
`

const testAccDataSourcePagerDutySpecialCharsConfig = `
data "pagerduty_vendor" "foo" {
  name = "Slack to PagerDuty (Legacy)"
}
`
