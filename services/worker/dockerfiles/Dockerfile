# Docker image to generate a srclib binary for language-specific
# Docker images

FROM golang:1.6-alpine

# Install srclib from source
RUN apk --update add git make

ARG URL
ADD $URL /toolchain/t
RUN (cd /toolchain && tar xfz t && rm t && mv * t) || true
RUN mkdir -p $GOPATH/src/sourcegraph.com/sourcegraph && mv /toolchain/t $GOPATH/src/sourcegraph.com/sourcegraph/srclib
WORKDIR $GOPATH/src/sourcegraph.com/sourcegraph/srclib

# Generate static binary
RUN go get github.com/laher/goxc
RUN make govendor
RUN goxc -q

# Run environment
VOLUME /out
ENTRYPOINT mv ./release/snapshot/linux-amd64 /out/srclib
