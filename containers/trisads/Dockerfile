FROM golang:1.14.5 AS builder

WORKDIR /srv/build

COPY . .
RUN go test ./... && go build -v ./cmd/trisads

FROM ubuntu:bionic

LABEL maintainer="TRISA <info@trisa.io>"
LABEL description="Simple TRISA Directory Service"

RUN apt-get update && apt-get install -y ca-certificates
RUN apt-get update && apt-get install -y wget gnupg

COPY --from=builder /srv/build/trisads /bin/

ENV TRISADS_DATABASE /data
RUN mkdir ${TRISADS_DATABASE}

ENTRYPOINT [ "/bin/trisads", "serve" ]