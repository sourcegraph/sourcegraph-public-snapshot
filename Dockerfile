FROM ubuntu:15.04

RUN apt-get update -qq \
    && apt-get install -qy \
            curl \
            git \
            make \
            nodejs \
            npm \
    && ln -rs /usr/bin/nodejs /usr/bin/node

# Install Protobuf v3.0.0-alpha-3, see .travis.yml for information on rebuilding.
RUN curl -O https://s3-us-west-2.amazonaws.com/public-dev/protobuf-bin-v3.0.0-alpha-3.tar.gz
RUN tar zxf protobuf-bin-v3.0.0-alpha-3.tar.gz -C /

# Install Go
RUN curl -Ls https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH=/usr/local/go/bin:/usr/local/protobuf/bin:$PATH \
    GOBIN=/usr/local/bin \
    SGPATH=/var/lib/sourcegraph \
    GOPATH=/sg

# Add an example repo
RUN mkdir -p $SGPATH/repos/github.com/gorilla/mux \
    && git clone https://github.com/gorilla/mux $SGPATH/repos/github.com/gorilla/mux

# Install src
ADD . /sg/src/src.sourcegraph.com/sourcegraph
WORKDIR /sg/src/src.sourcegraph.com/sourcegraph
RUN make dist PACKAGEFLAGS="--os linux" && mv release/snapshot/linux-amd64 $GOBIN/src

# Trust GitHub's SSH host key (for ssh cloning of repos during builds)
RUN install -Dm 600 package/etc/known_hosts /root/.ssh/known_hosts \
    && chmod 700 /root/.ssh

# Remove source dir (to avoid accidental source dependencies)
WORKDIR /
RUN rm -rf /sg

# srclib - Previous version broken, so only installing go for now
RUN /usr/local/bin/src srclib toolchain install go

EXPOSE 3000 3001 3100
CMD ["-v", "serve", "--http-addr=:3000", "--grpc-addr=:3100", "--addr=:3001"]
ENTRYPOINT ["/usr/local/bin/src"]
