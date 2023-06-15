# Dynamic Builds
ARG BUILDER_IMAGE=golang:1.19-buster
ARG FINAL_IMAGE=debian:buster-slim

# Build stage
FROM ${BUILDER_IMAGE} as builder

ARG GIT_REVISION=""

# Ensure ca-certificates are up to date
RUN update-ca-certificates

WORKDIR $GOPATH/src/github.com/trisacrypto/testnet

# Use modules for dependencies
COPY go.mod .
COPY go.sum .

ENV CGO_ENABLED=0
ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

# Copy only what is needed (ignore web and fixtures)
COPY cmd ./cmd
COPY pkg ./pkg
COPY proto ./proto

# Run tests to avoid building a broken container
RUN go test -timeout 180s ./...

# Build binary
RUN go build -v -o /go/bin/rvasp -ldflags="-X 'github.com/trisacrypto/testnet/pkg.GitVersion=$GIT_REVISION'" ./cmd/rvasp

# Final Stage
FROM ${FINAL_IMAGE} AS final

LABEL maintainer="TRISA <info@trisa.io>"
LABEL description="Robot VASP for TRISA demonstration and integration"

# Ensure ca-certificates are up to date
RUN set -x && apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /go/bin/rvasp /usr/local/bin/rvasp

CMD [ "/usr/local/bin/rvasp", "serve" ]