import { Editor } from '../../editor'
import { Message } from '../../sourcegraph-api'
import { ContextSearchOptions } from '../context-search-options'
import { getContextMessageWithResponse, populateCodeContextTemplate, truncateText, truncateTextStart } from '../prompt'

import { MARKDOWN_FORMAT_PROMPT, getNormalizedLanguageName } from './helpers'
import { Recipe, RecipePrompt } from './recipe'

export class ExplainCodeDetailed implements Recipe {
    public getID(): string {
        return 'explainCode'
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

        // Context messages
        const contextQuery = truncateText(selection.selectedText, maxInputTokens)
        const contextMessages = await getEmbeddingsContextMessages(contextQuery, {
            numCodeResults: 4,
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
            text: `Please explain the following ${languageName} code. Be very detailed and specific, and indicate when it is not clear to you what is going on. Format your response as an ordered list.\n\`\`\`\n${selection.selectedText}\n\`\`\`\n${MARKDOWN_FORMAT_PROMPT}`,
        }

        // Display text
        const displayText = `Explain the following code:\n\`\`\`\n${selection.selectedText}\n\`\`\``

        return {
            displayText,
            promptMessage,
            botResponsePrefix: '',
            contextMessages,
        }
    }
}

export class ExplainCodeHighLevel implements Recipe {
    public getID(): string {
        return 'explainCodeHighLevel'
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

        // Context messages
        const contextQuery = truncateText(selection.selectedText, maxInputTokens)
        const contextMessages = await getEmbeddingsContextMessages(contextQuery, {
            numCodeResults: 4,
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
            text: `Explain the following ${languageName} code at a high level. Only include details that are essential to an overal understanding of what's happening in the code.\n\`\`\`\n${selection.selectedText}\n\`\`\`\n${MARKDOWN_FORMAT_PROMPT}`,
        }

        // Display text
        const displayText = `Explain the following code at a high level:\n\`\`\`\n${selection.selectedText}\n\`\`\``

        return {
            displayText,
            promptMessage,
            botResponsePrefix: '',
            contextMessages,
        }
    }
}
