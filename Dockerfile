# Build Daemon
FROM golang:1.20-alpine AS build-dm
WORKDIR /build

# Copy source files
COPY cmd cmd
COPY internal internal
COPY pkg pkg
COPY go.mod .
COPY go.sum .
# Get go packages
RUN go mod download
# Build daemon
RUN go build -o ./bin/daemon ./cmd/daemon/main.go

# Set up runtime
FROM alpine:3 AS runtime
WORKDIR /app

# Copy daemon
COPY --from=build-dm /build/bin/daemon /app/daemon

# Prepare config directory
RUN mkdir -p /etc/config

ENTRYPOINT ["/app/daemon"]
CMD ["-c", "/etc/daemon/config.toml"]