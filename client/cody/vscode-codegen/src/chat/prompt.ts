import path from 'path'

import { Editor } from '../editor'
import { Embeddings } from '../embeddings'
import { IntentDetector } from '../intent-detector'
import { KeywordContextFetcher } from '../keyword-context'
import { Message } from '../sourcegraph-api'
import { EmbeddingsSearchResult } from '../sourcegraph-api/graphql/client'
import { isError } from '../utils'

import { ContextSearchOptions } from './context-search-options'
import { getRecipe } from './recipes/index'

const PROMPT_PREAMBLE_LENGTH = 230
const MAX_PROMPT_TOKEN_LENGTH = 7000 - PROMPT_PREAMBLE_LENGTH
const SOLUTION_TOKEN_LENGTH = 1000
const MAX_HUMAN_INPUT_TOKENS = 1000
const MAX_RECIPE_INPUT_TOKENS = 2000
const MAX_RECIPE_SURROUNDING_TOKENS = 500
const MAX_AVAILABLE_PROMPT_LENGTH = MAX_PROMPT_TOKEN_LENGTH - SOLUTION_TOKEN_LENGTH
export const MAX_CURRENT_FILE_TOKENS = 4000
const CHARS_PER_TOKEN = 4

export interface ContextMessage extends Message {
    filename?: string
}

/**
 * Each TranscriptChunk corresponds to a sequence of messages that should be considered as a unit during prompt construction.
 * - Typically, `actual` has length 1 and represents the actual message incorporated into the prompt.
 * - `context` is messages that include code snippets fetched as contextual knowledge.
 *    These should not be displayed in the chat GUI.
 * - `display` are messages that should replace `actual` in the chat GUI.
 */
export interface TranscriptChunk {
    actual: Message[]
    context: ContextMessage[]
    display?: Message[]
}

export class Transcript {
    private transcript: TranscriptChunk[] = []

    constructor(
        private contextType: 'embeddings' | 'keyword' | 'none' | 'blended',
        private embeddings: Embeddings | null,
        private intentDetector: IntentDetector,
        private keywords: KeywordContextFetcher,
        private editor: Editor
    ) {}

    public getTranscript(): TranscriptChunk[] {
        return this.transcript
    }

    public getDisplayMessages(): Message[] {
        return this.transcript.flatMap(({ display, actual }) => display || actual)
    }

    private getUnderlyingMessages(): Message[] {
        return this.transcript.flatMap(({ actual }) => actual)
    }

    public getLastContextFiles(): string[] {
        for (const chunk of [...this.transcript].reverse()) {
            if (chunk.actual.length === 0) {
                continue
            }
            if (chunk.actual[chunk.actual.length - 1].speaker === 'assistant') {
                continue
            }
            return chunk.context.flatMap(msg => msg.filename || [])
        }
        return []
    }

    private async getCodebaseContextMessages(query: string): Promise<ContextMessage[]> {
        const intentOrError = await this.intentDetector.detect(query)
        if (isError(intentOrError)) {
            console.error('Error detecting message intent:', intentOrError)
            return []
        }
        const { needsCodebaseContext, needsCurrentFileContext } = intentOrError
        if (needsCurrentFileContext) {
            const activeTextEditor = this.editor.getActiveTextEditor()
            if (!activeTextEditor) {
                return []
            }
            const truncatedDocumentText = truncateText(activeTextEditor.content, MAX_CURRENT_FILE_TOKENS)
            return [
                {
                    filename: path.basename(activeTextEditor.filePath),
                    speaker: 'human',
                    text: `Here is the current open file to add to your knowledge base:\n\`\`\`\n${truncatedDocumentText}\n\`\`\``,
                },
                {
                    speaker: 'assistant',
                    text: 'Ok, added this current open file to my knowledge base.',
                },
            ]
        }

        // Only load context messages for the first question in the transcript
        if (this.transcript.length > 0 || (!needsCurrentFileContext && !needsCodebaseContext)) {
            return []
        }

        const options = {
            numCodeResults: 8,
            numTextResults: 2,
        }

        switch (this.contextType) {
            case 'blended' || 'embeddings':
                return await this.getEmbeddingsContextMessages(query, options, needsCodebaseContext)
            case 'keyword':
                return await this.keywords.getContextMessages(query)
            default:
                return []
        }
    }

    // We split the context into multiple messages instead of joining them into a single giant message.
    // We can gradually eliminate them from the prompt, instead of losing them all at once with a single large messeage
    // when we run out of tokens.
    private async getEmbeddingsContextMessages(
        query: string,
        options: ContextSearchOptions,
        needsCodebaseContext: boolean = false
    ): Promise<ContextMessage[]> {
        // Not falling back to keywords context if contextType is set to 'embeddings'
        const fallbackToKeywords = this.contextType === 'blended' && needsCodebaseContext
        if (!this.embeddings) {
            console.log('no embeddings client for current codebase')
            // fallback to keyword context when embedding client not available but needs codebase context
            return fallbackToKeywords ? await this.keywords.getContextMessages(query) : []
        }
        if (!(await this.embeddings.isContextRequiredForQuery(query))) {
            console.log('embeddings: no context needed')
            return []
        }

        console.log('fetching embeddings context')
        const embeddingsSearchResults = await this.embeddings.search(
            query,
            options.numCodeResults,
            options.numTextResults
        )
        if (isError(embeddingsSearchResults)) {
            console.error('Error retrieving embeddings:', embeddingsSearchResults)
            return []
        }

        const combinedResults = embeddingsSearchResults.codeResults.concat(embeddingsSearchResults.textResults)

        if (!combinedResults.length && fallbackToKeywords) {
            return await this.keywords.getContextMessages(query)
        }

        return groupResultsByFile(combinedResults)
            .reverse() // Reverse results so that they appear in ascending order of importance (least -> most).
            .flatMap(groupedResults => {
                const contextTemplateFn = isMarkdownFile(groupedResults.fileName)
                    ? populateMarkdownContextTemplate
                    : populateCodeContextTemplate

                return groupedResults.results.flatMap<Message>(text =>
                    getContextMessageWithResponse(
                        contextTemplateFn(text, groupedResults.fileName),
                        groupedResults.fileName
                    )
                )
            })
    }

    private addMessage(chunk: TranscriptChunk): void {
        this.transcript.push(chunk)
    }

    // addHumanMessage adds a human message to the transcript, along the way computing any context
    // messages that should be incorporated into the prompt.
    // This should only be invoked with the last message was from 'assistant'.
    // Returns the prompt that should be sent to fetch the bot response (same as calling `getPrompt`)
    public async addHumanMessage(humanInput: string): Promise<Message[]> {
        const actualMessages = this.getUnderlyingMessages()
        if (actualMessages.length > 0 && actualMessages[actualMessages.length - 1].speaker === 'human') {
            throw new Error('attempt to add human message when last message was human')
        }

        const truncatedHumanInput = truncateText(humanInput, MAX_HUMAN_INPUT_TOKENS)
        const contextMessages = await this.getCodebaseContextMessages(humanInput)
        const humanMessage: Message = {
            speaker: 'human',
            text: contextMessages.length > 0 ? humanInput : truncatedHumanInput,
        }

        this.addMessage({
            actual: [humanMessage],
            context: contextMessages,
        })

        return this.getPrompt()
    }

    public addBotMessage(text: string): void {
        this.addMessage({
            actual: [{ speaker: 'assistant', text }],
            context: [],
        })
    }

    // getPrompt takes the current transcript (both hidden and displayed) and generates a prompt
    // to send to the server to generate the next bot message. This should only be invoked
    // when the last message in the transcript was from 'human'.
    //
    // The prompt construction algorithm is as follows:
    // - Iterate through chunks with most recent first
    // - Add the `actual` messages of the chunk to the prompt
    // - If the chunk has context messages, incorporate them if you haven't yet incorporated context messages from any other chunk.
    //   - Note: this means we only include context messages of the most recent chunk that has them
    // - Visit the next chunk. Repeat until you run out of token budget.
    // - At the end, incorporate the botResponsePrefix (which controls the first part of the bot response if you wish to constrain that).
    public getPrompt(botResponsePrefix = ''): Message[] {
        const reversePrompt: Message[] = []
        const reverseTranscript = [...this.transcript].reverse()
        let tokenBudget = MAX_AVAILABLE_PROMPT_LENGTH
        let incorporatedContext = false
        for (let i = 0; i < reverseTranscript.length; i++) {
            const chunk = reverseTranscript[i]
            for (const msg of [...chunk.actual].reverse()) {
                const tokenUsage = estimateTokensUsage(msg)
                if (tokenUsage <= tokenBudget) {
                    reversePrompt.push(msg)
                    tokenBudget -= tokenUsage
                } else {
                    break
                }
            }
            if (i === 0) {
                if (reversePrompt.length === 0) {
                    throw new Error('last message size exceeded token window')
                } else if (reversePrompt[0].speaker !== 'human') {
                    throw new Error('last message was not human')
                }
            }

            if (!incorporatedContext && chunk.context.length >= 2) {
                for (let j = chunk.context.length - 1; j >= 1; j -= 2) {
                    const humanMsg = chunk.context[j - 1]
                    const botMsg = chunk.context[j]
                    const combinedTokensUsage = estimateTokensUsage(humanMsg) + estimateTokensUsage(botMsg)

                    if (combinedTokensUsage <= tokenBudget) {
                        reversePrompt.push(botMsg, humanMsg)
                        tokenBudget -= combinedTokensUsage
                    } else {
                        break
                    }
                }
                incorporatedContext = true
            }
        }

        const prompt = [...reversePrompt].reverse()
        if (botResponsePrefix) {
            prompt.push({ speaker: 'assistant', text: botResponsePrefix })
        }
        return prompt
    }

    public async resetToRecipe(recipeID: string): Promise<{
        prompt: Message[]
        display: Message[]
        botResponsePrefix: string
    } | null> {
        const recipe = getRecipe(recipeID)
        if (!recipe) {
            return null
        }
        const prompt = await recipe.getPrompt(
            MAX_RECIPE_INPUT_TOKENS + MAX_RECIPE_SURROUNDING_TOKENS,
            this.editor,
            (query: string, options: ContextSearchOptions): Promise<Message[]> =>
                this.getEmbeddingsContextMessages(query, options)
        )
        if (!prompt) {
            return null
        }

        this.reset()
        const { displayText, contextMessages, promptMessage, botResponsePrefix } = prompt

        this.addMessage({
            display: [{ speaker: 'human', text: displayText }],
            actual: [promptMessage],
            context: contextMessages,
        })

        return {
            display: this.getDisplayMessages(),
            prompt: this.getPrompt(botResponsePrefix),
            botResponsePrefix,
        }
    }

    public reset(): void {
        this.transcript = []
    }

    public setEmbeddings(embeddings: Embeddings | null): void {
        this.embeddings = embeddings
    }

    public setIntentDetector(intentDetector: IntentDetector): void {
        this.intentDetector = intentDetector
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

function groupResultsByFile(results: EmbeddingsSearchResult[]): { fileName: string; results: string[] }[] {
    const originalFileOrder: string[] = []
    for (const result of results) {
        if (!originalFileOrder.includes(result.fileName)) {
            originalFileOrder.push(result.fileName)
        }
    }

    const resultsGroupedByFile = new Map<string, EmbeddingsSearchResult[]>()
    for (const result of results) {
        const results = resultsGroupedByFile.get(result.fileName)
        if (results === undefined) {
            resultsGroupedByFile.set(result.fileName, [result])
        } else {
            resultsGroupedByFile.set(result.fileName, results.concat([result]))
        }
    }

    return originalFileOrder.map(fileName => ({
        fileName,
        results: mergeConsecutiveResults(resultsGroupedByFile.get(fileName)!),
    }))
}

function mergeConsecutiveResults(results: EmbeddingsSearchResult[]): string[] {
    const sortedResults = results.sort((a, b) => a.startLine - b.startLine)
    const mergedResults = [results[0].content]

    for (let i = 1; i < sortedResults.length; i++) {
        const result = sortedResults[i]
        const previousResult = sortedResults[i - 1]

        if (result.startLine === previousResult.endLine) {
            mergedResults[mergedResults.length - 1] = mergedResults[mergedResults.length - 1] + result.content
        } else {
            mergedResults.push(result.content)
        }
    }

    return mergedResults
}

const MARKDOWN_EXTENSIONS = new Set(['md', 'markdown'])

function isMarkdownFile(filePath: string): boolean {
    const extension = path.extname(filePath).slice(1)
    return MARKDOWN_EXTENSIONS.has(extension)
}

const CODE_CONTEXT_TEMPLATE = `Use following code snippet from file \`{filePath}\`:
\`\`\`{language}
{text}
\`\`\``

export function populateCodeContextTemplate(code: string, filePath: string): string {
    const language = path.extname(filePath).slice(1)
    return CODE_CONTEXT_TEMPLATE.replace('{filePath}', filePath).replace('{language}', language).replace('{text}', code)
}

const MARKDOWN_CONTEXT_TEMPLATE = 'Use the following text from file `{filePath}`:\n{text}'

export function populateMarkdownContextTemplate(md: string, filePath: string): string {
    return MARKDOWN_CONTEXT_TEMPLATE.replace('{filePath}', filePath).replace('{text}', md)
}

export function getContextMessageWithResponse(text: string, filename?: string): ContextMessage[] {
    return [
        { speaker: 'human', text, filename },
        { speaker: 'assistant', text: 'Ok.' },
    ]
}
