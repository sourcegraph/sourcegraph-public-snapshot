# Cody

## Usage and features

- Autocomplete: `alt-\` to show autocompletion suggestions
- Chatbot: Click the robot icon in the primary side panel

## Install

See the [#announce-cody Slack channel](https://app.slack.com/client/T02FSM7DL/C04MZPE4JKD) for instructions.

## Development

There are four separate components:

- `vscode-codegen`: the VS Code extension
- `server`: the server that serves the completion and chat endpoints
- `embeddings`: generates the embeddings and serves the embeddings endpoint
- `common`: a library shared by the extension and server with common types

### Setup

1. Install [asdf](https://asdf-vm.com/)
1. Run `asdf install` (if needed, run `asdf plugin add NAME` for any missing plugins)
1. Set the following environment variables:

   ```
   export OPENAI_API_KEY=sk-...
   export ANTHROPIC_API_KEY=...
   export EMBEDDINGS_DIR=/path/to/embeddings/dir
   export CODY_USERS_PATH=/path/to/users.json
   ```

   See [Cody secrets](https://docs.google.com/document/d/1b5oqnE0kSUrgrb4Z2Alnhfods5e4Y5gx_oaIcQH4TZM/edit) (internal Google Doc) for these secret values.

1. Install dependencies:

   ```shell
   brew install wget  # Use your system's package manager
   pnpm install
   (cd embeddings && pip3 install -r requirements.txt)
   ```

### Build and run

Build and watch the TypeScript code (if you're running VS Code, this runs automatically in the background):

1. `pnpm exec tsc --build`

Run the server:

1. `cd server && CODY_PORT=9300 pnpm run dev`

Run the embeddings API:

1. Generate embeddings, including for at least 1 codebase. See [embeddings/README.md](embeddings/README.md).

   For example:

   ```shell
   cd embeddings
   python3 embed_repos.py --repos https://github.com/sourcegraph/conc --output-dir=$EMBEDDINGS_DIR
   python3 embed_context_dataset.py --output-dir=$EMBEDDINGS_DIR
   ```

   If you do this, ensure your `CODY_USERS_PATH` file has `github.com/sourcegraph/conc` in the `accessibleCodebaseIDs`.

1. `cd embeddings && asdf env python uvicorn api:app --reload --port 9301`

### Developing the [VS Code extension](vscode-codegen/)

1. Change your VS Code user settings to use your local dev servers:

   ```json
   "cody.serverEndpoint": "http://localhost:9300",
   "cody.embeddingsEndpoint": "http://localhost:9301",
   "cody.debug": true,
   ```

2. Run `pnpm install` from the root of this repository
3. Select `Launch Cody Extension` from the dropdown menu in the `RUN AND DEBUG` sidebar
   1. Remove `node_modeules` and rerun `pnpm install` if the start up failed
4. Refresh the extension to see updated changes

#### File structure

- `vscode-codegen/src`: source code of the components for the extension
  host
- `vscode-codegen/webviews`: source code of the extension UI (webviews),
  build with Vite and rollup.js using the `vscode-codegen/vite.config.ts` file at directory
  root
- `vscode-codegen/dist`: build outputs from both webpack and vite
- `vscode-codegen/resources`: everything in this directory will be move to
  the ./dist directory automatically during build time for easy packaging
- `vscode-codegen/index.html`: the entry file that Vite looks for to build
  the webviews. The extension host reads this file at run time and replace
  the variables inside the file with webview specific uri and info

### Testing the [VS Code extension](vscode-codegen/)

```
$ cd vscode-codegen
$ pnpm test
```

### Publishing the [VS Code extension](vscode-codegen/)

Increment the `version` in [`vscode-codegen/package.json`](vscode-codegen/package.json) and then run:

```shell
cd vscode-codegen
pnpm run vsce:package

# To try the packaged extension locally, disable any other installations of it and then run:
#   code --install-extension dist/kodj.vsix

# To publish the packaged extension to the VS Code Extension Marketplace:
pnpm exec vsce publish --packagePath dist/kodj.vsix
```

> NOTE: Since the extension has already been bundled, we will need to add the `--no-dependencies` flag during the packaging step to exclude the npm dependencies ([source](https://github.com/microsoft/vscode-vsce/issues/421#issuecomment-1038911725))
