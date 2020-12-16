#!/bin/bash

# Get the path to the project no matter where we're running this script from
PROJECT=$(realpath $(dirname $(realpath $0))/..)
CONTAINERS=$PROJECT/containers

# Build the base rvasp image
docker build -t trisacrypto/rvasp:latest -f $PROJECT/Dockerfile $PROJECT

# Build the alice, bob, and evil images
docker build -t trisacrypto/rvasp:alice -f $CONTAINERS/alice/Dockerfile $PROJECT
docker build -t trisacrypto/rvasp:bob -f $CONTAINERS/bob/Dockerfile $PROJECT
docker build -t trisacrypto/rvasp:evil -f $CONTAINERS/evil/Dockerfile $PROJECT