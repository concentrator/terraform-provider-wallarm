---
layout: "wallarm"
page_title: "Wallarm: wallarm_rule_bruteforce_counter"
subcategory: "Rule"
description: |-
  Provides the "Define brute-force attacks counter" rule resource.
---

# wallarm_rule_bruteforce_counter

Provides the resource to manage rules with the "Define brute-force attacks counter" action type. For detecting brute-force attacks, with every request, one of the statistical counters is incremented. By default, the counter name is automatically defined based on the domain name and the request path.

In cases where part of the URI does not matter, you can override the counter name for all such requests.

## Example Usage

```hcl
# Sets the `b:root` counter on the root `/` path

resource "wallarm_rule_bruteforce_counter" "root_counter" {
	counter = "b:root"

	action {
		type = "iequal"
		value = "/"
		point = {
			path = 0
		}
	}

}
```

## Argument Reference

* `counter` - (Required) Name of the counter always starts with `b:`.
* `client_id` - (Optional) ID of the client to apply the rules to. The value is required for multi-tenant scenarios.
* `action` - (Optional) Rule conditions. Possible attributes are described below.

**action**

`action` argument shares the available
conditions which can be applied. The conditions are:

* `type` - (Optional) Condition type. Can be: `equal`, `iequal`, `regex`, `absent`. Must be omitted for the `instance` parameter in `point`.
  For more details, see the offical [Wallarm documentation](https://docs.wallarm.com/user-guides/rules/add-rule/#condition-types)
  Example:
  `type = "absent"`
* `value` - (Optional) Value of the parameter to match with. Must be omitted for the `instance` parameter in `point` or if `type` is `absent`.
  Example:
  `value = "example.com"`
* `point` - (Optional) Request parameters that trigger the rule. Possible values are described below. For more details, see the official [Wallarm documentatioon](https://docs.wallarm.com/user-guides/rules/request-processing/#identifying-and-parsing-the-request-parts).

### Nested Objects

**point**

  * `header` - (Optional) Arbitrary HEADER parameter name.
  Example:
  `header = "HOST"`
  * `method` - (Optional) Request method. Can be: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`, `PATCH`.
  Example:
  `method = "POST"`
  * `path` - (Optional) Array with URL parts separated by the `/` symbol (the last URL part is not included in the array). If there is only one part in the URL, the array will be empty.
  Example:
  `path = 0`
  * `action_name` - (Optional) The last part of the URL after the `/` symbol and before the first period (`.`). This part of the URL is always present in the request even if its value is an empty string.
  Example:
  `action_name = "login"`
  * `action_ext` - (Optional) The part of the URL after the last period (`.`). It may be missing in the request.
  Example:
  `action_ext = "php"`
  * `proto` - (Optional) Version of the HTTP Protocol.
  Example:
  `proto = "1.1"`
  * `scheme` - (Optional) `http`/`https`.
  Example:
  `scheme = "https"` 
  * `uri` - (Optional) Part of the request URL without domain.
  Example:
  `uri = "/api/login"` 
  * `instance` - (Optional) ID of the application.
  Example:
  `instance = 42`

Example:

  ```hcl
  # ... omitted

  action {
    type = "equal"
    point = {
      scheme = "https"
    }
  }

  action {
    point = {
      instance = 9
    }
  }
  
  action {
    type = "absent"
    point = {
      path = 0
     }
  }

  action {
    type = "regex"
    point = {
      action_name = "masking"
    }
  }

  action {
    type = "absent"
    point = {
      action_ext = ""
    }
  }

  action {
    type = "iequal"
    value = "example.com"
    point = {
      header = "HOST"
    }
  }

  # ... omitted
  ```

> **_NOTE:_**
See below what limitations apply

When `type` is `absent`
`point` must contain key with the default value. For `action_name`, `action_ext`, `method`, `proto`, `scheme`, `uri` default value is `""` (empty string)

## Attributes Reference

* `rule_id` - ID of the created rule.
* `action_id` - The action ID (The conditions to apply on request).
* `rule_type` - Type of the created rule. For example, `rule_type = "brute_counter"`.

## Import

The rule can be imported using a composite ID formed of client ID, action ID, rule ID and rule type.

```
$ terraform import wallarm_rule_bruteforce_counter.root_counter 6039/563854/11086884
```

* `6039` - Client ID.
* `563854` - Action ID.
* `11086884` - Rule ID.