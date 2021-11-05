/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

brew install git
git clone https://github.com/sourcegraph/sourcegraph.git
git clone https://github.com/sourcegraph/dev-private.git

brew install --cask docker

brew install go
brew install node
brew install yarn

brew install gnu-sed
brew install comby
brew install pcre
brew install sqlite
brew install jq

brew install postgres redis
brew services start postgresql
brew services start redis

## sg setup-database
createdb --user=sourcegraph --owner=sourcegraph --host=localhost --encoding=UTF8 --template=template0 sourcegraph
createuser --superuser sourcegraph
psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"
createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph

## sg setup-caddy
./dev/add_https_domain_to_hosts.sh
./dev/caddy.sh trust



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
