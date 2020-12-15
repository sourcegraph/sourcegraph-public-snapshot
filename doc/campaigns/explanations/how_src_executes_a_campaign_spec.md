# How `src` executes a campaign spec

This document is meant to help with debugging and troubleshooting the writing and execution of campaign specs with [Sourcegraph CLI `src`](../../cli/index.md).

It explains what happens under the hood when a user uses applies or previews a campaign spec by running `src campaign apply` or `src campaign preview`.

## Overview

`src campaign apply` and `src campaign preview` execute a campaign spec the same way:

1. [Parse the campaign spec](#parse-campaign-spec)
1. [Resolve the namespace](#resolving-namespace)
1. [Prepare container images](#preparing-container-images)
1. [Resolve repositories](#resolving-repositories)
1. [Executing steps](#executing-steps)
1. [Sending changeset specs](#sending-changeset-specs)
1. [Sending campaign spec](#sending-campaign-spec)
1. [Preview or apply the campaign spec](#preview-or-apply-the-campaign-spec)

The difference is the last step: `src campaign apply` _applies_ the campaign spec where the `src campaign preview` only prints a URL at which you can preview what would happen if you apply it.

The rest of the document explains each step in more detail.

## Parse campaign spec

`src` reads in, parses and validates the campaign spec YAML specified with the `-f` flag.

It validates the campaign spec against its [schema](https://github.com/sourcegraph/src-cli/blob/main/schema/campaign_spec.schema.json) and does some semantic checks to make sure that, for example, `changesetTemplate` is specified if `steps` are specified, or that no feature is used that's not supported by the Sourcegraph instance.

## Resolving namespace

`src` resolves the given namespace in which to apply/preview the campaign spec by sending a GraphQL request to the Sourcegraph instance to fetch the ID for the given namespace name.

If no namespace is specified with `-namespace` (or `-n`) then the currently authenticated user is used as the namespace. See "[Connect to Sourcegraph](../../cli/quickstart.md#connect-to-sourcegraph)" in the CLI docs for details on how to authenticate.

## Preparing container images

If the campaign spec contains `steps`, then for each step `src` checks its `container` image to see whether it's already available locally.

To do that it runs `docker image inspect --format {{.Id}} -- <container-image-name>` to get the specific image ID for the container image.

If that fails with a "No such image" error, `src` tries to pull the image by running `docker image pull <container-image-name>` and then running `docker image inspect --format {{.Id}} -- <container-image-name>` again.

## Resolving repositories

`src` resolves each entry in the campaign spec's `on` property to produce a _unique list of repositories (!)_ in which to execute the campaign spec's `steps`.

With an `on` property like this

```yaml
on:
  - repositoriesMatchingQuery: lang:go fmt.Sprintf("%d", :[v]) patterntype:structural -file:vendor
  - repositoriesMatchingQuery: repohasfile:README.md
  - repository: github.com/sourcegraph/sourcegraph
  - repository: github.com/sourcegraph/automation-testing
    branch: thorstens-test-branch
```

`src` will do the following:

1. For each `repositoriesMatchingQuery` it will:
	1. Send a request to the Sourcegraph API to execute the search query.
	1. Collect each result's _repository_: the ID, the name, the default branch and the current revision of the default branch. If the search result _is a repository result_ (i.e. a search query of `type:repo` only produces repositories) that's used. If it's a file match the file match's repository is used.
	1. _Optional_: if the results are file matches, then their path in the repository is also saved, so that they can be used in the `steps` with [templating](../references/campaign_spec_templating.md).
1. For each `repository` without a `branch` it will:
	1. Send a request to the Sourcegraph API to get the repository's ID, name, its default branch and the current revision of the default branch.
1. For each `repository` _with_ a `branch` it will:
	1. Send a request to the Sourcegraph API to get the repository's ID, and name and the current revision of the specified `branch`.
1. It then creates a _unique_ list of all repositories yielded by the previous three steps by going through all repositories and comparing them, skipping those where no current revision of a branch could be resolved, checking whether they're on a supported code host. If they are on unsupported code hosts and no `-allow-unsupported` flag is given, then a warning is printed and the repositories are not added to the list.

## Executing steps

If a campaign spec contains `steps` then `src` executes the steps _locally_, on the machine on which `src` is run, for _each repository yielded by the previous "[Resolving repositories](#resolving-repositories)" step_.

If `-clear-cache` is _not_ set and it previously executed _the same `steps`_ for the _same repository_ at the _same revision of the base branch_, it will try to use cached results instead of re-executing the steps.

The following is what `src` does _for each repository_:

### 1. Download archive and prepare

1. Download archive of repository. What it does is equivalent to:

    ```
    curl -L -v -X GET -H 'Accept: application/zip' \
      -H 'Authorization: token <THE_SRC_TOKEN>' \
      'http://sourcegraph.example.com/github.com/my-org/my-repo@refs/heads/master/-/raw' \
      --output ~/tmp/my-repo.zip
    ```
2. Unzip archive, e.g. into `~/Library/Caches/sourcegraph/campaigns` (see `src campaign preview -h` for default value of cache directory, overwrite with `-cache`)
3. `cd` into unzipped archive
4. In the unzipped archive directory, create a git repository:
	- Configure `git` to not use local configuration (see [the code for explanations on what each variable does](https://github.com/sourcegraph/src-cli/blob/038180005c9ebf5c0f9e8d3b2eda63c109cea904/internal/campaigns/run_steps.go#L31-L44)):

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

For each step in the campaign specs `steps`:

1. Probe container image (the `container` property of the step) to see whether it has `/bin/sh` or `/bin/bash`
2. Write the step's `run` command to a temp file on the host, e.g. `/tmp-script`
3. Run `chmod 644 /tmp-script`
4. Run the Docker container:

    ```
    docker run --rm --init --workdir /work \
      --mount type=bind,source=/unzipped-archive-locally,target=/work \
      --mount type=bind,source=/tmp-script,target=/tmp-file-in-container \
      --entrypoint /bin/bash -- <IMAGE> /tmp-file-in-container
    ```
5. Add all produced changes to the git index: `git add --all`

### 3. Create final diff

In the unzipped archive:

1. Create a diff by running: `git diff --cached --no-prefix --binary`

### 4. Saving a changeset spec

The produced diff is added to the local cache so that re-executing the same steps in the same repository can be skipped if the base branch did not changed.

The diff is then combined with information about the repository in which the changes have been made (the name and ID of the repository, the revision of its base branch) and together with the `changesetTemplate` turned into a changeset spec: a description of what the changeset should look like.

## Importing changesets

If the campaign spec contains `importChangesets` then `src` goes through the list of `importChangesets` and for each entry it will:

1. Resolve the repository name, trying to get to get an ID, base branch, and revision for the given repository name.
1. Parse the `externalIDs`, checking that they're valid strings or numbers.
1. For each external ID it saves a changeset spec that describes that a changeset with the given external ID, in the given repository, should be imported and tracked in the campaign.

## Sending changeset specs

The previous two steps, "[Executing steps](#executing-steps)" and "[Importing changesets](#importing-changesets)", can produce changeset specs, each one describing either a changeset to create or to import.

These changeset specs are now uploaded to the connected Sourcegraph instance, one request per changeset spec.

Each request yields an ID that uniquely identifies the changeset spec on the Sourcegraph instance. These IDs are used for the next step.

## Sending campaign spec

The IDs of the changeset specs that were created in the previous step, "[Sending changeset specs](#sending-changeset-specs)", are collected into a list and used for the next request with which `src` uploads the campaign spec to the connected Sourcegraph instance.

`src` _creates_ the campaign spec on the Sourcegraph instance, together with the changeset spec IDs, so that the campaign spec fully describes the desired state of a campaign: its name, its description, and which changesets should be created or imported from which repository on which code host.

That request yields an ID that uniquely identifies this expanded version of the campaign spec.

## Preview or apply the campaign spec

If `src campaign apply` was used, then the ID of the campaign is then used to send another request to the Sourcegraph instance, to _apply_ the campaign spec.

If `src campaign preview` was used to execute and create the campaign spec, then a URL is printed, pointing to a preview page on the Sourcegraph instance on which we can see what _would_ happen if we were to apply the campaign spec.
