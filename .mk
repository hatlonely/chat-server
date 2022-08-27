NAME ?= chat-server
REGISTRY_ENDPOINT ?= docker.io
REGISTRY_NAMESPACE ?=
VERSION ?= $(shell git describe --tags | awk '{print(substr($$0,2,length($$0)))}')

define BUILD_VERSION
  version: $(shell git describe --tags)
gitremote: $(shell git remote -v | grep fetch | awk '{print $$2}')
   commit: $(shell git rev-parse HEAD)
 datetime: $(shell date '+%Y-%m-%d %H:%M:%S')
 hostname: $(shell hostname):$(shell pwd)
goversion: $(shell go version)
endef
export BUILD_VERSION

.PHONY: build
build: cmd/main.go $(wildcard internal/*/*.go) Makefile vendor
	mkdir -p build/bin && mkdir -p build/config && cp config/* build/config
	go build -ldflags "-X 'main.Version=$$BUILD_VERSION'" -o build/bin/${NAME} cmd/main.go

.PHONY: clean
clean:
	rm -rf build

vendor: go.mod go.sum
	go mod tidy
	go mod vendor

.PHONY: codegen
codegen: api/chat-server.proto
	if [ ! -z "$(shell docker ps --filter name=protobuf -q)" ]; then \
		docker stop protobuf; \
	fi
	docker run --name protobuf -d --rm docker.io/hatlonely/protobuf:1.1.1 tail -f /dev/null
	docker exec protobuf mkdir -p api
	docker cp $< protobuf:/$<
	docker exec protobuf bash -c "mkdir -p api/gen/go && mkdir -p api/gen/swagger"
	docker exec protobuf bash -c "protoc -I. --go_out api/gen/go --go_opt paths=source_relative $<"
	docker exec protobuf bash -c "cp api/gen/go/$(basename $<).pb.go api"
	docker exec protobuf bash -c "protoc -I. --go-grpc_out api/gen/go --go-grpc_opt paths=source_relative $<"
	docker cp protobuf:/api/gen api
	docker stop protobuf

.PHONY: image
image:
	docker build --build-arg version="$${BUILD_VERSION}" --tag=${REGISTRY_ENDPOINT}/${REGISTRY_NAMESPACE}/${NAME}:${VERSION} .
