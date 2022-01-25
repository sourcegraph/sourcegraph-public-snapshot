import { Observable, ReplaySubject } from 'rxjs'
import * as vscode from 'vscode'

import { AuthenticatedUser, currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'

import { requestGraphQLFromVSCode } from './requestGraphQl'

// Update authenticatedUser on accessToken changes
export function observeAuthenticatedUser({
    context,
}: {
    context: vscode.ExtensionContext
}): Observable<AuthenticatedUser | null> {
    const authenticatedUsers = new ReplaySubject<AuthenticatedUser | null>(1)

    function updateAuthenticatedUser(): void {
        requestGraphQLFromVSCode<CurrentAuthStateResult, CurrentAuthStateVariables>(currentAuthStateQuery, {})
            .then(authenticatedUserResult => {
                authenticatedUsers.next(authenticatedUserResult.data ? authenticatedUserResult.data.currentUser : null)
            })
            .catch(error => {
                console.log('core auth error', error)
                // TODO surface error?
                authenticatedUsers.next(null)
            })
    }

    // Initial authenticated user
    updateAuthenticatedUser()

    // Update authenticated user on access token changes
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration(config => {
            if (config.affectsConfiguration('sourcegraph.accessToken')) {
                updateAuthenticatedUser()
            }
        })
    )

    return authenticatedUsers
}
