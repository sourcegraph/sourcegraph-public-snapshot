import {
    gql as apolloGql,
    useQuery as useApolloQuery,
    useMutation as useApolloMutation,
    useLazyQuery as useApolloLazyQuery,
    DocumentNode,
    OperationVariables,
    QueryHookOptions as ApolloQueryHookOptions,
    QueryResult,
    MutationHookOptions as ApolloMutationHookOptions,
    MutationTuple,
    QueryTuple,
} from '@apollo/client'
import { useMemo } from 'react'

import { ApolloContext } from '../types'

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

export interface QueryHookOptions<TData = any, TVariables = OperationVariables>
    extends Omit<ApolloQueryHookOptions<TData, TVariables>, 'context'> {
    /**
     * Shared context information for apollo client. Since internal apollo
     * types have context as Record<string, any> we have to set this type
     * directly.
     */
    context?: ApolloContext
}

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

interface MutationHookOptions<TData = any, TVariables = OperationVariables>
    extends Omit<ApolloMutationHookOptions<TData, TVariables>, 'context'> {
    /**
     * Shared context information for apollo client. Since internal apollo
     * types have context as Record<string, any> we have to set this type
     * directly.
     */
    context?: ApolloContext
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
