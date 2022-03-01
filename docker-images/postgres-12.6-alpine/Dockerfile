FROM postgres:12.6-alpine@sha256:b5c57d326a4cb8cf2fd5cdedc758ed72146a69304ae434b177088b94a7c9e6b8

# We modify the postgres user/group to reconcile with our previous debian based images
# and avoid issues with customers migrating.
RUN apk add --no-cache nss su-exec shadow &&\
    groupmod -g 99 ping &&\
    usermod -u 999 postgres &&\
    groupmod -g 999 postgres &&\
    mkdir -p /data/pgdata-12 && chown -R postgres:postgres /data &&\
    chown -R postgres:postgres /var/lib/postgresql &&\
    chown -R postgres:postgres /var/run/postgresql

# @FIXME: Update redis image
# Pin busybox=1.32.1-r7 https://github.com/sourcegraph/sourcegraph/issues/27965
RUN apk add --upgrade --no-cache libxml2>=2.9.12 libgcrypt>=1.8.8 apk-tools>=2.12.7 busybox>=1.32.1

ENV POSTGRES_PASSWORD='' \
    POSTGRES_USER=sg \
    POSTGRES_DB=sg \
    PGDATA=/data/pgdata-12

COPY rootfs /
USER postgres
ENTRYPOINT ["/postgres.sh"]

