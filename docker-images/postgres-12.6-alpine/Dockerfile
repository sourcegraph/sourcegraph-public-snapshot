FROM postgres:12.6-alpine@sha256:faf54a67a58ecf4d21e3e68b9a3b33f81a11e8a786530044c4ee80d51b148e05

RUN apk add --no-cache nss su-exec shadow &&\
    groupmod -g 99 ping &&\
    usermod -u 999 postgres &&\
    groupmod -g 999 postgres &&\
    mkdir -p /data/pgdata-12 && chown -R postgres:postgres /data &&\
    chown -R postgres:postgres /var/lib/postgresql &&\
    chown -R postgres:postgres /var/run/postgresql

ENV POSTGRES_PASSWORD='' \
    POSTGRES_USER=sg \
    POSTGRES_DB=sg \
    PGDATA=/data/pgdata-12

COPY rootfs /

USER postgres

ENTRYPOINT ["/postgres.sh"]

