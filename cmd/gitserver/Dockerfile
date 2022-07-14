FROM index.docker.io/sourcegraph/gitserver:3.41.1@sha256:d9f35585bdda0d20aa2bc79b8928b5b34d3ab1a404f4714fd36bb6f2a040e32d

# We need root to add the correct version of curl, but we don't want to
# accidentally change the default USER of the docker container while we're doing so.
# So, we:
#  1) check to make sure the current user is what we expect
#  2) switch to root and install the patched version of curl
#  3) switch back to the old user, double checking to make sure that we actually are running as the old user
ARG BASE_USER=sourcegraph

RUN if [[ "$(whoami)" != "${BASE_USER}" ]]; then echo "expected user to be '$BASE_USER', got '$(whoami)'" && exit 1; fi

USER root
RUN for pkg in curl libcurl; do if apk info --installed $pkg; then apk add --no-cache "$pkg>=7.79.1-r2"; fi; done

USER $BASE_USER
RUN if [[ "$(whoami)" != "${BASE_USER}" ]]; then echo "expected user to be '$BASE_USER', got '$(whoami)'" && exit 1; fi

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/gitserver"]
