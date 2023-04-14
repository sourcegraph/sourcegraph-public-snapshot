<style>
.socials {
  display: flex;
  flex-direction: row;
}
.socials a {
  padding: 0.25rem;
  margin: 1rem;
  background: #dddddd;
  border-radius: 0.25rem;
  width: 3.5rem;
  height: 3.5rem;
  display: flex;
  align-items: center;
}
.socials a:hover {
  filter: brightness(0.75);
}
</style>

# Sourcegraph App

Sourcegraph App is a lightweight single-binary version of Sourcegraph for your local machine.

It runs alongside your IDE, to provide a better way to understand, browse, and search both the local you're working on as well as remote code on GitHub, GitLab, etc. - bringing Sourcegraph's powerful code search and IDE navigation straight to your laptop. [Learn more](https://about.sourcegraph.com/app)

<div class="socials">
  <a href="https://discord.com/invite/s2qDtYGnAE"><img alt="Discord" src="discord.svg"></img></a>
  <a href="https://twitter.com/sourcegraph"><img alt="Twitter" src="twitter.svg"></img></a>
  <a href="https://github.com/sourcegraph/app"><img alt="GitHub" src="github.svg"></img></a>
</div>

## Installation

Ensure you have `git` installed / on your path, then install.

**macOS:** via homebrew:

```sh
brew install sourcegraph/app/sourcegraph
```

**Linux:** via <a data-download-name="app-download-linux-deb" href="https://storage.googleapis.com/sourcegraph-app-releases/2023.03.23+209542.7216ba/sourcegraph_2023.03.23+209542.7216ba_linux_amd64.deb">deb pkg</a> installer:

```sh
dpkg -i <file>.deb
```

**Single-binary zip download:**

- <a data-download-name="app-download-mac-zip" href="https://storage.googleapis.com/sourcegraph-app-releases/2023.03.23+209542.7216ba/sourcegraph_2023.03.23+209542.7216ba_darwin_all.zip">macOS (universal)</a>
- <a data-download-name="app-download-linux-zip" href="https://storage.googleapis.com/sourcegraph-app-releases/2023.03.23+209542.7216ba/sourcegraph_2023.03.23+209542.7216ba_linux_amd64.zip" >linux (x64)</a>

## Usage

Start Sourcegraph by running the following in a terminal:

```sh
sourcegraph
```

**Sourcegraph will automatically add any repositories found below the directory you run `sourcegraph` in.**

Your browser should automatically open to http://localhost:3080 - this is the address of the Sourcegraph app.

### (Optional) batch changes & precise code intel

Batch changes and precise code intel require the following optional dependencies be installed and on your PATH:

- The `src` CLI ([installation](https://sourcegraph.com/github.com/sourcegraph/src-cli))
- `docker`

## Tips

### Get help & give feedback

Sourcegraph app is early-stages, if you run into any trouble or have ideas/feedback, we'd love to hear from you!

* [Join our community Discord](https://discord.com/invite/s2qDtYGnAE) for live help/discussion
* [Create a GitHub issue](https://github.com/sourcegraph/app/issues/new)

### Upgrading

#### macOS (app .dmg installer)

Navigate to your Applications directory using Finder, and delete the old version of Sourcegraph. Then [download and run the latest version](https://about.sourcegraph.com/app).

#### macOS (homebrew)

```
brew uninstall sourcegraph && brew update && brew install sourcegraph/app/sourcegraph
```

#### Linux (deb)

We do not have a PPA yet; so you will need to rerun the installation steps above to get the latest .deb version.

### Uninstallation

#### macOS (app .dmg installer)

Navigate to your Applications directory using Finder, and delete the old version of Sourcegraph.

#### macOS (homebrew)

```sh
brew uninstall sourcegraph
```

#### Linux

You can simply delete the `sourcegraph` binary from your system.

### Delete all Sourcegraph data

**Warning:** This will delete _all_ your Sourcegraph data and you will not be able to recover it!

### macOS

```sh
rm -rf $HOME/.sourcegraph-psql
rm -rf $HOME/Library/Application\ Support/sourcegraph-sp
rm -rf $HOME/Library/Caches/sourcegraph-sp
```

### Linux

We respect `$XDG_CACHE_HOME` and `$XDG_CONFIG_HOME`. If not set, we fall back to `$HOME/.cache` and `$HOME/.config`. Thus, you can delete all Sourcegraph data using:

```sh
rm -rf $HOME/.sourcegraph-psql
rm -rf $XDG_CACHE_HOME/sourcegraph-sp
rm -rf $XDG_CONFIG_HOME/sourcegraph-sp
rm -rf $HOME/.cache/sourcegraph-sp
rm -rf $HOME/.config/sourcegraph-sp
```
