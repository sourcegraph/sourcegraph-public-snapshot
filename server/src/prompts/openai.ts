import { Configuration, CreateCompletionRequest, OpenAIApi } from "openai";
import { CompletionsArgs, InflatedHistoryItem, LLMDebugInfo, ReferenceInfo } from "@sourcegraph/cody-common";
import {
  getCondensedText,
  tokenCost,
} from "./common";
import { createCompletion } from "./openai-ratelimit";
import { performance } from "perf_hooks";

export class OpenAIBackend {
  maxTokens = 8000; // token budget, TODO(beyang): adjust by subtracting lookback window, because otherwise we might "forget" first part of the prompt

  defaultCompletionParams: CreateCompletionRequest;
  oa: OpenAIApi;
  createPrompt: (opt: PromptOpt) => string;
  stopStrings: (uri: string) => string[] | undefined;
  extractFromResponse?: (responseText: string) => string;

  readonly expectedResponses: number;

  constructor(
    config: Configuration,
    defaultCompletionParams: CreateCompletionRequest,
    createPrompt: (opt: PromptOpt) => string,
    stopStrings: (uri: string) => string[] | undefined,
    extractFromResponse?: (responseText: string) => string
  ) {
    this.defaultCompletionParams = defaultCompletionParams;
    this.oa = new OpenAIApi(config);
    this.createPrompt = createPrompt;
    this.stopStrings = stopStrings;
    this.extractFromResponse = extractFromResponse;
    this.expectedResponses = defaultCompletionParams.n || 1;
  }

  async getCompletions({
    uri,
    prefix,
    history,
    references,
  }: CompletionsArgs): Promise<{
    debug: LLMDebugInfo;
    completions: string[];
  }> {
		const startTime = performance.now();
		const prompt = this.createPrompt({
			prefix,
			history,
			references,
			maxTokens: this.maxTokens,
			tokenCost,
		});
		const opt = {
			...this.defaultCompletionParams,
			stop: this.stopStrings(uri),
		};

		// using createCompletion instead of this.oa.createCompletion to ensure openai requests happen serial (see function docs)
		const response = await createCompletion(this.oa, {
			...opt,
			prompt,
		});

		const endTime = performance.now();
		return {
			debug: {
				prompt,
				llmOptions: opt,
				elapsedMillis: endTime - startTime,
			},
			completions: (
				(response.data as any).choices
					?.map((choice: any) => choice.text)
					.filter((e: any) => e) as string[]
			).map((e) =>
				this.extractFromResponse ? this.extractFromResponse(e) : e
			),
		};
	}
}

interface PromptOpt {
  prefix: string;
  history?: InflatedHistoryItem[];
  references?: ReferenceInfo[];

  maxTokens: number;
  tokenCost: (s: string, assumeExtraNewlines?: number) => number;
}

// eslint-disable-next-line @typescript-eslint/naming-convention
export const prompt_refs_history_inlinecomments =
  (lookback: number) =>
  ({
    prefix,
    history,
    references,

    maxTokens,
    tokenCost,
  }: PromptOpt): string => {
    let remainingBudget = maxTokens;
    const adjustedPrefix =
      prefix.length > lookback
        ? prefix.substring(prefix.length - lookback)
        : prefix;
    const promptComponents: string[] = [
      `// Code to complete:\n${adjustedPrefix}`,
    ];

    remainingBudget -= tokenCost(promptComponents.join("\n"));
    if (remainingBudget < 0) {
      throw new Error("not enough token budget for prefix");
    }
    if (!history) {
      return promptComponents.join("\n");
    }
    const dedupedReverseHistory: InflatedHistoryItem[] = [];
    const seenURIs: { [uri: string]: boolean } = {};
    for (const item of [...history].reverse()) {
      if (seenURIs[item.item.uri.toString()]) {
        continue;
      }
      seenURIs[item.item.uri.toString()] = true;
      dedupedReverseHistory.push(item);
    }

    const relativeProportions = {
      historySymbols: 0.5 * (history ? 1 : 0),
      references: 0.5 * (references ? 1 : 0),
    };
    const total = Object.values(relativeProportions).reduce((x, y) => x + y, 0);
    const proportions = Object.fromEntries(
      Object.entries(relativeProportions).map(([specifier, proportion]) => [
        specifier,
        proportion / total,
      ])
    );

    // references
    let refBudget = Math.round(remainingBudget * proportions.historySymbols);
    if (references && refBudget > 0) {
      const refComponents: string[] = [];
      for (const ref of [...references].reverse()) {
        const component = `// Example from ${ref.filename}\n${ref.text}`;
        const cost = tokenCost(component);
        if (cost > refBudget) {
          break;
        }
        refComponents.push(component);
        refBudget -= cost;
      }
      promptComponents.splice(0, 0, ...refComponents);
    }

    // history symbols
    const symbolPreface = "// Code from surrounding files:";
    let symbolBudget = Math.round(
      (remainingBudget - tokenCost(symbolPreface, 1)) *
        proportions.historySymbols
    );
    if (symbolBudget > 0) {
      const symbolComponents: string[] = [];
      for (const { item, symbols } of dedupedReverseHistory) {
        let isDone = false;
        if (symbols.length === 0) {
          continue;
        }
        for (const [i, symbol] of symbols.entries()) {
          const s =
            i === 0
              ? `// ${item.uri}\n${getCondensedText(symbol)}`
              : getCondensedText(symbol);
          const cost = tokenCost(s, 1);
          if (cost > symbolBudget) {
            isDone = true;
            break;
          }
          symbolComponents.push(s);
          symbolBudget -= cost;
        }
        if (isDone) {
          break;
        }
      }
      promptComponents.splice(0, 0, symbolPreface, ...symbolComponents);
    }
    return promptComponents.join("\n");
  };

// eslint-disable-next-line @typescript-eslint/naming-convention
export const prompt_refs_history_codeblocks =
  (lookback: number) =>
  ({
    prefix,
    history,
    references,

    maxTokens,
    tokenCost,
  }: PromptOpt): string => {
    let remainingBudget = maxTokens;
    const adjustedPrefix =
      prefix.length > lookback
        ? prefix.substring(prefix.length - lookback)
        : prefix;
    const promptComponents: string[] = [
      `Code to complete:\n\`\`\`\n${adjustedPrefix}`,
    ];

    remainingBudget -= tokenCost(promptComponents.join("\n"));
    if (remainingBudget < 0) {
      throw new Error("not enough token budget for prefix");
    }
    if (!history) {
      return promptComponents.join("\n");
    }
    const dedupedReverseHistory: InflatedHistoryItem[] = [];
    const seenURIs: { [uri: string]: boolean } = {};
    for (const item of [...history].reverse()) {
      if (seenURIs[item.item.uri.toString()]) {
        continue;
      }
      seenURIs[item.item.uri.toString()] = true;
      dedupedReverseHistory.push(item);
    }

    const relativeProportions = {
      historySymbols: 0.5 * (history ? 1 : 0),
      references: 0.5 * (references ? 1 : 0),
    };
    const total = Object.values(relativeProportions).reduce((x, y) => x + y, 0);
    const proportions = Object.fromEntries(
      Object.entries(relativeProportions).map(([specifier, proportion]) => [
        specifier,
        proportion / total,
      ])
    );

    // references
    let refBudget = Math.round(remainingBudget * proportions.historySymbols);
    if (references && refBudget > 0) {
      const refComponents: string[] = [];
      for (const ref of [...references].reverse()) {
        const component = `Example from ${ref.filename}\n\`\`\`\n${ref.text}\n\`\`\``;
        const cost = tokenCost(component);
        if (cost > refBudget) {
          break;
        }
        refComponents.push(component);
        refBudget -= cost;
      }
      promptComponents.splice(0, 0, ...refComponents);
    }

    // history symbols
    let symbolBudget = remainingBudget * proportions.historySymbols;
    if (symbolBudget > 0) {
      const symbolComponents: string[] = [];
      for (const { item, symbols } of dedupedReverseHistory) {
        let isDone = false;
        if (symbols.length === 0) {
          continue;
        }
        for (const [i, symbol] of symbols.entries()) {
          const s =
            i === 0
              ? `${item.uri}\n\`\`\`\n${getCondensedText(symbol)}\n\`\`\``
              : `\`\`\`\n${getCondensedText(symbol)}\n\`\`\``;
          const cost = tokenCost(s, 1);
          if (cost > symbolBudget) {
            isDone = true;
            break;
          }
          symbolComponents.push(s);
          symbolBudget -= cost;
        }
        if (isDone) {
          break;
        }
      }
      promptComponents.splice(0, 0, ...symbolComponents);
    }
    return promptComponents.join("\n");
  };

// eslint-disable-next-line @typescript-eslint/naming-convention
export const prompt_prefixonly =
  (lookback: number) =>
  ({
    prefix,

    maxTokens,
  }: PromptOpt): string => {
    let remainingBudget = maxTokens;
    const adjustedPrefix =
      prefix.length > lookback
        ? prefix.substring(prefix.length - lookback)
        : prefix;
    const promptComponents: string[] = [
      `Code to complete:\n\`\`\`\n${adjustedPrefix}`,
    ];

    return promptComponents.join("\n");
  };

export function newlineStopString(): string[] {
  return ["\n"];
}

export function endCodeBlockStopStrings(): string[] {
  return ["```"];
}

export function langKeywordStopStrings(uri: string): string[] | undefined {
  const ext = uri.split(".").pop()?.toLowerCase();
  // switch statement on ext
  switch (ext) {
    case "ts":
      return ["\nfunction ", "\nclass ", "\nexport ", "\n```"];
    case "go":
      return ["\nfunc ", "\ntype "];
		default:
			console.error(`no stopwords defined for language ${ext}, defaulting to no stopwords`)
			return undefined
  }
}
