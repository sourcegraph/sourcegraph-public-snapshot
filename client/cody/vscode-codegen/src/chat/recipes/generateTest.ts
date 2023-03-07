import path from 'path'

import { Message } from '@sourcegraph/cody-common'

import { ContextSearchOptions } from '../context-search-options'
import { getContextMessageWithResponse, populateCodeContextTemplate, truncateText, truncateTextStart } from '../prompt'

import { MARKDOWN_FORMAT_PROMPT, getNormalizedLanguageName, getActiveEditorSelection } from './helpers'
import { Recipe, RecipePrompt } from './recipe'

export class GenerateTest implements Recipe {
	getID(): string {
		return 'generateUnitTest'
	}
	async getPrompt(
		maxTokens: number,
		getEmbeddingsContextMessages: (query: string, options: ContextSearchOptions) => Promise<Message[]>
	): Promise<RecipePrompt | null> {
		const maxInputTokens = Math.round(0.8 * maxTokens)
		const maxSurroundingTokens = Math.round(0.2 * maxTokens)

		// Inputs
		const selection = await getActiveEditorSelection()
		if (!selection) {
			return null
		}

		// Display text
		const displayText = `Generate a unit test for the following code:\n\`\`\`\n${selection.selectedText}\n\`\`\``

		// Context messages
		const contextQuery = truncateText(selection.selectedText, maxInputTokens)
		const contextMessages = await getEmbeddingsContextMessages(contextQuery, {
			numCodeResults: 8,
			numMarkdownResults: 0,
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
			speaker: 'you',
			text: `Generate a unit test in ${languageName} for the following code:\n\`\`\`\n${selection.selectedText}\n\`\`\`\n${MARKDOWN_FORMAT_PROMPT}`,
		}

		// Response prefix
		const extension = getFileExtension(selection.fileName)
		const botResponsePrefix = `Here is the generated unit test:\n\`\`\`${extension}\n`

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
