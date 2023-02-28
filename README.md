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

## Development

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
   pnpm install
   (cd embeddings && pip3 install -r requirements.txt)
   ```

### Build and run

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

Run the VS Code extension:

1. Open this repository's root directory in VS Code.
1. In VS Code, run the `Debug: Start Debugging` command and select the `Run VS Code Extension` target.
1. Change your VS Code user settings to use your local dev servers:

   ```json
   "cody.serverEndpoint": "http://localhost:9300",
   "cody.embeddingsEndpoint": "http://localhost:9301",
   "cody.debug": true,
   ```

   - Note: You may find it more convenient to use a separate user profile for VS Code (or the Insiders build) so that you can continue using the released version of Cody in your usual editing workflow.

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
