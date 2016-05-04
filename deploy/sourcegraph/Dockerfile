FROM alpine:3.2
MAINTAINER Sourcegraph Team <support@sourcegraph.com>

RUN echo '@edge http://nl.alpinelinux.org/alpine/edge/main' >> /etc/apk/repositories

RUN apk add --update --update-cache \
  openssh \
  # Need edge for >2.7.1 See CVE-2016-2324, CVE-2016-2315
  git@edge \
  # for /etc/mime.types, needed by HTTP server
  mailcap \
  && rm -rf /var/cache/apk/*

RUN { \
    echo 'Host *'; \
    echo 'UseRoaming no'; \
  } >> /etc/ssh/ssh_config

# required for Git
ENV HOME "/root"

ENV STATIC_SITE_VERSION e7f2bcb297e66776b5019235edb7e2d3bd38f939
ENV STATIC_SITE_URL https://s3-us-west-2.amazonaws.com/sourcegraph-site/sourcegraph-site-$STATIC_SITE_VERSION.tar
RUN curl -Ls $STATIC_SITE_URL | tar x

# ADD wrapdocker /usr/local/bin/wrapdocker # currently not supported on Container Engine, add lxc package when using
ADD src /usr/local/bin/src

ENTRYPOINT ["src"]
CMD ["serve", "--http-addr=:80", "--https-addr=:443"]

VOLUME ["/root/.sourcegraph", "/var/run/docker.sock"]
EXPOSE 80 443 22
