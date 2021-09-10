package pagerduty

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func resourcePagerDutySchedule() *schema.Resource {
	return &schema.Resource{
		Create: resourcePagerDutyScheduleCreate,
		Read:   resourcePagerDutyScheduleRead,
		Update: resourcePagerDutyScheduleUpdate,
		Delete: resourcePagerDutyScheduleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"time_zone": {
				Type:     schema.TypeString,
				Required: true,
			},

			"overflow": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},

			"layer": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},

						"start": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateFunc:     validateRFC3339,
							DiffSuppressFunc: suppressScheduleLayerStartDiff,
						},

						"end": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateFunc:     validateRFC3339,
							DiffSuppressFunc: suppressRFC3339Diff,
						},

						"rotation_virtual_start": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateFunc:     validateRFC3339,
							DiffSuppressFunc: suppressRFC3339Diff,
						},

						"rotation_turn_length_seconds": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"users": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},

						"restriction": {
							Optional: true,
							Type:     schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},

									"start_time_of_day": {
										Type:     schema.TypeString,
										Required: true,
									},

									"start_day_of_week": {
										Type:     schema.TypeInt,
										Optional: true,
										ValidateFunc: validation.IntBetween(1, 7),
									},

									"duration_seconds": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"teams": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func buildScheduleStruct(d *schema.ResourceData) (*pagerduty.Schedule, error) {
	layers, err := expandScheduleLayers(d.Get("layer"))
	if err != nil {
		return nil, err
	}

	schedule := &pagerduty.Schedule{
		Name:           d.Get("name").(string),
		TimeZone:       d.Get("time_zone").(string),
		ScheduleLayers: layers,
	}

	if attr, ok := d.GetOk("description"); ok {
		schedule.Description = attr.(string)
	}

	if attr, ok := d.GetOk("teams"); ok {
		schedule.Teams = expandSchedTeams(attr.([]interface{}))
	}

	return schedule, nil
}

func resourcePagerDutyScheduleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	schedule, err := buildScheduleStruct(d)
	if err != nil {
		return err
	}

	o := &pagerduty.CreateScheduleOptions{}

	if v, ok := d.GetOk("overflow"); ok {
		o.Overflow = v.(bool)
	}

	log.Printf("[INFO] Creating PagerDuty schedule: %s", schedule.Name)

	schedule, _, err = client.Schedules.Create(schedule, o)
	if err != nil {
		return err
	}

	d.SetId(schedule.ID)

	return resourcePagerDutyScheduleRead(d, meta)
}

func resourcePagerDutyScheduleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Reading PagerDuty schedule: %s", d.Id())

	retryErr := resource.Retry(30*time.Second, func() *resource.RetryError {
		if schedule, _, err := client.Schedules.Get(d.Id(), &pagerduty.GetScheduleOptions{}); err != nil {
			time.Sleep(2 * time.Second)
			return resource.RetryableError(err)
		} else if schedule != nil {
			d.Set("name", schedule.Name)
			d.Set("time_zone", schedule.TimeZone)
			d.Set("description", schedule.Description)

			layers, err := flattenScheduleLayers(schedule.ScheduleLayers)
			if err != nil {
				return resource.NonRetryableError(err)
			}

			if err := d.Set("layer", layers); err != nil {
				return resource.NonRetryableError(err)
			}
			if err := d.Set("teams", flattenShedTeams(schedule.Teams)); err != nil {
				return resource.NonRetryableError(fmt.Errorf("error setting teams: %s", err))
			}

		}
		return nil
	})

	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}

	return nil
}

func resourcePagerDutyScheduleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	schedule, err := buildScheduleStruct(d)
	if err != nil {
		return err
	}

	opts := &pagerduty.UpdateScheduleOptions{}

	if v, ok := d.GetOk("overflow"); ok {
		opts.Overflow = v.(bool)
	}

	// A schedule layer can never be removed but it can be ended.
	// Here we determine which layer has been removed from the configuration
	// and we mark it as ended. This is to avoid diff issues.

	if d.HasChange("layer") {
		oraw, nraw := d.GetChange("layer")

		osl, err := expandScheduleLayers(oraw.([]interface{}))
		if err != nil {
			return err
		}

		nsl, err := expandScheduleLayers(nraw.([]interface{}))
		if err != nil {
			return err
		}

		// Checks to see if new schedule layers (nsl) include all old schedule layers (osl)
		for _, o := range osl {
			found := false
			for _, n := range nsl {
				// layer is found in both nsl and osl
				if o.ID == n.ID {
					found = true
				}
			}

			// If layer is not found in new schedule layers (nsl) set end value for layer
			if !found {
				end, err := timeToUTC(time.Now().Format(time.RFC3339))
				if err != nil {
					return err
				}
				o.End = end.String()
				schedule.ScheduleLayers = append(schedule.ScheduleLayers, o)
			}
		}
	}

	log.Printf("[INFO] Updating PagerDuty schedule: %s", d.Id())

	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if _, _, err := client.Schedules.Update(d.Id(), schedule, opts); err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})
	if retryErr != nil {
		time.Sleep(2 * time.Second)
		return retryErr
	}

	return nil
}

func resourcePagerDutyScheduleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pagerduty.Client)

	log.Printf("[INFO] Deleting PagerDuty schedule: %s", d.Id())

	// Retrying to give other resources (such as escalation policies) to delete
	retryErr := resource.Retry(2*time.Minute, func() *resource.RetryError {
		if _, err := client.Schedules.Delete(d.Id()); err != nil {
			if isErrCode(err, 400) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
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

func expandScheduleLayers(v interface{}) ([]*pagerduty.ScheduleLayer, error) {
	var scheduleLayers []*pagerduty.ScheduleLayer

	for _, sl := range v.([]interface{}) {
		rsl := sl.(map[string]interface{})

		// This is a temporary fix to prevent getting back the wrong rotation_virtual_start time.
		// The background here is that if a user specifies a rotation_virtual_start time to be:
		// "2017-09-01T10:00:00+02:00" the API returns back "2017-09-01T12:00:00+02:00".
		// With this fix in place, we get the correct rotation_virtual_start time, thus
		// eliminating the diff issues we've been seeing in the past.
		// This has been confirmed working by PagerDuty support.
		rvs, err := timeToUTC(rsl["rotation_virtual_start"].(string))
		if err != nil {
			return nil, err
		}

		scheduleLayer := &pagerduty.ScheduleLayer{
			ID:                        rsl["id"].(string),
			Name:                      rsl["name"].(string),
			Start:                     rsl["start"].(string),
			End:                       rsl["end"].(string),
			RotationVirtualStart:      rvs.String(),
			RotationTurnLengthSeconds: rsl["rotation_turn_length_seconds"].(int),
		}

		for _, slu := range rsl["users"].([]interface{}) {
			user := &pagerduty.UserReferenceWrapper{
				User: &pagerduty.UserReference{
					ID:   slu.(string),
					Type: "user",
				},
			}
			scheduleLayer.Users = append(scheduleLayer.Users, user)
		}

		for _, slr := range rsl["restriction"].([]interface{}) {
			rslr := slr.(map[string]interface{})

			restriction := &pagerduty.Restriction{
				Type:            rslr["type"].(string),
				StartTimeOfDay:  rslr["start_time_of_day"].(string),
				StartDayOfWeek:  rslr["start_day_of_week"].(int),
				DurationSeconds: rslr["duration_seconds"].(int),
			}

			scheduleLayer.Restrictions = append(scheduleLayer.Restrictions, restriction)
		}

		scheduleLayers = append(scheduleLayers, scheduleLayer)
	}

	return scheduleLayers, nil
}

func flattenScheduleLayers(v []*pagerduty.ScheduleLayer) ([]map[string]interface{}, error) {
	var scheduleLayers []map[string]interface{}

	for _, sl := range v {
		// A schedule layer can never be removed but it can be ended.
		// Here we check each layer and if it has been ended we don't read it back
		// because it's not relevant anymore.
		if sl.End != "" {
			end, err := timeToUTC(sl.End)
			if err != nil {
				return nil, err
			}

			if time.Now().UTC().After(end) {
				continue
			}
		}
		scheduleLayer := map[string]interface{}{
			"id":                           sl.ID,
			"name":                         sl.Name,
			"end":                          sl.End,
			"start":                        sl.Start,
			"rotation_virtual_start":       sl.RotationVirtualStart,
			"rotation_turn_length_seconds": sl.RotationTurnLengthSeconds,
		}

		var users []string

		for _, slu := range sl.Users {
			users = append(users, slu.User.ID)
		}

		scheduleLayer["users"] = users

		var restrictions []map[string]interface{}

		for _, slr := range sl.Restrictions {
			restriction := map[string]interface{}{
				"duration_seconds":  slr.DurationSeconds,
				"start_time_of_day": slr.StartTimeOfDay,
				"type":              slr.Type,
			}

			if slr.StartDayOfWeek > 0 {
				restriction["start_day_of_week"] = slr.StartDayOfWeek
			}

			restrictions = append(restrictions, restriction)
		}

		scheduleLayer["restriction"] = restrictions

		scheduleLayers = append(scheduleLayers, scheduleLayer)
	}

	// Reverse the final result and return it
	resultReversed := make([]map[string]interface{}, 0, len(scheduleLayers))

	for i := len(scheduleLayers) - 1; i >= 0; i-- {
		resultReversed = append(resultReversed, scheduleLayers[i])
	}

	return resultReversed, nil
}

// the expandShedTeams and flattenSchedTeams are based on the expandTeams and flattenTeams functions in the user
// resource. added these functions here for maintainability
func expandSchedTeams(v interface{}) []*pagerduty.TeamReference {
	var teams []*pagerduty.TeamReference

	for _, t := range v.([]interface{}) {
		team := &pagerduty.TeamReference{
			ID:   t.(string),
			Type: "team_reference",
		}
		teams = append(teams, team)
	}

	return teams
}

func flattenShedTeams(teams []*pagerduty.TeamReference) []string {
	res := make([]string, len(teams))
	for i, t := range teams {
		res[i] = t.ID
	}

	return res
}
