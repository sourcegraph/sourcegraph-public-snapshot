import { CodebaseContext } from '../../codebase-context'
import { ContextMessage, getContextMessageWithResponse } from '../../codebase-context/messages'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { CHARS_PER_TOKEN, MAX_AVAILABLE_PROMPT_LENGTH, MAX_CURRENT_FILE_TOKENS } from '../../prompt/constants'
import { populateCurrentEditorContextTemplate } from '../../prompt/templates'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'

export class NextQuestions implements Recipe {
    public id = 'next-questions'

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const promptPrefix = 'Assume I have an answer to the following request:'
        const promptSuffix =
            'Generate one to three follow-up discussion topics that the human can ask you to uphold the conversation. Keep the topics very concise (try not to exceed 5 words per topic) and phrase them as questions.'

        const maxTokenCount =
            MAX_AVAILABLE_PROMPT_LENGTH - (promptPrefix.length + promptSuffix.length) / CHARS_PER_TOKEN
        const truncatedText = truncateText(humanChatInput, maxTokenCount)
        const promptMessage = `${promptPrefix}\n\n\`\`\`\n${truncatedText}\n\`\`\`\n\n${promptSuffix}`

        const assistantResponsePrefix = 'Sure, here are great follow-up discussion topics and learning ideas:\n\n - '
        return Promise.resolve(
            new Interaction(
                { speaker: 'human', text: promptMessage },
                {
                    speaker: 'assistant',
                    prefix: assistantResponsePrefix,
                    text: assistantResponsePrefix,
                },
                this.getContextMessages(promptMessage, context.editor, context.intentDetector, context.codebaseContext)
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
