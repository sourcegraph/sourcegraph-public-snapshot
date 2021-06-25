FROM postgres:12.6-alpine@sha256:faf54a67a58ecf4d21e3e68b9a3b33f81a11e8a786530044c4ee80d51b148e05

RUN apk add --no-cache nss su-exec &&\
    mkdir -p /data/pgdata-12 && chown postgres:postgres /data

ENV POSTGRES_PASSWORD=sg \
    POSTGRES_USER=sg \
    POSTGRES_DB=sg \
    PGDATA=/data/pgdata-12

COPY rootfs /
USER postgres
ENTRYPOINT ["/postgres.sh"]
