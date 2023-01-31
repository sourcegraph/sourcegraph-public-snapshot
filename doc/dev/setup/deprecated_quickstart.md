# DEPRECATED Quickstart without `sg setup`

The following instructions are from our old quickstart guide before we had `sg setup` guiding new users through the setup process.

This guide is kept here until we're sure that `sg setup` is stable.

## Install dependencies

Sourcegraph has the following dependencies:

- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) (v2.18 or higher)
- [Go](https://golang.org/doc/install) (see current version in [`.tool-versions`](https://github.com/sourcegraph/sourcegraph/blob/main/.tool-versions))
- [Node JS](https://nodejs.org/en/download/) (see current recommended version in [.nvmrc](https://github.com/sourcegraph/sourcegraph/blob/main/.nvmrc))
- [make](https://www.gnu.org/software/make/)
- [Docker](https://docs.docker.com/engine/installation/) (v18 or higher)
  - For macOS we recommend using Docker for Mac instead of `docker-machine`
- [PostgreSQL](https://wiki.postgresql.org/wiki/Detailed_installation_guides) (v12 or higher)
- [Redis](http://redis.io/) (v5.0.7 or higher)
- [Yarn](https://yarnpkg.com) (v1.10.1 or higher)
- [SQLite](https://www.sqlite.org/index.html) tools
- [Comby](https://github.com/comby-tools/comby/) (v0.11.3 or higher)

Below are instructions to install these dependencies:

- [macOS](#macos)
- [Ubuntu](#ubuntu)
- Optional: [asdf](#optional-asdf) for an alternate way of managing dependencies, especially different versions of programming languages.

> NOTE: You can choose to install and run Redis and/or PostgreSQL **with or without Docker**. The following instructions will describe both options.
> What's the better option?
>
> - Running within a container provides some advantages such as storing the data separately from the container, you do not need to run it as a system service and its easy to use different database versions or multiple databases.
> - Running as system services might yield better performance, especially on macOS.
> - No matter which option you choose, docker is required because the development server starts additional docker containers.

### macOS

1.  Install [Homebrew](https://brew.sh).
2.  Install [Docker for Mac](https://docs.docker.com/docker-for-mac/).

    Alternatively, you can install it via `brew`

    ```
    brew install --cask docker
    ```

3.  Install Go, Yarn, Git, Comby, SQLite tools, and jq with the following command:

    ```
    brew install go yarn git gnu-sed comby pcre sqlite jq
    ```

4. Choose to run Postgres and Redis manually (Option a.) or via Docker (Option b.)

    a. **Install PostgreSQL and Redis Without Docker**

      1. Install PostgreSQL and Redis with the following commands:

          ```
          brew install postgresql
          brew install redis
          ```

      2. (optional) Start the services (and configure them to start automatically):

          ```
          brew services start postgresql
          brew services start redis
          ```

          (You can stop them later by calling `stop` instead of `start` above.

      3. Ensure `psql`, the PostgreSQL command line client, is on your `$PATH`.

          ```
          hash psql
          ```
          
          If this command emits no output, you are free to move to the next step. Otherwise, you installed a Homebrew recipe that does not modify your `$PATH` by default. Homebrew gives you the command to run to insert `psql` in your path in the "Caveats" section of `brew info postgresql`.
          Alternatively, you can use the command below. It might need to be adjusted depending on your Homebrew prefix (`/usr/local` below) and shell (bash below).

          ```
          hash psql || { echo 'export PATH="/usr/local/opt/postgresql/bin:$PATH"' >> ~/.bash_profile }
          source ~/.bash_profile
          ```

      4. Open a new Terminal window to ensure `psql` is now on your `$PATH`.


    b. **With Docker**

    Nothing to do here, since you already installed Docker for Mac.

    We provide a docker compose file at `dev/redis-postgres.yml` to make it easy to run Redis and PostgreSQL as Docker containers, with [docker compose](https://docs.docker.com/compose/install/).


6.  Install the Node Version Manager (`nvm`) using:

    ```
    NVM_VERSION="$(curl https://api.github.com/repos/nvm-sh/nvm/releases/latest | jq -r .name)"
    curl -L https://raw.githubusercontent.com/nvm-sh/nvm/"$NVM_VERSION"/install.sh -o install-nvm.sh
    sh install-nvm.sh
    ```

    After the install script is finished, re-source your shell profile (e.g.,
    `source ~/.zshrc`) or restart your terminal session to pick up the `nvm`
    definitions. Re-running the install script will update the installation.

    > NOTE: `nvm` is implemented as a shell function, so it may not show up in
    the output of `which nvm`. Use `type nvm` to verify whether it is set up.
    There is also a Homebrew package for `nvm`, but it is unsupported by the
    `nvm` maintainers.

    * For fish shell users, you will want to install `bass` which you can get via `omf`:

        ```
        curl -L https://get.oh-my.fish | fish
        omf install bass
        ```

    * Then add the following to your `config.fish`:

        ```sh
        function nvm
          bass source ~/.nvm/nvm.sh --no-use ';' nvm $argv
        end

        set -x NVM_DIR ~/.nvm
        ```

7.  Install the current recommended version of Node JS by running the following in the `sourcegraph/sourcegraph` repository clone (See [Get the code](#get-the-code) for cloning the repository):

    ```
    nvm install
    nvm use --delete-prefix
    ```

    After doing this, `node -v` should show the same version mentioned in
    `.nvmrc` at the root of the sourcegraph repository.

    > NOTE: Although there is a Homebrew package for Node, we advise using `nvm`
    instead, to ensure you get a Node version compatible with the current state
    of the sourcegraph repository.

### Ubuntu


1. Add package repositories:

    ```
    # Go
    sudo add-apt-repository ppa:longsleep/golang-backports

    # Docker
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
    sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

    # Yarn
    curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
    echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list
    ```

2. Update repositories:

    ```
    sudo apt-get update
    ```

3. Install dependencies:

    ```
    sudo apt install -y make git-all libpcre3-dev libsqlite3-dev pkg-config golang-go docker-ce docker-ce-cli containerd.io yarn jq libnss3-tools

    # Install comby
    curl -L https://github.com/comby-tools/comby/releases/download/0.11.3/comby-0.11.3-x86_64-linux.tar.gz | tar xvz

    # The extracted binary must be in your $PATH available as `comby`.
    # Here's how you'd move it to `/usr/local/bin` (which is most likely in your `$PATH`):
    chmod +x comby-*-linux
    mv comby-*-linux /usr/local/bin/comby

    # Install nvm (to manage Node.js)
    NVM_VERSION="$(curl https://api.github.com/repos/nvm-sh/nvm/releases/latest | jq -r .name)"
    curl -L https://raw.githubusercontent.com/nvm-sh/nvm/"$NVM_VERSION"/install.sh -o install-nvm.sh
    sh install-nvm.sh

    # In sourcegraph repository directory: install current recommendend version of Node JS
    nvm install
    ```

4. Choose to run Postgres and Redis manually (Option a.) or via Docker (Option b.)

    a. **Without Docker**

      1. Install PostgreSQL and Redis with the following commands:

          ```
          sudo apt install -y redis-server
          sudo apt install -y postgresql postgresql-contrib
          ```

      2. (optional) Start the services (and configure them to start automatically):

          ```
          sudo systemctl enable --now postgresql
          sudo systemctl enable --now redis-server.service
          ```

    b. **With Docker**

    We provide a docker compose file at `dev/redis-postgres.yml` to make it easy to run Redis and PostgreSQL as docker containers.

    > NOTE: Although Ubuntu provides a `docker-compose` package, we recommend to install the latest version via `pip` so that it is compatible with our compose file.

    See the official [docker compose documentation](https://docs.docker.com/compose/install/) for more details on different installation options.


### (optional) asdf

[asdf](https://github.com/asdf-vm/asdf) is a CLI tool that manages runtime versions for a number of different languages and tools. It can be likened to a language-agnostic version of [nvm](https://github.com/nvm-sh/nvm) or [pyenv](https://github.com/pyenv/pyenv).

We use asdf in buildkite to lock the versions of the tools that we use on a per-commit basis.

#### Install

##### asdf binary

See the [installation instructions on the official asdf documentation](https://asdf-vm.com/#/core-manage-asdf?id=install).

##### Plugins

sourcegraph/sourcegraph uses the following plugins:

- [Go](https://github.com/kennyp/asdf-golang)

```bash
asdf plugin add golang
```

- [NodeJS](https://github.com/asdf-vm/asdf-nodejs)

```bash
asdf plugin add nodejs

# Import the Node.js release team's OpenPGP keys to main keyring
bash ~/.asdf/plugins/nodejs/bin/import-release-team-keyring

# Have asdf read .nvmrc for auto-switching between node version
## Add the following to $HOME/.asdfrc:
legacy_version_file = yes
```

- [Yarn](https://github.com/twuni/asdf-yarn)

```bash
asdf plugin add yarn
```

#### Usage instructions

[asdf](https://github.com/asdf-vm/asdf) uses versions specified in [.tool-versions](https://github.com/sourcegraph/sourcegraph/blob/main/.tool-versions) whenever a command is run from one of `sourcegraph/sourcegraph`'s subdirectories.

You can install the all the versions specified in [.tool-versions](https://github.com/sourcegraph/sourcegraph/blob/main/.tool-versions) by running `asdf install`.

## Get the code

Run the following command in a folder where you want to keep a copy of the code. Command will create a new sub-folder (`sourcegraph`) in this folder.

```bash
git clone https://github.com/sourcegraph/sourcegraph.git
```

### For Sourcegraph employees: clone shared configuration

In order to run the local development environment as a Sourcegraph employee, you'll need to clone another repository: [`sourcegraph/dev-private`](https://github.com/sourcegraph/dev-private). It contains convenient preconfigured settings and code host connections.

It needs to be cloned into the same folder as `sourcegraph/sourcegraph`, so they sit alongside each other. To illustrate:

```
/dir
 |-- dev-private
 +-- sourcegraph
```

> NOTE: Ensure that you periodically pull the latest changes from [`sourcegraph/dev-private`](https://github.com/sourcegraph/dev-private) as the secrets are updated from time to time.

## Start Docker

### macOS

#### Option A: Docker for Mac

This is the easy way—just launch Docker.app and wait for it to finish loading.

#### Option B: docker-machine

The Docker daemon should be running in the background, which you can test by
running `docker ps`. If you're on OS X and using `docker-machine` instead of
Docker for Mac, you may have to run:

```bash
docker-machine start default
eval $(docker-machine env)
```

### Ubuntu/Linux

The docker daemon might already be running, but if necessary you can use the following commands to start it:

```sh
# as a system service
sudo systemctl enable --now docker

# manually
dockerd
```

If you have issues running Docker, try [adding your user to the docker group][dockerGroup], and/or [updating the socket file permissions][socketPermissions], or try running these commands under `sudo`.

[dockerGroup]: https://stackoverflow.com/a/48957722
[socketPermissions]: https://stackoverflow.com/a/51362528

## Initialize your database

### With Docker

The Sourcegraph server reads PostgreSQL connection configuration from the [`PG*` environment variables](http://www.postgresql.org/docs/current/static/libpq-envars.html).

The development server startup script as well as the docker compose file provide default settings, so it will work out of the box.

To initialize your database, you may have to set the appropriate environment variables before running the `createdb` command:

```sh
createdb --host=localhost --user=sourcegraph --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph
```

You can also use the `PGDATA_DIR` environment variable to specify a local folder (instead of a volume) to store the database files. See the `dev/redis-postgres.yml` file for more details.

This can also be spun up using [`sg run redis-postgres`](../background-information/sg/index.md), with the following `sg.config.override.yaml`:

```yaml
env:
    PGHOST: localhost
    PGPASSWORD: sourcegraph
    PGUSER: sourcegraph
```

### Without Docker

You need a fresh Postgres database and a database user that has full ownership of that database.

1. Create a database for the current Unix user

    ```
    # For Linux users, first access the postgres user shell
    sudo su - postgres
    # For Mac OS users
    sudo su - _postgres
    ```

    ```
    createdb
    ```

2. Create the Sourcegraph user and password

    ```
    createuser --superuser sourcegraph
    psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"
    ```

3. Create the Sourcegraph database

    ```
    createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph
    ```

4. Configure database settings in your environment

    The Sourcegraph server reads PostgreSQL connection configuration from the [`PG*` environment variables](http://www.postgresql.org/docs/current/static/libpq-envars.html).

    Our `sg` configuration in `sg.config.yaml` sets values that work with the setup described here, but if you want to use differnt values you can overwrite them in `sg.config.overwite.yaml`, like this:

    ```
    env:
      PGPORT=5432
      PGHOST=localhost
      PGUSER=sourcegraph
      PGPASSWORD=sourcegraph
      PGDATABASE=sourcegraph
      PGSSLMODE=disable
    ```

### More info

For more information about data storage, [read our full PostgreSQL page](../background-information/postgresql.md).

Migrations are applied automatically.


## Configure HTTPS reverse proxy

Sourcegraph's development environment ships with a [Caddy 2](https://caddyserver.com/) HTTPS reverse proxy that allows you to access your local sourcegraph instance via `https://sourcegraph.test:3443` (a fake domain with a self-signed certificate that's added to `/etc/hosts`).

If you'd like Sourcegraph to be accessible under `https://sourcegraph.test` (port 443) instead, you can [set up authbind](https://medium.com/@steve.mu.dev/setup-authbind-on-mac-os-6aee72cb828) and set the environment variable `SOURCEGRAPH_HTTPS_PORT=443`.

### Prerequisites

In order to configure the HTTPS reverse-proxy, you'll need to edit `/etc/hosts` and initialize Caddy 2.

### Add `sourcegraph.test` to `/etc/hosts`

`sourcegraph.test` needs to be added to `/etc/hosts` as an alias to `127.0.0.1`. There are two main ways of accomplishing this:

1. Manually append `127.0.0.1 sourcegraph.test` to `/etc/hosts`
1. Use the provided `./dev/add_https_domain_to_hosts.sh` convenience script (sudo may be required).

```bash
> ./dev/add_https_domain_to_hosts.sh

--- adding sourcegraph.test to '/etc/hosts' (you may need to enter your password)
Password:
Adding host(s) "sourcegraph.test" to IP address 127.0.0.1
--- printing '/etc/hosts'
...
127.0.0.1        localhost sourcegraph.test
...
```

### Initialize Caddy 2

[Caddy 2](https://caddyserver.com/) automatically manages self-signed certificates and configures your system so that your web browser can properly recognize them. The first time that Caddy runs, it needs `root/sudo` permissions to add
its keys to your system's certificate store. You can get this out the way after installing Caddy 2 by running the following command and entering your password if prompted:

```bash
./dev/caddy.sh trust
```

Note: If you are using Firefox and have a master password set, the following prompt will come up first:

```
Enter Password or Pin for "NSS Certificate DB":
```

Enter your Firefox master password here and proceed. See [this issue on GitHub](https://github.com/FiloSottile/mkcert/issues/50) for more information.

You might need to restart your web browsers in order for them to recognize the certificates.

## Start the server

### Configure `sg` to connect to databases

If you chose to run PostgreSQL and Redis **without Docker** they should already be running. You can jump the next section.

If you chose to run Redis and PostgreSQL **with Docker** to then we need to configure `sg` so it can connect to them.

1. In the `sourcegraph` folder, create a `sg.config.overwrite.yaml` file with the following contents (don't worry, `sg.config.overwrite.yaml` files are ignored by `git` and serve as a place for your local configuration):

    ```
    env:
        POSTGRES_HOST: localhost
        PGPASSWORD: sourcegraph
        PGUSER: sourcegraph
    ```

2. Start the databases:

    ```
    sg run redis-postgres
    ```

Keep this process running in a terminal window to keep the databases running. Follow the rest of the instructions in another terminal.

### Start the server

**If you are a Sourcegraph employee**: start the local development server for Sourcegraph Enterprise with the following command:

```
sg start
```

**If you are not a Sourcegraph employee and don't have access to [the `dev-private` repository](#get-the-code)**: you want to start Sourcegraph OSS, do this:

```
sg start oss
```

This will continuously compile your code and live reload your locally running instance of Sourcegraph.

Navigate your browser to https://sourcegraph.test:3443 to see if everything worked.

If `sg` exits with errors or outputs errors, take a look at [Troubleshooting](troubleshooting.md) or ask in the `#dev-experience` Slack channel.

### Running the server in different configurations

If you want to run the server in different configurations (with the monitoring stack, with code insights enabled, Sourcegraph OSS, ...), run the following:

```
sg start -help
```

That prints a list of possible configurations which you can start with `sg start`.

For example, you can start Sourcegraph in the mode it uses on Sourcegraph.com by running the following in one terminal window

```
sg start dotcom
```

and then, in another terminal window, start the monitoring stack:

```
sg start monitoring
```

## Additional resources

Congratulations on making it to the end of the quickstart guide!
Here are some additional resources to help you go further:

- [`sg`, the Sourcegraph developer tool](../background-information/sg/index.md)
- [Troubleshooting local development](troubleshooting.md)
- [Background information](../background-information/index.md) for more context

