FROM sourcegraph/builder:3254493778@sha256:10a11bf74d0b8c97af0e738aca547bc28399c943255b411e0fe21b1ab18ca60d as builder

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
