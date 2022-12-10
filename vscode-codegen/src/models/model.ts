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
export interface CompletionSupplier {
  getName(): string;
  getCompletions(
    document: TextDocument,
    position: Position,
    maxTokens: number,
    tries: number
  ): Promise<InlineCompletionItem[]>;
}
