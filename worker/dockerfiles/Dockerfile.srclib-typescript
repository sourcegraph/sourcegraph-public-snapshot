#
# Docker image for srclib-typescript
#

# Install dependencies
FROM node:5.7.0-slim

RUN apt-get update -y
RUN apt-get install -qq git make

# Install srclib-typescript executable
RUN npm install typescript -g
ENV SRCLIBPATH /srclib/toolchains
ARG TOOLCHAIN_URL
ADD $TOOLCHAIN_URL /toolchain/t
RUN (cd /toolchain && tar xfz t && rm t && mv * /toolchain/t) || true
RUN mkdir -p $SRCLIBPATH/sourcegraph.com/sourcegraph && mv /toolchain/t $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-typescript
WORKDIR $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-typescript
RUN make

# Install srclib binary (assumes this has been built on the Docker host)
ADD ./srclib /bin/srclib

ENTRYPOINT srclib config && srclib make
