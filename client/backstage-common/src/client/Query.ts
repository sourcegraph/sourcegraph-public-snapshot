import { gql } from 'graphql-request'
import type { AuthenticatedUser } from '../../../shared/src/auth'
import { currentAuthStateQuery } from '../../../shared/src/auth'

export interface Query<T> {
    gql(): string
    vars(): string
    marshal(data: any): T
}

export interface SearchResult {
    readonly repository: string
    readonly filename: string
    readonly fileContent: string
}

export class SearchQuery implements Query<SearchResult[]> {
    private readonly query: string

    constructor(query: string) {
        this.query = query
    }

    marshal(data: any): SearchResult[] {
        const results = new Array<SearchResult>()
        if (!data.search) {
            // TODO(@burmudar): remove - only temporary
            console.log('undefined data.search')
            return results
        }

        // TODO(@burmudar): remove - only temporary
        console.log('raw', data)

        for (const v of data.search.results.results) {
            results.push({ repository: v.repository.name, filename: v.file.name, fileContent: v.file.content })
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
    }
}

// UserQuery is purely used as a sanity check and will be removed in the future
// TODO(@burmudar): remove
export class UserQuery implements Query<string> {
    vars(): string {
        return ''
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

    marshal(data: any): string {
        if ('currentUser' in data) {
            return data.currentUser.username
        }
        throw new Error('username not found')
    }
}

export type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
export class AuthenticatedUserQuery implements Query<AuthenticatedUser> {
    gql(): string {
        return currentAuthStateQuery
    }

    vars(): string {
        return ''
    }

    marshal(data: any): AuthenticatedUser {
        return data.currentUser
    }
}
