import * as anthropic from '@anthropic-ai/sdk'

import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import {
    CompletionParameters,
    CompletionResponse,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

export function messagesToText(messages: Message[]): string {
    return messages
        .map(
            message =>
                `${message.speaker === 'human' ? anthropic.HUMAN_PROMPT : anthropic.AI_PROMPT}${
                    message.text === undefined ? '' : ' ' + message.text
                }`
        )
        .join('')
}

/**
 * The size of the Jaccard distance match window in number of lines. It determines how many
 * lines of the 'matchText' are considered at once when searching for a segment
 * that is most similar to the 'targetText'. In essence, it sets the maximum number
 * of lines that the best match can be. A larger 'windowSize' means larger potential matches
 */
export const SNIPPET_WINDOW_SIZE = 50

export function lastNLines(text: string, n: number): string {
    const lines = text.split('\n')
    return lines.slice(Math.max(0, lines.length - n)).join('\n')
}

/**
 * This function slices the suggestion string until the first n lines match the suffix string.
 *
 * It splits suggestion and suffix into lines, then iterates over the lines of suffix. For each line
 * of suffix, it checks if the next n lines of suggestion match. If so, it returns the first part of
 * suggestion up to those matching lines. If no match is found after iterating all lines of suffix,
 * the full suggestion is returned.
 *
 * For example, with:
 * suggestion = "foo\nbar\nbaz\nqux\nquux"
 * suffix = "baz\nqux\nquux"
 * n = 3
 *
 * It would return: "foo\nbar"
 *
 * Because the first 3 lines of suggestion ("baz\nqux\nquux") match suffix.
 */
export function sliceUntilFirstNLinesOfSuffixMatch(suggestion: string, suffix: string, n: number): string {
    const suggestionLines = suggestion.split('\n')
    const suffixLines = suffix.split('\n')

    for (let i = 0; i < suffixLines.length; i++) {
        let matchedLines = 0
        for (let j = 0; j < suggestionLines.length; j++) {
            if (suffixLines.length < i + matchedLines) {
                continue
            }
            if (suffixLines[i + matchedLines] === suggestionLines[j]) {
                matchedLines += 1
            } else {
                matchedLines = 0
            }
            if (matchedLines >= n) {
                return suggestionLines.slice(0, j - n + 1).join('\n')
            }
        }
    }

    return suggestion
}

export async function batchCompletions(
    client: SourcegraphNodeCompletionsClient,
    params: CompletionParameters,
    n: number,
    abortSignal: AbortSignal
): Promise<CompletionResponse[]> {
    const responses: Promise<CompletionResponse>[] = []
    for (let i = 0; i < n; i++) {
        responses.push(client.complete(params, abortSignal))
    }
    return Promise.all(responses)
}

export function isAbortError(error: Error): boolean {
    return (
        // http module
        error.message === 'aborted' ||
        // fetch
        error.message.includes('The operation was aborted') ||
        error.message.includes('The user aborted a request')
    )
}
