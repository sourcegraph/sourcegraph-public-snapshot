FROM amd64/ubuntu:19.04@sha256:d13c4ee6e14fea6b4e071d6276a9e2462c7312826b7761ecd32b3a1d24c6bd72
ENV DEBIAN_FRONTEND noninteractive
ENV LC_ALL C.UTF-8
ENV LANG C.UTF-8
RUN apt-get update && \
    apt-get install -y ca-certificates curl unzip xz-utils git openssh-client build-essential gpg && \
    rm -rf /var/lib/apt/lists/*
RUN groupadd --gid 1000 ubuntu && \
    useradd --uid 1000 --gid ubuntu --shell /bin/bash --create-home ubuntu
USER ubuntu
