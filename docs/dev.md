# Getting started with developing Sourcegraph

The best way to become familiar with the Sourcegraph repository is by
reading the code at https://sourcegraph.com/github.com/sourcegraph/sourcegraph.

## Environment

Before you can develop Sourcegraph you'll need to set up a
development environment.

### Manual setup

For Linux users or if you don't want to use Homebrew on macOS.

- [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- [Go](https://golang.org/doc/install) (v1.10.0 or higher)
- [Node JS](https://nodejs.org/en/download/) (v7.0.0 or higher)
- [make](https://www.gnu.org/software/make/)
- [Docker](https://docs.docker.com/engine/installation/) (v1.8 or higher)
  - if using Mac OS, we recommend using Docker for Mac instead of `docker-machine`
- [PostgreSQL](https://wiki.postgresql.org/wiki/Detailed_installation_guides) (v9.2 or higher)
- [Redis](http://redis.io/) (v3.0.7 or higher)

### Homebrew setup for macOS

This is a streamlined setup for Mac machines.

1.  Install [Docker for Mac](https://docs.docker.com/docker-for-mac/).
2.  Install [Homebrew](http://brew.sh).
3.  Install Go, Node, PostgreSQL, Redis, Git.

    ```
    brew install go node redis postgresql git gnu-sed
    ```

4.  Set up your [Go Workspace](https://golang.org/doc/code.html#Workspaces)

5.  Configure PostgreSQL and Redis to start automatically

    ```
    brew services start postgresql
    brew services start redis
    ```

## SSH keys

If you don't already have an SSH key on your machine (e.g. `~/.ssh/id_rsa`), [you should create one](https://help.github.com/articles/connecting-to-github-with-ssh/). This allows you to pull code from GitHub without typing in your password.

Note that you can use multiple SSH keys with different organizations, and have this handled automatically through your `.ssh/config` file. For instance:

```
Host    github.com
        User    git
        IdentityFile /home/YourNameHere/.ssh/github_rsa
```

## Get the code

You will want a local copy of the Sourcegraph tree. (If you don't have access to this yet, you need to be added to the Sourcegraph organization on GitHub.)

```
git clone git@github.com:sourcegraph/sourcegraph.git $GOPATH/src/github.com/sourcegraph/sourcegraph
cd $GOPATH/src/github.com/sourcegraph/sourcegraph
```

This is your "Sourcegraph repository directory".

## Create a BuildKite Account

Create an account [here](https://buildkite.com/signup), and then ask someone to add you the Sourcegraph organization. You can see the latest builds for Sourcegraph.com [here](https://buildkite.com/sourcegraph/sourcegraph).

## Create a NPM Account

Create an account [here](https://www.npmjs.com/signup), and then ask @sqs to add you to the Sourcegraph organization. You need to have recent NodeJS and `npm`; as of this writing, you want NodeJS 8 or 10, and npm 6. (See subhead below for instructions for Ubuntu.)

Run [`npm login`](https://docs.npmjs.com/cli/adduser) inside the Sourcegraph directory, and input your npmjs.org credentials.

### Getting recent NodeJS on Ubuntu

Ubuntu installs a fairly old NodeJS by default. To get a more recent version:

```
curl -sL https://deb.nodesource.com/setup_10.x | sudo -E bash -
sudo apt-get install -y nodejs
```

As of this writing, `setup_8.x` also works, but you may want to prefer the newer one.

Install Yarn:

```
npm i -g yarn@1.9.4
```

Noticing problems with `node_modules/` or package versions? Try:

```
yarn cache clean; rm -rf node_modules web/node_modules; yarn; cd web; yarn
```

If you want to test things across multiple node versions, consider using nvm:

https://github.com/creationix/nvm

## PostgreSQL

[Initialize and configure your database](https://github.com/sourcegraph/sourcegraph/blob/master/docs/storage.md).

## Redis

If you didn't install Redis through Homebrew in the [Homebrew setup for macOS](#homebrew-setup-for-macos), you can follow the [instructions to install Redis natively](http://redis.io/topics/quickstart). If you have Docker installed and are running Linux, however, the easiest way to get Redis up and running is probably:

```
dockerd # if docker isn't already running
docker run -p 6379:6379 -v $REDIS_DATA_DIR redis
```

_`$REDIS_DATA_DIR` should be an absolute path to a folder where you intend to store Redis data._

You need to have the redis image running when you run the Sourcegraph `dev/start.sh` script. If you do not have docker access without root, run these commands under `sudo`.

## Build

Make sure your [`$GOPATH` is set](https://golang.org/doc/code.html#GOPATH) and your `$PATH` includes `$GOPATH/bin`:

```
echo $GOPATH # should print something
echo $PATH # should include $GOPATH/bin
```

Then in your terminal run:

`cd $GOPATH/src/github.com/sourcegraph/sourcegraph`

The Docker daemon should be running in the background, which you can test by
running `docker ps`. If you're on OS X and using `docker-machine` instead of Docker for Mac,
you may have to run:

```
docker-machine start default
eval $(docker-machine env)
```

Then run the following commands (**NOTE: Node.js and Yarn must be installed for the this step**):

```
./dev/start.sh
```

This will continuously compile your code and live reload your locally running
instance of Sourcegraph. Navigate your browser to http://localhost:3080 to
see if everything worked.

### Troubleshooting

##### dial tcp 127.0.0.1:3090: connect: connection refused

This means the `frontend` server failed to start, for some reason. Look through the previous logs for possible explanations, such as failure to contact the `redis` server, or database migrations failing.

#### Database migration failures

While developing Sourcegraph, you may run into:

`frontend | failed to migrate the DB. Please contact hi@sourcegraph.com for further assistance:Dirty database version 1514702776. Fix and force version.`

You may have to run migrations manually. First, install the Go [migrate](https://github.com/golang-migrate/migrate/tree/master/cli#installation) CLI, and run something like:

Then try:

`dev/migrate.sh up`

If you get something like `error: Dirty database version 1514702776. Fix and force version.`, you need to roll things back and start from scratch.

```bash
dev/migrate.sh drop
dev/migrate.sh up
```

#### Internal Server Error

If you see this error when opening the app:

`500 Internal Server Error template: app.html:21:70: executing "app.html" at <version "styles/styl...>: error calling version: open ui/assets/styles/style.bundle.css: no such file or directory`

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

If you ever need to wipe your local database, run the following command.

```
./dev/drop-entire-local-database.sh
```

## Test

CI runs all our unit tests. To run tests within a directory (and recursively within its
subdirectories) on your machine:

```
go test ./cmd/...
```

To run a specific package's tests:

```
go test ./util/textutil
```

## Debugger

How to debug our code with Visual Studio Code.

### TypeScript

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

### Go

**Note: If you run into an error `could not launch process: decoding dwarf section info at offset 0x0: too short` make sure you are on the latest delve version**

- Install [Delve](https://github.com/derekparker/delve)
- Run `DELVE=frontend,searcher ./dev/start.sh` (`DELVE` accepts a comma-separated list of components as specified in [../dev/Procfile](../dev/Procfile))
- Set a breakpoint in VS Code (there's a bug where setting the breakpoint after attaching results in "Unverified breakpoint")
- Run "Attach to $component" in the VS Code debug view
- The process should start once the debugger is attached

Known issues:

- At the time of writing there is an issue with homebrew formula so workarounds are required.
  - Use homebrew and then google any errors you encounter.
- There doesn't seem to be a clean way to stop debugging (https://github.com/derekparker/delve/issues/1057).
  - The workaround is to manually kill the process when you are done.

## dep

We use [dep](https://github.com/golang/dep) to manage Go dependencies in this repository.

## Codegen

The Sourcegraph repository relies on code generation triggered by `go generate`. Code generation is used for a variety of tasks:

- generating code for mocking interfaces
- generate wrappers for interfaces (e.g., `./server/internal/middleware/*` packages)
- pack app templates and assets into binaries

To generate everything, just run:

```
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

## Code style guide

See [docs/style.md](style.md).

## Notes about alpha Windows support

Windows support for Sourcegraph is currently in alpha. A few extra
steps are required to run Sourcegraph in a Windows environment:

- Sourcegraph depends on some GNU and open-source command line tools
  like `make`, `sh`, and `find`. Install Cygwin from
  http://cygwin.org/ (tested with 2.2.1 x32).
- Install Git from http://git-scm.com rather than use Cygwin's
  default Git, which does not properly handle local repositories.
- Configure your Git to use Unix-style line endings.
- Unix symlinks must be converted to Windows symlimks. This does not
  happen automatically (see
  http://stackoverflow.com/questions/5917249/git-symlinks-in-windows).
  After checking out Sourcegraph, run `bash dev/git-windows-aliases.sh` to install the git aliases needed to do
  the conversion. Then run `git rm-symlinks` from the repository root
  to convert all symlinks into Windows symlinks and mark them as "not
  changed" (necessary to avoid issues related to running yarn-based
  tasks).

Note that multiple unit tests currently fail on Windows. We would be happy to accept contributions here! :)
