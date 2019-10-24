import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { queryGraphQL } from '../backend/graphql'

/**
 * Fetches symbols.
 */
export function fetchSymbols(
    repo: GQL.ID,
    rev: string,
    args: { first?: number; query?: string; includePatterns?: string[] }
): Observable<GQL.ISymbolConnection> {
    return queryGraphQL(
        gql`
            query Symbols($repo: ID!, $rev: String!, $first: Int, $query: String, $includePatterns: [String!]) {
                node(id: $repo) {
                    ... on Repository {
                        commit(rev: $rev) {
                            symbols(first: $first, query: $query, includePatterns: $includePatterns) {
                                pageInfo {
                                    hasNextPage
                                }
                                nodes {
                                    name
                                    containerName
                                    kind
                                    language
                                    location {
                                        resource {
                                            path
                                        }
                                        range {
                                            start {
                                                line
                                                character
                                            }
                                            end {
                                                line
                                                character
                                            }
                                        }
                                    }
                                    url
                                }
                            }
                        }
                    }
                }
            }
        `,
        { ...args, repo, rev }
    ).pipe(
        map(({ data, errors }) => {
            if (
                !data ||
                !data.node ||
                !(data.node as GQL.IRepository).commit ||
                !(data.node as GQL.IRepository).commit!.symbols ||
                !(data.node as GQL.IRepository).commit!.symbols.nodes
            ) {
                throw createAggregateError(errors)
            }
            return (data.node as GQL.IRepository).commit!.symbols
        })
    )
}
