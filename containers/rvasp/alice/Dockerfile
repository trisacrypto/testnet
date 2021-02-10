ARG BASE
FROM ${BASE}

LABEL maintainer="TRISA <info@trisa.io>"
LABEL description="Alice rVASP for TRISA demonstration and integration"

ENV RVASP_NAME alice
RUN rvasp initdb
RUN rvasp.py
