import { requestGraphQLCommon } from '@sourcegraph/http-client'
import { AuthenticatedUser, currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'

export async function getAuthenticatedUser(
    instanceURL: string,
    accessToken: string | null
): Promise<AuthenticatedUser | null> {
    if (!instanceURL) {
        return null
    }

    const result = await requestGraphQLCommon<CurrentAuthStateResult, CurrentAuthStateVariables>({
        request: currentAuthStateQuery,
        variables: {},
        baseUrl: instanceURL,
        headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
            'X-Sourcegraph-Should-Trace': new URLSearchParams(window.location.search).get('trace') || 'false',
            ...(accessToken && { Authorization: `token ${accessToken}` }),
        },
    }).toPromise()

    return result.data?.currentUser ?? null
}
