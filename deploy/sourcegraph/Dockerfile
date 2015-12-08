FROM sourcegraph/srclib-base

ENV STATIC_SITE_VERSION 64df72271be077a0a9542d24dd40e22c68aa494d
ENV STATIC_SITE_URL https://s3-us-west-2.amazonaws.com/sourcegraph-site/sourcegraph-site-$STATIC_SITE_VERSION.tar
RUN curl -Ls $STATIC_SITE_URL | tar x

# ADD wrapdocker /usr/local/bin/wrapdocker # currently not supported on Container Engine, add lxc package when using
ADD src /usr/local/bin/src

ENTRYPOINT ["src"]
CMD ["serve", "--http-addr=:3080"]

EXPOSE 3080 3443
