import { dataOrThrowErrors, requestGraphQLCommon } from '@sourcegraph/http-client'
import { map } from 'rxjs/operators'
import { UserQuery, Query, AuthenticatedUser, AuthenticatedUserQuery, SearchQuery, SearchResult } from './Query'

export interface Config {
    endpoint: string
    token: string
    sudoUsername?: string
}

export interface UserService {
    currentUsername(): Promise<string>
    getAuthenticatedUser(): Promise<AuthenticatedUser>
}

export const createService = (config: Config): SourcegraphService => {
    const { endpoint, token, sudoUsername } = config
    const base = new BaseClient(endpoint, token, sudoUsername || '')
    return new SourcegraphClient(base)
}

export const createDummySearch = (): SearchService => {
    return {
        searchQuery: async (_: string): Promise<SearchResult[]> => {
            console.log('DummySearch not doing anything')
            return []
        },
    }
}

export interface SearchService {
    searchQuery(query: string): Promise<SearchResult[]>
}

export interface SourcegraphService {
    Users: UserService
    Search: SearchService
}

class SourcegraphClient implements SourcegraphService, UserService, SearchService {
    private client: BaseClient
    Users: UserService = this
    Search: SearchService = this

    constructor(client: BaseClient) {
        this.client = client
    }

    async searchQuery(query: string): Promise<SearchResult[]> {
        const q = new SearchQuery(query)
        const results = await this.client.fetch(q)

        return results
    }

    async currentUsername(): Promise<string> {
        const q = new UserQuery()

        const data = await this.client.fetch(q)
        return data[0]
    }

    async getAuthenticatedUser(): Promise<AuthenticatedUser> {
        const q = new AuthenticatedUserQuery()
        const data = await this.client.fetch(q)
        return data
    }
}

type Headers = { [header: string]: string }

class BaseClient {
    private baseUrl: string
    private headers: Headers

    constructor(baseUrl: string, token: string, sudoUsername: string) {
        const authz =
            sudoUsername?.length > 0 ? `token - sudo user = "${sudoUsername}", token = "${token}"` : `token ${token} `
        this.baseUrl = baseUrl
        this.headers = {
            'X-Requested-With': `Sourcegraph - Backstage plugin DEV`,
            Authorization: authz,
        }
    }

    async fetch<T>(query: Query<T>): Promise<T> {
        return requestGraphQLCommon<T, string>({
            request: query.gql(),
            baseUrl: this.baseUrl,
            variables: query.vars(),
            headers: this.headers,
        })
            .pipe(map(dataOrThrowErrors), map(query.marshal))
            .toPromise()
    }
}
