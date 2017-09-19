import 'rxjs/add/observable/dom/ajax'
import 'rxjs/add/operator/map'
import { Observable } from 'rxjs/Observable'
import { sourcegraphContext } from '../util/sourcegraphContext'

/**
 * Interface for the response result of a GraphQL query
 */
export interface QueryResult {
    data?: GQL.IQuery
    errors?: GQL.IGraphQLResponseError[]
}

/**
 * Interface for the response result of a GraphQL mutation
 */
export interface MutationResult {
    data?: GQL.IMutation
    errors?: GQL.IGraphQLResponseError[]
}

/**
 * Does a GraphQL request to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL request (query or mutation)
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
function requestGraphQL(request: string, variables: any = {}): Observable<GQL.IGraphQLResponseRoot> {
    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    return Observable.ajax({
        method: 'POST',
        url: '/.api/graphql' + (nameMatch ? '?' + nameMatch[1] : ''),
        headers: {
            'Content-Type': 'application/json',
            ...sourcegraphContext.xhrHeaders
        },
        body: JSON.stringify({ query: request, variables })
    }).map(({ response }) => response)
}

/**
 * Does a GraphQL query to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param query The GraphQL query
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function queryGraphQL(query: string, variables: any = {}): Observable<QueryResult> {
    return requestGraphQL(query, variables) as Observable<QueryResult>
}

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param mutation The GraphQL mutation
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function mutateGraphQL(mutation: string, variables: any = {}): Observable<MutationResult> {
    return requestGraphQL(mutation, variables) as Observable<MutationResult>
}
