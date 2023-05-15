import fetch from 'isomorphic-fetch'

import { buildGraphQLUrl } from '@sourcegraph/http-client'

import { ConfigurationWithAccessToken } from '../../configuration'
import { isError } from '../../utils'

import {
    CURRENT_USER_ID_QUERY,
    IS_CONTEXT_REQUIRED_QUERY,
    REPOSITORY_ID_QUERY,
    SEARCH_EMBEDDINGS_QUERY,
    LOG_EVENT_MUTATION,
    REPOSITORY_EMBEDDING_EXISTS_QUERY,
} from './queries'

interface APIResponse<T> {
    data?: T
    errors?: { message: string; path?: string[] }[]
}

interface CurrentUserIdResponse {
    currentUser: { id: string } | null
}

interface RepositoryIdResponse {
    repository: { id: string } | null
}

interface RepositoryEmbeddingExistsResponse {
    repository: { id: string; embeddingExists: boolean } | null
}

interface EmbeddingsSearchResponse {
    embeddingsSearch: EmbeddingsSearchResults
}

interface LogEventResponse {}

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
        return new Error(response.errors.map(({ message }) => message).join(', '))
    }
    if (!response.data) {
        return new Error('response is missing data')
    }
    return extract(response.data)
}

export class SourcegraphGraphQLAPIClient {
    private dotcomUrl = 'https://sourcegraph.com'

    constructor(
        private config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>
    ) {}

    public onConfigurationChange(newConfig: typeof this.config): void {
        this.config = newConfig
    }

    public async getCurrentUserId(): Promise<string | Error> {
        return this.fetchSourcegraphAPI<APIResponse<CurrentUserIdResponse>>(CURRENT_USER_ID_QUERY, {}).then(response =>
            extractDataOrError(response, data =>
                data.currentUser ? data.currentUser.id : new Error('current user not found')
            )
        )
    }

    public async getRepoId(repoName: string): Promise<string | Error> {
        return this.fetchSourcegraphAPI<APIResponse<RepositoryIdResponse>>(REPOSITORY_ID_QUERY, {
            name: repoName,
        }).then(response =>
            extractDataOrError(response, data =>
                data.repository ? data.repository.id : new RepoNotFoundError(`repository ${repoName} not found`)
            )
        )
    }

    public async getRepoIdIfEmbeddingExists(repoName: string): Promise<string | null | Error> {
        return this.fetchSourcegraphAPI<APIResponse<RepositoryEmbeddingExistsResponse>>(
            REPOSITORY_EMBEDDING_EXISTS_QUERY,
            {
                name: repoName,
            }
        ).then(response =>
            extractDataOrError(response, data => (data.repository?.embeddingExists ? data.repository.id : null))
        )
    }

    public async logEvent(event: {
        event: string
        userCookieID: string
        url: string
        source: string
        argument?: string | {}
        publicArgument?: string | {}
    }): Promise<void | Error> {
        try {
            if (this.config.serverEndpoint === this.dotcomUrl) {
                await this.fetchSourcegraphAPI<APIResponse<LogEventResponse>>(LOG_EVENT_MUTATION, event).then(
                    response => {
                        extractDataOrError(response, data => {})
                    }
                )
            } else {
                await Promise.all([
                    this.fetchSourcegraphAPI<APIResponse<LogEventResponse>>(LOG_EVENT_MUTATION, event).then(
                        response => {
                            extractDataOrError(response, data => {})
                        }
                    ),
                    this.fetchSourcegraphDotcomAPI<APIResponse<LogEventResponse>>(LOG_EVENT_MUTATION, event).then(
                        response => {
                            extractDataOrError(response, data => {})
                        }
                    ),
                ])
            }
        } catch (error) {
            return error
        }
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

    private fetchSourcegraphAPI<T>(query: string, variables: Record<string, any>): Promise<T | Error> {
        const headers = new Headers(this.config.customHeaders as HeadersInit)
        headers.set('Content-Type', 'application/json; charset=utf-8')
        if (this.config.accessToken) {
            headers.set('Authorization', `token ${this.config.accessToken}`)
        }

        const url = buildGraphQLUrl({ request: query, baseUrl: this.config.serverEndpoint })
        return fetch(url, {
            method: 'POST',
            body: JSON.stringify({ query, variables }),
            headers,
        })
            .then(verifyResponseCode)
            .then(response => response.json() as T)
            .catch(error => new Error(`accessing Sourcegraph GraphQL API: ${error} (${url})`))
    }

    // make an anonymous request to the dotcom API
    private async fetchSourcegraphDotcomAPI<T>(query: string, variables: Record<string, any>): Promise<T | Error> {
        const url = buildGraphQLUrl({ request: query, baseUrl: this.dotcomUrl })
        return fetch(url, {
            method: 'POST',
            body: JSON.stringify({ query, variables }),
        })
            .then(verifyResponseCode)
            .then(response => response.json() as T)
            .catch(error => new Error(`error fetching Sourcegraph GraphQL API: ${error} (${url})`))
    }
}

function verifyResponseCode(response: Response): Response {
    if (!response.ok) {
        throw new Error(`HTTP status code: ${response.status}`)
    }
    return response
}

class RepoNotFoundError extends Error {}
export const isRepoNotFoundError = (value: unknown): value is RepoNotFoundError => value instanceof RepoNotFoundError
