import fetch from 'isomorphic-fetch'

import { buildGraphQLUrl } from '@sourcegraph/http-client'

import { ConfigurationWithAccessToken } from '../../configuration'
import { isError } from '../../utils'

import {
    CURRENT_USER_ID_QUERY,
    IS_CONTEXT_REQUIRED_QUERY,
    REPOSITORY_ID_QUERY,
    REPOSITORY_IDS_QUERY,
    REPOSITORY_NAMES_QUERY,
    SEARCH_ATTRIBUTION_QUERY,
    SEARCH_EMBEDDINGS_QUERY,
    LEGACY_SEARCH_EMBEDDINGS_QUERY,
    LOG_EVENT_MUTATION,
    REPOSITORY_EMBEDDING_EXISTS_QUERY,
    CURRENT_USER_ID_AND_VERIFIED_EMAIL_QUERY,
    CURRENT_SITE_VERSION_QUERY,
    CURRENT_SITE_HAS_CODY_ENABLED_QUERY,
    CURRENT_SITE_GRAPHQL_FIELDS_QUERY,
    GET_CODY_CONTEXT_QUERY,
    CURRENT_SITE_CODY_LLM_CONFIGURATION,
} from './queries'

interface APIResponse<T> {
    data?: T
    errors?: { message: string; path?: string[] }[]
}

interface SiteVersionResponse {
    site: { productVersion: string } | null
}

interface SiteGraphqlFieldsResponse {
    __type: { fields: { name: string }[] } | null
}

interface SiteHasCodyEnabledResponse {
    site: { isCodyEnabled: boolean } | null
}

interface CurrentUserIdResponse {
    currentUser: { id: string } | null
}

interface CurrentUserIdHasVerifiedEmailResponse {
    currentUser: { id: string; hasVerifiedEmail: boolean } | null
}

interface RepositoryIdResponse {
    repository: { id: string } | null
}

interface RepositoryIdsResponse {
    repositories: { nodes: { id: string; name: string }[] }
}

interface RepositoryNamesResponse {
    repositories: { nodes: { id: string; name: string }[] }
}

interface RepositoryEmbeddingExistsResponse {
    repository: { id: string; embeddingExists: boolean } | null
}

interface EmbeddingsSearchResponse {
    embeddingsSearch: EmbeddingsSearchResults
}

interface EmbeddingsMultiSearchResponse {
    embeddingsMultiSearch: EmbeddingsSearchResults
}

interface CodyFileChunkContext {
    __typename: 'FileChunkContext'
    blob: {
        path: string
        repository: {
            id: string
            name: string
        }
        commit: {
            id: string
            oid: string
        }
    }
    startLine: number
    endLine: number
    chunkContent: string
}

type GetCodyContextResult = CodyFileChunkContext | null

interface GetCodyContextResponse {
    getCodyContext: GetCodyContextResult[]
}

interface SearchAttributionResponse {
    snippetAttribution: {
        limitHit: boolean
        nodes: { repositoryName: string }[]
    }
}

interface LogEventResponse {}

export interface EmbeddingsSearchResult {
    repoName?: string
    revision?: string
    fileName: string
    startLine: number
    endLine: number
    content: string
}

export interface EmbeddingsSearchResults {
    codeResults: EmbeddingsSearchResult[]
    textResults: EmbeddingsSearchResult[]
}

export interface SearchAttributionResults {
    limitHit: boolean
    nodes: { repositoryName: string }[]
}

export interface CodyLLMSiteConfiguration {
    chatModel?: string
    chatModelMaxTokens?: number
    fastChatModel?: string
    fastChatModelMaxTokens?: number
    completionModel?: string
    completionModelMaxTokens?: number
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

    public onConfigurationChange(
        newConfig: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>
    ): void {
        this.config = newConfig
    }

    public isDotCom(): boolean {
        return new URL(this.config.serverEndpoint).origin === new URL(this.dotcomUrl).origin
    }

    public async getSiteVersion(): Promise<string | Error> {
        return this.fetchSourcegraphAPI<APIResponse<SiteVersionResponse>>(CURRENT_SITE_VERSION_QUERY, {}).then(
            response =>
                extractDataOrError(response, data =>
                    // Example values: "5.1.0" or "222587_2023-05-30_5.0-39cbcf1a50f0" for insider builds
                    data.site?.productVersion ? data.site?.productVersion : new Error('site version not found')
                )
        )
    }

    public async getSiteHasIsCodyEnabledField(): Promise<boolean | Error> {
        return this.fetchSourcegraphAPI<APIResponse<SiteGraphqlFieldsResponse>>(
            CURRENT_SITE_GRAPHQL_FIELDS_QUERY,
            {}
        ).then(response =>
            extractDataOrError(response, data => !!data.__type?.fields?.find(field => field.name === 'isCodyEnabled'))
        )
    }

    public async getSiteHasCodyEnabled(): Promise<boolean | Error> {
        return this.fetchSourcegraphAPI<APIResponse<SiteHasCodyEnabledResponse>>(
            CURRENT_SITE_HAS_CODY_ENABLED_QUERY,
            {}
        ).then(response => extractDataOrError(response, data => data.site?.isCodyEnabled ?? false))
    }

    public async getCurrentUserId(): Promise<string | Error> {
        return this.fetchSourcegraphAPI<APIResponse<CurrentUserIdResponse>>(CURRENT_USER_ID_QUERY, {}).then(response =>
            extractDataOrError(response, data =>
                data.currentUser ? data.currentUser.id : new Error('current user not found')
            )
        )
    }

    public async getCurrentUserIdAndVerifiedEmail(): Promise<{ id: string; hasVerifiedEmail: boolean } | Error> {
        return this.fetchSourcegraphAPI<APIResponse<CurrentUserIdHasVerifiedEmailResponse>>(
            CURRENT_USER_ID_AND_VERIFIED_EMAIL_QUERY,
            {}
        ).then(response =>
            extractDataOrError(response, data =>
                data.currentUser ? { ...data.currentUser } : new Error('current user not found with verified email')
            )
        )
    }

    public async getCodyLLMConfiguration(): Promise<undefined | CodyLLMSiteConfiguration | Error> {
        const response = await this.fetchSourcegraphAPI<APIResponse<any>>(CURRENT_SITE_CODY_LLM_CONFIGURATION)

        return extractDataOrError(response, data => data.site?.codyLLMConfiguration)
    }

    public async getRepoIds(names: string[]): Promise<{ id: string; name: string }[] | Error> {
        return this.fetchSourcegraphAPI<APIResponse<RepositoryIdsResponse>>(REPOSITORY_IDS_QUERY, {
            names,
            first: names.length,
        }).then(response => extractDataOrError(response, data => data.repositories?.nodes))
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

    public async getRepoNames(first: number): Promise<string[] | Error> {
        return this.fetchSourcegraphAPI<APIResponse<RepositoryNamesResponse>>(REPOSITORY_NAMES_QUERY, { first }).then(
            response =>
                extractDataOrError(
                    response,
                    data => data?.repositories?.nodes?.map((node: { id: string; name: string }) => node?.name) || []
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

    /**
     * Checks if Cody is enabled on the current Sourcegraph instance.
     *
     * @returns
     * enabled: Whether Cody is enabled.
     * version: The Sourcegraph version.
     *
     * This method first checks the Sourcegraph version using `getSiteVersion()`.
     * If the version is before 5.0.0, Cody is disabled.
     * If the version is 5.0.0 or newer, it checks for the existence of the `isCodyEnabled` field using `getSiteHasIsCodyEnabledField()`.
     * If the field exists, it calls `getSiteHasCodyEnabled()` to check its value.
     * If the field does not exist, Cody is assumed to be enabled for versions between 5.0.0 - 5.1.0.
     */
    public async isCodyEnabled(): Promise<{ enabled: boolean; version: string }> {
        // Check site version.
        const siteVersion = await this.getSiteVersion()
        if (isError(siteVersion)) {
            return { enabled: false, version: 'unknown' }
        }
        const insiderBuild = siteVersion.length > 12 || siteVersion.includes('dev')
        if (insiderBuild) {
            return { enabled: true, version: siteVersion }
        }
        // NOTE: Cody does not work on versions older than 5.0
        const versionBeforeCody = siteVersion < '5.0.0'
        if (versionBeforeCody) {
            return { enabled: false, version: siteVersion }
        }
        // Beta version is betwewen 5.0.0 - 5.1.0 and does not have isCodyEnabled field
        const betaVersion = siteVersion >= '5.0.0' && siteVersion < '5.1.0'
        const hasIsCodyEnabledField = await this.getSiteHasIsCodyEnabledField()
        // The isCodyEnabled field does not exist before version 5.1.0
        if (!betaVersion && !isError(hasIsCodyEnabledField) && hasIsCodyEnabledField) {
            const siteHasCodyEnabled = await this.getSiteHasCodyEnabled()
            return { enabled: !isError(siteHasCodyEnabled) && siteHasCodyEnabled, version: siteVersion }
        }
        return { enabled: insiderBuild || betaVersion, version: siteVersion }
    }

    public async logEvent(event: {
        event: string
        userCookieID: string
        url: string
        source: string
        argument?: string | {}
        publicArgument?: string | {}
    }): Promise<void | Error> {
        if (process.env.CODY_TESTING === 'true') {
            console.log(`not logging ${event.event} in test mode`)
            return
        }
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

    public async getCodyContext(
        repos: string[],
        query: string,
        codeResultsCount: number,
        textResultsCount: number
    ): Promise<GetCodyContextResult[] | Error> {
        return this.fetchSourcegraphAPI<APIResponse<GetCodyContextResponse>>(GET_CODY_CONTEXT_QUERY, {
            repos,
            query,
            codeResultsCount,
            textResultsCount,
        }).then(response => extractDataOrError(response, data => data.getCodyContext))
    }

    public async searchEmbeddings(
        repos: string[],
        query: string,
        codeResultsCount: number,
        textResultsCount: number
    ): Promise<EmbeddingsSearchResults | Error> {
        return this.fetchSourcegraphAPI<APIResponse<EmbeddingsMultiSearchResponse>>(SEARCH_EMBEDDINGS_QUERY, {
            repos,
            query,
            codeResultsCount,
            textResultsCount,
        }).then(response => extractDataOrError(response, data => data.embeddingsMultiSearch))
    }

    // (Naman): This is a temporary workaround for supporting vscode cody integrated with older version of sourcegraph which do not support the latest searchEmbeddings query.
    public async legacySearchEmbeddings(
        repo: string,
        query: string,
        codeResultsCount: number,
        textResultsCount: number
    ): Promise<EmbeddingsSearchResults | Error> {
        return this.fetchSourcegraphAPI<APIResponse<EmbeddingsSearchResponse>>(LEGACY_SEARCH_EMBEDDINGS_QUERY, {
            repo,
            query,
            codeResultsCount,
            textResultsCount,
        }).then(response => extractDataOrError(response, data => data.embeddingsSearch))
    }

    public async searchAttribution(snippet: string): Promise<SearchAttributionResults | Error> {
        return this.fetchSourcegraphAPI<APIResponse<SearchAttributionResponse>>(SEARCH_ATTRIBUTION_QUERY, {
            snippet,
        }).then(response => extractDataOrError(response, data => data.snippetAttribution))
    }

    public async isContextRequiredForQuery(query: string): Promise<boolean | Error> {
        return this.fetchSourcegraphAPI<APIResponse<IsContextRequiredForChatQueryResponse>>(IS_CONTEXT_REQUIRED_QUERY, {
            query,
        }).then(response => extractDataOrError(response, data => data.isContextRequiredForChatQuery))
    }

    private fetchSourcegraphAPI<T>(query: string, variables: Record<string, any> = {}): Promise<T | Error> {
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
