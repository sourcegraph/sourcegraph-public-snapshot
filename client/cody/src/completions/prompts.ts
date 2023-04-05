import * as anthropic from '@anthropic-ai/sdk'

import { ReferenceSnippet } from './context'

export interface Message {
    role: 'human' | 'ai'
    text: string | null
}

export function messagesToText(messages: Message[]): string {
    return messages
        .map(
            message =>
                `${message.role === 'human' ? anthropic.HUMAN_PROMPT : anthropic.AI_PROMPT}${
                    message.text === null ? '' : ' ' + message.text
                }`
        )
        .join('')
}

export interface PromptTemplate {
    make(bytesBudget: number, prefix: string, suffix: string, snippets: ReferenceSnippet[]): string
    postProcess(completion: string, prefix: string): string
}

export class SingleLinePromptTemplate implements PromptTemplate {
    make(bytesBudget: number, prefix: string, suffix: string, snippets: ReferenceSnippet[]): string {
        // TODO(beyang): escape 'Human:' and 'Assistant:'
        // console.log(`>>>${prefix}<<<`)

        const prefixLines = prefix.split('\n')
        if (prefixLines.length === 0) {
            throw new Error('no prefix lines')
        }

        const referenceSnippetMessages: Message[] = []
        let prefixMessages: Message[]
        if (prefixLines.length > 2) {
            const endLine = Math.max(Math.floor(prefixLines.length / 2), prefixLines.length - 5)
            prefixMessages = [
                {
                    role: 'human',
                    text:
                        `Complete the following file:\n` +
                        '```' +
                        `\n${prefixLines.slice(0, endLine).join('\n')}\n` +
                        '```',
                },
                {
                    role: 'ai',
                    text:
                        `Here is the completion of the file:\n` + '```' + `\n${prefixLines.slice(endLine).join('\n')}`,
                },
            ]
        } else {
            prefixMessages = [
                {
                    role: 'human',
                    text: 'Write some code',
                },
                {
                    role: 'ai',
                    text: `Here is some code:\n` + '```' + `\n${prefix}`,
                },
            ]
        }

        return messagesToText([...referenceSnippetMessages, ...prefixMessages])
    }
    postProcess(completion: string, prefix: string): string {
        if (completion.length > 0 && completion[0] === ' ' && prefix.length > 0 && prefix[prefix.length - 1] === ' ') {
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
    make(
        bytesBudget: number,
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
            const endLine = Math.max(Math.floor(prefixLines.length / 2), prefixLines.length - 5)
            prefixMessages = [
                {
                    role: 'human',
                    text:
                        `Complete the following file:\n` +
                        '```' +
                        `\n${prefixLines.slice(0, endLine).join('\n')}\n` +
                        '```',
                },
                {
                    role: 'ai',
                    text:
                        `Here is the completion of the file:\n` + '```' + `\n${prefixLines.slice(endLine).join('\n')}`,
                },
            ]
        } else {
            prefixMessages = [
                {
                    role: 'human',
                    text: 'Write some code',
                },
                {
                    role: 'ai',
                    text: `Here is some code:\n` + '```' + `\n${prefix}`,
                },
            ]
        }

        const promptNoSnippets = messagesToText([...referenceSnippetMessages, ...prefixMessages])
        let remainingBytes = bytesBudget - promptNoSnippets.length - 10 // extra 10 bytes of buffer cuz who knows
        for (const snippet of snippets) {
            const snippetMessages: Message[] = [
                {
                    role: 'human',
                    text:
                        `Add the following code snippet (from file ${snippet.filename}) to your knowledge base:\n` +
                        '```' +
                        `\n${snippet.text}\n` +
                        '```',
                },
                {
                    role: 'ai',
                    text: 'Okay, I have added it to my knowledge base.',
                },
            ]
            const numSnippetBytes = messagesToText(snippetMessages).length + 1
            console.log(`# numSnippetBytes: ${numSnippetBytes}, remainingBytes: ${remainingBytes}`)
            if (numSnippetBytes > remainingBytes) {
                break
            }
            referenceSnippetMessages.push(...snippetMessages)
            remainingBytes -= numSnippetBytes
        }

        return messagesToText([...referenceSnippetMessages, ...prefixMessages])
    }

    postProcess(completion: string): string {
        const endBlockIndex = completion.indexOf('```')
        if (endBlockIndex !== -1) {
            return completion.slice(0, endBlockIndex).trimEnd()
        }
        return completion.trimEnd()
    }
}
