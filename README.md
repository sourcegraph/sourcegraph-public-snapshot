# codebot

## Usage and features

- Autocomplete: `alt-\` to show autocompletion suggestions
- Chatbot: Click the robot icon in the primary side panel

Coming soon:

- Inline autocompletion
- Unit test generation
- Context-aware chatbot
- Instruction-driven refactoring

## Install

1. Set the following in your VS Code user settings JSON:

	```
	"conf.codebot.serverEndpoint": "sourcegraph-650f.ngrok.io",
	```
1. [Install the extension](https://code.visualstudio.com/docs/editor/extension-marketplace#_install-from-a-vsix) from the [latest VSIX file](https://github.com/sourcegraph/codebot/releases).

## Dev

There are four separate components:
- `vscode-codegen`: the VS Code extension
- `server`: the server that serves the completion and chat endpoints
- `embeddings`: generates the embeddings and serves the embeddings endpoint
- `common`: a library shared by the extension and server with common types

To run in development,
- `cd server && CLAUDE_KEY=<claude api key> OPENAI_KEY=<openai api key> npm run dev`
- `cd server && npm run watch:copy-static`
- `cd embeddings && export OPENAI_API_KEY=<open api key> export EMBEDDINGS_DIR=<path to embeddings dir>; export CODEBASE_IDS=<comma delimited codebase ids>; uvicorn api:app --reload`
- `cd vscode-codegen && npm run watch`
- `cd common && npm run dev`
- Launch the extension by opening VS Code to the `vscode-codegen`
  directory (`code vscode-codegen`) and selecting the "Run Extension"
  target.
