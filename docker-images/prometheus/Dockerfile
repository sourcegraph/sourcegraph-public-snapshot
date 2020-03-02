# This Dockerfile is adapted from https://github.com/prometheus/prometheus/blob/v2.15.2/Dockerfile with
# our own customizations.
#
# When upgrading prometheus, be sure to check https://github.com/prometheus/prometheus/blob/master/Dockerfile
# (replacing master with the version number) to see if any changes (added/removed files, etc.) need to be
# synchronized with this derivative Dockerfile.
FROM prom/prometheus:v2.15.2@sha256:35c52c0c2b76433bbfc44a5a7abc294c2f032ed80250be02b441db5dd91b203a AS upstream

# hadolint ignore=DL3007
FROM quay.io/prometheus/busybox-linux-amd64:latest@sha256:0c38f63cbe19e40123668a48c36466ef72b195e723cbfcbe01e9657a5f14cec6

LABEL org.opencontainers.image.url=https://sourcegraph.com/
LABEL org.opencontainers.image.source=https://github.com/sourcegraph/sourcegraph/
LABEL org.opencontainers.image.documentation=https://docs.sourcegraph.com/
LABEL com.sourcegraph.prometheus.version=v2.15.2

COPY --from=upstream /bin/prometheus /bin/prometheus
COPY --from=upstream /bin/promtool /bin/promtoool
COPY --from=upstream /etc/prometheus/prometheus.yml /etc/prometheus/prometheus.yml
COPY --from=upstream /usr/share/prometheus/console_libraries/ /usr/share/prometheus/console_libraries/
COPY --from=upstream /usr/share/prometheus/consoles/ /usr/share/prometheus/consoles/
COPY --from=upstream /LICENSE /LICENSE
COPY --from=upstream /NOTICE /NOTICE
# hadolint ignore=DL3010
COPY --from=upstream /npm_licenses.tar.bz2 /npm_licenses.tar.bz2

RUN ln -s /usr/share/prometheus/console_libraries /usr/share/prometheus/consoles/ /etc/prometheus/

# TODO(uwe): remove "USER root" line once https://github.com/prometheus/prometheus/issues/3441 is resolved
#
# This is needed currently because the upstream image has us running as "nobody"
# which cannot create the sourcegraph user below.
USER root
# Add the sourcegraph group, user, and create the home directory.
#
# We use a static GID/UID assignment to ensure files can be chown'd to this
# user on the host machine (where this user does not exist).
# See https://github.com/sourcegraph/sourcegraph/issues/1884
#
# Note: This mirrors what we do in e.g. our base alpine image: https://github.com/sourcegraph/sourcegraph/blob/master/docker-images/alpine/Dockerfile#L10-L15
RUN addgroup -g 101 -S sourcegraph && adduser -u 100 -S -G sourcegraph -h /home/sourcegraph sourcegraph
RUN mkdir -p /prometheus && chown -R sourcegraph:sourcegraph /prometheus
USER sourcegraph

COPY config/*_rules.yml /sg_config_prometheus/
COPY config/prometheus.yml /sg_config_prometheus/

COPY entry.sh /

ENTRYPOINT ["/entry.sh"]
# Note that upstream's 'VOLUME' directive was deliberately removed. Including it makes it impossible
# to chmod the directory to our 'sourcegraph' user.
WORKDIR    /prometheus
EXPOSE     9090
