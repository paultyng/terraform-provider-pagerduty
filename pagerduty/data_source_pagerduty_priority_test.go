package pagerduty

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourcePagerDutyPriority_Basic(t *testing.T) {
	dataSourceName := "data.pagerduty_priority.p1"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyPriorityConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "name", "P1"),
				),
			},
		},
	})
}
func TestAccDataSourcePagerDutyPriority_P2(t *testing.T) {
	dataSourceName := "data.pagerduty_priority.p2"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyP2Config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "name", "P2"),
				),
			},
		},
	})
}

const testAccDataSourcePagerDutyPriorityConfig = `
data "pagerduty_priority" "p1" {
  name = "p1"
}
`

const testAccDataSourcePagerDutyP2Config = `
data "pagerduty_priority" "p2" {
  name = "p2"
}
`
