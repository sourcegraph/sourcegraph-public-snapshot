import { getShortTimestamp } from '@sourcegraph/cody-shared/src/timestamp'

import { CodebaseContext } from '../../codebase-context'
import { ContextMessage, getContextMessageWithResponse } from '../../codebase-context/messages'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { MAX_CURRENT_FILE_TOKENS, MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { populateCodeContextTemplate } from '../../prompt/templates'
import { truncateText } from '../../prompt/truncation'
import { renderMarkdown } from '../markdown'
import { Interaction } from '../transcript/interaction'

import { Recipe } from './recipe'

export class ChatQuestion implements Recipe {
    public getID(): string {
        return 'chat-question'
    }

    public async getInteraction(
        humanChatInput: string,
        editor: Editor,
        intentDetector: IntentDetector,
        codebaseContext: CodebaseContext
    ): Promise<Interaction | null> {
        const timestamp = getShortTimestamp()
        const truncatedText = truncateText(humanChatInput, MAX_HUMAN_INPUT_TOKENS)
        const displayText = renderMarkdown(humanChatInput)

        return Promise.resolve(
            new Interaction(
                { speaker: 'human', text: truncatedText, displayText, timestamp },
                { speaker: 'assistant', text: '', displayText: '', timestamp },
                this.getContextMessages(truncatedText, editor, intentDetector, codebaseContext)
            )
        )
    }

    private async getContextMessages(
        text: string,
        editor: Editor,
        intentDetector: IntentDetector,
        codebaseContext: CodebaseContext
    ): Promise<ContextMessage[]> {
        const isCodebaseContextRequired = await intentDetector.isCodebaseContextRequired(text)
        if (!isCodebaseContextRequired) {
            return []
        }

        const editorContextMessages = this.getEditorContext(editor)
        const codebaseContextMessages = await codebaseContext.getContextMessages(text, {
            numCodeResults: 8,
            numTextResults: 2,
        })

        return editorContextMessages.concat(codebaseContextMessages)
    }

    private getEditorContext(editor: Editor): ContextMessage[] {
        const visibleContent = editor.getActiveTextEditorVisibleContent()
        if (!visibleContent) {
            return []
        }
        const truncatedContent = truncateText(visibleContent.content, MAX_CURRENT_FILE_TOKENS)
        return getContextMessageWithResponse(
            populateCodeContextTemplate(truncatedContent, visibleContent.fileName),
            visibleContent.fileName
        )
    }
}
