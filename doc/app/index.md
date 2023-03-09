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

It runs alongside your IDE, to provide a better way to understand, browse, and search both the local you're working on as well as remote code on GitHub, GitLab, etc. - bringing Sourcegraph's powerful code search and IDE navigation straight to your laptop.

<div class="socials">
  <a href="https://discord.com/invite/s2qDtYGnAE"><img alt="Discord" src="discord.svg"></img></a>
  <a href="https://twitter.com/sourcegraph"><img alt="Twitter" src="twitter.svg"></img></a>
  <a href="https://github.com/sourcegraph/app"><img alt="GitHub" src="github.svg"></img></a>
</div>

## Pre-release

This is a pre-release / early access version of what will be presented at our Starship event on Mar 23rd. We're still improving things, so [let us know if you encounter any issues](https://github.com/sourcegraph/app/issues/new).

For news and updates, be sure to [follow us on Twitter](https://twitter.com/sourcegraph) and [join our Discord](https://discord.com/invite/s2qDtYGnAE).

## Installation

Ensure you have `git` installed / on your path, then install.

**macOS:** via homebrew:

```sh
brew install sourcegraph/app/sourcegraph
```

**Linux:** via [deb pkg](https://storage.googleapis.com/sourcegraph-app-releases/2023.03.23+205301.ca3646/sourcegraph_2023.03.23+205301.ca3646_linux_amd64.deb) installer:

```sh
dpkg -i <file>.deb
```

**Single-binary zip download:**

* [macOS (universal)](https://storage.googleapis.com/sourcegraph-app-releases/2023.03.23+205301.ca3646/sourcegraph_2023.03.23+205301.ca3646_darwin_all.zip)
* [linux (x64)](https://storage.googleapis.com/sourcegraph-app-releases/2023.03.23+205301.ca3646/sourcegraph_2023.03.23+205301.ca3646_linux_amd64.zip)

## Usage

Start Sourcegraph by running the following in a terminal:

```sh
sourcegraph
```

**Sourcegraph will automatically add any repositories found below the directory you run `sourcegraph` in.**

Your browser should automatically open to http://localhost:3080 - this is the address of the Sourcegraph app.

## Optional - batch changes & precise code intel

Batch changes and precise code intel require the following optional dependencies be installed and on your PATH:

* The `src` CLI ([installation](https://github.com/sourcegraph/src-cli))
* `docker`

### Troubleshooting

Since the Sourcegraph app is early-stages, you may run into issues. If you do, please [let us know](https://github.com/sourcegraph/app/issues/new)!

### Known issues

#### macOS .zip download issues

macOS binaries are not yet code-signed, so you may need to right click on it -> open. If you use Homebrew, this is not an issue.

### Upgrading

**On macOS:** upgrade using Homebrew:

```
brew update && brew upgrade sourcegraph/app/sourcegraph
```

**On Linux:** we do not have a PPA yet, so you will need to rerun the installation steps above to get the latest .deb version.

## Feedback

Please let us know what you think in our [Discord](https://discord.com/invite/s2qDtYGnAE) or tweet [@sourcegraph](https://twitter.com/sourcegraph)!

_Sourcegraph employees:_ join `#dogfood-app` in Slack!
