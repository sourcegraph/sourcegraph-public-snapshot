import { SourcegraphGraphQLAPIClient, EmbeddingsSearchResults } from '../sourcegraph-api/graphql'

import { EmbeddingsSearch } from '.'

export class SourcegraphEmbeddingsSearchClient implements EmbeddingsSearch {
    constructor(private client: SourcegraphGraphQLAPIClient, private repoId: string | null, private web: boolean = true) { }

    public async search(
        query: string,
        codeResultsCount: number,
        textResultsCount: number
    ): Promise<EmbeddingsSearchResults | Error> {
        if (this.web) {
            return this.client.searchEmbeddings(this.repoId ? [this.repoId] : [], query, codeResultsCount, textResultsCount)
        }

        return { codeResults: [], textResults: [] }
        // return this.client.legacySearchEmbeddings(this.repoId, query, codeResultsCount, textResultsCount)
    }
}
