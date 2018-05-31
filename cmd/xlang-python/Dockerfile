FROM python:3.6-alpine
RUN apk add --no-cache git==2.8.6-r0 tini==0.9.0-r1 && pip install pip==9.0.3 pipenv==11.9.0 --upgrade

COPY python-langserver /src
WORKDIR /src

RUN pipenv install --deploy

EXPOSE 2087
ENTRYPOINT ["/sbin/tini", "--", "pipenv", "run", "python", "-X", "faulthandler", "python-langserver.py"]
CMD ["--mode", "tcp", "--addr", "2087"]
