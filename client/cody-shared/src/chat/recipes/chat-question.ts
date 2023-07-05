import { Uri } from 'vscode'

import { CodebaseContext } from '../../codebase-context'
import { ContextMessage, getContextMessageWithResponse } from '../../codebase-context/messages'
import { SelectionText, Editor, TextDocument } from '../../editor'
import { DocumentOffsets } from '../../editor/offsets'
import { IntentDetector } from '../../intent-detector'
import { MAX_CURRENT_FILE_TOKENS, MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import {
    populateCurrentEditorContextTemplate,
    populateCurrentEditorSelectedContextTemplate,
} from '../../prompt/templates'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext, RecipeID } from './recipe'

export class ChatQuestion implements Recipe {
    public id: RecipeID = 'chat-question'

    constructor(private debug: (filterLabel: string, text: string, ...args: unknown[]) => void) {}

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const truncatedText = truncateText(humanChatInput, MAX_HUMAN_INPUT_TOKENS)

        const active = context.editor.getActiveTextDocument()

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
                    active ? Editor.getTextDocumentSelectionText(active) : null
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
        selection: SelectionText | null
    ): Promise<ContextMessage[]> {
        const contextMessages: ContextMessage[] = []

        const isCodebaseContextRequired = firstInteraction || (await intentDetector.isCodebaseContextRequired(text))

        this.debug('ChatQuestion:getContextMessages', 'isCodebaseContextRequired', isCodebaseContextRequired)
        if (isCodebaseContextRequired) {
            const codebaseContextMessages = await codebaseContext.getContextMessages(text, {
                numCodeResults: 12,
                numTextResults: 3,
            })
            contextMessages.push(...codebaseContextMessages)
        }

        const isEditorContextRequired = intentDetector.isEditorContextRequired(text)
        this.debug('ChatQuestion:getContextMessages', 'isEditorContextRequired', isEditorContextRequired)
        if (isCodebaseContextRequired || isEditorContextRequired) {
            contextMessages.push(...ChatQuestion.getEditorContext(editor))
        }

        // Add selected text as context when available
        if (selection?.selectedText) {
            contextMessages.push(...ChatQuestion.getEditorSelectionContext(editor.getActiveTextDocument()!, selection))
        }

        return contextMessages
    }

    public static getEditorContext(editor: Editor): ContextMessage[] {
        const currentDocument = editor.getActiveTextDocument()
        if (!currentDocument?.visible) {
            return []
        }

        const filePath = Uri.parse(currentDocument.uri).fsPath
        const offset = new DocumentOffsets(currentDocument.content)

        const truncatedContent = truncateText(offset.jointRangeSlice(currentDocument.visible), MAX_CURRENT_FILE_TOKENS)
        return getContextMessageWithResponse(
            populateCurrentEditorContextTemplate(truncatedContent, filePath, currentDocument.repoName ?? undefined),
            {
                fileName: filePath,
                repoName: currentDocument.repoName ?? undefined,
                revision: currentDocument.revision ?? undefined,
            }
        )
    }

    public static getEditorSelectionContext(document: TextDocument, selection: SelectionText): ContextMessage[] {
        const filePath = Uri.parse(document.uri).fsPath
        const truncatedContent = truncateText(selection.selectedText, MAX_CURRENT_FILE_TOKENS)

        return getContextMessageWithResponse(
            populateCurrentEditorSelectedContextTemplate(truncatedContent, filePath, document.repoName ?? undefined),
            {
                fileName: filePath,
                ...selection,
            }
        )
    }
}
