FROM docker:23.0.1-dind@sha256:ed6220b0de0f309f0844cf8cf1a6b861e981fb7f5c28bec6acc97abc910bd0a8

RUN rm -f /usr/libexec/docker/cli-plugins/docker-compose && \
    rm -f /usr/libexec/docker/cli-plugins/docker-buildx

ARG COMMIT_SHA="unknown"
ARG DATE="unknown"
ARG VERSION="unknown"

VOLUME /var/lib/docker
EXPOSE 2375 2376

ENTRYPOINT ["dockerd-entrypoint.sh"]
CMD []
