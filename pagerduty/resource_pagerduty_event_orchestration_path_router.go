package pagerduty

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
	"log"
	"time"
)

func resourcePagerDutyEventOrchestrationPathRouter() *schema.Resource {
	return &schema.Resource{
		Read:   resourcePagerDutyEventOrchestrationPathRouterRead,
		Create: resourcePagerDutyEventOrchestrationPathRouterCreate,
		Update: resourcePagerDutyEventOrchestrationPathRouterUpdate,
		Delete: resourcePagerDutyEventOrchestrationPathRouterUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough, //TODO: resourcePagerDutyEventOrchestrationPathImport
		},
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"parent": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: PagerDutyEventOrchestrationPathParent,
				},
			},
			// "self": {
			// 	Type:     schema.TypeString,
			// 	Optional: true,
			// },
			"sets": {
				Type:     schema.TypeList,
				Required: true, //TODO: is it always going to have a set?
				//MaxItems: 1,    // TODO:Router can only have 'start' set, but not having max will help repurpose set code snippet
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"rules": {
							Type:     schema.TypeList,
							Required: true, // even if there are no rules, API returns rules as an empty list
							// MaxItems: 1000, // TODO: do we need this?! Router allows a max of 1000 rules
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true, // If the start set has no rules, empty list is returned by API for rules. TODO: there is a validation on id
									},
									"label": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"conditions": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: PagerDutyEventOrchestrationPathConditions,
										},
									},
									"actions": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1, //there can only be one action for router
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"route_to": {
													Type:     schema.TypeString,
													Required: true,
													//TODO: validate func, cannot be unrouted, should be some serviceID
												},
											},
										},
									},
									"disabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			// "catch_all": {
			// 	Type:     schema.TypeList,
			// 	Optional: true, //if not supplied, API creates it
			// 	MaxItems: 1,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"actions": {
			// 				Type:     schema.TypeList,
			// 				Optional: true, //if not provided, API defaults to unrouted
			// 				MaxItems: 1,
			// 				Elem: &schema.Resource{
			// 					Schema: map[string]*schema.Schema{
			// 						"route_to": {
			// 							Type:     schema.TypeString,
			// 							Optional: true, //if not provided, API defaults to unrouted
			// 						},
			// 					},
			// 				},
			// 			},
			// 		},
			// 	},
			// },
		},
	}
}

func resourcePagerDutyEventOrchestrationPathRouterRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		path := buildRouterPathStruct(d)
		log.Printf("[INFO] Reading PagerDuty Event Orchestration Path of type: %s for orchestration: %s", "router", path.Parent.ID)

		if routerPath, _, err := client.EventOrchestrationPaths.Get(path.Parent.ID, path.Type); err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
		} else if routerPath != nil {
			d.SetId(path.Parent.ID)
			// d.Set("type", routerPath.Type)
			// d.Set("self", path.Parent.Self+"/"+routerPath.Type)
			if routerPath.Sets != nil {
				d.Set("sets", flattenSets(routerPath.Sets))
			}
		}
		return nil
	})

}

func resourcePagerDutyEventOrchestrationPathRouterCreate(d *schema.ResourceData, meta interface{}) error {
	return resourcePagerDutyEventOrchestrationPathRouterUpdate(d, meta)
}

func resourcePagerDutyEventOrchestrationPathRouterUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	updatePath := buildRouterPathStructForUpdate(d)

	log.Printf("[INFO] Updating PagerDuty EventOrchestrationPath of type: %s for orchestration: %s", "router", updatePath.Parent.ID)

	return performRouterPathUpdate(d, updatePath, client)
}

func resourcePagerDutyEventOrchestrationPathRouterDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func performRouterPathUpdate(d *schema.ResourceData, routerPath *pagerduty.EventOrchestrationPath, client *pagerduty.Client) error {
	retryErr := resource.Retry(30*time.Second, func() *resource.RetryError {
		updatedPath, _, err := client.EventOrchestrationPaths.Update(routerPath.Parent.ID, "router", routerPath)
		if err != nil {
			return resource.RetryableError(err)
		}
		if updatedPath == nil {
			return resource.NonRetryableError(fmt.Errorf("No Event Orchestration Router found."))
		}
		// set props
		d.SetId(routerPath.Parent.ID)
		if routerPath.Sets != nil {
			d.Set("sets", flattenSets(routerPath.Sets))
		}

		//TODO: figure out rule ordering
		// else if rule.Position != nil && *updatedRouterPath.Position != *rule.Position && rule.CatchAll != true {
		// 	log.Printf("[INFO] PagerDuty ruleset rule %s position %d needs to be %d", updatedRouterPath.ID, *updatedRouterPath.Position, *rule.Position)
		// 	return resource.RetryableError(fmt.Errorf("Error updating ruleset rule %s position %d needs to be %d", updatedRouterPath.ID, *updatedRouterPath.Position, *rule.Position))
		// }
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}
	return nil
}

func buildRouterPathStruct(d *schema.ResourceData) *pagerduty.EventOrchestrationPath {
	orchPath := &pagerduty.EventOrchestrationPath{
		Type: d.Get("type").(string),
	}

	if attr, ok := d.GetOk("parent"); ok {
		orchPath.Parent = expandOrchestrationPathParent(attr)
	}

	return orchPath
}

func buildRouterPathStructForUpdate(d *schema.ResourceData) *pagerduty.EventOrchestrationPath {

	// get the path-parent
	orchPath := &pagerduty.EventOrchestrationPath{}

	if attr, ok := d.GetOk("parent"); ok {
		orchPath.Parent = expandOrchestrationPathParent(attr)
	}

	// build other props
	// if attr, ok := d.GetOk("self"); ok {
	// 	orchPath.Self = attr.(string)
	// } else {
	// 	orchPath.Self = orchPath.Parent.Self + "/" + orchPath.Type
	// }

	if attr, ok := d.GetOk("sets"); ok {
		orchPath.Sets = expandSets(attr.([]interface{}))
	}

	return orchPath
}

func expandOrchestrationPathParent(v interface{}) *pagerduty.EventOrchestrationPathReference {
	var parent *pagerduty.EventOrchestrationPathReference
	p := v.([]interface{})[0].(map[string]interface{})
	parent = &pagerduty.EventOrchestrationPathReference{
		ID:   p["id"].(string),
		Type: p["type"].(string),
		Self: p["self"].(string),
	}

	return parent
}

func expandSets(v interface{}) []*pagerduty.EventOrchestrationPathSet {
	var sets []*pagerduty.EventOrchestrationPathSet

	for _, set := range v.([]interface{}) {
		s := set.(map[string]interface{})

		orchPathSet := &pagerduty.EventOrchestrationPathSet{
			ID:    s["id"].(string),
			Rules: expandRules(s["rules"].(interface{})),
		}

		sets = append(sets, orchPathSet)
	}

	return sets
}

func expandRules(v interface{}) []*pagerduty.EventOrchestrationPathRule {
	var rules []*pagerduty.EventOrchestrationPathRule

	for _, rule := range v.([]interface{}) {
		r := rule.(map[string]interface{})

		ruleInSet := &pagerduty.EventOrchestrationPathRule{
			ID:       r["id"].(string),
			Label:    r["label"].(string),
			Disabled: r["disabled"].(bool),
			// TODO:
			Conditions: expandRouterConditions(r["conditions"].(interface{})),
			Actions:    expandRouterActions(r["actions"].([]interface{})),
		}

		rules = append(rules, ruleInSet)
	}

	return rules
}

func expandRouterActions(v interface{}) *pagerduty.EventOrchestrationPathRuleActions {
	var actions = new(pagerduty.EventOrchestrationPathRuleActions)
	for _, ai := range v.([]interface{}) {
		am := ai.(map[string]interface{})
		actions.RouteTo = am["route_to"].(string)
	}
	// am := v.([]interface{})
	// actions.RouteTo = am[0]["route_to"]

	return actions
}

func expandRouterConditions(v interface{}) []*pagerduty.EventOrchestrationPathRuleCondition {
	var conditions []*pagerduty.EventOrchestrationPathRuleCondition

	for _, cond := range v.([]interface{}) {
		c := cond.(map[string]interface{})

		cx := &pagerduty.EventOrchestrationPathRuleCondition{
			Expression: c["expression"].(string),
		}

		conditions = append(conditions, cx)
	}

	return conditions
}
func flattenSets(orchPathSets []*pagerduty.EventOrchestrationPathSet) []interface{} {
	var flattenedSets []interface{}

	for _, set := range orchPathSets {
		flattenedSet := map[string]interface{}{
			"id":    set.ID,
			"rules": flattenRules(set.Rules),
		}
		flattenedSets = append(flattenedSets, flattenedSet)
	}
	return flattenedSets
}

func flattenRules(rules []*pagerduty.EventOrchestrationPathRule) []interface{} {
	var flattenedRules []interface{}

	for _, rule := range rules {
		flattenedRule := map[string]interface{}{
			"id":         rule.ID,
			"label":      rule.Label,
			"disabled":   rule.Disabled,
			"conditions": flattenRouterConditions(rule.Conditions),
			"actions":    flattenRouterActions(rule.Actions),
		}
		flattenedRules = append(flattenedRules, flattenedRule)
	}

	return flattenedRules
}

func flattenRouterActions(actions *pagerduty.EventOrchestrationPathRuleActions) []map[string]interface{} {
	var actionsMap []map[string]interface{}

	am := make(map[string]interface{})

	//TODO: test errors, what if the action is not passed? or if they route to a path that doesn't exist etc.
	am["route_to"] = actions.RouteTo
	actionsMap = append(actionsMap, am)

	return actionsMap
}

func flattenRouterConditions(conditions []*pagerduty.EventOrchestrationPathRuleCondition) []interface{} {
	var flattendConditions []interface{}

	for _, condition := range conditions {
		flattendCondition := map[string]interface{}{
			"expression": condition.Expression,
		}
		flattendConditions = append(flattendConditions, flattendCondition)
	}

	return flattendConditions
}
