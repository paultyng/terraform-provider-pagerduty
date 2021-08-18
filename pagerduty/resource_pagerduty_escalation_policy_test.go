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
	resource.AddTestSweepers("pagerduty_escalation_policy", &resource.Sweeper{
		Name: "pagerduty_escalation_policy",
		F:    testSweepEscalationPolicy,
		Dependencies: []string{
			"pagerduty_service",
		},
	})
}

func testSweepEscalationPolicy(region string) error {
	config, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client, err := config.Client()
	if err != nil {
		return err
	}

	resp, _, err := client.EscalationPolicies.List(&pagerduty.ListEscalationPoliciesOptions{})
	if err != nil {
		return err
	}

	for _, escalation := range resp.EscalationPolicies {
		if strings.HasPrefix(escalation.Name, "test") || strings.HasPrefix(escalation.Name, "tf-") {
			log.Printf("Destroying escalation policy %s (%s)", escalation.Name, escalation.ID)
			if _, err := client.EscalationPolicies.Delete(escalation.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccPagerDutyEscalationPolicy_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	escalationPolicyUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyEscalationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyEscalationPolicyConfig(username, email, escalationPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyEscalationPolicyExists("pagerduty_escalation_policy.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "name", escalationPolicy),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "num_loops", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "rule.0.escalation_delay_in_minutes", "10"),
				),
			},

			{
				Config: testAccCheckPagerDutyEscalationPolicyConfigUpdated(username, email, escalationPolicyUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyEscalationPolicyExists("pagerduty_escalation_policy.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "name", escalationPolicyUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "description", "bar"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "num_loops", "2"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "rule.#", "2"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "rule.0.escalation_delay_in_minutes", "10"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "rule.1.escalation_delay_in_minutes", "20"),
				),
			},
		},
	})
}

func TestAccPagerDutyEscalationPolicyWithTeams_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	team := fmt.Sprintf("tf-%s", acctest.RandString(5))
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	escalationPolicyUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyEscalationPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyEscalationPolicyWithTeamsConfig(username, email, team, escalationPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyEscalationPolicyExists("pagerduty_escalation_policy.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "name", escalationPolicy),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "num_loops", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "rule.0.escalation_delay_in_minutes", "10"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "teams.#", "1"),
				),
			},
			{
				Config: testAccCheckPagerDutyEscalationPolicyWithTeamsConfigUpdated(username, email, team, escalationPolicyUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyEscalationPolicyExists("pagerduty_escalation_policy.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "name", escalationPolicyUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "description", "bar"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "num_loops", "2"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "rule.#", "2"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "rule.0.escalation_delay_in_minutes", "10"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "rule.1.escalation_delay_in_minutes", "20"),
					resource.TestCheckResourceAttr(
						"pagerduty_escalation_policy.foo", "teams.#", "0"),
				),
			},
		},
	})
}

func testAccCheckPagerDutyEscalationPolicyDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_escalation_policy" {
			continue
		}

		if _, _, err := client.EscalationPolicies.Get(r.Primary.ID, &pagerduty.GetEscalationPolicyOptions{}); err == nil {
			return fmt.Errorf("Escalation Policy still exists")
		}

	}
	return nil
}

func testAccCheckPagerDutyEscalationPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Escalation Policy ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, _, err := client.EscalationPolicies.Get(rs.Primary.ID, &pagerduty.GetEscalationPolicyOptions{})
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Escalation policy not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyEscalationPolicyConfig(name, email, escalationPolicy string) string {
	return fmt.Sprintf(`
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
  description = "foo"
  num_loops   = 1

  rule {
    escalation_delay_in_minutes = 10

    target {
      type = "user_reference"
      id   = pagerduty_user.foo.id
    }
  }
}
`, name, email, escalationPolicy)
}

func testAccCheckPagerDutyEscalationPolicyConfigUpdated(name, email, escalationPolicy string) string {
	return fmt.Sprintf(`
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

  rule {
    escalation_delay_in_minutes = 20

    target {
      type = "user_reference"
      id   = pagerduty_user.foo.id
    }
  }
}
`, name, email, escalationPolicy)
}

func testAccCheckPagerDutyEscalationPolicyWithTeamsConfig(name, email, team, escalationPolicy string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
  color       = "green"
  role        = "user"
  job_title   = "foo"
  description = "foo"
}

resource "pagerduty_team" "foo" {
  name        = "%s"
  description = "foo"
}

resource "pagerduty_escalation_policy" "foo" {
  name        = "%s"
  description = "foo"
  num_loops   = 1
	teams       = [pagerduty_team.foo.id]

  rule {
    escalation_delay_in_minutes = 10

    target {
      type = "user_reference"
      id   = pagerduty_user.foo.id
    }
  }
}
`, name, email, team, escalationPolicy)
}

func testAccCheckPagerDutyEscalationPolicyWithTeamsConfigUpdated(name, email, team, escalationPolicy string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
  color       = "green"
  role        = "user"
  job_title   = "foo"
  description = "foo"
}

resource "pagerduty_team" "foo" {
  name        = "%s"
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

  rule {
    escalation_delay_in_minutes = 20

    target {
      type = "user_reference"
      id   = pagerduty_user.foo.id
    }
  }
}
`, name, email, team, escalationPolicy)
}
