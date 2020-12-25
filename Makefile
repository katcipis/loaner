goversion=1.15.6
golangci_lint_version=1.33
short_sha=$(shell git rev-parse --short HEAD)
version?=$(short_sha)
img=katcipis/loaner:$(version)
vols=-v `pwd`:/app -w /app
run_go=docker run --rm $(vols) golang:$(goversion)
run_lint=docker run --rm $(vols) golangci/golangci-lint:v$(golangci_lint_version)
cov=coverage.out
covhtml=coverage.html


.PHONY: all
all: test lint

.PHONY: test
test:
	$(run_go) go test -coverprofile=$(cov) -race ./...

.PHONY: coverage
coverage: test
	@$(run_go) go tool cover -html=$(cov) -o=$(covhtml)
	@open $(covhtml) || xdg-open $(covhtml)

.PHONY: lint
lint:
	@$(run_lint) golangci-lint run ./...

.PHONY: image
image:
	docker build -t $(img) --build-arg GOVERSION=$(goversion) .
