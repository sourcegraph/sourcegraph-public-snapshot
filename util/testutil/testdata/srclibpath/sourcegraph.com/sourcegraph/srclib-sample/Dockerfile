FROM alpine:3.2

RUN apk --update add git curl bash

ENV SRCLIBPATH /srclib
ADD . /srclib/sourcegraph.com/sourcegraph/srclib-sample

RUN curl -Lo /tmp/srclib.gz https://srclib-release.s3.amazonaws.com/srclib/0.1.1-no-docker4/linux-amd64/srclib.gz && cd /tmp && gunzip -f srclib.gz && chmod +x srclib && mv srclib /bin/srclib

ENTRYPOINT srclib config && srclib make
