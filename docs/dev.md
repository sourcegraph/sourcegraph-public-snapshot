# Development README

The best way to become familiar with the Sourcegraph repository is by
reading the code at https://sourcegraph.com/github.com/sourcegraph/sourcegraph.

## Environment

Before you can develop Sourcegraph you'll need to set up a
development environment.

### Manual setup

For Linux users or if you don't want to use Homebrew on macOS.

- [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- [Go](https://golang.org/doc/install) (v1.7.0 or higher)
- [Node JS](https://nodejs.org/en/download/) (v7.0.0 or higher)
- [make](https://www.gnu.org/software/make/)
- [Docker](https://docs.docker.com/engine/installation/) (v1.8 or higher)
  - if using Mac OS, we recommend using Docker for Mac instead of `docker-machine`
- [PostgreSQL](https://wiki.postgresql.org/wiki/Detailed_installation_guides) (v9.2 or higher)
- [Redis](http://redis.io/) (v3.0.7 or higher)

### Homebrew setup for macOS

This is a streamlined setup for Mac machines.

1. Install [Docker for Mac](https://docs.docker.com/docker-for-mac/).
2. Install [Homebrew](http://homebrew.sh).
3. Install Go, Node, PostgreSQL, Redis, Git, Python.

    ```
    brew install go node redis postgresql git python gnu-sed
    ```

4. Install virtualenv.

    ```
    pip install virtualenv
    ```

5. Configure PostgreSQL and Redis to start automatically

    ```
    brew services start postgresql
	brew services start redis
    ```

### Optional (but recommended)

- [Visual Studio Code](https://code.visualstudio.com): this IDE comes with great out-of-the-box
  features for Go and TypeScript. Recommended extensions:
  - [Go](https://marketplace.visualstudio.com/items?itemName=lukehoban.Go)
  - [TSLint](https://marketplace.visualstudio.com/items?itemName=eg2.tslint)
  - [Debugger for Chrome](https://marketplace.visualstudio.com/items?itemName=msjsdiag.debugger-for-chrome)

## SSH keys

If you don't already have an SSH key on your machine (e.g. `~/.ssh/id_rsa`), [you should create one](https://help.github.com/articles/connecting-to-github-with-ssh/). This allows you to pull code from GitHub without typing in your password.

## Create a test GitHub account

This should be a separate GitHub user account for development whose username has the suffix `-test`

- get somebody to add you to the "sourcegraphtest" GitHub organization
- add a second profile to Chrome for your `*-test` GitHub user (https://cl.ly/3A3y1O040G3R),
    or download [Chrome Canary](https://www.google.com/chrome/browser/canary.html) to use for development

## Create a BuildKite Account 

Create an account [here](https://buildkite.com/signup), and then ask someone to add you the Sourcegraph organization. You can see the latest builds for Sourcegraph.com [here](https://buildkite.com/sourcegraph/sourcegraph). 

## Create a NPM Account

Create an account [here](https://www.npmjs.com/signup), and then ask @sqs to add you to the Sourcegraph organization. 

Run [`npm login`](https://docs.npmjs.com/cli/adduser) inside the Sourcegraph directory, and input your npmjs.org credentials. 

## Get the code

```
git clone git@github.com:sourcegraph/sourcegraph.git $GOPATH/src/sourcegraph.com/sourcegraph/sourcegraph
cd $GOPATH/src/sourcegraph.com/sourcegraph/sourcegraph
```

Running the preceding commands will build and install the `src` binary in `$GOPATH/bin`, which you will use in subsequent steps such as PostgreSQL setup.

## PostgreSQL

[Initialize and configure your database](https://github.com/sourcegraph/sourcegraph/blob/master/docs/storage.md).

## Redis

You can follow the [instructions to install Redis natively](http://redis.io/topics/quickstart). If you have Docker installed and are running Linux, however, the easiest way to get Redis up and running is probably:

```
sudo dockerd # if docker isn't already running
sudo docker run -p 6379:6379 -v $REDIS_DATA_DIR redis
```

**`$REDIS_DATA_DIR` should be an absolute path to a folder where you intend to store Redis data.**

## Build

Make sure your [`$GOPATH` is set](https://golang.org/doc/code.html#GOPATH) and your `$PATH` includes `$GOPATH/bin`:

```
echo $GOPATH # should print something
echo $PATH # should include $GOPATH/bin
```

Then in your terminal run:

`cd $GOPATH/src/sourcegraph.com/sourcegraph/sourcegraph`

The Docker daemon should be running in the background, which you can test by
running `docker ps`. If you're on OS X and using `docker-machine` instead of Docker for Mac,
 you may have to run:

```
docker-machine start default
eval $(docker-machine env)
```

Then run the following commands (**NOTE: Node.js and npm must be installed for the this step**):

```
./dev/start.sh
```

This will continuously compile your code and live reload your locally running
instance of Sourcegraph. Navigate your browser to http://localhost:3080 to
see if everything worked.

`./dev/start.sh` may ask you to run ulimit to increase the maximum number
of available file descriptors for a process. You can make this setting
permanent for every shell session by adding the following line to your
`.*rc` file (usually bash or zsh):

```bash
# increase max number of file descriptors for running a sourcegraph instance.
ulimit -n 10000
```

If you ever need to wipe your local database, run the following command:
```
./dev/drop-entire-local-database.sh
```

## Test

BuildKite runs all our unit tests, and its output can be viewed [here](https://buildkite.com/sourcegraph/sourcegraph). 

To run tests within a directory (and recursively within its
subdirectories):

```
go test ./app/...
```

To run a specific package's tests:

```
go test ./util/textutil
```

## Debugger

If you use VSCode and have the "Debugger for Chrome" extension, these steps allow you to
use the interactive visual debugger for our frontend codebase:

- Quit Chrome
- Launch Chrome (Canary) from the command line with a remote debugging port:
  - Mac OS: `/Applications/Google\ Chrome\ Canary.app/Contents/MacOS/Google\ Chrome\ Canary --remote-debugging-port=9222`
  - Windows: `start chrome.exe –remote-debugging-port=9222`
  - Linux: `chromium-browser --remote-debugging-port=9222`
- Go to http://localhost:3080
- Open the Debugger in VSCode: "View" > "Debug"
- Launch the `(ui) http://localhost:3080/*` debug configuration
- Set breakpoints, enjoy

## govendor

The Sourcegraph repository uses
[govendor](https://github.com/kardianos/govendor) to manage Go dependencies.

## Codegen

The Sourcegraph repository relies on code generation triggered by `go
generate`. Code generation is used for a variety of tasks:

* generating code for mocking interfaces
* generate wrappers for interfaces (e.g., `./server/internal/middleware/*` packages)
* pack app templates and assets into binaries

Then run:

```
./dev/generate.sh
```

Also, sometimes there are erroneous diffs. This occurs for a few
reasons, none of which are legitimate (i.e., they are tech debt items
we need to address):

* The codegen tool might emit code that depends on system configuration,
  such as the system timezone or packages you have in your GOPATH. We
  need to submit PRs to the tools to eliminate these issues.
* You might have existing but gitignored files that the codegen tools
  read on your disk that other developers don't have. (This occurs for
  app assets especially.)

If you think a diff is erroneous, don't commit it. Add a tech debt
item to the issue tracker and assign the person who you think is
responsible (or ask).

## Code standards

The Sourcegraph repository enforces some code standards via `make
test`, which is also run in CI.

Read more about our style in [docs/style.md](style.md).

## Windows notes

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
  After checking out Sourcegraph, run `bash
  dev/git-windows-aliases.sh` to install the git aliases needed to do
  the conversion. Then run `git rm-symlinks` from the repository root
  to convert all symlinks into Windows symlinks and mark them as "not
  changed" (necessary to avoid issues related to running npm-based
  tasks).
  
  P.S. #cuddy was here

Note that multiple unit tests currently fail on Windows.
