# Index
- [Golang](#Golang)
- [TypeScript and JavaScript](#TypeScript-and-JavaScript)

# Golang

## Manual indexing

Install [lsif-go](https://github.com/sourcegraph/lsif-go) with `go get github.com/sourcegraph/lsif-go/cmd/lsif-go` and ensure it's on your path.

Install [src-cli](https://github.com/sourcegraph/src-cli) with
```console
curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
```
OSX users can replace `linux` with `darwin` in the URL, and by replacing the endpoint with the Sourcegraph instance you intend to upload to you can guarantee tool compatibility. See the above linked repo for more info and Windows instructions.

Now `cd` into your Go project's root (where the go.mod file lives, if you have one) and run the following:
```console
lsif-go # generates a file named dump.lsif
# for private instances
src -endpoint=$SRC_ENDPOINT lsif upload
# for public instances
src lsif upload -github-token=$GITHUB_TOKEN
```
To upload LSIF data to the public sourcegraph.com instance, you must prove ownership of the repository with a GitHub token.

The upload command will provide a URL you can visit to see the upload's status, and when it's done you can visit the repo and check out the difference in code navigation quality! To troubleshoot issues, visit the more in depth [LSIF quickstart](./lsif_quickstart.md) guide and check out the documentation for the `lsif-go` and `src-cli` tools.

## Automated indexing

We provide the docker images `sourcegraph/lsif-go` and `sourcegraph/src-cli` to make automating this process in your favorite CI framework as easy as possible. Note that the `lsif-go` image bundles `src-cli` so the second image may not be necessary.

Here's some examples in a couple popular frameworks:

### GitHub Actions
```yaml
jobs:
  lsif-go:
    runs-on: ubuntu-latest
    # TODO: pin that container version!
    container: sourcegraph/lsif-go
    steps:
      - uses: actions/checkout@v1
      - name: Generate LSIF data
        run: lsif-go
      - name: Upload LSIF data
        run: src lsif upload -github-token=${{ secrets.GITHUB_TOKEN }}
```

# TypeScript and JavaScript

## Manual indexing

Install [lsif-node](https://github.com/sourcegraph/lsif-node) with `npm install -g @sourcegraph/lsif-tsc` or your favorite method of installing npm packages.

Install [src-cli](https://github.com/sourcegraph/src-cli) with
```console
curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
```
OSX users can replace `linux` with `darwin` in the URL, and by replacing the endpoint with the Sourcegraph instance you intend to upload to you can guarantee tool compatibility. See the above linked repo for more info and Windows instructions.

Now `cd` into your project's root (where the package.json/tsconfig.json) and run the following:
```console
# for typescript projects
lsif-tsc -p .
# for javascript projects
lsif-tsc **/*.js --allowJs --checkJs

# for private instances
src -endpoint=$SRC_ENDPOINT lsif upload
# for public instances
src lsif upload -github-token=$GITHUB_TOKEN
```
Check out the tool documentation if you're having trouble getting the `lsif-tsc` to work. It accepts any options `tsc` does, so it shouldn't be too hard to get it running on your project. To upload LSIF data to the public sourcegraph.com instance, you must prove ownership of the repository with a GitHub token.

The upload command will provide a URL you can visit to see the upload's status, and when it's done you can visit the repo and check out the difference in code navigation quality! To troubleshoot issues, visit the more in depth [LSIF quickstart](./lsif_quickstart.md) guide and check out the documentation for the `lsif-node` and `src-cli` tools.

## Automated indexing

We provide the docker images `sourcegraph/lsif-node` and `sourcegraph/src-cli` to make automating this process in your favorite CI framework as easy as possible. Note that the `lsif-node` image bundles `src-cli` so the second image may not be necessary.

Here's some examples in a couple popular frameworks:

### GitHub Actions
```yaml
jobs:
  lsif-node:
    runs-on: ubuntu-latest
    # TODO: pin that container version!
    container: sourcegraph/lsif-node
    steps:
      - uses: actions/checkout@v1
      - name: Install dependencies
        run: yarn
      - name: Generate LSIF data
        run: lsif-tsc -p .
      - name: Upload LSIF data
        run: src lsif upload -github-token=${{ secrets.GITHUB_TOKEN }}
```
Note that if you need to install your dependencies in a custom container, you can use our containers as github actions. Try these steps instead:
```yaml
jobs:
  lsif-node:
    runs-on: ubuntu-latest
    # TODO: pin that container version!
    container: my-awesome-container
    steps:
      - uses: actions/checkout@v1
      - name: Install dependencies
        run: <install dependencies>
      - name: Generate LSIF data
        uses: sourcegraph/lsif-node
        with:
          args: lsif-tsc -p .
      - name: Upload LSIF data
        uses: sourcegraph/src-cli
        with:
          args: src lsif upload -github-token=${{ secrets.GITHUB_TOKEN }}
```
