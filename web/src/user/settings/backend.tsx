import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { take } from 'rxjs/operators/take'
import { tap } from 'rxjs/operators/tap'
import { currentUser } from '../../auth'
import { gql, mutateGraphQL, queryGraphQL } from '../../backend/graphql'
import { configurationCascade } from '../../settings/configuration'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError } from '../../util/errors'

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
                configuration {
                    contents
                }
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
        query Configuration {
            configuration {
                ...ConfigurationCascadeFields
            }
        }
        ${configurationCascadeFragment}
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.configuration) {
                throw createAggregateError(errors)
            }
            return data.configuration
        })
    )
}

export const settingsFragment = gql`
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
                throw createAggregateError(errors)
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
                throw createAggregateError(errors)
            }
            refreshConfiguration().subscribe()
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
                        updateUser(username: $username, displayName: $displayName, avatarURL: $avatarURL) {
                            id
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
                throw createAggregateError(errors)
            }
            eventLogger.log('UserProfileUpdated', {
                auth: {
                    user: {
                        id: data.updateUser.sourcegraphID,
                        external_id: data.updateUser.externalID,
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

export function updatePassword(args: { oldPassword: string; newPassword: string }): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation updatePassword($oldPassword: String!, $newPassword: String!) {
                updatePassword(oldPassword: $oldPassword, newPassword: $newPassword) {
                    alwaysNil
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.updatePassword) {
                eventLogger.log('UpdatePasswordFailed')
                throw createAggregateError(errors)
            }
            eventLogger.log('PasswordUpdated')
        })
    )
}

/**
 * Set the verification state for a user email address.
 *
 * @param user the user's GraphQL ID
 * @param email the email address to edit
 * @param verified the new verification state for the user email
 */
export function setUserEmailVerified(user: GQLID, email: string, verified: boolean): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation SetUserEmailVerified($user: ID!, $email: String!, $verified: Boolean!) {
                setUserEmailVerified(user: $user, email: $email, verified: $verified) {
                    alwaysNil
                }
            }
        `,
        { user, email, verified }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
        })
    )
}

export function logUserEvent(event: GQL.IUserEventEnum): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation logUserEvent($event: UserEvent!, $userCookieID: String!) {
                logUserEvent(event: $event, userCookieID: $userCookieID) {
                    alwaysNil
                }
            }
        `,
        { event, userCookieID: eventLogger.uniqueUserCookieID() }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return
        })
    )
}

refreshConfiguration()
    .toPromise()
    .then(() => void 0, err => console.error(err))
