import { Editor } from '../../editor'
import { Message } from '../../sourcegraph-api'
import { ContextSearchOptions } from '../context-search-options'
import { getContextMessageWithResponse, populateCodeContextTemplate, truncateText, truncateTextStart } from '../prompt'

import { getFileExtension } from './generateTest'
import { MARKDOWN_FORMAT_PROMPT, getNormalizedLanguageName } from './helpers'
import { Recipe, RecipePrompt } from './recipe'

export class ImproveVariableNames implements Recipe {
    public getID(): string {
        return 'improveVariableNames'
    }
    public async getPrompt(
        maxTokens: number,
        editor: Editor,
        getEmbeddingsContextMessages: (query: string, options: ContextSearchOptions) => Promise<Message[]>
    ): Promise<RecipePrompt | null> {
        const maxInputTokens = Math.round(0.8 * maxTokens)
        const maxSurroundingTokens = Math.round(0.2 * maxTokens)

        // Inputs
        const selection = editor.getActiveTextEditorSelection()
        if (!selection) {
            return null
        }

        // Display text
        const displayText = `Improve the variable names in the following code:\n\`\`\`\n${selection.selectedText}\n\`\`\``

        // Context messages
        const contextQuery = truncateText(selection.selectedText, maxInputTokens)
        const contextMessages = await getEmbeddingsContextMessages(contextQuery, {
            numCodeResults: 8,
            numTextResults: 0,
        })
        contextMessages.push(
            ...[
                truncateTextStart(selection.precedingText, maxSurroundingTokens),
                truncateText(selection.followingText, maxSurroundingTokens),
            ].flatMap(text => getContextMessageWithResponse(populateCodeContextTemplate(text, selection.fileName)))
        )

        // Get query message
        const languageName = getNormalizedLanguageName(selection.fileName)
        const promptMessage: Message = {
            speaker: 'human',
            text: `Improve the variable names in this ${languageName} code by replacing the variable names with new identifiers which succinctly capture the purpose of the variable. We want the new code to be a drop-in replacement, so do not change names bound outside the scope of this code, like function names or members defined elsewhere. Only change the names of local variables and parameters:\n\n\`\`\`\n${selection.selectedText}\n\`\`\`\n${MARKDOWN_FORMAT_PROMPT}`,
        }

        // Response prefix
        const extension = getFileExtension(selection.fileName)
        const botResponsePrefix = `Here is the improved code:\n\`\`\`${extension}\n`

        return {
            displayText,
            promptMessage,
            botResponsePrefix,
            contextMessages,
        }
    }
}
