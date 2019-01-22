# We split deps and builder to speed up builds where dependencies don't
# change, since they'll be cached.
FROM golang:1.11-alpine3.8 as build
# hadolint ignore=DL3018
RUN apk update && apk add --no-cache git
WORKDIR /go/src/github.com/sourcegraph/sourcegraph
ENV GO111MODULE=on
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# At this step we are still missing generating production assets, and
# embedding them into go files. For now we rely on pre-build.sh to do those
# steps in CI. Once we port that over to Docker, we can just rely on docker
# build and not use any build scripts.

# Enterprise build uses its own version of frontend and includes more images.
ARG ADDITIONAL_IMAGES=github.com/sourcegraph/sourcegraph/cmd/frontend
# Enterprise build uses its own version of server
ARG SERVER_PKG=github.com/sourcegraph/sourcegraph/cmd/server
ARG VERSION
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go install \
    -buildmode "exe" \
    -a -tags "dist netgo" \
    -ldflags '-w -extldflags "-static" -X github.com/sourcegraph/sourcegraph/pkg/version.version='$VERSION \
    $SERVER_PKG \
    github.com/sourcegraph/sourcegraph/cmd/management-console \
    github.com/sourcegraph/sourcegraph/cmd/github-proxy \
    github.com/sourcegraph/sourcegraph/cmd/gitserver \
    github.com/sourcegraph/sourcegraph/cmd/query-runner \
    github.com/sourcegraph/sourcegraph/cmd/symbols \
    github.com/sourcegraph/sourcegraph/cmd/repo-updater \
    github.com/sourcegraph/sourcegraph/cmd/searcher \
    github.com/google/zoekt/cmd/zoekt-archive-index \
    github.com/google/zoekt/cmd/zoekt-sourcegraph-indexserver \
    github.com/google/zoekt/cmd/zoekt-webserver $ADDITIONAL_IMAGES

# This Dockerfile was generated from github.com/sourcegraph/godockerize. It
# was not written by a human, and as such looks janky. As you change this
# file, please don't be scared to make it more pleasant / remove hadolint
# ignores.
FROM alpine:3.8
RUN echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
    echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.6/main" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.6/community" >> /etc/apk/repositories

# hadolint ignore=DL3018
RUN apk add --no-cache 'postgresql-contrib=11.1-r0' 'postgresql=11.1-r0' 'redis=3.2.12-r0' bind-tools ca-certificates curl docker git@edge mailcap nginx openssh-client su-exec tini
# hadolint ignore=DL3017
RUN apk update && apk upgrade
RUN curl -o /usr/local/bin/syntect_server https://storage.googleapis.com/sourcegraph-artifacts/syntect_server/f85a9897d3c23ef84eb219516efdbb2d && chmod +x /usr/local/bin/syntect_server
# hadolint ignore=DL3003,DL3018,DL4006
RUN apk --no-cache add curl jansson-dev libseccomp-dev linux-headers autoconf pkgconfig make automake gcc g++ binutils && curl https://codeload.github.com/universal-ctags/ctags/tar.gz/7918d19fe358fae9bad1c264c4f5dc2dcde5cece | tar xz -C /tmp && cd /tmp/ctags-7918d19fe358fae9bad1c264c4f5dc2dcde5cece && ./autogen.sh && LDFLAGS=-static ./configure --program-prefix=universal- --enable-json --enable-seccomp && make -j8 && make install && cd && rm -rf /tmp/ctags-7918d19fe358fae9bad1c264c4f5dc2dcde5cece && apk --no-cache --purge del curl jansson-dev libseccomp-dev linux-headers autoconf pkgconfig make automake gcc g++ binutils
ARG VERSION
ENV LANG=en_US.utf8 VERSION=$VERSION
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/server"]
COPY --from=build /go/bin/* /usr/local/bin/
