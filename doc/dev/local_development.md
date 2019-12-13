# Getting started with developing Sourcegraph

Have a look around, our code is on [GitHub](https://sourcegraph.com/github.com/sourcegraph/sourcegraph).

## Outline

- [Environment](#environment)
- [Step 1: Install dependencies](#step-1-install-dependencies)
- [Step 2: Initialize your database](#step-2-initialize-your-database)
- [Step 3: (macOS) Start Docker](#step-3-macos-start-docker)
- [Step 4: Get the code](#step-4-get-the-code)
- [Step 5: Start the Server](#step-5-start-the-server)
- [Troubleshooting](#troubleshooting)
- [How to Run Tests](#how-to-run-tests)
- [CPU/RAM/bandwidth/battery usage](#cpurambandwidthbattery-usage)
- [How to debug live code](#how-to-debug-live-code)
- [Windows support](#windows-support)
- [Other nice things](#other-nice-things)

## Environment

Sourcegraph server is a collection of smaller binaries. The development server, [dev/start.sh](https://github.com/sourcegraph/sourcegraph/blob/master/dev/start.sh), initializes the environment and starts a process manager that runs all of the binaries. See the [Architecture doc](architecture/index.md) for a full description of what each of these services does. The sections below describe the dependencies you need to run `dev/start.sh`.

## Step 1: Install dependencies

Sourcegraph has the following dependencies:

- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) (v2.18 or higher)
- [Go](https://golang.org/doc/install) (v1.13 or higher)
- [Node JS](https://nodejs.org/en/download/) (version 8 or 10)
- [make](https://www.gnu.org/software/make/)
- [Docker](https://docs.docker.com/engine/installation/) (v18 or higher)
  - For macOS we recommend using Docker for Mac instead of `docker-machine`
- [PostgreSQL](https://wiki.postgresql.org/wiki/Detailed_installation_guides) (v11 or higher)
- [Redis](http://redis.io/) (v3.0.7 or higher)
- [Yarn](https://yarnpkg.com) (v1.10.1 or higher)
- [nginx](https://docs.nginx.com/nginx/admin-guide/installing-nginx/installing-nginx-open-source/) (v1.14 or higher)
- [SQLite](https://www.sqlite.org/index.html) tools
- [Golang Migrate](https://github.com/golang-migrate/migrate/) (v4.7.0 or higher)
- [Comby](https://github.com/comby-tools/comby/) (v0.11.3 or higher)

The following are two recommendations for installing these dependencies:

### macOS

1.  Install [Homebrew](https://brew.sh).
2.  Install [Docker for Mac](https://docs.docker.com/docker-for-mac/).

    optionally via `brew`

    ```bash
    brew cask install docker
    ```

3.  Install Go, Node, PostgreSQL, Redis, Git, nginx, golang-migrate, Comby, and SQLite tools with the following command:

    ```bash
    brew install go node yarn redis postgresql git gnu-sed nginx golang-migrate comby sqlite pcre FiloSottile/musl-cross/musl-cross
    ```

4.  Configure PostgreSQL and Redis to start automatically

    ```bash
    brew services start postgresql
    brew services start redis
    ```

    (You can stop them later by calling `stop` instead of `start` above.)

5.  Ensure `psql`, the PostgreSQL command line client, is on your `$PATH`.
    Homebrew does not put it there by default. Homebrew gives you the command to run to insert `psql` in your path in the "Caveats" section of `brew info postgresql`. Alternatively, you can use the command below. It might need to be adjusted depending on your Homebrew prefix (`/usr/local` below) and shell (bash below).

    ```bash
    hash psql || { echo 'export PATH="/usr/local/opt/postgresql/bin:$PATH"' >> ~/.bash_profile }
    source ~/.bash_profile
    ```

6.  Open a new Terminal window to ensure `psql` is now on your `$PATH`.

### Ubuntu


1. Add package repositories:

    ```bash
    # Go
    sudo add-apt-repository ppa:longsleep/golang-backports

    # Docker
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
    sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

    # Yarn
    curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
    echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list

    # Node.js
    curl -sL https://deb.nodesource.com/setup_10.x | sudo -E bash -
    ```

2. Update repositories:

    ```bash
    sudo apt-get update
    ```

3. Install dependencies:

    ```bash
    sudo apt install -y make git-all postgresql postgresql-contrib redis-server nginx libpcre3-dev libsqlite3-dev pkg-config golang-go musl-tools docker-ce docker-ce-cli containerd.io nodejs yarn

    # install golang-migrate (you must move the extracted binary into your $PATH)
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.7.0/migrate.linux-amd64.tar.gz | tar xvz

    # install comby (you must move the extracted binary into your $PATH)
    curl -L https://github.com/comby-tools/comby/releases/download/0.11.3/comby-0.11.3-x86_64-linux.tar.gz | tar xvz
    ```

4. Configure startup services

    ```bash
    sudo systemctl enable postgresql
    sudo systemctl enable redis-server.service
    ```

5. (optional) You can also run Redis using Docker

    In this case you should not enable the `redis-server.service` from the previous step.

    ```bash
    dockerd # if docker isn't already running
    docker run -p 6379:6379 -v $REDIS_DATA_DIR redis
    # $REDIS_DATA_DIR should be an absolute path to a folder where you intend to store Redis data
    ```

    You need to have Redis running when you start the dev server later on. If you have issues running Docker, try [adding your user to the docker group][dockerGroup], and/or [updating the socket file persimissions][socketPermissions], or try running these commands under `sudo`.

    [dockerGroup]: https://stackoverflow.com/a/48957722
    [socketPermissions]: https://stackoverflow.com/a/51362528

## Step 2: Initialize your database

You need a fresh Postgres database and a database user that has full ownership of that database.

1. Create a database for the current Unix user

    ```bash
    # For Linux users, first access the postgres user shell
    sudo su - postgres
    ```

    ```bash
    createdb
    ```

2. Create the Sourcegraph user and password

    ```bash
    createuser --superuser sourcegraph
    psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"
    ```

3. Create the Sourcegraph database

    ```bash
    createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph
    ```

4. Configure database settings in your environment

    The Sourcegraph server reads PostgreSQL connection configuration from the [`PG*` environment variables](http://www.postgresql.org/docs/current/static/libpq-envars.html).

    Add these, for example, in your `~/.bashrc`:

    ```bash
    export PGPORT=5432
    export PGHOST=localhost
    export PGUSER=sourcegraph
    export PGPASSWORD=sourcegraph
    export PGDATABASE=sourcegraph
    export PGSSLMODE=disable
    ```

    You can also use a tool like [`envdir`][envdir] or [a `.dotenv` file][dotenv] to
    source these env vars on demand when you start the server.

    [envdir]: https://cr.yp.to/daemontools/envdir.html
    [dotenv]: https://github.com/joho/godotenv

### More info

For more information about data storage, [read our full PostgreSQL Guide
page](postgresql.md).

Migrations are applied automatically.

## Step 3: (macOS) Start Docker

#### Option A: Docker for Mac

This is the easy way - just launch Docker.app and wait for it to finish loading.

#### Option B: docker-machine

The Docker daemon should be running in the background, which you can test by
running `docker ps`. If you're on OS X and using `docker-machine` instead of
Docker for Mac, you may have to run:

```bash
docker-machine start default
eval $(docker-machine env)
```

## Step 4: Get the code

```bash
git clone https://github.com/sourcegraph/sourcegraph.git
```

## Step 5: Start the Server

```bash
cd sourcegraph
./dev/start.sh
```

This will continuously compile your code and live reload your locally running instance of Sourcegraph.

Navigate your browser to http://localhost:3080 to see if everything worked.

## Troubleshooting

#### Problems with node_modules or Javascript packages

Noticing problems with <code>node_modules/</code> or package versions? Try
running this command to clear the local package cache.

```bash
yarn cache clean
rm -rf node_modules web/node_modules
yarn
cd web
yarn
```

#### dial tcp 127.0.0.1:3090: connect: connection refused

This means the `frontend` server failed to start, for some reason. Look through
the previous logs for possible explanations, such as failure to contact the
`redis` server, or database migrations failing.

#### Database migration failures

While developing Sourcegraph, you may run into:

`frontend | failed to migrate the DB. Please contact hi@sourcegraph.com for further assistance:Dirty database version 1514702776. Fix and force version.`

You may have to run migrations manually. First, install the Go [migrate](https://github.com/golang-migrate/migrate/tree/master/cli#installation) CLI, then run `dev/migrate.sh up`

If you get something like `error: Dirty database version 1514702776. Fix and force version.`, you need to roll things back and start from scratch.

```bash
dev/migrate.sh drop
dev/migrate.sh up
```

If you receive errors while migrating, try dropping the database

```bash
dev/drop-entire-local-database.sh
dev/migrate.sh up
```

#### Internal Server Error

If you see this error when opening the app:

`500 Internal Server Error template: app.html:21:70: executing "app.html" at <version "styles/styl...>: error calling version: open ui/assets/styles/app.bundle.css: no such file or directory`

that means Webpack hasn't finished compiling the styles yet (it takes about 3 minutes).
Simply wait a little while for a message from webpack like `web | Time: 180000ms` to appear
in the terminal.

#### Increase maximum available file descriptors.

`./dev/start.sh` may ask you to run ulimit to increase the maximum number
of available file descriptors for a process. You can make this setting
permanent for every shell session by adding the following line to your
`.*rc` file (usually `.bashrc` or `.zshrc`):

```bash
# increase max number of file descriptors for running a sourcegraph instance.
ulimit -n 10000
```

On Linux, it may also be necessary to increase `sysctl -n fs.inotify.max_user_watches`, which can be
done by running one of the following:

```bash
echo 524288 | sudo tee -a /proc/sys/fs/inotify/max_user_watches

# If the above doesn't work, you can also try this:
sudo sysctl fs.inotify.max_user_watches=524288
```

If you ever need to wipe your local database and Redis, run the following command.

```bash
./dev/drop-entire-local-database-and-redis.sh
```

## How to Run Tests

See [testing.md](testing.md) for details.

## CPU/RAM/bandwidth/battery usage

On first install, the program will use quite a bit of bandwidth to concurrently
download all of the Go and Node packages. After packages have been installed,
the Javascript assets will be compiled into a single Javascript file, which
can take up to 5 minutes, and can be heavy on the CPU at times.

After the initial install/compile is complete, the Docker for Mac binary uses
about 1.5GB of RAM. The numerous different Go binaries don't use that much RAM
or CPU each, about 5MB of RAM each.

Some users report [heavy battery usage running `gulp watch`][battery-usage].
[Double check that Spotlight is not indexing your Sourcegraph
repository][spotlight], as this can lead to additional, unnecessary, poll
events. We are investigating other causes of this issue.

[battery-usage]: https://github.com/sourcegraph/sourcegraph/issues/247
[spotlight]: https://www.macobserver.com/tips/how-to/stop-spotlight-indexing/

## How to debug live code

How to debug a program with Visual Studio Code:

### Debug TypeScript code

Requires "Debugger for Chrome" extension.

- Quit Chrome
- Launch Chrome (Canary) from the command line with a remote debugging port:
  - Mac OS: `/Applications/Google\ Chrome\ Canary.app/Contents/MacOS/Google\ Chrome\ Canary --remote-debugging-port=9222`
  - Windows: `start chrome.exe –remote-debugging-port=9222`
  - Linux: `chromium-browser --remote-debugging-port=9222`
- Go to http://localhost:3080
- Open the Debugger in VSCode: "View" > "Debug"
- Launch the `(ui) http://localhost:3080/*` debug configuration
- Set breakpoints, enjoy

### Debug Go code

Install [Delve](https://github.com/derekparker/delve):

```bash
xcode-select --install
go get -u github.com/go-delve/delve/cmd/dlv
```

Then install `pgrep`:

```bash
brew install proctools
```

Make sure to run `env DELVE=true dev/start.sh` to disable optimizations during compilation, otherwise Delve will have difficulty stepping through optimized functions (line numbers will be off, you won't be able to print local variables, etc.).

Now you can attach a debugger to any Go process (e.g. frontend, searcher, go-langserver) in 1 command:

```bash
dlv attach $(pgrep frontend)
```

Delve will pause the process once it attaches the debugger. Most used [commands](https://github.com/go-delve/delve/tree/master/Documentation/cli):

- `b cmd/frontend/db/access_tokens.go:52` to set a breakpoint on a line (`bp` lists all, `clearall` deletes all)
- `c` to continue execution of the program
- `Ctrl-C` pause the program to bring back the command prompt
- `n` to step over the next statement
- `s` to step into the next function call
- `stepout` to step out of the current function call
- `Ctrl-D` to exit

## Go dependency management

We use Go modules to manage Go dependencies in this repository.

## Codegen

The Sourcegraph repository relies on code generation triggered by `go generate`. Code generation is used for a variety of tasks:

- generating code for mocking interfaces
- generate wrappers for interfaces (e.g., `./server/internal/middleware/*` packages)
- pack app templates and assets into binaries

To generate everything, just run:

```bash
./dev/generate.sh
```

Note: Sometimes, there are erroneous diffs. This occurs for a few
reasons, none of which are legitimate (i.e., they are tech debt items
we need to address):

- The codegen tools might emit code that depends on system configuration,
  such as the system timezone or packages you have in your GOPATH. We
  need to submit PRs to the tools to eliminate these issues.
- You might have existing but gitignored files that the codegen tools
  read on your disk that other developers don't have. (This occurs for
  app assets especially.)

If you think a diff is erroneous, don't commit it. Add a tech debt
item to the issue tracker and assign the person who you think is
responsible (or ask).

## Windows support

Running Sourcegraph on Windows is not actively tested, but should be possible within the Windows Subsystem for Linux (WSL).
Sourcegraph currently relies on Unix specifics in several places, which makes it currently not possible to run Sourcegraph directly inside Windows without WSL.
We are happy to accept contributions here! :)

## Other nice things

### Offline development

Sometimes you will want to develop Sourcegraph but it just so happens you will be on a plane or a
train or perhaps a beach, and you will have no WiFi. And you may raise your fist toward heaven and
say something like, "Why, we can put a man on the moon, so why can't we develop high-quality code
search without an Internet connection?" But lower your hand back to your keyboard and fret no
further, for the year is 2019, and you *can* develop Sourcegraph with no connectivity by setting the
`OFFLINE` environment variable:

```bash
OFFLINE=true dev/start.sh
```
