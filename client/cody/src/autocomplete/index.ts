import vscode from 'vscode'

import { Completion, getAutocompleteContext } from '@sourcegraph/cody-shared/src/autocomplete'
import { CompletionsCache } from '@sourcegraph/cody-shared/src/autocomplete/cache'
import { InlineCompletionProvider } from '@sourcegraph/cody-shared/src/autocomplete/inline'
import { ProviderConfig } from '@sourcegraph/cody-shared/src/autocomplete/providers/provider'
import { isAbortError } from '@sourcegraph/cody-shared/src/autocomplete/utils'
import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { DocumentOffsets } from '@sourcegraph/cody-shared/src/editor/offsets'

import { debug } from '../log'
import { CodyStatusBar } from '../services/StatusBar'

import { VSCodeHistory } from './history'

interface CodyCompletionItemProviderConfig {
    providerConfig: ProviderConfig
    editor: Editor
    history: VSCodeHistory
    statusBar: CodyStatusBar
    codebaseContext: CodebaseContext
    responsePercentage?: number
    prefixPercentage?: number
    suffixPercentage?: number
    disableTimeouts?: boolean
    isCompletionsCacheEnabled?: boolean
    isEmbeddingsContextEnabled?: boolean
}

export class CodyCompletionItemProvider
    extends InlineCompletionProvider
    implements vscode.InlineCompletionItemProvider
{
    constructor(config: CodyCompletionItemProviderConfig) {
        const {
            providerConfig,
            editor,
            history,
            statusBar,
            codebaseContext,
            responsePercentage = 0.1,
            prefixPercentage = 0.6,
            suffixPercentage = 0.1,
            disableTimeouts = false,
            isEmbeddingsContextEnabled = true,
            isCompletionsCacheEnabled = true,
        } = config

        super(
            providerConfig,
            editor,
            history,
            codebaseContext,
            responsePercentage,
            prefixPercentage,
            suffixPercentage,
            disableTimeouts,
            isEmbeddingsContextEnabled,
            isCompletionsCacheEnabled ? new CompletionsCache() : undefined
        )

        this.startLoading = (label: string) => statusBar.startLoading(label)

        debug('CodyCompletionProvider:initialized', `provider: ${providerConfig.identifier}`)

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
        // Making it optional here to execute multiple suggestion in parallel from the CLI script.
        token?: vscode.CancellationToken
    ): Promise<vscode.InlineCompletionItem[]> {
        const abortController = new AbortController()

        if (token) {
            this.abortOpenInlineCompletions()
            token.onCancellationRequested(() => abortController.abort())
            this.abortOpenInlineCompletions = () => abortController.abort()
        }

        // CompletionLogger.clear()

        const currentEditor = vscode.window.activeTextEditor
        if (!currentEditor || currentEditor?.document.uri.scheme === 'cody') {
            return []
        }

        const doc = (await this.textEditor.getTextDocument(document.uri.toString()))!

        const docContext = getAutocompleteContext(doc, position, this.maxPrefixChars, this.maxSuffixChars)

        const offset = new DocumentOffsets(doc.content)

        const suffix = offset.jointRangeSlice(docContext.suffix)
        const prevLine = docContext.prevLine ? offset.jointRangeSlice(docContext.prevLine) : null

        // If we have a suffix in the same line as the cursor and the suffix contains any word
        // characters, do not attempt to make a completion. This means we only make completions if
        // we have a suffix in the same line for special characters like `)]}` etc.
        //
        // VS Code will attempt to merge the remainder of the current line by characters but for
        // words this will easily get very confusing.
        if (/\w/.test(suffix.slice(0, suffix.indexOf('\n')))) {
            return []
        }

        // VS Code does not show completions if we are in the process of writing a word or if a
        // selected completion info is present (so something is selected from the completions
        // dropdown list based on the lang server) and the returned completion range does not
        // contain the same selection.
        if (context.selectedCompletionInfo || !prevLine || /[A-Za-z]$/.test(prevLine)) {
            return []
        }

        // In this case, VS Code won't be showing suggestions anyway and we are more likely to want
        // suggested method names from the language server instead.
        if (context.triggerKind === vscode.InlineCompletionTriggerKind.Invoke || prevLine.endsWith('.')) {
            return []
        }

        try {
            const result = await this.getCompletions(
                abortController,
                {
                    uri: document.uri.toString(),
                    languageId: document.languageId,
                },
                docContext
            )
            if (!result) {
                return []
            }

            return toInlineCompletionItems(result.logId, result.completions)
        } catch (error) {
            this.stopLoading()

            if (isAbortError(error)) {
                return []
            }

            console.error(error)
            debug('CodyCompletionProvider:inline:error', `${error.toString()}\n${error.stack}`)
            return []
        }
    }
}

function toInlineCompletionItems(logId: string, completions: Completion[]): vscode.InlineCompletionItem[] {
    return completions.map(
        completion =>
            new vscode.InlineCompletionItem(completion.content, undefined, {
                title: 'Completion accepted',
                command: 'cody.autocomplete.inline.accepted',
                arguments: [{ codyLogId: logId, codyLines: completion.content.split('\n').length }],
            })
    )
}
