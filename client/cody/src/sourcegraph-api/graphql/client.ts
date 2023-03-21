import fetch, { Response } from 'node-fetch'

import { isError } from '../../utils'

import { IS_CONTEXT_REQUIRED_QUERY, REPOSITORY_ID_QUERY, SEARCH_EMBEDDINGS_QUERY } from './queries'

interface APIResponse<T> {
    data?: T
    errors?: string[]
}

interface RepositoryIdResponse {
    repository: { id: string } | null
}

interface EmbeddingsSearchResponse {
    embeddingsSearch: EmbeddingsSearchResults
}

export interface EmbeddingsSearchResult {
    fileName: string
    startLine: number
    endLine: number
    content: string
}

export interface EmbeddingsSearchResults {
    codeResults: EmbeddingsSearchResult[]
    textResults: EmbeddingsSearchResult[]
}

interface IsContextRequiredForChatQueryResponse {
    isContextRequiredForChatQuery: boolean
}

function extractDataOrError<T, R>(response: APIResponse<T> | Error, extract: (data: T) => R): R | Error {
    if (isError(response)) {
        return response
    }
    if (response.errors && response.errors.length > 0) {
        return new Error(response.errors.join(', '))
    }
    if (!response.data) {
        return new Error('response is missing data')
    }
    return extract(response.data)
}

export class SourcegraphGraphQLAPIClient {
    constructor(private instanceUrl: string, private accessToken: string) {}

    public async getRepoId(repoName: string): Promise<string | Error> {
        return this.fetchSourcegraphAPI<APIResponse<RepositoryIdResponse>>(REPOSITORY_ID_QUERY, {
            name: repoName,
        }).then(response =>
            extractDataOrError(response, data =>
                data.repository ? data.repository.id : new Error(`repository ${repoName} not found`)
            )
        )
    }

    public async searchEmbeddings(
        repo: string,
        query: string,
        codeResultsCount: number,
        textResultsCount: number
    ): Promise<EmbeddingsSearchResults | Error> {
        return this.fetchSourcegraphAPI<APIResponse<EmbeddingsSearchResponse>>(SEARCH_EMBEDDINGS_QUERY, {
            repo,
            query,
            codeResultsCount,
            textResultsCount,
        }).then(response => extractDataOrError(response, data => data.embeddingsSearch))
    }

    public async isContextRequiredForQuery(query: string): Promise<boolean | Error> {
        return this.fetchSourcegraphAPI<APIResponse<IsContextRequiredForChatQueryResponse>>(IS_CONTEXT_REQUIRED_QUERY, {
            query,
        }).then(response => extractDataOrError(response, data => data.isContextRequiredForChatQuery))
    }

    private async fetchSourcegraphAPI<T>(query: string, variables: Record<string, any>): Promise<T | Error> {
        return fetch(`${this.instanceUrl}/.api/graphql`, {
            headers: { authorization: `token ${this.accessToken}` },
            method: 'POST',
            body: JSON.stringify({ query, variables }),
        })
            .then(verifyResponseCode)
            .then(response => response.json() as T)
            .catch(() => new Error('error fetching Sourcegraph GraphQL API'))
    }
}

function verifyResponseCode(response: Response): Response {
    if (!response.ok) {
        throw new Error('HTTP status code: ' + response.status)
    }
    return response
}
