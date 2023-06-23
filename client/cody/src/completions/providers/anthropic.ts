import * as anthropic from '@anthropic-ai/sdk'

import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import { CompletionParameters } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { Completion } from '..'
import { truncateMultilineCompletion } from '../multiline'
import {
    OPENING_CODE_TAG,
    CLOSING_CODE_TAG,
    extractFromCodeBlock,
    fixBadCompletionStart,
    PrefixComponents,
    getHeadAndTail,
    trimUntilSuffix,
    trimStartUntilNewline,
} from '../text-processing'
import { batchCompletions, messagesToText } from '../utils'

import { Provider, ProviderConfig, ProviderOptions } from './provider'

const CHARS_PER_TOKEN = 4

function tokensToChars(tokens: number): number {
    return tokens * CHARS_PER_TOKEN
}

interface AnthropicOptions {
    contextWindowTokens: number
    completionsClient: SourcegraphNodeCompletionsClient
}

export class AnthropicProvider extends Provider {
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

    private postProcess(rawResponse: string): string {
        let completion = extractFromCodeBlock(rawResponse)

        const trimmedPrefixContainNewline = this.prefix.slice(this.prefix.trimEnd().length).includes('\n')
        if (trimmedPrefixContainNewline) {
            // The prefix already contains a `\n` that Claude was not aware of, so we remove any
            // leading `\n` that claude might add.
            completion = completion.trimStart()
        } else {
            completion = trimStartUntilNewline(completion)
        }

        // Remove bad symbols from the start of the completion string.
        completion = fixBadCompletionStart(completion)

        // Remove incomplete lines in single-line completions
        if (this.multilineMode === null) {
            const allowedNewlines = 2
            const lines = completion.split('\n')
            if (lines.length >= allowedNewlines) {
                completion = lines.slice(0, allowedNewlines).join('\n')
            }
        }

        // Trim start and end of the completion to remove all trailing whitespace.
        return completion.trimEnd()
    }

    public async generateCompletions(abortSignal: AbortSignal, n?: number): Promise<Completion[]> {
        // Create prompt
        const { messages: prompt } = this.createPrompt()
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
                    stopSequences: [anthropic.HUMAN_PROMPT, CLOSING_CODE_TAG],
                }
                break
            }
            default: {
                args = {
                    temperature: 0.5,
                    messages: prompt,
                    maxTokensToSample: Math.min(100, this.responseTokens),
                    stopSequences: [anthropic.HUMAN_PROMPT, CLOSING_CODE_TAG, '\n\n'],
                }
                break
            }
        }

        // Issue request
        const responses = await batchCompletions(this.completionsClient, args, n || this.n, abortSignal)

        // Post-process
        const ret = responses.map(resp => {
            const content = this.postProcess(resp.completion)

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
