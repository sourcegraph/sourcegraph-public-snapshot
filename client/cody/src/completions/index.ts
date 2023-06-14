import { LRUCache } from 'lru-cache'
import * as vscode from 'vscode'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { logEvent } from '../event-logger'
import { debug } from '../log'

import { CompletionsCache } from './cache'
import { getContext } from './context'
import { CompletionsDocumentProvider } from './docprovider'
import { History } from './history'
import * as CompletionLogger from './logger'
import { detectMultilineMode } from './multiline'
import { CompletionProvider, InlineCompletionProvider, ManualCompletionProvider } from './provider'

const LOG_MANUAL = { type: 'manual' }

/**
 * The size of the Jaccard distance match window in number of lines. It determines how many
 * lines of the 'matchText' are considered at once when searching for a segment
 * that is most similar to the 'targetText'. In essence, it sets the maximum number
 * of lines that the best match can be. A larger 'windowSize' means larger potential matches
 */
const WINDOW_SIZE = 50

function lastNLines(text: string, n: number): string {
    const lines = text.split('\n')
    return lines.slice(Math.max(0, lines.length - n)).join('\n')
}

export const inlineCompletionsCache = new CompletionsCache()

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
        private codebaseContext: CodebaseContext,
        private contextWindowTokens = 2048, // 8001
        private charsPerToken = 4,
        private responseTokens = 200,
        private prefixPercentage = 0.6,
        private suffixPercentage = 0.1,
        private disableTimeouts = false
    ) {
        this.promptTokens = this.contextWindowTokens - this.responseTokens
        this.maxPrefixTokens = Math.floor(this.promptTokens * this.prefixPercentage)
        this.maxSuffixTokens = Math.floor(this.promptTokens * this.suffixPercentage)

        vscode.workspace.onDidChangeTextDocument(event => {
            const document = event.document
            const changes = event.contentChanges

            if (changes.length <= 0) {
                return
            }

            const text = changes[0].text
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

        CompletionLogger.clear()

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

        const { prefix, suffix, prevLine: sameLinePrefix, prevNonEmptyLine } = docContext
        const sameLineSuffix = suffix.slice(0, suffix.indexOf('\n'))

        // Avoid showing completions when we're deleting code (Cody can only insert code at the
        // moment)
        const lastChange = this.lastContentChanges.get(document.fileName) ?? 'add'
        if (lastChange === 'del') {
            // When a line was deleted, only look up cached items and only include them if the
            // untruncated prefix matches. This fixes some weird issues where the completion would
            // render if you insert whitespace but not on the original place when you delete it
            // again
            const cachedCompletions = inlineCompletionsCache.get(prefix, false)
            if (cachedCompletions?.isExactPrefix) {
                return toInlineCompletionItems(cachedCompletions.logId, cachedCompletions.completions)
            }
            return []
        }

        const cachedCompletions = inlineCompletionsCache.get(prefix)
        if (cachedCompletions) {
            return toInlineCompletionItems(cachedCompletions.logId, cachedCompletions.completions)
        }

        const remainingChars = this.tokToChar(this.promptTokens)

        const completionNoSnippets = new InlineCompletionProvider(
            this.completionsClient,
            remainingChars,
            this.responseTokens,
            [],
            prefix,
            suffix,
            '\n',
            document.languageId
        )
        const emptyPromptLength = completionNoSnippets.emptyPromptLength()

        const contextChars = this.tokToChar(this.promptTokens) - emptyPromptLength

        const similarCode = await getContext({
            currentEditor,
            history: this.history,
            targetText: lastNLines(prefix, WINDOW_SIZE),
            windowSize: WINDOW_SIZE,
            maxChars: contextChars,
            codebaseContext: this.codebaseContext,
        })

        const completers: CompletionProvider[] = []
        let timeout: number
        let multilineMode: null | 'block' = null
        // VS Code does not show completions if we are in the process of writing a word or if a
        // selected completion info is present (so something is selected from the completions
        // dropdown list based on the lang server) and the returned completion range does not
        // contain the same selection.
        if (context.selectedCompletionInfo || /[A-Za-z]$/.test(sameLinePrefix)) {
            return []
        }
        // If we have a suffix in the same line as the cursor and the suffix contains any word
        // characters, do not attempt to make a completion. This means we only make completions if
        // we have a suffix in the same line for special characters like `)]}` etc.
        //
        // VS Code will attempt to merge the remainder of the current line by characters but for
        // words this will easily get very confusing.
        if (/\w/.test(sameLineSuffix)) {
            return []
        }
        // In this case, VS Code won't be showing suggestions anyway and we are more likely to want
        // suggested method names from the language server instead.
        if (context.triggerKind === vscode.InlineCompletionTriggerKind.Invoke || sameLinePrefix.endsWith('.')) {
            return []
        }

        if (
            (multilineMode = detectMultilineMode(
                prefix,
                prevNonEmptyLine,
                sameLinePrefix,
                sameLineSuffix,
                document.languageId
            ))
        ) {
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
                    document.languageId,
                    3,
                    multilineMode
                )
            )
        } else if (sameLinePrefix.trim() === '') {
            // The current line is empty
            timeout = 20
            completers.push(
                new InlineCompletionProvider(
                    this.completionsClient,
                    remainingChars,
                    this.responseTokens,
                    similarCode,
                    prefix,
                    suffix,
                    '',
                    document.languageId,
                    3 // tries
                )
            )
        } else {
            // The current line has a suffix
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
                    document.languageId,
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
                    document.languageId,
                    1 // tries
                )
            )
        }

        if (!this.disableTimeouts) {
            await new Promise<void>(resolve => setTimeout(resolve, timeout))
        }

        // We don't need to make a request at all if the signal is already aborted after the
        // debounce
        if (abortController.signal.aborted) {
            return []
        }

        const logId = CompletionLogger.start({ type: 'inline', multilineMode })

        const results = rankCompletions(
            (await Promise.all(completers.map(c => c.generateCompletions(abortController.signal)))).flat()
        )

        const visibleResults = filterCompletions(results)

        if (visibleResults.length > 0) {
            CompletionLogger.suggest(logId)
            inlineCompletionsCache.add(logId, visibleResults)
            return toInlineCompletionItems(logId, visibleResults)
        }

        CompletionLogger.noResponse(logId)
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
            '',
            currentEditor.document.languageId
        )
        const emptyPromptLength = completionNoSnippets.emptyPromptLength()

        const contextChars = this.tokToChar(this.promptTokens) - emptyPromptLength

        const similarCode = await getContext({
            currentEditor,
            history: this.history,
            targetText: lastNLines(prefix, WINDOW_SIZE),
            windowSize: WINDOW_SIZE,
            maxChars: contextChars,
            codebaseContext: this.codebaseContext,
        })

        const completer = new ManualCompletionProvider(
            this.completionsClient,
            remainingChars,
            this.responseTokens,
            similarCode,
            prefix,
            suffix,
            '',
            currentEditor.document.languageId
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

/**
 * Get the current document context based on the cursor position in the current document.
 *
 * This function is meant to provide a context around the current position in the document,
 * including a prefix, a suffix, the previous line, the previous non-empty line, and the next non-empty line.
 * The prefix and suffix are obtained by looking around the current position up to a max length
 * defined by `maxPrefixLength` and `maxSuffixLength` respectively. If the length of the entire
 * document content in either direction is smaller than these parameters, the entire content will be used.
 *
 * @param document - A `vscode.TextDocument` object, the document in which to find the context.
 * @param position - A `vscode.Position` object, the position in the document from which to find the context.
 * @param maxPrefixLength - A number representing the maximum length of the prefix to get from the document.
 * @param maxSuffixLength - A number representing the maximum length of the suffix to get from the document.
 *
 * @returns An object containing the current document context or null if there are no lines in the document.
 */
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

function toInlineCompletionItems(logId: string, completions: Completion[]): vscode.InlineCompletionItem[] {
    return completions.map(
        completion =>
            new vscode.InlineCompletionItem(completion.content, undefined, {
                title: 'Completion accepted',
                command: 'cody.completions.inline.accepted',
                arguments: [{ codyLogId: logId }],
            })
    )
}

function rankCompletions(completions: Completion[]): Completion[] {
    // TODO(philipp-spiess): Improve ranking to something more complex then just length
    return completions.sort((a, b) => b.content.split('\n').length - a.content.split('\n').length)
}

function filterCompletions(completions: Completion[]): Completion[] {
    return completions.filter(c => c.content.trim() !== '')
}
