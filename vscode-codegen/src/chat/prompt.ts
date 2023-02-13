import path from 'path'

import { Message } from '@sourcegraph/cody-common'

import { EmbeddingsClient, EmbeddingSearchResult } from '../embeddings-client'
import { ContextSearchOptions } from './context-search-options'
import { getRecipe } from './recipes/index'

const MAX_PROMPT_TOKEN_LENGTH = 7000
const SOLUTION_TOKEN_LENGTH = 1000
const MAX_HUMAN_INPUT_TOKENS = 1000
const MAX_RECIPE_INPUT_TOKENS = 2000
const MAX_RECIPE_SURROUNDING_TOKENS = 500
const MAX_AVAILABLE_PROMPT_LENGTH = MAX_PROMPT_TOKEN_LENGTH - SOLUTION_TOKEN_LENGTH
const CHARS_PER_TOKEN = 4

export class Prompt {
	private messages: Message[] = []

	constructor(private embeddingsClient: EmbeddingsClient | null) {}

	// We split the context into multiple messages instead of joining them into a single giant message.
	// We can gradually eliminate them from the prompt, instead of losing them all at once with a single large messeage
	// when we run out of tokens.
	private async getContextMessages(query: string, options: ContextSearchOptions): Promise<Message[]> {
		if (!this.embeddingsClient) {
			return []
		}

		const embeddingsSearchResults = await this.embeddingsClient.search(
			query,
			options.numCodeResults,
			options.numMarkdownResults
		)

		const filterFn = options.filterResults ? options.filterResults : () => true
		const combinedResults = embeddingsSearchResults.codeResults
			.concat(embeddingsSearchResults.markdownResults)
			.filter(filterFn)

		return groupResultsByFile(combinedResults)
			.reverse() // Reverse results so that they appear in ascending order of importance (least -> most).
			.flatMap(groupedResults => {
				const contextTemplateFn = isMarkdownFile(groupedResults.filePath)
					? populateMarkdownContextTemplate
					: populateCodeContextTemplate

				return groupedResults.results.flatMap<Message>(text =>
					getContextMessageWithResponse(contextTemplateFn(text, groupedResults.filePath))
				)
			})
	}

	private addInstructionsToHumanInput(humanInput: string): string {
		return `Answer the following question or statement only if you know the answer or can make a well-informed guess; otherwise tell me you don't know it.\n\n${humanInput}`
	}

	async getPromptForMessage(
		message: Message,
		contextMessages: Message[],
		botResponsePrefix: string = ''
	): Promise<Message[]> {
		const messageTokensUsage = estimateTokensUsage(message)
		// Since we are limited by the amount of tokens we can send to the backend, we have to truncate the prompt.
		// In order of priority, we have to fit in:
		//   1. The latest message.
		//   2. The context messages related to the message.
		//   3. Older chat messages.
		const newPromptMessages = []
		// We always want to include the human message in the prompt, so we decrease the available tokens budget
		// by the amount of tokens in the human message.
		let availablePromptTokensBudget = MAX_AVAILABLE_PROMPT_LENGTH - messageTokensUsage

		// The available messages for the next prompt consist of the older messages and the new context messages.
		// Since we traverse the available messages in reverse order (newer -> older) the context messages take a
		// precedence over the older messages.
		const availableMessages = this.messages.concat(contextMessages)
		for (let i = availableMessages.length - 1; i >= 1; i -= 2) {
			const humanMessage = availableMessages[i - 1]
			const botMessage = availableMessages[i]
			const combinedTokensUsage = estimateTokensUsage(humanMessage) + estimateTokensUsage(botMessage)

			// We stop adding pairs of messages once we exceed the available tokens budget.
			if (combinedTokensUsage <= availablePromptTokensBudget) {
				newPromptMessages.push(botMessage, humanMessage)
				availablePromptTokensBudget -= combinedTokensUsage
			} else {
				break
			}
		}

		// TODO: At this point, we could check if the context messages include any duplicates, and remove them.
		// TODO: If we manage to remove some of the duplicates, we could probably squeeze in a couple of older messages.
		// TODO: Although, that risks introducing more duplicates. So, the algorithm would very likely have to do multiple passes.
		// TODO: For now, we are happy with a single pass and keeping the duplicates.

		// Reverse the prompt messages, so they appear in chat order (older -> newer).
		this.messages = newPromptMessages.reverse()
		// Finally, add the human message at the end.
		this.messages.push(message)

		if (botResponsePrefix) {
			this.messages.push({ speaker: 'bot', text: botResponsePrefix })
		}

		return this.messages
	}

	async constructPromptForHumanInput(humanInput: string): Promise<Message[]> {
		const truncatedHumanInput = truncateText(humanInput, MAX_HUMAN_INPUT_TOKENS)

		// TODO: Add context from currently active text editor if embeddingsClient is not available
		const inputNeedsAdditionalContext = this.embeddingsClient
			? await this.embeddingsClient.queryNeedsAdditionalContext(truncatedHumanInput)
			: false

		// Load more context at the start of a chat to give the model as much information as possible.
		// On subsequent messages load less context to avoid too many duplicates and overwhelming
		// the conversation with context messages.
		const numCodeResults = this.messages.length === 0 ? 4 : 1
		const contextMessages = inputNeedsAdditionalContext
			? await this.getContextMessages(truncatedHumanInput, {
					numCodeResults,
					numMarkdownResults: 1,
			  })
			: []
		const humanMessage: Message = {
			speaker: 'you',
			text: contextMessages.length > 0 ? this.addInstructionsToHumanInput(humanInput) : truncatedHumanInput,
		}
		return this.getPromptForMessage(humanMessage, contextMessages)
	}

	async getPromptForRecipe(recipeID: string): Promise<{
		messages: Message[]
		displayText: string
		recipePrefix: string
	} | null> {
		const recipe = getRecipe(recipeID)
		if (!recipe) {
			return null
		}
		const prompt = await recipe.getPrompt(
			MAX_RECIPE_INPUT_TOKENS + MAX_RECIPE_SURROUNDING_TOKENS,
			(query: string, options: ContextSearchOptions): Promise<Message[]> =>
				this.getContextMessages(query, options)
		)
		if (!prompt) {
			return null
		}
		const { displayText, contextMessages, promptMessage, botResponsePrefix } = prompt

		const promptMessages = await this.getPromptForMessage(promptMessage, contextMessages, botResponsePrefix)
		return {
			messages: promptMessages,
			recipePrefix: botResponsePrefix,
			displayText,
		}
	}

	addBotResponse(text: string): void {
		const lastMessage = this.messages[this.messages.length - 1]
		if (lastMessage?.speaker === 'bot') {
			// Remove the last bot message that only contains the prefix, and append the full message below.
			this.messages = this.messages.slice(0, this.messages.length - 1)
		}
		this.messages.push({ speaker: 'bot', text })
	}

	reset(): void {
		this.messages = []
	}
}

export function truncateText(text: string, maxTokens: number): string {
	const maxLength = maxTokens * CHARS_PER_TOKEN
	return text.length <= maxLength ? text : text.slice(0, maxLength)
}

export function truncateTextStart(text: string, maxTokens: number): string {
	const maxLength = maxTokens * CHARS_PER_TOKEN
	return text.length <= maxLength ? text : text.slice(-maxLength - 1)
}

function estimateTokensUsage(message: Message): number {
	return Math.round(message.text.length / CHARS_PER_TOKEN)
}

function groupResultsByFile(results: EmbeddingSearchResult[]): { filePath: string; results: string[] }[] {
	const originalFileOrder: string[] = []
	for (const result of results) {
		if (originalFileOrder.indexOf(result.filePath) === -1) {
			originalFileOrder.push(result.filePath)
		}
	}

	const resultsGroupedByFile = new Map<string, EmbeddingSearchResult[]>()
	for (const result of results) {
		const results = resultsGroupedByFile.get(result.filePath)
		if (results === undefined) {
			resultsGroupedByFile.set(result.filePath, [result])
		} else {
			resultsGroupedByFile.set(result.filePath, results.concat([result]))
		}
	}

	return originalFileOrder.map(filePath => ({
		filePath,
		results: mergeConsecutiveResults(resultsGroupedByFile.get(filePath)!),
	}))
}

function mergeConsecutiveResults(results: EmbeddingSearchResult[]): string[] {
	const sortedResults = results.sort((a, b) => a['start'] - b['start'])
	const mergedResults = [results[0].text]

	for (let i = 1; i < sortedResults.length; i++) {
		const result = sortedResults[i]
		const previousResult = sortedResults[i - 1]

		if (result.start === previousResult.end) {
			mergedResults[mergedResults.length - 1] = mergedResults[mergedResults.length - 1] + result.text
		} else {
			mergedResults.push(result.text)
		}
	}

	return mergedResults
}

const MARKDOWN_EXTENSIONS = new Set(['md', 'markdown'])

function isMarkdownFile(filePath: string): boolean {
	const extension = path.extname(filePath).slice(1)
	return MARKDOWN_EXTENSIONS.has(extension)
}

const CODE_CONTEXT_TEMPLATE = `Add the following code snippet from file \`{filePath}\` to your knowledge base:
\`\`\`{language}
{text}
\`\`\``

export function populateCodeContextTemplate(code: string, filePath: string): string {
	const language = path.extname(filePath).slice(1)
	return CODE_CONTEXT_TEMPLATE.replace('{filePath}', filePath).replace('{language}', language).replace('{text}', code)
}

const MARKDOWN_CONTEXT_TEMPLATE = `Add the following text from file \`{filePath}\` to your knowledge base:\n{text}`

export function populateMarkdownContextTemplate(md: string, filePath: string): string {
	return MARKDOWN_CONTEXT_TEMPLATE.replace('{filePath}', filePath).replace('{text}', md)
}

export function getContextMessageWithResponse(text: string): Message[] {
	return [
		{ speaker: 'you', text: text },
		{ speaker: 'bot', text: 'Ok, adding previous message to my knowledge base.' },
	]
}
