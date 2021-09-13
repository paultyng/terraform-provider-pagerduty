package pagerduty

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyService() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyServiceCreate,
		Read:   resourcePagerDutyServiceRead,
		Update: resourcePagerDutyServiceUpdate,
		Delete: resourcePagerDutyServiceDelete,
		CustomizeDiff: func(diff *schema.ResourceDiff, i interface{}) error {
			in := diff.Get("incident_urgency_rule.#").(int)
			for i := 0; i <= in; i++ {
				t := diff.Get(fmt.Sprintf("incident_urgency_rule.%d.type", i)).(string)
				if t == "use_support_hours" && diff.Get(fmt.Sprintf("incident_urgency_rule.%d.urgency", i)).(string) != "" {
					return fmt.Errorf("general urgency cannot be set for a use_support_hours incident urgency rule type")
				}
			}
			return nil
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile(`^$|^[ ]+$|[/\\<>&]`), "Service name can't be blank or contain '\\', '/', '&', '<', '>' or non-printable characters. "),
			},
			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"alert_creation": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "create_incidents",
				ValidateFunc: validateValueFunc([]string{
					"create_alerts_and_incidents",
					"create_incidents",
				}),
			},
			"alert_grouping": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ValidateFunc: validateValueFunc([]string{
					"time",
					"intelligent",
					"rules",
				}),
			},
			"alert_grouping_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"alert_grouping_parameters": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validateValueFunc([]string{
								"time",
								"intelligent",
								"content_based",
							}),
						},
						"config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"timeout": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"fields": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"aggregate": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validateValueFunc([]string{
											"all",
											"any",
										}),
									},
								},
							},
						},
					},
				},
			},
			"auto_resolve_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "14400",
			},
			"last_incident_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"acknowledgement_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1800",
			},
			"escalation_policy": {
				Type:     schema.TypeString,
				Required: true,
			},
			"incident_urgency_rule": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"urgency": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"during_support_hours": {
							Type:     schema.TypeList,
							MaxItems: 1,
							MinItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"urgency": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"outside_support_hours": {
							Type:     schema.TypeList,
							MaxItems: 1,
							MinItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"urgency": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"support_hours": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"start_time": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"end_time": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"days_of_week": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 7,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
					},
				},
			},
			"scheduled_actions": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"to_urgency": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"at": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func buildServiceStruct(d *schema.ResourceData) (*pagerduty.Service, error) {
	service := pagerduty.Service{
		Name: d.Get("name").(string),
	}

	if attr, ok := d.GetOk("description"); ok {
		service.Description = attr.(string)
	}

	if attr, ok := d.GetOk("auto_resolve_timeout"); ok {
		if attr.(string) != "null" {
			if val, err := strconv.Atoi(attr.(string)); err == nil {
				service.AutoResolveTimeout = &val
			} else {
				return nil, err
			}
		}
	}

	if attr, ok := d.GetOk("acknowledgement_timeout"); ok {
		if attr.(string) != "null" {
			if val, err := strconv.Atoi(attr.(string)); err == nil {
				service.AcknowledgementTimeout = &val
			} else {
				return nil, err
			}
		}
	}

	if attr, ok := d.GetOk("alert_creation"); ok {
		service.AlertCreation = attr.(string)
	}

	if attr, ok := d.GetOk("alert_grouping"); ok {
		ag := attr.(string)
		service.AlertGrouping = &ag
	}
	if attr, ok := d.GetOk("alert_grouping_parameters"); ok {
		service.AlertGroupingParameters = expandAlertGroupingParameters(attr)
	} else {
		service.AlertGroupingParameters = &pagerduty.AlertGroupingParameters{}
	}
	// Using GetOkExists to allow for alert_grouping_timeout to be set to 0 if needed.
	if attr, ok := d.GetOkExists("alert_grouping_timeout"); ok {
		val := attr.(int)
		service.AlertGroupingTimeout = &val
	}

	if attr, ok := d.GetOk("escalation_policy"); ok {
		service.EscalationPolicy = &pagerduty.EscalationPolicyReference{
			ID:   attr.(string),
			Type: "escalation_policy_reference",
		}
	}

	if attr, ok := d.GetOk("incident_urgency_rule"); ok {
		service.IncidentUrgencyRule = expandIncidentUrgencyRule(attr)
		if service.IncidentUrgencyRule.Type == "use_support_hours" {
			service.ScheduledActions = make([]*pagerduty.ScheduledAction, 1)
		}
	}

	if attr, ok := d.GetOk("scheduled_actions"); ok {
		service.ScheduledActions = expandScheduledActions(attr)
	}

	if attr, ok := d.GetOk("support_hours"); ok {
		service.SupportHours = expandSupportHours(attr)
	}
	return &service, nil
}

func resourcePagerDutyServiceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	service, err := buildServiceStruct(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating PagerDuty service %s", service.Name)

	service, _, err = client.Services.Create(service)
	if err != nil {
		return err
	}

	d.SetId(service.ID)

	return resourcePagerDutyServiceRead(d, meta)
}

func resourcePagerDutyServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty service %s", d.Id())

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		service, _, err := client.Services.Get(d.Id(), &pagerduty.GetServiceOptions{})
		if err != nil {
			log.Printf("[WARN] Service read error")
			errResp := handleNotFoundError(err, d)
			if errResp != nil {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(errResp)
			}

			return nil
		}

		d.Set("name", service.Name)
		d.Set("html_url", service.HTMLURL)
		d.Set("status", service.Status)
		d.Set("created_at", service.CreatedAt)
		d.Set("escalation_policy", service.EscalationPolicy.ID)
		d.Set("description", service.Description)
		if service.AutoResolveTimeout == nil {
			d.Set("auto_resolve_timeout", "null")
		} else {
			d.Set("auto_resolve_timeout", strconv.Itoa(*service.AutoResolveTimeout))
		}
		d.Set("last_incident_timestamp", service.LastIncidentTimestamp)
		if service.AcknowledgementTimeout == nil {
			d.Set("acknowledgement_timeout", "null")
		} else {
			d.Set("acknowledgement_timeout", strconv.Itoa(*service.AcknowledgementTimeout))
		}
		d.Set("alert_creation", service.AlertCreation)
		if service.AlertGrouping != nil && *service.AlertGrouping != "" {
			d.Set("alert_grouping", *service.AlertGrouping)
		}
		if service.AlertGroupingTimeout == nil {
			d.Set("alert_grouping_timeout", "null")
		} else {
			d.Set("alert_grouping_timeout", *service.AlertGroupingTimeout)
		}
		if service.AlertGroupingParameters != nil {
			if err := d.Set("alert_grouping_parameters", flattenAlertGroupingParameters(service.AlertGroupingParameters)); err != nil {
				return resource.NonRetryableError(err)
			}
		}
		if service.IncidentUrgencyRule != nil {
			if err := d.Set("incident_urgency_rule", flattenIncidentUrgencyRule(service.IncidentUrgencyRule)); err != nil {
				return resource.NonRetryableError(err)
			}
		}

		if service.SupportHours != nil {
			if err := d.Set("support_hours", flattenSupportHours(service.SupportHours)); err != nil {
				return resource.NonRetryableError(err)
			}
		}

		if service.ScheduledActions != nil {
			if err := d.Set("scheduled_actions", flattenScheduledActions(service.ScheduledActions)); err != nil {
				return resource.NonRetryableError(err)
			}
		}
		return nil
	})
}

func resourcePagerDutyServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	service, err := buildServiceStruct(d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Updating PagerDuty service %s", d.Id())

	if _, _, err := client.Services.Update(d.Id(), service); err != nil {
		return err
	}

	return nil
}

func resourcePagerDutyServiceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty service %s", d.Id())

	if _, err := client.Services.Delete(d.Id()); err != nil {
		return err
	}

	d.SetId("")

	// giving the API time to catchup
	time.Sleep(time.Second)
	return nil
}
func expandAlertGroupingParameters(v interface{}) *pagerduty.AlertGroupingParameters {
	riur := v.([]interface{})[0].(map[string]interface{})
	alertGroupingParameters := &pagerduty.AlertGroupingParameters{
		Config: &pagerduty.AlertGroupingConfig{},
	}
	if len(riur["type"].(string)) > 0 {
		gt := riur["type"].(string)
		alertGroupingParameters.Type = &gt
	}

	if val, ok := riur["config"]; ok {
		alertGroupingParameters.Config = expandAlertGroupingConfig(val)
	}
	return alertGroupingParameters
}

func expandAlertGroupingConfig(v interface{}) *pagerduty.AlertGroupingConfig {
	alertGroupingConfig := &pagerduty.AlertGroupingConfig{}
	if len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
		return nil
	}
	riur := v.([]interface{})[0].(map[string]interface{})

	if val, ok := riur["fields"]; ok {
		for _, field := range val.([]interface{}) {
			alertGroupingConfig.Fields = append(alertGroupingConfig.Fields, field.(string))
		}
	}
	if val, ok := riur["aggregate"]; ok {
		agg := val.(string)
		alertGroupingConfig.Aggregate = &agg
	}
	if val, ok := riur["timeout"]; ok {
		to := val.(int)
		alertGroupingConfig.Timeout = &to
	}
	return alertGroupingConfig
}
func flattenAlertGroupingParameters(v *pagerduty.AlertGroupingParameters) interface{} {
	alertGroupingParameters := map[string]interface{}{"type": "", "config": []map[string]interface{}{{"aggregate": nil, "fields": nil, "timeout": nil}}}

	if v.Type != nil {
		alertGroupingParameters["type"] = v.Type
	}

	if v.Config != nil {
		alertGroupingParameters["config"] = flattenAlertGroupingConfig(v.Config)
	}

	return []interface{}{alertGroupingParameters}
}
func flattenAlertGroupingConfig(v *pagerduty.AlertGroupingConfig) interface{} {

	alertGroupingConfig := map[string]interface{}{
		"aggregate": v.Aggregate,
		"fields":    v.Fields,
		"timeout":   v.Timeout,
	}

	return []interface{}{alertGroupingConfig}
}
func expandIncidentUrgencyRule(v interface{}) *pagerduty.IncidentUrgencyRule {
	riur := v.([]interface{})[0].(map[string]interface{})
	incidentUrgencyRule := &pagerduty.IncidentUrgencyRule{
		Type: riur["type"].(string),
	}

	if val, ok := riur["urgency"]; ok {
		incidentUrgencyRule.Urgency = val.(string)
	}

	if val, ok := riur["during_support_hours"]; ok {
		if len(val.([]interface{})) > 0 {
			incidentUrgencyRule.DuringSupportHours = expandIncidentUrgencyType(val)
		}
	}

	if val, ok := riur["outside_support_hours"]; ok {
		if len(val.([]interface{})) > 0 {
			incidentUrgencyRule.OutsideSupportHours = expandIncidentUrgencyType(val)
		}
	}

	return incidentUrgencyRule
}

func flattenIncidentUrgencyRule(v *pagerduty.IncidentUrgencyRule) []interface{} {
	incidentUrgencyRule := map[string]interface{}{
		"type":    v.Type,
		"urgency": v.Urgency,
	}

	if v.DuringSupportHours != nil {
		incidentUrgencyRule["during_support_hours"] = flattenIncidentUrgencyType(v.DuringSupportHours)
	}

	if v.OutsideSupportHours != nil {
		incidentUrgencyRule["outside_support_hours"] = flattenIncidentUrgencyType(v.OutsideSupportHours)
	}

	return []interface{}{incidentUrgencyRule}
}

func expandIncidentUrgencyType(v interface{}) *pagerduty.IncidentUrgencyType {
	riut := v.([]interface{})[0].(map[string]interface{})
	incidentUrgencyType := &pagerduty.IncidentUrgencyType{}

	if v, ok := riut["type"]; ok {
		incidentUrgencyType.Type = v.(string)
	}

	if v, ok := riut["urgency"]; ok {
		incidentUrgencyType.Urgency = v.(string)
	}

	return incidentUrgencyType
}

func flattenIncidentUrgencyType(v *pagerduty.IncidentUrgencyType) []interface{} {
	incidentUrgencyType := map[string]interface{}{
		"type":    v.Type,
		"urgency": v.Urgency,
	}
	return []interface{}{incidentUrgencyType}
}

func expandSupportHours(v interface{}) *pagerduty.SupportHours {
	supportHours := &pagerduty.SupportHours{}

	rsh := v.([]interface{})[0].(map[string]interface{})

	if v, ok := rsh["type"]; ok {
		supportHours.Type = v.(string)
	}

	if v, ok := rsh["time_zone"]; ok {
		supportHours.TimeZone = v.(string)
	}

	if v, ok := rsh["start_time"]; ok {
		supportHours.StartTime = v.(string)
	}

	if v, ok := rsh["end_time"]; ok {
		supportHours.EndTime = v.(string)
	}

	if v, ok := rsh["days_of_week"]; ok {
		var daysOfWeek []int

		for _, dof := range v.([]interface{}) {
			daysOfWeek = append(daysOfWeek, dof.(int))
		}

		supportHours.DaysOfWeek = daysOfWeek
	}

	return supportHours
}

func flattenSupportHours(v *pagerduty.SupportHours) []interface{} {
	supportHours := map[string]interface{}{}

	if v.Type != "" {
		supportHours["type"] = v.Type
	}

	if v.TimeZone != "" {
		supportHours["time_zone"] = v.TimeZone
	}

	if v.StartTime != "" {
		supportHours["start_time"] = v.StartTime
	}

	if v.EndTime != "" {
		supportHours["end_time"] = v.EndTime
	}

	if len(v.DaysOfWeek) > 0 {
		supportHours["days_of_week"] = v.DaysOfWeek
	}

	return []interface{}{supportHours}
}

func expandScheduledActions(v interface{}) []*pagerduty.ScheduledAction {
	var scheduledActions []*pagerduty.ScheduledAction

	for _, sa := range v.([]interface{}) {
		rsa := sa.(map[string]interface{})

		scheduledAction := &pagerduty.ScheduledAction{
			Type:      rsa["type"].(string),
			ToUrgency: rsa["to_urgency"].(string),
			At:        expandScheduledActionAt(rsa["at"]),
		}

		scheduledActions = append(scheduledActions, scheduledAction)
	}

	return scheduledActions
}

func flattenScheduledActions(v []*pagerduty.ScheduledAction) []interface{} {
	var scheduledActions []interface{}

	for _, sa := range v {
		scheduledAction := map[string]interface{}{
			"type":       sa.Type,
			"to_urgency": sa.ToUrgency,
			"at":         flattenScheduledActionAt(sa.At),
		}
		scheduledActions = append(scheduledActions, scheduledAction)
	}

	return scheduledActions
}

func expandScheduledActionAt(v interface{}) *pagerduty.At {
	rat := v.([]interface{})[0].(map[string]interface{})
	return &pagerduty.At{
		Type: rat["type"].(string),
		Name: rat["name"].(string),
	}
}

func flattenScheduledActionAt(v *pagerduty.At) []interface{} {
	at := map[string]interface{}{"type": v.Type, "name": v.Name}
	return []interface{}{at}
}
