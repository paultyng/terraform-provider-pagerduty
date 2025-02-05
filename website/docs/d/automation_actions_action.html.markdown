---
layout: "pagerduty"
page_title: "PagerDuty: pagerduty_automation_actions_action"
sidebar_current: "docs-pagerduty-datasource-automation-actions-action"
description: |-
  Get information about an Automation Actions action that you have created.
---

# pagerduty\_automation\_actions\_action

Use this data source to get information about a specific [automation actions action][1].

## Example Usage

```hcl
data "pagerduty_automation_actions_action" "example" {
  id = "01CS1685B2UDM4I3XUUOXPPORM"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) The id of the automation actions action in the PagerDuty API.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the action.
* `name` - The name of the action.
* `type` - The type of object. The value returned will be `action`.
* `action_type` - The type of the action. The only allowed values are `process_automation` and `script`.
* `creation_time` - The time action was created. Represented as an ISO 8601 timestamp.
* `action_data_reference` - Action Data block. Action Data is documented below.
* `description` - (Optional) The description of the action.
* `runner_id` - (Optional) The Process Automation Actions runner to associate the action with.
* `runner_type` - (Optional) The type of the runner associated with the action.
* `action_classification` - (Optional) The category of the action. The only allowed values are `diagnostic` and `remediation`.
* `modify_time` - (Optional) The last time action has been modified. Represented as an ISO 8601 timestamp.
* `only_invocable_on_unresolved_incidents` - (Optional) Whether or not the action can be invoked on unresolved incidents.
* `map_to_all_services` - (Optional) If the action should be able to be run against all services or just specified ones.

Action Data (`action_data_reference`) supports the following:

  * `process_automation_job_id` - (Required for `process_automation` action_type) The ID of the Process Automation job to execute.
  * `process_automation_job_arguments` - (Optional) The arguments to pass to the Process Automation job execution.
  * `process_automation_node_filter` - (Optional) The expression that filters on which nodes a Process Automation Job executes [Learn more](https://docs.rundeck.com/docs/manual/05-nodes.html#node-filtering).
  * `script` - (Required for `script` action_type) Body of the script to be executed on the Runner. Max length is 16777215 characters.
  * `invocation_command` - (Optional) The command to execute the script with.

[1]: https://developer.pagerduty.com/api-reference/357ed15419f64-get-an-automation-action
