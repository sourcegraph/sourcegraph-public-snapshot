FROM python:3.6-alpine
RUN apk update && apk upgrade && apk add git && pip install --upgrade pip
RUN pip install pipenv --upgrade

ADD python-langserver /src
WORKDIR /src

RUN pipenv install --deploy 

EXPOSE 2087
CMD ["pipenv", "run", "python", "-X", "faulthandler", "python-langserver.py", "--mode", "tcp", "--addr", "2087"]
