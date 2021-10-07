import { memoize } from 'lodash'
import { Observable } from 'rxjs'

import { getGraphQLClient, GraphQLResult, requestGraphQLCommon } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

const getHeaders = (): { [header: string]: string } => ({
    ...window?.context?.xhrHeaders,
    Accept: 'application/json',
    'Content-Type': 'application/json',
    'X-Sourcegraph-Should-Trace': new URLSearchParams(window.location.search).get('trace') || 'false',
})

/**
 * Does a GraphQL request to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL request (query or mutation)
 * @param variables A key/value object with variable values
 * @returns Observable That emits the result or errors if the HTTP request failed
 * @template TResult The type of the query result (import from our auto-generated types).
 * @template TVariables The type of the query input variables (import from our auto-generated types).
 */
export const requestGraphQL = <TResult, TVariables = object>(
    request: string,
    variables?: TVariables
): Observable<GraphQLResult<TResult>> =>
    requestGraphQLCommon({
        request,
        variables,
        headers: getHeaders(),
    })

/**
 * Does a GraphQL query to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL query
 * @param variables A key/value object with variable values
 * @returns Observable That emits the result or errors if the HTTP request failed
 *
 * @deprecated Prefer using `requestGraphQL()` and passing auto-generated query types as type parameters.
 */
export const queryGraphQL = (request: string, variables?: {}): Observable<GraphQLResult<GQL.IQuery>> =>
    requestGraphQLCommon<GQL.IQuery>({
        request,
        variables,
        headers: getHeaders(),
    })

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL mutation
 * @param variables A key/value object with variable values
 * @returns Observable That emits the result or errors if the HTTP request failed
 *
 * @deprecated Prefer using `requestGraphQL()` and passing auto-generated query types as type parameters.
 */
export const mutateGraphQL = (request: string, variables?: {}): Observable<GraphQLResult<GQL.IMutation>> =>
    requestGraphQLCommon<GQL.IMutation>({
        request,
        variables,
        headers: getHeaders(),
    })

/**
 * Memoized Apollo Client getter. It should be executed once to restore the cache from the local storage.
 * After that, the same instance should be used by all consumers.
 */
export const getWebGraphQLClient = memoize(() =>
    getGraphQLClient({
        isAuthenticated: window.context.isAuthenticatedUser,
        headers: {
            ...window?.context?.xhrHeaders,
            'X-Sourcegraph-Should-Trace': new URLSearchParams(window.location.search).get('trace') || 'false',
            // Note: Do not use getHeaders() here due to a bug that Apollo has duplicating headers with different letter casing. Issue to fix: https://github.com/apollographql/apollo-client/issues/8447
        },
    })
)
