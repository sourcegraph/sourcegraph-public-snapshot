import { Observable } from 'rxjs'
import { ajax, AjaxResponse } from 'rxjs/ajax'
import { catchError, map } from 'rxjs/operators'
import { graphQLContent, GraphQLDocument, GraphQLResult } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { normalizeAjaxError } from '../../../shared/src/util/errors'

/**
 * Does a GraphQL request to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL request (query or mutation)
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function requestGraphQL<R extends GQL.IGraphQLResponseRoot>(
    request: GraphQLDocument,
    variables: any = {}
): Observable<R> {
    const nameMatch = request[graphQLContent].match(/^\s*(?:query|mutation)\s+(\w+)/)
    return ajax({
        method: 'POST',
        url: '/.api/graphql' + (nameMatch ? '?' + nameMatch[1] : ''),
        headers: {
            ...window.context.xhrHeaders,
            Accept: 'application/json',
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ query: request[graphQLContent], variables }),
    }).pipe(
        catchError<AjaxResponse, never>(err => {
            normalizeAjaxError(err)
            throw err
        }),
        map(({ response }) => response)
    )
}

/**
 * Does a GraphQL query to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param query The GraphQL query
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function queryGraphQL(query: GraphQLDocument, variables: any = {}): Observable<GraphQLResult<GQL.IQuery>> {
    return requestGraphQL(query, variables)
}

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param mutation The GraphQL mutation
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function mutateGraphQL(
    mutation: GraphQLDocument,
    variables: any = {}
): Observable<GraphQLResult<GQL.IMutation>> {
    return requestGraphQL(mutation, variables)
}
