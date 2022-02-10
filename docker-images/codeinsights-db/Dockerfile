# This updates to Alpine 3.15.0
# Is needed to avoid CVE-2021-38561
# See https://www.whitesourcesoftware.com/vulnerability-database/CVE-2021-38561 for more details
FROM timescale/timescaledb:2.5.0-pg12-oss

# hadolint ignore=DL3017
RUN apk -U upgrade && apk add --no-cache \
    busybox \
    wget

