import { AuthenticatedUser, currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'

import { requestGraphQLFromVSCode } from './requestGraphQl'

export async function getAuthenticatedUser(): Promise<AuthenticatedUser | null> {
    try {
        const authenticatedUserResult = await requestGraphQLFromVSCode<
            CurrentAuthStateResult,
            CurrentAuthStateVariables
        >(currentAuthStateQuery, {})
        return authenticatedUserResult.data ? authenticatedUserResult.data.currentUser : null
    } catch (error) {
        console.log('core auth error', error)
        // TODO surface error?
        return null
    }
}
