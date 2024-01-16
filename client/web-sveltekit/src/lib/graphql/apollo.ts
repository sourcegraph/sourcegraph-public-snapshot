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
    type MutationOptions,
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

function listBasedForwardConnection({
    keyArgs,
    cursorName,
}: {
    keyArgs: string[] | false
    cursorName: string
}): FieldPolicy {
    return {
        keyArgs,

        merge(existing, incoming, { args }) {
            if (!args) {
                return incoming
            }

            // args.after and pageInfo.endCursor seem to refer to the index of the
            // last item in the list.
            const nodes = existing ? [...existing.nodes] : []
            const offset = args[cursorName] ? +args[cursorName] : 0
            for (let i = 0; i < incoming.nodes.length; ++i) {
                nodes[offset + i] = incoming.nodes[i]
            }

            let pageInfo = existing?.pageInfo
            if (!pageInfo) {
                pageInfo = incoming.pageInfo
            } else if (pageInfo.endCursor) {
                if (incoming.pageInfo.endCursor && +incoming.pageInfo.endCursor > +pageInfo.endCursor) {
                    pageInfo = incoming.pageInfo
                }
            }

            return {
                ...incoming,
                nodes,
                pageInfo,
            }
        },
        read(existing, options) {
            if (!existing) {
                return existing
            }
            // This is a hack to allow processing `ancestor` in a paginated way as well
            // as in an infinity-scroll kind of way. For infinity scroll we want the
            // whole list to be returned whenever this field is requested (e.g. history panel).
            // For a paginated version we only want to return the n items following the current
            // cursor.
            // Queries who want the pagninated version simply alias the field to `<field>_paginated`
            if (options.field?.alias && /_paginated$/.test(options.field.alias.value)) {
                const from = options.args?.afterCursor ? +options.args.afterCursor : 0
                const to = from + options.args?.first
                const nodes = existing.nodes.slice(from, to)
                // If any of the nodes are missing it means we fetched previous data out-of-band.
                // Return undefined to force Apollo to fetch the requested data
                // [...nodes] is necessary because `.some` skips holes in arrays. [...nodes] makes
                // it so that those holes become `undefined` values instead.
                if (nodes.length === 0 || [...nodes].some(node => !node)) {
                    return undefined
                }
                return {
                    ...existing,
                    nodes,
                    pageInfo: existing.nodes[to]
                        ? {
                              ...existing.pageInfo,
                              endCursor: String(to),
                              hasNextPage: true,
                          }
                        : existing.pageInfo,
                }
            }
            return existing
        },
    }
}

export const getGraphQLClient = once(async (): Promise<GraphQLClient> => {
    const cache = new InMemoryCache({
        typePolicies: {
            GitCommit: {
                fields: {
                    ancestors: listBasedForwardConnection({
                        keyArgs: ['query', 'path', 'follow', 'after'],
                        cursorName: 'afterCursor',
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
            RepositoryComparison: {
                fields: {
                    fileDiffs: listBasedForwardConnection({
                        keyArgs: ['paths'],
                        cursorName: 'after',
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

export async function fromCache<T, V extends OperationVariables = OperationVariables>(
    query: DocumentNode,
    variables?: V,
    options?: Omit<QueryOptions<T, V>, 'query' | 'variables'>
): Promise<T | null> {
    return (await getGraphQLClient()).readQuery<T, V>({ query, variables, ...options })
}

export async function mutation<T, V extends OperationVariables = OperationVariables>(
    mutation: DocumentNode,
    variables?: V,
    options?: Omit<MutationOptions<T, V>, 'query' | 'variables'>
): Promise<T | null | undefined> {
    return (await getGraphQLClient()).mutate<T, V>({ mutation, variables, ...options }).then(result => {
        if (result.errors?.length ?? 0 > 0) {
            throw createAggregateError(result.errors)
        }
        return result.data
    })
}

export type { FetchPolicy }
export { gql }
