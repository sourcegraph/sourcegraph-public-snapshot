FROM postgres:12.6-alpine@sha256:2b21f2d617bbd930f60799519321fb630b13b494c38387f87520b3f21998e913

RUN apk add --no-cache nss &&\
    mkdir -p /data/pgdata-12 && chown postgres:postgres /data

ENV POSTGRES_PASSWORD=sg \
    POSTGRES_USER=sg \
    POSTGRES_DB=sg \
    PGDATA=/data/pgdata-12

# COPY rootfs /
# USER postgres
# ENTRYPOINT ["/postgres.sh"]
