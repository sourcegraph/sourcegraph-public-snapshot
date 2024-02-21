import {
    ApolloClient,
    InMemoryCache,
    createHttpLink,
    from,
    type HttpOptions,
    type NormalizedCacheObject,
} from '@apollo/client/core/index'
import { trimEnd, once } from 'lodash'

import { dev } from '$app/environment'
import { GRAPHQL_URI, checkOk } from '$lib/http-client'

import { getHeaders } from './shared'

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

const customFetch: HttpOptions['fetch'] = (uri, options) => fetch(uri, options).then(checkOk)

export type GraphQLClient = ApolloClient<NormalizedCacheObject>

/**
 * @deprecated Use `getGraphQLClient` from @lib/graphql instead.
 *
 * This is only used for compatibility with APIs that expect an ApolloClient.
 */
export const getGraphQLClient = once((): GraphQLClient => {
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
