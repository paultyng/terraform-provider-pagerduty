---
layout: "pagerduty"
page_title: "PagerDuty: pagerduty_webhook_subscription"
sidebar_current: "docs-pagerduty-resource-webhook-subscription"
description: |-
  Creates and manages a webhook subscription in PagerDuty.
---

# pagerduty\_webhook\_subscription

A [webhook subscription](https://developer.pagerduty.com/docs/ZG9jOjExMDI5NTkw-v3-overview) allow you to receive HTTP callbacks when incidents are created, updated and deleted. These are also known as V3 Webhooks.

## Example Usage

```hcl
data "pagerduty_service" "example" {
  name = "My Service"
}

resource "pagerduty_webhook_subscription" "foo" {
  delivery_method {
    type = "http_delivery_method"
    url = "https://example.com/receive_a_pagerduty_webhook"
    custom_header {
        name = "X-Foo"
        value = "foo"
    }
    custom_header {
        name = "X-Bar"
        value = "bar"
    }
  }
  description = "%s"
  events = [
    "incident.acknowledged",
    "incident.annotated",
    "incident.delegated",
    "incident.escalated",
    "incident.priority_updated",
    "incident.reassigned",
    "incident.reopened",
    "incident.resolved",
    "incident.responder.added",
    "incident.responder.replied",
    "incident.status_update_published",
    "incident.triggered",
    "incident.unacknowledged"
  ]
  active = true
  filter {
    id = data.pagerduty_service.example.id
    type = "service_reference"
  }
  type = "webhook_subscription"
}

```

## Argument Reference

The following arguments are supported:

  * `type` - (Required) The type indicating the schema of the object. The provider sets this as `webhook_subscription`, which is currently the only acceptable value. 
  * `active` - (Required) Determines whether the subscription will produce webhook events.
  * `delivery_method` - (Required) The object describing where to send the webhooks.
  * `description` - (Optional) A short description of the webhook subscription
  * `events` - (Required) A set of outbound event types the webhook will receive. The follow event types are possible: 
    * `incident.acknowledged`
    * `incident.annotated`
    * `incident.delegated`
    * `incident.escalated`
    * `incident.priority_updated`
    * `incident.reassigned`
    * `incident.reopened`
    * `incident.resolved`
    * `incident.responder.added`
    * `incident.responder.replied`
    * `incident.status_update_published`
    * `incident.triggered`
    * `incident.unacknowledged`
  * `filter` - (Required) determines which events will match and produce a webhook. There are currently three types of filters that can be applied to webhook subscriptions: `service_reference`, `team_reference` and `account_reference`.

### Webhook delivery method (`delivery_method`) supports the following:

* `temporarily_disabled` - (Required) Whether this webhook subscription is temporarily disabled. Becomes true if the delivery method URL is repeatedly rejected by the server.
* `type` - (Required) Indicates the type of the delivery method. Allowed and default value: `http_delivery_method`.
* `url` - (Required) The destination URL for webhook delivery.
* `custom_header` - (Optional) The custom_header of a webhook subscription define any optional headers that will be passed along with the payload to the destination URL.

### Webhook filter (`filter`) supports the following:

* `id` - (Optional) The id of the object being used as the filter. This field is required for all filter types except account_reference.
* `type` - (Required) The type of object being used as the filter. Allowed values are `account_reference`, `service_reference`, and `team_reference`.

## Attributes Reference

The following attributes are exported:

  * `id` - The ID of the webhook subscription.
  * `type` - The type indicating the schema of the object.
  * `active` - Determines whether this subscription will produce webhook events.
  * `description` - A short description of the webhook subscription.

## Import

Webhook Subscriptions can be imported using the `id`, e.g.

```
$ terraform import pagerduty_webhook_subscription.main PUABCDL
```
