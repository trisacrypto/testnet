FROM python:3.8

LABEL maintainer="TRISA <admin@trisa.io>"
LABEL description="Backend for Frontend (BFF) for rVASP Demo"

# Install rvaspy library
WORKDIR /usr/src/lib
COPY lib/python/rvaspy .
RUN pip install -e .

# Install the Flask app
WORKDIR /usr/src/app/
RUN mkdir flaskr

# Dependencies
COPY web/demo/bff/requirements.txt .
RUN pip install -r requirements.txt

# Application
COPY web/demo/bff/gunicorn_flaskr.sh .
COPY web/demo/bff/flaskr flaskr/

# Setup the environment
ENV FLASK_APP /usr/src/app/flaskr
ENV FLASK_ENV development

# Initialize the database
RUN [ "flask", "init-db" ]

ENTRYPOINT ["flask", "run", "--host", "0.0.0.0", "--port", "8000"]

