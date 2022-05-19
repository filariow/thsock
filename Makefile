GO := go
GOARCH := arm
GOOS := linux

.PHONY: build protos

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -trimpath -ldflags="-s -w" -o bin/thsock cmd/usock/main.go

ci: build
	scp bin/thsock rpi4:/home/pi/thsock

protos:
	mkdir -p pkg/thprotos
	protoc \
		-I protos \
		--go_opt=paths=source_relative \
		--go_out=pkg/thprotos \
		--go-grpc_opt=paths=source_relative \
		--go-grpc_out=pkg/thprotos \
		protos/*.proto

