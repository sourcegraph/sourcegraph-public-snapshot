import { InflatedSymbol, LLMDebugInfo, CompletionsArgs, CompletionLogProbs } from '@sourcegraph/cody-common'

export interface CompletionsBackend {
    expectedResponses: number
    getCompletions({ uri, prefix, history, references }: CompletionsArgs): Promise<{
        debug: LLMDebugInfo
        completions: string[]
    }>
}

export function getCondensedText(s: InflatedSymbol): string {
    const lines = s.text.split('\n')
    if (lines.length < 10) {
        return s.text
    }
    return [...lines.slice(0, 5), '// (omitted code)', ...lines.slice(-5, lines.length)].join('\n')
}

export const charsPerTokenOpenAI = 3

export function tokenCost(s: string, assumeExtraNewlines?: number): number {
    return Math.ceil(s.length + (assumeExtraNewlines || 0) / charsPerTokenOpenAI)
}

export function tokenCountToChars(tokenCount: number): number {
    return Math.floor(tokenCount * charsPerTokenOpenAI)
}

const indentWithContentRegex = /^(\s*)\S/
const maxPrefixLines = 5

export function enhanceCompletion(
    prefix: string,
    rawCompletion: string,
    stopPatterns: RegExp[]
): {
    prefixText: string
    insertText: string
} {
    const completion = dontRamble(rawCompletion, stopPatterns)
    const completionLines = completion.split('\n')
    let minCompletionIndent: string | null = null
    for (const line of completionLines.slice(1)) {
        // ignore the first line because completion might start mid-line
        const match = indentWithContentRegex.exec(line)
        if (!match || match.length !== 2) {
            continue
        }
        const indent = match[1]
        if (minCompletionIndent) {
            if (indent.length < minCompletionIndent.length) {
                minCompletionIndent = indent
            }
        } else {
            minCompletionIndent = indent
        }
    }
    if (minCompletionIndent === null) {
        return { prefixText: '', insertText: completion }
    }

    const prefixLines = prefix.split('\n')
    const firstLineMatch = indentWithContentRegex.exec(prefixLines[prefixLines.length - 1] + completionLines[0])
    if (firstLineMatch && firstLineMatch.length === 2) {
        const firstLineIndent = firstLineMatch[1]
        if (firstLineIndent <= minCompletionIndent) {
            return { prefixText: '', insertText: completion }
        }
    }

    let prefixStartLine = prefixLines.length - maxPrefixLines
    for (let i = prefixLines.length - 1; i >= Math.max(prefixLines.length - maxPrefixLines, 0); i--) {
        const match = indentWithContentRegex.exec(prefixLines[i])
        if (!match || match.length !== 2) {
            continue
        }
        const indent = match[1]
        if (indent.length <= minCompletionIndent.length) {
            prefixStartLine = i
            break
        }
    }
    return {
        prefixText: prefixLines.slice(prefixStartLine).join('\n'),
        insertText: completion,
    }
}

function dontRamble(s: string, stopPatterns: RegExp[]): string {
    let completionEndIndex = s.length
    for (const stopPattern of stopPatterns) {
        const match = stopPattern.exec(s)
        if (!match) {
            continue
        }
        if (match.index < completionEndIndex) {
            completionEndIndex = match.index
        }
    }
    return s.slice(0, Math.max(0, completionEndIndex))
}

export function truncateByProbability(
    minLogprob: number,
    logprobs?: CompletionLogProbs
): {
    truncatedInsertText: string
    removed: string
} {
    if (!logprobs) {
        throw new Error('logprobs undefined')
    }
    const { tokens, tokenLogprobs } = logprobs
    if (!tokens || !tokenLogprobs) {
        throw new Error('tokens or tokenLogprobs undefined')
    }
    if (tokens.length !== tokenLogprobs.length) {
        throw new Error('tokens and tokenLogprobs lengths not equal')
    }
    let logprobSum = 0
    let finalTokenIndex = tokens.length
    for (let i = 0; i < tokens.length; i++) {
        logprobSum += tokenLogprobs[i]
        if (logprobSum < minLogprob) {
            finalTokenIndex = i + 1
            break
        }
    }

    return {
        truncatedInsertText: tokens.slice(0, finalTokenIndex).join(''),
        removed: tokens.slice(finalTokenIndex).join(''),
    }
}

export function getCharCountLimitedPrefixAtLineBreak(prefix: string, charLimit: number): string {
    const lines = prefix.split('\n')
    const chosen: string[] = []
    let totalBytes = 0
    for (let i = lines.length - 1; i >= 0; i--) {
        const byteCount = lines[i].length + 1
        if (totalBytes + byteCount > charLimit) {
            break
        }

        chosen.push(lines[i])
        totalBytes += lines[i].length + 1
    }
    return chosen.reverse().join('\n')
}
