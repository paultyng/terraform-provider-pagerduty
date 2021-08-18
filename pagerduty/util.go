package pagerduty

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func timeToUTC(v string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return time.Time{}, err
	}

	return t.UTC(), nil
}

// validateRFC3339 validates that a date string has the correct RFC3339 layout
func validateRFC3339(v interface{}, k string) (we []string, errors []error) {
	value := v.(string)
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		errors = append(errors, fmt.Errorf("%s is not a valid format for argument: %s. Expected format: %s (RFC3339)", value, k, time.RFC3339))
	}

	return
}

func suppressRFC3339Diff(k, oldTime, newTime string, d *schema.ResourceData) bool {
	oldT, newT, err := parseRFC3339Time(k, oldTime, newTime)
	if err != nil {
		log.Printf(err.Error())
		return false
	}

	return oldT.Equal(newT)
}

// issue: https://github.com/PagerDuty/terraform-provider-pagerduty/issues/200
// The start value of schedule layer can't be set to a time in the past. So if the value passed in is before the current time then PagerDuty
// will set the start to the current time. Thus, we do not need to show diff if both newT and oldT is in the past, as it will not bring
// any real changes to the schedule layer.
func suppressScheduleLayerStartDiff(k, oldTime, newTime string, d *schema.ResourceData) bool {
	oldT, newT, err := parseRFC3339Time(k, oldTime, newTime)
	if err != nil {
		log.Printf(err.Error())
		return false
	}

	return oldT.Equal(newT) || (newT.Before(time.Now()) && oldT.Before(time.Now()))
}

func parseRFC3339Time(k, oldTime, newTime string) (time.Time, time.Time, error) {
	var t time.Time
	oldT, err := time.Parse(time.RFC3339, oldTime)
	if err != nil {
		return t, t, fmt.Errorf("[ERROR] Failed to parse %q (old %q). Expected format: %s (RFC3339)", oldTime, k, time.RFC3339)
	}

	newT, err := time.Parse(time.RFC3339, newTime)
	if err != nil {
		return t, t, fmt.Errorf("[ERROR] Failed to parse %q (new %q). Expected format: %s (RFC3339)", oldTime, k, time.RFC3339)
	}

	return oldT, newT, nil
}

func suppressLeadTrailSpaceDiff(k, old, new string, d *schema.ResourceData) bool {
	return old == strings.TrimSpace(new)
}

func suppressCaseDiff(k, old, new string, d *schema.ResourceData) bool {
	return old == strings.ToLower(new)
}

// Validate a value against a set of possible values
func validateValueFunc(values []string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (we []string, errors []error) {
		value := v.(string)
		valid := false
		for _, val := range values {
			if value == val {
				valid = true
				break
			}
		}

		if !valid {
			errors = append(errors, fmt.Errorf("%#v is an invalid value for argument %s. Must be one of %#v", value, k, values))
		}
		return
	}
}

// Takes the result of flatmap.Expand for an array of strings
// and returns a []string
func expandStringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		vs = append(vs, string(v.(string)))
	}
	return vs
}

func expandString(v string) []interface{} {
	var obj []interface{}
	if err := json.Unmarshal([]byte(v), &obj); err != nil {
		log.Printf("[ERROR] Could not unmarshal field %s: %v", v, err)
		return nil
	}

	return obj
}

func flattenSlice(v []interface{}) interface{} {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("[ERROR] Could not marshal field %s: %v", v, err)
		return nil
	}
	return string(b)
}
