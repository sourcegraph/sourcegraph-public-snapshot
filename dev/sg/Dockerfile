##
# EXPERIMENTAL Sourcegraph development kit Dockerfile
#
# Build it:
#
# $ docker build . -f dev/sg/Dockerfile -t sgk:latest
#
# Run it:
#
# $ docker run -d -v /var/run/docker.sock:/var/run/docker.sock -v /path/to/sourcegraph/:/opt/sourcegraph --name sgk sgk:latest

FROM ubuntu:latest

RUN apt-get update && apt-get install -y curl gnupg software-properties-common
RUN add-apt-repository ppa:longsleep/golang-backports

# Docker
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
RUN add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

# Yarn
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list

# Packages
RUN apt-get update && apt install -y make git-all libpcre3-dev libsqlite3-dev pkg-config musl-tools libnss3-tools \
                   golang-go docker-ce docker-ce-cli containerd.io yarn jq redis-server postgresql postgresql-contrib

# Configure database
# `sourcegraph` user needs to be a super-user to be able to create extensions
RUN /etc/init.d/postgresql start && \
  su postgres -c "createuser sourcegraph --superuser" && \
  su postgres -c "psql -c \"ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';\"" && \
  su postgres -c "createdb --owner=sourcegraph --encoding=UTF8 sourcegraph"

ENV PGPORT=5432
ENV PGHOST=localhost
ENV PGUSER=sourcegraph
ENV PGPASSWORD=sourcegraph
ENV PGDATABASE=sourcegraph
ENV PGSSLMODE=disable

# Install Node
COPY .nvmrc /opt/sourcegraph/.nvmrc

RUN curl -L https://raw.githubusercontent.com/nvm-sh/nvm/$(curl https://api.github.com/repos/nvm-sh/nvm/releases/latest | jq -r .name)/install.sh -o install-nvm.sh && \
  sh install-nvm.sh && rm install-nvm.sh && . /root/.nvm/nvm.sh && cd /opt/sourcegraph && nvm install

# Copy repository files
COPY . /opt/sourcegraph
WORKDIR /opt/sourcegraph

# Install sg
RUN echo 'alias sg="/root/go/bin/sg"' >> ~/.bashrc
ENV PATH="/root/go/bin:${PATH}"
RUN /opt/sourcegraph/dev/sg/install.sh

# Start databases
CMD ["/bin/bash", "-c", "/etc/init.d/redis-server start && /etc/init.d/postgresql start && sleep infinity"]
