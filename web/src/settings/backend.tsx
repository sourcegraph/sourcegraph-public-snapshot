import { Observable } from 'rxjs/Observable'
import { concat } from 'rxjs/operators/concat'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { take } from 'rxjs/operators/take'
import { currentUser, fetchCurrentUser } from '../auth'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { eventLogger } from '../tracking/eventLogger'

/**
 * Fetches an org by ID
 *
 * @return Observable that emits the org or `null` if it doesn't exist
 */
export function fetchOrg(id: number): Observable<GQL.IOrg | null> {
    return queryGraphQL(
        `
        query Org($id: Int!) {
            root {
                org(id: $id) {
                    id
                    name
                    slackWebhookURL
                    displayName
                    latestSettings {
                        id
                        configuration {
                            contents
                            highlighted
                        }
                    }
                    members {
                        id
                        createdAt
                        user {
                            auth0ID
                            username
                            email
                            displayName
                            avatarURL
                        }
                    }
                }
            }
        }
    `,
        { id }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.root || !data.root.org) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.root.org
        })
    )
}

export interface CreateOrgOptions {
    /** The name of the org */
    name: string
    /** The user's display name (e.g. full name) in the org profile */
    displayName: string
}

/**
 * Sends a GraphQL mutation to create an org and returns an Observable that emits the new org, then completes
 */
export function createOrg(options: CreateOrgOptions): Observable<GQL.IOrg> {
    return currentUser.pipe(
        take(1),
        mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in.')
            }

            return mutateGraphQL(
                `
                mutation createOrg(
                    $name: String!,
                    $displayName: String!
                ) {
                    createOrg(name: $name, displayName: $displayName) {
                        id
                        name
                    }
                }
            `,
                options
            )
        }),
        mergeMap(({ data, errors }) => {
            if (!data || !data.createOrg) {
                eventLogger.log('NewOrgFailed')
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            eventLogger.log('NewOrgCreated', {
                organization: {
                    org_id: data.createOrg.id,
                    org_name: data.createOrg.name,
                },
            })
            return fetchCurrentUser().pipe(concat([data.createOrg]))
        })
    )
}

export interface CreateUserOptions {
    /** The user's username */
    username: string
    /** The user's display name */
    displayName: string
}

/**
 * Sends a GraphQL mutation to create a user and returns an Observable that emits the new user, then completes
 */
export function createUser(options: CreateUserOptions): Observable<GQL.IUser> {
    return currentUser.pipe(
        take(1),
        mergeMap(user => {
            // This API is for user data backfill. You must be an authenticated user
            // to write a row to the users db; we use the authenticated actor to
            // fill auth0_id and email columns.
            if (!user) {
                throw new Error('User must be signed in.')
            }

            const variables = {
                ...options,
                avatarUrl: user.avatarURL,
            }
            return mutateGraphQL(
                `
                mutation createUser(
                    $username: String!,
                    $displayName: String!,
                    $avatarURL: String
                ) {
                    createUser(username: $username, displayName: $displayName, avatarURL: $avatarUrl) {
                        auth0ID
                        sourcegraphID
                        username
                    }
                }
            `,
                variables
            )
        }),
        map(({ data, errors }) => {
            if (!data || !data.createUser) {
                eventLogger.log('NewUserFailed')
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            eventLogger.log('NewUserCreated', {
                auth: {
                    user: {
                        id: data.createUser.sourcegraphID,
                        auth0_id: data.createUser.auth0ID,
                        username: data.createUser.username,
                        display_name: options.displayName,
                    },
                },
            })
            return data.createUser
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
                `
                mutation updateUser(
                    $username: String!,
                    $displayName: String!,
                    $avatarURL: String
                ) {
                    updateUser(username: $username, displayName: $displayName, avatarURL: $avatarUrl) {
                        auth0ID
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
                        auth0_id: data.updateUser.auth0ID,
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

/**
 * Sends a GraphQL mutation to invite a user to an org
 *
 * @param email The email to send the invitation to
 * @param orgID The ID of the org
 * @return Observable that emits `undefined`, then completes
 */
export function inviteUser(email: string, orgID: number): Observable<void> {
    return currentUser.pipe(
        take(1),
        mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in.')
            }

            const variables = {
                email,
                orgID,
            }
            return mutateGraphQL(
                `
                mutation inviteUser($email: String!, $orgID: Int!) {
                    inviteUser(email: $email, orgID: $orgID) {
                        alwaysNil
                    }
                }
            `,
                variables
            )
        }),
        map(({ data, errors }) => {
            const eventData = {
                organization: {
                    invite: {
                        user_email: email,
                    },
                    org_id: orgID,
                },
            }
            if (!data || (errors && errors.length > 0)) {
                eventLogger.log('InviteOrgMemberFailed', eventData)
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            eventLogger.log('OrgMemberInvited', eventData)
            return
        })
    )
    // For now, no need to re-fetch auth state after this fetch completes. The
    // inviteUser mutation only sends an email, it does not update current user
    // or org state.
}

export interface AcceptUserInviteOptions {
    /** The JWT */
    inviteToken: string
}

/**
 * Sends a GraphQL mutation to accept an invitation to an org
 *
 * @return An Observable that does not emit items and completes when done
 */
export function acceptUserInvite(options: AcceptUserInviteOptions): Observable<GQL.IOrgInviteStatus> {
    return currentUser.pipe(
        take(1),
        mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in')
            }
            return mutateGraphQL(
                `
                mutation AcceptUserInvite {
                    acceptUserInvite(
                        inviteToken: $inviteToken
                    ) {
                        emailVerified
                    }
                }
            `,
                options
            )
        }),
        map(({ data, errors }) => {
            if (!data || !data.acceptUserInvite) {
                eventLogger.log('AcceptInviteFailed')
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.acceptUserInvite
        })
    )
}

/**
 * Sends a GraphQL mutation to remove a user from an org
 *
 * @param orgID The org's ID
 * @param userID The user's ID to remove
 * @return An Observable that does emits `undefined` when done, then completes
 */
export function removeUserFromOrg(orgID: number, userID: string): Observable<never> {
    return mutateGraphQL(
        `
        mutation removeUserFromOrg {
            removeUserFromOrg(userID: $userID, orgID: $orgID) {
                alwaysNil
            }
        }
    `,
        {
            userID,
            orgID,
        }
    ).pipe(
        mergeMap(({ data, errors }) => {
            const eventData = {
                organization: {
                    remove: {
                        auth0_id: userID,
                    },
                    org_id: orgID,
                },
            }
            if (errors && errors.length > 0) {
                eventLogger.log('RemoveOrgMemberFailed', eventData)
                throw Object.assign(new Error(errors.map(e => e.message).join('\n')), { errors })
            }
            eventLogger.log('OrgMemberRemoved', eventData)
            // Reload user data
            return fetchCurrentUser()
        })
    )
}

/**
 * Sends a GraphQL mutation to update an org
 *
 * @param orgID The ID of the org
 * @param displayName The display name of the org
 * @param slackWebhookURL The Slack webhook URL to send Slack-formatted org actions/updates to
 * @return Observable that emits `undefined`, then completes
 */
export function updateOrg(orgID: number, displayName: string, slackWebhookURL: string): Observable<void> {
    return currentUser.pipe(
        take(1),
        mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in.')
            }

            const variables = {
                orgID,
                displayName,
                slackWebhookURL,
            }
            return mutateGraphQL(
                `
                mutation updateOrg($orgID: Int!, $displayName: String, $slackWebhookURL: String) {
                    updateOrg(orgID: $orgID, displayName: $displayName, slackWebhookURL: $slackWebhookURL) {
                        id
                    }
                }
            `,
                variables
            )
        }),
        map(({ data, errors }) => {
            const eventData = {
                organization: {
                    update: {
                        display_name: displayName,
                        slack_webhook_url: slackWebhookURL,
                    },
                    org_id: orgID,
                },
            }
            if (!data || (errors && errors.length > 0)) {
                eventLogger.log('UpdateOrgSettingsFailed', eventData)
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            eventLogger.log('OrgSettingsUpdated', eventData)
            return
        })
    )
}

export function updateOrgSettings(
    orgID: number,
    lastKnownSettingsID: number | null,
    contents: string
): Observable<void> {
    return mutateGraphQL(
        `
        mutation UpdateOrgSettings($orgID: Int!, $lastKnownSettingsID: Int, $contents: String!) {
            updateOrgSettings(orgID: $orgID, lastKnownSettingsID: $lastKnownSettingsID, contents: $contents) { }
        }
    `,
        { orgID, lastKnownSettingsID, contents }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return
        })
    )
}
