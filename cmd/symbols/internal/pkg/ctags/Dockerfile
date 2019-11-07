FROM sourcegraph/builder:13f3b144f4@sha256:d3cab7dd1d78fb2bc6d109ba364456194a26959122733d8b2a126f10ff384a80 as builder

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
