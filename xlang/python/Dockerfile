# Docker image for JavaScript/TypeScript lang/build server

FROM python:3.5-alpine
RUN apk update && apk upgrade && apk add git

ADD python-langserver /src
WORKDIR /src

RUN pip install --upgrade pip && pip install -r requirements.txt

EXPOSE 2087
CMD ["python3", "python-langserver.py", "--mode", "tcp", "--addr", "2087", "--remote", "1"]
