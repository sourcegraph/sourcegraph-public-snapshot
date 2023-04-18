import { CodebaseContext } from '../../codebase-context'
import { ContextMessage, getContextMessageWithResponse } from '../../codebase-context/messages'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { MAX_CURRENT_FILE_TOKENS, MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { populateCurrentEditorContextTemplate } from '../../prompt/templates'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'

export class ChatQuestion implements Recipe {
    public getID(): string {
        return 'chat-question'
    }

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const truncatedText = truncateText(humanChatInput, MAX_HUMAN_INPUT_TOKENS)

        return Promise.resolve(
            new Interaction(
                { speaker: 'human', text: truncatedText, displayText: humanChatInput },
                { speaker: 'assistant' },
                this.getContextMessages(truncatedText, context.editor, context.intentDetector, context.codebaseContext)
            )
        )
    }

    private async getContextMessages(
        text: string,
        editor: Editor,
        intentDetector: IntentDetector,
        codebaseContext: CodebaseContext
    ): Promise<ContextMessage[]> {
        const contextMessages: ContextMessage[] = []

        const isCodebaseContextRequired = await intentDetector.isCodebaseContextRequired(text)
        if (isCodebaseContextRequired) {
            const codebaseContextMessages = await codebaseContext.getContextMessages(text, {
                numCodeResults: 12,
                numTextResults: 3,
            })
            contextMessages.push(...codebaseContextMessages)
        }

        if (isCodebaseContextRequired || intentDetector.isEditorContextRequired(text)) {
            contextMessages.push(...this.getEditorContext(editor))
        }

        return contextMessages
    }

    private getEditorContext(editor: Editor): ContextMessage[] {
        const visibleContent = editor.getActiveTextEditorVisibleContent()
        if (!visibleContent) {
            return []
        }
        const truncatedContent = truncateText(visibleContent.content, MAX_CURRENT_FILE_TOKENS)
        return getContextMessageWithResponse(
            populateCurrentEditorContextTemplate(truncatedContent, visibleContent.fileName),
            visibleContent.fileName
        )
    }
}
