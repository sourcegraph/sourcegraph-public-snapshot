FROM alpine:3.8
RUN echo -e "@edge http://dl-cdn.alpinelinux.org/alpine/edge/main\n" >> /etc/apk/repositories && \
  echo -e "@edge http://dl-cdn.alpinelinux.org/alpine/edge/community\n" >> /etc/apk/repositories
RUN echo -e "http://dl-cdn.alpinelinux.org/alpine/edge/main\n" >> /etc/apk/repositories && \
  echo -e "http://dl-cdn.alpinelinux.org/alpine/edge/community\n" >> /etc/apk/repositories
RUN echo -e "http://dl-cdn.alpinelinux.org/alpine/v3.6/main\n" >> /etc/apk/repositories && \
  echo -e "http://dl-cdn.alpinelinux.org/alpine/v3.6/community\n" >> /etc/apk/repositories
RUN apk add --no-cache 'postgresql-contrib=11.1-r0' 'postgresql=11.1-r0' 'redis=3.2.12-r0' bind-tools ca-certificates curl docker git@edge mailcap nginx openssh-client su-exec tini
RUN apk update && apk upgrade
RUN curl -o /usr/local/bin/syntect_server https://storage.googleapis.com/sourcegraph-artifacts/syntect_server/f85a9897d3c23ef84eb219516efdbb2d && chmod +x /usr/local/bin/syntect_server
RUN apk --no-cache add curl jansson-dev libseccomp-dev linux-headers autoconf pkgconfig make automake gcc g++ binutils && curl https://codeload.github.com/universal-ctags/ctags/tar.gz/7918d19fe358fae9bad1c264c4f5dc2dcde5cece | tar xz -C /tmp && cd /tmp/ctags-7918d19fe358fae9bad1c264c4f5dc2dcde5cece && ./autogen.sh && LDFLAGS=-static ./configure --program-prefix=universal- --enable-json --enable-seccomp && make -j8 && make install && cd && rm -rf /tmp/ctags-7918d19fe358fae9bad1c264c4f5dc2dcde5cece && apk --no-cache --purge del curl jansson-dev libseccomp-dev linux-headers autoconf pkgconfig make automake gcc g++ binutils
ENV GO111MODULES=on LANG=en_US.utf8 VERSION=$VERSION
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/server"]
ADD server /usr/local/bin/
ADD management-console /usr/local/bin/
ADD github-proxy /usr/local/bin/
ADD gitserver /usr/local/bin/
ADD query-runner /usr/local/bin/
ADD symbols /usr/local/bin/
ADD repo-updater /usr/local/bin/
ADD searcher /usr/local/bin/
ADD zoekt-archive-index /usr/local/bin/
ADD zoekt-sourcegraph-indexserver /usr/local/bin/
ADD zoekt-webserver /usr/local/bin/
ADD frontend /usr/local/bin/
