import { Configuration } from '../configuration'
import { EmbeddingsSearch } from '../embeddings'
import { KeywordContextFetcher } from '../keyword-context'
import { isMarkdownFile, populateCodeContextTemplate, populateMarkdownContextTemplate } from '../prompt/templates'
import { Message } from '../sourcegraph-api'
import { EmbeddingsSearchResult } from '../sourcegraph-api/graphql/client'
import { isError } from '../utils'

import { ContextMessage, getContextMessageWithResponse } from './messages'

export interface ContextSearchOptions {
    numCodeResults: number
    numTextResults: number
}

export class CodebaseContext {
    constructor(
        private config: Pick<Configuration, 'useContext'>,
        private embeddings: EmbeddingsSearch | null,
        private keywords: KeywordContextFetcher
    ) {}

    public onConfigurationChange(newConfig: typeof this.config): void {
        this.config = newConfig
    }

    public async getContextMessages(query: string, options: ContextSearchOptions): Promise<ContextMessage[]> {
        switch (this.config.useContext) {
            case 'embeddings' || 'blended':
                return this.embeddings
                    ? this.getEmbeddingsContextMessages(query, options)
                    : this.getKeywordContextMessages(query, options)
            case 'keyword':
                return this.getKeywordContextMessages(query, options)
            default:
                return this.getEmbeddingsContextMessages(query, options)
        }
    }

    public checkEmbeddingsConnection(): boolean {
        return !!this.embeddings
    }

    // We split the context into multiple messages instead of joining them into a single giant message.
    // We can gradually eliminate them from the prompt, instead of losing them all at once with a single large messeage
    // when we run out of tokens.
    private async getEmbeddingsContextMessages(
        query: string,
        options: ContextSearchOptions
    ): Promise<ContextMessage[]> {
        if (!this.embeddings) {
            return []
        }

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

    private async getKeywordContextMessages(query: string, options: ContextSearchOptions): Promise<ContextMessage[]> {
        const results = await this.keywords.getContext(query, options.numCodeResults + options.numTextResults)
        return results.flatMap(({ content, fileName }) => {
            const messageText = populateCodeContextTemplate(content, fileName)
            return getContextMessageWithResponse(messageText, fileName)
        })
    }
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
