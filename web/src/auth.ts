import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { tap } from 'rxjs/operators/tap'
import { ReplaySubject } from 'rxjs/ReplaySubject'
import { mutateGraphQL, queryGraphQL } from './backend/graphql'

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
 * fetchCurrentUser can be called to fetch the current user and orgs state from the remote.
 * Emits no items, completes when done.
 */
export function fetchCurrentUser(): Observable<never> {
    return queryGraphQL(`
        query CurrentAuthState {
            root {
                currentUser {
                    auth0ID
                    sourcegraphID
                    username
                    avatarURL
                    email
                    username
                    displayName
                    latestSettings {
                        id
                        contents
                        highlighted
                    }
                    orgs {
                        id
                        name
                    }
                    tags {
                        id
                        name
                    }
                }
            }
        }
    `).pipe(
        tap(result => {
            if (!result.data) {
                throw new Error('invalid response received from graphql endpoint')
            }
            currentUser.next(result.data.root.currentUser)
        }),
        mergeMap(() => [])
    )
}

export function updateUserSettings(lastKnownSettingsID: number | null, contents: string): Observable<void> {
    return mutateGraphQL(
        `
        mutation UpdateUserSettings($lastKnownSettingsID: Int, $contents: String!) {
            updateUserSettings(lastKnownSettingsID: $lastKnownSettingsID, contents: $contents) { }
        }
    `,
        { lastKnownSettingsID, contents }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return
        })
    )
}
