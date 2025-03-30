FROM nginx:stable

LABEL maintainer="TRISA <admin@trisa.io>"
LABEL description="Redirect testnet.directory to trisa.dev"

COPY containers/docs-redirect/nginx.conf /etc/nginx/conf.d/default.conf

ENV NGINX_ENTRYPOINT_QUIET_LOGS=1
CMD [ "nginx", "-g", "daemon off;" ]