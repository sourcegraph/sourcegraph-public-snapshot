FROM redis:5-alpine@sha256:72fa17785a9a81e7cff8b4ee2fa9c7a8ed4c27776c7ece1f3cda2dbc5f3e9c31

RUN mkdir -p /redis-data && chown -R redis:redis /redis-data

# @FIXME: Update redis image
# Pin busybox=1.33.1-r6 https://github.com/sourcegraph/sourcegraph/issues/27965

RUN apk --upgrade --no-cache add tini apk-tools>=2.12.7-r0 libssl1.1>=1.1.1n-r0 libcrypto1.1>=1.1.1n-r0 busybox>=1.33.1-r6

USER redis
COPY redis.conf /etc/redis/redis.conf

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["redis-server", "/etc/redis/redis.conf"]
