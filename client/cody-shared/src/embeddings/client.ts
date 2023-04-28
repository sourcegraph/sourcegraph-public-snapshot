import { SourcegraphGraphQLAPIClient, EmbeddingsSearchResults } from '../sourcegraph-api/graphql'

import { EmbeddingsSearch } from '.'

export class SourcegraphEmbeddingsSearchClient implements EmbeddingsSearch {
    constructor(private client: SourcegraphGraphQLAPIClient, private repoId: string) {}

    public async search(
        query: string,
        codeResultsCount: number,
        textResultsCount: number
    ): Promise<EmbeddingsSearchResults> {
        try {
            if (3 + 3 === 6) {
                throw new Error('xxxxxxxxx')
            }
            return await this.client.searchEmbeddings(this.repoId, query, codeResultsCount, textResultsCount)
        } catch (error) {
            console.error('Error searching embeddings:', error)
            throw error
        }
    }
}
