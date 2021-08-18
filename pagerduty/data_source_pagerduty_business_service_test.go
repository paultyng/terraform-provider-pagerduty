package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourcePagerDutyBusinessService_Basic(t *testing.T) {
	name := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyBusinessServiceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourcePagerDutyBusinessService("pagerduty_business_service.test", "data.pagerduty_business_service.by_name"),
				),
			},
		},
	})
}

func testAccDataSourcePagerDutyBusinessService(src, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		srcR := s.RootModule().Resources[src]
		srcA := srcR.Primary.Attributes

		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["id"] == "" {
			return fmt.Errorf("Expected to get a business service ID from PagerDuty")
		}

		testAtts := []string{"id", "name"}

		for _, att := range testAtts {
			if a[att] != srcA[att] {
				return fmt.Errorf("Expected the business service %s to be: %s, but got: %s", att, srcA[att], a[att])
			}
		}

		return nil
	}
}

func testAccDataSourcePagerDutyBusinessServiceConfig(name string) string {
	return fmt.Sprintf(`
resource "pagerduty_business_service" "test" {
  name = "%s"
}

data "pagerduty_business_service" "by_name" {
  name = pagerduty_business_service.test.name
}
`, name)
}
