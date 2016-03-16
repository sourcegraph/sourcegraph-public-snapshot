FROM golang:1.6

RUN apt-get -qy update
RUN apt-get -qy install postgresql postgresql-contrib docker make nodejs npm
RUN ln -s /usr/bin/nodejs /usr/bin/node

RUN curl -fsSL https://get.docker.com/ | sh

# Set up PostgreSQL environment
ENV PGDATABASE=postgres
ENV PGHOST=localhost
ENV PGPORT=5432
ENV PGUSER=postgres
ENV PGSSLMODE=disable

RUN chown -R postgres /var/lib/postgresql
USER postgres
WORKDIR /var/lib/postgresql
RUN /usr/lib/postgresql/9.4/bin/pg_ctl -D db init
USER root

# Fetch Sourcegraph source
RUN go get -d sourcegraph.com/sourcegraph/sourcegraph
WORKDIR /go/src/sourcegraph.com/sourcegraph/sourcegraph
RUN make dep
RUN go get github.com/tools/godep
RUN godep go install ./cmd/src

ENV SRC_SKIP_REQS="Docker"
EXPOSE 3080

RUN su -c '/usr/lib/postgresql/9.4/bin/pg_ctl -D /var/lib/postgresql/db start' postgres && src pgsql reset && su -c '/usr/lib/postgresql/9.4/bin/pg_ctl -D /var/lib/postgresql/db stop' postgres

RUN echo "su -c '/usr/lib/postgresql/9.4/bin/pg_ctl -D /var/lib/postgresql/db -l /var/lib/postgresql/logfile start' -s /bin/bash postgres" >> /etc/bash.bashrc
