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

This is a service responsible for creating payment plans for loans.

# API

The reference documentation for the API can be found [here](docs/api.md).

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
curl http://localhost:8080/v1/loan/plan -X POST -d '{"todo":"example"}'
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
