FROM gcr.io/distroless/static as final
CMD ["/go/bin/thctl"]

FROM golang:1.18 as builder
WORKDIR /app
COPY "go.mod" "go.mod"
COPY "go.sum" "go.sum"
RUN go mod download
RUN go mod verify

COPY "cmd/thctl" "cmd/thctl"
COPY "internal" "internal"
COPY "pkg" "pkg"

RUN CGO_ENABLED=0 go build -a -installsuffix cgo \
	-trimpath \
	-ldflags="-s -w" \
	-o /tmp/thctl \
	cmd/thctl/main.go

FROM final
COPY --from=builder /tmp/thctl /go/bin/thctl
