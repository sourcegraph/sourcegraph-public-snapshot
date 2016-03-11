FROM alpine:3.2
MAINTAINER Sourcegraph Team <support@sourcegraph.com>

RUN apk add --update openssh git mailcap && rm -rf /var/cache/apk/* # install mailcap for /etc/mime.types, needed by HTTP server

RUN { \
		echo 'Host *'; \
		echo 'UseRoaming no'; \
	} >> /etc/ssh/ssh_config

ENV STATIC_SITE_VERSION 9fc7a38cafbcfcdd48642c4c6c30c30e29bcaf63
ENV STATIC_SITE_URL https://s3-us-west-2.amazonaws.com/sourcegraph-site/sourcegraph-site-$STATIC_SITE_VERSION.tar
RUN curl -Ls $STATIC_SITE_URL | tar x

# ADD wrapdocker /usr/local/bin/wrapdocker # currently not supported on Container Engine, add lxc package when using
ADD src /usr/local/bin/src

ENTRYPOINT ["src"]
CMD ["serve", "--http-addr=:80", "--https-addr=:443"]

VOLUME ["/root/.sourcegraph", "/var/run/docker.sock"]
EXPOSE 80 443 22
