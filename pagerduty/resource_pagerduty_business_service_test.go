package pagerduty

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func init() {
	resource.AddTestSweepers("pagerduty_business_service", &resource.Sweeper{
		Name: "pagerduty_business_service",
		F:    testSweepBusinessService,
	})
}

func testSweepBusinessService(region string) error {
	config, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client, err := config.Client()
	if err != nil {
		return err
	}

	resp, _, err := client.BusinessServices.List()
	if err != nil {
		return err
	}

	for _, businessService := range resp.BusinessServices {
		if strings.HasPrefix(businessService.Name, "test") || strings.HasPrefix(businessService.Name, "tf-") {
			log.Printf("Destroying business service %s (%s)", businessService.Name, businessService.ID)
			if _, err := client.BusinessServices.Delete(businessService.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccPagerDutyBusinessService_Basic(t *testing.T) {
	name := fmt.Sprintf("tf-%s", acctest.RandString(5))
	description := fmt.Sprintf("tf-%s", acctest.RandString(5))
	pointOfContact := fmt.Sprintf("tf-%s", acctest.RandString(5))
	nameUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	descriptionUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	pointOfContactUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyBusinessServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyBusinessServiceConfig(name, description, pointOfContact),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyBusinessServiceExists("pagerduty_business_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_business_service.foo", "name", name),
					resource.TestCheckResourceAttr(
						"pagerduty_business_service.foo", "description", description),
					resource.TestCheckResourceAttr(
						"pagerduty_business_service.foo", "point_of_contact", pointOfContact),
					resource.TestCheckResourceAttrSet(
						"pagerduty_business_service.foo", "self"),
				),
			},
			{
				Config: testAccCheckPagerDutyBusinessServiceConfigUpdated(nameUpdated, descriptionUpdated, pointOfContactUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyBusinessServiceExists("pagerduty_business_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_business_service.foo", "name", nameUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_business_service.foo", "description", descriptionUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_business_service.foo", "point_of_contact", pointOfContactUpdated),
					resource.TestCheckResourceAttrSet(
						"pagerduty_business_service.foo", "self"),
				),
			},
		},
	})
}

func TestAccPagerDutyBusinessService_WithTeam(t *testing.T) {
	businessService := fmt.Sprintf("tf-%s", acctest.RandString(5))
	teamName := fmt.Sprintf("tf-%s", acctest.RandString(5))
	description := fmt.Sprintf("tf-%s", acctest.RandString(5))
	pointOfContact := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyBusinessServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyBusinessServiceWithTeamConfig(businessService, teamName, description, pointOfContact),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyBusinessServiceExists("pagerduty_business_service.bar"),
					resource.TestCheckResourceAttr(
						"pagerduty_business_service.bar", "name", businessService),
					resource.TestCheckResourceAttr(
						"pagerduty_business_service.bar", "description", description),
					resource.TestCheckResourceAttr(
						"pagerduty_business_service.bar", "point_of_contact", pointOfContact),
					resource.TestCheckResourceAttrSet(
						"pagerduty_business_service.bar", "self"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyBusinessServiceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Business Service ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, _, err := client.BusinessServices.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Business Service not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyBusinessServiceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_business_service" {
			continue
		}

		if _, _, err := client.BusinessServices.Get(r.Primary.ID); err == nil {
			return fmt.Errorf("Business service still exists")
		}

	}
	return nil
}
func testAccCheckPagerDutyBusinessServiceConfig(name, description, poc string) string {
	return fmt.Sprintf(`
resource "pagerduty_business_service" "foo" {
	name = "%s"
	description = "%s"
	point_of_contact = "%s"
}
`, name, description, poc)
}

func testAccCheckPagerDutyBusinessServiceConfigUpdated(name, description, poc string) string {
	return fmt.Sprintf(`
resource "pagerduty_business_service" "foo" {
	name = "%s"
	description = "%s"
	point_of_contact = "%s"
}
`, name, description, poc)
}

func testAccCheckPagerDutyBusinessServiceWithTeamConfig(businessServiceName, teamName, description, poc string) string {
	return fmt.Sprintf(`

resource "pagerduty_team" "bar" {
	name = "%s"
}
	
resource "pagerduty_business_service" "bar" {
	name = "%s"
	description = "%s"
	point_of_contact = "%s"
	team = pagerduty_team.bar.id
}
`, teamName, businessServiceName, description, poc)
}
