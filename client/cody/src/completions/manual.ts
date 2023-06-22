import vscode from 'vscode'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { Completion } from '@sourcegraph/cody-shared/src/completions'
import { ReferenceSnippet, getContext } from '@sourcegraph/cody-shared/src/completions/context'
import { logCompletionEvent } from '@sourcegraph/cody-shared/src/completions/logger'
import {
    batchCompletions,
    sliceUntilFirstNLinesOfSuffixMatch,
    SNIPPET_WINDOW_SIZE,
    messagesToText,
} from '@sourcegraph/cody-shared/src/completions/utils'
import { SourcegraphNodeCompletionsClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/nodeClient'
import { Message } from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

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

        const completionNoSnippets = new LegacyManualCompletionProvider(
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

        const completer = new LegacyManualCompletionProvider(
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

export class LegacyManualCompletionProvider {
    constructor(
        protected completionsClient: SourcegraphNodeCompletionsClient,
        protected promptChars: number,
        protected responseTokens: number,
        protected snippets: ReferenceSnippet[],
        protected prefix: string,
        protected suffix: string,
        protected injectPrefix: string,
        protected languageId: string,
        protected defaultN: number = 1
    ) {}

    public emptyPromptLength(): number {
        const promptNoSnippets = messagesToText(this.createPromptPrefix())
        return promptNoSnippets.length - 10 // extra 10 chars of buffer cuz who knows
    }

    // Creates the resulting prompt and adds as many snippets from the reference
    // list as possible.
    protected createPrompt(): Message[] {
        const prefixMessages = this.createPromptPrefix()
        const referenceSnippetMessages: Message[] = []

        let remainingChars = this.promptChars - this.emptyPromptLength()

        if (this.suffix.length > 0) {
            let suffix = ''
            // We throw away the first 5 lines of the suffix to avoid the LLM to
            // just continue the completion by appending the suffix.
            const suffixLines = this.suffix.split('\n')
            if (suffixLines.length > 5) {
                suffix = suffixLines.slice(5).join('\n')
            }

            if (suffix.length > 0) {
                const suffixContext: Message[] = [
                    {
                        speaker: 'human',
                        text:
                            'Add the following code snippet to your knowledge base:\n' +
                            '```' +
                            `\n${suffix}\n` +
                            '```',
                    },
                    {
                        speaker: 'assistant',
                        text: '```\n// Ok```',
                    },
                ]

                const numSnippetChars = messagesToText(suffixContext).length + 1
                if (numSnippetChars <= remainingChars) {
                    referenceSnippetMessages.push(...suffixContext)
                    remainingChars -= numSnippetChars
                }
            }
        }

        for (const snippet of this.snippets) {
            const snippetMessages: Message[] = [
                {
                    speaker: 'human',
                    text:
                        `Add the following code snippet (from file ${snippet.fileName}) to your knowledge base:\n` +
                        '```' +
                        `\n${snippet.content}\n` +
                        '```',
                },
                {
                    speaker: 'assistant',
                    text: 'Okay, I have added it to my knowledge base.',
                },
            ]
            const numSnippetChars = messagesToText(snippetMessages).length + 1
            if (numSnippetChars > remainingChars) {
                break
            }
            referenceSnippetMessages.push(...snippetMessages)
            remainingChars -= numSnippetChars
        }

        return [...referenceSnippetMessages, ...prefixMessages]
    }

    protected createPromptPrefix(): Message[] {
        // TODO(beyang): escape 'Human:' and 'Assistant:'
        const prefix = this.prefix.trim()

        const prefixLines = prefix.split('\n')
        if (prefixLines.length === 0) {
            throw new Error('no prefix lines')
        }

        let prefixMessages: Message[]
        if (prefixLines.length > 2) {
            const endLine = Math.max(Math.floor(prefixLines.length / 2), prefixLines.length - 5)
            prefixMessages = [
                {
                    speaker: 'human',
                    text:
                        'Complete the following file:\n' +
                        '```' +
                        `\n${prefixLines.slice(0, endLine).join('\n')}\n` +
                        '```',
                },
                {
                    speaker: 'assistant',
                    text: `Here is the completion of the file:\n\`\`\`\n${prefixLines.slice(endLine).join('\n')}`,
                },
            ]
        } else {
            prefixMessages = [
                {
                    speaker: 'human',
                    text: 'Write some code',
                },
                {
                    speaker: 'assistant',
                    text: `Here is some code:\n\`\`\`\n${prefix}`,
                },
            ]
        }

        return prefixMessages
    }

    private postProcess(completion: string): string {
        let suggestion = completion
        const endBlockIndex = completion.indexOf('```')
        if (endBlockIndex !== -1) {
            suggestion = completion.slice(0, endBlockIndex)
        }

        // Remove trailing whitespace before newlines
        suggestion = suggestion
            .split('\n')
            .map(line => line.trimEnd())
            .join('\n')

        return sliceUntilFirstNLinesOfSuffixMatch(suggestion, this.suffix, 5)
    }

    public async generateCompletions(abortSignal: AbortSignal, n?: number): Promise<Completion[]> {
        const prefix = this.prefix.trim()

        // Create prompt
        const prompt = this.createPrompt()
        const textPrompt = messagesToText(prompt)
        if (textPrompt.length > this.promptChars) {
            throw new Error('prompt length exceeded maximum alloted chars')
        }

        // Issue request
        const responses = await batchCompletions(
            this.completionsClient,
            {
                messages: prompt,
                maxTokensToSample: this.responseTokens,
            },
            // We over-fetch the number of completions to account for potential
            // empty results
            (n || this.defaultN) + 2,
            abortSignal
        )
        // Post-process
        return responses
            .flatMap(resp => {
                const completion = this.postProcess(resp.completion)
                if (completion.trim() === '') {
                    return []
                }

                return [
                    {
                        prefix,
                        messages: prompt,
                        content: this.postProcess(resp.completion),
                        stopReason: resp.stopReason,
                    },
                ]
            })
            .slice(0, 3)
    }
}
