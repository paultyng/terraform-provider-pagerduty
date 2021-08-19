---
layout: "pagerduty"
page_title: "PagerDuty: pagerduty_service_dependency"
sidebar_current: "docs-pagerduty-resource-service-dependency"
description: |-
  Creates and manages a business service dependency in PagerDuty.
---

# pagerduty\_service\_dependency

A [service dependency](https://developer.pagerduty.com/api-reference/reference/REST/openapiv3.json/paths/~1service_dependencies~1associate/post) is a relationship between two services that this service uses, or that are used by this service, and are critical for successful operation.


## Example Usage

```hcl
resource "pagerduty_service_dependency" "foo" {
	dependency {
		dependent_service {
			id = pagerduty_business_service.foo.id
			type = "business_service"
		}
		supporting_service {
			id = pagerduty_service.foo.id
			type = "service"
		}
	}
}

resource "pagerduty_service_dependency" "bar" {
	dependency {
		dependent_service {
			id = pagerduty_business_service.foo.id
			type = "business_service"
		}
		supporting_service {
			id = pagerduty_service.two.id
			type = "service"
		}
	}
}
```

## Argument Reference

The following arguments are supported:

  * `dependency` - (Required) The relationship between the `supporting_service` and `dependent_service`. One and only one dependency block must be defined.
  * `supporting_service` - (Required) The service that supports the dependent service.
  * `dependent_service` - (Required) The service that dependents on the supporting service.

## Attributes Reference

The following attributes are exported:

  * `id` - The ID of the service dependency.

***NOTE: Due to the API supporting this resource, it does not support updating. To make changes to a `service_dependency` you'll need to destroy and then create a new one.***

## Import

Service dependencies can be imported using the related supporting service id, supporting service type (`business_service` or `service`) and the dependency id separated by a dot, e.g.

```
$ terraform import pagerduty_service_dependency.main P4B2Z7G.business_service.D5RTHKRNGU4PYE90PJ
```
