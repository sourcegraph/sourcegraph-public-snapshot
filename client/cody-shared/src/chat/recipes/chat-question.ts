import { spawnSync } from 'child_process'

import * as vscode from 'vscode'

import { isErrorLike } from '@sourcegraph/common'

import { CodebaseContext } from '../../codebase-context'
import { ContextMessage, getContextMessageWithResponse } from '../../codebase-context/messages'
import { ActiveTextEditorSelection, Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { MAX_CURRENT_FILE_TOKENS, MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import {
    populateCurrentEditorContextTemplate,
    populateCurrentEditorSelectedContextTemplate,
} from '../../prompt/templates'
import { truncateText } from '../../prompt/truncation'
import { PreciseContextResult, SourcegraphGraphQLAPIClient } from '../../sourcegraph-api/graphql/client'
import { Interaction } from '../transcript/interaction'

import { convertGitCloneURLToCodebaseName } from './helpers'
import { Recipe, RecipeContext, RecipeID } from './recipe'

export class ChatQuestion implements Recipe {
    public id: RecipeID = 'chat-question'

    constructor(private debug: (filterLabel: string, text: string, ...args: unknown[]) => void) {}

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const truncatedText = truncateText(humanChatInput, MAX_HUMAN_INPUT_TOKENS)

        return Promise.resolve(
            new Interaction(
                { speaker: 'human', text: truncatedText, displayText: humanChatInput },
                { speaker: 'assistant' },
                this.getContextMessages(
                    truncatedText,
                    context.editor,
                    context.firstInteraction,
                    context.intentDetector,
                    context.codebaseContext,
                    context.editor.getActiveTextEditorSelection() || null
                ),
                []
            )
        )
    }

    private async getContextMessages(
        text: string,
        editor: Editor,
        firstInteraction: boolean,
        intentDetector: IntentDetector,
        codebaseContext: CodebaseContext,
        selection: ActiveTextEditorSelection | null
    ): Promise<ContextMessage[]> {
        const contextMessages: ContextMessage[] = []

        const fullConfig = {
            serverEndpoint: 'https://sourcegraph.test:3443',
            accessToken: '<CHANGE_THIS_TOKEN>',
            debug: false,
            customHeaders: {},
        }
        const graphqlClient = new SourcegraphGraphQLAPIClient(fullConfig)

        const activeFileContent = vscode.window.visibleTextEditors[0].document.getText()
        const filePath = vscode.window.visibleTextEditors[0].document.uri.fsPath

        let preciseContext: PreciseContextResult[] = []
        const workspaceRoot = editor.getWorkspaceRootPath()
        if (workspaceRoot) {
            // Get codebase from config or fallback to getting repository name from git clone URL
            const gitCommand = spawnSync('git', ['remote', 'get-url', 'origin'], { cwd: workspaceRoot })
            const gitOutput = gitCommand.stdout.toString().trim()
            const repository = convertGitCloneURLToCodebaseName(gitOutput) || ''
            const codebaseNameSplit = repository.split('/')
            const repoName = codebaseNameSplit.length ? codebaseNameSplit[codebaseNameSplit.length - 1] : ''
            const activeFile = trimPath(filePath, repoName)

            const gitOIDCommand = spawnSync('git', ['rev-parse', 'HEAD'], { cwd: workspaceRoot })
            const commitID = gitOIDCommand.stdout.toString().trim()

            const response = await graphqlClient.getPreciseContext(repository, commitID, activeFile, activeFileContent)
            if (!isErrorLike(response)) {
                preciseContext = response
            }
        }

        // Add selected text as context when available
        if (selection?.selectedText) {
            contextMessages.push(...ChatQuestion.getEditorSelectionContext(selection))
        }

        const isCodebaseContextRequired = firstInteraction || (await intentDetector.isCodebaseContextRequired(text))

        this.debug('ChatQuestion:getContextMessages', 'isCodebaseContextRequired', isCodebaseContextRequired)
        if (isCodebaseContextRequired) {
            const codebaseContextMessages = await codebaseContext.getContextMessages(text, {
                numCodeResults: 12,
                numTextResults: 3,
            })
            contextMessages.push(...codebaseContextMessages)

            for (const context of preciseContext) {
                contextMessages.push({
                    speaker: 'human',
                    file: {
                        fileName: '',
                        repoName: context.repository,
                    },
                    text: `Here is the code snippet: ${context.text}`,
                })
                contextMessages.push({ speaker: 'assistant', text: 'okay' })
            }

            // contextMessages.push({
            //     speaker: 'human',
            //     file: {
            //         fileName: 'filename',
            //         repoName: 'repoName',
            //         revision: 'revision',
            //     },
            //     text: `
            //         Here is the path to the file ${test.data.search.results.results[0].file.path}.
            //         The kind of the symbol is a ${test.data.search.results.results[0].symbols.kind}.
            //         The name of the symbol is a ${test.data.search.results.results[0].symbols.name}.
            //         It is located in ${test.data.search.results.results[0].symbols.url}
            //         This is the content of the file ${test.data.search.results.results[0].file.content}.
            //     `,
            // })
            // contextMessages.push({
            //     speaker: 'assistant',
            //     text: 'okay',
            // })
        }

        const isEditorContextRequired = intentDetector.isEditorContextRequired(text)
        this.debug('ChatQuestion:getContextMessages', 'isEditorContextRequired', isEditorContextRequired)
        if (isCodebaseContextRequired || isEditorContextRequired) {
            contextMessages.push(...ChatQuestion.getEditorContext(editor))
        }

        return contextMessages
    }

    public static getEditorContext(editor: Editor): ContextMessage[] {
        const visibleContent = editor.getActiveTextEditorVisibleContent()
        if (!visibleContent) {
            return []
        }
        const truncatedContent = truncateText(visibleContent.content, MAX_CURRENT_FILE_TOKENS)
        return getContextMessageWithResponse(
            populateCurrentEditorContextTemplate(truncatedContent, visibleContent.fileName),
            visibleContent
        )
    }

    public static getEditorSelectionContext(selection: ActiveTextEditorSelection): ContextMessage[] {
        const truncatedContent = truncateText(selection.selectedText, MAX_CURRENT_FILE_TOKENS)
        return getContextMessageWithResponse(
            populateCurrentEditorSelectedContextTemplate(truncatedContent, selection.fileName),
            selection
        )
    }
}

function trimPath(path: string, folderName: string) {
    const folderIndex = path.indexOf(folderName)

    if (folderIndex === -1) {
        return path
    }

    // Add folderName.length for length of folder name and +1 for the slash
    return path.slice(Math.max(0, folderIndex + folderName.length + 1))
}
