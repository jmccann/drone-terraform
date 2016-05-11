.PHONY: all clean deps fmt vet test docker

EXECUTABLE ?= drone-terraform
IMAGE ?= plugins/$(EXECUTABLE)
COMMIT ?= $(shell git rev-parse --short HEAD)

LDFLAGS = -X "main.buildCommit=$(COMMIT)"
PACKAGES = $(shell go list ./... | grep -v /vendor/)

all: deps build test

clean:
	go clean -i ./...

deps:
	go get -d github.com/hashicorp/terraform
	go get -t ./...

fmt:
	go fmt $(PACKAGES)

vet:
	go vet $(PACKAGES)

test:
	@for PKG in $(PACKAGES); do go test -cover -coverprofile $$GOPATH/src/$$PKG/coverage.out $$PKG || exit 1; done;

docker:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o terraform github.com/hashicorp/terraform
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-s -w $(LDFLAGS)'
	docker build --rm -t $(IMAGE) .

$(EXECUTABLE): $(wildcard *.go)
	go build -o terraform github.com/hashicorp/terraform
	go build -ldflags '-s -w $(LDFLAGS)'

build: $(EXECUTABLE)
