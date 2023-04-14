import * as anthropic from '@anthropic-ai/sdk'
import { ChatCompletionRequestMessage } from 'openai'
import * as vscode from 'vscode'

import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import {
    CodeCompletionParameters,
    CodeCompletionResponse,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { ReferenceSnippet, getContext } from './context'
import { CompletionsDocumentProvider } from './docprovider'
import { History } from './history'
import { Message, messagesToText } from './prompts'

function lastNLines(text: string, n: number): string {
    const lines = text.split('\n')
    return lines.slice(Math.max(0, lines.length - n)).join('\n')
}

const estimatedLLMResponseLatencyMS = 700

export class CodyCompletionItemProvider implements vscode.InlineCompletionItemProvider {
    private promptTokens: number
    private maxPrefixTokens: number
    private maxSuffixTokens: number
    constructor(
        private completionsClient: SourcegraphNodeCompletionsClient,
        private documentProvider: CompletionsDocumentProvider,
        private history: History,
        private contextWindowTokens = 2048, // 8001
        private charsPerToken = 4,
        private responseTokens = 200,
        private prefixPercentage = 0.6,
        private suffixPercentage = 0.1
    ) {
        this.promptTokens = this.contextWindowTokens - this.responseTokens
        this.maxPrefixTokens = Math.floor(this.promptTokens * this.prefixPercentage)
        this.maxSuffixTokens = Math.floor(this.promptTokens * this.suffixPercentage)
    }

    public async provideInlineCompletionItems(
        document: vscode.TextDocument,
        position: vscode.Position,
        context: vscode.InlineCompletionContext,
        token: vscode.CancellationToken
    ): Promise<vscode.InlineCompletionItem[]> {
        try {
            return await this.provideInlineCompletionItemsInner(document, position, context, token)
        } catch (error) {
            await vscode.window.showErrorMessage(`Error in provideInlineCompletionItems: ${error}`)
            return []
        }
    }

    private tokToChar(toks: number): number {
        return toks * this.charsPerToken
    }

    private async provideInlineCompletionItemsInner(
        document: vscode.TextDocument,
        position: vscode.Position,
        context: vscode.InlineCompletionContext,
        token: vscode.CancellationToken
    ): Promise<vscode.InlineCompletionItem[]> {
        const docContext = getCurrentDocContext(
            document,
            position,
            this.tokToChar(this.maxPrefixTokens),
            this.tokToChar(this.maxSuffixTokens)
        )
        if (!docContext) {
            return []
        }

        const { prefix, prevLine: precedingLine } = docContext
        let waitMs: number
        const remainingChars = this.tokToChar(this.promptTokens)
        const completers: CompletionProvider[] = []
        if (precedingLine.trim() === '') {
            // Start of line: medium debounce
            waitMs = 1500
            completers.push(
                new EndOfLineCompletionProvider(
                    this.completionsClient,
                    remainingChars,
                    this.responseTokens,
                    prefix,
                    '',
                    2
                )
            )
        } else if (context.triggerKind === vscode.InlineCompletionTriggerKind.Invoke || precedingLine.endsWith('.')) {
            // Do nothing
            return []
        } else {
            // End of line: long debounce, complete until newline
            waitMs = 3000
            completers.push(
                new EndOfLineCompletionProvider(
                    this.completionsClient,
                    remainingChars,
                    this.responseTokens,
                    prefix,
                    '',
                    2 // 2 tries
                ),
                new EndOfLineCompletionProvider(
                    this.completionsClient,
                    remainingChars,
                    this.responseTokens,
                    prefix,
                    '\n', // force a new line in the case we are at end of line
                    2 // 2 tries
                )
            )
        }

        const aborter = new AbortController()
        token.onCancellationRequested(() => aborter.abort())

        // TODO(beyang): trigger on context quality (better context means longer completion)

        const waiter = new Promise<void>(resolve =>
            setTimeout(() => resolve(), Math.max(0, waitMs - estimatedLLMResponseLatencyMS))
        )
        await waiter
        const results = (await Promise.all(completers.map(c => c.generateCompletions()))).flat()
        return results.map(r => new vscode.InlineCompletionItem(r.content))
    }

    public async fetchAndShowCompletions(): Promise<void> {
        const currentEditor = vscode.window.activeTextEditor
        if (!currentEditor || currentEditor?.document.uri.scheme === 'cody') {
            return
        }
        const filename = currentEditor.document.fileName
        const ext = filename.split('.').pop() || ''
        const completionsUri = vscode.Uri.parse('cody:Completions.md')
        this.documentProvider.clearCompletions(completionsUri)

        const doc = await vscode.workspace.openTextDocument(completionsUri)
        await vscode.window.showTextDocument(doc, {
            preview: false,
            viewColumn: 2,
        })

        // TODO(beyang): make getCurrentDocContext fetch complete line prefix
        const docContext = getCurrentDocContext(
            currentEditor.document,
            currentEditor.selection.start,
            this.tokToChar(this.maxPrefixTokens),
            this.tokToChar(this.maxSuffixTokens)
        )
        if (docContext === null) {
            console.error('not showing completions, no currently open doc')
            return
        }
        const { prefix } = docContext

        // TODO: better remaining context calculation
        const systemMessage: ChatCompletionRequestMessage = {
            role: 'system',
            content: 'Complete whatever code you obtain from the user up to the end of the function or block scope.',
        }
        const l = (systemMessage.role + ': ' + systemMessage.content + '\n').length + prefix.length + 'user: \n'.length
        const contextChars = this.tokToChar(this.promptTokens) - l

        const windowSize = 20
        const similarCode = await getContext(
            currentEditor,
            this.history,
            lastNLines(prefix, windowSize),
            windowSize,
            contextChars
        )

        const remainingChars = this.tokToChar(this.promptTokens)

        const completer = new MultilineCompletionProvider(
            this.completionsClient,
            remainingChars,
            this.responseTokens,
            similarCode,
            prefix
        )

        try {
            const completions = await completer.generateCompletions(3)
            this.documentProvider.addCompletions(completionsUri, ext, completions, {
                suffix: '',
                elapsedMillis: 0,
                llmOptions: null,
            })
        } catch (error) {
            await vscode.window.showErrorMessage(`Error in provideInlineCompletionItems: ${error}`)
        }
    }
}

function getCurrentDocContext(
    document: vscode.TextDocument,
    position: vscode.Position,
    maxPrefixLength: number,
    maxSuffixLength: number
): {
    prefix: string
    suffix: string
    prevLine: string
    prevNonEmptyLine: string
    nextNonEmptyLine: string
} | null {
    const offset = document.offsetAt(position)
    const prefixLines = document.getText(new vscode.Range(new vscode.Position(0, 0), position)).split('\n')
    if (prefixLines.length === 0) {
        console.error('no lines')
        return null
    }

    const suffixLines = document
        .getText(new vscode.Range(position, document.positionAt(document.getText().length)))
        .split('\n')
    let nextNonEmptyLine = ''
    if (suffixLines.length > 0) {
        for (const line of suffixLines) {
            if (line.trim().length > 0) {
                nextNonEmptyLine = line
                break
            }
        }
    }

    let prevNonEmptyLine = ''
    for (let i = prefixLines.length - 1; i >= 0; i--) {
        const line = prefixLines[i]
        if (line.trim().length > 0) {
            prevNonEmptyLine = line
            break
        }
    }

    const prevLine = prefixLines[prefixLines.length - 1]

    let prefix: string
    if (offset > maxPrefixLength) {
        let total = 0
        let startLine = prefixLines.length
        for (let i = prefixLines.length - 1; i >= 0; i--) {
            if (total + prefixLines[i].length > maxPrefixLength) {
                break
            }
            startLine = i
            total += prefixLines[i].length
        }
        prefix = prefixLines.slice(startLine).join('\n')
    } else {
        prefix = document.getText(new vscode.Range(new vscode.Position(0, 0), position))
    }

    let totalSuffix = 0
    let endLine = 0
    for (let i = 0; i < suffixLines.length; i++) {
        if (totalSuffix + suffixLines[i].length > maxSuffixLength) {
            break
        }
        endLine = i + 1
        totalSuffix += suffixLines[i].length
    }
    const suffix = suffixLines.slice(0, endLine).join('\n')

    return {
        prefix,
        suffix,
        prevLine,
        prevNonEmptyLine,
        nextNonEmptyLine,
    }
}

async function batchCompletions(
    client: SourcegraphNodeCompletionsClient,
    params: CodeCompletionParameters,
    n: number
): Promise<CodeCompletionResponse[]> {
    const responses: Promise<CodeCompletionResponse>[] = []
    for (let i = 0; i < n; i++) {
        responses.push(client.complete(params))
    }
    return Promise.all(responses)
}

export interface Completion {
    prompt: string
    content: string
    stopReason?: string
}

interface CompletionProvider {
    generateCompletions(n?: number): Promise<Completion[]>
}

export class MultilineCompletionProvider implements CompletionProvider {
    constructor(
        private completionsClient: SourcegraphNodeCompletionsClient,
        private promptChars: number,
        private responseTokens: number,
        private snippets: ReferenceSnippet[],
        private prefix: string,
        private defaultN: number = 1
    ) {}

    private makePrompt(): string {
        // TODO(beyang): escape 'Human:' and 'Assistant:'
        const prefix = this.prefix.trim()

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
                        'Complete the following file:\n' +
                        '```' +
                        `\n${prefixLines.slice(0, endLine).join('\n')}\n` +
                        '```',
                },
                {
                    role: 'ai',
                    text: `Here is the completion of the file:\n\`\`\`\n${prefixLines.slice(endLine).join('\n')}`,
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
                    text: `Here is some code:\n\`\`\`\n${prefix}`,
                },
            ]
        }

        const promptNoSnippets = messagesToText([...referenceSnippetMessages, ...prefixMessages])
        let remainingChars = this.promptChars - promptNoSnippets.length - 10 // extra 10 chars of buffer cuz who knows
        for (const snippet of this.snippets) {
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
            const numSnippetChars = messagesToText(snippetMessages).length + 1
            if (numSnippetChars > remainingChars) {
                break
            }
            referenceSnippetMessages.push(...snippetMessages)
            remainingChars -= numSnippetChars
        }

        return messagesToText([...referenceSnippetMessages, ...prefixMessages])
    }
    private postProcess(completion: string): string {
        const endBlockIndex = completion.indexOf('```')
        if (endBlockIndex !== -1) {
            return completion.slice(0, endBlockIndex).trimEnd()
        }
        return completion.trimEnd()
    }

    public async generateCompletions(n?: number): Promise<Completion[]> {
        // Create prompt
        const prompt = this.makePrompt()
        if (prompt.length > this.promptChars) {
            throw new Error('prompt length exceeded maximum alloted chars')
        }

        // Issue request
        const responses = await batchCompletions(
            this.completionsClient,
            {
                prompt,
                stopSequences: [anthropic.HUMAN_PROMPT],
                maxTokensToSample: this.responseTokens,
                model: 'claude-instant-v1.0',
                temperature: 1, // default value (source: https://console.anthropic.com/docs/api/reference)
                topK: -1, // default value
                topP: -1, // default value
            },
            n || this.defaultN
        )
        // Post-process
        return responses.map(resp => ({
            prompt,
            content: this.postProcess(resp.completion),
            stopReason: resp.stopReason,
        }))
    }
}

export class EndOfLineCompletionProvider implements CompletionProvider {
    constructor(
        private completionsClient: SourcegraphNodeCompletionsClient,
        private promptChars: number,
        private responseTokens: number,
        private prefix: string,
        private injectPrefix: string,
        private defaultN: number = 1
    ) {}

    private makePrompt(): string {
        // TODO(beyang): escape 'Human:' and 'Assistant:'
        const prefixLines = this.prefix.split('\n')
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
                        'Complete the following file:\n' +
                        '```' +
                        `\n${prefixLines.slice(0, endLine).join('\n')}\n` +
                        '```',
                },
                {
                    role: 'ai',
                    text:
                        'Here is the completion of the file:\n' +
                        '```' +
                        `\n${prefixLines.slice(endLine).join('\n')}${this.injectPrefix}`,
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
                    text: `Here is some code:\n\`\`\`\n${this.prefix}${this.injectPrefix}`,
                },
            ]
        }

        return messagesToText([...referenceSnippetMessages, ...prefixMessages])
    }

    private postProcess(completion: string): string {
        // Sometimes Claude emits an extra space
        if (
            completion.length > 0 &&
            completion.startsWith(' ') &&
            this.prefix.length > 0 &&
            this.prefix.endsWith(' ')
        ) {
            completion = completion.slice(1)
        }
        // Insert the injected prefix back in
        if (this.injectPrefix.length > 0) {
            completion = this.injectPrefix + completion
        }
        // Strip out trailing markdown block and trim trailing whitespace
        const endBlockIndex = completion.indexOf('```')
        if (endBlockIndex !== -1) {
            return completion.slice(0, endBlockIndex).trimEnd()
        }
        return completion.trimEnd()
    }

    public async generateCompletions(n?: number): Promise<Completion[]> {
        // Create prompt
        const prompt = this.makePrompt()
        if (prompt.length > this.promptChars) {
            throw new Error('prompt length exceeded maximum alloted chars')
        }

        // Issue request
        const responses = await batchCompletions(
            this.completionsClient,
            {
                prompt,
                stopSequences: [anthropic.HUMAN_PROMPT, '\n'],
                maxTokensToSample: this.responseTokens,
                model: 'claude-instant-v1.0',
                temperature: 1,
                topK: -1,
                topP: -1,
            },
            n || this.defaultN
        )
        // Post-process
        return responses.map(resp => ({
            prompt,
            content: this.postProcess(resp.completion),
            stopReason: resp.stopReason,
        }))
    }
}
