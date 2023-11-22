# Are you bumping postgres minor or major version?
# Please review the changes in /usr/local/share/postgresql/postgresql.conf.sample
# If there is any change, you should ping @team/delivery
# And Delivery will make sure changes are reflected in our deploy repository
FROM postgres:12.11-alpine@sha256:41eb5461276b12ff88cb38a82f7d939f2acb0dfddcedb93d1f4d69418c51cc5d

ARG PING_UID=99
ARG POSTGRES_UID=999

# We modify the postgres user/group to reconcile with our previous debian based images
# and avoid issues with customers migrating.
RUN apk add --no-cache nss su-exec shadow &&\
    groupmod -g $PING_UID ping &&\
    usermod -u $POSTGRES_UID postgres &&\
    groupmod -g $POSTGRES_UID postgres &&\
    mkdir -p /data/pgdata-12 && chown -R postgres:postgres /data &&\
    chown -R postgres:postgres /var/lib/postgresql &&\
    chown -R postgres:postgres /var/run/postgresql

RUN apk add --upgrade --no-cache \
    'busybox>=1.35.0-r17' \
    'libxslt>=1.1.35' \
    'libxml2>=2.9.14-r2' \
    'libgcrypt>=1.8.8' \
    'apk-tools>=2.12.7' \
    'krb5-libs>=1.19.4-r0' \
    'ncurses-libs>=6.3_p20220521-r1'

ENV POSTGRES_PASSWORD='' \
    POSTGRES_USER=sg \
    POSTGRES_DB=sg \
    PGDATA=/data/pgdata-12

COPY rootfs /
USER postgres
ENTRYPOINT ["/postgres.sh"]
