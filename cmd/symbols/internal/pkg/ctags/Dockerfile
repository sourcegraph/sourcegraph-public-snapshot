FROM alpine:3.9@sha256:644fcb1a676b5165371437feaa922943aaf7afcfa8bfee4472f6860aad1ef2a0 AS ctags

# hadolint ignore=DL3003,DL3018,DL4006
RUN apk --no-cache add --virtual build-deps curl jansson-dev libseccomp-dev linux-headers autoconf pkgconfig make automake gcc g++ binutils && curl https://codeload.github.com/universal-ctags/ctags/tar.gz/08a69b5a030cf56c079e3006238b992b34c9cc51 | tar xz -C /tmp && cd /tmp/ctags-08a69b5a030cf56c079e3006238b992b34c9cc51 && ./autogen.sh && LDFLAGS=-static ./configure --program-prefix=universal- --enable-json --enable-seccomp && make -j8 && make install && cd && rm -rf /tmp/ctags-08a69b5a030cf56c079e3006238b992b34c9cc51 && apk --no-cache --purge del build-deps

WORKDIR /
COPY .ctags.d /.ctags.d
