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
export const MAX_REFERENCE_PAGE_REQUESTS = 10

export interface ReferencesResponse {
    references: {
        nodes: LocationConnectionNode[]
        pageInfo: { endCursor?: string }
    }
}

const referencesQuery = gql`
    query LegacyReferences(
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
                        references(line: $line, character: $character, after: $after) {
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

/** Retrieve references for the current hover position. */

export async function* referencesForPosition(
    textDocument: sourcegraph.TextDocument,
    position: sourcegraph.Position,
    queryGraphQL: QueryGraphQLFn<GenericLSIFResponse<ReferencesResponse | null>> = sgQueryGraphQL
): AsyncGenerator<sourcegraph.Location[] | null, void, undefined> {
    yield* concat(
        getQueryPage(textDocument, position, queryGraphQL, referencePageForPosition)(MAX_REFERENCE_PAGE_REQUESTS)
    )
}

/** Retrieve a single page of references for the current hover position. */
export async function referencePageForPosition(
    textDocument: sourcegraph.TextDocument,
    position: sourcegraph.Position,
    after: string | undefined,
    queryGraphQL: QueryGraphQLFn<GenericLSIFResponse<ReferencesResponse | null>> = sgQueryGraphQL
): Promise<LocationCursor> {
    return referenceResponseToLocations(
        textDocument,
        await queryLSIF(
            {
                query: referencesQuery,
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
 * Convert a GraphQL reference response into a set of Sourcegraph locations and end cursor.
 *
 * @param textDocument The current document.
 * @param lsifObject The resolved LSIF object.
 */
function referenceResponseToLocations(
    textDocument: sourcegraph.TextDocument,
    lsifObject: ReferencesResponse | null
): { locations: sourcegraph.Location[] | null; endCursor?: string } {
    if (!lsifObject) {
        return { locations: null }
    }

    return {
        locations: lsifObject.references.nodes.map(node => nodeToLocation(textDocument, node)),
        endCursor: lsifObject.references.pageInfo.endCursor,
    }
}
