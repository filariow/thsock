FROM gcr.io/distroless/static as final
CMD ["/go/bin/thsampler"]

FROM golang:1.18 as builder
WORKDIR /app
COPY "go.mod" "go.mod"
COPY "go.sum" "go.sum"
RUN go mod download
RUN go mod verify

COPY "cmd/thsampler" "cmd/thsampler"
COPY "internal" "internal"
COPY "pkg" "pkg"

RUN CGO_ENABLED=0 go build \
	-trimpath \
	-ldflags="-s -w" \
	-o /tmp/thsampler \
	cmd/thsampler/main.go

FROM final
COPY --from=builder /tmp/thsampler /go/bin/thsampler
