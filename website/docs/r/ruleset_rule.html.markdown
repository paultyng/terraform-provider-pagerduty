---
layout: "pagerduty"
page_title: "PagerDuty: pagerduty_ruleset_rule"
sidebar_current: "docs-pagerduty-resource-ruleset-rule"
description: |-
  Creates and manages a ruleset rule in PagerDuty.
---

# pagerduty\_ruleset_rule

An [event rule](https://support.pagerduty.com/docs/rulesets#section-create-event-rules) allows you to set actions that should be taken on events that meet your designated rule criteria.

## Example Usage

```hcl
resource "pagerduty_team" "foo" {
  name = "Engineering (Seattle)"
}

resource "pagerduty_ruleset" "foo" {
  name = "Primary Ruleset"
  team {
    id = pagerduty_team.foo.id
  }
}

# The pagerduty_ruleset_rule.foo rule defined below
# repeats daily from 9:30am - 11:30am using the America/New_York timezone.
# Thus it requires a time_static instance to represent 9:30am on an arbitrary date in that timezone.
# April 11th, 2019 was EDT (UTC-4) https://www.timeanddate.com/worldclock/converter.html?iso=20190411T133000&p1=179
resource "time_static" "eastern_time_at_0930" {
  rfc3339 = "2019-04-11T09:30:00-04:00"
}

resource "pagerduty_ruleset_rule" "foo" {
  ruleset  = pagerduty_ruleset.foo.id
  position = 0
  disabled = "false"
  time_frame {
    scheduled_weekly {
      # Every Tuesday, Thursday, & Saturday
      weekdays = [2, 4, 6]
      # Starting at 9:30am
      start_time = time_static.eastern_time_at_0930.unix * 1000
      # Until 11:30am (2 hours later)
      duration = 2 * 60 * 60 * 1000
      # in this timezone
      # (either EST or EDT depending on when your event arrives)
      timezone = "America/New_York"
    }
  }
  conditions {
    operator = "and"
    subconditions {
      operator = "contains"
      parameter {
        value = "disk space"
        path  = "payload.summary"
      }
    }
    subconditions {
      operator = "contains"
      parameter {
        value = "db"
        path  = "payload.source"
      }
    }
  }
  variable {
    type = "regex"
    name = "Src"
    parameters {
      value = "(.*)"
      path  = "payload.source"
    }
  }
  actions {
    route {
      value = pagerduty_service.foo.id
    }
    severity {
      value = "warning"
    }
    annotate {
      value = "From Terraform"
    }
    extractions {
      target = "dedup_key"
      source = "details.host"
      regex  = "(.*)"
    }
    extractions {
      target   = "summary"
      template = "Warning: Disk Space Low on {{Src}}"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `ruleset` - (Required) The ID of the ruleset that the rule belongs to.
* `conditions` - (Required) Conditions evaluated to check if an event matches this event rule. Is always empty for the catch all rule, though.
* `position` - (Optional) Position/index of the rule within the ruleset.
* `disabled` - (Optional) Indicates whether the rule is disabled and would therefore not be evaluated.
* `time_frame` - (Optional) Settings for [scheduling the rule](https://support.pagerduty.com/docs/rulesets#section-scheduled-event-rules).
* `actions` - (Optional) Actions to apply to an event if the conditions match.
* `variable` - (Optional) Populate variables from event payloads and use those variables in other event actions. *NOTE: A rule can have multiple `variable` objects.*

### Conditions (`conditions`) supports the following:
* `operator` - Operator to combine sub-conditions. Can be `and` or `or`.
* `subconditions` - List of sub-conditions that define the the condition.

### Sub-Conditions (`subconditions`) supports the following:
* `operator` - Type of operator to apply to the sub-condition. Can be `exists`,`nexists`,`equals`,`nequals`,`contains`,`ncontains`,`matches`, or `nmatches`.
* `parameter` - Parameter for the sub-condition. It requires both a `path` and `value` to be set.

### Action (`actions`) supports the following:
* `route` (Optional) - The ID of the service where the event will be routed.
* `priority` (Optional) - The ID of the priority applied to the event.
* `severity` (Optional)  - The [severity level](https://support.pagerduty.com/docs/rulesets#section-set-severity-with-event-rules) of the event. Can be either `info`,`warning`,`error`, or `critical`.
* `annotate` (Optional) - Note added to the event.
* `extractions` (Optional) - Allows you to copy important data from one event field to another. Extraction objects may use *either* of the following field structures:
  * `source` - Field where the data is being copied from. Must be a [PagerDuty Common Event Format (PD-CEF)](https://support.pagerduty.com/docs/pd-cef) field.
  * `target` - Field where the data is being copied to. Must be a [PagerDuty Common Event Format (PD-CEF)](https://support.pagerduty.com/docs/pd-cef) field.
  * `regex` - The conditions that need to be met for the extraction to happen. Must use valid [RE2 regular expression syntax](https://github.com/google/re2/wiki/Syntax).

  *- **OR** -*

  * `template` - A customized field message. This can also include variables extracted from the payload by using string interpolation.
  * `target` - Field where the data is being copied to. Must be a [PagerDuty Common Event Format (PD-CEF)](https://support.pagerduty.com/docs/pd-cef) field.

  *NOTE: A rule can have multiple `extraction` objects attributed to it.*

* `suppress` (Optional) - Controls whether an alert is [suppressed](https://support.pagerduty.com/docs/rulesets#section-suppress-but-create-triggering-thresholds-with-event-rules) (does not create an incident). Note: If a threshold is set, the rule must also have a `route` action.
  * `value` - Boolean value that indicates if the alert should be suppressed before the indicated threshold values are met.
  * `threshold_value` (Optional) - The number of alerts that should be suppressed. Must be greater than 0.
  * `threshold_time_amount` (Optional) - The number value of the `threshold_time_unit` before an incident is created. Must be greater than 0.
  * `threshold_time_unit` (Optional)  - The `seconds`,`minutes`, or `hours` the `threshold_time_amount` should be measured.
* `event_action` (Optional) - An object with a single `value` field. The value sets whether the resulting alert status is `trigger` or `resolve`.
* `suspend` (Optional) - An object with a single `value` field. The value sets the length of time to suspend the resulting alert before triggering. Note: A rule with a `suspend` action must also have a `route` action.

### Time Frame (`time_frame`) supports the following:
* `scheduled_weekly` (Optional) - Values for executing the rule on a recurring schedule.
  * `weekdays` - An integer array representing which days during the week the rule executes. For example `weekdays = [1,3,7]` would execute on Monday, Wednesday and Sunday.
  * `timezone` - [The name of the timezone](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) for the given schedule, which will be used to determine UTC offset including adjustment for daylight saving time. For example: `timezone = "America/Toronto"`
  * `start_time` - A Unix timestamp in milliseconds which is combined with the `timezone` to determine the time this rule will start on each specified `weekday`. Note that the _date_ of the timestamp you specify does **not** matter, except that it lets you determine whether or not daylight saving time is in effect so that you use the correct UTC offset for the timezone you specify. In practice, you may want to use [the `time_static` resource](https://registry.terraform.io/providers/hashicorp/time/latest/docs/resources/static) to generate this value, as demonstrated in the `resource.pagerduty_ruleset_rule.foo` code example at the top of this page. To generate this timestamp manually, if you want your rule to apply starting at 9:30am in the `America/New_York` timezone, use your programing language of choice to determine a Unix timestamp that represents 9:30am in that timezone, like [1554989400000](https://www.epochconverter.com/timezones?q=1554989400000&tz=America%2FNew_York).
  * `duration` - Length of time the schedule will be active in milliseconds. For example `duration = 2 * 60 * 60 * 1000` if you want your rule to apply for 2 hours, from the specified `start_time`.
* `active_between` (Optional) - Values for executing the rule during a specific time period.
  * `start_time` - Beginning of the scheduled time when the rule should execute.  Unix timestamp in milliseconds.
  * `end_time` - Ending of the scheduled time when the rule should execute.  Unix timestamp in milliseconds.

## Attributes Reference

The following attributes are exported:

  * `id` - The ID of the rule.
  * `catch_all` - Indicates whether the rule is the last rule of the ruleset that serves as a catch-all. It has limited functionality compared to other rules.

## Import

Ruleset rules can be imported using using the related `ruleset` ID and the `ruleset_rule` ID separated by a dot, e.g.

```
$ terraform import pagerduty_ruleset_rule.main a19cdca1-3d5e-4b52-bfea-8c8de04da243.19acac92-027a-4ea0-b06c-bbf516519601
```
