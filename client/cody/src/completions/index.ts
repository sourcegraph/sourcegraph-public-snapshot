import * as vscode from 'vscode'

import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { logEvent } from '../event-logger'

import { CompletionsCache } from './cache'
import { getContext } from './context'
import { CompletionsDocumentProvider } from './docprovider'
import { History } from './history'
import { CompletionProvider, EndOfLineCompletionProvider, MultilineCompletionProvider } from './provider'

const LOG_INLINE = { type: 'inline' }
const LOG_MULTILINE = { type: 'multiline' }

function lastNLines(text: string, n: number): string {
    const lines = text.split('\n')
    return lines.slice(Math.max(0, lines.length - n)).join('\n')
}

const estimatedLLMResponseLatencyMS = 700
const inlineCompletionsCache = new CompletionsCache()

export class CodyCompletionItemProvider implements vscode.InlineCompletionItemProvider {
    private promptTokens: number
    private maxPrefixTokens: number
    private maxSuffixTokens: number
    private abortOpenInlineCompletions: () => void = () => {}
    private abortOpenMultilineCompletion: () => void = () => {}

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
            if (error.message === 'aborted') {
                return []
            }

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
        this.abortOpenInlineCompletions()
        const abortController = new AbortController()
        token.onCancellationRequested(() => abortController.abort())
        this.abortOpenInlineCompletions = () => abortController.abort()

        const currentEditor = vscode.window.activeTextEditor
        if (!currentEditor || currentEditor?.document.uri.scheme === 'cody') {
            return []
        }

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

        const cachedCompletions = inlineCompletionsCache.get(prefix)
        if (cachedCompletions) {
            return cachedCompletions.map(toInlineCompletionItem)
        }

        const remainingChars = this.tokToChar(this.promptTokens)

        const completionNoSnippets = new MultilineCompletionProvider(
            this.completionsClient,
            remainingChars,
            this.responseTokens,
            [],
            prefix,
            '\n'
        )
        const emptyPromptLength = completionNoSnippets.emptyPromptLength()

        const contextChars = this.tokToChar(this.promptTokens) - emptyPromptLength

        const windowSize = 20
        const similarCode = await getContext(
            currentEditor,
            this.history,
            lastNLines(prefix, windowSize),
            windowSize,
            contextChars
        )

        let waitMs: number
        const completers: CompletionProvider[] = []
        if (precedingLine.trim() === '') {
            // Start of line: medium debounce
            waitMs = 500
            completers.push(
                new EndOfLineCompletionProvider(
                    this.completionsClient,
                    remainingChars,
                    this.responseTokens,
                    similarCode,
                    prefix,
                    '',
                    2 // tries
                )
            )
        } else if (context.triggerKind === vscode.InlineCompletionTriggerKind.Invoke || precedingLine.endsWith('.')) {
            // Do nothing
            return []
        } else {
            // End of line: long debounce, complete until newline
            waitMs = 1000
            completers.push(
                new EndOfLineCompletionProvider(
                    this.completionsClient,
                    remainingChars,
                    this.responseTokens,
                    similarCode,
                    prefix,
                    '',
                    2 // tries
                ),
                // Create a completion request for the current prefix with a new line added. This
                // will make for faster recommendations when the user presses enter.
                new EndOfLineCompletionProvider(
                    this.completionsClient,
                    remainingChars,
                    this.responseTokens,
                    similarCode,
                    prefix,
                    '\n', // force a new line in the case we are at end of line
                    1 // tries
                )
            )
        }

        // TODO(beyang): trigger on context quality (better context means longer completion)

        await new Promise<void>(resolve =>
            setTimeout(() => resolve(), Math.max(0, waitMs - estimatedLLMResponseLatencyMS))
        )

        // We don't need to make a request at all if the signal is already aborted after the
        // debounce
        if (abortController.signal.aborted) {
            return []
        }

        logEvent('CodyVSCodeExtension:completion:started', LOG_INLINE, LOG_INLINE)

        const results = (await Promise.all(completers.map(c => c.generateCompletions(abortController.signal)))).flat()

        inlineCompletionsCache.add(results)

        logEvent('CodyVSCodeExtension:completion:suggested', LOG_INLINE, LOG_INLINE)

        return results.map(toInlineCompletionItem)
    }

    public async fetchAndShowCompletions(): Promise<void> {
        this.abortOpenMultilineCompletion()
        const abortController = new AbortController()
        this.abortOpenMultilineCompletion = () => abortController.abort()

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

        const remainingChars = this.tokToChar(this.promptTokens)

        const completionNoSnippets = new MultilineCompletionProvider(
            this.completionsClient,
            remainingChars,
            this.responseTokens,
            [],
            prefix,
            ''
        )
        const emptyPromptLength = completionNoSnippets.emptyPromptLength()

        const contextChars = this.tokToChar(this.promptTokens) - emptyPromptLength

        const windowSize = 20
        const similarCode = await getContext(
            currentEditor,
            this.history,
            lastNLines(prefix, windowSize),
            windowSize,
            contextChars
        )

        const completer = new MultilineCompletionProvider(
            this.completionsClient,
            remainingChars,
            this.responseTokens,
            similarCode,
            prefix,
            ''
        )

        try {
            logEvent('CodyVSCodeExtension:completion:started', LOG_MULTILINE, LOG_MULTILINE)
            const completions = await completer.generateCompletions(abortController.signal, 3)
            this.documentProvider.addCompletions(completionsUri, ext, completions, {
                suffix: '',
                elapsedMillis: 0,
                llmOptions: null,
            })
            logEvent('CodyVSCodeExtension:completion:suggested', LOG_MULTILINE, LOG_MULTILINE)
        } catch (error) {
            if (error.message === 'aborted') {
                return
            }

            await vscode.window.showErrorMessage(`Error in fetchAndShowCompletions: ${error}`)
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

export interface Completion {
    prefix: string
    prompt: string
    content: string
    stopReason?: string
}

function toInlineCompletionItem(completion: Completion): vscode.InlineCompletionItem {
    return new vscode.InlineCompletionItem(completion.content, undefined, {
        title: 'Completion accepted',
        command: 'cody.completions.inline.accepted',
    })
}
