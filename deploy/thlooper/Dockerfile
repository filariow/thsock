FROM gcr.io/distroless/static as final
CMD ["/go/bin/thlooper"]

FROM golang:1.18 as builder
WORKDIR /app
COPY "go.mod" "go.mod"
COPY "go.sum" "go.sum"
RUN go mod download
RUN go mod verify

COPY "cmd/thlooper" "cmd/thlooper"
COPY "internal" "internal"
COPY "pkg" "pkg"

RUN CGO_ENABLED=0 go build \
	-trimpath \
	-ldflags="-s -w" \
	-o /tmp/thlooper \
	cmd/thlooper/main.go

FROM final
COPY --from=builder /tmp/thlooper /go/bin/thlooper
