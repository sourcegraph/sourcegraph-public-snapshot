import { SourcegraphGraphQLAPIClient, EmbeddingsSearchResults } from '../sourcegraph-api/graphql'

import { Embeddings } from '.'

export class EmbeddingsClient implements Embeddings {
    constructor(private client: SourcegraphGraphQLAPIClient, private repoId: string) {}

    public async search(
        query: string,
        codeResultsCount: number,
        textResultsCount: number
    ): Promise<EmbeddingsSearchResults | Error> {
        return this.client.searchEmbeddings(this.repoId, query, codeResultsCount, textResultsCount)
    }

    public async isContextRequiredForQuery(query: string): Promise<boolean | Error> {
        return this.client.isContextRequiredForQuery(query)
    }
}
