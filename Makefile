GO := go
GOARCH := arm
GOOS := linux

.PHONY: protos ci-thsampler ci-thlooper ci-ihm build-ctl build-srv

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

ci-build-srv:
	docker run --rm -it -v $(realpath .):/workspace -w /workspace  golang:1.18 make build-srv
	sudo install ./bin/thsock /usr/local/bin/thsock 

build-ctl:
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_ENABLED=0 \
	$(GO) build \
		-trimpath \
		-ldflags="-s -w" \
		-o bin/thsampler \
		cmd/thsampler/main.go

ci-thsampler:
	docker build -f deploy/thsampler/Dockerfile -t unina/thsampler:dev .
	docker save --output /tmp/thsampler.dev.tar unina/thsampler:dev
	sudo k3s ctr images import /tmp/thsampler.dev.tar
	sudo rm -f /tmp/thsampler.dev.tar

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

