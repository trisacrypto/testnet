FROM golang:1.16-buster AS builder

WORKDIR /srv/build

COPY . .
RUN go test ./... && go build -v ./cmd/rvasp

FROM ubuntu:focal

LABEL maintainer="TRISA <info@trisa.io>"
LABEL description="Robot VASP for TRISA demonstration and integration"

RUN apt-get update
RUN apt-get install -y ca-certificates
RUN apt-get install -y wget gnupg

COPY --from=builder /srv/build/rvasp /bin/

ENTRYPOINT [ "/bin/rvasp", "serve" ]