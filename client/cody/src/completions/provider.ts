import * as anthropic from '@anthropic-ai/sdk'

import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import {
    CompletionParameters,
    CompletionResponse,
    Message,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { Completion } from '.'
import { ReferenceSnippet } from './context'
import { truncateMultilineCompletion } from './multiline'
import { messagesToText } from './prompts'

const COMPLETIONS_PREAMBLE = `You are Cody, a code completion AI developed by Sourcegraph.
You only respond in a single Markdown code blocks to all questions.
All answers must be valid {lang} programs.
DO NOT respond with anything other than code.`

const BAD_COMPLETION_START = /^(\p{Emoji_Presentation}|\u{200B}|\+ |- |. )+(\s)+/u

export abstract class CompletionProvider {
    constructor(
        protected completionsClient: SourcegraphNodeCompletionsClient,
        protected promptChars: number,
        protected responseTokens: number,
        protected snippets: ReferenceSnippet[],
        protected prefix: string,
        protected suffix: string,
        protected injectPrefix: string,
        protected languageId: string,
        protected defaultN: number = 1
    ) {}

    // Returns the content specific prompt excluding additional referenceSnippets
    protected abstract createPromptPrefix(): Message[]

    public emptyPromptLength(): number {
        const promptNoSnippets = messagesToText(this.createPromptPrefix())
        return promptNoSnippets.length - 10 // extra 10 chars of buffer cuz who knows
    }

    // Creates the resulting prompt and adds as many snippets from the reference
    // list as possible.
    protected createPrompt(): Message[] {
        const prefixMessages = this.createPromptPrefix()
        const referenceSnippetMessages: Message[] = []

        let remainingChars = this.promptChars - this.emptyPromptLength()

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
                        `Add the following code snippet (from file ${snippet.filename}) to your knowledge base:\n` +
                        '```' +
                        `\n${snippet.text}\n` +
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

    public abstract generateCompletions(abortSignal: AbortSignal, n?: number): Promise<Completion[]>
}

export class ManualCompletionProvider extends CompletionProvider {
    protected createPromptPrefix(): Message[] {
        // TODO(beyang): escape 'Human:' and 'Assistant:'
        const prefix = this.prefix.trim()

        const prefixLines = prefix.split('\n')
        if (prefixLines.length === 0) {
            throw new Error('no prefix lines')
        }

        let prefixMessages: Message[]
        if (prefixLines.length > 2) {
            const endLine = Math.max(Math.floor(prefixLines.length / 2), prefixLines.length - 5)
            prefixMessages = [
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
                    text: `Here is the completion of the file:\n\`\`\`\n${prefixLines.slice(endLine).join('\n')}`,
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
                    text: `Here is some code:\n\`\`\`\n${prefix}`,
                },
            ]
        }

        return prefixMessages
    }

    private postProcess(completion: string): string {
        let suggestion = completion
        const endBlockIndex = completion.indexOf('```')
        if (endBlockIndex !== -1) {
            suggestion = completion.slice(0, endBlockIndex)
        }

        // Remove trailing whitespace before newlines
        suggestion = suggestion
            .split('\n')
            .map(line => line.trimEnd())
            .join('\n')

        return sliceUntilFirstNLinesOfSuffixMatch(suggestion, this.suffix, 5)
    }

    public async generateCompletions(abortSignal: AbortSignal, n?: number): Promise<Completion[]> {
        const prefix = this.prefix.trim()

        // Create prompt
        const prompt = this.createPrompt()
        const textPrompt = messagesToText(prompt)
        if (textPrompt.length > this.promptChars) {
            throw new Error('prompt length exceeded maximum alloted chars')
        }

        // Issue request
        const responses = await batchCompletions(
            this.completionsClient,
            {
                messages: prompt,
                maxTokensToSample: this.responseTokens,
            },
            // We over-fetch the number of completions to account for potential
            // empty results
            (n || this.defaultN) + 2,
            abortSignal
        )
        // Post-process
        return responses
            .flatMap(resp => {
                const completion = this.postProcess(resp.completion)
                if (completion.trim() === '') {
                    return []
                }

                return [
                    {
                        prefix,
                        messages: prompt,
                        content: this.postProcess(resp.completion),
                        stopReason: resp.stopReason,
                    },
                ]
            })
            .slice(0, 3)
    }
}

export class InlineCompletionProvider extends CompletionProvider {
    constructor(
        completionsClient: SourcegraphNodeCompletionsClient,
        promptChars: number,
        responseTokens: number,
        snippets: ReferenceSnippet[],
        prefix: string,
        suffix: string,
        injectPrefix: string,
        languageId: string,
        defaultN: number = 1,
        protected multilineMode: null | 'block' | 'statement' = null
    ) {
        super(
            completionsClient,
            promptChars,
            responseTokens,
            snippets,
            prefix,
            suffix,
            injectPrefix,
            languageId,
            defaultN
        )
    }

    protected createPromptPrefix(): Message[] {
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
                    text: `\`\`\`\n${prefixLines.slice(endLine).join('\n')}${this.injectPrefix}`,
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
                    text: `Here is some code:\n\`\`\`\n${this.prefix}${this.injectPrefix}`,
                },
            ]
        }

        return prefixMessages
    }

    private postProcess(completion: string): null | string {
        // Extract a few common parts for the processing
        const currentLinePrefix = this.prefix.slice(this.prefix.lastIndexOf('\n') + 1)
        const firstNlInSuffix = this.suffix.indexOf('\n') + 1
        const nextNonEmptyLine =
            this.suffix
                .slice(firstNlInSuffix)
                .split('\n')
                .find(line => line.trim().length > 0) ?? ''

        // Sometimes Claude emits an extra space
        let hasOddIndentation = false
        if (
            completion.length > 0 &&
            completion.startsWith(' ') &&
            this.prefix.length > 0 &&
            (this.prefix.endsWith(' ') || this.prefix.endsWith('\t'))
        ) {
            completion = completion.slice(1)
            hasOddIndentation = true
        }

        // Experimental: Trim start of the completion to remove all trailing whitespace nonsense
        completion = completion.trimStart()

        // Detect bad completion start
        if (BAD_COMPLETION_START.test(completion)) {
            completion = completion.replace(BAD_COMPLETION_START, '')
        }

        // Insert the injected prefix back in
        if (this.injectPrefix.length > 0) {
            completion = this.injectPrefix + completion
        }

        // Strip out trailing markdown block and trim trailing whitespace
        const endBlockIndex = completion.indexOf('```')
        if (endBlockIndex !== -1) {
            completion = completion.slice(0, endBlockIndex)
        }

        if (this.multilineMode !== null) {
            completion = truncateMultilineCompletion(
                completion,
                hasOddIndentation,
                this.prefix,
                nextNonEmptyLine,
                this.languageId
            )
        }

        // If a completed line matches the next non-empty line of the suffix 1:1, we remove
        const lines = completion.split('\n')
        const matchedLineIndex = lines.findIndex((line, index) => {
            if (index === 0) {
                line = currentLinePrefix + line
            }
            if (line.trim() !== '' && nextNonEmptyLine.trim() !== '') {
                // We need a trimEnd here because the machine likes to add trailing whitespace.
                //
                // TODO: Fix this earlier in the post process run but this needs to be careful not
                // to alter the meaning
                return line.trimEnd() === nextNonEmptyLine
            }
            return false
        })
        if (matchedLineIndex !== -1) {
            completion = lines.slice(0, matchedLineIndex).join('\n')
        }

        return completion.trimEnd()
    }

    public async generateCompletions(abortSignal: AbortSignal, n?: number): Promise<Completion[]> {
        const prefix = this.prefix + this.injectPrefix

        // Create prompt
        const prompt = this.createPrompt()
        if (prompt.length > this.promptChars) {
            throw new Error('prompt length exceeded maximum alloted chars')
        }

        // Issue request
        const responses = await batchCompletions(
            this.completionsClient,
            {
                messages: prompt,
                stopSequences:
                    this.multilineMode !== null ? [anthropic.HUMAN_PROMPT, '\n\n\n'] : [anthropic.HUMAN_PROMPT, '\n'],
                maxTokensToSample: this.responseTokens,
                temperature: 1,
                topK: -1,
                topP: -1,
            },
            n || this.defaultN,
            abortSignal
        )

        // Post-process
        return responses.flatMap(resp => {
            const content = this.postProcess(resp.completion)

            if (content === null) {
                return []
            }

            return [
                {
                    prefix,
                    messages: prompt,
                    content,
                    stopReason: resp.stopReason,
                },
            ]
        })
    }
}

async function batchCompletions(
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
