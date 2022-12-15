import * as vscode from "vscode";
import {
  TextDocument,
  Position,
  InlineCompletionContext,
  CancellationToken,
  ProviderResult,
  InlineCompletionItem,
  InlineCompletionList,
} from "vscode";
import * as openai from "openai";
import { cp } from "fs";
import { CompletionSupplier } from "./models/model";
import { OpenAICompletionSupplier } from "./models/codegen";
import {
  CodegenDocumentProvider,
  completionsDisplayString,
} from "./docprovider";

export async function generateTest(documentProvider: CodegenDocumentProvider) {
  const editor = vscode.window.activeTextEditor;
  const ext = editor?.document.fileName.split(".").pop();
  const snippet = editor?.document.getText(editor.selection);
  const prompt = `Here is the code I wish to test:
  \`\`\`
  ${snippet}
  \`\`\`

  Write the code for the unit test:
  \`\`\`
  `;

  const config = new openai.Configuration({
    apiKey: vscode.workspace
      .getConfiguration()
      .get("conf.codebot.openai.apiKey"),
  });

  const generatedUri = vscode.Uri.parse(`codegen:unittests.${ext}`);
  documentProvider.setDocument(generatedUri, null);
  const oa = new openai.OpenAIApi(config);
  const completions = oa
    .createCompletion({
      model: "code-davinci-002",
      prompt: prompt,
      temperature: 0.2,
      max_tokens: 700,
      stop: "```",
      n: 1,
    })
    .then((response) => response.data.choices.map((choice) => choice.text))
    .then((choiceText) => choiceText.filter((c) => c).join("\n\n"));
  documentProvider.setDocument(generatedUri, completions);

  const doc = await vscode.workspace.openTextDocument(generatedUri);
  await vscode.window.showTextDocument(doc, {
    preview: false,
    viewColumn: 2,
  });
}
