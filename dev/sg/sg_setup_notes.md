## homebrew

/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

## base depencencies
brew install git
brew install gnu-sed
brew install comby
brew install pcre
brew install sqlite
brew install jq

## clone repositories
git clone https://github.com/sourcegraph/sourcegraph.git
git clone https://github.com/sourcegraph/dev-private.git

## docker

brew install --cask docker

## Postgres

brew install postgres
brew services start postgresql

## Redis

brew install postgres redis
brew services start redis

## Programming languages

brew install go
brew install node
brew install yarn

## sg setup-database

```
createdb --user=sourcegraph --owner=sourcegraph --host=localhost --encoding=UTF8 --template=template0 sourcegraph
createuser --superuser sourcegraph
psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"
createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph
```

## Proxy for local development

```
./dev/add_https_domain_to_hosts.sh
./dev/caddy.sh trust
```

## linux

sudo apt install -y
  make
  git-all
  libpcre3-dev
  libsqlite3-dev
  pkg-config
  golang-go
  docker-ce
  docker-ce-cli
  containerd.io
  yarn
  jq
  libnss3-tools
