FROM alpine:3.2
MAINTAINER Sourcegraph Team <help@sourcegraph.com>

RUN apk add --update git mailcap && rm -rf /var/cache/apk/* # install mailcap for /etc/mime.types, needed by HTTP server

ENV STATIC_SITE_VERSION 7fc2fba9d768b29f6768175ad255415137a9b5db
ENV STATIC_SITE_URL https://s3-us-west-2.amazonaws.com/sourcegraph-site/sourcegraph-site-$STATIC_SITE_VERSION.tar
RUN curl -Ls $STATIC_SITE_URL | tar x

# ADD wrapdocker /usr/local/bin/wrapdocker # currently not supported on Container Engine, add lxc package when using
ADD src /usr/local/bin/src

ENTRYPOINT ["src"]
CMD ["serve", "--http-addr=:80", "--https-addr=:443", "--ssh-addr=:22"]

VOLUME ["/root/.sourcegraph", "/var/run/docker.sock"]
EXPOSE 80 443 22
