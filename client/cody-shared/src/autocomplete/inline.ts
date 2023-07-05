import { LRUCache } from 'lru-cache'

import { CodebaseContext } from '../codebase-context'
import { Editor, History, LightTextDocument } from '../editor'
import { DocumentOffsets } from '../editor/offsets'

import { AutocompleteContext, Completion } from '.'
import { CompletionsCache } from './cache'
import { getContext } from './context'
import { detectMultilineMode } from './multiline'
import { Provider, ProviderConfig } from './providers/provider'
import { sharedPostProcess } from './shared-post-process'
import { SNIPPET_WINDOW_SIZE } from './utils'

export interface InlineCompletionResult {
    logId: string
    completions: Completion[]
}

export class InlineCompletionProvider {
    public promptChars: number
    public maxPrefixChars: number
    public maxSuffixChars: number
    public abortOpenInlineCompletions: () => void = () => {}
    public startLoading: (text: string) => () => void = () => () => {}
    public stopLoading: () => void = () => {}
    public lastContentChanges: LRUCache<string, 'add' | 'del'> = new LRUCache<string, 'add' | 'del'>({
        max: 10,
    })

    constructor(
        public providerConfig: ProviderConfig,
        public textEditor: Editor,
        public history: History,
        public codebaseContext: CodebaseContext,
        public responsePercentage: number,
        public prefixPercentage: number,
        public suffixPercentage: number,
        public disableTimeouts: boolean,
        public isEmbeddingsContextEnabled: boolean,
        public inlineCompletionsCache?: CompletionsCache
    ) {
        this.promptChars =
            providerConfig.maximumContextCharacters - providerConfig.maximumContextCharacters * responsePercentage
        this.maxPrefixChars = Math.floor(this.promptChars * this.prefixPercentage)
        this.maxSuffixChars = Math.floor(this.promptChars * this.suffixPercentage)
    }

    // document: vscode.TextDocument,
    // position: vscode.Position,
    // context: vscode.InlineCompletionContext,
    // token?: vscode.CancellationToken
    public async getCompletions(
        abortController: AbortController,
        document: LightTextDocument,
        docContext: AutocompleteContext
    ): Promise<InlineCompletionResult | null> {
        const offset = new DocumentOffsets(docContext.content)

        const languageId = document.languageId
        const {
            prefix: prefixRange,
            suffix: suffixRange,
            prevLine: sameLinePrefixRange,
            prevNonEmptyLine: prevNonEmptyLineRange,
        } = docContext

        const prefix = offset.jointRangeSlice(prefixRange)
        const suffix = offset.jointRangeSlice(suffixRange)
        const sameLinePrefix = sameLinePrefixRange ? offset.jointRangeSlice(sameLinePrefixRange) : ''
        const prevNonEmptyLine = prevNonEmptyLineRange ? offset.jointRangeSlice(prevNonEmptyLineRange) : ''

        const sameLineSuffix = suffix.slice(0, suffix.indexOf('\n'))

        // Avoid showing completions when we're deleting code (Cody can only insert code at the
        // moment)
        const lastChange = this.lastContentChanges.get(document.uri) ?? 'add'
        if (lastChange === 'del') {
            // When a line was deleted, only look up cached items and only include them if the
            // untruncated prefix matches. This fixes some weird issues where the completion would
            // render if you insert whitespace but not on the original place when you delete it
            // again
            const cachedCompletions = this.inlineCompletionsCache?.get(prefix, false)
            if (cachedCompletions?.isExactPrefix) {
                return { logId: cachedCompletions.logId, completions: cachedCompletions.completions }
            }
            return null
        }

        const cachedCompletions = this.inlineCompletionsCache?.get(prefix)
        if (cachedCompletions) {
            return { logId: cachedCompletions.logId, completions: cachedCompletions.completions }
        }

        const completers: Provider[] = []
        let timeout: number

        const workspace = this.textEditor.getActiveWorkspace()

        if (!workspace) {
            return null
        }

        const sharedProviderOptions = {
            prefix,
            suffix,
            fileName: workspace.relativeTo(document.uri)!,
            languageId,
            responsePercentage: this.responsePercentage,
            prefixPercentage: this.prefixPercentage,
            suffixPercentage: this.suffixPercentage,
        }

        const multilineMode = detectMultilineMode(
            this.textEditor,
            prefix,
            prevNonEmptyLine,
            sameLinePrefix,
            sameLineSuffix,
            languageId,
            this.providerConfig.enableExtendedMultilineTriggers
        )
        if (multilineMode === 'block') {
            timeout = 100
            completers.push(
                this.providerConfig.create({
                    ...sharedProviderOptions,
                    n: 3,
                    multilineMode,
                })
            )
        } else {
            // The current line has a suffix
            timeout = 20
            completers.push(
                this.providerConfig.create({
                    ...sharedProviderOptions,
                    n: 3,
                    multilineMode: null,
                })
            )
        }

        if (!this.disableTimeouts) {
            await new Promise<void>(resolve => setTimeout(resolve, timeout))
        }

        // We don't need to make a request at all if the signal is already aborted after the
        // debounce
        if (abortController.signal.aborted) {
            return null
        }

        const { context: similarCode } = await getContext({
            currentEditor: this.textEditor,
            prefix,
            suffix,
            history: this.history,
            jaccardDistanceWindowSize: SNIPPET_WINDOW_SIZE,
            maxChars: this.promptChars,
            codebaseContext: this.codebaseContext,
            isEmbeddingsContextEnabled: this.isEmbeddingsContextEnabled,
        })
        if (abortController.signal.aborted) {
            return null
        }

        // const logId = CompletionLogger.start({
        //     type: 'inline',
        //     multilineMode,
        //     providerIdentifier: this.providerConfig.identifier,
        //     languageId,
        //     contextSummary,
        // })
        const logId = 'TODO'
        const stopLoading = this.startLoading('Completions are being generated')
        this.stopLoading = stopLoading

        // Overwrite the abort handler to also update the loading state
        const previousAbort = this.abortOpenInlineCompletions
        this.abortOpenInlineCompletions = () => {
            previousAbort()
            stopLoading()
        }

        const completions = (
            await Promise.all(completers.map(c => c.generateCompletions(abortController.signal, similarCode)))
        ).flat()

        // Shared post-processing logic
        const processedCompletions = completions.map(completion =>
            sharedPostProcess({
                textEditor: this.textEditor,
                prefix,
                suffix,
                multiline: multilineMode !== null,
                languageId,
                completion,
            })
        )

        // Filter results
        const visibleResults = filterCompletions(processedCompletions)

        // Rank results
        const rankedResults = rankCompletions(visibleResults)

        stopLoading()

        if (rankedResults.length > 0) {
            // CompletionLogger.suggest(logId)
            this.inlineCompletionsCache?.add(logId, rankedResults)
            return { logId, completions: rankedResults }
        }

        // CompletionLogger.noResponse(logId)
        return null
    }
}

function rankCompletions(completions: Completion[]): Completion[] {
    // TODO(philipp-spiess): Improve ranking to something more complex then just length
    return completions.sort((a, b) => b.content.split('\n').length - a.content.split('\n').length)
}

function filterCompletions(completions: Completion[]): Completion[] {
    return completions.filter(c => c.content.trim() !== '')
}
