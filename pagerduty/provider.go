package pagerduty

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
	"github.com/heimweh/go-pagerduty/persistentconfig"
)

const (
	IsMuxed    = true
	IsNotMuxed = false
)

// Provider represents a resource provider in Terraform
func Provider(isMux bool) *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"skip_credentials_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PAGERDUTY_TOKEN", nil),
			},

			"user_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PAGERDUTY_USER_TOKEN", nil),
			},

			"service_region": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PAGERDUTY_SERVICE_REGION", ""),
			},

			"use_app_oauth_scoped_token": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pd_client_id": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("PAGERDUTY_CLIENT_ID", nil),
						},
						"pd_client_secret": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("PAGERDUTY_CLIENT_SECRET", nil),
						},
						"pd_subdomain": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("PAGERDUTY_SUBDOMAIN", nil),
						},
					},
				},
			},

			"api_url_override": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"insecure_tls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"pagerduty_escalation_policy":                          dataSourcePagerDutyEscalationPolicy(),
			"pagerduty_licenses":                                   dataSourcePagerDutyLicenses(),
			"pagerduty_user_contact_method":                        dataSourcePagerDutyUserContactMethod(),
			"pagerduty_team":                                       dataSourcePagerDutyTeam(),
			"pagerduty_vendor":                                     dataSourcePagerDutyVendor(),
			"pagerduty_service":                                    dataSourcePagerDutyService(),
			"pagerduty_service_integration":                        dataSourcePagerDutyServiceIntegration(),
			"pagerduty_automation_actions_action":                  dataSourcePagerDutyAutomationActionsAction(),
			"pagerduty_automation_actions_runner":                  dataSourcePagerDutyAutomationActionsRunner(),
			"pagerduty_business_service":                           dataSourcePagerDutyBusinessService(),
			"pagerduty_event_orchestration":                        dataSourcePagerDutyEventOrchestration(),
			"pagerduty_event_orchestration_global_cache_variable":  dataSourcePagerDutyEventOrchestrationGlobalCacheVariable(),
			"pagerduty_event_orchestration_integration":            dataSourcePagerDutyEventOrchestrationIntegration(),
			"pagerduty_event_orchestration_service_cache_variable": dataSourcePagerDutyEventOrchestrationServiceCacheVariable(),
			"pagerduty_event_orchestrations":                       dataSourcePagerDutyEventOrchestrations(),
			"pagerduty_incident_custom_field":                      dataSourcePagerDutyIncidentCustomField(),
			"pagerduty_incident_workflow":                          dataSourcePagerDutyIncidentWorkflow(),
			"pagerduty_priority":                                   dataSourcePagerDutyPriority(),
			"pagerduty_ruleset":                                    dataSourcePagerDutyRuleset(),
			"pagerduty_team_members":                               dataSourcePagerDutyTeamMembers(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"pagerduty_addon":                                         resourcePagerDutyAddon(),
			"pagerduty_escalation_policy":                             resourcePagerDutyEscalationPolicy(),
			"pagerduty_maintenance_window":                            resourcePagerDutyMaintenanceWindow(),
			"pagerduty_schedule":                                      resourcePagerDutySchedule(),
			"pagerduty_service":                                       resourcePagerDutyService(),
			"pagerduty_service_integration":                           resourcePagerDutyServiceIntegration(),
			"pagerduty_team":                                          resourcePagerDutyTeam(),
			"pagerduty_team_membership":                               resourcePagerDutyTeamMembership(),
			"pagerduty_user":                                          resourcePagerDutyUser(),
			"pagerduty_user_contact_method":                           resourcePagerDutyUserContactMethod(),
			"pagerduty_event_rule":                                    resourcePagerDutyEventRule(),
			"pagerduty_ruleset":                                       resourcePagerDutyRuleset(),
			"pagerduty_ruleset_rule":                                  resourcePagerDutyRulesetRule(),
			"pagerduty_business_service":                              resourcePagerDutyBusinessService(),
			"pagerduty_response_play":                                 resourcePagerDutyResponsePlay(),
			"pagerduty_service_event_rule":                            resourcePagerDutyServiceEventRule(),
			"pagerduty_slack_connection":                              resourcePagerDutySlackConnection(),
			"pagerduty_business_service_subscriber":                   resourcePagerDutyBusinessServiceSubscriber(),
			"pagerduty_webhook_subscription":                          resourcePagerDutyWebhookSubscription(),
			"pagerduty_event_orchestration":                           resourcePagerDutyEventOrchestration(),
			"pagerduty_event_orchestration_integration":               resourcePagerDutyEventOrchestrationIntegration(),
			"pagerduty_event_orchestration_global":                    resourcePagerDutyEventOrchestrationPathGlobal(),
			"pagerduty_event_orchestration_router":                    resourcePagerDutyEventOrchestrationPathRouter(),
			"pagerduty_event_orchestration_unrouted":                  resourcePagerDutyEventOrchestrationPathUnrouted(),
			"pagerduty_event_orchestration_service":                   resourcePagerDutyEventOrchestrationPathService(),
			"pagerduty_event_orchestration_global_cache_variable":     resourcePagerDutyEventOrchestrationGlobalCacheVariable(),
			"pagerduty_event_orchestration_service_cache_variable":    resourcePagerDutyEventOrchestrationServiceCacheVariable(),
			"pagerduty_automation_actions_runner":                     resourcePagerDutyAutomationActionsRunner(),
			"pagerduty_automation_actions_action":                     resourcePagerDutyAutomationActionsAction(),
			"pagerduty_automation_actions_action_team_association":    resourcePagerDutyAutomationActionsActionTeamAssociation(),
			"pagerduty_automation_actions_runner_team_association":    resourcePagerDutyAutomationActionsRunnerTeamAssociation(),
			"pagerduty_incident_workflow":                             resourcePagerDutyIncidentWorkflow(),
			"pagerduty_incident_workflow_trigger":                     resourcePagerDutyIncidentWorkflowTrigger(),
			"pagerduty_automation_actions_action_service_association": resourcePagerDutyAutomationActionsActionServiceAssociation(),
			"pagerduty_incident_custom_field":                         resourcePagerDutyIncidentCustomField(),
			"pagerduty_incident_custom_field_option":                  resourcePagerDutyIncidentCustomFieldOption(),
		},
	}

	if isMux {
		delete(p.DataSourcesMap, "pagerduty_business_service")
		delete(p.DataSourcesMap, "pagerduty_escalation_policy")
		delete(p.DataSourcesMap, "pagerduty_licenses")
		delete(p.DataSourcesMap, "pagerduty_priority")
		delete(p.DataSourcesMap, "pagerduty_service")
		delete(p.DataSourcesMap, "pagerduty_service_integration")
		delete(p.DataSourcesMap, "pagerduty_vendor")

		delete(p.ResourcesMap, "pagerduty_addon")
		delete(p.ResourcesMap, "pagerduty_business_service")
		delete(p.ResourcesMap, "pagerduty_team")
		delete(p.ResourcesMap, "pagerduty_team_membership")
	}

	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigureContextFunc(ctx, d, terraformVersion)
	}

	return p
}

func isErrCode(err error, code int) bool {
	if e, ok := err.(*pagerduty.Error); ok && e.ErrorResponse.Response.StatusCode == code {
		return true
	}

	return false
}

func isMalformedNotFoundError(err error) bool {
	// There are some errors that doesn't stick to expected error interface and
	// fallback to a simple text error message that can be capture by this regexp.
	if err == nil {
		return false
	}

	re := regexp.MustCompile(".*: 404 Not Found$")
	return re.Match([]byte(err.Error()))
}

func isMalformedForbiddenError(err error) bool {
	if err == nil {
		return false
	}

	re := regexp.MustCompile(".*: 403 Forbidden$")
	return re.Match([]byte(err.Error()))
}

func genError(err error, d *schema.ResourceData) error {
	return fmt.Errorf("Error reading: %s: %s", d.Id(), err)
}

func handleNotFoundError(err error, d *schema.ResourceData) error {
	if isErrCode(err, 404) || isMalformedNotFoundError(err) {
		log.Printf("[WARN] Removing %s because it's gone", d.Id())
		d.SetId("")
		return nil
	}
	return genError(err, d)
}

func providerConfigureContextFunc(_ context.Context, data *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	serviceRegion := strings.ToLower(data.Get("service_region").(string))

	var regionApiUrl string
	if serviceRegion == "us" || serviceRegion == "" {
		regionApiUrl = ""
	} else {
		regionApiUrl = serviceRegion + "."
	}

	config := Config{
		ApiUrl:              "https://api." + regionApiUrl + "pagerduty.com",
		AppUrl:              "https://app." + regionApiUrl + "pagerduty.com",
		SkipCredsValidation: data.Get("skip_credentials_validation").(bool),
		Token:               data.Get("token").(string),
		UserToken:           data.Get("user_token").(string),
		UserAgent:           fmt.Sprintf("(%s %s) Terraform/%s", runtime.GOOS, runtime.GOARCH, terraformVersion),
		ApiUrlOverride:      data.Get("api_url_override").(string),
		ServiceRegion:       serviceRegion,
		InsecureTls:         data.Get("insecure_tls").(bool),
	}

	useAuthTokenType := pagerduty.AuthTokenTypeAPIToken
	if attr, ok := data.GetOk("use_app_oauth_scoped_token"); ok {
		config.AppOauthScopedTokenParams = expandAppOauthTokenParams(attr)
		config.AppOauthScopedTokenParams.Region = serviceRegion
		useAuthTokenType = pagerduty.AuthTokenTypeUseAppCredentials
		if err := validateAuthMethodConfig(data); err != nil {
			diag := diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "`token` and `use_app_oauth_scoped_token` are both configured at the same time",
				Detail:   err.Error(),
			}
			diags = append(diags, diag)
		}
	}

	config.APITokenType = &useAuthTokenType

	log.Println("[INFO] Initializing PagerDuty client")
	return &config, diags
}

func expandAppOauthTokenParams(v interface{}) *persistentconfig.AppOauthScopedTokenParams {
	aotp := &persistentconfig.AppOauthScopedTokenParams{}

	i := v.([]interface{})[0]
	if isNilFunc(i) {
		return nil
	}
	mi := i.(map[string]interface{})

	aotp.ClientID = mi["pd_client_id"].(string)
	aotp.ClientSecret = mi["pd_client_secret"].(string)
	aotp.PDSubDomain = mi["pd_subdomain"].(string)

	return aotp
}

var validationAuthMethodConfigWarning = "PagerDuty Provider has been set to authenticate API calls utilizing API token and App Oauth token at same time, in this scenario the use of App Oauth token is prioritised over API token authentication configuration. It is recommended to explicitely set just one authentication method.\nWe also suggest you to check your environment variables in case `token` being automatically read by Provider configuration through `PAGERDUTY_TOKEN` environment variable."

func validateAuthMethodConfig(data *schema.ResourceData) error {
	_, isSetAPIToken := data.GetOk("token")
	_, isSetUseAppOauthScopedToken := data.GetOk("use_app_oauth_scoped_token")

	if isSetUseAppOauthScopedToken && isSetAPIToken {
		return fmt.Errorf(validationAuthMethodConfigWarning)
	}

	return nil
}
