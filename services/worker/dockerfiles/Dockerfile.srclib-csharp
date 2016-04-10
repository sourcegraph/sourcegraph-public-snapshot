#
# Docker image for srclib-csharp
#

# Dependencies
FROM microsoft/aspnet:1.0.0-rc1-update1-coreclr

RUN apt-get update -y
RUN apt-get install -qq git make wget tar

# Add global assembly cache
RUN mkdir /gac
RUN cd /gac && wget https://www.dropbox.com/s/s02lvl8s140f0sc/v4.5.tar.gz && tar -xvf v4.5.tar.gz

# Install srclib-csharp executable
ENV SRCLIBPATH /srclib/toolchains
ARG TOOLCHAIN_URL
ADD $TOOLCHAIN_URL /toolchain/t
RUN (cd /toolchain && tar xfz t && rm t && mv * /toolchain/t) || true
RUN mkdir -p $SRCLIBPATH/sourcegraph.com/sourcegraph && mv /toolchain/t $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-csharp
WORKDIR $SRCLIBPATH/sourcegraph.com/sourcegraph/srclib-csharp
RUN make

# Install srclib binary (assumes this has been built on the Docker host)
ADD ./srclib /bin/srclib

ENTRYPOINT srclib config && srclib make
