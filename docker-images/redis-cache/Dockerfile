FROM redis:5-alpine@sha256:1c465cac273d3e0a63ca9896cb78eb75bad9c80e9a0b3ea0b5a0c6d1f12c3232

RUN mkdir -p /redis-data && chown -R redis:redis /redis-data

# @FIXME: Update redis image
# Pin busybox=1.33.1-r6 https://github.com/sourcegraph/sourcegraph/issues/27965

RUN apk --upgrade --no-cache add tini apk-tools>=2.12.7-r0 libssl1.1>=1.1.1n-r0 libcrypto1.1>=1.1.1n-r0 busybox>=1.33.1-r6


USER redis
COPY redis.conf /etc/redis/redis.conf

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["redis-server", "/etc/redis/redis.conf"]
