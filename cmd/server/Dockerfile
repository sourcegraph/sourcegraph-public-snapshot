FROM alpine:3.9 AS ctags

# hadolint ignore=DL3003,DL3018,DL4006
RUN apk --no-cache add --virtual build-deps curl jansson-dev libseccomp-dev linux-headers autoconf pkgconfig make automake gcc g++ binutils && curl https://codeload.github.com/universal-ctags/ctags/tar.gz/08a69b5a030cf56c079e3006238b992b34c9cc51 | tar xz -C /tmp && cd /tmp/ctags-08a69b5a030cf56c079e3006238b992b34c9cc51 && ./autogen.sh && LDFLAGS=-static ./configure --program-prefix=universal- --enable-json --enable-seccomp && make -j8 && make install && cd && rm -rf /tmp/ctags-08a69b5a030cf56c079e3006238b992b34c9cc51 && apk --no-cache --purge del build-deps

FROM alpine:3.9
RUN echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
    echo "@edge http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.6/main" >> /etc/apk/repositories && \
    echo "http://dl-cdn.alpinelinux.org/alpine/v3.6/community" >> /etc/apk/repositories

# hadolint ignore=DL3018
RUN apk add --no-cache 'postgresql-contrib=9.6.10-r0' 'postgresql=9.6.10-r0' 'redis=3.2.12-r0' bind-tools ca-certificates git@edge mailcap nginx openssh-client su-exec tini

# hadolint ignore=DL3022
COPY --from=sourcegraph/syntect_server:d74791c /syntect_server /usr/local/bin/
COPY --from=ctags /usr/local/bin/universal-* /usr/local/bin/

ENV GO111MODULES=on LANG=en_US.utf8 VERSION=$VERSION
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/server"]
COPY * /usr/local/bin/
