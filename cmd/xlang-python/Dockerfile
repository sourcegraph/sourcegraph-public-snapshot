FROM python:3.6-alpine
RUN apk add --no-cache git && pip install pip pipenv --upgrade

ADD python-langserver /src
WORKDIR /src

RUN pipenv install --deploy 

EXPOSE 2087
ENTRYPOINT ["pipenv", "run", "python", "-X", "faulthandler", "python-langserver.py"]
CMD ["--mode", "tcp", "--addr", "2087"]
