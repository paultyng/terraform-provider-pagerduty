---
layout: "pagerduty"
page_title: "PagerDuty: pagerduty_event_orchestration_unrouted"
sidebar_current: "docs-pagerduty-resource-event-orchestration-unrouted"
description: |-
  Creates and manages an Unrouted Orchestration for a Global Event Orchestration in PagerDuty.
---

# pagerduty_event_orchestration_unrouted

An Unrouted Orchestration allows users to create a set of Event Rules that will be evaluated against all events that don't match any rules in the Orchestration's Router.

The Unrouted Orchestration evaluates events sent to it against each of its rules, beginning with the rules in the "start" set. When a matching rule is found, it can modify and enhance the event and can route the event to another set of rules within this Unrouted Orchestration for further processing.

## Example of configuring Unrouted Rules for an Orchestration

In this example of an Unrouted Orchestration, the rule matches only if the condition is matched.
Alerts created for events that do not match the rule will have severity level set to `info` as defined in `catch_all` block.

```hcl
resource "pagerduty_event_orchestration_unrouted" "unrouted" {
 event_orchestration = pagerduty_event_orchestration.my_monitor.id
  set {
    id = "start"
    rule {
      label = "Update the summary of un-matched Critical alerts so they're easier to spot"
      condition {
        expression = "event.severity matches 'critical'"
      }
      actions {
        severity = "critical"
        extraction {
          target = "event.summary"
          template = "[Critical Unrouted] {{event.summary}}"
        }
      }
    }
  }
  catch_all {
    actions {
      severity = "info"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `event_orchestration` - (Required) The Event Orchestration to which this Unrouted Orchestration belongs to.
* `set` - (Required) An Unrouted Orchestration must contain at least a "start" set, but can contain any number of additional sets that are routed to by other rules to form a directional graph.
* `catch_all` - (Required) the `catch_all` actions will be applied if an Event reaches the end of any set without matching any rules in that set.

### Set (`set`) supports the following:
* `id` - (Required) The ID of this set of rules. Rules in other sets can route events into this set using the rule's `route_to` property.
* `rule` - (Optional) The Unrouted Orchestration evaluates Events against these Rules, one at a time, and applies all the actions for first rule it finds where the event matches the rule's conditions. If no rules are provided as part of Terraform configuration, the API returns empty list of rules.

### Rule (`rule`) supports the following:
* `label` - (Optional) A description of this rule's purpose.
* `condition` - (Optional) Each of these conditions is evaluated to check if an event matches this rule. The rule is considered a match if any of these conditions match. If none are provided, the event will `always` match against the rule.
* `actions` - (Required) Actions that will be taken to change the resulting alert and incident, when an event matches this rule.
* `disabled` - (Optional) Indicates whether the rule is disabled and would therefore not be evaluated.

### Condition (`condition`) supports the following:
* `expression`- (Required) A [PCL condition](https://developer.pagerduty.com/docs/ZG9jOjM1NTE0MDc0-pcl-overview) string.

### Actions (`actions`) supports the following:
* `route_to` - (Optional) The ID of a Set from this Unrouted Orchestration whose rules you also want to use with events that match this rule.
* `severity` - (Optional) sets Severity of the resulting alert. Allowed values are: `info`, `error`, `warning`, `critical`
* `event_action` - (Optional) sets whether the resulting alert status is trigger or resolve. Allowed values are: `trigger`, `resolve`
* `variable` - (Optional) Populate variables from event payloads and use those variables in other event actions.
  * `name` - (Required) The name of the variable
  * `path` - (Required) Path to a field in an event, in dot-notation. This supports both [PD-CEF](https://support.pagerduty.com/docs/pd-cef) and non-CEF fields. Eg: Use `event.summary` for the `summary` CEF field. Use `raw_event.fieldname` to read from the original event `fieldname` data.
  * `type` - (Required) Only `regex` is supported
  * `value` - (Required) The Regex expression to match against. Must use valid [RE2 regular expression](https://github.com/google/re2/wiki/Syntax) syntax.
* `extraction` - (Optional) Replace any CEF field or Custom Details object field using custom variables.
  * `target` - (Required) The PagerDuty Common Event Format [PD-CEF](https://support.pagerduty.com/docs/pd-cef) field that will be set with the value from the `template` or based on `regex` and `source` fields.
  * `template` - (Optional) A string that will be used to populate the `target` field. You can reference variables or event data within your template using double curly braces. For example:
    * Use variables named `ip` and `subnet` with a template like: `{{variables.ip}}/{{variables.subnet}}`
    * Combine the event severity & summary with template like: `{{event.severity}}:{{event.summary}}`
  * `regex` - (Optional) A [RE2 regular expression](https://github.com/google/re2/wiki/Syntax) that will be matched against field specified via the `source` argument. If the regex contains one or more capture groups, their values will be extracted and appended together. If it contains no capture groups, the whole match is used. This field can be ignored for `template` based extractions.
  * `source` - (Optional) The path to the event field where the `regex` will be applied to extract a value. You can use any valid [PCL path](https://developer.pagerduty.com/docs/ZG9jOjM1NTE0MDc0-pcl-overview#paths) like `event.summary` and you can reference previously-defined variables using a path like `variables.hostname`. This field can be ignored for `template` based extractions.

### Catch All (`catch_all`) supports the following:
* `actions` - (Required) These are the actions that will be taken to change the resulting alert and incident. `catch_all` supports all actions described above for `rule` _except_ `route_to` action.

## Attributes Reference

The following attributes are exported:
* `self` - The URL at which the Unrouted Event Orchestration is accessible.
* `rule`
  * `id` - The ID of the rule within the set.

## Import

Unrouted Orchestration can be imported using the `id` of the Event Orchestration, e.g.

```
$ terraform import pagerduty_event_orchestration_unrouted.unrouted 1b49abe7-26db-4439-a715-c6d883acfb3e
```
