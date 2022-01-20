FROM timescale/timescaledb:2.0.0-pg12-oss

# hadolint ignore=DL3017
RUN apk -U upgrade && apk add --no-cache \
    'busybox>=1.34.1'
    # Required due to ssl_clent upgrade with busybox
    'wget>=1.21.1'

