# Dockerfile to run additional initailization after Sourcegraph Origin starts up.
#
# Environment expectations:
# - expects local repositories to be mounted at $GIT_PARENT_DIRECTORY
# - expects PostgreSQL environment variables to be set and the appropriate Postgres server instance to be available at the specified coordinates

FROM alpine:3.5
RUN apk update
RUN apk add curl bash postgresql-client findutils coreutils

ADD *.sh /scripts/
WORKDIR /scripts

# Run the init script and then stick around in case we want to run additional scripts.
CMD "/bin/bash" "-c" "./init.sh; while true; do sleep 86400; done;"
