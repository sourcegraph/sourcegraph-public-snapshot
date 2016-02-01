FROM alpine:3.2
MAINTAINER Sourcegraph Team <help@sourcegraph.com>

RUN apk add --update openssh git mailcap postgresql postgresql-contrib && rm -rf /var/cache/apk/* # install mailcap for /etc/mime.types, needed by HTTP server

RUN { \
		echo 'Host *'; \
		echo 'UseRoaming no'; \
	} >> /etc/ssh/ssh_config

ADD src /usr/local/bin/src
ADD src-init.sh /usr/local/bin/src-init.sh

ENTRYPOINT ["src-init.sh"]
CMD ["serve", "--http-addr=:80", "--https-addr=:443", "--ssh-addr=:22"]

VOLUME ["/root/.sourcegraph", "/var/run/docker.sock"]
EXPOSE 80 443 22
