import { EmbeddingsSearchResults } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql/client'

export interface EmbeddingsSearch {
    search(query: string, codeResultsCount: number, textResultsCount: number): Promise<EmbeddingsSearchResults | Error>
}
