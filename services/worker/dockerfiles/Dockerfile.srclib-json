#
# Docker image for srclib-json
#

# Install dependencies
FROM golang:1.6-alpine

RUN apk --update add git make

# Install srclib-json executable
ENV GOPATH /srclib/toolchains
ENV PATH $PATH:$GOPATH/bin
ENV SRCLIBPATH $GOPATH/src
ARG TOOLCHAIN_URL
ADD $TOOLCHAIN_URL /toolchain/t
RUN (cd /toolchain && tar xfz t && rm t && mv * t) || true
RUN mkdir -p $SRCLIBPATH/sourcegraph.com/sourcegraph && mv /toolchain/t $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-json
WORKDIR $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-json
RUN make clean && make

# Install srclib binary (assumes this has been built on the Docker host)
ADD ./srclib /bin/srclib

# Run environment
ENV GOPATH /drone
ENTRYPOINT srclib config && srclib make