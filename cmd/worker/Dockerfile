FROM index.docker.io/sourcegraph/worker:3.41.1@sha256:a92fe878bfc1abecf77ea44927ba2781fa982cd3a93b9379fe95ba37d9c3b29f

# We need root to add the correct version of curl, but we don't want to
# accidentally change the default USER of the docker container while we're doing so.
# So, we:
#  1) check to make sure the current user is what we expect
#  2) switch to root and install the patched version of curl
#  3) switch back to the old user, double checking to make sure that we actually are running as the old user
ARG BASE_USER=sourcegraph

RUN if [[ "$(whoami)" != "${BASE_USER}" ]]; then echo "expected user to be '$BASE_USER', got '$(whoami)'" && exit 1; fi

USER root
RUN apk del curl && apk add --no-cache --verbose 'curl>=7.79.1-r2'

USER $BASE_USER
RUN if [[ "$(whoami)" != "${BASE_USER}" ]]; then echo "expected user to be '$BASE_USER', got '$(whoami)'" && exit 1; fi

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/worker"]
