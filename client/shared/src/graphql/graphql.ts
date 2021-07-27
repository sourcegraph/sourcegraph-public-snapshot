import {
    gql as apolloGql,
    useQuery as useApolloQuery,
    useMutation as useApolloMutation,
    DocumentNode,
    ApolloClient,
    NormalizedCacheObject,
    OperationVariables,
    QueryHookOptions,
    QueryResult,
    MutationHookOptions,
    MutationTuple,
    ApolloQueryResult,
    FetchResult as ApolloFetchResult,
    split,
    HttpLink,
} from '@apollo/client'
import { BatchHttpLink } from '@apollo/client/link/batch-http'
import { getOperationDefinition } from '@apollo/client/utilities'
import { GraphQLError } from 'graphql'
import { useMemo } from 'react'
import { from, Observable } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { map } from 'rxjs/operators'
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
    return fromFetch(baseUrl ? new URL(apiURL, baseUrl).href : apiURL, {
        ...options,
        method: 'POST',
        body: JSON.stringify({ query: request, variables }),
        selector: response => checkOk(response).json(),
    })
}

/**
 * Map non-conforming GraphQL responses to a GraphQLResult.
 */
export const apolloToGraphQLResult = <T>(response: ApolloQueryResult<T> | ApolloFetchResult<T>): GraphQLResult<T> => {
    if (response.errors) {
        return {
            data: undefined,
            errors: response.errors,
        }
    }

    if (!response.data) {
        // This should never happen unless `errorPolicy` is set to `ignore` when calling `client.mutate`
        // We don't support this configuration through this wrapper function, so OK to error here.
        throw new Error('Invalid response. Neither data nor errors is present in the response.')
    }

    return {
        data: response.data,
        errors: undefined,
    }
}

export interface RequestGraphQLContext {
    batchRequests?: boolean
}

export function requestGraphQLApollo<T, V = object>({
    request,
    variables,
    client,
    context,
}: GraphQLRequestOptions & {
    request: string
    variables?: V
    client: ApolloClient<NormalizedCacheObject>
    context?: RequestGraphQLContext
}): Observable<GraphQLResult<T>> {
    const document = getDocumentNode(request)
    const queryDefinition = getOperationDefinition(document)

    if (!queryDefinition) {
        throw new Error('No GraphQL operation found.')
    }

    if (queryDefinition.operation === 'query') {
        return from(client.query({ query: document, variables, fetchPolicy: 'no-cache', context })).pipe(
            map(apolloToGraphQLResult)
        )
    }

    if (queryDefinition.operation === 'mutation') {
        return from(client.mutate({ mutation: document, variables, fetchPolicy: 'no-cache', context })).pipe(
            map(apolloToGraphQLResult)
        )
    }

    // We don't support the GraphQL subscription operation.
    // If we have a use-case for this in the future, we should use useSubscription from Apollo
    throw new Error(`Unsupported GraphQL operation: ${queryDefinition.operation}`)
}

export const graphQLClient = ({ headers }: { headers: RequestInit['headers'] }): ApolloClient<NormalizedCacheObject> =>
    new ApolloClient({
        uri: GRAPHQL_URI,
        cache,
        link: split(
            operation => operation.getContext().batchRequests,
            new BatchHttpLink({
                uri: ({ operationName }) => `${GRAPHQL_URI}?${operationName}`,
                headers,
                batchMax: 5, // No more than 5 operations per batch
                batchInterval: 20, // Wait no more than 20ms after first batched operation
            }),
            new HttpLink({
                uri: ({ operationName }) => `${GRAPHQL_URI}?${operationName}`,
                headers,
            })
        ),
    })

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
