import { Observable, ReplaySubject } from 'rxjs'
import { catchError, map, mergeMap, tap } from 'rxjs/operators'
import { dataOrThrowErrors, gql, queryGraphQL } from './backend/graphql'
import * as GQL from './backend/graphqlschema'

/**
 * Always represents the latest
 * state of the currently authenticated user.
 *
 * Note that currentUser is not designed to survive across changes in the
 * currently authenicated user. Sign in, sign out, and account changes are
 * all expected to refresh the app.
 */
export const currentUser = new ReplaySubject<GQL.IUser | null>(1)

/**
 * refreshCurrentUser can be called to fetch the current user, orgs, and config
 * state from the remote. Emits no items, completes when done.
 */
export function refreshCurrentUser(): Observable<never> {
    return queryGraphQL(gql`
        query CurrentAuthState {
            currentUser {
                __typename
                id
                sourcegraphID
                username
                avatarURL
                email
                username
                displayName
                siteAdmin
                tags
                url
                organizations {
                    nodes {
                        id
                        name
                    }
                }
                session {
                    canSignOut
                }
                viewerCanAdminister
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        tap(data => currentUser.next(data.currentUser)),
        catchError(error => {
            currentUser.next(null)
            return []
        }),
        mergeMap(() => [])
    )
}

const initialSiteConfigAuthPublic = window.context.site['auth.public']

/**
 * Whether auth is required to perform any action.
 *
 * If an HTTP request might be triggered by an unauthenticated user on a server with auth.public ==
 * false, the caller must first check authRequired. If authRequired is true, then the component must
 * not initiate the HTTP request. This prevents the browser's devtools console from showing HTTP 401
 * errors, which mislead the user into thinking there is a problem (and make debugging any actual
 * issue much harder).
 */
export const authRequired = currentUser.pipe(map(user => user === null && !initialSiteConfigAuthPublic))

// Populate currentUser.
if (window.context.isAuthenticatedUser) {
    refreshCurrentUser()
        .toPromise()
        .then(() => void 0, err => console.error(err))
} else {
    currentUser.next(null)
}
