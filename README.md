# Cody

## Usage and features

- Autocomplete: `alt-\` to show autocompletion suggestions
- Chatbot: Click the robot icon in the primary side panel

## Install

1. Set the following in your VS Code user settings JSON:

   ```json
   "cody.serverEndpoint": "",
   "cody.embeddingsEndpoint": "",
   ```

1. Set the following in your workspace settings JSON:

   ```json
   "cody.codebase": "github.com/example/repo",
   ```

1. [Install the extension](https://code.visualstudio.com/docs/editor/extension-marketplace#_install-from-a-vsix) from the [latest VSIX file](https://github.com/sourcegraph/codebot/releases).

1. Run the "Cody: Set access token" command to set the access token and reload the editor.

## Development

There are four separate components:

- `vscode-codegen`: the VS Code extension
- `server`: the server that serves the completion and chat endpoints
- `embeddings`: generates the embeddings and serves the embeddings endpoint
- `common`: a library shared by the extension and server with common types

Set the following environment variables:

```
export OPENAI_API_KEY=sk-...
export ANTHROPIC_API_KEY=...
export EMBEDDINGS_DIR=/path/to/embedings/dir
export CODY_USERS_PATH=/path/to/users.json
```

To run the server:

- `cd server && npm run dev`

To run the embeddings API:

- `cd embeddings && uvicorn api:app --reload`

To build the common library:

- `cd common && npm run dev`

To build the extension:

- `cd vscode-codegen && npm run watch`
- `cd vscode-codegen && npm run watch:copy-static`
- Launch the extension by opening VS Code to the `vscode-codegen` directory (`code vscode-codegen`) and selecting the "Run Extension" target.
