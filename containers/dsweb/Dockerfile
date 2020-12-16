FROM node:14 AS builder

WORKDIR /usr/src/app

# Dependencies
COPY web/trisads/package*.json ./
RUN npm install

# Set production environment variable for build context
ENV TRISADS_API_ENDPOINT="https://proxy.vaspdirectory.net"

# Build app with browserify
COPY web/trisads/ .
RUN npm run build

FROM nginx:stable

LABEL maintainer="TRISA <admin@trisa.io>"
LABEL description="Prototype TRISA Directory Service User Interface"

COPY --from=builder /usr/src/app/build/ /var/www/build/
COPY --from=builder /usr/src/app/index.html /var/www/
COPY --from=builder /usr/src/app/css/ /var/www/css/
COPY --from=builder /usr/src/app/img/ /var/www/img/
COPY --from=builder /usr/src/app/nginx/nginx.conf /etc/nginx/conf.d/default.conf

ENV NGINX_ENTRYPOINT_QUIET_LOGS=1
CMD [ "nginx", "-g", "daemon off;" ]