FROM alpine:3.2
MAINTAINER Sourcegraph Team <help@sourcegraph.com>

RUN apk add --update openssh git mailcap postgresql postgresql-contrib && rm -rf /var/cache/apk/* # install mailcap for /etc/mime.types, needed by HTTP server

RUN { \
		echo 'Host *'; \
		echo 'UseRoaming no'; \
	} >> /etc/ssh/ssh_config

# ADD wrapdocker /usr/local/bin/wrapdocker # currently not supported on Container Engine, add lxc package when using
ADD src /usr/local/bin/src

ENTRYPOINT ["src"]
CMD ["serve", "--http-addr=:80", "--https-addr=:443", "--ssh-addr=:22"]

VOLUME ["/root/.sourcegraph", "/var/run/docker.sock"]
EXPOSE 80 443 22
