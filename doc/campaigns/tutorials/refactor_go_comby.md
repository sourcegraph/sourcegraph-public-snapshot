# Refactor Go code using Comby

<style>
.markdown-body pre.chroma {
  font-size: 0.75em;
}
</style>

### Introduction

This campaign uses Sourcegraph's [structural search](../../code_search/reference/structural.md) and [Comby](https://comby.dev/) to rewrite Go statements from

```go
fmt.Sprintf("%d", number)
```

to

```go
strconv.Itoa(number)
```

The statements are semantically equivalent, but the latter is clearer.

Since the replacements could require importing the `strconv` package, it uses [`goimports`](https://godoc.org/golang.org/x/tools/cmd/goimports) to update the list of imported packages in all `*.go` files.

### Prerequisites

We recommend that use the latest version of Sourcegraph when working with campaigns and that you have a basic understanding of how to create campaign specs and run them. See the following documents for more information:

1. ["Quickstart"](../quickstart.md)
1. ["Introduction to campaigns"](../explanations/introduction_to_campaigns.md)

### Create the campaign spec

Save the following campaign spec YAML as `sprintf-to-itoa.campaign.yaml`:

```yaml
name: sprintf-to-itoa
description: |
  This campaign uses [Comby](https://comby.dev) to replace `fmt.Sprintf` calls
  in Go code with the equivalent but clearer `strconv.Iota` call.

on:
  - repositoriesMatchingQuery: lang:go fmt.Sprintf("%d", :[v]) patterntype:structural

steps:
  - run: comby -in-place 'fmt.Sprintf("%d", :[v])' 'strconv.Itoa(:[v])' .go -matcher .go -exclude-dir .,vendor
    container: comby/comby
  - run: goimports -w $(find . -type f -name '*.go' -not -path "./vendor/*")
    container: unibeautify/goimports

changesetTemplate:
  title: Replace fmt.Sprintf with equivalent strconv.Itoa
  body: This campaign replaces `fmt.Sprintf` with `strconv.Itoa`
  branch: campaigns/sprintf-to-itoa
  commit:
    message: Replacing fmt.Sprintf with strconv.Iota
  published: false
```

### Create the campaign

1. In your terminal, run this command:

    <pre>src campaign preview -f sprintf-to-itoa.campaign.yaml -namespace <em>USERNAME_OR_ORG</em></pre>

    > The `namespace` is either your Sourcegraph username or the name of a Sourcegraph organisation under which you want to create the campaign. If you're not sure what to choose, use your username.
1. Wait for it to run and compute the changes for each repository.
1. Open the preview URL that the command printed out.
1. Examine the preview. Confirm that the changesets are the ones you intended to track. If not, edit the campaign spec and then rerun the command above.
1. Click the **Apply spec** button to create the campaign.
1. Feel free to then publish the changesets (i.e. create pull requests and merge requests) by [modifying the `published` attribute in the campaign spec](../references/campaign_spec_yaml_reference.md#changesettemplate-published) and re-running the `src campaign preview` command.
