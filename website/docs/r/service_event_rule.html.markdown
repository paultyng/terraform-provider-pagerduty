---
layout: "pagerduty"
page_title: "PagerDuty: pagerduty_service_event_rule"
sidebar_current: "docs-pagerduty-resource-service-event-rule"
description: |-
  Creates and manages a service event rule in PagerDuty.
---

# pagerduty\_service_event_rule

A [service event rule](https://support.pagerduty.com/docs/rulesets#service-event-rules) allows you to set actions that should be taken on events for a service that meet the designated rule criteria.

## Example Usage

```hcl
resource "pagerduty_service" "example" {
	name                    = "Checkout API Service"
	auto_resolve_timeout    = 14400
	acknowledgement_timeout = 600
	escalation_policy       = pagerduty_escalation_policy.example.id
	alert_creation          = "create_alerts_and_incidents"
}

resource "pagerduty_service_event_rule" "foo" {
	service = pagerduty_service.example.id
	position = 0
	disabled = true
	conditions {
		operator = "and"
		subconditions {
			operator = "contains"
			parameter {
				value = "disk space"
				path = "summary"
			}
		}
	}
	variable {
		type = "regex"
		name = "Src"
		parameters {
			value = "(.*)"
			path = "source"
		}
	}
	actions {
		annotate {
			value = "From Terraform"
		}
		extractions {
			target = "dedup_key"
			source = "source"
			regex = "(.*)"
		}
		extractions {
			target = "summary"
			template = "Warning: Disk Space Low on {{Src}}"
		}
	}
}

resource "pagerduty_service_event_rule" "bar" {
	service = pagerduty_service.foo.id
	position = 1
	disabled = true
	conditions {
		operator = "and"
		subconditions {
			operator = "contains"
			parameter {
				value = "cpu spike"
				path = "summary"
			}
		}
	}
	actions {
		annotate {
			value = "From Terraform"
		}
	}
}
```

## Argument Reference

The following arguments are supported:

* `service` - (Required) The ID of the service that the rule belongs to.
* `conditions` - (Required) Conditions evaluated to check if an event matches this event rule.
* `position` - (Optional) Position/index of the rule within the service.
* `disabled` - (Optional) Indicates whether the rule is disabled and would therefore not be evaluated.
* `time_frame` - (Optional) Settings for [scheduling the rule](https://support.pagerduty.com/docs/rulesets#section-scheduled-event-rules). 
* `actions` - (Optional) Actions to apply to an event if the conditions match.
* `variable` - (Optional) Populate variables from event payloads and use those variables in other event actions. *NOTE: A rule can have multiple `variable` objects.*

### Conditions (`conditions`) supports the following:
* `operator` - Operator to combine sub-conditions. Can be `and` or `or`.
* `subconditions` - List of sub-conditions that define the the condition. 

### Sub-Conditions (`subconditions`) supports the following:
* `operator` - Type of operator to apply to the sub-condition. Can be `exists`,`nexists`,`equals`,`nequals`,`contains`,`ncontains`,`matches`, or `nmatches`.
* `parameter` - Parameter for the sub-condition. It requires both a `path` and `value` to be set. The `path` value must be a [PagerDuty Common Event Format (PD-CEF)](https://support.pagerduty.com/docs/pd-cef) field. 

### Action (`actions`) supports the following:
* `priority` (Optional) - The ID of the priority applied to the event.
* `severity` (Optional)  - The [severity level](https://support.pagerduty.com/docs/rulesets#section-set-severity-with-event-rules) of the event. Can be either `info`,`error`,`warning`, or `critical`.
* `annotate` (Optional) - Note added to the event.
* `extractions` (Optional) - Allows you to copy important data from one event field to another. Extraction objects may use *either* of the following field structures:
	* `source` - Field where the data is being copied from. Must be a [PagerDuty Common Event Format (PD-CEF)](https://support.pagerduty.com/docs/pd-cef) field.
	* `target` - Field where the data is being copied to. Must be a [PagerDuty Common Event Format (PD-CEF)](https://support.pagerduty.com/docs/pd-cef) field.
	* `regex` - The conditions that need to be met for the extraction to happen. Must use valid [RE2 regular expression syntax](https://github.com/google/re2/wiki/Syntax).

	*- **OR** -*  
	
	* `template` - A customized field message. This can also include variables extracted from the payload by using string interpolation.
	* `target` - Field where the data is being copied to. Must be a [PagerDuty Common Event Format (PD-CEF)](https://support.pagerduty.com/docs/pd-cef) field.

	*NOTE: A rule can have multiple `extraction` objects attributed to it.*

* `suppress` (Optional) - Controls whether an alert is [suppressed](https://support.pagerduty.com/docs/rulesets#section-suppress-but-create-triggering-thresholds-with-event-rules) (does not create an incident).
	* `value` - Boolean value that indicates if the alert should be suppressed before the indicated threshold values are met.
	* `threshold_value` - The number of alerts that should be suppressed.
	* `threshold_time_amount` - The number value of the `threshold_time_unit` before an incident is created.
	* `threshold_time_unit` - The `seconds`,`minutes`, or `hours` the `threshold_time_amount` should be measured. 
* `event_action` (Optional) - An object with a single `value` field. The value sets whether the resulting alert status is `trigger` or `resolve`.
* `suspend` (Optional) - An object with a single `value` field. The value sets the length of time to suspend the resulting alert before triggering.

### Variable ('variable') supports the following:
* `name` (Optional) - The name of the variable.
* `type` (Optional) - Type of operation to populate the variable. Usually `regex`.
* `parameters` (Optional) - The parameters for performing the operation to populate the variable.
	* `value` - The value for the operation. For example, an RE2 regular expression for regex-type variables.
	* `path` - Path to a field in an event, in dot-notation. For Event Rules on a Service, this will have to be a [PD-CEF field](https://support.pagerduty.com/docs/pd-cef).

### Time Frame (`time_frame`) supports the following:
* `scheduled_weekly` (Optional) - Values for executing the rule on a recurring schedule.
	* `weekdays` - An integer array representing which days during the week the rule executes. For example `weekdays = [1,3,7]` would execute on Monday, Wednesday and Sunday.
	* `timezone` - Timezone for the given schedule.
	* `start_time` - Time when the schedule will start. Timestamp in milliseconds relative to the beginning of the day. For example, if you have a rule with a `start_time` of `0` and a `duration` of `60,000` then that rule would be active from `00:00` to `00:01`. If the `start_time` was `3,600,000` the it would be active starting at `01:00`.
	* `duration` - Length of time the schedule will be active.  Duration in milliseconds.
* `active_between` (Optional) - Values for executing the rule during a specific time period.
	* `start_time` - Beginning of the scheduled time when the rule should execute.  Timestamp in milliseconds relative to the beginning of the day.
	* `end_time` - Ending of the scheduled time when the rule should execute.  Timestamp in milliseconds relative to the beginning of the day.

## Attributes Reference

The following attributes are exported:

  * `id` - The ID of the rule.

## Import

Service event rules can be imported using using the related `service` id and the `service_event_rule` id separated by a dot, e.g.

```
$ terraform import pagerduty_service_event_rule.main a19cdca1-3d5e-4b52-bfea-8c8de04da243.19acac92-027a-4ea0-b06c-bbf516519601
```
