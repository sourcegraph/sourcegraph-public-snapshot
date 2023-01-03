import { performance } from "perf_hooks";
import {
  CreateSimpleSamplingStreamOptions,
  ModelParams,
  createSimpleSamplingStream,
} from "@completion/sampling";
import { extractUntilTripleBacktick } from "./extract";
import { CompletionsArgs, InflatedHistoryItem, LLMDebugInfo, Message, ReferenceInfo } from "common";
import {
  getCondensedText,
} from "./common";
import { ErrorEvent } from "ws";

interface ClaudeMock {
  createSimpleSamplingStream: (
    options: CreateSimpleSamplingStreamOptions<any>
  ) => void;
}

export class ClaudeBackend {
  accessToken: string;
  modelParams: ModelParams;
  mock?: ClaudeMock;

  readonly expectedResponses: number;

  constructor(
    accessToken: string,
    modelParams: ModelParams,
    mock?: ClaudeMock
  ) {
    this.accessToken = accessToken;
    this.modelParams = modelParams;
;
    this.mock = mock;
    this.expectedResponses = 1;
  }

  chat(messages: Message[], callbacks: ChatCallbacks): void {
    // basic verification
    let lastSpeaker: "bot" | "you" | undefined;
    for (const msg of messages) {
      if (msg.speaker === lastSpeaker) {
        throw new Error(`duplicate speaker ${lastSpeaker}`);
      }
      lastSpeaker = msg.speaker;
    }
    if (lastSpeaker !== "you") {
      throw new Error("last speaker was not human");
    }

    const promptComponents: string[] = [];
    for (const msg of messages) {
      promptComponents.push(
        `\n\n${msg.speaker === "bot" ? "Assistant:" : "Human"}: ${msg.text}`
      );
    }
    promptComponents.push("\n\nAssistant: ");
    const prompt = promptComponents.join("");
    const stream =
      this.mock?.createSimpleSamplingStream || createSimpleSamplingStream;
    stream({
      basicAuth: `Basic ${this.accessToken}`,
      prompt,
      model: this.modelParams,
      onSampleChange: (text) => callbacks.onChange(text),
      onSampleComplete: (text) => callbacks.onComplete(text),
      onError: (err, _, originalErrorEvent) =>
        callbacks.onError(err, originalErrorEvent),
    });
  }

  async refactor(text: string, transformation: string): Promise<string> {
    const prompt = `Rewrite this code with the following transformation: ${transformation}\n\`\`\`\n${text}\n\`\`\`\nPlace the result between \`\`\` delimiters.`;
    const stream =
      this.mock?.createSimpleSamplingStream || createSimpleSamplingStream;
    const response = await new Promise<string>((resolve, reject) => {
      stream({
        basicAuth: `Basic ${this.accessToken}`,
        prompt: `\n\nHuman: ${prompt}\n\nAssistant:`,
        model: this.modelParams,
        onSampleComplete: (completion) => resolve(completion),
        onError: (err) => reject(err),
      });
    });
    console.log("# response:", response);
    const start = response.indexOf("```");
    if (start === -1) {
      throw new Error(`Could not extract result from response:\n${response}`);
    }
    const end = response.indexOf("```", start + 3);
    return response.substring(start + 3, end);
  }

  async getCompletions({
    uri,
    prefix,
    history,
    references,
  }: CompletionsArgs, promptGenerator: PromptGenerateFunction): Promise<{
    debug: LLMDebugInfo;
    completions: string[];
  }> {
    const startTime = performance.now();
    const maxLookback = 5000;
    const maxLookforward = 2000;
    if (prefix.length > maxLookback) {
      prefix = prefix.substring(prefix.length - maxLookback);
    }

    const prompt = promptGenerator({ prefix, history, references }); // NEXT: pass history and references through to here

    const stream =
      this.mock?.createSimpleSamplingStream || createSimpleSamplingStream;
    const completion = await new Promise<string>((resolve, reject) => {
      stream({
        model: this.modelParams,
        basicAuth: `Basic ${this.accessToken}`,
        prompt: `\n\nHuman: ${prompt}\n\nAssistant:`,
        onSampleComplete: (completion) => resolve(completion),
        onError: (err) => reject(err),
      });
    });
    const endTime = performance.now();

    return {
      completions: [extractUntilTripleBacktick(completion)], // NEXT: pass this in
      debug: {
        elapsedMillis: endTime - startTime,
        prompt: prompt,
        llmOptions: this.modelParams,
      },
    };
  }
}

interface PromptGeneratorOpt {
  maxInputTokens: number;
  lookback: number; // in chars, not tokens
  tokenCost: (s: string, assumeExtraNewlines?: number) => number;
}
interface PromptArgs {
  prefix: string;
  history?: InflatedHistoryItem[];
  references?: ReferenceInfo[];
}

type PromptGenerateFunction = ({
  prefix,
  history,
  references,
}: PromptArgs) => string;

export const prompt_claude_prefixonly =
  ({
    maxInputTokens: maxTokens,
    lookback,
    tokenCost,
  }: PromptGeneratorOpt): PromptGenerateFunction =>
  ({ prefix, history, references }: PromptArgs): string => {
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

interface ChatCallbacks {
  onChange: (text: string) => void;
  onComplete: (text: string) => void;
  onError: (message: string, originalErrorEvent?: ErrorEvent) => void;
}
