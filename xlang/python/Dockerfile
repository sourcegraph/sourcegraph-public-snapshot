# Docker image for JavaScript/TypeScript lang/build server

FROM python:3.5-alpine

ADD python-langserver /src
WORKDIR /src

RUN pip install --upgrade pip && pip install -r requirements.txt

EXPOSE 2087
CMD ["python3", "python-langserver.py", "--mode", "tcp", "--addr", "2087"]
