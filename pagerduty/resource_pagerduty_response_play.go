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

func resourcePagerDutyResponsePlay() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyResponsePlayCreate,
		Read:   resourcePagerDutyResponsePlayRead,
		Update: resourcePagerDutyResponsePlayUpdate,
		Delete: resourcePagerDutyResponsePlayDelete,
		Importer: &schema.ResourceImporter{
			State: resourcePagerDutyResponsePlayImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "response_play",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"from": {
				Type:     schema.TypeString,
				Required: true,
			},
			"team": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"subscriber": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"subscribers_message": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"responder": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"num_loops": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"on_call_handoff_notifications": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"escalation_rule": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"escalation_delay_in_minutes": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"target": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"service": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"team": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"responders_message": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"runnability": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"conference_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"conference_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func buildResponsePlayStruct(d *schema.ResourceData) *pagerduty.ResponsePlay {
	responsePlay := &pagerduty.ResponsePlay{
		Name:      d.Get("name").(string),
		FromEmail: d.Get("from").(string),
	}
	if attr, ok := d.GetOk("type"); ok {
		responsePlay.Type = attr.(string)
	}
	if attr, ok := d.GetOk("description"); ok {
		responsePlay.Description = attr.(string)
	}
	if attr, ok := d.GetOk("team"); ok {
		responsePlay.Team = &pagerduty.TeamReference{
			ID:   attr.(string),
			Type: "team",
		}
	}
	if attr, ok := d.GetOk("subscriber"); ok {
		responsePlay.Subscribers = expandSubscribers(attr.([]interface{}))
	}
	if attr, ok := d.GetOk("subscribers_message"); ok {
		responsePlay.SubscribersMessage = attr.(string)
	}

	if attr, ok := d.GetOk("responder"); ok {
		responsePlay.Responders = expandResponders(attr.([]interface{}))
	}

	if attr, ok := d.GetOk("responders_message"); ok {
		responsePlay.RespondersMessage = attr.(string)
	}

	if attr, ok := d.GetOk("runnability"); ok {
		responsePlay.Runnability = attr.(string)
	}

	if attr, ok := d.GetOk("conference_number"); ok {
		responsePlay.ConferenceNumber = attr.(string)
	}

	if attr, ok := d.GetOk("conference_url"); ok {
		responsePlay.ConferenceURL = attr.(string)
	}

	return responsePlay
}

func resourcePagerDutyResponsePlayCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	responsePlay := buildResponsePlayStruct(d)

	log.Printf("[INFO] Creating PagerDuty response play: %s", responsePlay.ID)

	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if responsePlay, _, err := client.ResponsePlays.Create(responsePlay); err != nil {
			return resource.RetryableError(err)
		} else if responsePlay != nil {
			d.SetId(responsePlay.ID)
			d.Set("from", responsePlay.FromEmail)
			log.Printf("[INFO] Created PagerDuty response play: %s (from: %s)", d.Id(), responsePlay.FromEmail)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}
	return resourcePagerDutyResponsePlayRead(d, meta)
}

func resourcePagerDutyResponsePlayRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	from := d.Get("from").(string)
	log.Printf("[INFO] Reading PagerDuty response play: %s (from: %s)", d.Id(), from)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		if responsePlay, _, err := client.ResponsePlays.Get(d.Id(), from); err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
		} else if responsePlay != nil {
			if responsePlay.Team != nil {
				d.Set("team", []interface{}{responsePlay.Team})
			}
			log.Printf("[INFO] Read PagerDuty response play initial subscribers: %s", d.Get("subscriber"))
			if err := d.Set("subscriber", flattenSubscribers(responsePlay.Subscribers)); err != nil {
				return resource.NonRetryableError(err)
			}
			log.Printf("[INFO] Read PagerDuty response play initial responders: %s", d.Get("responder"))
			if err := d.Set("responder", flattenResponders(responsePlay.Responders)); err != nil {
				return resource.NonRetryableError(err)
			}
			d.Set("from", from)
			d.Set("name", responsePlay.Name)
			d.Set("type", responsePlay.Type)
			d.Set("description", responsePlay.Description)
			d.Set("subscribers_message", responsePlay.SubscribersMessage)
			d.Set("responders_message", responsePlay.RespondersMessage)
			d.Set("runnability", responsePlay.Runnability)
			d.Set("conference_number", responsePlay.ConferenceNumber)
			d.Set("conference_url", responsePlay.ConferenceURL)

		}
		return nil
	})
}

func resourcePagerDutyResponsePlayUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	responsePlay := buildResponsePlayStruct(d)

	log.Printf("[INFO] Updating PagerDuty response play: %s", d.Id())

	retryErr := resource.Retry(30*time.Second, func() *resource.RetryError {
		if _, _, err := client.ResponsePlays.Update(d.Id(), responsePlay); err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}
	return resourcePagerDutyResponsePlayRead(d, meta)
}

func resourcePagerDutyResponsePlayDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty response play: %s", d.Id())
	from := d.Get("from").(string)

	retryErr := resource.Retry(30*time.Second, func() *resource.RetryError {
		if _, err := client.ResponsePlays.Delete(d.Id(), from); err != nil {
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

func expandSubscribers(v interface{}) []*pagerduty.SubscriberReference {
	var subscribers []*pagerduty.SubscriberReference

	for _, si := range v.([]interface{}) {
		sm := si.(map[string]interface{})
		sub := &pagerduty.SubscriberReference{
			ID:   sm["id"].(string),
			Type: sm["type"].(string),
		}
		subscribers = append(subscribers, sub)
	}

	return subscribers
}

func expandResponders(v interface{}) []*pagerduty.Responder {
	var responders []*pagerduty.Responder

	for _, ri := range v.([]interface{}) {
		rm := ri.(map[string]interface{})
		resp := &pagerduty.Responder{
			ID:                         rm["id"].(string),
			Type:                       rm["type"].(string),
			Description:                rm["description"].(string),
			Name:                       rm["name"].(string),
			NumLoops:                   rm["num_loops"].(int),
			OnCallHandoffNotifications: rm["on_call_handoff_notifications"].(string),
		}
		if rm["escalation_rules"] != nil {
			// calling expandEscalationRules in resource_pagerduty_escalation_policy
			resp.EscalationRules = expandEscalationRules(rm["escalation_rules"].([]interface{}))
		}
		if rm["service"] != nil {
			resp.Services = expandRSServices(rm["service"].([]interface{}))
		}
		if rm["teams"] != nil {
			// calling expandTeams in resource_pagerduty_escalation_policy
			resp.Teams = expandTeams(rm["teams"].([]interface{}))
		}
		log.Printf("[INFO] PagerDuty response play expandResponders: %v", resp.ID)
		responders = append(responders, resp)
	}

	return responders
}

func expandRSServices(v interface{}) []*pagerduty.ServiceReference {
	var services []*pagerduty.ServiceReference

	for _, si := range v.([]interface{}) {
		sm := si.(map[string]interface{})
		sr := &pagerduty.ServiceReference{
			ID:   sm["id"].(string),
			Type: sm["type"].(string),
		}
		services = append(services, sr)
	}

	return services
}

func flattenSubscribers(sref []*pagerduty.SubscriberReference) []interface{} {
	var subs []interface{}

	for _, s := range sref {
		flattenedSub := map[string]interface{}{
			"id":   s.ID,
			"type": s.Type,
		}
		subs = append(subs, flattenedSub)
	}
	return subs
}

func flattenResponders(rlist []*pagerduty.Responder) []interface{} {
	var resps []interface{}

	for _, r := range rlist {
		flattenedR := map[string]interface{}{
			"id":                            r.ID,
			"type":                          r.Type,
			"name":                          r.Name,
			"num_loops":                     r.NumLoops,
			"description":                   r.Description,
			"on_call_handoff_notifications": r.OnCallHandoffNotifications,
		}
		// EscalationRules
		if r.EscalationRules != nil {
			// flattenEscalationRules in resource_pagerduty_escalation_policy
			flattenedR["escalation_rules"] = flattenEscalationRules(r.EscalationRules)
		}
		// Services
		if r.Services != nil {
			flattenedR["services"] = flattenRSServices(r.Services)
		}
		// Teams
		if r.Teams != nil {
			flattenedR["teams"] = flattenRSTeams(r.Teams)
		}
		log.Printf("[INFO] PagerDuty response play flattenedR: %s", flattenedR)
		resps = append(resps, flattenedR)
	}

	return resps
}

func flattenRSServices(services []*pagerduty.ServiceReference) []interface{} {
	var flatServiceList []interface{}

	for _, s := range services {
		flatService := map[string]interface{}{
			"id":   s.ID,
			"type": s.Type,
		}
		flatServiceList = append(flatServiceList, flatService)
	}
	return flatServiceList
}

func flattenRSTeams(teams []*pagerduty.TeamReference) []interface{} {
	var flatTeamList []interface{}

	for _, t := range teams {
		flatTeam := map[string]interface{}{
			"id":   t.ID,
			"type": t.Type,
		}
		flatTeamList = append(flatTeamList, flatTeam)
	}
	return flatTeamList
}

func resourcePagerDutyResponsePlayImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*pagerduty.Client)

	ids := strings.SplitN(d.Id(), ".", 2)

	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing pagerduty_response_play. Expecting an importation ID formed as '<response_play_id>.<from_email>'")
	}
	rid, from := ids[0], ids[1]
	log.Printf("[INFO] Importing PagerDuty response play: %s (From: %s)", rid, from)

	_, _, err := client.ResponsePlays.Get(rid, from)
	if err != nil {
		return []*schema.ResourceData{}, err
	}
	// These are set because an import also calls Read behind the scenes
	d.SetId(rid)
	d.Set("from", from)

	return []*schema.ResourceData{d}, nil
}
