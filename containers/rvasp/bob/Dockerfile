ARG BASE
FROM ${BASE}

LABEL maintainer="TRISA <info@trisa.io>"
LABEL description="Bob rVASP for TRISA demonstration and integration"

ENV RVASP_NAME bob
RUN rvasp initdb
RUN rvasp.py
