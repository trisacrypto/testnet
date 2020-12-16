FROM golang:1.14.5 AS builder

WORKDIR /srv/build

COPY . .
RUN go test ./... && go build -v ./cmd/rvasp

FROM ubuntu:focal

LABEL maintainer="TRISA <info@trisa.io>"
LABEL description="Robot VASP for TRISA demonstration and integration"

RUN apt-get update
RUN apt-get install -y ca-certificates
RUN apt-get install -y wget gnupg
RUN apt-get install -y python3

COPY --from=builder /srv/build/rvasp /bin/
COPY --from=builder /srv/build/fixtures/initdb.py /bin/

RUN mkdir /data
ENV DATABASE_URL=/data/rvasp.db

ENTRYPOINT [ "/bin/rvasp", "serve" ]