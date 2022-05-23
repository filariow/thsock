GO := go
GOARCH := arm
GOOS := linux

.PHONY: protos ci-thl ci-ihm build-ctl build-srv

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

ci-ths:
	docker build -f deploy/thsampler/Dockerfile -t unina/thsampler:dev .
	docker save --output /tmp/thsampler.dev.tar unina/thsampler:dev
	sudo k3s ctr images import /tmp/thsampler.dev.tar
	sudo rm -f /tmp/thsampler.dev.tar

ci-thl:
	docker build -f deploy/thlooper/Dockerfile -t unina/thlooper:mqtt .
	docker save --output /tmp/thlooper.mqtt.tar unina/thlooper:mqtt	
	sudo k3s ctr images import /tmp/thlooper.mqtt.tar
	sudo rm -f /tmp/thlooper.mqtt.tar

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

