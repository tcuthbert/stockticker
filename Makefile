ifneq (,$(wildcard ./.env))
    include .env
    export
endif
GITHUB_USERNAME=tcuthbert
BINARY_NAME=stockticker
GOARCH=amd64
GOOS=linux

VERSION := $(shell git describe --dirty)
SEMVER := $(VERSION:v%=%)
COMMIT := $(shell git rev-parse --verify HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

DOCKER_IMAGE_TAG ?= $(SEMVER)

GOPATH = $(shell go env GOPATH)
PROJECT_ROOT = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BUILD_DIR := ${PROJECT_ROOT}/build

K8S_APIKEY = kubernetes/base/apikey.txt

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION} -X main.COMMIT=${COMMIT} -X main.BRANCH=${BRANCH}"

all: clean test vet build

$(BUILD_DIR):
	@if [ ! -d "$@" ]; then \
		mkdir -p "$@"; \
	fi

$(K8S_APIKEY):
	@if [ ! -f "$@" -a ! -z ${APIKEY} ]; then \
		echo -n ${APIKEY} > $@; \
	fi

build: $(BUILD_DIR)
	GOARCH=${GOARCH} GOOS=${GOOS} go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} .

build-docker: DOCKER_FLAGS ?= --no-cache
build-docker: Dockerfile
	docker build ${DOCKER_FLAGS} -t ghcr.io/${GITHUB_USERNAME}/${BINARY_NAME}:${DOCKER_IMAGE_TAG} .

push-docker:
	docker push ghcr.io/${GITHUB_USERNAME}/${BINARY_NAME}:${DOCKER_IMAGE_TAG} 

k8s-deploy-%: $(K8S_APIKEY)
	mkctl create -k kubernetes/$(*F)/

k8s-delete-dev: $(K8S_APIKEY)
	mkctl delete -k kubernetes/dev/

k8s-kustomize-%: $(K8S_APIKEY)
	mkctl kustomize kubernetes/$(*F)/

k8s-curl-test:
	@mkctl get svc | awk '/^stockticker/{printf "http://%s", $$3}' | xargs curl -s | jq .

run: build
	${BUILD_DIR}/${BINARY_NAME}

clean: $(BUILD_DIR)
	go clean || true
	if [ -f $</${BINARY_NAME} ]; then \
		rm -v $</${BINARY_NAME}; \
	fi

test:
	go test ./...

coverage:
	go test ./... -coverprofile=coverage.out

dep:
	go mod download && go mod verify

vet:
	go vet

fmt:
	go fmt ./...

lint:
	golangci-lint run --enable-all

.PHONY: all build run clean test test_coverage dep vet fmt lint
