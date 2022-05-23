GO := go
GOARCH := arm
GOOS := linux

.PHONY: protos ci-thctl ci-thlooper ci-ihm build-ctl build-srv

build-srv:
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_ENABLED=1 \
	$(GO) build \
		-a \
		-installsuffix cgo \
		-trimpath \
		-ldflags="-s -w" \
		-o bin/thsock  \
		cmd/usock/main.go

build-ctl:
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_ENABLED=0 \
	$(GO) build \
		-trimpath \
		-ldflags="-s -w" \
		-o bin/thctl \
		cmd/thctl/main.go

ci-thctl:
	docker build -f deploy/thctl/Dockerfile -t unina/thctl:dev .
	docker save --output /tmp/thctl.dev.tar unina/thctl:dev
	sudo k3s ctr images import /tmp/thctl.dev.tar
	sudo rm -f /tmp/thctl.dev.tar

ci-thlooper:
	docker build -f deploy/thlooper/Dockerfile -t unina/thlooper:dev .
	docker save --output /tmp/thlooper.dev.tar unina/thlooper:dev	
	sudo k3s ctr images import /tmp/thlooper.dev.tar
	sudo rm -f /tmp/thlooper.dev.tar

ci-ihm:
	docker build -f deploy/iothubbroker/Dockerfile -t unina/iothubmqtt:dev .
	docker save --output /tmp/iothubmqtt.dev.tar unina/iothubmqtt:dev
	sudo k3s ctr images import /tmp/iothubmqtt.dev.tar
	sudo rm -f /tmp/iothubmqtt.dev.tar

protos:
	mkdir -p pkg/thprotos
	protoc \
		-I protos \
		--go_opt=paths=source_relative \
		--go_out=pkg/thprotos \
		--go-grpc_opt=paths=source_relative \
		--go-grpc_out=pkg/thprotos \
		protos/*.proto

