import { Observable, ReplaySubject } from 'rxjs'
import * as vscode from 'vscode'

import { gql } from '@sourcegraph/http-client'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult, CurrentAuthStateVariables } from '@sourcegraph/shared/src/graphql-operations'

import { scretTokenKey } from '../webview/platform/AuthProvider'

import { requestGraphQLFromVSCode } from './requestGraphQl'

// Minimal auth state for the VS Code extension.
// Uses only old fields for backwards compatibility with old GraphQL API versions.
const currentAuthStateQuery = gql`
    query CurrentAuthState {
        currentUser {
            __typename
            id
            databaseID
            username
            avatarURL
            email
            displayName
            siteAdmin
            tags
            url
            settingsURL
            organizations {
                nodes {
                    id
                    name
                    displayName
                    url
                    settingsURL
                }
            }
            session {
                canSignOut
            }
            viewerCanAdminister
            tags
        }
    }
`

// Update authenticatedUser on accessToken changes
export function observeAuthenticatedUser(secretStorage: vscode.SecretStorage): Observable<AuthenticatedUser | null> {
    const authenticatedUsers = new ReplaySubject<AuthenticatedUser | null>(1)

    function updateAuthenticatedUser(): void {
        requestGraphQLFromVSCode<CurrentAuthStateResult, CurrentAuthStateVariables>(currentAuthStateQuery, {})
            .then(async authenticatedUserResult => {
                if (!authenticatedUserResult.data) {
                    await secretStorage.delete(scretTokenKey)
                }
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

    secretStorage.onDidChange(event => {
        if (event.key === scretTokenKey) {
            const token = secretStorage.get(scretTokenKey)
            if (token !== undefined) {
                updateAuthenticatedUser()
            }
        }
    })

    return authenticatedUsers
}
