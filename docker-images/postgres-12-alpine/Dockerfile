# Are you bumping postgres minor or major version?
# Please review the changes in /usr/local/share/postgresql/postgresql.conf.sample
# If there is any change, you should ping @team/delivery
# And Delivery will make sure changes are reflected in our deploy repository
FROM postgres:12.7-alpine@sha256:b815f145ef6311e24e4bc4d165dad61b2d8e4587c96cea2944297419c5c93054

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

RUN apk add --upgrade --no-cache busybox>=1.34.1 &&\
    apk add --upgrade --no-cache libxslt>=1.1.35 &&\
    apk add --upgrade --no-cache libxml2>=2.9.12 &&\
    apk add --upgrade --no-cache libgcrypt>=1.8.8 &&\
    apk add --upgrade --no-cache apk-tools>=2.12.7

ENV POSTGRES_PASSWORD='' \
    POSTGRES_USER=sg \
    POSTGRES_DB=sg \
    PGDATA=/data/pgdata-12

COPY rootfs /
USER postgres
ENTRYPOINT ["/postgres.sh"]
