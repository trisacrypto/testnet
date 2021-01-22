FROM node:14 AS builder

WORKDIR /usr/src/app

# Dependendencies
COPY web/demo/npm/package*.json ./
ENV PATH node_modules/.bin:$PATH
RUN npm install --silent

# Build the React app
COPY web/demo/npm/ .
RUN npm run build

FROM nginx:stable

LABEL maintainer="TRISA <admin@trisa.io>"
LABEL description="rVASP Demo React Frontend Application"

COPY --from=builder /usr/src/app/build/ /var/www/
COPY --from=builder /usr/src/app/nginx/nginx.conf /etc/nginx/conf.d/default.conf

ENV NGINX_ENTRYPOINT_QUIET_LOGS=1
CMD [ "nginx", "-g", "daemon off;" ]
