import { Observable, ReplaySubject } from 'rxjs'
import { catchError, map, mergeMap, tap } from 'rxjs/operators'
import { gql, queryGraphQL } from './backend/graphql'
import * as GQL from './backend/graphqlschema'
import { createAggregateError } from './util/errors'

/**
 * Always represents the latest
 * state of the currently authenticated user.
 *
 * Unlike sourcegraphContext.user, the global currentUser object contains
 * locally mutable properties such as email, displayName, and avatarUrl, all
 * of which are expected to change over the course of a user's session.
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
                orgs {
                    id
                    name
                    tags {
                        name
                    }
                }
                tags {
                    id
                    name
                }
                session {
                    canSignOut
                }
            }
        }
    `).pipe(
        tap(({ data, errors }) => {
            if (!data) {
                throw createAggregateError(errors)
            }
            currentUser.next(data.currentUser)
        }),
        catchError(error => {
            currentUser.next(null)
            return []
        }),
        mergeMap(() => [])
    )
}

/** Whether auth is required to perform any action. */
export const authRequired = currentUser.pipe(map(user => user === null && !window.context.site['auth.public']))

// Populate currentUser synchronously at page load time if possible.
if (!window.context.user) {
    currentUser.next(null)
}

refreshCurrentUser()
    .toPromise()
    .then(() => void 0, err => console.error(err))
