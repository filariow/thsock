FROM gcr.io/distroless/static as final
CMD ["/go/bin/iothubbroker"]

FROM golang:1.18 as builder
WORKDIR /app
COPY "go.mod" "go.mod"
COPY "go.sum" "go.sum"
RUN go mod download
RUN go mod verify

COPY "cmd/iothubbroker" "cmd/iothubbroker"
COPY "internal" "internal"
COPY "pkg" "pkg"

RUN CGO_ENABLED=0 go build -a -installsuffix cgo \
	-trimpath \
	-ldflags="-s -w" \
	-o /tmp/iothubbroker \
	cmd/iothubbroker/main.go

FROM final
COPY --from=builder /tmp/iothubbroker /go/bin/iothubbroker
