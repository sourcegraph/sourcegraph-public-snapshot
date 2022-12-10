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
import { CompletionSupplier } from "./model";

export class CodeGenCompletionSupplier implements CompletionSupplier {
  log: (...args: any[]) => void;
  serverUrl: string;
  model: string;
  constructor(serverUrl: string, model: string, log: (...args: any[]) => void) {
    this.serverUrl = serverUrl;
    this.model = model;
    this.log = log;
  }

  getName(): string {
    return "CodeGen";
  }

  async getCompletions(
    document: TextDocument,
    position: Position,
    maxTokens: number,
    tries: number
  ): Promise<InlineCompletionItem[]> {
    this.log("getCompletions:", document.uri, position, tries);
    const prompt = document.getText(
      new vscode.Range(0, 0, position.line, position.character)
    );
    const oa = new openai.OpenAIApi(
      new openai.Configuration({ basePath: this.serverUrl })
    );
    const args = {
      model: this.model,
      prompt: prompt,
      max_tokens: maxTokens,
      stop: getStopSequences(position.line, document),
      n: tries,
    };
    const response = await oa.createCompletion(args);
    const completions =
      response.data.choices
        ?.map((choice) => choice.text)
        .map(
          (choiceText) =>
            new vscode.InlineCompletionItem(
              choiceText as string,
              new vscode.Range(position, position)
            )
        ) || [];

    this.log(`got ${completions.length} completions`);
    return completions;
  }
}

export function getStopSequences(position: number, doc: vscode.TextDocument) {
  const stopSequences: string[] = [];

  // If the position is the last line, we don't use stop sequences
  if (doc.lineCount === position + 1) {
    return stopSequences;
  }
  // If the nextLine is not empty, it will be used as a stop sequence
  const nextLine = doc.lineAt(position + 1).text;
  if (nextLine) {
    stopSequences.push(nextLine.toString());
  }
  return stopSequences;
}
