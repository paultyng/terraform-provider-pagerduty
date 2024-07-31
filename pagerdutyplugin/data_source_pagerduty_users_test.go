package pagerduty

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourcePagerDutyUsers_Basic(t *testing.T) {
	timeZone := "America/New_York"
	teamname1 := fmt.Sprintf("tf-team-%s", acctest.RandString(5))
	teamname2 := fmt.Sprintf("tf-team-%s", acctest.RandString(5))

	username1 := fmt.Sprintf("tf-user1-%s", acctest.RandString(5))
	email1 := fmt.Sprintf("%s@foo.test", username1)
	title1 := fmt.Sprintf("%s-title", username1)
	timeZone1 := timeZone
	description1 := fmt.Sprintf("%s-description", username1)

	username2 := fmt.Sprintf("tf-user2-%s", acctest.RandString(5))
	email2 := fmt.Sprintf("%s@foo.test", username2)
	title2 := fmt.Sprintf("%s-title", username2)
	timeZone2 := timeZone
	description2 := fmt.Sprintf("%s-description", username2)

	username3 := fmt.Sprintf("tf-user3-%s", acctest.RandString(5))
	email3 := fmt.Sprintf("%s@foo.test", username3)
	title3 := fmt.Sprintf("%s-title", username3)
	description3 := fmt.Sprintf("%s-description", username3)
	timeZone3 := timeZone

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePagerDutyUsersConfig(teamname1, teamname2, username1, email1, title1, timeZone1, description1, username2, email2, title2, timeZone2, description2, username3, email3, title3, timeZone3, description3),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourcePagerDutyUsersExists("data.pagerduty_users.test_all_users"),
					testAccDataSourcePagerDutyUsersExists("data.pagerduty_users.test_by_1_team"),
					testAccDataSourcePagerDutyUsersExists("data.pagerduty_users.test_by_2_team"),
					resource.TestCheckResourceAttrSet(
						"data.pagerduty_users.test_all_users", "users.#"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.pagerduty_users.test_all_users",
						"users.*",
						map[string]string{
							"name": username1,
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.pagerduty_users.test_all_users",
						"users.*",
						map[string]string{
							"name": username2,
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.pagerduty_users.test_all_users",
						"users.*",
						map[string]string{
							"name": username3,
						}),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_1_team", "users.#", "1"),
					resource.TestCheckResourceAttrSet(
						"data.pagerduty_users.test_by_1_team", "users.0.id"),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_1_team", "users.0.name", username2),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_1_team", "users.0.email", email2),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_1_team", "users.0.role", "user"),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_1_team", "users.0.job_title", title2),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_1_team", "users.0.time_zone", timeZone2),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_1_team", "users.0.description", description2),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.#", "2"),
					resource.TestCheckResourceAttrSet(
						"data.pagerduty_users.test_by_2_team", "users.0.id"),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.0.name", username2),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.0.email", email2),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.0.role", "user"),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.0.job_title", title2),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.0.time_zone", timeZone2),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.0.description", description2),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.1.name", username3),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.1.email", email3),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.1.role", "user"),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.1.job_title", title3),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.1.time_zone", timeZone3),
					resource.TestCheckResourceAttr(
						"data.pagerduty_users.test_by_2_team", "users.1.description", description3),
				),
			},
		},
	})
}

func testAccDataSourcePagerDutyUsersExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		if a["id"] == "" {
			return fmt.Errorf("Expected to get a user ID from PagerDuty")
		}

		return nil
	}
}

func testAccDataSourcePagerDutyUsersConfig(teamname1, teamname2, username1, email1, title1, timeZone1, description1, username2, email2, title2, timeZone2, description2, username3, email3, title3, timeZone3, description3 string) string {
	return fmt.Sprintf(`
    resource "pagerduty_team" "test1" {
        name = "%s"
    }
    resource "pagerduty_team" "test2" {
        name = "%s"
    }

    data "pagerduty_license" "stakeholder" {
        name = "Digital Operations (Stakeholder)"
    }

    resource "pagerduty_user" "test_wo_team" {
      name = "%s"
      email = "%s"
      job_title = "%s"
      time_zone = "%s"
      description = "%s"
      license = data.pagerduty_license.stakeholder.id
    }
    resource "pagerduty_user" "test_w_team1" {
      name = "%s"
      email = "%s"
      job_title = "%s"
      time_zone = "%s"
      description = "%s"
      license = data.pagerduty_license.stakeholder.id
    }
    resource "pagerduty_user" "test_w_team2" {
      name = "%s"
      email = "%s"
      job_title = "%s"
      time_zone = "%s"
      description = "%s"
      license = data.pagerduty_license.stakeholder.id
    }

    resource "pagerduty_team_membership" "test1" {
      team_id = pagerduty_team.test1.id
      user_id = pagerduty_user.test_w_team1.id
    }
    resource "pagerduty_team_membership" "test2" {
      depends_on = [pagerduty_team_membership.test1]
      team_id = pagerduty_team.test2.id
      user_id = pagerduty_user.test_w_team2.id
    }

    data "pagerduty_users" "test_all_users" {
      depends_on = [pagerduty_user.test_w_team1, pagerduty_user.test_wo_team]
    }

    data "pagerduty_users" "test_by_1_team" {
      depends_on = [pagerduty_team_membership.test1]
      team_ids = [pagerduty_team.test1.id]
    }
    data "pagerduty_users" "test_by_2_team" {
      depends_on = [pagerduty_team_membership.test2]
      team_ids = [pagerduty_team.test1.id, pagerduty_team.test2.id]
    }
`,
		teamname1, teamname2,
		username1, email1, title1, timeZone1, description1,
		username2, email2, title2, timeZone2, description2,
		username3, email3, title3, timeZone3, description3,
	)
}
