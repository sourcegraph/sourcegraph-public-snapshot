import { LRUCache } from 'lru-cache'
import * as vscode from 'vscode'

import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { logEvent } from '../event-logger'
import { debug } from '../log'

import { CompletionsCache } from './cache'
import { getContext } from './context'
import { CompletionsDocumentProvider } from './docprovider'
import { History } from './history'
import { CompletionProvider, InlineCompletionProvider, ManualCompletionProvider } from './provider'

const LOG_MANUAL = { type: 'manual' }

function lastNLines(text: string, n: number): string {
    const lines = text.split('\n')
    return lines.slice(Math.max(0, lines.length - n)).join('\n')
}

const inlineCompletionsCache = new CompletionsCache()

export class CodyCompletionItemProvider implements vscode.InlineCompletionItemProvider {
    private promptTokens: number
    private maxPrefixTokens: number
    private maxSuffixTokens: number
    private abortOpenInlineCompletions: () => void = () => {}
    private abortOpenManualCompletion: () => void = () => {}
    private lastContentChanges: LRUCache<string, 'add' | 'del'> = new LRUCache<string, 'add' | 'del'>({
        max: 10,
    })

    constructor(
        private webviewErrorMessager: (error: string) => Promise<void>,
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

        vscode.workspace.onDidChangeTextDocument(event => {
            const document = event.document
            const text = event.contentChanges[0].text

            this.lastContentChanges.set(document.fileName, text.length > 0 ? 'add' : 'del')
        })
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
            console.error(error)
            debug('CodyCompletionProvider:inline:error', `${error.toString()}\n${error.stack}`)
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

        const { prefix, suffix, prevLine: precedingLine } = docContext

        // Avoid showing completions when we're deleting code (Cody can only insert code at the
        // moment)
        const lastChange = this.lastContentChanges.get(document.fileName) ?? 'add'
        if (lastChange === 'del') {
            // When a line was deleted, only look up cached items and only include them if the
            // untruncated prefix matches. This fixes some weird issues where the completion would
            // render if you insert whitespace but not on the original place when you delete it
            // again
            const cachedCompletions = inlineCompletionsCache.get(prefix, false)
            if (cachedCompletions) {
                return cachedCompletions.map(toInlineCompletionItem)
            }
            return []
        }

        const cachedCompletions = inlineCompletionsCache.get(prefix)
        if (cachedCompletions) {
            return cachedCompletions.map(toInlineCompletionItem)
        }

        const remainingChars = this.tokToChar(this.promptTokens)

        const completionNoSnippets = new InlineCompletionProvider(
            this.completionsClient,
            remainingChars,
            this.responseTokens,
            [],
            prefix,
            suffix,
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

        let timeout: number
        const completers: CompletionProvider[] = []

        // VS Code does not show completions if we are in the process of writing a word or if a
        // selected completion info is present (so something is selected from the completions
        // dropdown list based on the lang server) and the returned completion range does not
        // contain the same selection.
        if (context.selectedCompletionInfo || /[A-Za-z]$/.test(precedingLine)) {
            return []
        }

        let multilineMode: null | 'block' = null

        // TODO(philipp-spiess): Add a better detection for start-of-block and don't require C like
        // languages.
        const multilineEnabledLanguage =
            document.languageId === 'typescript' || document.languageId === 'javascript' || document.languageId === 'go'
        if (
            multilineEnabledLanguage &&
            // Only trigger multiline inline suggestions for empty lines
            precedingLine.trim() === '' &&
            // Only trigger multiline inline suggestions for the beginning of blocks
            prefix.trim().at(prefix.trim().length - 1) === '{'
        ) {
            timeout = 500
            multilineMode = 'block'
            completers.push(
                new InlineCompletionProvider(
                    this.completionsClient,
                    remainingChars,
                    this.responseTokens,
                    similarCode,
                    prefix,
                    suffix,
                    '',
                    3,
                    multilineMode
                )
            )
        } else if (precedingLine.trim() === '') {
            // Start of line: medium debounce
            timeout = 200
            completers.push(
                new InlineCompletionProvider(
                    this.completionsClient,
                    remainingChars,
                    this.responseTokens,
                    similarCode,
                    prefix,
                    suffix,
                    '',
                    2 // tries
                )
            )
        } else if (context.triggerKind === vscode.InlineCompletionTriggerKind.Invoke || precedingLine.endsWith('.')) {
            // Do nothing
            return []
        } else {
            // End of line: long debounce, complete until newline
            timeout = 500
            completers.push(
                new InlineCompletionProvider(
                    this.completionsClient,
                    remainingChars,
                    this.responseTokens,
                    similarCode,
                    prefix,
                    suffix,
                    '',
                    2 // tries
                ),
                // Create a completion request for the current prefix with a new line added. This
                // will make for faster recommendations when the user presses enter.
                new InlineCompletionProvider(
                    this.completionsClient,
                    remainingChars,
                    this.responseTokens,
                    similarCode,
                    prefix,
                    suffix,
                    '\n', // force a new line in the case we are at end of line
                    1 // tries
                )
            )
        }

        await new Promise<void>(resolve => setTimeout(resolve, timeout))

        // We don't need to make a request at all if the signal is already aborted after the
        // debounce
        if (abortController.signal.aborted) {
            return []
        }

        const logParams = {
            type: 'inline',
            multilineMode,
        }

        logEvent('CodyVSCodeExtension:completion:started', logParams, logParams)
        const start = Date.now()

        const results = rankCompletions(
            (await Promise.all(completers.map(c => c.generateCompletions(abortController.signal)))).flat()
        )

        if (hasVisibleCompletions(results)) {
            const logParamsWithTimings = { ...logParams, latency: Date.now() - start, timeout }
            logEvent('CodyVSCodeExtension:completion:suggested', logParamsWithTimings, logParamsWithTimings)
            inlineCompletionsCache.add(results)
            return results.map(toInlineCompletionItem)
        }
        return []
    }

    public async fetchAndShowManualCompletions(): Promise<void> {
        this.abortOpenManualCompletion()
        const abortController = new AbortController()
        this.abortOpenManualCompletion = () => abortController.abort()

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
        const { prefix, suffix } = docContext

        const remainingChars = this.tokToChar(this.promptTokens)

        const completionNoSnippets = new ManualCompletionProvider(
            this.completionsClient,
            remainingChars,
            this.responseTokens,
            [],
            prefix,
            suffix,
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

        const completer = new ManualCompletionProvider(
            this.completionsClient,
            remainingChars,
            this.responseTokens,
            similarCode,
            prefix,
            suffix,
            ''
        )

        try {
            logEvent('CodyVSCodeExtension:completion:started', LOG_MANUAL, LOG_MANUAL)
            const completions = await completer.generateCompletions(abortController.signal, 3)
            this.documentProvider.addCompletions(completionsUri, ext, completions, {
                suffix: '',
                elapsedMillis: 0,
                llmOptions: null,
            })
            logEvent('CodyVSCodeExtension:completion:suggested', LOG_MANUAL, LOG_MANUAL)
        } catch (error) {
            if (error.message === 'aborted') {
                return
            }

            await this.webviewErrorMessager(`FetchAndShowCompletions - ${error}`)
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
    messages: Message[]
    content: string
    stopReason?: string
}

function toInlineCompletionItem(completion: Completion): vscode.InlineCompletionItem {
    return new vscode.InlineCompletionItem(completion.content, undefined, {
        title: 'Completion accepted',
        command: 'cody.completions.inline.accepted',
    })
}

function rankCompletions(completions: Completion[]): Completion[] {
    // TODO(philipp-spiess): Improve ranking to something more complex then just length
    return completions.sort((a, b) => b.content.split('\n').length - a.content.split('\n').length)
}

function hasVisibleCompletions(completions: Completion[]): boolean {
    return completions.length > 0 && !!completions.find(c => c.content.trim() !== '')
}
