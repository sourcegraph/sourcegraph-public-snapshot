import { Observable } from 'rxjs'
import { GraphQLResult, requestGraphQL as requestGraphQLCommon } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'

const getHeaders = () => ({
    ...window.context.xhrHeaders,
    Accept: 'application/json',
    'Content-Type': 'application/json',
})

/**
 * Does a GraphQL request to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL request (query or mutation)
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export const requestGraphQL = <T extends GQL.IQuery | GQL.IMutation>(
    request: string,
    variables?: {}
): Observable<GraphQLResult<T>> =>
    requestGraphQLCommon({
        request,
        variables,
        headers: getHeaders(),
    })

/**
 * Does a GraphQL query to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param query The GraphQL query
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export const queryGraphQL = (request: string, variables?: {}): Observable<GraphQLResult<GQL.IQuery>> =>
    requestGraphQLCommon({
        request,
        variables,
        headers: getHeaders(),
    })

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param mutation The GraphQL mutation
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export const mutateGraphQL = (request: string, variables?: {}): Observable<GraphQLResult<GQL.IMutation>> =>
    requestGraphQLCommon({
        request,
        variables,
        headers: getHeaders(),
    })
