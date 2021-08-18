---
layout: "pagerduty"
page_title: "PagerDuty: pagerduty_extension"
sidebar_current: "docs-pagerduty-resource-extension"
description: |-
  Creates and manages a service extension in PagerDuty.
---

# pagerduty\_extension

An [extension](https://developer.pagerduty.com/api-reference/reference/REST/openapiv3.json/paths/~1extensions/post) can be associated with a service.

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


resource "pagerduty_extension" "slack"{
  name = "My Web App Extension"
  endpoint_url = "https://generic_webhook_url/XXXXXX/BBBBBB"
  extension_schema = data.pagerduty_extension_schema.webhook.id
  extension_objects    = [pagerduty_service.example.id]

  config = <<EOF
{
	"restrict": "any",
	"notify_types": {
			"resolve": false,
			"acknowledge": false,
			"assignments": false
	},
	"access_token": "XXX"
}
EOF

}
```

## Argument Reference

The following arguments are supported:

  * `name` - (Optional) The name of the service extension.
  * `endpoint_url` - (Required|Optional) The url of the extension.
  **Note:** The [endpoint URL is Optional API wise](https://api-reference.pagerduty.com/#!/Extensions/post_extensions) in most cases. But in some cases it is a _Required_ parameter. For example, `pagerduty_extension_schema` named `Generic V2 Webhook` doesn't accept `pagerduty_extension` with no `endpoint_url`, but one with named `Slack` accepts.
  * `extension_schema` - (Required) This is the schema for this extension.
  * `extension_objects` - (Required) This is the objects for which the extension applies (An array of service ids).
  * `config` - (Optional) The configuration of the service extension as string containing plain JSON-encoded data.
  * `summary`- A short-form, server-generated string that provides succinct, important information about an object suitable for primary labeling of an entity in a client. In many cases, this will be identical to `name`, though it is not intended to be an identifier.

    **Note:** You can use the `pagerduty_extension_schema` data source to locate the appropriate extension vendor ID.
## Attributes Reference

The following attributes are exported:

  * `id` - The ID of the extension.
  * `html_url` - URL at which the entity is uniquely displayed in the Web app

## Import

Extensions can be imported using the id.e.g.

```
$ terraform import pagerduty_extension.main PLBP09X
```
