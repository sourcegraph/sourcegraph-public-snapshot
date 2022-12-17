import * as vscode from "vscode";
import { Utils } from "vscode-uri";
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

export class OpenAICompletionSupplier implements CompletionSupplier {
  log: (...args: any[]) => void;
  config: openai.Configuration;
  model: string;
  name: string;
  constructor(
    config: openai.Configuration,
    model: string,
    name: string,
    log: (...args: any[]) => void
  ) {
    this.config = config;
    this.model = model;
    this.name = name;
    this.log = log;
  }

  getName(): string {
    return this.name;
  }

  async getCompletions(
    document: TextDocument,
    position: Position,
    maxTokens: number,
    tries: number
  ): Promise<InlineCompletionItem[]> {
    this.log("getCompletions:", document.uri, position, tries);
    const offset = document.offsetAt(position);
    const maxLookback = 500;
    const startOffset = offset > maxLookback ? offset - maxLookback : 0;
    const startPosition = document.positionAt(startOffset);
    const prompt = document.getText(
      new vscode.Range(
        startPosition.line,
        startPosition.character,
        position.line,
        position.character
      )
    );
    const oa = new openai.OpenAIApi(this.config);

    this.log("prompt:", prompt);
    try {
      const response = await oa.createCompletion({
        model: this.model,
        prompt: prompt,
        temperature: 0.2,
        max_tokens: maxTokens,
        // stop: getStopSequences(position.line, document), // OpenAI doesn't seem to like this sometimes for some reason
        n: tries,
      });
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

      this.log(
        `Model ${this.getName()} returned ${completions.length} completions`
      );
      return completions;
    } catch (e) {
      console.error(
        `Model ${this.getName()} failed to fetch completions: ${e}`
      );
      throw e;
    }
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
