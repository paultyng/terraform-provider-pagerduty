package pagerduty

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutyServiceIntegration() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyServiceIntegrationCreate,
		Read:   resourcePagerDutyServiceIntegrationRead,
		Update: resourcePagerDutyServiceIntegrationUpdate,
		Delete: resourcePagerDutyServiceIntegrationDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePagerDutyServiceIntegrationImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"service": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				ConflictsWith: []string{"vendor"},
				ValidateFunc: validateValueFunc([]string{
					"aws_cloudwatch_inbound_integration",
					"cloudkick_inbound_integration",
					"event_transformer_api_inbound_integration",
					"events_api_v2_inbound_integration",
					"generic_email_inbound_integration",
					"generic_events_api_inbound_integration",
					"keynote_inbound_integration",
					"nagios_inbound_integration",
					"pingdom_inbound_integration",
					"sql_monitor_inbound_integration",
				}),
			},
			"vendor": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"type"},
				Computed:      true,
			},
			"integration_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"integration_email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"html_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email_incident_creation": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"email_filter_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"email_parsing_fallback": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"email_parser": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validateValueFunc([]string{
								"resolve",
								"trigger",
							}),
						},
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"match_predicate": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: false,
							MaxItems: 1,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"predicate": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: false,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"matcher": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"part": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validateValueFunc([]string{
														"body",
														"from_addresses",
														"subject",
													}),
												},
												"predicate": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"matcher": {
																Type:     schema.TypeString,
																Required: true,
															},
															"part": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validateValueFunc([]string{
																	"body",
																	"from_addresses",
																	"subject",
																}),
															},
															"type": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validateValueFunc([]string{
																	"contains",
																	"exactly",
																	"regex",
																}),
															},
														},
													},
												},
												"type": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validateValueFunc([]string{
														"contains",
														"exactly",
														"not",
														"regex",
													}),
												},
											},
										},
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validateValueFunc([]string{
											"all",
											"any",
										}),
									},
								},
							},
						},
						"value_extractor": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ends_before": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"part": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validateValueFunc([]string{
											"body",
											"subject",
										}),
									},
									"regex": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"starts_after": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validateValueFunc([]string{
											"between",
											"entire",
											"regex",
										}),
									},
									"value_name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"email_filter": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subject_mode": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validateValueFunc([]string{
								"always",
								"match",
								"no-match",
							}),
						},
						"subject_regex": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"body_mode": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validateValueFunc([]string{
								"always",
								"match",
								"no-match",
							}),
						},
						"body_regex": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"from_email_mode": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validateValueFunc([]string{
								"always",
								"match",
								"no-match",
							}),
						},
						"from_email_regex": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func buildServiceIntegrationStruct(d *schema.ResourceData) *pagerduty.Integration {
	serviceIntegration := &pagerduty.Integration{
		Name: d.Get("name").(string),
		Type: "service_integration",
		Service: &pagerduty.ServiceReference{
			Type: "service",
			ID:   d.Get("service").(string),
		},
	}

	if attr, ok := d.GetOk("email_incident_creation"); ok {
		serviceIntegration.EmailIncidentCreation = attr.(string)
	}

	if attr, ok := d.GetOk("email_filter_mode"); ok {
		serviceIntegration.EmailFilterMode = attr.(string)
	}

	if attr, ok := d.GetOk("email_parsing_fallback"); ok {
		serviceIntegration.EmailParsingFallback = attr.(string)
	}

	if attr, ok := d.GetOk("email_parser"); ok {
		parcers, err := expandEmailParsers(attr)
		if err != nil {
			log.Printf("[ERR] Parce PagerDuty service integration email parcers fail %s", err)
		}
		serviceIntegration.EmailParsers = parcers
	}

	if attr, ok := d.GetOk("email_filter"); ok {
		filters, err := expandEmailFilters(attr)
		if err != nil {
			log.Printf("[ERR] Parce PagerDuty service integration email filters fail %s", err)
		}
		serviceIntegration.EmailFilters = filters
	}

	if attr, ok := d.GetOk("integration_key"); ok {
		serviceIntegration.IntegrationKey = attr.(string)
	}

	if attr, ok := d.GetOk("integration_email"); ok {
		serviceIntegration.IntegrationEmail = attr.(string)
	}

	if attr, ok := d.GetOk("type"); ok {
		serviceIntegration.Type = attr.(string)
	}

	if attr, ok := d.GetOk("vendor"); ok {
		serviceIntegration.Vendor = &pagerduty.VendorReference{
			ID:   attr.(string),
			Type: "vendor",
		}
	}

	return serviceIntegration
}

func expandEmailParsers(v interface{}) ([]*pagerduty.EmailParser, error) {
	var emailParsers []*pagerduty.EmailParser

	for _, ep := range v.([]interface{}) {
		rep := ep.(map[string]interface{})

		repid := rep["id"].(int)
		emailParser := &pagerduty.EmailParser{
			ID:     &repid,
			Action: rep["action"].(string),
		}

		mp := rep["match_predicate"].([]interface{})[0].(map[string]interface{})

		matchPredicate := &pagerduty.MatchPredicate{
			Type: mp["type"].(string),
		}

		for _, p := range mp["predicate"].([]interface{}) {
			rp := p.(map[string]interface{})

			predicate := &pagerduty.Predicate{
				Type: rp["type"].(string),
			}
			if predicate.Type == "not" {
				mp := rp["predicate"].([]interface{})[0].(map[string]interface{})
				predicate2 := &pagerduty.Predicate{
					Type:    mp["type"].(string),
					Part:    mp["part"].(string),
					Matcher: mp["matcher"].(string),
				}
				predicate.Predicates = append(predicate.Predicates, predicate2)
			} else {
				predicate.Part = rp["part"].(string)
				predicate.Matcher = rp["matcher"].(string)
			}

			matchPredicate.Predicates = append(matchPredicate.Predicates, predicate)
		}

		emailParser.MatchPredicate = matchPredicate

		if rep["value_extractor"] != nil {
			for _, ve := range rep["value_extractor"].([]interface{}) {
				rve := ve.(map[string]interface{})

				extractor := &pagerduty.ValueExtractor{
					Type:      rve["type"].(string),
					ValueName: rve["value_name"].(string),
					Part:      rve["part"].(string),
				}

				if extractor.Type == "regex" {
					extractor.Regex = rve["regex"].(string)
				} else {
					extractor.StartsAfter = rve["starts_after"].(string)
					extractor.EndsBefore = rve["ends_before"].(string)
				}

				emailParser.ValueExtractors = append(emailParser.ValueExtractors, extractor)
			}
		}

		emailParsers = append(emailParsers, emailParser)
	}

	return emailParsers, nil
}

func expandEmailFilters(v interface{}) ([]*pagerduty.EmailFilter, error) {
	var emailFilters []*pagerduty.EmailFilter

	for _, ef := range v.([]interface{}) {
		ref := ef.(map[string]interface{})

		emailFilter := &pagerduty.EmailFilter{
			ID:             ref["id"].(string),
			SubjectMode:    ref["subject_mode"].(string),
			SubjectRegex:   ref["subject_regex"].(string),
			BodyMode:       ref["body_mode"].(string),
			BodyRegex:      ref["body_regex"].(string),
			FromEmailMode:  ref["from_email_mode"].(string),
			FromEmailRegex: ref["from_email_regex"].(string),
		}

		emailFilters = append(emailFilters, emailFilter)
	}

	return emailFilters, nil
}

func resourcePagerDutyServiceIntegrationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	serviceIntegration := buildServiceIntegrationStruct(d)

	log.Printf("[INFO] Creating PagerDuty service integration %s", serviceIntegration.Name)

	service := d.Get("service").(string)

	retryErr := resource.Retry(1*time.Minute, func() *resource.RetryError {
		if serviceIntegration, _, err := client.Services.CreateIntegration(service, serviceIntegration); err != nil {
			if isErrCode(err, 400) {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		} else if serviceIntegration != nil {
			d.SetId(serviceIntegration.ID)
		}
		return nil
	})

	if retryErr != nil {
		return retryErr
	}

	return resourcePagerDutyServiceIntegrationRead(d, meta)
}

func resourcePagerDutyServiceIntegrationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty service integration %s", d.Id())

	service := d.Get("service").(string)

	o := &pagerduty.GetIntegrationOptions{}

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		serviceIntegration, _, err := client.Services.GetIntegration(service, d.Id(), o)
		if err != nil {
			log.Printf("[WARN] Service integration read error")
			errResp := handleNotFoundError(err, d)
			if errResp != nil {
				time.Sleep(2 * time.Second)
				return resource.RetryableError(errResp)
			}

			return nil
		}

		if err := d.Set("name", serviceIntegration.Name); err != nil {
			return err
		}

		if err := d.Set("type", serviceIntegration.Type); err != nil {
			return err
		}

		if serviceIntegration.Service != nil {
			if err := d.Set("service", serviceIntegration.Service.ID); err != nil {
				return err
			}
		}

		if serviceIntegration.Vendor != nil {
			if err := d.Set("vendor", serviceIntegration.Vendor.ID); err != nil {
				return err
			}
		}

		if serviceIntegration.IntegrationKey != "" {
			if err := d.Set("integration_key", serviceIntegration.IntegrationKey); err != nil {
				return err
			}
		}

		if serviceIntegration.IntegrationEmail != "" {
			if err := d.Set("integration_email", serviceIntegration.IntegrationEmail); err != nil {
				return err
			}
		}

		if serviceIntegration.EmailIncidentCreation != "" {
			if err := d.Set("email_incident_creation", serviceIntegration.EmailIncidentCreation); err != nil {
				return err
			}
		}

		if serviceIntegration.EmailFilterMode != "" {
			if err := d.Set("email_filter_mode", serviceIntegration.EmailFilterMode); err != nil {
				return err
			}
		}

		if serviceIntegration.EmailParsingFallback != "" {
			if err := d.Set("email_parsing_fallback", serviceIntegration.EmailParsingFallback); err != nil {
				return err
			}
		}

		if serviceIntegration.HTMLURL != "" {
			if err := d.Set("html_url", serviceIntegration.HTMLURL); err != nil {
				return err
			}
		}

		if serviceIntegration.EmailFilters != nil {
			if err := d.Set("email_filter", flattenEmailFilters(serviceIntegration.EmailFilters)); err != nil {
				return err
			}
		}

		if serviceIntegration.EmailParsers != nil {
			if err := d.Set("email_parser", flattenEmailParsers(serviceIntegration.EmailParsers)); err != nil {
				return err
			}
		}

		return nil
	})
}

func flattenEmailFilters(v []*pagerduty.EmailFilter) []map[string]interface{} {
	var emailFilters []map[string]interface{}

	for _, ef := range v {
		emailFilter := map[string]interface{}{
			"id":               ef.ID,
			"subject_mode":     ef.SubjectMode,
			"subject_regex":    ef.SubjectRegex,
			"body_mode":        ef.BodyMode,
			"body_regex":       ef.BodyRegex,
			"from_email_mode":  ef.FromEmailMode,
			"from_email_regex": ef.FromEmailRegex,
		}

		emailFilters = append(emailFilters, emailFilter)
	}

	return emailFilters
}

func flattenEmailParsers(v []*pagerduty.EmailParser) []map[string]interface{} {
	var emailParsers []map[string]interface{}

	for _, ef := range v {
		emailParser := map[string]interface{}{
			"id":     ef.ID,
			"action": ef.Action,
		}

		matchPredicate := map[string]interface{}{
			"type": ef.MatchPredicate.Type,
		}

		var predicates []map[string]interface{}

		for _, p := range ef.MatchPredicate.Predicates {
			predicate := map[string]interface{}{
				"type": p.Type,
			}

			if p.Type == "not" && len(p.Predicates) > 0 {
				var predicates2 []map[string]interface{}
				predicate2 := map[string]interface{}{
					"type":    p.Predicates[0].Type,
					"part":    p.Predicates[0].Part,
					"matcher": p.Predicates[0].Matcher,
				}

				predicates2 = append(predicates2, predicate2)

				predicate["predicate"] = predicates2

			} else {
				predicate["part"] = p.Part
				predicate["matcher"] = p.Matcher
			}

			predicates = append(predicates, predicate)
		}

		matchPredicate["predicate"] = predicates

		emailParser["match_predicate"] = []interface{}{matchPredicate}

		var valueExtractors []map[string]interface{}

		for _, ve := range ef.ValueExtractors {
			extractor := map[string]interface{}{
				"type":         ve.Type,
				"value_name":   ve.ValueName,
				"part":         ve.Part,
				"starts_after": ve.StartsAfter,
				"ends_before":  ve.EndsBefore,
			}

			valueExtractors = append(valueExtractors, extractor)
		}

		emailParser["value_extractor"] = valueExtractors

		emailParsers = append(emailParsers, emailParser)
	}

	return emailParsers
}

func resourcePagerDutyServiceIntegrationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	serviceIntegration := buildServiceIntegrationStruct(d)

	service := d.Get("service").(string)

	log.Printf("[INFO] Updating PagerDuty service integration %s", d.Id())

	if _, _, err := client.Services.UpdateIntegration(service, d.Id(), serviceIntegration); err != nil {
		return err
	}

	return nil
}

func resourcePagerDutyServiceIntegrationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	service := d.Get("service").(string)

	log.Printf("[INFO] Removing PagerDuty service integration %s", d.Id())

	if _, err := client.Services.DeleteIntegration(service, d.Id()); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func resourcePagerDutyServiceIntegrationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*pagerduty.Client)

	ids := strings.Split(d.Id(), ".")

	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing pagerduty_service_integration. Expecting an importation ID formed as '<service_id>.<integration_id>'")
	}
	sid, id := ids[0], ids[1]

	_, _, err := client.Services.GetIntegration(sid, id, nil)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	// These are set because an import also calls Read behind the scenes
	d.SetId(id)
	if err := d.Set("service", sid); err != nil {
		return []*schema.ResourceData{}, err
	}

	return []*schema.ResourceData{d}, nil
}
