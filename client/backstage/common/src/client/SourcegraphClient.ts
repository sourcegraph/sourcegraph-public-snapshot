import { Config } from '@backstage/config';
import { GraphQLClient, gql } from 'graphql-request';

export interface Query<T> {
  gql(): string
  vars(): string
  Marshal(data: any): T[]
}

export interface SearchResult {
  readonly repository: string
  readonly fileContent: string

}

export class SearchQuery implements Query<SearchResult> {
  private readonly query: string

  constructor(query: string) {
    this.query = query;
  }

  Marshal(data: any): SearchResult[] {
    const results = new Array<SearchResult>();

    for (let v in data.search.results.results) {
      let { repository, file: { fileContent }
      } = v as any
      results.push({ repository, fileContent })
    }

    return results
  }

  vars(): any {
    return { search: this.query }
  }

  gql(): string {
    return gql`
      query ($search: String!) {
        search(query: $search) {
          results {
            __typename
            ... on FileMatch {
              repository
            }
            file {
              content
            }
          }
        }
      }
    `
  }

}

class UserQuery implements Query<string> {
  Marshal(data: any): string[] {
    if ("currentUser" in data) {
      return [data.currentUser.username];
    }
    throw new Error("username not found")
  }
  vars(): string {
    return ""
  }
  gql(): string {
    return gql`
    query {
      currentUser {
        username
      }
    }
    `
  }

}


export class SourcegraphClient {
  private client: GraphQLClient

  static create(config: Config): SourcegraphClient {
    const endpoint = config.getString("sourcegraph.endpoint")
    const token = config.getString("sourcegraph.token")
    const sudoUsername = config.getOptionalString("sourcegraph.sudoUsername")

    return new SourcegraphClient(endpoint, token, sudoUsername || "")
  }

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

  async ping(): Promise<string> {
    const q = new UserQuery()

    const data = await this.fetch(q)
    return data[0]
  }

  async fetch<T>(q: Query<T>): Promise<T[]> {
    const data = await this.client.request(q.gql(), q.vars())

    return q.Marshal(data)
  }
}

