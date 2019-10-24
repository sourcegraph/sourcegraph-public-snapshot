import { GraphQLResult, requestGraphQL } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Observable } from 'rxjs'
import { Config } from '../../../../shared/src/e2e/config'
import { PlatformContext } from '../../../../shared/src/platform/context'

/**
 * A GraphQL client to be used from regression test scripts.
 */
export interface GraphQLClient {
    /**
     * mimics the `mutateGraphQL` function used by the Sourcegraph backend, but substitutes
     * in CLI-read HTTP headers rather than use the ones in the browser context (which doesn't
     *  exist).
     */
    mutateGraphQL(request: string, variables?: {}): Observable<GraphQLResult<GQL.IMutation>>

    /**
     * mimics the `queryGraphQL` function used by the Sourcegraph backend, but substitutes
     * in CLI-read HTTP headers rather than use the ones in the browser context (which doesn't
     * exist).
     */
    queryGraphQL(request: string, variables?: {}): Observable<GraphQLResult<GQL.IQuery>>

    /**
     * Mimics the {@link PlatformContext#requestGraphQL} function that is injected in shared backend functions,
     * but substitutes in CLI-read HTTP headers rather than use the ones in the browser context (which doesn't
     * exist).
     */
    requestGraphQL: PlatformContext['requestGraphQL']
}

/**
 * Returns a {@link GraphQLClient} for use in regression test scripts.
 */
export const createGraphQLClient = ({
    sourcegraphBaseUrl: baseUrl,
    sudoToken,
    sudoUsername,
}: Pick<Config, 'sourcegraphBaseUrl' | 'sudoToken' | 'sudoUsername'>): GraphQLClient => {
    const headers = {
        Authorization: `token-sudo user="${sudoUsername}",token="${sudoToken}"`,
    }
    return {
        mutateGraphQL: (request, variables) =>
            requestGraphQL({
                request,
                variables,
                headers,
                baseUrl,
            }),
        queryGraphQL: (request, variables) =>
            requestGraphQL({
                request,
                variables,
                headers,
                baseUrl,
            }),
        requestGraphQL: ({ request, variables }) =>
            requestGraphQL({
                request,
                variables,
                headers,
                baseUrl,
            }),
    }
}
