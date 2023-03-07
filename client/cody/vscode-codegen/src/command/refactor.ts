/* eslint-disable unicorn/no-empty-file */
// import * as vscode from "vscode";
// import { ClaudeBackend } from "../prompts/claude";

// export async function refactor(claude: ClaudeBackend) {
//   const editor = vscode.window.activeTextEditor;
//   const snippet = editor?.document.getText(editor.selection);
//   if (!snippet) {
//     return;
//   }

//   // show vscode input
//   const description = await vscode.window.showInputBox({
//     prompt: "What would you like to change?",
//   });
//   if (!description) {
//     return;
//   }

//   const result = await claude.refactor(snippet, description);

//   // replace the editor's current selection with result
//   await editor?.edit((editBuilder) => {
//     editBuilder.replace(editor.selection, result);
//   });
// }
