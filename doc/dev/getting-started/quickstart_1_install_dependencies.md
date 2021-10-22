# Quickstart step 1: Install dependencies

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

## macOS

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

          Homebrew does not put it there by default. Homebrew gives you the command to run to insert `psql` in your path in the "Caveats" section of `brew info postgresql`.
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

7.  Install the current recommended version of Node JS by running the following in the `sourcegraph/sourcegraph` repository clone (See [Cloning our repository](quickstart_2_clone_repository.md) for cloning the repository):

    ```
    nvm install
    nvm use --delete-prefix
    ```

    After doing this, `node -v` should show the same version mentioned in
    `.nvmrc` at the root of the sourcegraph repository.

    > NOTE: Although there is a Homebrew package for Node, we advise using `nvm`
    instead, to ensure you get a Node version compatible with the current state
    of the sourcegraph repository.

## Ubuntu


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


## (optional) asdf

[asdf](https://github.com/asdf-vm/asdf) is a CLI tool that manages runtime versions for a number of different languages and tools. It can be likened to a language-agnostic version of [nvm](https://github.com/nvm-sh/nvm) or [pyenv](https://github.com/pyenv/pyenv).

We use asdf in buildkite to lock the versions of the tools that we use on a per-commit basis.

### Install

#### asdf binary

See the [installation instructions on the official asdf documentation](https://asdf-vm.com/#/core-manage-asdf?id=install).

#### Plugins

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

### Usage instructions

[asdf](https://github.com/asdf-vm/asdf) uses versions specified in [.tool-versions](https://github.com/sourcegraph/sourcegraph/blob/main/.tool-versions) whenever a command is run from one of `sourcegraph/sourcegraph`'s subdirectories.

You can install the all the versions specified in [.tool-versions](https://github.com/sourcegraph/sourcegraph/blob/main/.tool-versions) by running `asdf install`.

[< Previous](index.md) | [Next >](quickstart_2_clone_repository.md)
