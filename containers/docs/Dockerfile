FROM nginx:alpine as builder

# Build Dependencies
RUN apk add --update wget git

# Fetch and install Hugo
ARG HUGO_VERSION="0.79.0"
RUN wget --quiet "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_${HUGO_VERSION}_Linux-64bit.tar.gz" && \
    tar xzf hugo_${HUGO_VERSION}_Linux-64bit.tar.gz && \
    rm -r hugo_${HUGO_VERSION}_Linux-64bit.tar.gz && \
    mv hugo /usr/bin

# Move files over and initialize submodule for theme
WORKDIR /site
COPY docs/ .

# Build app with Hugo
RUN hugo

FROM nginx:alpine

LABEL maintainer="TRISA <admin@trisa.io>"
LABEL description="TRISA Test Net Documentation"

COPY --from=builder /site/public /usr/share/nginx/html

WORKDIR /usr/share/nginx/html