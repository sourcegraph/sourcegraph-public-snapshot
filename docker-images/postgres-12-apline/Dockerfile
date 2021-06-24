FROM postgres:12-alpine


RUN apk add nss

RUN mkdir -p /data/pgdata-12 && chown postgres:postgres /data


ENV POSTGRES_PASSWORD=sg \
    POSTGRES_USER=sg \
    POSTGRES_DB=sg \
    PGDATA=/data/pgdata-12

# COPY rootfs /
# USER postgres
# ENTRYPOINT ["/postgres.sh"]
