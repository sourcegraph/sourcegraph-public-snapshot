import { OpenAIApi } from 'openai'
import * as vscode from 'vscode'

import { CompletionsDocumentProvider } from './docprovider'

export class CodyCompletionItemProvider implements vscode.InlineCompletionItemProvider {
    private maxPrefixTokens: number
    private maxSuffixTokens: number
    constructor(
        private openai: OpenAIApi,
        private documentProvider: CompletionsDocumentProvider,
        private model = 'gpt-3.5-turbo',
        private contextWindowTokens = 2048, // 8001
        private bytesPerToken = 4,
        private responseTokens = 200,
        private prefixPercentage = 0.9,
        private suffixPercentage = 0.1
    ) {
        const promptTokens = this.contextWindowTokens - this.responseTokens
        this.maxPrefixTokens = Math.floor(promptTokens * this.prefixPercentage)
        this.maxSuffixTokens = Math.floor(promptTokens * this.suffixPercentage)
    }

    async provideInlineCompletionItems(
        document: vscode.TextDocument,
        position: vscode.Position,
        context: vscode.InlineCompletionContext,
        token: vscode.CancellationToken
    ): Promise<vscode.InlineCompletionItem[]> {
        try {
            return this.provideInlineCompletionItemsInner(document, position, context, token)
        } catch (error) {
            vscode.window.showErrorMessage(error)
            return []
        }
    }

    private tokToByte(toks: number): number {
        return Math.floor(toks * this.bytesPerToken)
    }

    private async provideInlineCompletionItemsInner(
        document: vscode.TextDocument,
        position: vscode.Position,
        context: vscode.InlineCompletionContext,
        token: vscode.CancellationToken
    ): Promise<vscode.InlineCompletionItem[]> {
        // Require manual invocation
        if (context.triggerKind === vscode.InlineCompletionTriggerKind.Automatic) {
            return []
        }

        const docContext = getCurrentDocContext(
            document,
            position,
            this.tokToByte(this.maxPrefixTokens),
            this.tokToByte(this.maxSuffixTokens)
        )
        if (!docContext) {
            return []
        }

        const { prefix, prevLine: precedingLine } = docContext
        let waitMs: number
        let completionPrefix = '' // text to require as the first part of the completion
        if (precedingLine.trim() === '') {
            // Start of line: medium debounce, allow multiple lines
            waitMs = 1000
        } else if (context.triggerKind === vscode.InlineCompletionTriggerKind.Invoke || precedingLine.endsWith('.')) {
            // Middle of line: long debounce, next line
            waitMs = 100
        } else {
            // End of line: long debounce, next line
            completionPrefix = '\n'
sto            waitMs = 2000

            // TODO(beyang): handle this as a special case, try 2 completions, one with newline inserted, one without
        }

        const aborter = new AbortController()
        token.onCancellationRequested(() => aborter.abort())

        const waiter = new Promise<void>(resolve => setTimeout(() => resolve(), waitMs))
        const completionsPromise = this.openai.createChatCompletion(
            {
                model: this.model,
                messages: [
                    {
                        role: 'system',
                        content: 'Complete whatever code you obtain from the user through the end of the line.',
                    },
                    {
                        role: 'user',
                        content: prefix + completionPrefix,
                    },
                ],
                max_tokens: Math.min(this.contextWindowTokens - this.maxPrefixTokens, this.responseTokens),
                n: 1,
            },
            {
                signal: aborter.signal,
            }
        )
        await waiter
        let completions
        try {
            completions = await completionsPromise
        } catch (error) {
            throw new Error(`error fetching completions from OpenAI: ${error}`)
        }
        if (token.isCancellationRequested) {
            return []
        }

        if (completions.data.choices.length === 0) {
            throw new Error('no completions')
        }

        const inlineCompletions: vscode.InlineCompletionItem[] = []
        for (const choice of completions.data.choices) {
            if (!choice.message?.content) {
                continue
            }
            inlineCompletions.push(new vscode.InlineCompletionItem(completionPrefix + choice.message.content))
        }
        return inlineCompletions
    }

    async fetchAndShowCompletions(): Promise<void> {
        const currentEditor = vscode.window.activeTextEditor
        if (!currentEditor || currentEditor?.document.uri.scheme === 'cody') {
            return
        }
        const filename = currentEditor.document.fileName
        const ext = filename.split('.').pop() || ''
        const completionsUri = vscode.Uri.parse('cody:Completions.md')
        this.documentProvider.clearCompletions(completionsUri)

        const doc = vscode.workspace.openTextDocument(completionsUri)
        doc.then(doc => {
            vscode.window.showTextDocument(doc, {
                preview: false,
                viewColumn: 2,
            })
        })

        const docContext = getCurrentDocContext(
            currentEditor.document,
            currentEditor.selection.start,
            this.tokToByte(this.maxPrefixTokens),
            this.tokToByte(this.maxSuffixTokens)
        )
        if (docContext === null) {
            console.error('not showing completions, no currently open doc')
            return
        }
        const { prefix, suffix, prevLine, prevNonEmptyLine } = docContext

        try {
            const completion = await this.openai.createChatCompletion({
                model: this.model,
                messages: [
                    {
                        role: 'system',
                        content:
                            'Complete whatever code you obtain from the user up to the end of the function or block scope.',
                    },
                    {
                        role: 'user',
                        content: prefix,
                    },
                ],
                max_tokens: Math.min(this.contextWindowTokens - this.maxPrefixTokens, this.responseTokens),
                n: 3,
            })

            // Trim lines that go past current indent
            for (const choice of completion.data.choices) {
                if (!choice.message?.content) {
                    continue
                }
                const indent = getIndent(prevNonEmptyLine)
                choice.message.content = trimToIndent(choice.message.content, indent)
            }

            this.documentProvider.addCompletions(completionsUri, ext, prevLine, completion.data, {
                prompt: prefix,
                suffix,
                elapsedMillis: 0,
                llmOptions: null,
            })
        } catch (error) {
            if (error.response) {
                console.error(error.response.status)
                console.error(error.response.data)
            } else {
                console.error(error.message)
            }
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

function getIndent(line: string): string {
    const match = /^(\s*)/.exec(line)
    if (!match || match.length < 2) {
        return ''
    }
    return match[1]
}

function trimToIndent(text: string, indent: string): string {
    const lines = text.split('\n')
    // Iterate through the lines starting at the second line (always include the first line)
    for (let i = 1; i < lines.length; i++) {
        if (lines[i].trim().length === 0) {
            continue
        }
        const lineIndent = getIndent(lines[i])
        if (indent.indexOf(lineIndent) === 0 && lineIndent.length < indent.length) {
            return lines.slice(0, i).join('\n')
        }
    }
    return text
}
