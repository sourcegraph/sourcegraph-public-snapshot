# Automation

> Automation is currently available in private beta for select enterprise customers.

[Sourcegraph automation](https://about.sourcegraph.com/product/automation) allows large-scale code changes across many repositories and different code hosts.

**Important**: If you're on Sourcegraph 3.12 or older, you might also want to look at the old documentation: "[Automation documentation for Sourcegraph 3.12](https://docs.sourcegraph.com/@3.12/user/automation)"

## Configuration

In order to use the Automation preview, a site-admin of your Sourcegraph instance must enable it in the site configuration settings e.g. `sourcegraph.example.com/site-admin/configuration`

```json
{
  "experimentalFeatures": {
      "automation": "enabled"
  }
}
```

Without any further configuration Automation is **only be accessible to site-admins.** If you want to grant read-only access to non-site-admins, use the following site configuration setting:

```json
{
  "automation.readAccess.enabled": true
}
```

## Usage

There are two types of Automation campaigns:

- Manual campaigns to which you can manually add changesets (pull requests) and track their progress.
- Campaigns created from a set of patches. With the `src` CLI tool you can not only create the campaign from an existing set of patches, but you can also _generate the patches_ for a number of repositories.

### Creating a manual campaign

1. Go to `/campaigns` on your Sourcegraph instance and click on the "New campaign" button
2. Fill in a name for the campaign and a description
3. Create the campaign
4. Track changesets by adding them to the campaign through the form on the Campaign page

### Creating a campaign from a set of patches

**Required**: The [`src` CLI tool](https://github.com/sourcegraph/src-cli). Make sure it is setup to point at your Sourcegraph instance by setting the `SRC_ENDPOINT` environment variable.

Short overview:

1. Create an `action.json` file that contains an action definition.
1. _Optional_: See repositories the action would run over: `src actions scope-query -f action.json`
1. Create a set of patches by executing the action over repositories: `src actions exec -f action.json > patches.json`
1. Save the patches in Sourcegraph by creating a campaign plan based on these patches: `src campaign plan create-from-patches < patches.json`
1. Create a campaign from the campaign plan: `src campaigns create -plan=<plan-ID-returned-by-previous-command>`

Read on for the longer version.

> **Note about scalability**: the patches are generated on the machine on which the `src` CLI tool is being run. That means archives of the repositories in a campaign have to be downloaded from your Sourcegraph instance to the local machine. We're working on remote execution of campaign actions. For now feel free to use a bigger machine, possibly closer to your Sourcegraph instance so that downloading archives of repositories is faster, and use the `-j` parameter to tune the number of parallel jobs being executed.

#### Defining an action

The first thing we need is a definition of an "action". An action is what produces a patch and describes what commands or Docker containers to run over which repositories.

> **Note**: At the moment only two `"type"`s of steps are supported: `"docker"` and `"command"`. See `src actions exec -help` for more information.

Here is an example defintion of an action:

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

This action runs over every repository that has `go-` in its name and doesn't have an `INSTALL.md` file.

The first step, a `"command"` step, creates an `INSTALL.md` file in the root directory of each repository by running `sh` in a temporary copy of each repository. **This is executed on the machine on which `src` is being run.** Note that the first element in `"args"` is the command itself.

The second step, again a `"command"` step, runs the `sed` command to replace text in the `README.md` file in the root of each repository (the `-i ''` argument is only necessary for BSD versions of `sed` that usually come with macOS). Please note that the executed command is simply `sed` which means its arguments are _not_ expanded, as they would be in a shell. To achieve that, execute the `sed` as part of a shell invocation (using `sh -c` and passing in a single argument, for example, like in the first step).

The third step builds a Docker image from the specified `"dockerfile"` and starts a container with this image in which the repository is mounted under `/work`.

The fourth step pulls the `golang:1.13-alpine` image from Docker hub, starts a container from it and runs `go fix /work/...` in it.

As you can see from these examples, the "output" of an action is the modified, local copy of a repository.

Save that definition in a file called `action.json` (or any other name of your choosing).

#### Executing an action to produce patches

With our action defined we can now execute it:

```
$ src actions exec -f action.json
```

This command is going to:

1. Download or build the required Docker images, if necessary.
2. Download a ZIP archive of the repositories matched by the `"scopeQuery"` from the Sourcegraph instance to a local temporary directory in `/tmp`.
3. Execute the action in each repository in parallel (the maximum number of parallel jobs can be configured with `-j`, the default is number of cores on the machine), with each step in an action being executed sequentially on the same copy of a repository. If a step in an action is of type `"command"` the command will be executed in the temporary directory that contains the copy of the repository. If the type is `"docker"` then a container will be launched in which the repository is mounted under `/work`.
4. Produce a diff for each repository between a fresh copy of the repository's contents and directory in which the action ran.

The output can either be saved into a file by redirecting it:

```
$ src actions exec -f action.json > patches.json
```

Or it can be piped straight into the next command we're going to use to save the patches on the Sourcegraph instance:

```
$ src actions exec -f action.json | src campaign plan create-from-patches
```

#### Creating a campaign plan from patches

The next step is to save the set of patches on the Sourcegraph instance so they can be run together as a campaign.

To do that we use the following command:

```
$ src campaign plan create-from-patches < patches.json
```

Or, again, pipe the patches directly into it.

When the command successfully ran, it will print a URL with which you can preview the changesets that would be created on the codehosts, or a command for the `src` tool to create a campaign from the campaign plan.

#### Creating a campaign

If you're happy with the campaign plan and its patches, it's time to create changesets (pull requests) on the code hosts by creating a campaign:

```
$ src campaigns create -name='My campaign name' \
   -desc='My first CLI-created campaign'
   -plan=Q2FtcGFpZ25QbGFuOjg=
```

If you have `$EDITOR` configured you can use the configured editor to edit the name and Markdown description of the campaign:

```
$ src campaigns create -plan=Q2FtcGFpZ25QbGFuOjg=
```

The `src campaigns create` command will create a campaign on the Sourcegraph instance and asychronously create a pull request for each patch on the code hosts on which the repositories are hosted.

Check progress by opening the campaign on your Sourcegraph instance.

## Example: Add a GitHub action to upload LSIF data to Sourcegraph

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

```
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
1. Run the action and create a campaign plan: `src actions exec -f action.json | src campaign plan create-from-patches`
1. Follow the printed instructions to create and run the campaign on Sourcegraph

## Example: Refactor Go code with Comby

Our goal for this campaign is to simplify Go code by using [Comby](https://comby.dev/) to rewrite calls to `fmt.Sprintf("%d", arg)` with `strconv.Itoa(:[v])`. The semantics are the same, but one more cleanly expresses the intention behind the code.

> **Note**: Learn more about Comby and what it's capable of at [comby.dev](https://comby.dev/)

To do that we use two Docker containers. One container launches Comby to rewrite the the code in Go files and the other runs [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) to update the `import` statements in the updated Go code so that `strconv` is correctly important and, possibly, `fmt` is removed.

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

Save the defintion in a file, for example `go-comby.action.json`.

Now we can execute the action and turn it into a campaign:

1. Make sure that the `"scopeQuery"` returns the repositories we want to run over: `src actions scope-query -f go-comby.action.json`
1. Execute the action and create a campaign plan: `src actions exec -f action.json | src campaign plan create-from-patches`
1. Follow the printed instructions to create and run the campaign on Sourcegraph

## Example: Use ESLint to automatically migrate to a new TypeScript version

Our goal for this campaign is to convert all TypeScript code synced to our Sourcegraph instance to make use of new TypeScript features. To do this we convert the code, then update the TypeScript version.

To convert the code we install and run ESLint with the desired `typescript-eslint` rules, using the [`--fix` flag](https://eslint.org/docs/user-guide/command-line-interface#fix) to automatically fix problems. We then update the TypeScript version using [`yarn upgrade`](https://legacy.yarnpkg.com/en/docs/cli/upgrade/).

The first thing we need is a Docker container in which we can freely install and run ESLint. Here is the `Dockerfile`:

```
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
src actions scope-query -timeout 15m -f eslint-fix-typescript.action.json | src campaign plan create-from-patches
```

> **Note**: we're giving the action a generous timeout of 15 minutes per repository, since it needs to download and install all dependencies. With a still-empty caching directory that might take a few minutes.

You should now see that the Docker container we built is being executed in a local, temporary copy of each repository. After executing, the diff it generated will be turned into a campaign plan you can preview on our Sourcegraph instance.

## Caching dependencies across multiple steps using `cacheDirs`

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

## Note for Automation developers

If you are looking to run automation on a larger scale in the local dev environment, follow the [guide on automation development](../dev/automation_development.md).
