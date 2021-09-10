package pagerduty

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func init() {
	resource.AddTestSweepers("pagerduty_extension", &resource.Sweeper{
		Name: "pagerduty_extension",
		F:    testSweepExtension,
	})
}

func testSweepExtension(region string) error {
	config, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client, err := config.Client()
	if err != nil {
		return err
	}

	resp, _, err := client.Extensions.List(&pagerduty.ListExtensionsOptions{})
	if err != nil {
		return err
	}

	for _, extension := range resp.Extensions {
		if strings.HasPrefix(extension.Name, "test") || strings.HasPrefix(extension.Name, "tf-") {
			log.Printf("Destroying extension %s (%s)", extension.Name, extension.ID)
			if _, err := client.Extensions.Delete(extension.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccPagerDutyExtension_Basic(t *testing.T) {
	extension_name := resource.PrefixedUniqueId("tf-")
	extension_name_updated := resource.PrefixedUniqueId("tf-")
	name := resource.PrefixedUniqueId("tf-")
	url := "https://example.com/recieve_a_pagerduty_webhook"
	url_updated := "https://example.com/webhook_foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyExtensionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyExtensionConfig(name, extension_name, url, "false", "any"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyExtensionExists("pagerduty_extension.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_extension.foo", "name", extension_name),
					resource.TestCheckResourceAttr(
						"pagerduty_extension.foo", "extension_schema", "PJFWPEP"),
					resource.TestCheckResourceAttr(
						"pagerduty_extension.foo", "endpoint_url", url),
					resource.TestCheckResourceAttr(
						"pagerduty_extension.foo", "config", "{\"notify_types\":{\"acknowledge\":false,\"assignments\":false,\"resolve\":false},\"restrict\":\"any\"}"),
					resource.TestCheckResourceAttr(
						"pagerduty_extension.foo", "html_url", ""),
				),
			},
			{
				Config: testAccCheckPagerDutyExtensionConfig(name, extension_name_updated, url_updated, "true", "pd-users"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyExtensionExists("pagerduty_extension.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_extension.foo", "name", extension_name_updated),
					resource.TestCheckResourceAttr(
						"pagerduty_extension.foo", "extension_schema", "PJFWPEP"),
					resource.TestCheckResourceAttr(
						"pagerduty_extension.foo", "endpoint_url", url_updated),
					resource.TestCheckResourceAttr(
						"pagerduty_extension.foo", "config", "{\"notify_types\":{\"acknowledge\":true,\"assignments\":true,\"resolve\":true},\"restrict\":\"pd-users\"}"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyExtensionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_extension" {
			continue
		}

		if _, _, err := client.Extensions.Get(r.Primary.ID); err == nil {
			return fmt.Errorf("Extension still exists")
		}

	}
	return nil
}

func testAccCheckPagerDutyExtensionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No extension ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, _, err := client.Extensions.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Extension not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyExtensionConfig(name string, extension_name string, url string, notify_types string, restrict string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%[1]v"
  email       = "%[1]v@foo.com"
  color       = "green"
  role        = "user"
  job_title   = "foo"
  description = "foo"
}

resource "pagerduty_escalation_policy" "foo" {
  name        = "%[1]v"
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
  name                    = "%[1]v"
  description             = "foo"
  auto_resolve_timeout    = 1800
  acknowledgement_timeout = 1800
  escalation_policy       = pagerduty_escalation_policy.foo.id

  incident_urgency_rule {
    type    = "constant"
    urgency = "high"
  }
}

data "pagerduty_extension_schema" "foo" {
	name = "Generic V2 Webhook"
}

resource "pagerduty_extension" "foo"{
  name = "%s"
  endpoint_url = "%s"
  extension_schema = data.pagerduty_extension_schema.foo.id
  extension_objects = [pagerduty_service.foo.id]
  config = <<EOF
{
	"restrict": "%[4]v",
	"notify_types": {
			"resolve": %[5]v,
			"acknowledge": %[5]v,
			"assignments": %[5]v
	}
}
EOF
}

`, name, extension_name, url, restrict, notify_types)
}
