FROM alpine:3.4

RUN apk add --no-cache ca-certificates
RUN apk add --update \
	curl \
	python \
	python-dev \
	py-pip \
	build-base \
  && pip install virtualenv \
  && rm -rf /var/cache/apk/*

ADD requirements.txt /e2e/requirements.txt
ADD Makefile /e2e/Makefile

WORKDIR /e2e
RUN make dep

ADD ./*.py /e2e/

ENTRYPOINT ["/e2e/.env/bin/python", "run.py"]
