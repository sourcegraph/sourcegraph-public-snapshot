FROM alpine:3.4
MAINTAINER Sourcegraph Team <support@sourcegraph.com>

RUN apk add --update --update-cache \
  openssh \
  git \
  # for /etc/mime.types, needed by HTTP server
  mailcap \
  && rm -rf /var/cache/apk/*

RUN { \
    echo 'Host *'; \
    echo 'UseRoaming no'; \
  } >> /etc/ssh/ssh_config

# required for Git
ENV HOME "/root"

# ADD wrapdocker /usr/local/bin/wrapdocker # currently not supported on Container Engine, add lxc package when using
ADD src /usr/local/bin/src

ENTRYPOINT ["src"]
CMD ["serve", "--http-addr=:80", "--https-addr=:443"]

VOLUME ["/root/.sourcegraph", "/var/run/docker.sock"]
EXPOSE 80 443 22
