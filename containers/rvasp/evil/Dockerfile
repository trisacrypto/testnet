ARG BASE
FROM ${BASE}

LABEL maintainer="TRISA <info@trisa.io>"
LABEL description="Evil rVASP for TRISA demonstration and integration"

ENV RVASP_NAME evil
RUN rvasp initdb
RUN rvasp.py
