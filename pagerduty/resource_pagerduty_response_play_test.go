package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func TestAccPagerDutyResponsePlay_Basic(t *testing.T) {
	name := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyResponsePlayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyResponsePlayConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyResponsePlayExists("pagerduty_response_play.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_response_play.foo", "name", name),
					resource.TestCheckResourceAttr(
						"pagerduty_response_play.foo", "from", name+"@foo.com"),
					resource.TestCheckResourceAttr(
						"pagerduty_response_play.foo", "responder.#", "2"),
				),
			},
			{
				Config: testAccCheckPagerDutyResponsePlayConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyResponsePlayExists("pagerduty_response_play.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_response_play.foo", "name", name),
					resource.TestCheckResourceAttr(
						"pagerduty_response_play.foo", "from", name+"@foo.com"),
					resource.TestCheckResourceAttr(
						"pagerduty_response_play.foo", "responder.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_response_play.foo", "subscriber.#", "1"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyResponsePlayDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_response_play" {
			continue
		}
		u, _ := s.RootModule().Resources["pagerduty_user.foo"]
		ua := u.Primary.Attributes

		if _, _, err := client.ResponsePlays.Get(r.Primary.ID, ua["email"]); err == nil {
			return fmt.Errorf("response play still exists")
		}

	}
	return nil
}

func testAccCheckPagerDutyResponsePlayExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No response play ID is set")
		}
		u, _ := s.RootModule().Resources["pagerduty_user.foo"]
		ua := u.Primary.Attributes

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, _, err := client.ResponsePlays.Get(rs.Primary.ID, ua["email"])
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("response play not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyResponsePlayConfig(name string) string {
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

resource "pagerduty_response_play" "foo" {
  name = "%[1]v"
  from = pagerduty_user.foo.email
  responder {
	  type = "user_reference"
	  id = pagerduty_user.foo.id
  }
  responder {
	  type = "escalation_policy_reference"
	  id = pagerduty_escalation_policy.foo.id
  }
  subscriber {
	type = "user_reference"
	id = pagerduty_user.foo.id
  }
runnability = "services"	
}
`, name)
}

func testAccCheckPagerDutyResponsePlayConfigUpdated(name string) string {
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
	
resource "pagerduty_response_play" "foo" {
	name = "%[1]v"
	from = pagerduty_user.foo.email
	responder {
		type = "escalation_policy_reference"
		id = pagerduty_escalation_policy.foo.id
	}
	subscriber {
		type = "user_reference"
		id = pagerduty_user.foo.id
	}
	runnability = "services"
}
`, name)
}
