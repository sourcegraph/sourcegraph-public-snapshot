import { Observable } from 'rxjs/Observable'
import { ajax } from 'rxjs/observable/dom/ajax'
import { map } from 'rxjs/operators/map'

export interface GraphQL {
    /** Don't set this yourself, use the template string tag */
    __graphQL: string
}

/**
 * Use this template string tag for all GraphQL queries
 */
export const gql = (template: TemplateStringsArray, ...substitutions: any[]): GraphQL => ({
    __graphQL: String.raw(template, ...substitutions),
})

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
function requestGraphQL(request: GraphQL, variables: any = {}): Observable<GQL.IGraphQLResponseRoot> {
    const nameMatch = request.__graphQL.match(/^\s*(?:query|mutation)\s+(\w+)/)
    return ajax({
        method: 'POST',
        url: '/.api/graphql' + (nameMatch ? '?' + nameMatch[1] : ''),
        headers: {
            'Content-Type': 'application/json',
            ...window.context.xhrHeaders,
        },
        body: JSON.stringify({ query: request.__graphQL, variables }),
    }).pipe(map(({ response }) => response))
}

/**
 * Does a GraphQL query to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param query The GraphQL query
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function queryGraphQL(query: GraphQL, variables: any = {}): Observable<QueryResult> {
    return requestGraphQL(query, variables) as Observable<QueryResult>
}

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param mutation The GraphQL mutation
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function mutateGraphQL(mutation: GraphQL, variables: any = {}): Observable<MutationResult> {
    return requestGraphQL(mutation, variables) as Observable<MutationResult>
}
