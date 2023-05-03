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
        const selection = context.editor.getActiveTextEditorSelection()
        if (!humanChatInput || !selection) {
            await context.editor.showWarningMessage('Failed to start file-chat.')
            return null
        }
        // Check if this is a fix-up request
        if (humanChatInput.startsWith('/fix ') || humanChatInput.startsWith('/f ')) {
            return new Fixup().getInteraction(humanChatInput, context)
        }

        const truncatedText = truncateText(humanChatInput, MAX_HUMAN_INPUT_TOKENS)
        const MAX_RECIPE_CONTENT_TOKENS = MAX_RECIPE_INPUT_TOKENS + MAX_RECIPE_SURROUNDING_TOKENS * 2
        const truncatedFile = truncateText(
            selection.precedingText + selection.selectedText + selection.followingText,
            MAX_RECIPE_CONTENT_TOKENS
        )
        const truncatedSelectedText = truncateText(selection.selectedText, MAX_RECIPE_CONTENT_TOKENS)

        // Prompt for Cody
        const promptText = FileChat.prompt
            .replace('{humanInput}', truncatedText)
            .replace('{selected}', truncatedSelectedText)
            .replace('{content}', truncatedFile)
            .replace('{fileName}', selection.fileName)

        // Text display in UI fpr human
        const displayText = humanChatInput + FileChat.displayPrompt.replace('{selectedText}', selection.selectedText)

        return Promise.resolve(
            new Interaction(
                {
                    speaker: 'human',
                    text: promptText,
                    displayText,
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

    // Prompt Templates
    public static readonly prompt = `
    Here's the code from file {filenName}:
    \`\`\`
    {content}
    \`\`\`

    I have questions about this part of the file:
    \`\`\`
    {selectedText}
    \`\`\`

    As my coding assistant, please help me with my questions:
    {humanInput}

    ## Instruction
    - Do not enclose your answer with tags.
    - Do not remove code that might be being used by the other part of the code that was not shared.
    - Your answers and suggestions should based on the provided context only.
    - You may make references to other part of the shared code.
    - Do not suggest code that are not related to any of the shared context.
    - Do not suggest anything that would break the working code.
    `

    public static readonly displayPrompt = `
    \nQuestions based on the code below:\n\`\`\`\n{selectedText}\n\`\`\`\n
    `
}
