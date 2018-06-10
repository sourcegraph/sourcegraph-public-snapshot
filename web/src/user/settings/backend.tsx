import { Observable } from 'rxjs'
import { filter, map, mergeMap, tap } from 'rxjs/operators'
import { authRequired } from '../../auth'
import { gql, mutateGraphQL, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { configurationCascade } from '../../settings/configuration'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError } from '../../util/errors'

/**
 * Refreshes the configuration from the server, which propagates throughout the
 * app to all consumers of configuration settings.
 */
export function refreshConfiguration(): Observable<never> {
    return authRequired.pipe(
        filter(authRequired => !authRequired),
        mergeMap(() => fetchConfiguration()),
        tap(result => configurationCascade.next(result)),
        mergeMap(() => [])
    )
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
 * Fetches the settings for the user.
 *
 * @return Observable that emits the settings or `null` if it doesn't exist
 */
export function fetchUserSettings(user: GQL.ID): Observable<GQL.ISettings | null> {
    return queryGraphQL(
        gql`
            query UserSettings($user: ID!) {
                node(id: $user) {
                    ... on User {
                        latestSettings {
                            ...SettingsFields
                        }
                    }
                }
            }
            ${settingsFragment}
        `,
        { user }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node) {
                throw createAggregateError(errors)
            }
            return (data.node as GQL.IUser).latestSettings
        })
    )
}

/**
 * Updates the settings for the user.
 *
 * @return Observable that emits the newly updated settings
 */
export function updateUserSettings(
    user: GQL.ID,
    lastKnownSettingsID: number | null,
    contents: string
): Observable<GQL.ISettings> {
    return mutateGraphQL(
        gql`
            mutation UpdateUserSettings($user: ID!, $lastKnownSettingsID: Int, $contents: String!) {
                updateUserSettings(user: $user, lastKnownSettingsID: $lastKnownSettingsID, contents: $contents) {
                    ...SettingsFields
                }
            }
            ${settingsFragment}
        `,
        { user, lastKnownSettingsID, contents }
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

interface UpdateUserOptions {
    username: string | null
    /** The user's display name */
    displayName: string | null
    /** The user's avatar URL */
    avatarURL: string | null
}

/**
 * Sends a GraphQL mutation to update a user's profile
 */
export function updateUser(user: GQL.ID, args: UpdateUserOptions): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation updateUser($user: ID!, $username: String, $displayName: String, $avatarURL: String) {
                updateUser(user: $user, username: $username, displayName: $displayName, avatarURL: $avatarURL) {
                    alwaysNil
                }
            }
        `,
        { ...args, user }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.updateUser) {
                eventLogger.log('UpdateUserFailed')
                throw createAggregateError(errors)
            }
            eventLogger.log('UserProfileUpdated')
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
export function setUserEmailVerified(user: GQL.ID, email: string, verified: boolean): Observable<void> {
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

/**
 * Log a user action (used to allow site admins on a Sourcegraph instance
 * to see a count of unique users on a daily, weekly, and monthly basis).
 *
 * Not used at all for public/sourcegraph.com usage.
 */
export function logUserEvent(event: GQL.UserEvent): void {
    if (window.context.sourcegraphDotComMode) {
        return
    }
    mutateGraphQL(
        gql`
            mutation logUserEvent($event: UserEvent!, $userCookieID: String!) {
                logUserEvent(event: $event, userCookieID: $userCookieID) {
                    alwaysNil
                }
            }
        `,
        { event, userCookieID: eventLogger.getAnonUserID() }
    )
        .pipe(
            map(({ data, errors }) => {
                if (!data || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return
            })
        )
        .subscribe()
}

refreshConfiguration()
    .toPromise()
    .then(() => void 0, err => console.error(err))
