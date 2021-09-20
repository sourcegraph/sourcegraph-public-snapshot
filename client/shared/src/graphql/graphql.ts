import {
    gql as apolloGql,
    useQuery as useApolloQuery,
    useMutation as useApolloMutation,
    useLazyQuery as useApolloLazyQuery,
    DocumentNode,
    ApolloClient,
    createHttpLink,
    NormalizedCacheObject,
    OperationVariables,
    QueryHookOptions,
    QueryResult,
    MutationHookOptions,
    MutationTuple,
    QueryTuple,
} from '@apollo/client'
import { GraphQLError } from 'graphql'
import { once } from 'lodash'
import { useMemo } from 'react'
import { Observable } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { Omit } from 'utility-types'

import { checkOk } from '../backend/fetch'
import { createAggregateError } from '../util/errors'

import { cache } from './cache'

/**
 * Use this template string tag for all GraphQL queries.
 */
export const gql = (template: TemplateStringsArray, ...substitutions: any[]): string =>
    String.raw(template, ...substitutions)

export interface SuccessGraphQLResult<T> {
    data: T
    errors: undefined
}
export interface ErrorGraphQLResult {
    data: undefined
    errors: readonly GraphQLError[]
}

export type GraphQLResult<T> = SuccessGraphQLResult<T> | ErrorGraphQLResult

/**
 * Guarantees that the GraphQL query resulted in an error.
 */
export function isErrorGraphQLResult<T>(result: GraphQLResult<T>): result is ErrorGraphQLResult {
    return !!(result as ErrorGraphQLResult).errors && (result as ErrorGraphQLResult).errors.length > 0
}

export function dataOrThrowErrors<T>(result: GraphQLResult<T>): T {
    if (isErrorGraphQLResult(result)) {
        throw createAggregateError(result.errors)
    }
    return result.data
}

export const createInvalidGraphQLQueryResponseError = (queryName: string): GraphQLError =>
    new GraphQLError(`Invalid GraphQL response: query ${queryName}`)

export const createInvalidGraphQLMutationResponseError = (queryName: string): GraphQLError =>
    new GraphQLError(`Invalid GraphQL response: mutation ${queryName}`)

export interface GraphQLRequestOptions extends Omit<RequestInit, 'method' | 'body'> {
    baseUrl?: string
}

const GRAPHQL_URI = '/.api/graphql'

/**
 * This function should not be called directly as it does not
 * add the necessary headers to authorize the GraphQL API call.
 * Use `requestGraphQL()` in `client/web/src/backend/graphql.ts` instead.
 */
export function requestGraphQLCommon<T, V = object>({
    request,
    baseUrl,
    variables,
    ...options
}: GraphQLRequestOptions & {
    request: string
    variables?: V
}): Observable<GraphQLResult<T>> {
    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    const apiURL = `${GRAPHQL_URI}${nameMatch ? '?' + nameMatch[1] : ''}`

    return fromFetch<GraphQLResult<T>>(baseUrl ? new URL(apiURL, baseUrl).href : apiURL, {
        ...options,
        method: 'POST',
        body: JSON.stringify({ query: request, variables }),
        selector: response => checkOk(response).json(),
    })
}

interface GetGraphqlClientOptions {
    headers: RequestInit['headers']
    baseUrl?: string
}

export type GraphQLClient = ApolloClient<NormalizedCacheObject>

export const getGraphQLClient = once(
    async (options: GetGraphqlClientOptions): Promise<GraphQLClient> => {
        const { headers, baseUrl } = options
        const uri = baseUrl ? new URL(GRAPHQL_URI, baseUrl).href : GRAPHQL_URI

        const apolloClient = new ApolloClient({
            uri,
            cache,
            defaultOptions: {
                /**
                 * The default `fetchPolicy` is `cache-first`, which returns a cached response
                 * and doesn't trigger cache update. This is undesirable default behavior because
                 * we want to keep our cache updated to avoid confusing the user with stale data.
                 * `cache-and-network` allows us to return a cached result right away and then update
                 * all consumers with the fresh data from the network request.
                 */
                watchQuery: {
                    fetchPolicy: 'cache-and-network',
                },
                /**
                 * `client.query()` returns promise, so it can only resolve one response.
                 * Meaning we cannot return the cached result first and then update it with
                 * the response from the network as it's done in `client.watchQuery()`.
                 * So we always need to make a network request to get data unless another
                 * `fetchPolicy` is specified in the `client.query()` call.
                 */
                query: {
                    fetchPolicy: 'network-only',
                },
            },
            link: createHttpLink({
                uri: ({ operationName }) => `${uri}?${operationName}`,
                headers,
            }),
        })

        return Promise.resolve(apolloClient)
    }
)

type RequestDocument = string | DocumentNode

/**
 * Returns a `DocumentNode` value to support integrations with GraphQL clients that require this.
 *
 * @param document The GraphQL operation payload
 * @returns The created `DocumentNode`
 */
export const getDocumentNode = (document: RequestDocument): DocumentNode => {
    if (typeof document === 'string') {
        return apolloGql(document)
    }
    return document
}

const useDocumentNode = (document: RequestDocument): DocumentNode =>
    useMemo(() => getDocumentNode(document), [document])

/**
 * Send a query to GraphQL and respond to updates.
 * Wrapper around Apollo `useQuery` that supports `DocumentNode` and `string` types.
 *
 * @param query GraphQL operation payload.
 * @param options Operation variables and request configuration
 * @returns GraphQL response
 */
export function useQuery<TData = any, TVariables = OperationVariables>(
    query: RequestDocument,
    options: QueryHookOptions<TData, TVariables>
): QueryResult<TData, TVariables> {
    const documentNode = useDocumentNode(query)
    return useApolloQuery(documentNode, options)
}

export function useLazyQuery<TData = any, TVariables = OperationVariables>(
    query: RequestDocument,
    options: QueryHookOptions<TData, TVariables>
): QueryTuple<TData, TVariables> {
    const documentNode = useDocumentNode(query)
    return useApolloLazyQuery(documentNode, options)
}

/**
 * Send a mutation to GraphQL and respond to updates.
 * Wrapper around Apollo `useMutation` that supports `DocumentNode` and `string` types.
 *
 * @param mutation GraphQL operation payload.
 * @param options Operation variables and request configuration
 * @returns GraphQL response
 */
export function useMutation<TData = any, TVariables = OperationVariables>(
    mutation: RequestDocument,
    options?: MutationHookOptions<TData, TVariables>
): MutationTuple<TData, TVariables> {
    const documentNode = useDocumentNode(mutation)
    return useApolloMutation(documentNode, options)
}
