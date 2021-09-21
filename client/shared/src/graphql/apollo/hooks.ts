import {
    gql as apolloGql,
    useQuery as useApolloQuery,
    useMutation as useApolloMutation,
    useLazyQuery as useApolloLazyQuery,
    DocumentNode,
    OperationVariables,
    QueryHookOptions,
    QueryResult,
    MutationHookOptions,
    MutationTuple,
    QueryTuple,
} from '@apollo/client'
import { useMemo } from 'react'

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
 * Unlike with `useQuery`, when you call `useLazyQuery`, it does not immediately execute its associated query.
 * Wrapper around Apollo `useLazyQuery` that supports `DocumentNode` and `string` types.
 *
 * @param query GraphQL operation payload.
 * @param options Operation variables and request configuration.
 * @returns returns a query function in its result tuple that you call whenever you're ready to execute the query.
 */
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
