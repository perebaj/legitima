# TOOLS VERSIONS
export GO_VERSION=1.21.0
export GOLANGCI_LINT_VERSION=v1.54.0

SHELL = /bin/bash


# configuration/aliases
version=$(shell git rev-parse --short HEAD)
base_image=registry.heroku.com/legitima/web
image=$(base_image):latest
devimage=legitima-dev
# To avoid downloading deps everytime it runs on containers
gopkg=$(devimage)-gopkg
gocache=$(devimage)-gocache
devrun=docker run $(devrunopts) --rm \
	-v `pwd`:/app \
	-v $(gopkg):/go/pkg \
	-v $(gocache):/root/.cache/go-build \
	$(devimage)

covreport ?= coverage.txt

all: lint test image

## run isolated tests
.PHONY: test
test:
	go test ./... -timeout 10s -race -shuffle on

## Format go code
.PHONY: fmt
fmt:
	goimports -w .

## builds the service
.PHONY: service
service:
	go build -o ./cmd/legitima/legitima ./cmd/legitima

## runs the service locally
.PHONY: run
run: service
	./cmd/legitima/legitima

## tidy up go modules
.PHONY: mod
mod:
	go mod tidy

## lint the whole project
.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

## generates coverage report
.PHONY: test/coverage
test/coverage:
	go test -count=1 -coverprofile=$(covreport) ./...

## generates coverage report and shows it on the browser locally
.PHONY: test/coverage/show
test/coverage/show: test/coverage
	go tool cover -html=$(covreport)


## Build the service image
.PHONY: image
image:
	docker build . \
		--build-arg GO_VERSION=$(GO_VERSION) \
		-t $(image)

## Build a production ready container image and run it locally for testing.
.PHONY: image/run
image/run: image
	docker run --rm -ti \
		-v $(gopkg):/go/pkg \
		$(image)

## Publish the service image
.PHONY: image/publish
image/publish: image
	docker push $(image)

## Create the dev container image
.PHONY: dev/image
dev/image:
	docker build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION) \
		-t $(devimage) \
		-f Dockerfile.dev \
		.

## Create a shell inside the dev container
.PHONY: dev
dev: devrunopts=-ti
dev: dev/image
	$(devrun)

## run a make target inside the dev container.
dev/%: dev/image
	$(devrun) make ${*}

## Create a new migration, use make migration/new name=<migration_name>
.PHONY: migration/new
migration/new:
	@echo "Creating new migration..."
	go run github.com/golang-migrate/migrate/v4/cmd/migrate \
		create \
		-dir ./mysql/migrations \
		-ext sql \
		-seq \
		$(name)

## Start containers, additionaly you can provide rebuild=true to force rebuild
.PHONY: dev/start
dev/start:
	@echo "Starting development server..."
	@if [ "$(rebuild)" = "true" ]; then \
		docker-compose up -d --build; \
	else \
		docker-compose up -d; \
	fi

## Stop containers
.PHONY: dev/stop
dev/stop:
	@echo "Stopping development server..."
	@docker-compose down

## Display help for all targets
.PHONY: help
help:
	@awk '/^.PHONY: / { \
		msg = match(lastLine, /^## /); \
			if (msg) { \
				cmd = substr($$0, 9, 100); \
				msg = substr(lastLine, 4, 1000); \
				printf "  ${GREEN}%-30s${RESET} %s\n", cmd, msg; \
			} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
