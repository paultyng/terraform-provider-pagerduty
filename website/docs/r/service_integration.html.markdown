---
layout: "pagerduty"
page_title: "PagerDuty: pagerduty_service_integration"
sidebar_current: "docs-pagerduty-resource-service-integration"
description: |-
  Creates and manages a service integration in PagerDuty.
---

# pagerduty\_service_integration

A [service integration](https://v2.developer.pagerduty.com/v2/page/api-reference#!/Services/post_services_id_integrations) is an integration that belongs to a service.

## Example Usage

```hcl
resource "pagerduty_user" "example" {
  name  = "Earline Greenholt"
  email = "125.greenholt.earline@graham.name"
  teams = [pagerduty_team.example.id]
}

resource "pagerduty_escalation_policy" "foo" {
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

resource "pagerduty_service_integration" "example" {
  name    = "Generic API Service Integration"
  type    = "generic_events_api_inbound_integration"
  service = pagerduty_service.example.id
}

resource "pagerduty_service_integration" "apiv2" {
  name = "API V2"
  type = "events_api_v2_inbound_integration"
  integration_key = "12345678910testtesttesttesttes"
  service = pagerduty_service.example.id
}

resource "pagerduty_service_integration" "email_x" {
  name = "Email X"
  type = "generic_email_inbound_integration"
  integration_email = "ecommerce@subdomain.pagerduty.com"
  service = pagerduty_service.example.id
}

data "pagerduty_vendor" "datadog" {
  name = "Datadog"
}

resource "pagerduty_service_integration" "datadog" {
  name    = data.pagerduty_vendor.datadog.name
  service = pagerduty_service.example.id
  vendor  = data.pagerduty_vendor.datadog.id
}

data "pagerduty_vendor" "cloudwatch" {
  name = "Cloudwatch"
}

resource "pagerduty_service_integration" "cloudwatch" {
  name    = data.pagerduty_vendor.cloudwatch.name
  service = pagerduty_service.example.id
  vendor  = data.pagerduty_vendor.cloudwatch.id
}

data "pagerduty_vendor" "email" {
  name = "Email"
}

resource "pagerduty_service_integration" "email" {
  name    = "${data.pagerduty_vendor.email.name}"
  service = "${pagerduty_service.example.id}"
  vendor  = "${data.pagerduty_vendor.email.id}"
  
  integration_email       = "s1@your_account.pagerduty.com"
  email_incident_creation = "use_rules"
  email_filter_mode       = "and-rules-email"
  email_filter {
    body_mode        = "always"
    body_regex       = null
    from_email_mode  = "match"
    from_email_regex = "(@foo.com*)"
    subject_mode     = "match"
    subject_regex    = "(CRITICAL*)"
  }
  email_filter {
    body_mode        = "always"
    body_regex       = null
    from_email_mode  = "match"
    from_email_regex = "(@bar.com*)"
    subject_mode     = "match"
    subject_regex    = "(CRITICAL*)"
  }
  
  email_parser {
    action = "resolve"
    match_predicate {
      type = "any"
      predicate {
        matcher = "foo"
        part    = "subject"
        type    = "contains"
      }
      predicate {
        type = "not"
        predicate {
          matcher = "(bar*)"
          part    = "body"
          type    = "regex"
        }
      }
    }
    value_extractor {
      ends_before  = "end"
      part         = "subject"
      starts_after = "start"
      type         = "between"
      value_name   = "incident_key"
    }
    value_extractor {
      ends_before  = "end"
      part         = "subject"
      starts_after = "start"
      type         = "between"
      value_name   = "FieldName1"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

  * `service` - (Required) The ID of the service the integration should belong to.
  * `name` - (Optional) The name of the service integration.
  * `type` - (Optional) The service type. Can be:
  `aws_cloudwatch_inbound_integration`,
  `cloudkick_inbound_integration`,
  `event_transformer_api_inbound_integration`,
  `events_api_v2_inbound_integration` (requires service `alert_creation` to be `create_alerts_and_incidents`),
  `generic_email_inbound_integration`,
  `generic_events_api_inbound_integration`,
  `keynote_inbound_integration`,
  `nagios_inbound_integration`,
  `pingdom_inbound_integration`or `sql_monitor_inbound_integration`.

    **Note:** This is meant for **generic** service integrations.
    To integrate with a **vendor** (e.g Datadog or Amazon Cloudwatch) use the `vendor` field instead.

  * `vendor` - (Optional) The ID of the vendor the integration should integrate with (e.g Datadog or Amazon Cloudwatch).
  * `integration_key` - (Optional) This is the unique key used to route events to this integration when received via the PagerDuty Events API.
  * `integration_email` - (Optional) This is the unique fully-qualified email address used for routing emails to this integration for processing.
  
  * `email_incident_creation` - (Optional) This is the unique fully-qualified email address used for routing emails to this integration for processing.
  * `email_filter_mode` - (Optional) This is the unique fully-qualified email address used for routing emails to this integration for processing.
  * `email_parsing_fallback` - (Optional) Can be `open_new_incident` or `discard`.
    
  Email filters (`email_filter`) supports the following:
  
  * `body_mode` - (Required) Can be `always` or `match`.
  * `body_regex` - (Optional) Should be a valid regex or `null`
  * `from_email_mode` - (Required) Can be `always` or `match`.
  * `from_email_regex` - (Optional) Should be a valid regex or `null`
  * `subject_mode` - (Required) Can be `always` or `match`.
  * `subject_regex` - (Optional) Should be a valid regex or `null`
  
  Email parsers (`email_parser`) supports the following:
    
  * `action` - (Required) Can be `resolve` or `trigger`.
   
  Match predicate (`match_predicate`) supports the following:
      
  * `type` - (Required) Can be `any` or `all`.
  
  Predicates (`predicate`) supports the following:
      
  * `type` - (Required) Can be `contains`, `exactly`, `regex` or `not`. If type is `not` predicate should contain child predicate with all parameters.
  * `matcher` - (Optional) Predicate value or valid regex.
  * `part` - (Optional) Can be `subject`, `body` or `from_addresses`.
  
  Value extractors (`value_extractor`) supports the following:
      
  * `type` - (Required) Can be `between`, `entire` or `regex`.
  * `part` - (Required) Can be `subject` or `body`.
  * `value_name` - (Required) First value extractor should have name `incident_key` other value extractors should contain custom names.
  * `ends_before` - (Optional) 
  * `starts_after` - (Optional)
  * `regex` - (Optional) If `type` has value `regex` this value should contain valid regex.
 
  
  **Note:** You can use the `pagerduty_vendor` data source to locate the appropriate vendor ID.
## Attributes Reference

The following attributes are exported:

  * `id` - The ID of the service integration.
  * `integration_key` - This is the unique key used to route events to this integration when received via the PagerDuty Events API.
  * `integration_email` - This is the unique fully-qualified email address used for routing emails to this integration for processing.
  * `html_url` - URL at which the entity is uniquely displayed in the Web app

To configure an event, please use the `integration_key` in the following interpolation:

```hcl
https://events.pagerduty.com/integration/${pagerduty_service_integration.slack.integration_key}/enqueue
```

## Import

Services can be imported using their related `service` id and service integration `id` separated by a dot, e.g.

```
$ terraform import pagerduty_service_integration.main PLSSSSS.PLIIIII
```
