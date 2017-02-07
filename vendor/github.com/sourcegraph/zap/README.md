# Zap: multiplayer, realtime coding

Zap lets you share your Git repository you have write access to and editor state with any number of other editors and viewers. Each operation you perform (keystrokes, saving a file, creating a file, etc.) is replicated to all other participants. Conflicting edits are resolved immediately, just like how Google Docs does it. Think Google Docs meets Git.

## Getting started

These steps help you start using Zap on your repositories.

1. Clone the repository https://github.com/sgtest/xyztest to your local machine.
1. Install Zap: `go get -v -u github.com/sourcegraph/zap/cmd/zap` (if that fails due to lack of authentication, then run `cd $GOPATH/src/github.com/sourcegraph && git clone git@github.com:sourcegraph/zap.git && go get github.com/sourcegraph/zap/cmd/zap`)
1. Launch the Zap server: `zap server -v`
1. Install the Visual Studio Code extension: Cmd/Ctrl-P for quick open, then `ext install sqs.vscode-zap`

Then, in the directory of `sgtest/xyztest`, or any repository you want to use Zap on:

1. Tell Zap to start watching it: `zap init`
1. Configure the upstream repository: `zap remote set origin wss://sourcegraph.com/.api/zap github.com/sgtest/xyztest` (replace the last two parameters, the URL and repo name, with the appropriate values)
1. Set the current Zap branch to push upstream: `zap checkout -upstream origin -overwrite -create master@sqs` (use your unix username in place of sqs)
1. Open the repository in Visual Studio Code, hit alt + s to open Sourcegraph, and watch as cursors, selections, and edits are instantly synced from your editor to Sourcegraph

Notes:

* On Sourcegraph, you must run `features.zap.enable()` and `features.extensions.enable()`
* You must open Visual Studio Code after running `zap server -v`

## Hacking

### vscode-zap (Visual Studio Code extension)

To use the Visual Studio Code extension:

1. `cd ext/vscode && yarn`
2. Open the `./ext/vscode` directory in VSCode
3. Hit F5 (or go to the "Debug: Start Debugging" action in the launcher with the "Launch Extension" task chosen) to launch another instance with the extension running.

#### Release

This assumes you have run `lerna bootstrap` already.

1. `make publish-npm` to release the libzap and vscode-zap npm packages
1. `make publish-vscode` to release the vscode extension

## How it works

Zap's core concepts are heavily inspired by Git's:

* **workspace:** like a Git tree plus editor state and history: the contents of all files on disk in a directory tree, plus the contents of unsaved files in your editor, cursor positions for each user, and the current Git HEAD
* **op** (short for "operation"): like a Git commit diff, but capable of representing changes to anything in the workspace (including cursor position and changes to unsaved files)
* **rev** (short for "revision"): a sequential integer N that refers to a prior state of a ref (after only the first N ops were applied)
* **ref** (short for "reference"): a name that refers to some version of a workspace (similar to how a Git ref refers to a commit and, indirectly, tree)
  * **branch:** a ref of the form `refs/heads/branch` (where `branch` can be anything) that points to an active line of work
  * **user branch:** a ref of the form `refs/heads/branch@user` (where `branch` and `user` can be anything) that by convention mirrors a single user's workspace (and should not be edited by anyone else)
  * **shared branch:** all other refs of the form `refs/heads/branch` (where `branch` does not contain `@`), which by convention allow multiple clients editing
* **repository:** like a Git repository, but stores Zap refs and ops instead of Git refs and objects

#### Notes

* `git checkout B` always resets and switches to the Zap branch `B@user`, clobbering any existing workspace state.
  * If `git checkout B` switched to the shared ref `B`, then we'd have to handle merging the user's worktree with B's workspace state.
  * It's OK to clobber the `B@user` workspace state because by convention that branch should always mirror the user's workspace, and Zap will immediately update that branch with the user's current workspace.

## Development notes

zap is highly prone to many problems:

* race conditions, given that it's a distributed system where even the local state comes from multiple sources (the FS and the editor)
* platform incompatibilities (macOS FS notifications are different and hundreds of msec slower than inotify on Linux)
* editor incompatibilities (vscode encodes edits in slightly quirky ways that don't 100% map to our representation of edits, and vscode is one of the good ones!)
* complexity, given that we are using OT and Git and are modeling a complex thing (local workspace state)

It is easy to introduce bugs and can be hard to debug them. So, it's important that we:

* write good tests
* handle edge cases (or at least identify them and panic/error)
* add good logging

### Overview

Zap is distributed like Git: a Zap repository can have any number of remote repositories that it bidirectionally syncs with.

Typically, you run a local Zap server on your own machine, which is a single process that monitors multiple repositories on disk and communicates with your text editor. Each local Zap repository will also have a remote Zap repository hosted on a remote Zap server. The remote Zap server is the central point through which clients communicate with each other.

The remote server runs against the Git bare repository (the upstream repo that you push to). Zap clients (i.e., local servers) connect to the server over a socket and speak JSON-RPC 2.0 with it.

When you run `zap server` in your Git clone directory, it connects to the server and watches to the "workspace" on the server corresponding to your repository. Each workspace on the server is a separate stream of events. When you start watching, you receive all past events, so that you can sync up to the current state of the workspace.

Once the client is running, it receives changes from multiple sources and handles them appropriately:

* changes to your Git worktree (via fsnotify): encode into an op and send to server
* changes to unsaved files in your editor (via an editor extension), plus selection/cursor changes: encode into an op and send to server
* operations received from the server (that originated on other clients): perform the changes against the file system and (if there are changes to buffered files) forward to your editor to be applied there

Thanks to the careful definition of OT, as long as we have implemented our primitives (WorkspaceOp compose and transform) and apply/record/send these correctly, we can have any number of clients simultaneously editing and viewing changes.

### Implementation details

### Snapshot

zap maintains a "snapshot," which is a Git commit that refers to the last-known state of the worktree.

The snapshot only stores the worktree state. This means we need to take a snapshot each time a file changes on disk (or when a Git commit is made). It does not record cursor selections and keystrokes, which are much more frequent operations and would overwhelm Git with 100,000s of commits per day.

Taking snapshots is necessary because we must be able to compute a diff in response to an fsnotify event. By the time we receive the fsnotify event, the file has already changed, so we need a way to always know a file's previous contents. We could store a copy of each file in memory or in a temporary directory, but that'd be expensive. So, Git to the rescue! Git is optimized for doing exactly this.

We create a snapshot in 2 ways:

1. By diffing the worktree against the previous snapshot (including unstaged and untracked files), when a file changes.
2. By applying an op to the previous snapshot in memory, when we receive an op from the remote server. (Why not just apply the op to the worktree directly and then create a snapshot by using method 1 above? Because there might be other changes made to the worktree while we're applying the op's changes, and our snapshot would then incorporate changes that weren't in the op. This would mean we wouldn't send some of those worktree changes to the server, and our local workspace would diverge erroneously from the server's workspace.)

### Using alongside Git

Zap is designed to be easy to use for Git repositories.

By convention, Zap branch names correspond to Git branch names:

* When a user `alice` checks out a Git branch `b` (using `git checkout b`), their Zap client automatically checks out the Zap user branch `b@alice`.
* If two users `alice` and `bob` want to collaboratively edit a Git branch `b` using Zap, they will typically use a Zap shared branch named `b` (which they check out by both running `zap checkout b`).

The Zap client and server implementations also rely on Git:

* The Zap server periodically writes the Zap workspace state for each Zap branch `b` to a Git commit at the Git ref `refs/zap/snaps/b`. TODO(sqs): implement this
* When a user on a Zap branch `b` creates a Git commit locally, the Zap client automatically pushes it to the upstream's Git ref `refs/zap/b`.

Applications that currently support Git can be gracefully enhanced with Zap support as follows.

* TODO

### Tests

Zap has unit tests and integration tests for the server, clients, and extensions. To run them:

```
make test
```

#### Integration tests

The integration tests in `./cmd/zap` exec the `zqp` program to test it. See TestSync for a well commented test.

The tests use Unix sockets for communication. To test the other modes, use the following environment variables for `go test`:

* `ZAP_SERVER_LISTEN=ws://localhost:0`
* `ZAP_SERVER_LISTEN=tcp://localhost:0`

Other tips:

* Run with the race detector: `go install -race ./cmd/zap && ZAP_TEST_STALE_OK=t go test ./cmd/zap`
* Inspect/debug the state of Git repos in temp dirs created by the tests: pass the `-keep-temp` flag to `go test`.
* Test flakiness on macOS? File system notifications (via macOS's kqueue) have a longer delay than on Linux. You may need to increase the sleep durations between the shell commands and operations in integration test cases.

#### Workspace OT tests

Our generalization of OT to WorkspaceOp is complex and is heavily tested in `ot/workspace_test.go`.
