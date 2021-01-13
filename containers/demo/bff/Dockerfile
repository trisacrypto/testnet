FROM python:3.8

WORKDIR /
RUN mkdir /flaskr

COPY requirements.txt .

RUN pip install -r requirements.txt

ENV FLASK_APP flaskr
ENV FLASK_ENV development

COPY flaskr/ /flaskr

RUN [ "flask", "init-db" ]
ENTRYPOINT [ "flask", "run", "--host=0.0.0.0"]
#ENTRYPOINT [ "waitress-serve", "--port=80", "--call", "flaskr:create_app" ]
