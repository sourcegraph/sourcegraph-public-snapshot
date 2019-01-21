# This Dockerfile was generate from github.com/sourcegraph/godockerize. It was
# not written by a human, and as such looks janky. As you change this file,
# please don't be scared to make it more pleasant / remove hadolint ingores.

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
ENV GO111MODULES=on LANG=en_US.utf8 VERSION=$VERSION
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/server"]
COPY server /usr/local/bin/
COPY management-console /usr/local/bin/
COPY github-proxy /usr/local/bin/
COPY gitserver /usr/local/bin/
COPY query-runner /usr/local/bin/
COPY symbols /usr/local/bin/
COPY repo-updater /usr/local/bin/
COPY searcher /usr/local/bin/
COPY zoekt-archive-index /usr/local/bin/
COPY zoekt-sourcegraph-indexserver /usr/local/bin/
COPY zoekt-webserver /usr/local/bin/
COPY frontend /usr/local/bin/
