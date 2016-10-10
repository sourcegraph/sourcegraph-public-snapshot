FROM alpine:3.4

RUN apk add --no-cache ca-certificates
RUN apk add --update \
    python \
    python-dev \
    py-pip \
    build-base \
  && pip install virtualenv \
  && rm -rf /var/cache/apk/*

ADD ./*.py /e2e/
ADD Makefile /e2e/Makefile
ADD requirements.txt /e2e/requirements.txt

WORKDIR /e2e
RUN make dep

ENTRYPOINT ["/e2e/.env/bin/python", "run.py"]
