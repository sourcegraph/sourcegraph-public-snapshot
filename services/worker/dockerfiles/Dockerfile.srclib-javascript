#
# Docker image for srclib-javascript
#

#
# Dependencies
#
FROM node:5.7.0-slim

RUN apt-get update -y
RUN apt-get install -qq git

#
# Install srclib-javascript executable
#
ENV SRCLIBPATH /srclib/toolchains
ARG TOOLCHAIN_URL
ADD $TOOLCHAIN_URL /toolchain/t
RUN (cd /toolchain && tar xfz t && rm t && mv * /toolchain/t) || true
RUN mkdir -p $SRCLIBPATH/sourcegraph.com/sourcegraph && mv /toolchain/t $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-javascript
WORKDIR $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-javascript
RUN npm install

#
# Install srclib binary (assumes this has been built on the Docker host)
#
ADD ./srclib /bin/srclib

# Run environment
ENTRYPOINT srclib config && srclib make
