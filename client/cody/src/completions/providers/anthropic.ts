import * as anthropic from '@anthropic-ai/sdk'

import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import { CompletionParameters } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { Completion } from '..'
import { checkOddIndentation, truncateMultilineCompletion } from '../multiline'
import {
    OPENING_CODE_TAG,
    CLOSING_CODE_TAG,
    extractFromCodeBlock,
    fixBadCompletionStart,
    PrefixComponents,
    getHeadAndTail,
    trimUntilSuffix,
} from '../text-processing'
import { batchCompletions, messagesToText } from '../utils'

import { AbstractProvider, ProviderConfig, ProviderOptions } from './provider'

const CHARS_PER_TOKEN = 4

function tokensToChars(tokens: number): number {
    return tokens * CHARS_PER_TOKEN
}

interface AnthropicOptions {
    contextWindowTokens: number
    completionsClient: SourcegraphNodeCompletionsClient
}

export class AnthropicProvider extends AbstractProvider {
    private promptChars: number
    private responseTokens: number
    private completionsClient: SourcegraphNodeCompletionsClient

    constructor(options: ProviderOptions, anthropicOptions: AnthropicOptions) {
        super(options)
        this.promptChars =
            tokensToChars(anthropicOptions.contextWindowTokens) -
            Math.floor(tokensToChars(anthropicOptions.contextWindowTokens) * options.responsePercentage)
        this.responseTokens = Math.floor(anthropicOptions.contextWindowTokens * options.responsePercentage)
        this.completionsClient = anthropicOptions.completionsClient
    }

    public emptyPromptLength(): number {
        const { messages } = this.createPromptPrefix()
        const promptNoSnippets = messagesToText(messages)
        return promptNoSnippets.length - 10 // extra 10 chars of buffer cuz who knows
    }

    private createPromptPrefix(): { messages: Message[]; prefix: PrefixComponents } {
        // TODO(beyang): escape 'Human:' and 'Assistant:'
        const prefixLines = this.prefix.split('\n')
        if (prefixLines.length === 0) {
            throw new Error('no prefix lines')
        }

        const { head, tail, overlap } = getHeadAndTail(this.prefix)
        const prefixMessages: Message[] = [
            {
                speaker: 'human',
                text: `You are Cody, a code completion AI developed by Sourcegraph. You write code in between tags like this:${OPENING_CODE_TAG}/* Code goes here */${CLOSING_CODE_TAG}`,
            },
            {
                speaker: 'assistant',
                text: 'I am Cody, a code completion AI developed by Sourcegraph.',
            },
            {
                speaker: 'human',
                text: `Complete this code: ${OPENING_CODE_TAG}${head.trimmed}${CLOSING_CODE_TAG}.`,
            },
            {
                speaker: 'assistant',
                text: `Okay, here is some code: ${OPENING_CODE_TAG}${tail.trimmed}`,
            },
        ]
        return { messages: prefixMessages, prefix: { head, tail, overlap } }
    }

    // Creates the resulting prompt and adds as many snippets from the reference
    // list as possible.
    protected createPrompt(): { messages: Message[]; prefix: PrefixComponents } {
        const { messages: prefixMessages, prefix } = this.createPromptPrefix()
        const referenceSnippetMessages: Message[] = []

        let remainingChars = this.promptChars - this.emptyPromptLength()

        for (const snippet of this.snippets) {
            const snippetMessages: Message[] = [
                {
                    speaker: 'human',
                    text: `Here is a reference snippet of code: ${OPENING_CODE_TAG}${snippet.content}${CLOSING_CODE_TAG}`,
                },
                {
                    speaker: 'assistant',
                    text: 'I have added the snippet to my knowledge base.',
                },
            ]
            const numSnippetChars = messagesToText(snippetMessages).length + 1
            if (numSnippetChars > remainingChars) {
                break
            }
            referenceSnippetMessages.push(...snippetMessages)
            remainingChars -= numSnippetChars
        }

        return { messages: [...referenceSnippetMessages, ...prefixMessages], prefix }
    }

    private postProcess(rawResponse: string, prefix: PrefixComponents): string {
        let completion = extractFromCodeBlock(rawResponse)

        // Sometimes Claude emits a single space in the completion. We call this an "odd indentation"
        // completion and try to fix the response.
        const hasOddIndentation = checkOddIndentation(completion, prefix)

        // Remove bad symbols from the start of the completion string.
        completion = fixBadCompletionStart(completion)

        // Trim start of the completion to remove all trailing whitespace.
        completion = completion.trimStart()

        // Handle multiline completion indentation and remove overlap with suffx.
        if (this.multilineMode === 'block') {
            completion = truncateMultilineCompletion(
                completion,
                this.prefix,
                this.suffix,
                hasOddIndentation,
                this.languageId
            )
            completion = trimUntilSuffix(completion, this.suffix)
        }

        // Remove incomplete lines in single-line completions
        if (this.multilineMode === null) {
            const allowedNewlines = 2
            const lines = completion.split('\n')
            if (lines.length >= allowedNewlines) {
                completion = lines.slice(0, allowedNewlines).join('\n')
            }
        }

        return completion.trimEnd()
    }

    public async generateCompletions(abortSignal: AbortSignal, n?: number): Promise<Completion[]> {
        // Create prompt
        const { messages: prompt, prefix } = this.createPrompt()
        if (prompt.length > this.promptChars) {
            throw new Error('prompt length exceeded maximum alloted chars')
        }

        let args: CompletionParameters
        switch (this.multilineMode) {
            case 'block': {
                args = {
                    temperature: 0.5,
                    messages: prompt,
                    maxTokensToSample: this.responseTokens,
                    stopSequences: [anthropic.HUMAN_PROMPT],
                }
                break
            }
            default: {
                args = {
                    temperature: 0.5,
                    messages: prompt,
                    maxTokensToSample: Math.min(100, this.responseTokens),
                    // '\n' contributed the most to a sub 1 second response latency
                    stopSequences: [anthropic.HUMAN_PROMPT, '\n', '\n\n'],
                }
                break
            }
        }

        // Issue request
        const responses = await batchCompletions(this.completionsClient, args, n || this.n, abortSignal)

        // Post-process
        const ret = responses.map(resp => {
            const content = this.postProcess(resp.completion, prefix)

            if (content === null) {
                return []
            }

            return [
                {
                    prefix: this.prefix,
                    messages: prompt,
                    content,
                    stopReason: resp.stopReason,
                },
            ]
        })

        return ret.flat()
    }
}

export function createProviderConfig(anthropicOptions: AnthropicOptions): ProviderConfig {
    return {
        create(options: ProviderOptions) {
            return new AnthropicProvider(options, anthropicOptions)
        },
        maximumContextCharacters: tokensToChars(anthropicOptions.contextWindowTokens),
        identifier: 'anthropic',
    }
}
