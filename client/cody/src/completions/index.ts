import { LRUCache } from 'lru-cache'
import * as vscode from 'vscode'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { debug } from '../log'
import { CodyStatusBar } from '../services/StatusBar'

import { CompletionsCache } from './cache'
import { getContext } from './context'
import { getCurrentDocContext } from './document'
import { History } from './history'
import * as CompletionLogger from './logger'
import { detectMultilineMode } from './multiline'
import { CompletionProvider, InlineCompletionProvider } from './provider'
import { SNIPPET_WINDOW_SIZE, lastNLines } from './utils'

export const inlineCompletionsCache = new CompletionsCache()

export class CodyCompletionItemProvider implements vscode.InlineCompletionItemProvider {
    private promptTokens: number
    private maxPrefixTokens: number
    private maxSuffixTokens: number
    private abortOpenInlineCompletions: () => void = () => {}
    private lastContentChanges: LRUCache<string, 'add' | 'del'> = new LRUCache<string, 'add' | 'del'>({
        max: 10,
    })

    constructor(
        private completionsClient: SourcegraphNodeCompletionsClient,
        private history: History,
        private statusBar: CodyStatusBar,
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
            targetText: lastNLines(prefix, SNIPPET_WINDOW_SIZE),
            windowSize: SNIPPET_WINDOW_SIZE,
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

        // We don't need to make a request at all if the signal is already aborted after the
        // debounce
        if (abortController.signal.aborted) {
            return []
        }

        const logId = CompletionLogger.start({ type: 'inline', multilineMode })
        const stopLoading = this.statusBar.startLoading('Completions are being generated')

        // Overwrite the abort handler to also update the loading state
        const previousAbort = this.abortOpenInlineCompletions
        this.abortOpenInlineCompletions = () => {
            previousAbort()
            stopLoading()
        }

        if (!this.disableTimeouts) {
            await new Promise<void>(resolve => setTimeout(resolve, timeout))
        }

        const results = rankCompletions(
            (await Promise.all(completers.map(c => c.generateCompletions(abortController.signal)))).flat()
        )

        const visibleResults = filterCompletions(results)

        stopLoading()

        if (visibleResults.length > 0) {
            CompletionLogger.suggest(logId)
            inlineCompletionsCache.add(logId, visibleResults)
            return toInlineCompletionItems(logId, visibleResults)
        }

        CompletionLogger.noResponse(logId)
        return []
    }
}

export interface Completion {
    prefix: string
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
