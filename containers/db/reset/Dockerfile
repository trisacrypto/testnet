FROM golang:1.16-buster AS builder
WORKDIR /rvasp/build
COPY . .
RUN go test ./... && go build -v ./cmd/rvasp

FROM ubuntu:focal
LABEL maintainer="TRISA <info@trisa.io>"
LABEL description="Job to reset the Robot VASP working database"
RUN apt-get update && apt-get install -y postgresql-client
COPY --from=builder /rvasp/build/rvasp /bin/
COPY --from=builder /rvasp/build/scripts/reset-db.sh /bin/
COPY --from=builder /rvasp/build/pkg/rvasp/fixtures /fixtures

CMD ["/bin/reset-db.sh"]