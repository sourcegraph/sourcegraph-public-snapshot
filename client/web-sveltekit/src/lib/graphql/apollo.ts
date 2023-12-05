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

export const getGraphQLClient = once(async (): Promise<GraphQLClient> => {
    const cache = new InMemoryCache({
        typePolicies: {
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
