import * as anthropic from '@anthropic-ai/sdk'

import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'

import { ReferenceSnippet } from './context'

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

export interface PromptTemplate {
    make(charsBudget: number, prefix: string, suffix: string, snippets: ReferenceSnippet[]): string
    postProcess(completion: string, prefix: string): string
}

export class SingleLinePromptTemplate implements PromptTemplate {
    public make(charsBudget: number, prefix: string, suffix: string, snippets: ReferenceSnippet[]): string {
        // TODO(beyang): escape 'Human:' and 'Assistant:'
        // console.log(`>>>${prefix}<<<`)

        const prefixLines = prefix.split('\n')
        if (prefixLines.length === 0) {
            throw new Error('no prefix lines')
        }

        const referenceSnippetMessages: Message[] = []
        let prefixMessages: Message[]
        if (prefixLines.length > 2) {
            const lastHumanLine = Math.max(Math.floor(prefixLines.length / 2), prefixLines.length - 5)
            prefixMessages = [
                {
                    speaker: 'human',
                    text:
                        'Complete the following file:\n' +
                        '```' +
                        `\n${prefixLines.slice(0, lastHumanLine).join('\n')}\n` +
                        '```',
                },
                {
                    speaker: 'assistant',
                    text:
                        'Here is the completion of the file:\n' +
                        '```' +
                        `\n${prefixLines.slice(lastHumanLine).join('\n')}`,
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

        return messagesToText([...referenceSnippetMessages, ...prefixMessages])
    }
    public postProcess(completion: string, prefix: string): string {
        if (completion.length > 0 && completion.startsWith(' ') && prefix.length > 0 && prefix.endsWith(' ')) {
            completion = completion.slice(1)
        }
        const endBlockIndex = completion.indexOf('```')
        if (endBlockIndex !== -1) {
            return completion.slice(0, endBlockIndex).trimEnd()
        }
        return completion.trimEnd()
    }
}

export class KnowledgeBasePromptTemplate implements PromptTemplate {
    public make(
        charsBudget: number,
        prefix: string,
        suffix: string, // TODO(beyang)
        snippets: ReferenceSnippet[]
    ): string {
        // TODO(beyang): escape 'Human:' and 'Assistant:'
        prefix = prefix.trim()

        const prefixLines = prefix.split('\n')
        if (prefixLines.length === 0) {
            throw new Error('no prefix lines')
        }

        const referenceSnippetMessages: Message[] = []
        let prefixMessages: Message[]
        if (prefixLines.length > 2) {
            const lastHumanLine = Math.max(Math.floor(prefixLines.length / 2), prefixLines.length - 5)
            prefixMessages = [
                {
                    speaker: 'human',
                    text:
                        'Complete the following file:\n' +
                        '```' +
                        `\n${prefixLines.slice(0, lastHumanLine).join('\n')}\n` +
                        '```',
                },
                {
                    speaker: 'assistant',
                    text:
                        'Here is the completion of the file:\n' +
                        '```' +
                        `\n${prefixLines.slice(lastHumanLine).join('\n')}`,
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

        const promptNoSnippets = messagesToText([...referenceSnippetMessages, ...prefixMessages])
        let remainingChars = charsBudget - promptNoSnippets.length - 10 // extra 10 chars of buffer cuz who knows
        for (const snippet of snippets) {
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

        return messagesToText([...referenceSnippetMessages, ...prefixMessages])
    }

    public postProcess(completion: string): string {
        const endBlockIndex = completion.indexOf('```')
        if (endBlockIndex !== -1) {
            return completion.slice(0, endBlockIndex).trimEnd()
        }
        return completion.trimEnd()
    }
}
