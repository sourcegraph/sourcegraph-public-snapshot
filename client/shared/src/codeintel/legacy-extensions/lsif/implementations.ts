import gql from 'tagged-template-noop'

import type * as sourcegraph from '../api'
import { queryGraphQL as sgQueryGraphQL, type QueryGraphQLFn } from '../util/graphql'
import { concat } from '../util/ix'

import { queryLSIF, type GenericLSIFResponse } from './api'
import { nodeToLocation, type LocationConnectionNode, getQueryPage, type LocationCursor } from './locations'

/**
 * The maximum number of chained GraphQL requests to make for a single
 * requests query. The page count for a result set should generally be
 * relatively low unless it's a VERY popular library and LSIF data is
 * ubiquitous (which is our goal).
 */
export const MAX_IMPLEMENTATION_PAGE_REQUESTS = 10

export interface ImplementationsResponse {
    implementations: {
        nodes: LocationConnectionNode[]
        pageInfo: { endCursor?: string }
    }
}

const implementationsQuery = gql`
    query LegacyImplementations(
        $repository: String!
        $commit: String!
        $path: String!
        $line: Int!
        $character: Int!
        $after: String
    ) {
        repository(name: $repository) {
            commit(rev: $commit) {
                blob(path: $path) {
                    lsif {
                        implementations(line: $line, character: $character, after: $after) {
                            nodes {
                                resource {
                                    path
                                    repository {
                                        name
                                    }
                                    commit {
                                        oid
                                    }
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
                            pageInfo {
                                endCursor
                            }
                        }
                    }
                }
            }
        }
    }
`

/** Retrieve implementations for the current hover position. */

export async function* implementationsForPosition(
    textDocument: sourcegraph.TextDocument,
    position: sourcegraph.Position,
    queryGraphQL: QueryGraphQLFn<GenericLSIFResponse<ImplementationsResponse | null>> = sgQueryGraphQL
): AsyncGenerator<sourcegraph.Location[] | null, void, undefined> {
    yield* concat(
        getQueryPage(
            textDocument,
            position,
            queryGraphQL,
            implementationPageForPosition
        )(MAX_IMPLEMENTATION_PAGE_REQUESTS)
    )
}

/** Retrieve a single page of implementations for the current hover position. */
export async function implementationPageForPosition(
    textDocument: sourcegraph.TextDocument,
    position: sourcegraph.Position,
    after: string | undefined,
    queryGraphQL: QueryGraphQLFn<GenericLSIFResponse<ImplementationsResponse | null>> = sgQueryGraphQL
): Promise<LocationCursor> {
    return implementationResponseToLocations(
        textDocument,
        await queryLSIF(
            {
                query: implementationsQuery,
                uri: textDocument.uri,
                after,
                line: position.line,
                character: position.character,
            },
            queryGraphQL
        )
    )
}

/**
 * Convert a GraphQL implementation response into a set of Sourcegraph locations and end cursor.
 *
 * @param doc The current document.
 * @param lsifObj The resolved LSIF object.
 */
function implementationResponseToLocations(
    textDocument: sourcegraph.TextDocument,
    lsifObject: ImplementationsResponse | null
): { locations: sourcegraph.Location[] | null; endCursor?: string } {
    if (!lsifObject) {
        return { locations: null }
    }

    return {
        locations: lsifObject.implementations.nodes.map(node => nodeToLocation(textDocument, node)),
        endCursor: lsifObject.implementations.pageInfo.endCursor,
    }
}
