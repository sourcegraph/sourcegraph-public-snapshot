import { Observable, ReplaySubject } from 'rxjs'
import { catchError, map, mergeMap, tap } from 'rxjs/operators'

import { dataOrThrowErrors } from '@sourcegraph/http-client'
import { AuthenticatedUser as SharedAuthenticatedUser, currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import { CurrentAuthStateResult } from '@sourcegraph/shared/src/graphql-operations'

import { requestGraphQL } from './backend/graphql'
import { JsContextCurrentUser } from './jscontext'

/**
 * Always represents the latest state of the currently authenticated user.
 *
 * Note that authenticatedUser is not designed to survive across changes in the currently authenticated user. Sign
 * in, sign out, and account changes all require a full-page reload in the browser to take effect.
 */
export const authenticatedUser = new ReplaySubject<AuthenticatedUser | null>(1)

/**
 * Represent the current user info on the initial application load. Instead waiting for the `currentAuthStateQuery` query
 * we use the value provided us from the server. Subsequent updates are received via `authenticatedUser` subject.
 *
 * TODO: migrate the `currentAuthStateQuery` to Apollo Client and then user pre-loaded to initialize Apollo cache.
 */
export const authenticatedUserValue = jsContextCurrentUserToAuthenticatedUser(window.context.CurrentUser)
authenticatedUser.next(authenticatedUserValue)

export type AuthenticatedUser = SharedAuthenticatedUser

/**
 * Fetches the current user, orgs, and config state from the remote. Emits no items, completes when done.
 */
export function refreshAuthenticatedUser(): Observable<never> {
    return requestGraphQL<CurrentAuthStateResult>(currentAuthStateQuery).pipe(
        map(dataOrThrowErrors),
        tap(data => authenticatedUser.next(data.currentUser)),
        catchError(() => {
            authenticatedUser.next(null)
            return []
        }),
        mergeMap(() => [])
    )
}

/**
 * Whether auth is required to perform any action.
 *
 * If an HTTP request might be triggered by an unauthenticated user on a server that is not Sourcegraph.com
 * the caller must first check authRequired. If authRequired is true, then the component must
 * not initiate the HTTP request. This prevents the browser's devtools console from showing HTTP 401
 * errors, which mislead the user into thinking there is a problem (and make debugging any actual
 * issue much harder).
 */
export const authRequired = authenticatedUser.pipe(
    map(user => user === null && typeof window !== 'undefined' && !window.context?.sourcegraphDotComMode)
)

/**
 * Convert `JsContextCurrentUser` to `AuthenticatedUser` received from the GraphQL query `currentAuthStateQuery`.
 * Using pre-loaded user information allows us to skip this query on the inital render of the application.
 */
function jsContextCurrentUserToAuthenticatedUser(user: JsContextCurrentUser): AuthenticatedUser | null {
    if (!user) {
        return null
    }

    return {
        __typename: 'User',
        id: user.ID,
        databaseID: user.DatabaseID,
        username: user.Username,
        avatarURL: user.AvatarURL,
        displayName: user.DisplayName,
        siteAdmin: user.SiteAdmin,
        tags: user.Tags,
        url: user.URL,
        settingsURL: user.SettingsURL,
        organizations: {
            __typename: 'OrgConnection',
            nodes: user.Organizations.map(org => ({
                __typename: 'Org',
                id: org.ID,
                name: org.Name,
                displayName: org.DisplayName,
                url: org.URL,
                settingsURL: org.SettingsURL,
            })),
        },
        session: {
            canSignOut: user.CanSignOut,
        },
        viewerCanAdminister: user.ViewerCanAdminister,
        tosAccepted: user.TosAccepted,
        searchable: user.Searchable,
        emails: user.Emails.map(emailItem => ({
            email: emailItem.Email,
            verified: emailItem.Verified,
            isPrimary: emailItem.IsPrimary,
        })),
        latestSettings: {
            id: user.LatestSettings.ID,
            contents: user.LatestSettings.Contents,
        },
    }
}
