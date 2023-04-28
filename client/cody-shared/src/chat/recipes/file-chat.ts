import { CodebaseContext } from '../../codebase-context'
import { ContextMessage } from '../../codebase-context/messages'
import { MAX_HUMAN_INPUT_TOKENS, MAX_RECIPE_INPUT_TOKENS, MAX_RECIPE_SURROUNDING_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { Fixup } from './fixup'
import { Recipe, RecipeContext } from './recipe'

export class FileChat implements Recipe {
    public id = 'file-chat'

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        if (humanChatInput.startsWith('/fix')) {
            return new Fixup().getInteraction(humanChatInput, context)
        }

        const truncatedText = truncateText(humanChatInput, MAX_HUMAN_INPUT_TOKENS)
        const selection = context.editor.getActiveTextEditorSelection()
        if (!truncatedText || !selection) {
            await context.editor.showWarningMessage('Failed to start chat.')
            return null
        }

        const MAX_RECIPE_CONTENT_TOKENS = MAX_RECIPE_INPUT_TOKENS + MAX_RECIPE_SURROUNDING_TOKENS * 2
        const truncatedFile = truncateText(selection.selectedText, MAX_RECIPE_CONTENT_TOKENS)

        const linePrompt = `I am current looking at this part of the file:\n\`\`\`\n${selection.selectedText}\n\`\`\`\n`
        const contentPrompt = `Here is the code from the file (path: ${selection.fileName}) I am working on:\n\`\`\`\n${truncatedFile}\n\`\`\`\n`
        const prompt =
            contentPrompt +
            linePrompt +
            `\nAnswer my (follow-up) questions based on the shared code: \n${truncatedText}\n`

        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: prompt,
                    displayText:
                        humanChatInput +
                        `\n\nQuestions based on the code below:\n\`\`\`\n${selection.selectedText}\n\`\`\`\n`,
                },
                { speaker: 'assistant' },
                this.getContextMessages(selection.selectedText, context.codebaseContext)
            )
        )
    }

    private async getContextMessages(text: string, codebaseContext: CodebaseContext): Promise<ContextMessage[]> {
        const contextMessages: ContextMessage[] = await codebaseContext.getContextMessages(text, {
            numCodeResults: 12,
            numTextResults: 3,
        })
        return contextMessages
    }
}
