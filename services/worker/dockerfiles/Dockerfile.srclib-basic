#
# Docker image for srclib-basic
#

#
# Dependencies
#
FROM maven:3-jdk-8

RUN apt-get update -y
RUN apt-get install -qq git wget make

#
# Install srclib-basic executable
#
ENV SRCLIBPATH /srclib/toolchains
ARG TOOLCHAIN_URL
ADD $TOOLCHAIN_URL /toolchain/t
RUN (cd /toolchain && tar xfz t && rm t && mv * t) || true
RUN mkdir -p $SRCLIBPATH/sourcegraph.com/sourcegraph && mv /toolchain/t $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-basic
WORKDIR $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-basic
RUN make

#
# Install srclib binary (assumes this has been built on the Docker host)
#
ADD ./srclib /bin/srclib

# Run environment
ENTRYPOINT srclib config && srclib make
