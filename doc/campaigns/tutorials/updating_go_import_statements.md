# Updating Go import statements using Comby

<style>
.markdown-body pre.chroma {
  font-size: 0.75em;
}
</style>

### Introduction

This campaign rewrites Go import paths for the `log15` package from `gopkg.in/inconshreveable/log15.v2` to `github.com/inconshreveable/log15` using [Comby](https://comby.dev/).

It can handle single-package import statements like this one:

```go
import "gopkg.in/inconshreveable/log15.v2"
```

Single-package imports with an alias:


```go
import log15 "gopkg.in/inconshreveable/log15.v2"
```

And multi-package import statements with our without an alias:

```go
import (
	"io"

	"github.com/pkg/errors"
	"gopkg.in/inconshreveable/log15.v2"
)
```

### Prerequisites

We recommend that use the latest version of Sourcegraph when working with campaigns and that you have a basic understanding of how to create campaign specs and run them. See the following documents for more information:

1. ["Quickstart"](../quickstart.md)
1. ["Introduction to campaigns"](../explanations/introduction_to_campaigns.md)

### Create the campaign spec

Save the following campaign spec YAML as `update-log15-import.campaign.yaml`:

```yaml
name: update-log15-import
description: This campaign updates Go import paths for the `log15` package from `gopkg.in/inconshreveable/log15.v2` to `github.com/inconshreveable/log15` using [Comby](https://comby.dev/)

# Find all repositories that contain the import we want to change.
on:
  - repositoriesMatchingQuery: lang:go gopkg.in/inconshreveable/log15.v2

# In each repository
steps:
  # we first replace the import when it's part of a multi-package import statement
  - run: comby -in-place 'import (:[before]"gopkg.in/inconshreveable/log15.v2":[after])' 'import (:[before]"github.com/inconshreveable/log15":[after])' .go -matcher .go -exclude-dir .,vendor
    container: comby/comby
  # ... and when it's a single import line.
  - run: comby -in-place 'import:[alias]"gopkg.in/inconshreveable/log15.v2"' 'import:[alias]"github.com/inconshreveable/log15"' .go -matcher .go -exclude-dir .,vendor
    container: comby/comby

# Describe the changeset (e.g., GitHub pull request) you want for each repository.
changesetTemplate:
  title: Update import path for log15 package to use GitHub
  body: Updates Go import paths for the `log15` package from `gopkg.in/inconshreveable/log15.v2` to `github.com/inconshreveable/log15` using [Comby](https://comby.dev/)
  branch: campaigns/update-log15-import # Push the commit to this branch.
  commit:
    message: Fix import path for log15 package
  published: false
```

### Create the campaign

1. In your terminal, run this command:

    <pre>src campaign preview -f update-log15-import.campaign.yaml -namespace <em>USERNAME_OR_ORG</em></pre>

    > The `namespace` is either your Sourcegraph username or the name of a Sourcegraph organisation under which you want to create the campaign. If you're not sure what to choose, use your username.
1. Wait for it to run and compute the changes for each repository.
1. Open the preview URL that the command printed out.
1. Examine the preview. Confirm that the changesets are the ones you intended to track. If not, edit the campaign spec and then rerun the command above.
1. Click the **Apply spec** button to create the campaign.
1. Feel free to then publish the changesets (i.e. create pull requests and merge requests) by [modifying the `published` attribute in the campaign spec](../references/campaign_spec_yaml_reference.md#changesettemplate-published) and re-running the `src campaign preview` command.
