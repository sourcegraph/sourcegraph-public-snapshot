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

# Install Java 8 (to bootstrap building our own one, below) -- COPIED from sourcegraph.com/sourcegraph/srclib-java Dockerfile
RUN add-apt-repository ppa:webupd8team/java
RUN apt-get update -qq
# auto accept oracle jdk license
RUN echo oracle-java8-installer shared/accepted-oracle-license-v1-1 select true | /usr/bin/debconf-set-selections
RUN apt-get install -y oracle-java8-installer

# Install Gradle 2.2 -- COPIED from sourcegraph.com/sourcegraph/srclib-java Dockerfile
RUN apt-get install -qq unzip
RUN curl -L -o gradle.zip https://services.gradle.org/distributions/gradle-2.2-bin.zip
RUN unzip gradle.zip
RUN mv gradle-2.2 /gradle
ENV PATH /gradle/bin:${PATH}
RUN ln -s /usr/lib/jvm/java-8-oracle /usr/lib/jvm/default-java

# Deps for building .deb
RUN apt-get install -yq ruby-dev gcc && gem install fpm
COPY package/linux/scripts/install_daemonize.sh /tmp/install_daemonize.sh
RUN bash /tmp/install_daemonize.sh /usr/local/bin

# Install src
ENV GOPATH=/usr/local
RUN go get github.com/shurcooL/vfsgen github.com/gogo/protobuf/protoc-gen-gogo sourcegraph.com/sourcegraph/prototools/cmd/protoc-gen-dump sourcegraph.com/sourcegraph/gopathexec github.com/sqs/go-selfupdate # OPTIMIZATION
COPY . /usr/local/src/src.sourcegraph.com/sourcegraph
WORKDIR /usr/local/src/src.sourcegraph.com/sourcegraph
RUN make dist PACKAGEFLAGS="--os linux --skip-protoc --ignore-dirty"

# Build .deb package for src
RUN make -C package linux/bin/src
RUN make -C package/linux dist/src-snapshot.deb
RUN dpkg -i package/linux/dist/0.0-snapshot/src.deb

# Copy config so we can create the config files on the host if we are
# host-mounting the config dir.
RUN cp -R /etc/sourcegraph /etc/sourcegraph.default

# Remove source dir (to avoid accidental source dependencies)
WORKDIR /
RUN rm -rf /usr/local/src/src.sourcegraph.com

USER sourcegraph
ENV GOPATH=$HOME
RUN src srclib toolchain install go
RUN src srclib toolchain install java

VOLUME ["/etc/sourcegraph", "/home/sourcegraph/.sourcegraph"]

EXPOSE 3080 3443

# Invoke src in a similar way to how the .deb would invoke it as a
# system service (but avoid the complexity of Docker+upstart/systemd
# by invoking src directly).
USER root
CMD chown -R sourcegraph:sourcegraph /etc/sourcegraph /home/sourcegraph/.sourcegraph && cp --no-clobber /etc/sourcegraph.default/* /etc/sourcegraph && su sourcegraph -c "source /etc/sourcegraph/config.env && src --config /etc/sourcegraph/config.ini serve"
