import * as anthropic from '@anthropic-ai/sdk'

import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import { CompletionParameters } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { Completion } from '.'
import { ReferenceSnippet } from './context'
import { messagesToText } from './prompts'
import { CompletionProvider, batchCompletions } from './provider'

interface TrimmedString {
    trimmed: string
    leadSpace: string
    rearSpace: string
}

interface PrefixComponents {
    head: TrimmedString
    tail: TrimmedString
    overlap?: string
}

export class NewCompletionProvider implements CompletionProvider {
    constructor(
        protected completionsClient: SourcegraphNodeCompletionsClient,
        protected promptChars: number,
        protected responseTokens: number,
        protected snippets: ReferenceSnippet[],
        protected prefix: string,
        protected suffix: string,
        protected languageId: string,
        protected defaultN: number = 1,
        private completionType: 'single-line' | 'multi-line'
    ) {}

    // Returns the content specific prompt excluding additional referenceSnippets
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
                text: 'You are Cody, a code completion AI developed by Sourcegraph. You write code in between tags like this:<CODE5711>/* Code goes here */</CODE5711>',
            },
            {
                speaker: 'assistant',
                text: 'I am Cody, a code completion AI developed by Sourcegraph.',
            },
            {
                speaker: 'human',
                text: `Complete this code: <CODE5711>${head.trimmed}</CODE5711>.`,
            },
            {
                speaker: 'assistant',
                text: `Okay, here is some code: <CODE5711>${tail.trimmed}`,
            },
        ]
        return { messages: prefixMessages, prefix: { head, tail, overlap } }
    }

    public emptyPromptLength(): number {
        const { messages } = this.createPromptPrefix()
        const promptNoSnippets = messagesToText(messages)
        return promptNoSnippets.length - 10 // extra 10 chars of buffer cuz who knows
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
                    text: `Here is a reference snippet of code: <CODE5711>${snippet.text}</CODE5711>`,
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

        if (prefix.tail.rearSpace.length > 0) {
            const rearSpaceLines = prefix.tail.rearSpace.split('\n')
            const currentLine = rearSpaceLines[rearSpaceLines.length - 1]
            if (currentLine.match(/^\s*$/)) {
                // New line + indent: trim all leading new lines and match indent
                completion = completion.replace(/^\n*/, '')
                const completionIndent = completion.match(/^\s*/)![0]
                if (currentLine.length <= completionIndent.length) {
                    completion = completionIndent.slice(currentLine.length) + completion.slice(completionIndent.length)
                }
            }
        }

        completion = trimIndent(completion, this.suffix)
        completion = trimUntilSuffix(completion, this.suffix)

        // Remove incomplete lines in single-line completions
        if (this.completionType === 'single-line') {
            const allowedNewlines = 2
            const lines = completion.split('\n')
            if (lines.length >= allowedNewlines) {
                completion = lines.slice(0, allowedNewlines).join('\n')
            }
        }

        completion = completion.trimEnd()
        return completion
    }

    public async generateCompletions(abortSignal: AbortSignal, n?: number): Promise<Completion[]> {
        // Create prompt
        const { messages: prompt, prefix } = this.createPrompt()
        if (prompt.length > this.promptChars) {
            throw new Error('prompt length exceeded maximum alloted chars')
        }

        let args: CompletionParameters
        switch (this.completionType) {
            case 'single-line': {
                args = {
                    temperature: 0.5,
                    messages: prompt,
                    maxTokensToSample: 100,
                    stopSequences: [anthropic.HUMAN_PROMPT, '\n\n'],
                }
                break
            }
            case 'multi-line': {
                args = {
                    temperature: 0.5,
                    messages: prompt,
                    maxTokensToSample: 2000,
                    stopSequences: [anthropic.HUMAN_PROMPT],
                }
                break
            }
            default:
                throw new Error(`unrecognized completion type ${this.completionType}`)
        }

        // Issue request
        const responses = await batchCompletions(this.completionsClient, args, n || this.defaultN, abortSignal)

        // Post-process
        const ret = await Promise.all(
            responses.map(async resp => {
                const content = await this.postProcess(resp.completion, prefix)

                if (content === null) {
                    return []
                }

                return [
                    {
                        startOffset: 0 - prefix.tail.rearSpace.length,
                        prefix: this.prefix,
                        messages: prompt,
                        content,
                        stopReason: resp.stopReason,
                    },
                ]
            })
        )
        return ret.flat()
    }
}

function extractFromCodeBlock(completion: string): string {
    if (completion.includes('<CODE5711>')) {
        console.error('TODO invalid 1: ', completion)
        return ''
    }
    let end = completion.indexOf('</CODE5711>')
    if (end === -1) {
        end = completion.length
    }
    return completion.substring(0, end).trimEnd()
}

// Split string into head and tail. The tail is at most the last 2 non-empty lines of the snippet
function getHeadAndTail(s: string): PrefixComponents {
    const lines = s.split('\n')
    const tailThreshold = 2

    let nonEmptyCount = 0
    let tailStart = -1
    for (let i = lines.length - 1; i >= 0; i--) {
        if (lines[i].trim().length > 0) {
            nonEmptyCount++
        }
        if (nonEmptyCount >= tailThreshold) {
            tailStart = i
            break
        }
    }

    if (tailStart === -1) {
        return { head: trimSpace(s), tail: trimSpace(s), overlap: s }
    }

    return { head: trimSpace(lines.slice(0, tailStart).join('\n')), tail: trimSpace(lines.slice(tailStart).join('\n')) }
}

function trimSpace(s: string): TrimmedString {
    const trimmed = s.trim()
    const headEnd = s.indexOf(trimmed)
    return { trimmed, leadSpace: s.slice(0, headEnd), rearSpace: s.slice(headEnd + trimmed.length) }
}

function trimIndent(insertion: string, suffix: string): string {
    let suffixIndent = 0
    for (const line of suffix.split('\n')) {
        if (line.trim().length > 0) {
            const indentMatch = line.match(/^\s*/)
            if (indentMatch && indentMatch.length >= 1) {
                suffixIndent = indentMatch[0].length
            }
            break
        }
    }

    const insertionLines = insertion.split('\n')
    let insertionEnd = insertionLines.length
    // Skip over first line, because we expect that to always be included and to have no leading whitespace
    for (let i = 1; i < insertionLines.length; i++) {
        const line = insertionLines[i]
        if (line.trim().length === 0) {
            continue
        }
        const indentMatch = line.match(/^\s*/)
        if (indentMatch && indentMatch.length >= 1) {
            if (indentMatch[0].length < suffixIndent) {
                insertionEnd = i
                break
            }
        }
    }
    return insertionLines.slice(0, insertionEnd).join('\n')
}

function trimUntilSuffix(insertion: string, suffix: string): string {
    insertion = insertion.trimEnd()
    let firstNonEmptySuffixLine = ''
    for (const line of suffix.split('\n')) {
        if (line.trim().length > 0) {
            firstNonEmptySuffixLine = line
            break
        }
    }
    if (firstNonEmptySuffixLine.length === 0) {
        return insertion
    }

    const insertionLines = insertion.split('\n')
    let insertionEnd = insertionLines.length
    for (let i = 0; i < insertionLines.length; i++) {
        const line = insertionLines[i]
        if (line === firstNonEmptySuffixLine) {
            insertionEnd = i
            break
        }
    }
    return insertionLines.slice(0, insertionEnd).join('\n')
}
