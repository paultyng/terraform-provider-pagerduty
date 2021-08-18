package pagerduty

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyServiceEventRule() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyServiceEventRuleCreate,
		Read:   resourcePagerDutyServiceEventRuleRead,
		Update: resourcePagerDutyServiceEventRuleUpdate,
		Delete: resourcePagerDutyServiceEventRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePagerDutyServiceEventRuleImport,
		},
		Schema: map[string]*schema.Schema{
			"service": {
				Type:     schema.TypeString,
				Required: true,
			},
			"position": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"conditions": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"operator": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"subconditions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"operator": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"parameter": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"path": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"value": {
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
				},
			},
			"time_frame": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scheduled_weekly": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"timezone": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"start_time": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"duration": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"weekdays": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
									},
								},
							},
						},
						"active_between": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start_time": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"end_time": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"actions": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"suppress": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"threshold_value": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"threshold_time_unit": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validateValueFunc([]string{
											"minutes",
											"seconds",
											"hours",
										}),
									},
									"threshold_time_amount": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"severity": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validateValueFunc([]string{
											"info",
											"error",
											"warning",
											"critical",
										}),
									},
								},
							},
						},
						"priority": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"annotate": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"event_action": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validateValueFunc([]string{
											"trigger",
											"resolve",
										}),
									},
								},
							},
						},
						"extractions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"source": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"regex": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"template": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"suspend": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"variable": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"parameters": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"value": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"path": {
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

func buildServiceEventRuleStruct(d *schema.ResourceData) *pagerduty.ServiceEventRule {
	rule := &pagerduty.ServiceEventRule{
		Service: &pagerduty.ServiceReference{
			Type: "service_reference",
			ID:   d.Get("service").(string),
		},
		Conditions: expandConditions(d.Get("conditions").([]interface{})),
	}

	if attr, ok := d.GetOk("actions"); ok {
		rule.Actions = expandActions(attr.([]interface{}))
	}
	if attr, ok := d.GetOk("time_frame"); ok {
		rule.TimeFrame = expandTimeFrame(attr.([]interface{}))
	}
	if attr, ok := d.GetOk("variable"); ok {
		rule.Variables = expandRuleVariables(attr.([]interface{}))
	}

	pos := d.Get("position").(int)
	rule.Position = &pos

	if attr, ok := d.GetOk("disabled"); ok {
		rule.Disabled = attr.(bool)
	}

	return rule
}
func resourcePagerDutyServiceEventRuleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	rule := buildServiceEventRuleStruct(d)

	log.Printf("[INFO] Creating PagerDuty service event rule for service: %s", rule.Service.ID)

	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if rule, _, err := client.Services.CreateEventRule(rule.Service.ID, rule); err != nil {
			return resource.RetryableError(err)
		} else if rule != nil {
			d.SetId(rule.ID)
			// Verifying the position that was defined in terraform is the same position set in PagerDuty
			pos := d.Get("position").(int)
			if *rule.Position != pos {
				if err := resourcePagerDutyServiceEventRuleUpdate(d, meta); err != nil {
					return resource.NonRetryableError(err)
				}
			}
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}
	return resourcePagerDutyServiceEventRuleRead(d, meta)
}

func resourcePagerDutyServiceEventRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty service event rule: %s", d.Id())
	serviceID := d.Get("service").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		if rule, _, err := client.Services.GetEventRule(serviceID, d.Id()); err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
		} else if rule != nil {
			if rule.Conditions != nil {
				d.Set("conditions", flattenConditions(rule.Conditions))
			}
			if rule.Actions != nil {
				d.Set("actions", flattenActions(rule.Actions))
			}
			if rule.TimeFrame != nil {
				d.Set("time_frame", flattenTimeFrame(rule.TimeFrame))
			}
			if rule.Variables != nil {
				d.Set("variable", flattenRuleVariables(rule.Variables))
			}
			d.Set("position", rule.Position)
			d.Set("disabled", rule.Disabled)
			d.Set("service", serviceID)
		}
		return nil
	})
}

func resourcePagerDutyServiceEventRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	rule := buildServiceEventRuleStruct(d)

	log.Printf("[INFO] Updating PagerDuty service event rule: %s", d.Id())
	serviceID := d.Get("service").(string)

	retryErr := resource.Retry(30*time.Second, func() *resource.RetryError {
		if updatedRule, _, err := client.Services.UpdateEventRule(serviceID, d.Id(), rule); err != nil {
			return resource.RetryableError(err)
		} else if rule.Position != nil && *updatedRule.Position != *rule.Position {
			log.Printf("[INFO] Service Event Rule %s position %v needs to be %v", updatedRule.ID, *updatedRule.Position, *rule.Position)
			return resource.RetryableError(fmt.Errorf("Error updating service event rule %s position %d needs to be %d", updatedRule.ID, *updatedRule.Position, *rule.Position))
		}

		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}
	return nil
}

func resourcePagerDutyServiceEventRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty service event rule: %s", d.Id())
	serviceID := d.Get("service").(string)

	retryErr := resource.Retry(30*time.Second, func() *resource.RetryError {
		if _, err := client.Services.DeleteEventRule(serviceID, d.Id()); err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}
	d.SetId("")

	return nil
}

func resourcePagerDutyServiceEventRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*pagerduty.Client)

	ids := strings.Split(d.Id(), ".")

	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing pagerduty_service_event_rule. Expecting an importation ID formed as '<service_id>.<service_event_rule_id>'")
	}
	serviceID, ruleID := ids[0], ids[1]

	_, _, err := client.Services.GetEventRule(serviceID, ruleID)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(ruleID)
	d.Set("service", serviceID)

	return []*schema.ResourceData{d}, nil
}
