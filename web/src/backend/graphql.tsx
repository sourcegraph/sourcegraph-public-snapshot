import { Observable } from 'rxjs'
import { ajax, AjaxResponse } from 'rxjs/ajax'
import { catchError, map } from 'rxjs/operators'
import { normalizeAjaxError } from '../util/errors'
import * as GQL from './graphqlschema'

const graphQLContent = Symbol('graphQLContent')
export interface GraphQLDocument {
    [graphQLContent]: string
}

/**
 * Use this template string tag for all GraphQL queries
 */
export const gql = (template: TemplateStringsArray, ...substitutions: any[]): GraphQLDocument => ({
    [graphQLContent]: String.raw(template, ...substitutions.map(s => s[graphQLContent] || s)),
})

/**
 * Interface for the response result of a GraphQL query
 */
interface QueryResult {
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
function requestGraphQL(request: GraphQLDocument, variables: any = {}): Observable<GQL.IGraphQLResponseRoot> {
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
export function queryGraphQL(query: GraphQLDocument, variables: any = {}): Observable<QueryResult> {
    return requestGraphQL(query, variables) as Observable<QueryResult>
}

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param mutation The GraphQL mutation
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export function mutateGraphQL(mutation: GraphQLDocument, variables: any = {}): Observable<MutationResult> {
    return requestGraphQL(mutation, variables) as Observable<MutationResult>
}
