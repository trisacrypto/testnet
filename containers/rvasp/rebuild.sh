#!/bin/bash

ask() {
    local prompt default reply

    if [[ ${2:-} = 'Y' ]]; then
        prompt='Y/n'
        default='Y'
    elif [[ ${2:-} = 'N' ]]; then
        prompt='y/N'
        default='N'
    else
        prompt='y/n'
        default=''
    fi

    while true; do

        # Ask the question (not using "read -p" as it uses stderr not stdout)
        echo -n "$1 [$prompt] "

        # Read the answer (use /dev/tty in case stdin is redirected from somewhere else)
        read -r reply </dev/tty

        # Default?
        if [[ -z $reply ]]; then
            reply=$default
        fi

        # Check if the reply is valid
        case "$reply" in
            Y*|y*) return 0 ;;
            N*|n*) return 1 ;;
        esac

    done
}

# Get the tag as the first argument or from git if none is supplied
if [ -z "$1" ]; then
    TAG=$(git rev-parse --short HEAD)
else
    TAG=$1
fi

# Confirm that we're continuing with the tag
if ! ask "Continue with tag $TAG?" N; then
    exit 1
fi

# Set some helpful variables
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
REPO=$(realpath "$DIR/../..")

GIT_REVISION=$(git rev-parse --short HEAD)

# Build the base rvasp image
docker build --platform linux/amd64 --build-arg GIT_REVISION=${GIT_REVISION} -t trisa/rvasp:$TAG -f $DIR/Dockerfile $REPO
docker build --platform linux/amd64 -t trisa/rvasp-migrate:$TAG -f $DIR/../db/Dockerfile $REPO

# Retag the alice, bob, and evil images
docker tag trisa/rvasp:$TAG gcr.io/trisa-gds/rvasp:$TAG
docker tag trisa/rvasp-migrate:$TAG gcr.io/trisa-gds/rvasp-migrate:$TAG

docker push trisa/rvasp:$TAG
docker push trisa/rvasp-migrate:$TAG

docker push gcr.io/trisa-gds/rvasp:$TAG
docker push gcr.io/trisa-gds/rvasp-migrate:$TAG