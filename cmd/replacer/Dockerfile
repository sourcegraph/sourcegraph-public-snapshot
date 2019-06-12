FROM sourcegraph/alpine:3.9

# hadolint ignore=DL3018
RUN apk --no-cache add pcre-dev

# hadolint ignore=DL3022
COPY --from=comby/comby:0.5.0 /usr/local/bin/comby /usr/local/bin/comby

EXPOSE 3185
USER sourcegraph
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/replacer"]
COPY replacer /usr/local/bin/
