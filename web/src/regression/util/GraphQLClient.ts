import { GraphQLResult, requestGraphQL } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { Observable } from 'rxjs'

/**
 * A GraphQL client to be used from regression test scripts.
 */
export class GraphQLClient {
    constructor(public baseURL: string, public sudoToken: string, public username: string) {}

    /**
     * mimics the `mutateGraphQL` function used by the Sourcegraph backend, but substitutes
     * in CLI-read HTTP headers rather than use the ones in the browser context (which doesn't
     *  exist).
     */
    public mutateGraphQL(request: string, variables?: {}): Observable<GraphQLResult<GQL.IMutation>> {
        return requestGraphQL({
            request,
            variables,
            headers: {
                Authorization: `token-sudo user="${this.username}",token="${this.sudoToken}"`,
            },
            baseUrl: this.baseURL,
        })
    }

    /**
     * mimics the `queryGraphQL` function used by the Sourcegraph backend, but substitutes
     * in CLI-read HTTP headers rather than use the ones in the browser context (which doesn't
     * exist).
     */
    public queryGraphQL(request: string, variables?: {}): Observable<GraphQLResult<GQL.IQuery>> {
        return requestGraphQL({
            request,
            variables,
            headers: {
                Authorization: `token-sudo user="${this.username}",token="${this.sudoToken}"`,
            },
            baseUrl: this.baseURL,
        })
    }
}
