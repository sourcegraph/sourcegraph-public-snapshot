# Sourcegraph App

## ⚠️ Experimental ⚠️

Sourcegraph App is **highly experimental** and early-stages-**we do not advise you try it yet.** If you're interested, reach out to us on [Twitter](https://twitter.com/sourcegraph) or [Discord](https://discord.com/invite/s2qDtYGnAE) and we'll let you know when it's ready to try out!

<span class="badge badge-warning">The following is intended primarily for Sourcegraph employees</span>

## What is it?

Sourcegraph App runs alongside your IDE and provides a better way to understand, browse, and search both your local and remote code. It is lightweight, and provides IDE-like code navigation, powerful regexp/commit/diff search, etc. all on your local machine.

We're still working out all the details, this is a **very early-stages version** which doesn't live up to that vision yet. Still, we'd love if you try it out and let us know how it goes and what you think!

Learn more about [how this fits into our product strategy](https://handbook.sourcegraph.com/departments/engineering/teams/growth/app/)

## Installation

**macOS:** via homebrew:

```sh
brew install sourcegraph/sourcegraph-app/sourcegraph
```

**Linux:** via [deb pkg](https://storage.googleapis.com/sourcegraph-app-releases/0.0.196391-snapshot+20230131-f10a97/sourcegraph_0.0.196391-snapshot+20230131-f10a97_linux_amd64.deb) installer:

```sh
dpkg -i sourcegraph_0.0.196391-snapshot+20230131-f10a97_linux_amd64.deb
```

**Single-binary zip download:**

* [macOS (universal)](https://storage.googleapis.com/sourcegraph-app-releases/0.0.196391-snapshot+20230131-f10a97/sourcegraph_0.0.196391-snapshot+20230131-f10a97_darwin_all.zip)
* [linux (x64)](https://storage.googleapis.com/sourcegraph-app-releases/0.0.196391-snapshot+20230131-f10a97/sourcegraph_0.0.196391-snapshot+20230131-f10a97_linux_amd64.zip)

## Prerequisites

Ensure you have the following:

1. `src` CLI available on your PATH ([installation](https://github.com/sourcegraph/src-cli))
2. `docker` is installed and on your PATH
3. Redis is running, e.g. via Docker:

```sh
docker run -p 127.0.0.1:6379:6379 -d redis redis-server --save 60 1 --loglevel warning
```

## Usage

Start Sourcegraph by running the following in a terminal:

```sh
sourcegraph
```

Navigate to http://localhost:3080 and you can add your remote repositories from there (we're still working on ability to add local code.)

### Troubleshooting

If it doesn't start, make sure:

* Redis is running on port 6379
* `docker` and `git` are installed and on your path
* `src` is installed and on your path 

### What works

* Adding repositories from GitHub, GitLab, etc.
* Code navigation
* Search
* Batch changes
* Precise code intel

### Known issues

* Can't add local code yet, only remote code
* Syntax highlighting is broken
* We're working on eliminating the Redis, `src` CLI, and Docker dependencies
* macOS binaries are not code-signed yet, so you may need to right-click -> open the binary if you do not use Homebrew.

## Feedback

You can provide feedback and get help in our [Discord](https://discord.com/invite/s2qDtYGnAE) or tweet [@sourcegraph](https://twitter.com/sourcegraph).

_Sourcegraph employees:_ join `#app` in Slack!
