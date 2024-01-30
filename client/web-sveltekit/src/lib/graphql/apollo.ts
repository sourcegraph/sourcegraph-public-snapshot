import type { KeyArgsFunction, KeySpecifier } from '@apollo/client/cache/inmemory/policies'
import {
    gql,
    ApolloClient,
    InMemoryCache,
    createHttpLink,
    from,
    type HttpOptions,
    type NormalizedCacheObject,
    type OperationVariables,
    type QueryOptions,
    type DocumentNode,
    type FetchPolicy,
    type FieldPolicy,
} from '@apollo/client/core/index'
import { trimEnd, once } from 'lodash'

import { dev } from '$app/environment'
import { createAggregateError } from '$lib/common'
import { GRAPHQL_URI, checkOk } from '$lib/http-client'

interface BuildGraphQLUrlOptions {
    request?: string
    baseUrl?: string
}
/**
 * Constructs GraphQL Request URL
 */
function buildGraphQLUrl({ request, baseUrl }: BuildGraphQLUrlOptions): string {
    const nameMatch = request ? request.match(/^\s*(?:query|mutation)\s+(\w+)/) : ''
    const apiURL = `${GRAPHQL_URI}${nameMatch ? '?' + nameMatch[1] : ''}`
    return baseUrl ? new URL(trimEnd(baseUrl, '/') + apiURL).href : apiURL
}

function getHeaders(): { [header: string]: string } {
    const headers: { [header: string]: string } = {
        ...window?.context?.xhrHeaders,
        Accept: 'application/json',
        'Content-Type': 'application/json',
    }
    const parameters = new URLSearchParams(window.location.search)
    const trace = parameters.get('trace')
    if (trace) {
        headers['X-Sourcegraph-Should-Trace'] = trace
    }
    const feat = parameters.getAll('feat')
    if (feat.length) {
        headers['X-Sourcegraph-Override-Feature'] = feat.join(',')
    }
    return headers
}
const customFetch: HttpOptions['fetch'] = (uri, options) => fetch(uri, options).then(checkOk)

export type GraphQLClient = ApolloClient<NormalizedCacheObject>

/**
 * Creates a field policy for a list-like forward connections. It concatenates the
 * incoming nodes with the existing nodes, and updates the pageInfo.
 */
function listLikeForwardConnection({ keyArgs }: { keyArgs: KeySpecifier | KeyArgsFunction | false }): FieldPolicy {
    return {
        keyArgs,

        merge(existing, incoming) {
            if (!existing) {
                return incoming
            }

            if (existing.pageInfo.endCursor === incoming.pageInfo.endCursor) {
                // If the endCursor is the same, we assume that the incoming
                // nodes are the same as the existing nodes. This can happen
                // when the same query is executed multiple times in a row.
                // In this case, we return the existing nodes to prevent
                // incorrect cache updates.
                return existing
            }

            return {
                ...incoming,
                nodes: [...existing.nodes, ...incoming.nodes],
            }
        },
    }
}

export const getGraphQLClient = once(async (): Promise<GraphQLClient> => {
    const cache = new InMemoryCache({
        typePolicies: {
            GitCommit: {
                fields: {
                    ancestors: listLikeForwardConnection({
                        keyArgs: args => {
                            // This key function treats an empty path the same as an
                            // omitted path.
                            // keyArgs: ['query', 'path', 'follow', 'after'],
                            const keyArgs: Record<string, any> = {}
                            if (args) {
                                for (const key of ['query', 'path', 'follow', 'after']) {
                                    if (key in args && (key !== 'path' || args[key] !== '')) {
                                        keyArgs[key] = args[key]
                                    }
                                }
                            }
                            return JSON.stringify(keyArgs)
                        },
                    }),
                },
            },
            GitTree: {
                // GitTree object's don't have an ID, but canonicalURL is unique
                keyFields: ['canonicalURL'],
            },
            GitBlob: {
                // GitBlob object's don't have an ID, but canonicalURL is unique
                keyFields: ['canonicalURL'],
            },
            HighlightedFile: {
                // Necessary to cache the highlight results of multiple line
                // highlight requests
                merge: true,
            },
            GitBlobLSIFData: {
                merge: true,
            },
            // Signature is not normalized. Data from multiple requests needs
            // to be merged, not replaced, in order to not lose data.
            Signature: {
                merge: true,
            },
            // Person is not normalized. Data from multiple requests needs
            // to be merged, not replaced, in order to not lose data.
            Person: {
                merge: true,
            },
            RepositoryComparison: {
                fields: {
                    fileDiffs: listLikeForwardConnection({
                        keyArgs: ['paths'],
                    }),
                },
            },
        },
        possibleTypes: {
            TreeEntry: ['GitTree', 'GitBlob'],
        },
    })

    // TODO: Persist data locally after figuring out how to determine user authentication state

    const uri = buildGraphQLUrl({})
    return new ApolloClient({
        connectToDevTools: dev,
        uri,
        cache,
        link: from([
            createHttpLink({
                uri: ({ operationName }) => `${uri}?${operationName}`,
                headers: getHeaders(),
                fetch: customFetch,
            }),
        ]),
    })
})

export async function query<T, V extends OperationVariables = OperationVariables>(
    query: DocumentNode,
    variables?: V,
    options?: Omit<QueryOptions<T, V>, 'query' | 'variables'>
): Promise<T> {
    return (await getGraphQLClient()).query<T, V>({ query, variables, ...options }).then(result => {
        if (result.errors && result.errors.length > 0) {
            throw createAggregateError(result.errors)
        }
        return result.data
    })
}

export type { FetchPolicy }
export { gql }
