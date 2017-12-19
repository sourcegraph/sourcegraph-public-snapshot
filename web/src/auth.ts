import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { tap } from 'rxjs/operators/tap'
import { ReplaySubject } from 'rxjs/ReplaySubject'
import { mutateGraphQL, queryGraphQL } from './backend/graphql'
import { configurationCascade } from './settings/configuration'

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

export const configurationGQL = `
    configuration {
        defaults { contents }
        subjects {
            __typename
            ... on Org {
                id
                name
            }
            ... on User {
                id
                username
            }
            latestSettings {
                id
            }
        }
        merged {
            contents
            messages
        }
    }`

/**
 * fetchCurrentUser can be called to fetch the current user, orgs, and config
 * state from the remote. Emits no items, completes when done.
 */
export function fetchCurrentUser(): Observable<never> {
    return queryGraphQL(
        `
        query CurrentAuthState {
            currentUser {
                __typename
                auth0ID
                sourcegraphID
                username
                avatarURL
                email
                username
                displayName
                verified
                latestSettings {
                    id
                    configuration {
                        contents
                    }
                }
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
            }
            ${configurationGQL}
        }
    `
    ).pipe(
        tap(({ data, errors }) => {
            if (!data) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            currentUser.next(data.currentUser)
            configurationCascade.next(data.configuration)
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
