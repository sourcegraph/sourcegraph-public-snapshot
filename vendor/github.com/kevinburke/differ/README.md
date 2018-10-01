# differ

Differ makes it easy to run a command and error if it generated a change in a
git worktree. You can use this in tests or the build process to verify that
a given build step was run correctly. For example you may want to verify that
all files in a Go project have run `go fmt`. Run:

```
differ go fmt ./...
```

This will execute `go fmt ./...` and error if it modifies any file tracked by
Git.

Other uses:

- Restore and revendor all vendored libraries and error if a git diff is
generated.
- Check whether new CSS files have been generated from SCSS, HTML files from
  Markdown, JS files from Coffeescript, or any other compilation step.

## Usage

Run the same command you would usually run but put `differ` before it, for
example:

```
differ go generate ./...
```

differ will exit with a non-zero return code if:

- your command exits with an error

- "git status" errors, for example if you run it in a directory that is not
  a Git repository.

- "git status" says that there are untracked or modified files present

## Installation

Find your target operating system (darwin, windows, linux) and desired bin
directory, and modify the command below as appropriate:

    curl --silent --location --output /usr/local/bin/differ https://github.com/kevinburke/differ/releases/download/0.4/differ-linux-amd64 && chmod 755 /usr/local/bin/differ

On Travis, you may want to create `$HOME/bin` and write to that, since
/usr/local/bin isn't writable with their container-based infrastructure.

The latest version is 0.4.

If you have a Go development environment, you can also install via source code:

    go get -u github.com/kevinburke/differ
