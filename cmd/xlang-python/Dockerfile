FROM python:3.6-alpine3.7
RUN apk add --no-cache git==2.15.2-r0 tini==0.16.1-r0 shadow==4.5-r0 && pip install pip==9.0.3 pipenv==11.9.0 --upgrade

RUN addgroup -S langserver && adduser -S -G langserver -h /home/langserver langserver

COPY python-langserver /home/langserver/src
RUN chown -R langserver:langserver /home/langserver
WORKDIR /home/langserver/src

USER langserver

ENV PIPENV_VENV_IN_PROJECT=true
RUN pipenv install --deploy

EXPOSE 2087
ENTRYPOINT ["/sbin/tini", "--", "pipenv", "run", "python", "-X", "faulthandler", "python-langserver.py"]
CMD ["--mode", "tcp", "--addr", "2087"]
