BINARY=prme
DOCKER_TAG=pr-me
DOCKER_REGISTRY=ivanfetch
VERSION= $(shell git describe --tags --dirty 2>/dev/null || echo snapshot)
GIT_COMMIT=$(shell git rev-parse HEAD)
LDFLAGS="-s -w -X prme.Version=$(VERSION) -X prme.GitCommit=$(GIT_COMMIT)"

all: build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:go.sum
	go vet ./...

go.sum:go.mod
	go get -t prme

.PHONY: test
test:go.sum
	go test ./...

.PHONY: integrationtest
integrationtest:go.sum
	go test -tags integration ./...

.PHONY: binary
binary:go.sum
	go build -ldflags $(LDFLAGS) -o $(BINARY) cmd/main.go

.PHONY: build
build: fmt vet test binary

.PHONY: docker-build
docker-build: fmt vet integrationtest binary
	docker build -t $(DOCKER_TAG):$(VERSION) .
	@echo Run this container using a command like: docker run -it --rm -e GH_TOKEN=xyz $(DOCKER_TAG):$(VERSION)

.PHONY: docker-push
docker-push:
	docker tag $(DOCKER_TAG):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_TAG):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_TAG):$(VERSION)

