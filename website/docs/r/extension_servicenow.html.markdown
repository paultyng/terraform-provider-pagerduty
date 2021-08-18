---
layout: "pagerduty"
page_title: "PagerDuty: pagerduty_extension_servicenow"
sidebar_current: "docs-pagerduty-resource-extension-servicenow"
description: |-
  Creates and manages a ServiceNow service extension in PagerDuty.
---

# pagerduty\_extension\_servicenow

A special case for [extension](https://developer.pagerduty.com/api-reference/reference/REST/openapiv3.json/paths/~1extensions/post) for ServiceNow.

## Example Usage

```hcl
data "pagerduty_extension_schema" "webhook" {
  name = "Generic V2 Webhook"
}

resource "pagerduty_user" "example" {
  name  = "Howard James"
  email = "howard.james@example.domain"
}

resource "pagerduty_escalation_policy" "example" {
  name      = "Engineering Escalation Policy"
  num_loops = 2

  rule {
    escalation_delay_in_minutes = 10

    target {
      type = "user"
      id   = pagerduty_user.example.id
    }
  }
}

resource "pagerduty_service" "example" {
  name                    = "My Web App"
  auto_resolve_timeout    = 14400
  acknowledgement_timeout = 600
  escalation_policy       = pagerduty_escalation_policy.example.id
}


resource "pagerduty_extension_servicenow" "snow"{
  name = "My Web App Extension"
  extension_schema = data.pagerduty_extension_schema.webhook.id
  extension_objects = [pagerduty_service.example.id]
  snow_user = "meeps"
  snow_password = "zorz"
  sync_options = "manual_sync"
  target = "https://foo.servicenow.com/webhook_foo"
  task_type = "incident"
  referer = "None"
}
```

## Argument Reference

The following arguments are supported:

  * `name` - (Optional) The name of the service extension.
  * `extension_schema` - (Required) This is the schema for this extension.
  * `extension_objects` - (Required) This is the objects for which the extension applies (An array of service ids).
  * `snow_user` - (Required) The ServiceNow username.
  * `snow_password` - (Required) The ServiceNow password.
  * `summary`- A short-form, server-generated string that provides succinct, important information about an object suitable for primary labeling of an entity in a client. In many cases, this will be identical to `name`, though it is not intended to be an identifier.
  * `sync_options` - (Required) The ServiceNow sync option.
  * `target` - (Required) Target Webhook URL
  * `task_type` - (Required) The ServiceNow task type, typically `incident`.
  * `referer` - (Required) The ServiceNow referer.

## Attributes Reference

The following attributes are exported:

  * `id` - The ID of the extension.
  * `html_url` - URL at which the entity is uniquely displayed in the Web app

## Import

Extensions can be imported using the id.e.g.

```
$ terraform import pagerduty_extension_servicenow.main PLBP09X
```
