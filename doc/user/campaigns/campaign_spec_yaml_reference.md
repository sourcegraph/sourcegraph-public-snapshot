# Campaign spec YAML reference

<style>.markdown-body h2 { margin-top: 50px; }</style>

[Sourcegraph campaigns](../index.md) use [campaign specs](../index.md#campaign-specs) to define campaigns.

This page is a reference guide to the campaign spec YAML format in which campaign specs are defined. If you're new to YAML and want a short introduction, see "[Learn YAML in five minutes](https://learnxinyminutes.com/docs/yaml/)."

## [`name`](#name)

The name of the campaign, which is unique among all campaigns in the namespace. A campaign's name is case-preserving.

### Examples

```yaml
name: update-go-import-statements
```

```yaml
name: update-node.js
```

## [`description`](#description)

The description of the campaign. It's rendered as Markdown.

### Examples

```yaml
description: This campaign changes all `fmt.Sprintf` calls to `strconv.Iota`.
```

```yaml
description: |
  This campaign changes all imports from
  
  `gopkg.in/sourcegraph/sourcegraph-in-x86-asm`
  
  to
  
  `github.com/sourcegraph/sourcegraph-in-x86-asm`
```

## [`on`](#on)

The set of repositories (and branches) to run the campaign on, specified as a list of search queries (that match repositories) and/or specific repositories.

### Examples

```yaml
on:
  - repositoriesMatchingQuery: lang:go fmt.Sprintf("%d", :[v]) patterntype:structural
  - repository: github.com/sourcegraph/sourcegraph
```

## [`on.repositoriesMatchingQuery`](#on-repositoriesMatchingQuery)

A Sourcegraph search query that matches a set of repositories (and branches). Each matched repository branch is added to the list of repositories that the campaign will be run on.

See "[Code search](../search/index.md)" for more information on Sourcegraph search queries.

### Examples

```yaml
on:
  - repositoriesMatchingQuery: file:README.md -repo:github.com/sourcegraph/src-cli
```

```yaml
on:
  - repositoriesMatchingQuery: lang:typescript file:web const changesetStatsFragment
```

## [`on.repository`](#on-repository)

A specific repository (and branch) that is added to the list of repositories that the campaign will be run on.

A `branch` attribute specifies the branch on the repository to propose changes to. If unset, the repository's default branch is used.

### Examples

```yaml
on:
  - repository: github.com/sourcegraph/src-cli
```

```yaml
on:
  - repository: github.com/sourcegraph/sourcegraph
    branch: 3.19-beta
  - repository: github.com/sourcegraph/src-cli
```

## [`steps`](#steps)

The sequence of commands to run (for each repository branch matched in the `on` property) to produce the campaign's changes.

### Examples

```yaml
steps:
  - run: echo "Hello World!" >> README.md
    container: alpine:3
```

```yaml
steps:
  - run: comby -in-place 'fmt.Sprintf("%d", :[v])' 'strconv.Itoa(:[v])' .go -matcher .go -exclude-dir .,vendor
    container: comby/comby
  - run: gofmt -w ./
    container: golang:1.15-alpine
```

```yaml
steps:
  - run: ./update_dependency.sh
    container: our-custom-image
    env:
      OLD_VERSION: 1.31.7
      NEW_VERSION: 1.33.0
```

## [`steps.run`](#steps-run)

The shell command to run in the container. It can also be a multi-line shell script. The working directory is the root directory of the repository checkout.

## [`steps.container`](#steps-run)

The Docker image used to launch the Docker container in which the shell command is run.

The image has to have either the `/bin/sh` or the `/bin/bash` shell.

It is executed using `docker` on the machine on which the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) is executed. If the image exists locally, that is used. Otherwise it's pulled using `docker pull`.

## [`steps.env`](#steps-env)

Environment variables to set in the environment when running this command.

## [`importChangesets`](#importChangesets)

An array describing which already-existing changesets should be imported from the code host into the campaign.

### Examples

```yaml
importChangesets:
  - repository: github.com/sourcegraph/sourcegraph
    externalIDs: [13323, "13343", 13342, 13380]
  - repository: github.com/sourcegraph/src-cli
    externalIDs: [260, 271]
```

## [`importChangesets.repository`](#importChangesets-repository)

The repository name as configured on your Sourcegraph instance.

## [`importChangesets.externalIDs`](#importChangesets-externalIDs)

The changesets to import from the code host. For GitHub this is the pull request number, for GitLab this is the merge request number, for Bitbucket Server this is the pull request number.

## [`changesetTemplate`](#changesetTemplate)

A template describing how to create (and update) changesets with the file changes produced by the command steps.

This defines what the changesets on the code hosts (pull requests on GitHub, merge requests on Gitlab, ...) will look like.

### Examples

```yaml
changesetTemplate:
  title: Replace equivalent fmt.Sprintf calls with strconv.Itoa
  body: This campaign replaces `fmt.Sprintf("%d", integer)` calls with semantically equivalent `strconv.Itoa` calls
  branch: campaigns/sprintf-to-itoa
  commit:
    message: Replacing fmt.Sprintf with strconv.Iota
  published: false
```

```yaml
changesetTemplate:
  title: Update rxjs in package.json to newest version
  body: This pull request updates rxjs to the newest version, `6.6.2`.
  branch: campaigns/update-rxjs
  commit:
    message: Update rxjs to 6.6.2
  published: false
```

```yaml
changesetTemplate:
  title: Run go fmt over all Go files
  body: Regular `go fmt` run over all our Go files.
  branch: go-fmt
  commit:
    message: Run go fmt
  published: true
```

## [`changesetTemplate.title`](#changesetTemplate-title)

The title of the changeset on the code host.

## [`changesetTemplate.body`](#changesetTemplate-body)

The body (description) of the changeset on the code host. If the code supports Markdown you can use it here.

## [`changesetTemplate.branch`](#changesetTemplate-branch)

The name of the Git branch to create or update on each repository with the changes.

## [`changesetTemplate.commit`](#changesetTemplate-commit)

The Git commit to create with the changes.

## [`changesetTemplate.commit.message`](#changesetTemplate-commit-message)

The Git commit message.

## [`changesetTemplate.published`](#changesetTemplate-published)

Whether to publish the changeset.

An unpublished changeset can be previewed on Sourcegraph by any person who can view the campaign, but its commit, branch, and pull request aren't created on the code host.

A published changeset results in a commit, branch, and pull request being created on the code host.
