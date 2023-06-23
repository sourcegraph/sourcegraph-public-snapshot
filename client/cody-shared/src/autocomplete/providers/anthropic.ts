import * as anthropic from '@anthropic-ai/sdk'

import { Completion } from '..'
import { Message } from '../../sourcegraph-api'
import { SourcegraphNodeCompletionsClient } from '../../sourcegraph-api/completions/nodeClient'
import { messagesToText } from '../utils'

import { Provider, ProviderConfig, ProviderOptions } from './provider'

const COMPLETIONS_PREAMBLE = `You are Cody, a code completion AI developed by Sourcegraph.
You only respond in a single Markdown code blocks to all questions.
All answers must be valid {lang} programs.
DO NOT respond with anything other than code.`

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

    private emptyPromptLength(injectPrefix?: string): number {
        const promptNoSnippets = messagesToText(this.createPromptPrefix(injectPrefix))
        return promptNoSnippets.length - 10 // extra 10 chars of buffer cuz who knows
    }

    private createPromptPrefix(injectPrefix: string = ''): Message[] {
        // TODO(beyang): escape 'Human:' and 'Assistant:'
        const prefixLines = this.prefix.split('\n')
        if (prefixLines.length === 0) {
            throw new Error('no prefix lines')
        }

        let prefixMessages: Message[]
        if (prefixLines.length > 2) {
            const endLine = Math.max(Math.floor(prefixLines.length / 2), prefixLines.length - 5)
            prefixMessages = [
                {
                    speaker: 'human',
                    text: COMPLETIONS_PREAMBLE.replace('{lang}', this.languageId),
                },
                {
                    speaker: 'assistant',
                    text: '```\n// Ok```',
                },
                {
                    speaker: 'human',
                    text:
                        'Complete the following file:\n' +
                        '```' +
                        `\n${prefixLines.slice(0, endLine).join('\n')}\n` +
                        '```',
                },
                {
                    speaker: 'assistant',
                    text: `\`\`\`\n${prefixLines.slice(endLine).join('\n')}${injectPrefix}`,
                },
            ]
        } else {
            prefixMessages = [
                {
                    speaker: 'human',
                    text: 'Write some code',
                },
                {
                    speaker: 'assistant',
                    text: `Here is some code:\n\`\`\`\n${this.prefix}${injectPrefix}`,
                },
            ]
        }

        return prefixMessages
    }

    // Creates the resulting prompt and adds as many snippets from the reference
    // list as possible.
    protected createPrompt(injectPrefix?: string): Message[] {
        const prefixMessages = this.createPromptPrefix(injectPrefix)
        const referenceSnippetMessages: Message[] = []

        let remainingChars = this.promptChars - this.emptyPromptLength(injectPrefix)

        if (this.suffix.length > 0) {
            let suffix = ''
            // We throw away the first 5 lines of the suffix to avoid the LLM to
            // just continue the completion by appending the suffix.
            const suffixLines = this.suffix.split('\n')
            if (suffixLines.length > 5) {
                suffix = suffixLines.slice(5).join('\n')
            }

            if (suffix.length > 0) {
                const suffixContext: Message[] = [
                    {
                        speaker: 'human',
                        text:
                            'Add the following code snippet to your knowledge base:\n' +
                            '```' +
                            `\n${suffix}\n` +
                            '```',
                    },
                    {
                        speaker: 'assistant',
                        text: '```\n// Ok```',
                    },
                ]

                const numSnippetChars = messagesToText(suffixContext).length + 1
                if (numSnippetChars <= remainingChars) {
                    referenceSnippetMessages.push(...suffixContext)
                    remainingChars -= numSnippetChars
                }
            }
        }

        for (const snippet of this.snippets) {
            const snippetMessages: Message[] = [
                {
                    speaker: 'human',
                    text:
                        `Add the following code snippet (from file ${snippet.fileName}) to your knowledge base:\n` +
                        '```' +
                        `\n${snippet.content}\n` +
                        '```',
                },
                {
                    speaker: 'assistant',
                    text: 'Okay, I have added it to my knowledge base.',
                },
            ]
            const numSnippetChars = messagesToText(snippetMessages).length + 1
            if (numSnippetChars > remainingChars) {
                break
            }
            referenceSnippetMessages.push(...snippetMessages)
            remainingChars -= numSnippetChars
        }

        return [...referenceSnippetMessages, ...prefixMessages]
    }

    public async generateCompletions(abortSignal: AbortSignal): Promise<Completion[]> {
        // TODO: Bring back the logic with injectPrefix \n when the current line is non empty

        // Create prompt
        const prompt = this.createPrompt()
        if (prompt.length > this.promptChars) {
            throw new Error('prompt length exceeded maximum allowed chars')
        }

        const params = {
            messages: prompt,
            stopSequences:
                this.multilineMode !== null ? [anthropic.HUMAN_PROMPT, '\n\n\n'] : [anthropic.HUMAN_PROMPT, '\n'],
            maxTokensToSample: this.responseTokens,
            temperature: 1,
            topK: -1,
            topP: -1,
        }

        // Issue requests
        const responses = await Promise.all(
            Array.from({ length: this.n }).map(() => this.completionsClient.complete(params, abortSignal))
        )

        return responses.map(resp => ({
            prefix: this.prefix,
            content: resp.completion,
            stopReason: resp.stopReason,
        }))
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
