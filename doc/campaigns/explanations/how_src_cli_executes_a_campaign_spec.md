# How src-cli executes a campaign spec

## Overview

This document explains what happens under the hood when a user uses the [Sourcegraph CLI `src`](../cli/index.md) to apply or preview a campaign spec with `src campaign apply` or `src campaign preview`. It's meant to help debugging when writing and executing campaign specs.

`src` executes the `steps` in a campaign spec locally before sending up the changeset specs (which include the produced diff) and the campaign spec to the Sourcegraph instance.


## Executing campaign spec steps in a single repository

### 1. Download archive and prepare

1. Download archive of repository. What it does is equivalent to:

    ```
    curl -L -v -X GET -H 'Accept: application/zip' \
      -H 'Authorization: token <THE_SRC_TOKEN>' \
      'http://sourcegraph.example.com/github.com/my-org/my-repo@refs/heads/master/-/raw' \
      --output ~/tmp/my-repo.zip
    ```
2. Unzip archive, e.g. into `~/Library/Caches/sourcegraph/campaigns` (see `src campaign preview -h` for default value of cache dir, overwrite with `-cache`)
3. `cd` into unzipped archive
4. In the unzipped archive directory, create a git repository:
	- Configure `git` to not use local config (see [the code for explanations on what each variable does](https://github.com/sourcegraph/src-cli/blob/038180005c9ebf5c0f9e8d3b2eda63c109cea904/internal/campaigns/run_steps.go#L31-L44)):

    ```
    export GIT_CONFIG_NOSYSTEM=1 \
           GIT_CONFIG=/dev/null \
           GIT_AUTHOR_NAME=Sourcegraph \
           GIT_AUTHOR_EMAIL=campaigns@sourcegraph.com \
           GIT_COMMITTER_NAME=Sourcegraph \
           GIT_COMMITTER_EMAIL=campaigns@sourcegraph.com
    ```
  - Run `git init`
  - Run `git config --local user.name Sourcegraph`
  - Run `git config --local user.email campaigns@sourcegraph.com`
  - Run `git add --force --all`
  - Run `git commit --quiet --all -m sourcegraph-campaigns`

### 2. Run the steps

For each step:

1. Probe container image to see whether it has `bin/sh` or `bin/bash`
2. Write `steps.run` to a temp file, e.g. `/tmp-script`
3. Run `chmod 644 /tmp-script`
4. Run the Docker container:

    ```
    docker run --rm --init --workdir /work \
      --mount type=bind,source=/unzipped-archive-locally,target=/work \
      --mount type=bind,source=/tmp-script,target=/tmp-file-in-container \
      --entrypoint /bin/bash -- <IMAGE> /tmp-file-in-container
    ```
5. Add all the changes, run: `git add --all`

### 3. Create final diff

In the unzipped archive:

1. Create a diff by running: `git diff --cached --no-prefix --binary`
