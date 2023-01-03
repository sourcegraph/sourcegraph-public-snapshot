import { InflatedSymbol, LLMDebugInfo, CompletionsArgs } from "common";


export interface CompletionsBackend {
  expectedResponses: number;
  getCompletions({
    uri,
    prefix,
    history,
    references,
  }: CompletionsArgs): Promise<{
    debug: LLMDebugInfo;
    completions: string[];
  }>;
}

export function getCondensedText(s: InflatedSymbol): string {
  const lines = s.text.split("\n");
  if (lines.length < 10) {
    return s.text;
  }
  return [
    ...lines.slice(0, 5),
    "// (omitted code)",
    ...lines.slice(lines.length - 5, lines.length),
  ].join("\n");
}

export function tokenCost(s: string, assumeExtraNewlines?: number): number {
  const charsPerToken = 4;
  return s.length + (assumeExtraNewlines || 0) / charsPerToken;
}
