FROM alpine:3.2

RUN apk add --update git docker mailcap && rm -rf /var/cache/apk/* # install mailcap for /etc/mime.types, needed by HTTP server
ENV GOPATH /usr/local

ENV STATIC_SITE_VERSION c79bd3f8c89141fd8e2be3328bf117a23e64f599
ENV STATIC_SITE_URL https://s3-us-west-2.amazonaws.com/sourcegraph-site/sourcegraph-site-$STATIC_SITE_VERSION.tar
RUN curl -Ls $STATIC_SITE_URL | tar x

# ADD wrapdocker /usr/local/bin/wrapdocker # currently not supported on Container Engine, add lxc package when using
ADD src /usr/local/bin/src

RUN for lang in "python" "go" "java"; do src toolchain get sourcegraph.com/sourcegraph/srclib-$lang; done

ENTRYPOINT ["src"]
CMD ["serve", "--http-addr=:3080"]

EXPOSE 3080 3443
