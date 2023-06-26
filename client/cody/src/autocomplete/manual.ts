import vscode from 'vscode'

import { Completion, CurrentDocumentContextWithLanguage, History } from '@sourcegraph/cody-shared/src/autocomplete'
import { ManualCompletionService } from '@sourcegraph/cody-shared/src/autocomplete/manual'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'

import { CompletionsDocumentProvider } from './docprovider'
import { getCurrentDocContext } from './document'
import { textEditor } from './text_editor'

const COMPLETIONS_URI = vscode.Uri.parse('cody:Completions.md')

export class ManualCompletionServiceVSCode extends ManualCompletionService {
    constructor(
        webviewErrorMessenger: (error: string) => Promise<void>,
        completionsClient: SourcegraphNodeCompletionsClient,
        private documentProvider: CompletionsDocumentProvider,
        history: History,
        codebaseContext: CodebaseContext
    ) {
        super(textEditor, webviewErrorMessenger, completionsClient, history, codebaseContext)
    }

    async getCurrentDocumentContext(): Promise<CurrentDocumentContextWithLanguage | null> {
        const currentEditor = vscode.window.activeTextEditor
        if (!currentEditor || currentEditor?.document.uri.scheme === 'cody') {
            return null
        }

        const filename = currentEditor.document.fileName
        const ext = filename.split('.').pop() || ''
        this.documentProvider.clearCompletions(COMPLETIONS_URI)

        const doc = await vscode.workspace.openTextDocument(COMPLETIONS_URI)
        await vscode.window.showTextDocument(doc, {
            preview: false,
            viewColumn: 2,
        })

        const ctx = getCurrentDocContext(
            currentEditor.document,
            currentEditor.selection.start,
            this.tokToChar(this.maxPrefixTokens),
            this.tokToChar(this.maxSuffixTokens)
        )

        if (!ctx) {
            return null
        }

        // TODO(beyang): make getCurrentDocContext fetch complete line prefix
        return {
            ...ctx,
            languageId: currentEditor.document.languageId,
            markdownLanguage: ext,
        }
    }

    emitCompletions(docContext: CurrentDocumentContextWithLanguage, completions: Completion[]): void {
        this.documentProvider.addCompletions(COMPLETIONS_URI, docContext.markdownLanguage, completions, {
            suffix: '',
            elapsedMillis: 0,
            llmOptions: null,
        })
    }
}
