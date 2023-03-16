import path from 'path'

import { Editor } from '../../editor'
import { Message } from '../../sourcegraph-api'
import { ContextSearchOptions } from '../context-search-options'
import { getContextMessageWithResponse, populateCodeContextTemplate, truncateText, truncateTextStart } from '../prompt'

import { MARKDOWN_FORMAT_PROMPT, getNormalizedLanguageName } from './helpers'
import { Recipe, RecipePrompt } from './recipe'

export class GenerateDocstring implements Recipe {
    getID(): string {
        return 'generateDocstring'
    }
    async getPrompt(
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
        const displayText = `Generate documentation for the following code:\n\`\`\`\n${selection.selectedText}\n\`\`\``

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
        const extension = getFileExtension(selection.fileName)
        const languageName = getNormalizedLanguageName(selection.fileName)
        const promptPrefix = `Generate a comment documenting the parameters and functionality of the following ${languageName} code:`
        let additionalInstructions = `Use the ${languageName} documentation style to generate a ${languageName} comment.`
        if (extension === 'java') {
            additionalInstructions = 'Use the JavaDoc documentation style to generate a Java comment.'
        } else if (extension === 'py') {
            additionalInstructions = 'Use a Python docstring to generate a Python multi-line string.'
        }
        const promptMessage: Message = {
            speaker: 'human',
            text: `${promptPrefix}\n\`\`\`\n${selection.selectedText}\n\`\`\`\n Only generate the documentation, do not generate the code. ${additionalInstructions} ${MARKDOWN_FORMAT_PROMPT}`,
        }

        // Response prefix
        let docStart = ''
        if (extension === 'java' || extension.startsWith('js') || extension.startsWith('ts')) {
            docStart = '/*'
        } else if (extension === 'py') {
            docStart = '"""\n'
        } else if (extension === 'go') {
            docStart = '// '
        }
        const botResponsePrefix = `Here is the generated documentation:\n\`\`\`${extension}\n${docStart}`

        return {
            displayText,
            promptMessage,
            botResponsePrefix,
            contextMessages,
        }
    }
}

export function getFileExtension(fileName: string): string {
    return path.extname(fileName).slice(1).toLowerCase()
}
