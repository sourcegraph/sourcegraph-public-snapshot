import { Configuration } from '../configuration'
import { EmbeddingsSearch } from '../embeddings'
import { GraphContextFetcher } from '../graph-context'
import { FilenameContextFetcher, KeywordContextFetcher, ContextResult } from '../local-context'
import { isMarkdownFile, populateCodeContextTemplate, populateMarkdownContextTemplate } from '../prompt/templates'
import { Message } from '../sourcegraph-api'
import { EmbeddingsSearchResult } from '../sourcegraph-api/graphql/client'
import { UnifiedContextFetcher } from '../unified-context'
import { isError } from '../utils'

import { ContextMessage, ContextFile, getContextMessageWithResponse } from './messages'

export interface ContextSearchOptions {
    numCodeResults: number
    numTextResults: number
}

export class CodebaseContext {
    private embeddingResultsError = ''
    constructor(
        private config: Pick<Configuration, 'useContext' | 'serverEndpoint'>,
        private codebase: string | undefined,
        private embeddings: EmbeddingsSearch | null,
        private keywords: KeywordContextFetcher | null,
        private filenames: FilenameContextFetcher | null,
        private graph: GraphContextFetcher | null,
        private unifiedContextFetcher?: UnifiedContextFetcher | null,
        private rerank?: (query: string, results: ContextResult[]) => Promise<ContextResult[]>
    ) {}

    public getCodebase(): string | undefined {
        return this.codebase
    }

    public onConfigurationChange(newConfig: typeof this.config): void {
        this.config = newConfig
    }

    private mergeContextResults(keywordResults: ContextResult[], filenameResults: ContextResult[]): ContextResult[] {
        // Just take the single most relevant filename suggestion for now. Otherwise, because our reranking relies solely
        // on filename, the filename results would dominate the keyword results.
        const merged = filenameResults.slice(-1).concat(keywordResults)

        const uniques = new Map<string, ContextResult>()
        for (const result of merged) {
            uniques.set(result.fileName, result)
        }

        return Array.from(uniques.values())
    }

    /**
     * Returns list of context messages for a given query, sorted in *reverse* order of importance (that is,
     * the most important context message appears *last*)
     */
    public async getContextMessages(query: string, options: ContextSearchOptions): Promise<ContextMessage[]> {
        switch (this.config.useContext) {
            case 'unified':
                return this.getUnifiedContextMessages(query, options)
            case 'keyword':
                return this.getLocalContextMessages(query, options)
            case 'none':
                return []
            default:
                return this.embeddings
                    ? this.getEmbeddingsContextMessages(query, options)
                    : this.getLocalContextMessages(query, options)
        }
    }

    public checkEmbeddingsConnection(): boolean {
        return !!this.embeddings
    }

    public getEmbeddingSearchErrors(): string {
        return this.embeddingResultsError.trim()
    }

    public async getSearchResults(
        query: string,
        options: ContextSearchOptions
    ): Promise<{ results: ContextResult[] | EmbeddingsSearchResult[]; endpoint: string }> {
        if (this.embeddings && this.config.useContext !== 'keyword') {
            return {
                results: await this.getEmbeddingSearchResults(query, options),
                endpoint: this.config.serverEndpoint,
            }
        }
        return {
            results:
                (await this.keywords?.getSearchContext(query, options.numCodeResults + options.numTextResults)) || [],
            endpoint: this.config.serverEndpoint,
        }
    }

    // We split the context into multiple messages instead of joining them into a single giant message.
    // We can gradually eliminate them from the prompt, instead of losing them all at once with a single large messeage
    // when we run out of tokens.
    private async getEmbeddingsContextMessages(
        query: string,
        options: ContextSearchOptions
    ): Promise<ContextMessage[]> {
        const combinedResults = await this.getEmbeddingSearchResults(query, options)

        return groupResultsByFile(combinedResults)
            .reverse() // Reverse results so that they appear in ascending order of importance (least -> most).
            .flatMap(groupedResults => this.makeContextMessageWithResponse(groupedResults))
    }

    private async getEmbeddingSearchResults(
        query: string,
        options: ContextSearchOptions
    ): Promise<EmbeddingsSearchResult[]> {
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
            this.embeddingResultsError = `Error retrieving embeddings: ${embeddingsSearchResults}`
            return []
        }
        this.embeddingResultsError = ''
        return embeddingsSearchResults.codeResults.concat(embeddingsSearchResults.textResults)
    }

    private makeContextMessageWithResponse(groupedResults: { file: ContextFile; results: string[] }): ContextMessage[] {
        const contextTemplateFn = isMarkdownFile(groupedResults.file.fileName)
            ? populateMarkdownContextTemplate
            : populateCodeContextTemplate

        return groupedResults.results.flatMap<Message>(text =>
            getContextMessageWithResponse(
                contextTemplateFn(text, groupedResults.file.fileName, groupedResults.file.repoName),
                groupedResults.file
            )
        )
    }

    private async getUnifiedContextMessages(query: string, options: ContextSearchOptions): Promise<ContextMessage[]> {
        if (!this.unifiedContextFetcher) {
            return []
        }

        const results = await this.unifiedContextFetcher.getContext(
            query,
            options.numCodeResults,
            options.numTextResults
        )

        if (isError(results)) {
            console.error('Error retrieving context:', results)
            return []
        }

        return results.flatMap(({ content, filePath, repoName, revision }) => {
            const messageText = isMarkdownFile(filePath)
                ? populateMarkdownContextTemplate(content, filePath, repoName)
                : populateCodeContextTemplate(content, filePath, repoName)

            return getContextMessageWithResponse(messageText, { fileName: filePath, repoName, revision })
        })
    }

    private async getLocalContextMessages(query: string, options: ContextSearchOptions): Promise<ContextMessage[]> {
        const keywordResultsPromise = this.getKeywordSearchResults(query, options)
        const filenameResultsPromise = this.getFilenameSearchResults(query, options)

        const [keywordResults, filenameResults] = await Promise.all([keywordResultsPromise, filenameResultsPromise])

        const combinedResults = this.mergeContextResults(keywordResults, filenameResults)
        const rerankedResults = await (this.rerank ? this.rerank(query, combinedResults) : combinedResults)
        const messages = resultsToMessages(rerankedResults)
        return messages
    }

    private async getKeywordSearchResults(query: string, options: ContextSearchOptions): Promise<ContextResult[]> {
        if (!this.keywords) {
            return []
        }
        const results = await this.keywords.getContext(query, options.numCodeResults + options.numTextResults)
        return results
    }

    private async getFilenameSearchResults(query: string, options: ContextSearchOptions): Promise<ContextResult[]> {
        if (!this.filenames) {
            return []
        }
        const results = await this.filenames.getContext(query, options.numCodeResults + options.numTextResults)
        return results
    }

    /** NOTE(auguste): not part of getContextMessages as this is more intricate; need to think of better abstraction */
    public async getGraphContextMessages(): Promise<ContextMessage[]> {
        // NOTE(auguste): I recommend checking out populateCodeContextTemplate and using
        // that in the long-term, but this will do for now :)

        if (!this.graph) {
            return []
        }

        const contextMessages: ContextMessage[] = []
        const preciseContext = await this.graph.getContext()

        for (const context of preciseContext) {
            contextMessages.push({
                speaker: 'human',
                file: {
                    fileName: '',
                    repoName: context.repository,
                },
                text: `Here is the code snippet: ${context.text}`,
            })
            contextMessages.push({ speaker: 'assistant', text: 'okay' })
        }

        return contextMessages

        // contextMessages.push({
        //     speaker: 'human',
        //     file: {
        //         fileName: 'filename',
        //         repoName: 'repoName',
        //         revision: 'revision',
        //     },
        //     text: `
        //         Here is the path to the file ${test.data.search.results.results[0].file.path}.
        //         The kind of the symbol is a ${test.data.search.results.results[0].symbols.kind}.
        //         The name of the symbol is a ${test.data.search.results.results[0].symbols.name}.
        //         It is located in ${test.data.search.results.results[0].symbols.url}
        //         This is the content of the file ${test.data.search.results.results[0].file.content}.
        //     `,
        // })
        // contextMessages.push({
        //     speaker: 'assistant',
        //     text: 'okay',
        // })
    }
}

function groupResultsByFile(results: EmbeddingsSearchResult[]): { file: ContextFile; results: string[] }[] {
    const originalFileOrder: ContextFile[] = []
    for (const result of results) {
        if (!originalFileOrder.find((ogFile: ContextFile) => ogFile.fileName === result.fileName)) {
            originalFileOrder.push({ fileName: result.fileName, repoName: result.repoName, revision: result.revision })
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

    return originalFileOrder.map(file => ({
        file,
        results: mergeConsecutiveResults(resultsGroupedByFile.get(file.fileName)!),
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

function resultsToMessages(results: ContextResult[]): ContextMessage[] {
    return results.flatMap(({ content, fileName, repoName, revision }) => {
        const messageText = populateCodeContextTemplate(content, fileName, repoName)
        return getContextMessageWithResponse(messageText, { fileName, repoName, revision })
    })
}
