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
	resource.AddTestSweepers("pagerduty_user", &resource.Sweeper{
		Name: "pagerduty_user",
		F:    testSweepUser,
		Dependencies: []string{
			"pagerduty_team",
			"pagerduty_schedule",
			"pagerduty_escalation_policy",
		},
	})
}

func testSweepUser(region string) error {
	config, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client, err := config.Client()
	if err != nil {
		return err
	}

	resp, _, err := client.Users.List(&pagerduty.ListUsersOptions{})
	if err != nil {
		return err
	}

	for _, user := range resp.Users {
		if strings.HasPrefix(user.Name, "test") || strings.HasPrefix(user.Name, "tf") {
			log.Printf("Destroying user %s (%s)", user.Name, user.ID)
			if _, err := client.Users.Delete(user.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccPagerDutyUser_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	usernameSpaces := " " + username + " "
	usernameUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@Foo.com", username)
	emailUpdated := fmt.Sprintf("%s@foo.com", usernameUpdated)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserConfig(usernameSpaces, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserExists("pagerduty_user.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "name", username),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "email", strings.ToLower(email)),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "color", "green"),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "role", "user"),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "job_title", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "description", "foo"),
					resource.TestCheckResourceAttrSet(
						"pagerduty_user.foo", "html_url"),
				),
			},
			{
				Config: testAccCheckPagerDutyUserConfigUpdated(usernameUpdated, emailUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserExists("pagerduty_user.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "name", usernameUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "email", emailUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "color", "red"),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "role", "observer"),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "job_title", "bar"),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "description", "bar"),
				),
			},
		},
	})
}

func TestAccPagerDutyUserWithTeams_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	team1 := fmt.Sprintf("tf-%s", acctest.RandString(5))
	team2 := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyUserWithTeamsConfig(team1, username, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserExists("pagerduty_user.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "name", username),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "email", email),
				),
			},
			{
				Config: testAccCheckPagerDutyUserWithTeamsConfigUpdated(team1, team2, username, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserExists("pagerduty_user.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "name", username),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "email", email),
				),
			},
			{
				Config: testAccCheckPagerDutyUserWithNoTeamsConfig(team1, team2, username, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyUserExists("pagerduty_user.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "name", username),
					resource.TestCheckResourceAttr(
						"pagerduty_user.foo", "email", email),
				),
			},
		},
	})
}

func testAccCheckPagerDutyUserDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_user" {
			continue
		}

		if _, _, err := client.Users.Get(r.Primary.ID, &pagerduty.GetUserOptions{}); err == nil {
			return fmt.Errorf("User still exists")
		}

	}
	return nil
}

func testAccCheckPagerDutyUserExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No user ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, _, err := client.Users.Get(rs.Primary.ID, &pagerduty.GetUserOptions{})
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("User not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyUserConfig(username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
  color       = "green"
  role        = "user"
  job_title   = "foo"
  description = "foo"
  time_zone   = "Europe/Berlin"
}`, username, email)
}

func testAccCheckPagerDutyUserConfigUpdated(username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_user" "foo" {
  name        = "%s"
  email       = "%s"
  color       = "red"
  role        = "observer"
  job_title   = "bar"
  description = "bar"
  time_zone   = "Europe/Dublin"
}`, username, email)
}

func testAccCheckPagerDutyUserWithTeamsConfig(team, username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_team" "foo" {
  name = "%s"
}

resource "pagerduty_user" "foo" {
  name  = "%s"
  email = "%s"
}

resource "pagerduty_team_membership" "foo" {
  user_id = pagerduty_user.foo.id
  team_id = pagerduty_team.foo.id
}
`, team, username, email)
}

func testAccCheckPagerDutyUserWithTeamsConfigUpdated(team1, team2, username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_team" "foo" {
  name = "%s"
}

resource "pagerduty_team" "bar" {
  name = "%s"
}

resource "pagerduty_user" "foo" {
  name  = "%s"
  email = "%s"
}

resource "pagerduty_team_membership" "foo" {
  user_id = pagerduty_user.foo.id
  team_id = pagerduty_team.foo.id
}

resource "pagerduty_team_membership" "bar" {
  user_id = pagerduty_user.foo.id
  team_id = pagerduty_team.bar.id
}
`, team1, team2, username, email)
}

func testAccCheckPagerDutyUserWithNoTeamsConfig(team1, team2, username, email string) string {
	return fmt.Sprintf(`
resource "pagerduty_team" "foo" {
  name = "%s"
}

resource "pagerduty_team" "bar" {
  name = "%s"
}

resource "pagerduty_user" "foo" {
  name  = "%s"
  email = "%s"
}
`, team1, team2, username, email)
}
