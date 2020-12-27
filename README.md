<!-- mdtocstart -->

# Table of Contents

- [Loaner](#loaner)
- [API](#api)
- [Development](#development)
    - [Dependencies](#dependencies)
    - [Running tests](#running-tests)
    - [Linting](#linting)
    - [Releasing](#releasing)
    - [Running Locally](#running-locally)
- [Deployment](#deployment)

<!-- mdtocend -->

# Loaner

![CI Status](https://github.com/katcipis/loaner/workflows/CI/badge.svg)[![Go Report Card](https://goreportcard.com/badge/github.com/katcipis/loaner)](https://goreportcard.com/report/github.com/katcipis/loaner)

Loaner is a service responsible for creating payment plans for loans.

The reference documentation for the API can be found [here](api.md).

# Development

## Dependencies

To run the automation provided here you need to have installed:

* [Make](https://www.gnu.org/software/make/)
* [Docker](https://docs.docker.com/get-docker/)

It is recommended to just use the provided automation through make,
it will help you achieve consistent results across different hosts
and even different operational systems (it is also used on the CI).

If you fancy building things (or running tests) with no extra layers
you can install the latest [Go](https://golang.org/doc/install) and run
Go commands like:

```sh
go test -race ./...
```

## Running tests

Just run:

```
make test
```

To check locally the coverage from all the tests run:

```
make coverage
```

And it should open the coverage analysis in your browser.


## Linting

To lint code, just run:

```
make lint
```

## Releasing

To create an image ready for production deployment just run:

```
make image
```

And one will be created tagged with the git short revision of the
code used to build it, you can also specify an explicit version
if you want:

```
make image version=1.12.0
```

## Running Locally

If you want to explore a locally running version of the service just run:

```
make run
```

And the service will be available at port 8080.

Here is an example of how to make a request to the service with cURL:

```sh
curl http://localhost:8080/loan-plan -X POST -d '{"loanAmount":"5000","nominalRate":"5.0","duration":24,"startDate": "2018-01-01T00:00:01Z"}'
```

# Deployment

To deploy the service you can use Docker images or build the
service directly on the host and copy the binary.

To publish the image, which later can be used on deployment
(eg: on Kubernetes) run:

```sh
make publish
```

To build just the binary run:

```sh
make build
```

And test the binary with:

```sh
./cmd/loaner/loaner --version
```

# Design

One of the main design principles that I like to apply in code
is the single responsibility principle. Even though it is debatable
what a "single responsibility" is (context sensitive)
it is fairly obvious to me that core logic should be decoupled from delivery
mechanisms.

So the core logic should not be coupled to the fact
that the service is delivered through HTTP and the response payloads
are JSON, etc. Changes on how the logic is delivered to clients should
not be reflected on changes on the core logic (loan planning).

To enforce that I usually first develop all the core logic without
writing a single HTTP related code and add this layer on the end
doing a final end to end integration test.

That is why there is two separate packages, the **loan** one with all the
core logic of the service and the **api** one that exports the loan logic
through HTTP.


# FAQ

## Why decimal lib ?

I'm not extensively experienced with financial calculations but I know
enough about them and floating point precision issues to know
that it is not safe to do calculation regarding money using floating point.
It is very easy for cumulative precision errors to sum up and making
people lose (or gain) money. So I did some searching and found the
library used here.

## Why vendor ?

When I started programming in Go I had very mixed feelings with vendoring
since it was the first language that presented this idea as a first class
concept. In time I learned to appreciate it, even when pull requests got
big because of changes on vendor that helped the reality of the complexity
introduced by the dependencies to sink in (I even found that a third party
library panicked because of that).

With the advent of Go modules vendoring stopped being considered one
of the main ways to handle dependencies, but I still appreciate its
simplicity and it inter-operates really well with Go modules (you can
use both).

Also running tests and linting inside containers gets faster without having
to handle go mod caching complications (by default each container run re-downloads
dependencies). 

## Constraint on start date

The start date day is constrained on the range 01-28 to avoid having to deal
with leap years and months with different amount of days (30/31). It makes
calculating the payment days easier. Almost all systems that I use that will
charge me monthly, like credit cards and loans, just give me a few days along
the month as options for payment days, so it seems like a reasonable constraint.
