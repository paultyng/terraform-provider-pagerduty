package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

// Testing Business Service Dependencies
func TestAccPagerDutyBusinessServiceDependency_Basic(t *testing.T) {
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))
	businessService := fmt.Sprintf("tf-%s", acctest.RandString(5))
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyBusinessServiceDependencyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyBusinessServiceDependencyConfig(service, businessService, username, email, escalationPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyBusinessServiceDependencyExists("pagerduty_service_dependency.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_dependency.foo", "dependency.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_dependency.foo", "dependency.0.supporting_service.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_dependency.foo", "dependency.0.dependent_service.#", "1"),
				),
			},
		},
	})
}
func testAccCheckPagerDutyBusinessServiceDependencyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service Relationship ID is set")
		}
		businessService, _ := s.RootModule().Resources["pagerduty_business_service.foo"]

		client := testAccProvider.Meta().(*pagerduty.Client)

		depResp, _, err := client.ServiceDependencies.GetServiceDependenciesForType(businessService.Primary.ID, "business_service")
		if err != nil {
			return fmt.Errorf("Business Service not found: %v", err)
		}
		var foundRel *pagerduty.ServiceDependency

		// loop serviceRelationships until relationship.IDs match
		for _, rel := range depResp.Relationships {
			if rel.ID == rs.Primary.ID {
				foundRel = rel
				break
			}
		}
		if foundRel == nil {
			return fmt.Errorf("Service Dependency not found: %v", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPagerDutyBusinessServiceDependencyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_service_dependency" {
			continue
		}
		businessService, _ := s.RootModule().Resources["pagerduty_business_service.foo"]

		// get business service
		dependencies, _, err := client.ServiceDependencies.GetServiceDependenciesForType(businessService.Primary.ID, "business_service")
		if err != nil {
			// if the business service doesn't exist, that's okay
			return nil
		}
		// get business service dependencies
		for _, rel := range dependencies.Relationships {
			if rel.ID == r.Primary.ID {
				return fmt.Errorf("supporting service relationship still exists")
			}
		}

	}
	return nil
}
func testAccCheckPagerDutyBusinessServiceDependencyConfig(service, businessService, username, email, escalationPolicy string) string {
	return fmt.Sprintf(`
resource "pagerduty_business_service" "foo" {
	name = "%s"
}

resource "pagerduty_user" "foo" {
	name        = "%s"
	email       = "%s"
	color       = "green"
	role        = "user"
	job_title   = "foo"
	description = "foo"
}

resource "pagerduty_escalation_policy" "foo" {
	name        = "%s"
	description = "bar"
	num_loops   = 2
	rule {
		escalation_delay_in_minutes = 10
		target {
			type = "user_reference"
			id   = pagerduty_user.foo.id
		}
	}
}
resource "pagerduty_service" "foo" {
	name = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id
	alert_creation          = "create_incidents"
}
resource "pagerduty_service_dependency" "foo" {
	dependency {
		dependent_service {
			id = pagerduty_business_service.foo.id
			type = "business_service"
		}
		supporting_service {
			id = pagerduty_service.foo.id
			type = "service"
		}
	}
}
`, businessService, username, email, escalationPolicy, service)
}

// Testing Technical Service Dependencies
func TestAccPagerDutyTechnicalServiceDependency_Basic(t *testing.T) {
	dependentService := fmt.Sprintf("tf-%s", acctest.RandString(5))
	supportingService := fmt.Sprintf("tf-%s", acctest.RandString(5))
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyTechnicalServiceDependencyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyTechnicalServiceDependencyConfig(dependentService, supportingService, username, email, escalationPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyTechnicalServiceDependencyExists("pagerduty_service_dependency.bar"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_dependency.bar", "dependency.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_dependency.bar", "dependency.0.supporting_service.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service_dependency.bar", "dependency.0.dependent_service.#", "1"),
				),
			},
		},
	})
}
func testAccCheckPagerDutyTechnicalServiceDependencyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service Relationship ID is set")
		}
		supportService, _ := s.RootModule().Resources["pagerduty_service.supportBar"]

		client := testAccProvider.Meta().(*pagerduty.Client)

		depResp, _, err := client.ServiceDependencies.GetServiceDependenciesForType(supportService.Primary.ID, "service")
		if err != nil {
			return fmt.Errorf("Technical Service not found: %v", err)
		}
		var foundRel *pagerduty.ServiceDependency

		// loop serviceRelationships until relationship.IDs match
		for _, rel := range depResp.Relationships {
			if rel.ID == rs.Primary.ID {
				foundRel = rel
				break
			}
		}
		if foundRel == nil {
			return fmt.Errorf("Service Dependency not found: %v", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPagerDutyTechnicalServiceDependencyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_service_dependency" {
			continue
		}
		supportService, _ := s.RootModule().Resources["pagerduty_service.supportBar"]

		// get service dependencies
		dependencies, _, err := client.ServiceDependencies.GetServiceDependenciesForType(supportService.Primary.ID, "service")
		if err != nil {
			// if the dependency doesn't exist, that's okay
			return nil
		}
		// find desired dependency
		for _, rel := range dependencies.Relationships {
			if rel.ID == r.Primary.ID {
				return fmt.Errorf("supporting service relationship still exists")
			}
		}

	}
	return nil
}
func testAccCheckPagerDutyTechnicalServiceDependencyConfig(dependentService, supportingService, username, email, escalationPolicy string) string {
	return fmt.Sprintf(`


resource "pagerduty_user" "bar" {
	name        = "%s"
	email       = "%s"
	color       = "green"
	role        = "user"
	job_title   = "foo"
	description = "foo"
}

resource "pagerduty_escalation_policy" "bar" {
	name        = "%s"
	description = "bar-desc"
	num_loops   = 2
	rule {
		escalation_delay_in_minutes = 10
		target {
			type = "user_reference"
			id   = pagerduty_user.bar.id
		}
	}
}
resource "pagerduty_service" "supportBar" {
	name = "%s"
	description             = "supportBarDesc"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.bar.id
	alert_creation          = "create_incidents"
}
resource "pagerduty_service" "dependBar" {
	name = "%s"
	description             = "dependBarDesc"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.bar.id
	alert_creation          = "create_incidents"
}
resource "pagerduty_service_dependency" "bar" {
	dependency {
		dependent_service {
			id = pagerduty_service.dependBar.id
			type = "service"
		}
		supporting_service {
			id = pagerduty_service.supportBar.id
			type = "service"
		}
	}
}
`, username, email, escalationPolicy, supportingService, dependentService)
}
