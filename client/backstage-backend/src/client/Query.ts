import { gql } from '@apollo/client'
import type { DocumentNode } from '@apollo/client'

export interface Query<T> {
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
    static raw: string = `
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
        `

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
        return gql(SearchQuery.raw)
    }
}

export class SearchRepoQuery extends SearchQuery {
    static raw: string = `
            query ($search: String!) {
                search(query: $search) {
                    results {
                        results {
                            __typename
                            ... on FileMatch {
                                repository {
                                    name
                                }
                            }
                        }
                    }
                }
            }
        `

    constructor(query: string) {
        super(query)
    }

    gql(): DocumentNode {
        return gql(SearchRepoQuery.raw)
    }
}

export class UserQuery implements Query<string> {
    static raw: string = `
            query {
                currentUser {
                    username
                }
            }
        `

    vars(): any {
        return {}
    }

    gql(): DocumentNode {
        return gql(UserQuery.raw)
    }

    marshal(data: any): string {
        if ('currentUser' in data) {
            const { currentUser } = data
            return currentUser.username
        }
        throw new Error('currentUser field missing')
    }
}
