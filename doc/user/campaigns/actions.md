# Actions

## What are actions?

The `src` CLI offers the ability to _execute actions_ with the `src action exec` command to **produce a set of patches**, one per repository.

These **patches can then be turned into _changesets_ (pull requests)** on the code hosts on which the repositories are hosted by creating a campaign (see "[Creating a campaign from patches](./creating_campaign_from_patches.md)").

An _action_ is made up of two things:

- A Sourcegraph search query (the `scopeQuery`).
- A series of steps to be executed in each repository yielded by the search query.

Here is an example definition of an action:

```json
{
  "scopeQuery": "lang:go gopkg.in\/inconshreveable\/log15.v2",
  "steps": [
    {
      "type": "docker",
      "image": "comby/comby",
      "args": [
        "-in-place",
        "import (:[before]\"gopkg.in/inconshreveable/log15.v2\":[after])",
        "import (:[before]\"github.com/inconshreveable/log15\":[after])",
        ".go",
        "-matcher",
        ".go",
        "-d", "/work",
        "-exclude-dir", ".,vendor"
      ]
    },
    {
      "type": "command",
      "args": ["goimports", "-w", "."]
    }
  ]
}
```

This action uses [Comby](https://comby.dev) to update a Go import path.

The `"scopeQuery"` yields every repository in which the old import path is mentioned in a Go file.

The first step, of type `"docker"`, executes the Docker image `comby/comby` (with each repository mounted under `/work`) to rewrite the import path from `gopkg.in/inconshreveable/log15.v2` to `github.com/inconshreveable/log15`.

The second step, a `"command"`, then runs [`goimports`](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) to ensure that updating the import paths worked and that the code is correctly formatted.

Since it is a `"command"`, the `goimports` step doesn't use a Docker container, but instead runs the `goimports` executable in a temporary directory on the machine on which the `src` CLI is executed.

## Requirements

To execute actions with the `src` CLI the following is required:

- `src` CLI, setup to point to your Sourcegraph instance.
- [git](https://git-scm.com/).
- [Docker](https://www.docker.com/), if you want to execute Docker containers.

## Defining an action

An action definition is a JSON file ([JSON schema for actions](https://raw.githubusercontent.com/sourcegraph/src-cli/master/schema/actions.schema.json)) and needs to specify:

- `"scopeQuery"` - a Sourcegraph search query to generate a list of repositories over which to run the action.
- `"steps"` - a list of action steps to execute in each repository.

A single step can either be a of type `"command"`, which means the step is executed on the machine on which `src actions exec` is executed, or it can be of type `"docker"` which then runs a container in which the repository is mounted.

To create a new, empty action definition you can use the following helper command:

```
$ src actions create
```

That will create a new `action.json` file in the current directory (see `src actions create -h` for how specify a filename) with a reference to the [JSON schema](https://raw.githubusercontent.com/sourcegraph/src-cli/master/schema/actions.schema.json).

### The scope query

The scope query, specified as `"scopeQuery"` in an action definition, is a [Sourcegraph search](../search/index.md) query that's executed to yield a list of repositories in which to execute an action.

It doesn't have to use `type:repo` to only search for repositories, because `src action exec` will construct a _unique list of repositories_ associated with each search result. If the scope query, for example, yields eight search results in three different repositories, `src action exec` will execute the action in these three repositories.

Examples:

- `lang:go fmt.Sprintf` returns repositories in which `fmt.Sprintf` appears in a Go file.
- `lang:go fmt.Sprintf repo:github.com/my-org/[a-c]` returns repositories in which `fmt.Sprintf` appears in a Go file, if they are part of the `my-org` GitHub organization and their name begins with `a`, `b` or `c`.
- `repohasfile:yarn.lock` returns repositories in which a file called `yarn.lock` exists.

See "[Search query syntax](../search/queries.md)" and "[Search examples](../search/examples.md)" for more information.

If you want to see which repositories are yielded by a `"scopeQuery"` in an action definition without execution the action, use `src action scope-query`:

```
$ src action scope-query -f my-action-definition.json
```

This will return a list of repositories.

Since action definitions are JSON files and require the `"scopeQuery"` to be escaped, it often helps to use `src action scope-query` in combination with the `-v` flag to see exactly which query is sent to the Sourcegraph instance:

```
$ src -v action scope-query -f my-action-definition.json
```

### Docker steps

A Docker step specification requires three attributes: `"type"`, `"image"` and `"args"`.

The `"type"` is `"docker"`.

`"image"` is the Docker image that is executed.

`"args"` is a list of arguments to be passed to `docker run`.

Here is an example Docker step in an action definition:

```json
{
  "type": "docker",
  "image": "alpine:3",
  "args": ["sh", "-c", "find /work -iname '*.txt' -type f | xargs -n 1 sed -i s/this/that/g"]
}
```

This step will effectively execute the following for each repository yielded by the scope query:

```
docker run -it --rm --workdir /work --mount type=bind,source=<REPOSITORY_PATH>,target=/work -- alpine:3 \
  sh -c 'find /work -iname '*.txt' -type f | xargs -n 1 sed -i s/this/that/g'
```

Note the `<REPOSITORY_PATH>` placeholder: when executing an action, each repository yielded by the scope query is extracted into a temporary directory, which gets reused for all steps in an action.

That temporary directory is then mounted into each Docker container under `/work`. That's why the `find` command in the example searches in `/work`.

If you need more directories (outside of the repository) to be persisted across multiple `"docker"` steps, you can use the `"cacheDirs"` property of a step definition. Example:

```json
{
  "scopeQuery": "lang:go repo:sourcegraph/sourcegraph$",
  "steps": [
    {
      "type": "docker",
      "image": "alpine:3",
      "args": ["sh", "-c", "echo 'hello from cache 1' > /cache1/hello.txt && echo 'hello from cache 2' > /cache2/hello.txt"],
      "cacheDirs": ["/cache1", "/cache2"]
    },
    {
      "type": "docker",
      "image": "alpine:3",
      "args": ["sh", "-c", "cp /cache1/hello.txt /work/hello_cache1.txt && cp /cache2/hello.txt /work/hello_cache2.txt"],
      "cacheDirs": ["/cache1", "/cache2"]
    }
  ]
}
```

Note the `cacheDirs` properties used by each step. For every entry in `cacheDirs`, `src action exec` will create a temporary directory that persists across all steps and which is then mounted under the specified name into the container.

This can be used, for example, to cache package manager installations across multiple steps.

### Command steps

A command step specification requires two attributes: `"type"` and `"args"`.

The `"type"` is `"command"`.

`"args"` is a list of consisting of the command to be executed and its arguments.

Example:

```json
{
  "type": "command",
  "args": ["sed", "-i", "", "s/this/that/g README.md"]
}
```

This will execute `sed -i '' s/this/that/g README.md` _on the machine on which `src action exec` is being executed_. There are no containers involved.

The current working directory for each `"command"` step is the root of each repository (extracted into a temporary directory on the machine on which `src action exec` is executed) yielded by the scope query.

## Executing

After creating an action definition and saving it to a file, it can be executed by running `src action exec`:

```
src action exec -f my-action-definition.json
```

What this does is the following:

1. Send the `"scopeQuery"` to the configured Sourcegraph instance and turn the results into a unique list of repositories.
1. Download a ZIP archive of each repository in that list and, _on the machine on which `src action exec` is executed_, extract it to a local temporary directory in `/tmp`.
1. Execute the action for each repository in parallel, with each step in an action being executed sequentially on the same copy of the repository.
1. Produce a patch for each repository by creating _a diff between a fresh copy of the repository and the directory in which the action ran_.

Run `src action exec -h` to get a complete overview of which flags are accepted by `src action exec`, but noteworthy are:

* `-keep-logs` causes `src action exec` to redirect STDOUT/STDERR of each step to a log file and not clean that up after execution has finished.
* `-j` specifies the number of parallel jobs, where one job is the action being executed in a single repository.
* `-clear-cache` clears the cache before executing an action.
* `-create-patchset` creates a patchset out of the produced patches, which can then be used to create a campaign.

If you are experimenting with actions and want to get more insight, pass the `-v` flag to `src` to see the what each step prints to STDOUT and STDOUT:

```
src -v action exec -f my-action-definition.json
```

### Where to run `src action exec`

The steps of an action are executed on the machine where the `src` CLI is installed and executed.

For most usecases that involve a lot of repositories and action steps that require a lot of resources, we recommend that `src` CLI should be run on a Linux machine with considerable CPU, RAM, and network bandwidth to reduce the execution time. Putting this machine in the same network as your Sourcegraph instance will also improve performance.

Another factor affecting execution time is the number of jobs executed in parallel, which is by default the number of cores on the machine. This can be adjusted using the `-j` parameter.

## Creating patchsets

In order to create a patchset out of the patches produced by executing an action, pipe the output to `src campaign patchset create-from-patches`:

```
src actions exec -f my-action-definition.json | src campaign patchset create-from-patches
```

Or pass the `-create-patchset` flag directly to `src action exec`:

```
src actions exec -f my-action-definition.json -create-patchset
```

If the action failed to execute in one of the repositories, `src action exec` will ask for confirmation to create a patchset anyway, if `-create-patchset` is given. In order to _always create a patchset_, without asking for confirmation, use the `-force-create-patchset` flag:

```
src actions exec -f my-action-definition.json -force-create-patchset
```
