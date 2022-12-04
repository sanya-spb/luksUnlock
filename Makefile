PROJECT?=luksUnlock
PROJECTNAME=$(shell basename "$(PROJECT)")

TARGETOS?=linux
TARGETARCH?=amd64

CGO_ENABLED=0

RELEASE := $(shell git tag -l | tail -1 | grep -E "v.+"|| echo devel)
COMMIT := git-$(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COPYRIGHT := "sanya-spb"

## cert: Make certs
cert:
	mkdir -p ./cert/dropbear_export
	ssh-keygen -a 100 -f ./cert/id_ed25519 -t ed25519
	cat ./cert/id_ed25519.pub > ./cert/dropbear_export/authorized_keys

## build: Build luksUnlock
build:
	mkdir -p ./bin
	GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=${CGO_ENABLED} go build \
		-ldflags "-s -w \
		-X ${PROJECT}/pkg/version.version=${RELEASE} \
		-X ${PROJECT}/pkg/version.commit=${COMMIT} \
		-X ${PROJECT}/pkg/version.buildTime=${BUILD_TIME} \
		-X ${PROJECT}/pkg/version.copyright=${COPYRIGHT}" \
		-o ./bin/luksUnlock ./cmd/

## image: Build luksUnlock images
image:
	docker build -t luksUnlock \
	--build-arg RELEASE=${RELEASE} \
	--build-arg COMMIT=${COMMIT} \
	--build-arg BUILD_TIME=${BUILD_TIME} \
	.
	@echo "\n\nTo start container:"
	@echo 'docker run -dit --restart unless-stopped -p 8080:8080 -v $(pwd)/conf:/app/data/conf --name luksUnlock luksUnlock:latest'

## run: Run luksUnlock
run:
	go run ./cmd/ -config ./configs/config.yaml

## clean: Clean build files
clean:
	go clean
	rm -v ./bin/* 2> /dev/null || true

## test: Run unit test
test:
	go test -v -short ${PROJECT}/...

## integration: Run integration test
integration:
	# go test -v -run Integration ${PROJECT}/...

## help: Show this
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
