import { getGraphQLClient, GraphQLClient } from '@sourcegraph/http-client'
import polyfillEventSource from '@sourcegraph/shared/src/polyfills/vendor/eventSource'
import { generateCache } from '@sourcegraph/shared/src/backend/apolloCache'
import { UserQuery, Query, SearchQuery, SearchResult, SearchResults } from './Query'
import { SearchEvent, SearchMatch, StreamSearchOptions, search, LATEST_VERSION, MessageHandlers, messageHandlers, switchAggregateSearchResults, AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { of, OperatorFunction, Observable, pipe } from 'rxjs'
import { map } from 'rxjs/operators'

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

export interface PageInfo {
  start: number
  perPage: number
}

export type SearchMatches = SearchMatch[]
export interface PaginatedResult<T> {
  results: T
  next: PageInfo
}

export interface SearchService {
  searchQuery(query: string): Promise<SearchResults>
  doQuery<T>(query: Query<T>): Promise<T>
  searchStream(query: string): Observable<SearchMatch[]>
}

export { SearchEvent, AggregateStreamingSearchResults }


export interface SourcegraphService {
  Users: UserService
  Search: SearchService
}

export const createService = async (config: Config): Promise<SourcegraphService> => {
  const { endpoint, token, sudoUsername } = config

  const graphqlClient = await GraphQLAPIClient.create(endpoint, token, sudoUsername ?? '')
  const streamClient = new StreamAPIClient(endpoint, token, sudoUsername ?? '')
  return new SourcegraphClient(graphqlClient, streamClient)
}

export class SourcegraphClient implements SourcegraphService, UserService, SearchService {
  private graphql: GraphQLAPIClient
  private streamer: StreamAPIClient
  Users: UserService = this
  Search: SearchService = this

  constructor(client: GraphQLAPIClient, streamClient: StreamAPIClient) {
    this.graphql = client
    this.streamer = streamClient
  }

  searchStream(query: string): Observable<SearchMatch[]> {
    return this.streamer.search(query)
  }

  async searchQuery(query: string): Promise<SearchResult[]> {
    return await this.doQuery(new SearchQuery(query))
  }

  async doQuery<T>(query: Query<T>): Promise<T> {
    return await this.graphql.fetch(query)
  }

  async currentUsername(): Promise<string> {
    const q = new UserQuery()

    const result = await this.graphql.fetch(q)
    return result
  }
}


export class ClientFactory {
  private baseURL: string
  private token: string
  private sudoUsername?: string

  constructor(url: string, token: string, sudoUsername?: string) {
    this.baseURL = url
    this.token = token
    this.sudoUsername = sudoUsername
  }

  async createGraphQLAPIClient(): Promise<GraphQLAPIClient> {
    return await GraphQLAPIClient.create(this.baseURL, this.token, this.sudoUsername ?? "")
  }

  async createStreamAPIClient(): Promise<StreamAPIClient> {
    return new StreamAPIClient(this.baseURL, this.token, this.sudoUsername ?? "")
  }

}

function authZHeader(token: string, sudoUsername: string): Record<string, string> {
  const authz = sudoUsername.length > 0 ? `token - sudo user = "${sudoUsername}", token = "${token}"` : `token ${token}`
  return { Authorization: authz }
}

export class GraphQLAPIClient {
  private client: GraphQLClient

  static async create(url: string, token: string, sudoUsername: string): Promise<GraphQLAPIClient> {
    const headers: RequestInit['headers'] = {
      'X-Requested-With': `Sourcegraph - Backstage plugin DEV`,
      ...authZHeader(token, sudoUsername)
    }

    try {
      const client: GraphQLClient = await getGraphQLClient({
        baseUrl: url,
        headers: headers,
        isAuthenticated: true,
        cache: generateCache(),
      })
      return new GraphQLAPIClient(client)
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
}

class StreamAPIClient {
  private readonly baseURL: string


  constructor(url: string, token: string, sudoUsername?: string) {
    console.log('stream ctor')
    polyfillEventSource(
      {
        'X-Requested-With': 'Sourcegraph Backstage DEV',
        ...authZHeader(token, sudoUsername ?? "")
      },
      undefined,
    )
    this.baseURL = url
  }

  doStream(query: string): Observable<SearchEvent> {
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
      // matches: (type, eventSource, observer) => {
      //   return observeMessages(type, eventSource).subscribe(data => {
      //     return observer.next(data)
      //   })
      // }
    }
    return search(of(query), opts, handlers)
  }

  paginate(query: string, start: number = 0, perPage: number = 30): Observable<AggregateStreamingSearchResults> {
    const fn: OperatorFunction<SearchEvent, AggregateStreamingSearchResults> = pipe(
      switchAggregateSearchResults,
    )
    const stream = this.doStream(query).pipe(fn)
    return stream
  }

  search(query: string): Observable<SearchMatch[]> {
    const fn: OperatorFunction<SearchEvent, AggregateStreamingSearchResults> = pipe(
      switchAggregateSearchResults,
    )
    const stream = this.doStream(query).pipe(fn, map((r: AggregateStreamingSearchResults) => r.results))
    return stream
  }

}
