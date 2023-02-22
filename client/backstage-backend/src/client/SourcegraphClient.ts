import { getGraphQLClient, GraphQLClient } from '@sourcegraph/http-client'
import { generateCache } from '@sourcegraph/shared/src/backend/apolloCache'
import { UserQuery, Query, SearchQuery, SearchResult, SearchResults } from './Query'
import { SearchEvent, StreamSearchOptions, search, LATEST_VERSION, MessageHandlers, messageHandlers, observeMessages, SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { fromEvent, Observable, of, OperatorFunction } from 'rxjs'
import { SearchPatternType } from '@sourcegraph/../shared/src/graphql-operations'

export interface Config {
  endpoint: string
  token: string
  sudoUsername?: string
}

export interface QueryOptions {
  Size: Number
}

export interface UserService {
  currentUsername(): Promise<string>
}

export interface SearchService {
  searchQuery(query: string): Promise<SearchResults>
  doQuery<T>(query: Query<T>): Promise<T>
}


export interface SourcegraphService {
  Users: UserService
  Search: SearchService
}

export const createService = async (config: Config): Promise<SourcegraphService> => {
  const { endpoint, token, sudoUsername } = config
  const base = await BaseClient.create(endpoint, token, sudoUsername || '')
  return new SourcegraphClient(base)
}

export class SourcegraphClient implements SourcegraphService, UserService, SearchService {
  private client: BaseClient
  Users: UserService = this
  Search: SearchService = this

  constructor(client: BaseClient) {
    this.client = client
  }

  async searchQuery(query: string): Promise<SearchResult[]> {
    return await this.doQuery(new SearchQuery(query))
  }

  async doQuery<T>(query: Query<T>): Promise<T> {
    return await this.client.fetch(query)
  }

  async currentUsername(): Promise<string> {
    const q = new UserQuery()

    const result = await this.client.fetch(q)
    return result
  }
}

export class BaseClient {
  private client: GraphQLClient
  private baseURL: string

  static async create(baseURL: string, token: string, sudoUsername: string): Promise<BaseClient> {
    const authz =
      sudoUsername?.length > 0 ? `token - sudo user = "${sudoUsername}", token = "${token}"` : `token ${token}`
    const headers: RequestInit['headers'] = {
      'X-Requested-With': `Sourcegraph - Backstage plugin DEV`,
      Authorization: authz,
    }

    this.baseURL = baseURL

    try {
      const client: GraphQLClient = await getGraphQLClient({
        baseUrl: baseURL,
        headers: headers,
        isAuthenticated: true,
        cache: generateCache(),
      })
      return new BaseClient(client)
    } catch (e) {
      throw new Error(`failed to create graphsql client: ${e}`)
    }
  }
  constructor(client: GraphQLClient) {
    this.client = client
  }

  async fetch<T>(query: Query<T>): Promise<T> {
    const { data } = await this.client.query({
      query: query.gql(),
      variables: query.vars(),
    })
    if (!data) {
      throw new Error('grapql request failed: no data')
    }
    return query.marshal(data)
  }

  async stream(query: string): Promise<any> {
    const opts: StreamSearchOptions = {
      version: LATEST_VERSION,
      patternType: SearchPatternType.standard,
      caseSensitive: false,
      trace: undefined,
      sourcegraphURL: this.baseURL,
      chunkMatches: true
    }

    const handlers: MessageHandlers = {
      ...messageHandlers,
      matches: (type, eventSource, observer) => {
        return observeMessages(type, eventSource).subscribe(data => { observer.next(data) })
      }
    }
    const f: OperatorFunction<SearchEvent, SearchMatch> = (e: SearchEvent): SearchMatch | {} => {
      return e.data
    }
    const r = search(of(`${query}`), opts, handlers).pipe(f)
    return r.toPromise()
  }
