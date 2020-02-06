EXECUTABLE ?= jalapeno
BINARY ?= bin/$(EXECUTABLE)
REPO=iejalapeno
IMAGE=topology
TAG=3.1

all: build

build:
	CGO_ENABLED=0 go build --ldflags '${EXTLDFLAGS}' -o ${BINARY} github.com/ciscosrapps/topology

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${BINARY} github.com/ciscosrapps/topology

container:
	docker build -t ${REPO}/${IMAGE}:${TAG} .
