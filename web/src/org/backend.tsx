import { Observable } from 'rxjs/Observable'
import { concat } from 'rxjs/operators/concat'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { take } from 'rxjs/operators/take'
import { currentUser, refreshCurrentUser } from '../auth'
import { gql, mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { eventLogger } from '../tracking/eventLogger'

/**
 * Fetches an org by ID
 *
 * @return Observable that emits the org or `null` if it doesn't exist
 */
export function fetchOrg(id: string): Observable<GQL.IOrg | null> {
    return queryGraphQL(
        gql`
            query Org($id: ID!) {
                org(id: $id) {
                    id
                    name
                    slackWebhookURL
                    displayName
                    latestSettings {
                        id
                        configuration {
                            contents
                        }
                    }
                    members {
                        id
                        createdAt
                        user {
                            id
                            authID
                            username
                            email
                            displayName
                            avatarURL
                        }
                    }
                    tags {
                        name
                    }
                }
            }
        `,
        { id }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.org) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.org
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
                gql`
                    mutation createOrg($name: String!, $displayName: String!) {
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
            return refreshCurrentUser().pipe(concat([data.createOrg]))
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
export function inviteUser(email: string, orgID: string): Observable<GQL.IInviteUserResult> {
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
                gql`
                    mutation inviteUser($email: String!, $orgID: ID!) {
                        inviteUser(email: $email, orgID: $orgID) {
                            acceptInviteURL
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
            if (!data || !data.inviteUser || (errors && errors.length > 0)) {
                eventLogger.log('InviteOrgMemberFailed', eventData)
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            eventLogger.log('OrgMemberInvited', eventData)
            return data.inviteUser
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
 */
export function acceptUserInvite(options: AcceptUserInviteOptions): Observable<void> {
    return currentUser.pipe(
        take(1),
        mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in')
            }
            return mutateGraphQL(
                gql`
                    mutation AcceptUserInvite {
                        acceptUserInvite(inviteToken: $inviteToken) { }
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
            return
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
export function removeUserFromOrg(orgID: GQLID, userID: GQLID): Observable<never> {
    return mutateGraphQL(
        gql`
            mutation removeUserFromOrg($userID: ID!, $orgID: ID!) {
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
                        user_id: userID,
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
            return refreshCurrentUser()
        })
    )
}

/**
 * Sends a GraphQL mutation to update an org
 *
 * @param id The ID of the org
 * @param displayName The display name of the org
 * @param slackWebhookURL The Slack webhook URL to send Slack-formatted org actions/updates to
 * @return Observable that emits `undefined`, then completes
 */
export function updateOrg(id: string, displayName: string, slackWebhookURL: string): Observable<void> {
    return currentUser.pipe(
        take(1),
        mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in.')
            }

            const variables = {
                id,
                displayName,
                slackWebhookURL,
            }
            return mutateGraphQL(
                gql`
                    mutation updateOrg($id: ID!, $displayName: String, $slackWebhookURL: String) {
                        updateOrg(id: $id, displayName: $displayName, slackWebhookURL: $slackWebhookURL) {
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
                    org_id: id,
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

export function updateOrgSettings(id: string, lastKnownSettingsID: number | null, contents: string): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation UpdateOrgSettings($id: ID!, $lastKnownSettingsID: Int, $contents: String!) {
                updateOrgSettings(id: $id, lastKnownSettingsID: $lastKnownSettingsID, contents: $contents) { }
            }
        `,
        { id, lastKnownSettingsID, contents }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return
        })
    )
}
