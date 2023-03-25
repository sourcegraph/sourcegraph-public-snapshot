import {
    SourcegraphGraphQLAPIClient,
    EmbeddingsSearchResults,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'

import { EmbeddingsSearch } from '.'

export class SourcegraphEmbeddingsSearchClient implements EmbeddingsSearch {
    constructor(private client: SourcegraphGraphQLAPIClient, private repoId: string) {}

    public async search(
        query: string,
        codeResultsCount: number,
        textResultsCount: number
    ): Promise<EmbeddingsSearchResults | Error> {
        return this.client.searchEmbeddings(this.repoId, query, codeResultsCount, textResultsCount)
    }
}
