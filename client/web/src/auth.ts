import { type Observable, ReplaySubject } from 'rxjs'
import { catchError, map, mergeMap, tap } from 'rxjs/operators'

import { dataOrThrowErrors } from '@sourcegraph/http-client'
import { type AuthenticatedUser as SharedAuthenticatedUser, currentAuthStateQuery } from '@sourcegraph/shared/src/auth'
import type { CurrentAuthStateResult } from '@sourcegraph/shared/src/graphql-operations'

import { requestGraphQL } from './backend/graphql'

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
 * TODO: migrate the `currentAuthStateQuery` to Apollo Client and then use pre-loaded data to initialize the Apollo cache.
 */
export const authenticatedUserValue = (typeof window !== 'undefined' && window.context?.currentUser) || null
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
