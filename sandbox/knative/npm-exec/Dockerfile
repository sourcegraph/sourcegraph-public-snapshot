FROM sourcegraph/a8n-exec-server:latest as builder

FROM node:12.11.0-buster
USER root
ARG NPM_VERSION
RUN yarn global add npm@$NPM_VERSION
COPY --from=builder /work/server /server
USER node

CMD ["/server"]
