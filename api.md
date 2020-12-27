# Loaner API

The loaner API provides services related to loans, like
creating loan plans.

## Core Concepts

This API strives to follow the [REST](http://en.wikipedia.org/wiki/Representational_State_Transfer)
architectural style. It has resource-oriented
URLs and uses [JSON](https://www.json.org/) as the representation for resources.
It uses standard HTTP response codes and verbs.

JSON request/responses are documented by listing the fields and
the expected type for that field (along with an optional annotation
if the field is expected to be optional).

For example, with this specification:

```
{
    "field" : <boolean>,
    "optionalField" : <string>(optional)
}
```

You can expect a JSON like this:

```json
{
    "field": true,
    "optionalField" : "data"
}
```

All JSON fields documented as part of request/response bodies are
to be considered obligatory, unless they are explicitly
documented as optional. Also all fields of type **<string>** are
expected to be non-empty by default, unless stated otherwise.

Two domain specific types (not native on JSON) are defined here,
"<decimal>" and "<date>". When a field has type "<decimal>" you
can expect an string representing decimal values,
like "5000.10" and "5.0".

When a field has type "<date>" you can expect an string representation
of date following the [RFC 3339](https://tools.ietf.org/html/rfc3339),
for example: "2018-01-01T00:00:01Z".


# Error Handling

When an error occurs you can always expect an HTTP status code indicating the
nature of the failure and also a response body with an error message
giving some more information on what went wrong (when appropriate).

It follows this schema:

```
{
    "error": {
        "message" : <string>
    }
}
```

The **message** is intended for human inspection, no programmatic decision
should be made using their contents. Services integrating with this API
can depend on the error response schema, but the contents of the
message itself should be handled as opaque strings.


## Creating a loan plan

To create a loan plan, send the following request:

```
POST /loan-plan
```

With the following request body:

```json
{
    "loanAmount": <decimal>,
    "nominalRate": <decimal>,
    "duration": <int>,
    "startDate": <date>
}
```

Example of request body:

```json
{
    "loanAmount": "5000",
    "nominalRate": "5.0",
    "duration": 24,
    "startDate": "2018-01-01T00:00:01Z"
}
```

In case of success you can expect an status code 200/OK and the following response:

```json
{
    "borrowerPayments": [
        {
            "borrowerPaymentAmount": <decimal>,
            "date": <date>,
            "initialOutstandingPrincipal": <decimal>,
            "interest": <decimal>,
            "principal": <decimal>,
            "remainingOutstandingPrincipal": <decimal>
        }
    ]
}
```

Example of response body:

```json
{
    "borrowerPayments":[
        {
            "borrowerPaymentAmount":"219.36",
            "date":"2018-01-01T00:00:00Z",
            "initialOutstandingPrincipal":"5000.00",
            "interest":"20.83",
            "principal":"198.53",
            "remainingOutstandingPrincipal":"4801.47"
        },
        {
            "borrowerPaymentAmount":"219.36",
            "date":"2018-02-01T00:00:00Z",
            "initialOutstandingPrincipal":"4801.47",
            "interest":"20.01",
            "principal":"199.35",
            "remainingOutstandingPrincipal":"4602.12"
        },
        {
            "borrowerPaymentAmount":"219.28",
            "date":"2019-12-01T00:00:00Z",
            "initialOutstandingPrincipal":"218.37",
            "interest":"0.91",
            "principal":"218.37",
            "remainingOutstandingPrincipal":"0"
        }
    ]
}
```
