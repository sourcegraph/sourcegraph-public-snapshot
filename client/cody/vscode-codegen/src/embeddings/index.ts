import { EmbeddingsSearchResults } from '../sourcegraph-api/graphql/client'

export interface Embeddings {
    search(query: string, codeResultsCount: number, textResultsCount: number): Promise<EmbeddingsSearchResults | Error>
    isContextRequiredForQuery(query: string): Promise<boolean | Error>
}
