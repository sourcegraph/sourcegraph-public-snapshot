# We use weaves ubuntu image as the base. This image must be kept as light on startup as possible,
# it directly impacts the startup time of the VMs.
FROM weaveworks/ignite-ubuntu:20.04-amd64@sha256:4f5f5ed56fae650ae122daa28a785192dda081be4f0b37dca2eb25ea57840500

# Install latest git.
RUN set -ex && \
  apt-get update && \
  apt-get install -y software-properties-common && \
  add-apt-repository ppa:git-core/ppa && \
  apt-get update -y && \
  apt-get install -y git

# Install docker.
# hadolint ignore=DL3008,DL3009
RUN set -ex && \
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \
  add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" && \
  apt-get update -y && \
  apt-cache policy docker-ce && \
  apt-get install -y docker-ce docker-ce-cli containerd.io

ARG SRC_CLI_VERSION

# Install src-cli at the given version.
RUN set -ex && \
  curl -f -L -o src-cli.tar.gz "https://github.com/sourcegraph/src-cli/releases/download/${SRC_CLI_VERSION}/src-cli_${SRC_CLI_VERSION}_linux_amd64.tar.gz" && \
  tar -xvzf src-cli.tar.gz src && \
  mv src /usr/local/bin/src && \
  chmod +x /usr/local/bin/src && \
  rm -rf src-cli.tar.gz

# A little cleanup.
RUN set -ex && \
  apt-get -y autoremove && \
  apt-get clean && \
  rm -rf /var/cache/* && \
  rm -rf /var/lib/apt/lists/*
