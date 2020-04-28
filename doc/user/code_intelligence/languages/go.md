# Go LSIF Indexing

This guide is meant to provide specific instructions to get you producing index data in LSIF as quickly as possible. The [LSIF quick start](../lsif_quickstart.md) and [CI configuration](../adding_lsif_to_workflows.md) guides provide more in depth descriptions of each step and a lot of helpful context that we haven't duplicated in each language guide.

## Manual indexing

1. Install [lsif-go](https://github.com/sourcegraph/lsif-go) with `go get github.com/sourcegraph/lsif-go/cmd/lsif-go` and ensure it's on your path.

1. Install the [Sourcegraph CLI](https://github.com/sourcegraph/src-cli) with
   ```
   curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
   chmod +x /usr/local/bin/src
   ```
   - **macOS**: replace `linux` with `darwin` in the URL
   - **Windows**: visit [the CLI's repo](https://github.com/sourcegraph/src-cli) for further instructions

1. `cd` into your Go project's root (where the go.mod file lives, if you have one) and run:
   ```
   lsif-go # generates a file named dump.lsif
   ```

1. Upload the data to a Sourcegraph instance with
   ```
   # for private instances
   src -endpoint=<your sourcegraph endpoint> lsif upload
   # for public instances
   src lsif upload -github-token=<your github token>
   ```
   Visit the [LSIF quickstart](../lsif_quickstart.md) for more information about the upload command.

The upload command will provide a URL you can visit to see the upload's status, and when it's done you can visit the repo and check out the difference in code navigation quality! To troubleshoot issues, visit the more in depth [LSIF quickstart](../lsif_quickstart.md) guide and check out the documentation for the `lsif-go` and `src-cli` tools.

## Automated indexing

We provide the docker images `sourcegraph/lsif-go` and `sourcegraph/src-cli` to make automating this process in your favorite CI framework as easy as possible. Note that the `lsif-go` image bundles `src-cli` so the second image may not be necessary.

Here's some examples in a couple popular frameworks, just substitute the indexer and upload commands with what works for your project locally. If you end up implementing this in a different framework, feel free to edit this page with instructions!

### GitHub Actions
```yaml
jobs:
  lsif-go:
    runs-on: ubuntu-latest
    container: sourcegraph/lsif-go:latest
    steps:
      - uses: actions/checkout@v1
      - name: Generate LSIF data
        run: lsif-go
      - name: Upload LSIF data
        run: src lsif upload -github-token=${{ secrets.GITHUB_TOKEN }}
```

### CircleCI
```yaml
jobs:
  lsif-go:
    docker:
      - image: sourcegraph/lsif-go:latest
    steps:
      - checkout
      - run: lsif-go
      - run: src lsif upload -github-token=<<parameters.github-token>>

workflows:
  lsif-node:
    jobs:
      - lsif-node
```

### Travis CI
```yaml
services:
  - docker

jobs:
  include:
    - stage: lsif-go
      script:
      - |
        docker run --rm -v $(pwd):/src -w /src sourcegraph/lsif-go:latest /bin/sh -c \
          "lsif-go; src lsif upload -github-token=$GITHUB_TOKEN"
```
