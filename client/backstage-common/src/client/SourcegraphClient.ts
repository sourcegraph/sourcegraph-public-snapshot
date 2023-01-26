import { GraphQLClient } from 'graphql-request';
import { UserQuery, Query, AuthenticatedUser, AuthenticatedUserQuery, SearchQuery, SearchResult } from './Query';



export interface Config {
  endpoint: string,
  token: string,
  sudoUsername?: string
}

export interface UserService {
  CurrentUsername(): Promise<string>
  GetAuthenticatedUser(): Promise<AuthenticatedUser>
}

export const createService = (config: Config): SourcegraphService => {
  const { endpoint, token, sudoUsername } = config
  const base = new BaseClient(endpoint, token, sudoUsername || "")
  return new SourcegraphClient(base)
}

export interface SearchService {
  SearchQuery(query: string): Promise<SearchResult[]>
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

  async SearchQuery(query: string): Promise<SearchResult[]> {
    const q = new SearchQuery(query)
    const data = await this.client.fetch(q)

    return q.Marshal(data)
  }

  async CurrentUsername(): Promise<string> {
    const q = new UserQuery()

    const data = await this.client.fetch(q)
    return data[0]
  }

  async GetAuthenticatedUser(): Promise<AuthenticatedUser> {
    const q = new AuthenticatedUserQuery()
    const data = await this.client.fetch(q)
    return data[0]
  }

}

class BaseClient {
  private client: GraphQLClient

  constructor(baseUrl: string, token: string, sudoUsername: string) {
    const authz = sudoUsername?.length > 0 ? `token-sudo user="${sudoUsername}",token="${token}"` : `token ${token}`
    const apiUrl = `${baseUrl}/.api/graphql`
    this.client = new GraphQLClient(apiUrl,
      {
        headers: {
          'X-Requested-With': `Sourcegraph - Backstage plugin DEV`,
          Authorization: authz,
        }
      })
  }

  async fetch<T>(q: Query<T>): Promise<T[]> {
    const data = await this.client.request(q.gql(), q.vars())

    return q.Marshal(data)
  }
}

