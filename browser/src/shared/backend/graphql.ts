import { Observable } from 'rxjs'
import {
    GraphQLDocument,
    GraphQLRequestOptions,
    GraphQLResult,
    requestGraphQL as requestGraphQLCommon,
} from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { sourcegraphUrl } from '../util/context'
import { getHeaders } from './headers'

const options: GraphQLRequestOptions = {
    headers: getHeaders(),
    baseUrl: sourcegraphUrl,
    requestOptions: {
        crossDomain: true,
        withCredentials: true,
        async: true,
    },
}

/**
 * Does a GraphQL request to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param request The GraphQL request (query or mutation)
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export const requestGraphQL = <T extends GQL.IQuery | GQL.IMutation>(
    request: GraphQLDocument,
    variables?: any
): Observable<GraphQLResult<T>> =>
    requestGraphQLCommon({
        request,
        variables,
        ...options,
    })

/**
 * Does a GraphQL query to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param query The GraphQL query
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export const queryGraphQL = (request: GraphQLDocument, variables?: any): Observable<GraphQLResult<GQL.IQuery>> =>
    requestGraphQLCommon({
        request,
        variables,
        ...options,
    })

/**
 * Does a GraphQL mutation to the Sourcegraph GraphQL API running under `/.api/graphql`
 *
 * @param mutation The GraphQL mutation
 * @param variables A key/value object with variable values
 * @return Observable That emits the result or errors if the HTTP request failed
 */
export const mutateGraphQL = (request: GraphQLDocument, variables?: any): Observable<GraphQLResult<GQL.IMutation>> =>
    requestGraphQLCommon({
        request,
        variables,
        ...options,
    })
