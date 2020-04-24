# Creating a campaign from patches

A campaign can be created from a set of patches, one per repository. For each patch, a changeset (what code hosts call _pull request_ or _merge request_) will be created on the code host on which the repository is hosted.

Here is the short version for how to create a patch set and turn that into changesets by creating a campaign:

1. Create an action definition and save to a JSON file (e.g. `action.json`).
1. _Optional_: See repositories the action would run over:

  ```
  src actions scope-query -f action.json
  ```
1. Create a set of patches by executing the action:

  ```
  src actions exec -f action.json -create-patchset
  ```
1. Click on the URL that's printed to create a campaign.

Read on for more detailed steps and documentation and see "[Actions](./actions.md)" for more information about how to define and execute actions.

## Requirements

If you have not done so already, first [install](https://github.com/sourcegraph/src-cli), [set up and configure](https://github.com/sourcegraph/src-cli#setup) the `src` CLI to point to your Sourcegraph instance.

## 1. Defining an action

The first thing we need is a definition of an "action". An action contains a list of steps to run in each repository returned by the results of the `scopeQuery` search string. There are two types of steps: `docker` and `command`. See "[Actions](./actions.md)" for more information.

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
      "image": "golang:1.13-alpine",
      "args": ["go", "fix", "/work/..."]
    }
  ]
}
```

**Save that definition in a file called `action.json` (or any other name of your choosing).**

This action will be executed for each repository that has `go-` in its name and doesn't have an `INSTALL.md` file.

- **The first step** (a `command` step) creates an `INSTALL.md` file in the root directory of each repository by running `sh` in a temporary copy of each repository. This is executed on the machine on which `src` is being run.
- **The second step**, again a `"command"` step, runs the `sed` command to replace text in the `README.md` file in the root of each repository (the `-i ''` argument is only necessary for BSD versions of `sed` that usually come with macOS).

    > NOTE: The executed command is simply `sed` which means its arguments are _not_ expanded, as they would be in a shell. To achieve that, execute the `sed` as part of a shell invocation (using `sh -c` and passing in a single argument, for example, like in the first step).

- **The third step** starts a Docker container based on the `golang:1.13-alpine` image and runs `go fix /work/...` in it.

As you can see from these examples, the "output" of an action is the modified, local copy of a repository.


## 2. Executing an action to produce patches

With our action file defined, we can now execute it:

```
src actions exec -f action.json
```

This command is going to:

1. Download a ZIP archive of each repository returned by the `"scopeQuery"` and extract it to a local temporary directory in `/tmp`.
1. Execute the action for each repository in parallel, with each step in an action being executed sequentially on the same copy of a repository.
1. Produce a patch for each repository with creating diff between a fresh copy of the repository and the directory in which the action ran.

(See "[Actions](./actions.md)" for more information about how actions are executed.)

The output, a set of patches, can either be saved into a file by redirecting it:

```
src actions exec -f action.json > patches.json
```

## 3. Creating a patch set from patches

The next step is to save the set of patches on the Sourcegraph instance so they can be turned into a campaign.

To do that, run:

```
src campaign patchset create-from-patches < patches.json
```

Or leverage the cache of `src action exec` and re-run the command to pipe the patches directly into `src campaign patchset create-from-patches`:

```
src actions exec -f action.json | src campaign patchset create-from-patches
```

You can also use `src action exec` with the `-create-patchset` flag, which is equivalent to the last command:

```
src actions exec -f action.json -create-patchset
```

Once completed, the output will contain:

- The URL to preview the changesets that would be created on the code hosts.
- The command for the `src` SLI to create a campaign from the patch set.

## 4. Publishing a campaign

If you're happy with the preview of the campaign, it's time to trigger the creation of changesets (pull requests) on the code host(s).

That is done by creating and publishing a campaign with the given patchset.

You can either do that in the Sourcegraph UI or on the CLI:

```
src campaigns create -name='My campaign name' \
  -desc='My first CLI-created campaign' \
  -patchset=Q2FtcGFpZ25QbGFuOjg= \
  -branch=my-first-campaign
```

Creating this campaign will asynchronously create a changeset (pull request) for each repository that has a patch in the patch set. You can check the progress of campaign completion by viewing the campaign on your Sourcegraph instance.

The `-branch` flag specifies the branch name that will be used for each pull request. If a branch with that name already exists for a repository, a fallback will be generated by appending a counter at the end of the name, e.g.: `my-first-campaign-1`.

If you have defined the `$EDITOR` environment variable, the configured editor will be used to edit the name and Markdown description of the campaign:

```sh
src campaigns create -patchset=Q2FtcGFpZ25QbGFuOjg= -branch=my-first-campaign
```
