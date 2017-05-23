FROM python:3.6-alpine
RUN apk update && apk upgrade && apk add git && pip install --upgrade pip

ADD python-langserver/requirements.txt /tmp/requirements.txt
RUN pip install -r /tmp/requirements.txt && rm /tmp/requirements.txt

ADD python-langserver /src
WORKDIR /src

EXPOSE 2087
CMD ["python3", "-X", "faulthandler", "python-langserver.py", "--mode", "tcp", "--addr", "2087", "--lightstep_token", "d60b0b2477a7ccb05d7783917f648816"]
