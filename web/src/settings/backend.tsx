import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { take } from 'rxjs/operators/take'
import { tap } from 'rxjs/operators/tap'
import { currentUser } from '../auth'
import { gql, mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { eventLogger } from '../tracking/eventLogger'
import { configurationCascade } from './configuration'

/**
 * Refreshes the configuration from the server, which propagates throughout the
 * app to all consumers of configuration settings.
 */
export function refreshConfiguration(): Observable<never> {
    return fetchConfiguration().pipe(tap(result => configurationCascade.next(result)), mergeMap(() => []))
}

const configurationCascadeFragment = gql`
    fragment ConfigurationCascadeFields on ConfigurationCascade {
        defaults {
            contents
        }
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
    }
`

/**
 * Fetches the configuration from the server. Callers should use refreshConfiguration
 * instead of calling this function, to ensure that the result is propagated consistently
 * throughout the app instead of only being returned to the caller.
 *
 * @return Observable that emits the configuration
 */
function fetchConfiguration(): Observable<GQL.IConfigurationCascade> {
    return queryGraphQL(gql`
        query Configuration() {
            configuration {
                ...ConfigurationCascadeFields
            }
        }
        ${configurationCascadeFragment}
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.configuration) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.configuration
        })
    )
}

const settingsFragment = gql`
    fragment SettingsFields on Settings {
        id
        configuration {
            contents
        }
    }
`

/**
 * Fetches the settings for the current user.
 *
 * @return Observable that emits the settings or `null` if it doesn't exist
 */
export function fetchUserSettings(): Observable<GQL.ISettings | null> {
    return queryGraphQL(
        gql`
            query CurrentUserSettings() {
                currentUser {
                    latestSettings {
                        ...SettingsFields
                    }
                }
            }
            ${settingsFragment}
        `
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.currentUser) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.currentUser.latestSettings
        })
    )
}

/**
 * Updates the settings for the current user.
 *
 * @return Observable that emits the newly updated settings
 */
export function updateUserSettings(lastKnownSettingsID: number | null, contents: string): Observable<GQL.ISettings> {
    return mutateGraphQL(
        gql`
            mutation UpdateUserSettings($lastKnownSettingsID: Int, $contents: String!) {
                updateUserSettings(lastKnownSettingsID: $lastKnownSettingsID, contents: $contents) {
                    ...SettingsFields
                }
            }
            ${settingsFragment}
        `,
        { lastKnownSettingsID, contents }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0) || !data.updateUserSettings) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.updateUserSettings
        })
    )
}

export interface UpdateUserOptions {
    username: string
    /** The user's display name */
    displayName: string
    /** The user's avatar URL */
    avatarUrl?: string
}

/**
 * Sends a GraphQL mutation to update a user's profile
 */
export function updateUser(options: UpdateUserOptions): Observable<GQL.IUser> {
    return currentUser.pipe(
        take(1),
        mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in.')
            }

            const variables = {
                ...options,
                avatarUrl: options.avatarUrl || user.avatarURL,
            }
            return mutateGraphQL(
                gql`
                    mutation updateUser($username: String!, $displayName: String!, $avatarURL: String) {
                        updateUser(username: $username, displayName: $displayName, avatarURL: $avatarUrl) {
                            authID
                            sourcegraphID
                            username
                        }
                    }
                `,
                variables
            )
        }),
        map(({ data, errors }) => {
            if (!data || !data.updateUser) {
                eventLogger.log('UpdateUserFailed')
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            eventLogger.log('UserProfileUpdated', {
                auth: {
                    user: {
                        id: data.updateUser.sourcegraphID,
                        auth_id: data.updateUser.authID,
                        username: data.updateUser.username,
                        display_name: options.displayName,
                        avatar_url: options.avatarUrl,
                    },
                },
            })
            return data.updateUser
        })
    )
}

export function logUserEvent(event: GQL.IUserEventEnum): Observable<void> {
    if (!currentUser) {
        throw new Error('User must be signed in')
    }
    return mutateGraphQL(
        gql`
            mutation logUserEvent {
                logUserEvent(event: $event) {
                    alwaysNil
                }
            }
        `,
        { event }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return
        })
    )
}

refreshConfiguration()
    .toPromise()
    .then(() => void 0, err => console.error(err))
