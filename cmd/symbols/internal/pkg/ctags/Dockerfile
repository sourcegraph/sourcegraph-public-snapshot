FROM sourcegraph/builder:1be4c9ba43@sha256:0172c10814a1fc71adfe82c19c69bcf3f1496745c99e865e32836466b148778b as builder

ARG CTAGS_VERSION="03f933a96d3ef87adbf9d167462d45ce69577edb"
ENV CTAGS_VERSION ${CTAGS_VERSION}

ENV OUTPUT_DIR=/output
RUN mkdir -p ${OUTPUT_DIR}

ENV CTAGS_D_OUTPUT_PATH ${OUTPUT_DIR}/.ctags.d

WORKDIR /src
COPY . .
RUN internal/pkg/ctags/build.sh

FROM alpine:3.9@sha256:644fcb1a676b5165371437feaa922943aaf7afcfa8bfee4472f6860aad1ef2a0

COPY --from=builder /output/.ctags.d /.ctags.d
COPY --from=builder /output/bin/universal-* /usr/local/bin/
