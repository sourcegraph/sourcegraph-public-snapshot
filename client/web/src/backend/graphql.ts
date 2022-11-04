import { memoize } from 'lodash'
import { Observable } from 'rxjs'

import { getGraphQLClient, GraphQLResult, requestGraphQLCommon } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'

const getHeaders = (): Headers => {
    const headers = new Headers({
        ...window?.context?.xhrHeaders,
        Accept: 'application/json',
        'Content-Type': 'application/json',
    })
    const searchParameters = new URLSearchParams(window.location.search)
    if (searchParameters.get('trace') === '1') {
        headers.set('X-Sourcegraph-Should-Trace', 'true')
    }
    for (const feature of searchParameters.getAll('feat')) {
        headers.append('X-Sourcegraph-Override-Feature', feature)
    }
    return headers
}

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
        headers: getHeaders(),
    })
)
