#!/bin/bash

docker build -t trisa/rvasp:latest -f ./containers/rvasp/Dockerfile .
docker compose -f ./containers/docker-compose.yml build