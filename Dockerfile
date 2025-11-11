FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.25 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app/
COPY go.mod ./
RUN go mod download
COPY **/*.go ./
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o dns-updater
RUN chmod +x dns-updater

FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch
WORKDIR /app/
COPY --from=builder /app/dns-updater /app/dns-updater
ENTRYPOINT ["/app/dns-updater"]
