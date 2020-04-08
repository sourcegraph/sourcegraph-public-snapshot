# Campaigns

>NOTE: **IMPORTANT** If you are using Sourcegraph 3.14 use [the 3.14 documentation instead](https://docs.sourcegraph.com/@3.14/user/campaigns)

> **Campaigns are currently available in private beta for select enterprise customers.** (This feature was previously known as "Automation".)

## What are Campaigns?

Campaigns are part of [Sourcegraph code change management](https://about.sourcegraph.com/product/code-change-management) and let you make large-scale code changes across many repositories and different code hosts.

You provide the code to make the change and Campaigns provide the plumbing to turn it into a large-scale code change campaign and monitor its progress.

## Getting started with Campaigns

1. Make sure that the Campaigns feature flag is enabled: [Configuration](#Configuration)
1. Optional, but highly recommended for optimal syncing performance between your code host and Sourcegraph, setup the webhook integration:
  * GitHub: [Configuring GitHub webhooks](https://docs.sourcegraph.com/admin/external_service/github#webhooks).
  * Bitbucket Server: [Setup the `bitbucket-server-plugin`](https://github.com/sourcegraph/bitbucket-server-plugin), [create a webhook](https://github.com/sourcegraph/bitbucket-server-plugin/blob/master/src/main/java/com/sourcegraph/webhook/README.md#create) and configure the `"plugin"` settings for your [Bitbucket Server code host connection](https://docs.sourcegraph.com/admin/external_service/bitbucket_server#configuration).
1. Setup the `src` CLI on your machine: [Installation and setup instructions](https://github.com/sourcegraph/src-cli/#installation)
1. Create your first campaign: [Creating campaigns](#creating-campaigns)
1. Take a look at example campaigns: [Example campaigns](#example-campaigns)

## Configuration

In order to use Campaigns, a site-admin of your Sourcegraph instance must enable it in the site configuration settings e.g. `sourcegraph.example.com/site-admin/configuration`

```json
{
  "experimentalFeatures": {
      "automation": "enabled"
  }
}
```

Without any further configuration, campaigns are **only accessible to site-admins.** If you want to grant read-only access to non-site-admins, use the following site configuration setting:

```json
{
  "campaigns.readAccess.enabled": true
}
```

## Creating campaigns

There are two types of campaigns:

### Campaigns created from a set of patches

When a Campaign is created from a set of patches, one per repository, Sourcegraph will create changesets (pull requests) on the associated code hosts and track their progress in the newly created campaign, where you can manage them.

With the `src` CLI tool, you can not only create the campaign from an existing set of patches, but you can also _generate the patches_ for a number of repositories.

### Manual campaigns

Manual campaigns provide the ability to manage and monitor changesets (pull requests) that already exist on code hosts by manually adding them to a campaign.

## Creating a manual campaign

1. Go to `/campaigns` on your Sourcegraph instance and click on the "New campaign" button
2. Fill in a name for the campaign and a description
3. Create the campaign
4. Track changesets by adding them to the campaign through the form on the Campaign page

## Creating a campaign using the src CLI

If you have not already, first [install](https://github.com/sourcegraph/src-cli), [set up and configure](https://github.com/sourcegraph/src-cli#setup) the `src` CLI to point to your Sourcegraph instance.

To create a campaign via the CLI:

1. Create an action JSON file (e.g. `action.json`) that contains an action definition
1. _Optional_: See repositories the action would run over: `src actions scope-query -f action.json`
1. Create a set of patches by executing the action over repositories: `src actions exec -f action.json > patches.json`
1. Save the patches in Sourcegraph by creating a patch set: `src campaign patchset create-from-patches < patches.json`
1. Create a campaign based on the patch set: `src campaigns create -branch=<branch-name> -patchset=<patchset-ID-returned-by-previous-command>`

Read on detailed steps and documentation.

## Where to best run campaigns

The patches for a campaign are generated on the machine where the `src` CLI is executed, which in turn, downloads zip archives and runs each step against each repository. For most usecases we recommend that `src` CLI should be run on a Linux machine with considerable CPU, RAM, and network bandwidth to reduce the execution time. Putting this machine in the same network as your Sourcegraph instance will also improve performance.

Another factor affecting execution time is the number of jobs executed in parallel, which is by default the number of cores on the machine. This can be adjusted using the `-j` parameter.

To make it simpler for customers, we're [working on remote execution of campaign](https://github.com/sourcegraph/src-cli/pull/128) of campaign actions and would love your feedback.

## Defining an action

The first thing we need is a definition of an "action". An action contains a list of steps to run in each repository returned by the results of the `scopeQuery` search string.

There are two types of steps: `docker` and `command`. See `src actions exec -help` for more information.

Here is an example of a multi-step action definition using the `docker` and `command` types:

```json
{
  "scopeQuery": "repo:go-* -repohasfile:INSTALL.md",
  "steps": [
    {
      "type": "command",
      "args": ["sh", "-c", "echo '# Installation' > INSTALL.md"]
    },
    {
      "type": "command",
      "args": ["sed", "-i", "", "s/No install instructions/See INSTALL.md/", "README.md"]
    },
    {
      "type": "docker",
      "dockerfile": "FROM alpine:3 \n CMD find /work -iname '*.md' -type f | xargs -n 1 sed -i s/this/that/g"
    },
    {
      "type": "docker",
      "image": "golang:1.13-alpine",
      "args": ["go", "fix", "/work/..."]
    }
  ]
}
```

This action will execute on every repository that has `go-` in its name and doesn't have an `INSTALL.md` file.

1. The first step (a `command` step) creates an `INSTALL.md` file in the root directory of each repository by running `sh` in a temporary copy of each repository. **This is executed on the machine on which `src` is being run.** Note that the first element in `"args"` is the command itself.

2. The second step, again a `"command"` step, runs the `sed` command to replace text in the `README.md` file in the root of each repository (the `-i ''` argument is only necessary for BSD versions of `sed` that usually come with macOS). Please note that the executed command is simply `sed` which means its arguments are _not_ expanded, as they would be in a shell. To achieve that, execute the `sed` as part of a shell invocation (using `sh -c` and passing in a single argument, for example, like in the first step).

3. The third step builds a Docker image from the specified `"dockerfile"` and starts a container with this image in which the repository is mounted under `/work`.

4. The fourth step starts a Docker container based on the `golang:1.13-alpine` image and runs `go fix /work/...` in it.

As you can see from these examples, the "output" of an action is the modified, local copy of a repository.

Save that definition in a file called `action.json` (or any other name of your choosing).

## Executing an action to produce patches

With our action file defined, we can now execute it:

```sh
src actions exec -f action.json
```

This command is going to:

1. Build the required Docker image if necessary.
1. Download a ZIP archive of the repositories matched by the `"scopeQuery"` from the Sourcegraph instance to a local temporary directory in `/tmp`.
1. Execute the action for each repository in parallel (the number of parallel jobs can be configured with `-j`, the default is number of cores on the machine), with each step in an action being executed sequentially on the same copy of a repository. If a step in an action is of type `"command"` the command will be executed in the temporary directory that contains the copy of the repository. If the type is `"docker"` then a container will be launched in which the repository is mounted under `/work`.
1. Produce a patch for each repository with a diff between a fresh copy of the repository's contents and directory in which the action ran.

The output can either be saved into a file by redirecting it:

```sh
src actions exec -f action.json > patches.json
```

Or it can be piped straight into the next command we're going to use to save the patches on the Sourcegraph instance:

```sh
src actions exec -f action.json | src campaign patchset create-from-patches
```

## Creating a patch set from patches

The next step is to save the set of patches on the Sourcegraph instance so they can be turned into a campaign.

To do that, run:

```sh
src campaign patchset create-from-patches < patches.json
```

Or, again, pipe the patches directly into it:

```sh
src actions exec -f action.json | src campaign patchset create-from-patches
```

Once completed, the output will contain:

- The URL to preview the changesets that would be created on the code hosts.
- The command for the `src` SLI to create a campaign from the patch set.

## Publishing a campaign

If you're happy with the preview of the campaign, it's time to trigger the creation of changesets (pull requests) on the code host(s) by creating and publishing the campaign:

```sh
src campaigns create -name='My campaign name' \
   -desc='My first CLI-created campaign' \
   -patchset=Q2FtcGFpZ25QbGFuOjg= \
   -branch=my-first-campaign
```

Creating this campaign will asynchronously create a pull request for each repository that has a patch in the patch set. You can check the progress of campaign completion by viewing the campaign on your Sourcegraph instance.

The `-branch` flag specifies the branch name that will be used for each pull request. If a branch with that name already exists for a repository, a fallback will be generated by appending a counter at the end of the name, e.g.: `my-first-campaign-1`.

If you have defined the `$EDITOR` environment variable, the configured editor will be used to edit the name and Markdown description of the campaign:

```sh
src campaigns create -patchset=Q2FtcGFpZ25QbGFuOjg= -branch=my-first-campaign
```

## Campaign drafts

A campaign can be created as a draft, either by adding the `-draft` flag to the `src campaign create` command, or by selecting `Create draft` in the web UI. When a campaign is a draft, no changesets will be created until the campaign is published, or each changeset is individually published. This can be done in the Sourcegraph campaign web interface.

## Updating a campaign

You can also apply a new patch set to an existing campaign. Following the creation of the patch set with the `src campaign patchset create-from-patches` command, a URL will be output that will guide you to the web UI to allow you to change an existing campaign's patch set.

On this page, click "Preview" for the campaign that will be updated. From there, the delta of existing and new changesets will be displayed. Click "Update" to finalize the proposed changes.

Edits to the name and description of a campaign can also be made in the web UI with the changes reflected in each changeset. The branch name of a draft campaign with a patch set can also be edited, but only if the campaign doesn't contain any published changesets.

## Clearing the campaign action cache

Patches are intelligently cached based on the `scopeQuery` and defined `steps`, but the need to clear the cache to run the steps from scratch may be required.

Clearing the cache requires either manually emptying the cache directory or using a different one. If no `-cache` flag is passed to `src actions exec`, the default location of the cache is used which can be found for your platform by running:

```sh
src actions exec -help
```

## Example campaigns

The following examples demonstrate various types of campaigns for different languages using both commands and Docker images. They also provide commentary on considerations such as adjusting the duration (`-timeout`) for actions that exceed the 15 minute default limit.

### Adding a GitHub action to upload LSIF data to Sourcegraph

Our goal for this campaign is to add a GitHub Action that generates and uploads LSIF data to Sourcegraph by adding a `.github/workflows/lsif.yml` file to each repository that doesn't have it yet.

The first thing we need is the definition of an action that we can execute with the [`src` CLI tool](https://github.com/sourcegraph/src-cli) and its `src actions exec` subcommand.

Here is an `action.json` file that runs a Docker container based on the Docker image called `add-lsif-to-build-pipeline-action` in each repository that has a `go.mod` file, `github` in its name and no `.github/workflows/lsif.yml` file:

```json
{
  "scopeQuery": "repohasfile:go.mod repo:github -repohasfile:.github/workflows/lsif.yml",
  "steps": [
    {
      "type": "docker",
      "image": "add-lsif-to-build-pipeline-action"
    }
  ]
}
```

Save that as `action.json`.

In order to build the Docker image, we first need to create a file called `github-action-workflow-golang.yml` with the following content:

```yaml
name: LSIF
on:
  - push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v1
      - name: Generate LSIF data
        uses: sourcegraph/lsif-go-action@master
        with:
          verbose: "true"
      - name: Upload LSIF data
        uses: sourcegraph/lsif-upload-action@master
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
```

This is the definition of the GitHub action.

Next we create the `Dockerfile`:

```Dockerfile
FROM alpine:3
ADD ./github-action-workflow-golang.yml /tmp/workflows/

CMD mkdir -p .github/workflows && \
  DEST=.github/workflows/lsif.yml; \
  if [ ! -f .github/workflows/lsif.yml ]; then \
    cp /tmp/workflows/github-action-workflow-golang.yml $DEST; \
  else \
    echo Doing nothing because existing LSIF workflow found at $DEST; \
  fi
```

Now we're ready to run the campaign:

1. Build the Docker image: `docker build -t add-lsif-to-build-pipeline-action .`
1. Run the action and create a patch set: `src actions exec -f action.json | src campaign patchset create-from-patches`
1. Follow the printed instructions to create and run the campaign on Sourcegraph

### Refactor Go code using Comby

Our goal for this campaign is to simplify Go code by using [Comby](https://comby.dev/) to rewrite calls to `fmt.Sprintf("%d", arg)` with `strconv.Itoa(:[v])`. The semantics are the same, but one more cleanly expresses the intention behind the code.

> **Note**: Learn more about Comby and what it's capable of at [comby.dev](https://comby.dev/)

To do that we use two Docker containers. One container launches Comby to rewrite the the code in Go files and the other runs [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) to update the `import` statements in the updated Go code so that `strconv` is correctly imported and, possibly, `fmt` is removed.

Here is the `action.json` file that defines this as an action:

```json
{
  "scopeQuery": "lang:go fmt.Sprintf",
  "steps": [
    {
      "type": "docker",
      "image": "comby/comby",
      "args": ["-in-place", "fmt.Sprintf(\"%d\", :[v])", "strconv.Itoa(:[v])", "-matcher", ".go", "-d", "/work"]
    },
    {
      "type": "docker",
      "image": "cytopia/goimports",
      "args": ["-w", "/work"]
    }
  ]
}
```

Please note that the `"scopeQuery"` makes sure that the repositories over which we run the action all contain Go code in which we have a call to `fmt.Sprintf`. That narrows the list of repositories down considerably, even though we still need to search through the whole repository with Comby. (We're aware that this is a limitation and are working on improving the workflows involving exact search results.)

Save the definition in a file, for example `go-comby.action.json`.

Now we can execute the action and turn it into a campaign:

1. Make sure that the `"scopeQuery"` returns the repositories we want to run over: `src actions scope-query -f go-comby.action.json`
1. Execute the action and create a patchset: `src actions exec -f action.json | src campaign patchset create-from-patches`
1. Follow the printed instructions to create and run the campaign on Sourcegraph

### Using ESLint to automatically migrate to a new TypeScript version

Our goal for this campaign is to convert all TypeScript code synced to our Sourcegraph instance to make use of new TypeScript features. To do this we convert the code, then update the TypeScript version.

To convert the code we install and run ESLint with the desired `typescript-eslint` rules, using the [`--fix` flag](https://eslint.org/docs/user-guide/command-line-interface#fix) to automatically fix problems. We then update the TypeScript version using [`yarn upgrade`](https://legacy.yarnpkg.com/en/docs/cli/upgrade/).

The first thing we need is a Docker container in which we can freely install and run ESLint. Here is the `Dockerfile`:

```Dockerfile
FROM node:12-alpine3.10
CMD package_json_bkup=$(mktemp) && \
  cp package.json $package_json_bkup && \
  yarn -s --non-interactive --pure-lockfile --ignore-optional --ignore-scripts --ignore-engines --ignore-platform --no-progress && \
  yarn add --ignore-workspace-root-test --non-interactive --ignore-optional --ignore-scripts --ignore-engines --ignore-platform --pure-lockfile -D @typescript-eslint/parser @typescript-eslint/eslint-plugin --no-progress eslint && \
  node_modules/.bin/eslint \
  --fix \
  --plugin @typescript-eslint \
  --parser @typescript-eslint/parser \
  --parser-options '{"ecmaVersion": 8, "sourceType": "module", "project": "tsconfig.json"}' \
  --rule '@typescript-eslint/prefer-optional-chain: 2' \
  --rule '@typescript-eslint/no-unnecessary-type-assertion: 2' \
  --rule '@typescript-eslint/no-unnecessary-type-arguments: 2' \
  --rule '@typescript-eslint/no-unnecessary-condition: 2' \
  --rule '@typescript-eslint/no-unnecessary-type-arguments: 2' \
  --rule '@typescript-eslint/prefer-includes: 2' \
  --rule '@typescript-eslint/prefer-readonly: 2' \
  --rule '@typescript-eslint/prefer-string-starts-ends-with: 2' \
  --rule '@typescript-eslint/prefer-nullish-coalescing: 2' \
  --rule '@typescript-eslint/no-non-null-assertion: 2' \
  '**/*.ts'; \
  mv $package_json_bkup package.json; \
  rm -rf node_modules; \
  yarn upgrade --latest --ignore-workspace-root-test --non-interactive --ignore-optional --ignore-scripts --ignore-engines --ignore-platform --no-progress typescript && \
  rm -rf node_modules
```

When turned into an image and run as a container, the instructions in this Dockerfile will do the following:

1. Copy the current `package.json` to a backup location so that we can install ESLint without changes to the original `package.json`
1. Install all dependencies & add `eslint` with the `typescript-eslint` plugin
1. Run `eslint --fix` with a set of TypeScript rules to detect and fix problems over all `*.ts` files
1. Restore the original `package.json` from its backup location
1. Run `yarn upgrade` to update the `typescript` version

Before we can run it as an action we need to turn it into a Docker image, by running the following command in the directory where the `Dockerfile` was saved:

```sh
docker build -t eslint-fix-action .
```

That builds a Docker image and names it `eslint-fix-action`.

Once that is done we're ready to define our action:

```json
{
  "scopeQuery": "repohasfile:yarn\\.lock repohasfile:tsconfig\\.json",
  "steps": [
    {
      "type": "docker",
      "image": "eslint-fix-action"
    }
  ]
}
```

The `"scopeQuery"` ensures that the action will only be run over repositories containing both a `yarn.lock` and a `tsconfig.json` file. This narrows the scope down to only the TypeScript projects in which we can use `yarn` to install dependencies. Feel free to narrow it down further by using more filters, such as `repo:my-project-name` to only run over repositories that have `my-project-name` in their name.

The action only has a single step to execute in each repository: it runs the Docker container we just built (called `eslint-fix-action`) on the machine on which `src` is executed.

Save that definition in a file called `eslint-fix-typescript.action.json` and we're ready to execute it.

First we make sure that we match all the repositories we want:

```sh
src actions scope-query -f eslint-fix-typescript.action.json
```

If that list looks good, we're ready to execute the action:

```sh
src actions exec -timeout 15m -f eslint-fix-typescript.action.json | src campaign patchset create-from-patches
```

> **Note**: we're giving the action a generous timeout of 15 minutes per repository, since it needs to download and install all dependencies. With a still-empty caching directory that might take a few minutes.

You should now see that the Docker container we built is being executed in a local, temporary copy of each repository. After executing, the patches it produced will be turned into a patch set you can preview on our Sourcegraph instance.

#### Caching dependencies across multiple steps using `cacheDirs`

If you find yourself writing an action definition that relies on a project's dependencies to be installed for every step it can be helpful to cache these dependencies in a directory outside of the repository.

For `"docker"` steps you can use the `cacheDirs` attribute:

```json
{
  "scopeQuery": "repohasfile:package.json",
  "steps": [
    {
      "type": "docker",
      "image": "yarn-install-dependencies",
      "cacheDirs": [ "/cache" ]
    },
    {
      "type": "docker",
      "image": "eslint-fix-action",
      "cacheDirs": [ "/cache" ]
    },
    {
      "type": "docker",
      "image": "upgrade-typescript",
      "cacheDirs": [ "/cache" ]
    }
  ]
}
```

This defines three separate steps that all use the same `"cacheDirs"`. The specified directory `"/cache"` will be created on the machine on which `src` is executing in a temporary location and mounted in each of the three containers under `/cache`.

As an example, here is the first part of `Dockerfile` that makes use of this `/cache` directory when installing dependencies with `yarn`:

```
FROM node:12-alpine3.10
VOLUME /cache
ENV YARN_CACHE_FOLDER=/cache/yarn
ENV NPM_CONFIG_CACHE=/cache/npm
CMD yarn -s --non-interactive --pure-lockfile --ignore-optional --ignore-scripts --ignore-engines --ignore-platform --no-progress

# [...]
```

This uses `/cache` as the `YARN_CACHE_FOLDER` and `NPM_CONFIG_CACHE` folder. It installs the dependencies with `yarn`, thus populating the `/cache` folder.

Subsequent action steps that use this preamble in their `Dockerfile` will run faster because they can leverage the cache folder.
