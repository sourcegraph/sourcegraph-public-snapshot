FROM sourcegraph/a8n-exec-server:latest as builder

FROM node:12.11.0-buster
USER root
ARG YARN_VERSION
RUN yarn global add yarn@$YARN_VERSION
COPY --from=builder /work/server /server
RUN chmod +x /server
USER node

CMD ["/server"]
