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

## Usage

There are two types of Automation campaigns:

- Manual campaigns to which you can manually add changesets (pull requests) and track their progress
- Campaigns created from a set of patches

### Creating a manual campaign

1. Go to `/campaigns` on your Sourcegraph instance and click on the "New campaign" button
2. Fill in a name for the campaign and a description
3. Create the campaign
4. Track changesets by adding them to the campaign through the form on the Campaign page

### Creating a campaign from a set of patches

**Required**: The [`src` CLI tool](https://github.com/sourcegraph/src-cli). 

Short overview:

1. Create an `action.json` file that contains an action definition.
2. Create a set of patches by executing the action over repositories: `src actions exec -f action.json > patches.json`
3. Save the patches in Sourcegraph by creating a campaign plan based on these patches: `src campaign plan create-from-patches < patches.json`
4. Create a campaign from the campaign plan: `src campaigns create -name='Campaign name' -desc='Description' -plan=<plan-ID-returned-by-previous-command>`

Read on for the longer version.

**Note about scalability**: the patches are generated on the machine on which the `src` CLI tool is being run. You can tune the number of parallel jobs with the `-j` parameter. If you still run into performance issues, feel free to use a bigger machine, possibly closer to your Sourcegraph instance so that downloading archives of repositories is faster.

#### Defining an action

The first thing we need is a definition of an "action". An action is what produces a patch and describes what commands and Docker containers to run over which repositories.

Example:

```json
{
  "scopeQuery": "repo:go-* -repohasfile:INSTALL.md",
  "steps": [
    {
      "type": "command",
      "args": ["sh", "-c", "echo '# Installation' > INSTALL.md"]
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

The first step creates an `INSTALL.md` file by running `sh` in each repository on the machine on which `src` is executed.

The second step builds a Docker image from the specified `"dockerfile"` and starts a container with this image in which the repository is mounted under `/work`.

The third step pulls the `golang:1.13-alpine` image from Docker hub, starts a container from it and runs `go fix /work/...` in it.

Save that definition in a file called `action.json` (or any other name of your choosing).

#### Executing an action to produce patches

With our action defined we can now execute it:

```
$ src actions exec -f action.json
```

This command is going to:

1. Download or build the required Docker images.
2. Download a copy of the repositories matched by the `"scopeQuery"` from the Sourcegraph instance.
3. Execute the action in each repository in parallel (the maximum number of parallel jobs can be configured with `-j`, the default is number of cores on the machine)
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

This will create a campaign on the Sourcegraph instance and asychronously create a pull request for each patch on the code hosts on which the repositories are hosted.

Check progress by opening the campaign on your Sourcegraph instance.

## Example: Campaign to add a GitHub action to upload LSIF data to Sourcegraph

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
1. Run the action and create a campaign plan: `src actions exec -f action.json`
1. Follow the printed instructions to create and run the campaign on Sourcegraph

## Note for Automation developers

If you are looking to run automation on a larger scale in the local dev environment, follow the [guide on automation development](../dev/automation_development.md).
