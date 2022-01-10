FROM timescale/timescaledb:2.0.0-pg12-oss

# hadolint ignore=DL3017
RUN apk -U upgrade && apk add --no-cache \
    # Requires the use of the edge repo for CVE-2021-28831 until a patched version is brought into the 3.12 mirror.
    'busybox>=1.34.1' --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main \
    # Required due to ssl_clent upgrade with busybox
    'wget>=1.21.1' --repository=http://dl-cdn.alpinelinux.org/alpine/v3.14/main

