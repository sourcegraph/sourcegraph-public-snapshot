import vscode from 'vscode'

import { getContext } from '@sourcegraph/cody-shared/src/autocomplete/context'
import { logCompletionEvent } from '@sourcegraph/cody-shared/src/autocomplete/logger'
import { ManualCompletionProvider } from '@sourcegraph/cody-shared/src/autocomplete/manual'
import { SNIPPET_WINDOW_SIZE } from '@sourcegraph/cody-shared/src/autocomplete/utils'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { CompletionsDocumentProvider } from './docprovider'
import { getCurrentDocContext } from './document'
import { VSCodeHistory } from './history'
import { textEditor } from './text_editor'

const LOG_MANUAL = { type: 'manual' }

export class ManualCompletionService {
    private promptTokens: number
    private maxPrefixTokens: number
    private maxSuffixTokens: number
    private abortOpenManualCompletion: () => void = () => {}

    constructor(
        private webviewErrorMessenger: (error: string) => Promise<void>,
        private completionsClient: SourcegraphNodeCompletionsClient,
        private documentProvider: CompletionsDocumentProvider,
        private history: VSCodeHistory,
        private codebaseContext: CodebaseContext,
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

    private tokToChar(tokens: number): number {
        return tokens * this.charsPerToken
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

        const { context: similarCode } = await getContext({
            currentEditor: textEditor,
            prefix,
            suffix,
            history: this.history,
            jaccardDistanceWindowSize: SNIPPET_WINDOW_SIZE,
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
            logCompletionEvent('started', LOG_MANUAL)
            const completions = await completer.generateCompletions(abortController.signal, 3)
            this.documentProvider.addCompletions(completionsUri, ext, completions, {
                suffix: '',
                elapsedMillis: 0,
                llmOptions: null,
            })
            logCompletionEvent('suggested', LOG_MANUAL)
        } catch (error) {
            if (error.message === 'aborted') {
                return
            }

            await this.webviewErrorMessenger(`FetchAndShowCompletions - ${error}`)
        }
    }
}
