FROM python:3.6-alpine3.7
RUN apk add --no-cache git==2.15.2-r0 tini==0.16.1-r0 shadow==4.5-r0 && pip install pip==9.0.3 pipenv==11.9.0 --upgrade

COPY python-langserver /home/langserver/src
WORKDIR /home/langserver/src

ENV PIPENV_VENV_IN_PROJECT=true
RUN pipenv install --deploy

# Allow all users to read/write/execute all files (to enable user override at runtime)
RUN chmod -R 777 /home/langserver/src
RUN mkdir -p /.cache && chmod -R 777 /.cache

EXPOSE 2087
ENTRYPOINT ["/sbin/tini", "--", "pipenv", "run", "python", "-X", "faulthandler", "python-langserver.py"]
CMD ["--mode", "tcp", "--addr", "2087"]
