FROM redis:5-alpine@sha256:ed670ae81585e6bc3d4847fde5de862ac711cdd9e4c3b5fc3acc9f72c4ce5030

RUN mkdir -p /redis-data && chown -R redis:redis /redis-data

# @FIXME: Update redis image
# Pin busybox=1.33.1-r6 https://github.com/sourcegraph/sourcegraph/issues/27965

RUN apk --upgrade --no-cache add tini apk-tools>=2.12.7-r0 libssl1.1>=1.1.1n-r0 libcrypto1.1>=1.1.1n-r0 busybox>=1.33.1-r6


USER redis
COPY redis.conf /etc/redis/redis.conf

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["redis-server", "/etc/redis/redis.conf"]
