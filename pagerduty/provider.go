package pagerduty

import (
	"fmt"
	"log"
	"runtime"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

// Provider represents a resource provider in Terraform
func Provider() terraform.ResourceProvider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"skip_credentials_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PAGERDUTY_TOKEN", nil),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"pagerduty_escalation_policy":   dataSourcePagerDutyEscalationPolicy(),
			"pagerduty_schedule":            dataSourcePagerDutySchedule(),
			"pagerduty_user":                dataSourcePagerDutyUser(),
			"pagerduty_user_contact_method": dataSourcePagerDutyUserContactMethod(),
			"pagerduty_team":                dataSourcePagerDutyTeam(),
			"pagerduty_vendor":              dataSourcePagerDutyVendor(),
			"pagerduty_extension_schema":    dataSourcePagerDutyExtensionSchema(),
			"pagerduty_service":             dataSourcePagerDutyService(),
			"pagerduty_service_integration": dataSourcePagerDutyServiceIntegration(),
			"pagerduty_business_service":    dataSourcePagerDutyBusinessService(),
			"pagerduty_priority":            dataSourcePagerDutyPriority(),
			"pagerduty_ruleset":             dataSourcePagerDutyRuleset(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"pagerduty_addon":                  resourcePagerDutyAddon(),
			"pagerduty_escalation_policy":      resourcePagerDutyEscalationPolicy(),
			"pagerduty_maintenance_window":     resourcePagerDutyMaintenanceWindow(),
			"pagerduty_schedule":               resourcePagerDutySchedule(),
			"pagerduty_service":                resourcePagerDutyService(),
			"pagerduty_service_integration":    resourcePagerDutyServiceIntegration(),
			"pagerduty_team":                   resourcePagerDutyTeam(),
			"pagerduty_team_membership":        resourcePagerDutyTeamMembership(),
			"pagerduty_user":                   resourcePagerDutyUser(),
			"pagerduty_user_contact_method":    resourcePagerDutyUserContactMethod(),
			"pagerduty_user_notification_rule": resourcePagerDutyUserNotificationRule(),
			"pagerduty_extension":              resourcePagerDutyExtension(),
			"pagerduty_extension_servicenow":   resourcePagerDutyExtensionServiceNow(),
			"pagerduty_event_rule":             resourcePagerDutyEventRule(),
			"pagerduty_ruleset":                resourcePagerDutyRuleset(),
			"pagerduty_ruleset_rule":           resourcePagerDutyRulesetRule(),
			"pagerduty_business_service":       resourcePagerDutyBusinessService(),
			"pagerduty_service_dependency":     resourcePagerDutyServiceDependency(),
			"pagerduty_response_play":          resourcePagerDutyResponsePlay(),
			"pagerduty_service_event_rule":     resourcePagerDutyServiceEventRule(),
			"pagerduty_slack_connection":       resourcePagerDutySlackConnection(),
		},
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}

	return p
}

func isErrCode(err error, code int) bool {
	if e, ok := err.(*pagerduty.Error); ok && e.ErrorResponse.Response.StatusCode == code {
		return true
	}

	return false
}

func handleNotFoundError(err error, d *schema.ResourceData) error {
	if isErrCode(err, 404) {
		log.Printf("[WARN] Removing %s because it's gone", d.Id())
		d.SetId("")
		return nil
	}

	return fmt.Errorf("Error reading: %s: %s", d.Id(), err)
}

func providerConfigure(data *schema.ResourceData, terraformVersion string) (interface{}, error) {
	config := Config{
		SkipCredsValidation: data.Get("skip_credentials_validation").(bool),
		Token:               data.Get("token").(string),
		UserAgent:           fmt.Sprintf("(%s %s) Terraform/%s", runtime.GOOS, runtime.GOARCH, terraformVersion),
	}

	log.Println("[INFO] Initializing PagerDuty client")
	return config.Client()
}
