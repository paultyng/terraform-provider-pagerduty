package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPagerDutyExtension_import(t *testing.T) {
	extension_name := fmt.Sprintf("tf-%s", acctest.RandString(5))
	name := fmt.Sprintf("tf-%s", acctest.RandString(5))
	url := "https://example.com/receive_a_pagerduty_webhook"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyExtensionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyExtensionConfig(name, extension_name, url, "false", "any"),
			},

			{
				ResourceName:      "pagerduty_extension.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
