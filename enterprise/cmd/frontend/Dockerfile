FROM index.docker.io/sourcegraph/frontend:3.41.1@sha256:b3d59106f3490d743c27cb3822416b4976a433d0c99bb85ab7babfa2fea1e97f

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
RUN if apk info --installed libcrypto1.1; then apk add --no-cache 'libcrypto1.1>=1.1.1q-r0' 'libssl1.1>=1.1.1q-r0'; fi
USER $BASE_USER
RUN if [[ "$(whoami)" != "${BASE_USER}" ]]; then echo "expected user to be '$BASE_USER', got '$(whoami)'" && exit 1; fi
