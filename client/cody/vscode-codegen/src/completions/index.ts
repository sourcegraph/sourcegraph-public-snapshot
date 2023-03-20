import { OpenAIApi } from 'openai'
import * as vscode from 'vscode'

import { CompletionsDocumentProvider } from './docprovider'

export class CodyCompletionItemProvider implements vscode.InlineCompletionItemProvider {
    private maxPrefixTokens: number
    private maxSuffixTokens: number
    constructor(
        private openai: OpenAIApi,
        private documentProvider: CompletionsDocumentProvider,
        private model = 'code-cushman-001', // code-davinci-002
        private contextWindowTokens = 2048, // 8001
        private bytesPerToken = 4,
        private responseTokens = 200,
        private prefixPercentage = 0.9,
        private suffixPercentage = 0.1,
        private codeContextPercentage = 0.0
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

        const docContext = getCurrentDocContext(document, position, this.tokToByte(this.maxPrefixTokens))
        if (!docContext) {
            return []
        }

        const { prefix, prevLine: precedingLine } = docContext
        let waitMs: number
        let stop = null
        let completionPrefix = '' // text to require as the first part of the completion
        if (precedingLine.trim() === '') {
            // Start of line: medium debounce, allow multiple lines
            waitMs = 1000
        } else if (context.triggerKind === vscode.InlineCompletionTriggerKind.Invoke || precedingLine.endsWith('.')) {
            // Middle of line: long debounce, next line
            waitMs = 100
            stop = ['\n']
        } else {
            // End of line: long debounce, next line
            completionPrefix = '\n'
            stop = ['\n']
            waitMs = 2000

            // TODO(beyang): handle this as a special case, try 2 completions, one with newline inserted, one without
        }

        const aborter = new AbortController()
        token.onCancellationRequested(() => aborter.abort())

        const waiter = new Promise<void>(resolve => setTimeout(() => resolve(), waitMs))
        const completionsPromise = this.openai.createCompletion(
            {
                model: this.model,
                prompt: prefix + completionPrefix,
                max_tokens: Math.min(this.contextWindowTokens - this.maxPrefixTokens, this.responseTokens),
                n: 1,
                stop,
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
            // throw new
            //     Error(`error condition after completion fetch: no choices (response: ${JSON.stringify(completions.data)})`)
            throw new Error('no completions')
        }

        const inlineCompletions: vscode.InlineCompletionItem[] = []
        for (const choice of completions.data.choices) {
            if (!choice.text) {
                continue
            }
            inlineCompletions.push(new vscode.InlineCompletionItem(completionPrefix + choice.text))
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
            this.tokToByte(this.maxPrefixTokens)
        )
        if (docContext === null) {
            console.error('not showing completions, no currently open doc')
            return
        }
        const { prefix, prevLine, prevNonEmptyLine } = docContext

        try {
            const completion = await this.openai.createCompletion({
                model: this.model,
                prompt: prefix,
                max_tokens: Math.min(this.contextWindowTokens - this.maxPrefixTokens, this.responseTokens),
                n: 3,
            })

            // Trim lines that go past current indent
            for (const choice of completion.data.choices) {
                if (!choice.text) {
                    continue
                }
                const indent = getIndent(prevNonEmptyLine)
                choice.text = trimToIndent(choice.text, indent)
            }

            this.documentProvider.addCompletions(completionsUri, ext, prevLine, completion.data, {
                prompt: prefix,
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
    maxPrefixLength: number
): {
    prefix: string
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
        return {
            prefix: prefixLines.slice(startLine).join('\n'),
            prevLine,
            prevNonEmptyLine,
            nextNonEmptyLine,
        }
    }

    return {
        prefix: document.getText(new vscode.Range(new vscode.Position(0, 0), position)),
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
