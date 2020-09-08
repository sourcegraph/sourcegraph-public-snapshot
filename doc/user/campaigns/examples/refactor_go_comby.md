# Refactor Go code using Comby

This campaign rewrites Go statements from

```go
fmt.Sprintf("%d", number)
```

to

```go
strconv.Itoa(number)
```

since they are equivalent.

Since the replacements could change the formatting of the code, it also runs `gofmt` over the repository.

## Campaign spec

```yaml
name: sprintf-to-itoa
description: Run `comby` to replace `fmt.Sprintf("%d", integer)` calls with `strconv.Iota`

# Find all repositories that contain the `fmt.Sprintf` statement using structural search
on:
  - repositoriesMatchingQuery: lang:go fmt.Sprintf("%d", :[v]) patterntype:structural

steps:
  - run: comby -in-place 'fmt.Sprintf("%d", :[v])' 'strconv.Itoa(:[v])' .go -matcher .go -exclude-dir .,vendor
    container: comby/comby
  - run: gofmt -w ./
    container: golang:1.15-alpine

# Describe the changeset (e.g., GitHub pull request) you want for each repository.
changesetTemplate:
  title: Replace equivalent fmt.Sprintf calls with strconv.Itoa
  body: This campaign replaces `fmt.Sprintf("%d", integer)` calls with semantically equivalent `strconv.Itoa` calls
  branch: campaigns/sprintf-to-itoa # Push the commit to this branch.
  commit:
    message: Replacing fmt.Sprintf with strconv.Iota
  published: false
```

## Instructions

1. Save the campaign spec above as `YOUR_CAMPAIGN_SPEC.campaign.yaml`.
1. Create a campaign from the campaign spec by running the following [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) command:

    <pre><code>src campaign preview -f <em>YOUR_CAMPAIGN_SPEC.campaign.yaml</em> -namespace USERNAME_OR_ORG</code></pre>

1. Open the preview URL that the command printed out.
1. Examine the preview. Confirm that the changesets are the ones you intended to track. If not, edit the campaign spec and then rerun the command above.
1. Click the **Create campaign** button.
