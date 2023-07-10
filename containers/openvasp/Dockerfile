# Dynamic Builds
ARG BUILDER_IMAGE=golang:1.19-buster
ARG FINAL_IMAGE=debian:buster-slim

# Build stage
FROM ${BUILDER_IMAGE} as builder

WORKDIR $GOPATH/src/github.com/trisacrypto/testnet

# Use modules for dependencies
COPY go.mod .
COPY go.sum .

# Copy only what is needed (ignore web and fixtures)
COPY cmd ./cmd
COPY pkg ./pkg

# Build binary
RUN go build -v -o /go/bin/openvasp -ldflags="-X 'github.com/trisacrypto/testnet/pkg.GitVersion=$GIT_REVISION'" ./cmd/openvasp

# Final Stage
FROM ${FINAL_IMAGE} AS final

# Copy the binary to the production image from the builder stage.
COPY --from=builder /go/bin/openvasp /usr/local/bin/openvasp

CMD [ "/usr/local/bin/openvasp", "serve" ]

