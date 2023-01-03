# codebot

# Usage and features

- Autocomplete: `alt-\` to show autocompletion suggestions
- Chatbot: Click the robot icon in the primary side panel

Coming soon:

- Inline autocompletion
- Unit test generation
- Context-aware chatbot
- Instruction-driven refactoring

## Dev

There are three separate components within one node workspace:
- `vscode-codegen`: the VS Code extension
- `server`: the server that speaks to the VS Code extension and in turn speaks to the LLM API
- `common`: a library shared by the extension and server with common types

To run in development,
- `cd server && CLAUDE_KEY=<claude api key> OPENAI_KEY=<openai api key> npm run dev`
- `cd vscode-codegen && npm run watch`
- `cd common && npm run dev`
- Launch the extension by opening VS Code to the `vscode-codegen`
  directory (`code vscode-codegen`) and selecting the "Run Extension"
  target.
