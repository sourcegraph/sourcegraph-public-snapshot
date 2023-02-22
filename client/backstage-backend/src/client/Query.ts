import { gql } from '@apollo/client'
import type { DocumentNode } from '@apollo/client'

export interface Query<T> {
  toString(): string
  gql(): DocumentNode
  vars(): string
  marshal(data: any): T
}

export type SearchResults = Array<SearchResult>
export interface SearchResult {
  readonly __typename?: string
  readonly repository: string
  readonly filename?: string
  readonly fileContent?: string
}

export class SearchQuery implements Query<SearchResults> {
  private readonly query: string
  static gql: DocumentNode = gql(`
            query ($search: String!) {
                search(query: $search) {
                    results {
                        results {
                            __typename
                            ... on FileMatch {
                                repository {
                                    name
                                }
                                file {
                                    name
                                    content
                                }
                            }
                        }
                    }
                }
            }
        `)

  constructor(query: string) {
    this.query = query
  }

  marshal(data: any): SearchResults {
    const results = new Array<SearchResult>()
    if (!data.search) {
      return results
    }

    for (const value of data.search.results.results) {
      let filename = ''
      let fileContent = ''
      if ('file' in value) {
        filename = value.file.name
        fileContent = value.file.content
      }

      results.push({ repository: value.repository.name, filename: filename, fileContent: fileContent })
    }

    return results
  }

  vars(): any {
    return { search: this.query }
  }

  gql(): DocumentNode {
    return SearchQuery.gql
  }

  toString(): string {
    return this.query
  }
}

export class SearchRepoQuery extends SearchQuery {
  static gql: DocumentNode = gql(`
            query ($search: String!) {
                search(query: $search) {
                    results {
                        results {
                            __typename
                            repository {
                                name
                            }
                        }
                    }
                }
            }
        `)

  constructor(query: string) {
    super(query)
  }

  gql(): DocumentNode {
    return SearchRepoQuery.gql
  }
}

export class UserQuery implements Query<string> {
  static gql: DocumentNode = gql(`
            query {
                currentUser {
                    username
                }
            }
        `)

  vars(): any {
    return {}
  }

  gql(): DocumentNode {
    return UserQuery.gql
  }

  marshal(data: any): string {
    if ('currentUser' in data) {
      const { currentUser } = data
      return currentUser.username
    }
    throw new Error('currentUser field missing')
  }

  toString(): string {
    return ""
  }
}
