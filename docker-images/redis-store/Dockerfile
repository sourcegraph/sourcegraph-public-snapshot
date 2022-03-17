FROM redis:5-alpine@sha256:fea243676a4d2d67f5990ddcbd4a56db9423b7f25e55758491e39988efc1cfbe

RUN mkdir -p /redis-data && chown -R redis:redis /redis-data

# @FIXME: Update redis image
# Pin busybox=1.33.1-r6 https://github.com/sourcegraph/sourcegraph/issues/27965

RUN apk --upgrade --no-cache add tini apk-tools>=2.12.7-r0 libssl1.1>=1.1.1n-r0 libcrypto1.1>=1.1.1n-r0 busybox>=1.33.1-r6

USER redis
COPY redis.conf /etc/redis/redis.conf

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["redis-server", "/etc/redis/redis.conf"]
