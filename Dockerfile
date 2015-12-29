# Sourcegraph Dockerfile
#
# See https://src.sourcegraph.com/sourcegraph/.docs/install/docker/
# for Docker installation instructions.
#

FROM ubuntu:14.04
MAINTAINER Sourcegraph Team <help@sourcegraph.com>

RUN apt-get update -q \
    && apt-get install -qy --no-install-recommends \
            curl \
            git \
            make \
            wget \
            ca-certificates

RUN apt-get install -qq curl git python-software-properties software-properties-common
RUN curl -sL https://deb.nodesource.com/setup_4.x | bash -
RUN apt-get install -y nodejs

# Install Go
RUN curl -Ls https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH=$PATH:/usr/local/go/bin

## Install the protobuf compiler (required for building the binary)
##
## NOTE: Disabled because it takes a long time and is usually only
## needed/wanted on Sourcegraph.com.
# COPY dev/install_protobuf.sh /tmp/install_protobuf.sh
# RUN apt-get install -qy unzip autoconf libtool
# RUN /tmp/install_protobuf.sh /

# Trust GitHub's SSH host key (for ssh cloning of repos during builds)
COPY package/etc/known_hosts /tmp/known_hosts
RUN install -Dm 600 /tmp/known_hosts /root/.ssh/known_hosts \
    && chmod 700 /root/.ssh

# Build and install src
ENV GOPATH=/usr/local
COPY . /usr/local/src/src.sourcegraph.com/sourcegraph
WORKDIR /usr/local/src/src.sourcegraph.com/sourcegraph
RUN GOPATH=$PWD/Godeps/_workspace:$GOPATH go install ./sgtool
RUN sgtool package --os linux --skip-protoc --ignore-dirty
RUN mv release/snapshot/linux-amd64 /usr/local/bin/src
RUN mkdir /etc/sourcegraph
RUN cp package/linux/default.ini /etc/sourcegraph/config.ini

# Copy config so we can create the config files on the host if we are
# host-mounting the config dir.
RUN cp -R /etc/sourcegraph /etc/sourcegraph.default

# Remove source dir (to avoid accidental source dependencies)
WORKDIR /
RUN rm -rf /usr/local/src/src.sourcegraph.com

VOLUME ["/etc/sourcegraph", "/home/sourcegraph/.sourcegraph"]

EXPOSE 3080 3443

CMD cp --no-clobber /etc/sourcegraph.default/* /etc/sourcegraph && src --config /etc/sourcegraph/config.ini serve --http-addr=:80 --https-addr=:443 --ssh-addr=:22
