# Getting started with server-side Batch Changes

<aside class="experimental">This feature is experimental. Follow the <a href="server_side#setup">setup guide</a> to get started.</aside>

## Creating your first server-side Batch Change

To get started, click on the "Create batch change" button on the Batch Changes page, or go to `/batch-changes/create`.
You will be prompted to choose a name for your namespace and optionally define a custom namespace to put your batch change in.

<!-- TODO: Screenshot of the batch change create form here. -->

Once done, click "Create".

### Editing the spec file

You should see the editor view now. This view consists of three main areas:

- The library sidebar panel on the left
- The editor in the middle
- The workspaces preview panel on the right

<!-- TODO: Screenshot of the editor view with 3 red boxes around the panels. -->

You can pick from the examples in the library pane to get started quickly, or begin working on your batch spec in the editor right away. The editor will provide documentation as you hover over tokens in the YAML spec and supports auto-completion.

### Previewing workspaces

Once satisfied or to test your batch change's scope, you can at any time run a new preview from the right hand side panel. After resolution, it will show all the workspaces in repositories that matched the given `on` statements. You can search through them and determine if your query is satisfying before starting execution. You can also exclude single workspaces from this list.

<!-- TODO: Screenshot of the editor page with workspaces resolved. -->

### Executing your batch spec

When the spec is ready to run, make sure the preview is up to date and then click "Run batch spec". You will be taken to the execution screen.

On this page, you see

- Run statistics at the top
- All the workspaces including status and diff stat in the left panel
- Details on a particular workspace on the right hand side panel where you can see
  - Steps with
    - Command
    - Logs
    - Output variables
    - Per-step diffs
  - Results
  - Execution timelines for debugging

Once finished, you can proceed to the batch spec preview as you know it from before.

<!-- TODO: Screenshot of the execution page here. -->

### Previewing and applying the batch spec

On this page, you can review the changes proposed one more time and also review the operations taken by Sourcegraph on each changeset. Once satisfied, click "Apply".

Congratulations, you got your first server-side Batch Change going 🎊

<!-- TODO: Screenshot of the preview page here. -->

## Updating your server-side batch change

This section is a work in progress.
