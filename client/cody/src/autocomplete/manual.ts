import vscode from 'vscode'

import { History } from '@sourcegraph/cody-shared/src/autocomplete'
import { logCompletionEvent } from '@sourcegraph/cody-shared/src/autocomplete/logger'
import { ManualCompletionService } from '@sourcegraph/cody-shared/src/autocomplete/manual'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { CompletionsDocumentProvider } from './docprovider'
import { getCurrentDocContext } from './document'
import { textEditor } from './text_editor'

const LOG_MANUAL = { type: 'manual' }
const COMPLETIONS_URI = vscode.Uri.parse('cody:Completions.md')

export class ManualCompletionServiceVSCode extends ManualCompletionService {
    private abortOpenManualCompletion: () => void = () => {}

    constructor(
        private webviewErrorMessenger: (error: string) => Promise<void>,
        completionsClient: SourcegraphNodeCompletionsClient,
        private documentProvider: CompletionsDocumentProvider,
        history: History,
        codebaseContext: CodebaseContext
    ) {
        super(textEditor, completionsClient, history, codebaseContext)
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
        this.documentProvider.clearCompletions(COMPLETIONS_URI)

        const doc = await vscode.workspace.openTextDocument(COMPLETIONS_URI)
        await vscode.window.showTextDocument(doc, {
            preview: false,
            viewColumn: 2,
        })

        const docContext = getCurrentDocContext(
            currentEditor.document,
            currentEditor.selection.start,
            this.tokToChar(this.maxPrefixTokens),
            this.tokToChar(this.maxSuffixTokens)
        )

        if (!docContext) {
            return
        }

        const completer = await this.getManualCompletionProvider({
            ...docContext,
            languageId: currentEditor.document.languageId,
            markdownLanguage: ext,
        })

        if (!completer) {
            return
        }

        try {
            logCompletionEvent('started', LOG_MANUAL)
            const completions = await completer.generateCompletions(abortController.signal, 3)
            this.documentProvider.addCompletions(COMPLETIONS_URI, ext, completions, {
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
