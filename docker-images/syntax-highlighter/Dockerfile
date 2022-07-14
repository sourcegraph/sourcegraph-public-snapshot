FROM index.docker.io/sourcegraph/syntax-highlighter:3.41.1@sha256:ce2c417dd5a2e133f4b2fe2a252d2bac8ba40b253b6a892b60df9f437b28d22d

# We need root to add the correct version of curl, but we don't want to
# accidentally change the default USER of the docker container while we're doing so.
# So, we:
#  1) check to make sure the current user is what we expect
#  2) switch to root and install the patched version of curl
#  3) switch back to the old user, double checking to make sure that we actually are running as the old user

# this is odd, we always ran as root here?
ARG BASE_USER=root

RUN if [[ "$(whoami)" != "${BASE_USER}" ]]; then echo "expected user to be '$BASE_USER', got '$(whoami)'" && exit 1; fi

USER root
RUN apk del curl && apk add --no-cache --verbose 'curl>=7.79.1-r2'

USER $BASE_USER
RUN if [[ "$(whoami)" != "${BASE_USER}" ]]; then echo "expected user to be '$BASE_USER', got '$(whoami)'" && exit 1; fi

CMD ["sh", "-c", "/http-server-stabilizer -listen=:9238 -prometheus-app-name=syntect_server -workers=$WORKERS -- env ROCKET_PORT={{.Port}} /syntect_server"]
