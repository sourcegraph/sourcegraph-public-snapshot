FROM ubuntu:15.04

RUN apt-get update -qq \
    && apt-get install -qy \
            curl \
            git \
            make \
            nodejs \
            npm \
    && ln -rs /usr/bin/nodejs /usr/bin/node

# Install Go
RUN curl -Ls https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH=/usr/local/go/bin:$PATH \
    GOBIN=/usr/local/bin \
    GOPATH=/go \
    SGPATH=/var/lib/sourcegraph

# Install protoc
RUN curl -Ls https://s3-us-west-2.amazonaws.com/public-dev/protobuf-bin-v3.0.0-alpha-3.tar.gz | tar zx -C /

# Pre-emptively fetch some dependencies of creating the release, so we don't
# need to refetch each time we run the image
RUN go get \
    github.com/tools/godep \
    sourcegraph.com/sourcegraph/prototools/cmd/protoc-gen-dump

# Setup cross compilation
RUN make dist

# We will mount the src into this directory, rather than including it as part
# of the image
RUN mkdir -p /go/src/sourcegraph.com/sourcegraph
VOLUME /sourcegraph
WORKDIR /go/src/sourcegraph.com/sourcegraph

# Setup our wrapper script which ensures we have a clean copy of sourcegraph
# at /go/src/sourcegraph.com/sourcegraph/ as well as copying the artifacts
# back into the volume
ADD pristine_checkout_wrapper.sh /
ENTRYPOINT ["/pristine_checkout_wrapper.sh"]
