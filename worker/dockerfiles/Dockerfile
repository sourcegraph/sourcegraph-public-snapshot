# Docker image to generate a srclib binary for language-specific
# Docker images

FROM golang:1.6-alpine

# Install srclib from source
RUN apk --update add git make
RUN go get -d sourcegraph.com/sourcegraph/srclib/cmd/srclib
WORKDIR $GOPATH/src/sourcegraph.com/sourcegraph/srclib
ENV SRCLIB_VERSION 6538858f0c410cac5c63440317b8d009e889d3fb
RUN git checkout $SRCLIB_VERSION

# Generate static binary
RUN go get github.com/laher/goxc
RUN goxc -q

# Run environment
VOLUME /out
ENTRYPOINT mv ./release/snapshot/linux-amd64 /out/srclib
