# -- NOTICE -- NOTICE -- NOTICE -- NOTICE -- NOTICE -- NOTICE -- NOTICE
#
# This image must be kept as light on startup as possible, it directly impacts
# the startup time of the VMs.
#
# -- NOTICE -- NOTICE -- NOTICE -- NOTICE -- NOTICE -- NOTICE -- NOTICE

# NOTE: This section is taken from https://github.com/weaveworks/ignite/blob/main/images/ubuntu/Dockerfile.
# We maintain this on our end, so we can determine when we want to pull new ubuntu versions, sice
# upstream ignite doesn't push new ones frequently.
FROM amd64/ubuntu:20.04@sha256:a06ae92523384c2cd182dcfe7f8b2bf09075062e937d5653d7d0db0375ad2221

# udev is needed for booting a "real" VM, setting up the ttyS0 console properly
# kmod is needed for modprobing modules
# systemd is needed for running as PID 1 as /sbin/init
# Also, other utilities are installed
# hadolint ignore=DL3008,DL3009,DL3015
RUN apt-get update && \
    apt-get install -y \
        ca-certificates \
        curl \
        gnupg \
        lsb-release \
        dbus \
        kmod \
        iproute2 \
        iputils-ping \
        net-tools \
        openssh-server \
        rng-tools \
        sudo \
        systemd \
        udev \
        vim-tiny \
        software-properties-common \
        wget && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN echo "" > /etc/machine-id && echo "" > /var/lib/dbus/machine-id

# This container image doesn't have locales installed. Disable forwarding the
# user locale env variables or we get warnings such as:
#  bash: warning: setlocale: LC_ALL: cannot change locale
RUN sed -i -e 's/^AcceptEnv LANG LC_\*$/#AcceptEnv LANG LC_*/' /etc/ssh/sshd_config

# Set the root password to root when logging in through the VM's ttyS0 console
# hadolint ignore=DL4006
RUN echo "root:root" | chpasswd

### END UPSTREAM SECTION

## Configures the serial TTY to auto login. This makes it easier to use ignite attach.
## The password is well-known anyways so this doesn't pose a security threat.
RUN mkdir -p "/etc/systemd/system/serial-getty@ttyS0.service.d"
COPY autologin.conf /etc/systemd/system/serial-getty@ttyS0.service.d/autologin.conf

# hadolint ignore=DL3008,DL3009
RUN add-apt-repository ppa:git-core/ppa && \
  apt-get update && \
  apt-get install -y --no-install-recommends git

# hadolint ignore=DL4006
RUN mkdir -p /etc/apt/keyrings && curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
# hadolint ignore=DL4006
RUN echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
# hadolint ignore=DL3008,DL3009
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        docker-ce \
        docker-ce-cli \
        containerd.io && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ARG SRC_CLI_VERSION
RUN set -ex && \
    curl -f -L -o src-cli.tar.gz "https://github.com/sourcegraph/src-cli/releases/download/${SRC_CLI_VERSION}/src-cli_${SRC_CLI_VERSION}_linux_amd64.tar.gz" && \
    tar -xvzf src-cli.tar.gz src && \
    mv src /usr/local/bin/src && \
    chmod +x /usr/local/bin/src && \
    rm -rf src-cli.tar.gz

RUN set -ex && \
  apt-get -y autoremove && \
  apt-get clean && \
  rm -rf /var/cache/* && \
  rm -rf /var/lib/apt/lists/*
