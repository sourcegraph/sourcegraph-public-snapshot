# Campaign spec templating

<style>.markdown-body h2 { margin-top: 50px; }</style>

[Sourcegraph campaigns](../index.md) use [campaign specs](../index.md#campaign-specs) to define campaigns.

This page is a reference guide to the **experimental** templating support in campaign spec YAML files, that was added in Sourcegraph **3.22**.

> NOTE: This feature is **experimental** and might change in the future.

## Overview

In Sourcegraph 3.22 with src-cli 3.21.5 and later it's possible to use template variables in campaign spec YAML files.

```yaml
on:
  - repositoriesMatchingQuery: lang:go fmt.Sprintf("%d", :[v]) patterntype:structural -file:vendor

steps:
  - run: comby -in-place 'fmt.Sprintf("%d", :[v])' 'strconv.Itoa(:[v])' ${{ repository.search_result_paths }}
    container: comby/comby
  - run: goimports -w ${{ previous_step.modified_files }}
    container: unibeautify/goimports
```

## Fields with template support

Template variables are supported in the following fields:

- [`steps.run`](campaign_spec_yaml_reference.md#steps-run)
- [`steps.env`](campaign_spec_yaml_reference.md#steps-run) values
- [`steps.files`](campaign_spec_yaml_reference.md#steps-run) values

## Template variables

The following template variables are available:

- `${{ repository.search_result_paths }}`
- `${{ repository.name }}`
- `${{ previous_step.modified_files }}`
- `${{ previous_step.added_files }}`
- `${{ previous_step.deleted_files }}`
- `${{ previous_step.stdout }}`
- `${{ previous_step.stderr}}`

## Template helper functions

- `${{ join repository.search_result_paths "\n" }}`
- `${{ split repository.name "/" }}`

## Example campaign specs

```yaml
steps:
  # Run comby over the search results in each repository:
  - run: comby -in-place -config /tmp/go-sprintf.toml -f ${{ join repository.search_result_paths "," }}
    container: comby/comby
    files:
      # Create files inside the container by specifying path and content here:
      /tmp/go-sprintf.toml: |
        [sprintf_to_strconv]
        match='fmt.Sprintf("%d", :[v])'
        rewrite='strconv.Itoa(:[v])'

  - run: echo "comby found the following problems:" >> CHANGELOG.md && cat /tmp/comby-output.txt >> CHANGELOG.md
    container: alpine:3
    files:
      # files also support templating:
      /tmp/comby-output.txt: ${{ previous_step.stdout }}

  - run: echo $MY_MODIFIED_FILES >> modified_files.txt
    container: alpine:3
    env:
      # env vars also support templating:
      MY_MODIFIED_FILES: ${{ join previous_step.modified_files "\n" }}
```

```yaml
# Step 1: build the `ruplacer` Docker image
#   $ cat Dockerfile
#   FROM rust
#   RUN cargo install ruplacer
#
#   $ docker build -t ruplacer .
#
#
#   (Why use ruplacer? Because it supports `--subvert` which allows us to
#   replace `camelCase`, `snake_case`, `ThisCase`, `nocase`.)
#
# Step 2: use the src-cli prototype with templating support
#
#   $ cd src-cli && git fetch && git checkout mrnugget/templates-and-files
#   $ go build ./cmd/src -o ~/bin/src
#
# Step 3: update the `repositoriesMatchingQuery` to include or exclude file types.
#
# Step 4: run this campaign
#
#   $ src campaign preview -f update-language.campaign.yaml
name: update-language
description: This campaign changes occurrences of whitelist & blacklist to allowlist & denylist.

on:
  - repositoriesMatchingQuery: whitelist OR blacklist -file:scss$ -file:html$ repo:github.com/sourcegraph/sourcegraph

steps:
  - run: |
      cat /tmp/search-results | while read file;
      do
        ruplacer --subvert whitelist allowlist --go ${file} || echo "nothing to replace";
        ruplacer --subvert blacklist denylist --go ${file} || echo "nothing to replace";
      done
    container: ruplacer
    files:
      /tmp/search-results: ${{ join repository.search_result_paths "\n" }}

# Describe the changeset (e.g., GitHub pull request) you want for each repository.
changesetTemplate:
  title: Replace usage of whitelist & blacklist with allowlist & denylist
  body: This replaces usages 
  branch: campaigns/allowlist-denylist
  commit:
    message: Use allowlist/denylist in wording
  published: false
```
