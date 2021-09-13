package pagerduty

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func init() {
	resource.AddTestSweepers("pagerduty_service", &resource.Sweeper{
		Name: "pagerduty_service",
		F:    testSweepService,
	})
}

func testSweepService(region string) error {
	config, err := sharedConfigForRegion(region)
	if err != nil {
		return err
	}

	client, err := config.Client()
	if err != nil {
		return err
	}

	resp, _, err := client.Services.List(&pagerduty.ListServicesOptions{})
	if err != nil {
		return err
	}

	for _, service := range resp.Services {
		if strings.HasPrefix(service.Name, "test") || strings.HasPrefix(service.Name, "tf-") {
			log.Printf("Destroying service %s (%s)", service.Name, service.ID)
			if _, err := client.Services.Delete(service.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccPagerDutyService_Basic(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))
	serviceUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyServiceConfig(username, email, escalationPolicy, service),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", service),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "auto_resolve_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "acknowledgement_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_creation", "create_incidents"),
					resource.TestCheckNoResourceAttr(
						"pagerduty_service.foo", "alert_grouping"),
					resource.TestCheckNoResourceAttr(
						"pagerduty_service.foo", "alert_grouping_timeout"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.type", "constant"),
					resource.TestCheckResourceAttrSet(
						"pagerduty_service.foo", "html_url"),
				),
			},
			{
				Config: testAccCheckPagerDutyServiceConfigUpdated(username, email, escalationPolicy, serviceUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", serviceUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "description", "bar"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "auto_resolve_timeout", "3600"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "acknowledgement_timeout", "3600"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_creation", "create_incidents"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.type", "constant"),
				),
			},
			{
				Config: testAccCheckPagerDutyServiceConfigUpdatedWithDisabledTimeouts(username, email, escalationPolicy, serviceUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", serviceUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "description", "bar"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "auto_resolve_timeout", "null"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "acknowledgement_timeout", "null"),
				),
			},
		},
	})
}

func TestAccPagerDutyService_AlertGrouping(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckPagerDutyAbility(t, "preview_intelligent_alert_grouping") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyServiceConfigWithAlertGrouping(username, email, escalationPolicy, service),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", service),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "auto_resolve_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "acknowledgement_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_creation", "create_alerts_and_incidents"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_grouping", "time"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_grouping_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.type", "constant"),
				),
			},
			{
				Config: testAccCheckPagerDutyServiceConfigWithAlertGroupingUpdated(username, email, escalationPolicy, service),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", service),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "auto_resolve_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "acknowledgement_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_creation", "create_alerts_and_incidents"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_grouping", "intelligent"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_grouping_timeout", "1900"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.type", "constant"),
				),
			},
		},
	})
}

func TestAccPagerDutyService_AlertContentGrouping(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyServiceConfigWithAlertContentGrouping(username, email, escalationPolicy, service),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", service),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "auto_resolve_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "acknowledgement_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_creation", "create_alerts_and_incidents"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_grouping", "rules"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_grouping_parameters.0.config.0.aggregate", "all"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_grouping_parameters.0.type", "content_based"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_grouping_parameters.0.config.0.fields.0", "custom_details.field1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.type", "constant"),
				),
			},
		},
	})
}

func TestAccPagerDutyService_BasicWithIncidentUrgencyRules(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))
	serviceUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyServiceWithIncidentUrgencyRulesConfig(username, email, escalationPolicy, service),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", service),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "auto_resolve_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "acknowledgement_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_creation", "create_incidents"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.0.type", "constant"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.0.urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.0.type", "constant"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.0.urgency", "low"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.type", "use_support_hours"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.at.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.at.0.name", "support_hours_start"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.to_urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.type", "urgency_change"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.#", "5"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.0", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.1", "2"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.2", "3"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.3", "4"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.4", "5"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.end_time", "17:00:00"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.start_time", "09:00:00"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.time_zone", "America/Lima"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.type", "fixed_time_per_day"),
				),
			},
			{
				Config:      testAccCheckPagerDutyServiceWithIncidentUrgencyRulesConfigError(username, email, escalationPolicy, serviceUpdated),
				ExpectError: regexp.MustCompile("general urgency cannot be set for a use_support_hours incident urgency rule type"),
			},
			{
				Config: testAccCheckPagerDutyServiceWithIncidentUrgencyRulesWithoutScheduledActionsConfig(username, email, escalationPolicy, service),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", service),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "auto_resolve_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "acknowledgement_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_creation", "create_incidents"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.0.type", "constant"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.0.urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.0.type", "constant"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.0.urgency", "severity_based"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.type", "use_support_hours"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.#", "5"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.0", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.1", "2"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.2", "3"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.3", "4"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.4", "5"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.end_time", "17:00:00"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.start_time", "09:00:00"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.time_zone", "America/Lima"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.type", "fixed_time_per_day"),
				),
			},
			{
				Config: testAccCheckPagerDutyServiceWithIncidentUrgencyRulesConfigUpdated(username, email, escalationPolicy, serviceUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", serviceUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "description", "bar bar bar"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "auto_resolve_timeout", "3600"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "acknowledgement_timeout", "3600"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_creation", "create_incidents"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.0.type", "constant"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.0.urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.0.type", "constant"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.0.urgency", "low"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.type", "use_support_hours"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.at.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.at.0.name", "support_hours_start"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.to_urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.type", "urgency_change"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.#", "5"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.0", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.1", "2"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.2", "3"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.3", "4"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.4", "5"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.end_time", "17:00:00"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.start_time", "09:00:00"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.time_zone", "America/Lima"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.type", "fixed_time_per_day"),
				),
			},
		},
	})
}

func TestAccPagerDutyService_FromBasicToCustomIncidentUrgencyRules(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))
	serviceUpdated := fmt.Sprintf("tf-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyServiceConfig(username, email, escalationPolicy, service),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", service),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "description", "foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "auto_resolve_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "acknowledgement_timeout", "1800"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_creation", "create_incidents"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.type", "constant"),
				),
			},
			{
				Config: testAccCheckPagerDutyServiceWithIncidentUrgencyRulesConfigUpdated(username, email, escalationPolicy, serviceUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", serviceUpdated),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "description", "bar bar bar"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "auto_resolve_timeout", "3600"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "acknowledgement_timeout", "3600"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "alert_creation", "create_incidents"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.0.type", "constant"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.during_support_hours.0.urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.0.type", "constant"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.outside_support_hours.0.urgency", "low"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "incident_urgency_rule.0.type", "use_support_hours"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.at.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.at.0.name", "support_hours_start"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.to_urgency", "high"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "scheduled_actions.0.type", "urgency_change"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.#", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.#", "5"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.0", "1"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.1", "2"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.2", "3"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.3", "4"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.days_of_week.4", "5"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.end_time", "17:00:00"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.start_time", "09:00:00"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.time_zone", "America/Lima"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "support_hours.0.type", "fixed_time_per_day"),
				),
			},
		},
	})
}

func TestAccPagerDutyService_SupportHoursChange(t *testing.T) {
	username := fmt.Sprintf("tf-%s", acctest.RandString(5))
	email := fmt.Sprintf("%s@foo.com", username)
	escalationPolicy := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service := fmt.Sprintf("tf-%s", acctest.RandString(5))
	service_id := ""
	p_service_id := &service_id
	updated_service_id := ""
	p_updated_service_id := &updated_service_id

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPagerDutyServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckPagerDutyServiceWithIncidentUrgencyRulesConfig(username, email, escalationPolicy, service),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", service),
					testAccCheckPagerDutyServiceSaveServiceId(p_service_id, "pagerduty_service.foo"),
				),
			},
			{
				Config: testAccCheckPagerDutyServiceWithSupportHoursConfigUpdated(username, email, escalationPolicy, service),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPagerDutyServiceExists("pagerduty_service.foo"),
					resource.TestCheckResourceAttr(
						"pagerduty_service.foo", "name", service),
					testAccCheckPagerDutyServiceSaveServiceId(p_updated_service_id, "pagerduty_service.foo"),
				),
			},
		},
	})

	if service_id != updated_service_id {
		t.Error(fmt.Errorf("Expected service id to be %s, but found %s", service_id, updated_service_id))
	}
}

func testAccCheckPagerDutyServiceSaveServiceId(p *string, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, _, err := client.Services.Get(rs.Primary.ID, &pagerduty.GetServiceOptions{})
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Service not found: %v - %v", rs.Primary.ID, found)
		}

		*p = found.ID

		return nil
	}
}

func testAccCheckPagerDutyServiceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*pagerduty.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "pagerduty_service" {
			continue
		}

		if _, _, err := client.Services.Get(r.Primary.ID, &pagerduty.GetServiceOptions{}); err == nil {
			return fmt.Errorf("Service still exists")
		}

	}
	return nil
}

func testAccCheckPagerDutyServiceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service ID is set")
		}

		client := testAccProvider.Meta().(*pagerduty.Client)

		found, _, err := client.Services.Get(rs.Primary.ID, &pagerduty.GetServiceOptions{})
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Service not found: %v - %v", rs.Primary.ID, found)
		}

		return nil
	}
}

func testAccCheckPagerDutyServiceConfig(username, email, escalationPolicy, service string) string {
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
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id
	alert_creation          = "create_incidents"
}
`, username, email, escalationPolicy, service)
}

func testAccCheckPagerDutyServiceConfigWithAlertGrouping(username, email, escalationPolicy, service string) string {
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
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id
	alert_creation          = "create_alerts_and_incidents"
	alert_grouping          = "time"
	alert_grouping_timeout  = 1800
}
`, username, email, escalationPolicy, service)
}

func testAccCheckPagerDutyServiceConfigWithAlertContentGrouping(username, email, escalationPolicy, service string) string {
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
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id
	alert_creation          = "create_alerts_and_incidents"
	alert_grouping_parameters { 
        type = "content_based"
        config { 
            aggregate = "all"
            fields = ["custom_details.field1"]
        }
    }
}
`, username, email, escalationPolicy, service)
}
func testAccCheckPagerDutyServiceConfigWithAlertGroupingUpdated(username, email, escalationPolicy, service string) string {
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
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id
	alert_creation          = "create_alerts_and_incidents"
	alert_grouping          = "intelligent"
	alert_grouping_timeout  = 1900
}
`, username, email, escalationPolicy, service)
}

func testAccCheckPagerDutyServiceConfigUpdated(username, email, escalationPolicy, service string) string {
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
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "bar"
	auto_resolve_timeout    = 3600
	acknowledgement_timeout = 3600

	escalation_policy       = pagerduty_escalation_policy.foo.id
	incident_urgency_rule {
		type    = "constant"
		urgency = "high"
	}
}
`, username, email, escalationPolicy, service)
}

func testAccCheckPagerDutyServiceConfigUpdatedWithDisabledTimeouts(username, email, escalationPolicy, service string) string {
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
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "bar"
	auto_resolve_timeout    = "null"
	acknowledgement_timeout = "null"

	escalation_policy       = pagerduty_escalation_policy.foo.id
	incident_urgency_rule {
		type    = "constant"
		urgency = "high"
	}
}
`, username, email, escalationPolicy, service)
}

func testAccCheckPagerDutyServiceWithIncidentUrgencyRulesConfig(username, email, escalationPolicy, service string) string {
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
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id

	incident_urgency_rule {
		type = "use_support_hours"

		during_support_hours {
			type    = "constant"
			urgency = "high"
		}
		outside_support_hours {
			type    = "constant"
			urgency = "low"
		}
	}

	support_hours {
		type         = "fixed_time_per_day"
		time_zone    = "America/Lima"
		start_time   = "09:00:00"
		end_time     = "17:00:00"
		days_of_week = [ 1, 2, 3, 4, 5 ]
	}

	scheduled_actions {
		type = "urgency_change"
		to_urgency = "high"
		at {
			type = "named_time"
			name = "support_hours_start"
		}
	}
}
`, username, email, escalationPolicy, service)
}

func testAccCheckPagerDutyServiceWithIncidentUrgencyRulesConfigError(username, email, escalationPolicy, service string) string {
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
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id

	incident_urgency_rule {
		type    = "use_support_hours"
		urgency = "high"
		during_support_hours {
			type    = "constant"
			urgency = "high"
		}
		outside_support_hours {
			type    = "constant"
			urgency = "low"
		}
	}

	support_hours {
		type         = "fixed_time_per_day"
		time_zone    = "America/Lima"
		start_time   = "09:00:00"
		end_time     = "17:00:00"
		days_of_week = [ 1, 2, 3, 4, 5 ]
	}

	scheduled_actions {
		type = "urgency_change"
		to_urgency = "high"
		at {
			type = "named_time"
			name = "support_hours_start"
		}
	}
}
`, username, email, escalationPolicy, service)
}

func testAccCheckPagerDutyServiceWithIncidentUrgencyRulesWithoutScheduledActionsConfig(username, email, escalationPolicy, service string) string {
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
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id

	incident_urgency_rule {
		type = "use_support_hours"

		during_support_hours {
			type    = "constant"
			urgency = "high"
		}
		outside_support_hours {
			type    = "constant"
			urgency = "severity_based"
		}
	}

	support_hours {
		type         = "fixed_time_per_day"
		time_zone    = "America/Lima"
		start_time   = "09:00:00"
		end_time     = "17:00:00"
		days_of_week = [ 1, 2, 3, 4, 5 ]
	}
}
`, username, email, escalationPolicy, service)
}

func testAccCheckPagerDutyServiceWithIncidentUrgencyRulesConfigUpdated(username, email, escalationPolicy, service string) string {
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
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "bar bar bar"
	auto_resolve_timeout    = 3600
	acknowledgement_timeout = 3600
	escalation_policy       = pagerduty_escalation_policy.foo.id

	incident_urgency_rule {
		type = "use_support_hours"
		during_support_hours {
			type    = "constant"
			urgency = "high"
		}
		outside_support_hours {
			type    = "constant"
			urgency = "low"
		}
	}

	support_hours {
		type         = "fixed_time_per_day"
		time_zone    = "America/Lima"
		start_time   = "09:00:00"
		end_time     = "17:00:00"
		days_of_week = [ 1, 2, 3, 4, 5 ]
	}

	scheduled_actions {
		type = "urgency_change"
		to_urgency = "high"
		at {
			type = "named_time"
			name = "support_hours_start"
		}
	}
}
`, username, email, escalationPolicy, service)
}

func testAccCheckPagerDutyServiceWithSupportHoursConfigUpdated(username, email, escalationPolicy, service string) string {
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
}

resource "pagerduty_service" "foo" {
	name                    = "%s"
	description             = "foo"
	auto_resolve_timeout    = 1800
	acknowledgement_timeout = 1800
	escalation_policy       = pagerduty_escalation_policy.foo.id

	incident_urgency_rule {
		type = "constant"
		urgency = "high"
	}

}
`, username, email, escalationPolicy, service)
}
