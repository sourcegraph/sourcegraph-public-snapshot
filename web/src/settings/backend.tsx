import 'rxjs/add/operator/concat'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/take'
import { Observable } from 'rxjs/Observable'
import { currentUser, fetchCurrentUser } from '../auth'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { events } from '../tracking/events'

/**
 * Fetches an org by ID
 *
 * @return Observable that emits the org or `null` if it doesn't exist
 */
export function fetchOrg(id: number): Observable<GQL.IOrg | null> {
    return queryGraphQL(`
        query Org($id: Int!) {
            root {
                org(id: $id) {
                    id
                    name
                    members {
                        id
                        userID
                        username
                        email
                        displayName
                        avatarURL
                        createdAt
                    }
                }
            }
        }
    `, { id })
        .map(({ data, errors }) => {
            if (!data || !data.root || !data.root.org) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.root.org
        })
}

export interface CreateOrgOptions {
    /** The name of the org */
    name: string
    /** The user's new username in the org profile */
    username: string
    /** The user's display name (e.g. full name) in the org profile */
    displayName: string
    /** The user's email in the org profile */
    email: string
}

/**
 * Sends a GraphQL mutation to create an org and returns an Observable that emits the new org, then completes
 */
export function createOrg(options: CreateOrgOptions): Observable<GQL.IOrg> {
    return currentUser
        .take(1)
        .mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in.')
            }

            const variables = {
                ...options,
                avatarUrl: user.avatarURL
            }
            return mutateGraphQL(`
                mutation createOrg(
                    $name: String!,
                    $username: String!,
                    $email: String!
                    $displayName: String!,
                    $avatarUrl: String!
                ) {
                    createOrg(name: $name, username: $username, email: $email, displayName: $displayName, avatarUrl: $avatarUrl) {
                        id
                        name
                    }
                }
            `, variables)
        })
        .mergeMap(({ data, errors }) => {
            if (!data || !data.createOrg) {
                events.NewOrgFailed.log()
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            events.NewOrgCreated.log({
                organization: {
                    org_id: data.createOrg.id,
                    org_name: data.createOrg.name
                }
            })
            return fetchCurrentUser().concat([data.createOrg])
        })
}

/**
 * Sends a GraphQL mutation to invite a user to an org
 *
 * @param email The email to send the invitation to
 * @param orgID The ID of the org
 * @return Observable that emits `undefined`, then completes
 */
export function inviteUser(email: string, orgID: number): Observable<void> {
    return currentUser
        .take(1)
        .mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in.')
            }

            const variables = {
                email,
                orgID
            }
            return mutateGraphQL(`
                mutation inviteUser($email: String!, $orgID: Int!) {
                    inviteUser(email: $email, orgID: $orgID) {
                        alwaysNil
                    }
                }
            `, variables)
        })
        .map(({ data, errors }) => {
            const eventData = {
                organization: {
                    invite: {
                        user_email: email
                    },
                    org_id: orgID
                }
            }
            if (!data || (errors && errors.length > 0)) {
                events.InviteOrgMemberFailed.log(eventData)
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            events.OrgMemberInvited.log(eventData)
            return
        })
    // For now, no need to re-fetch auth state after this fetch completes. The
    // inviteUser mutation only sends an email, it does not update current user
    // or org state.
}

export interface AcceptUserInviteOptions {
    /** The JWT */
    inviteToken: string
    /** The user's new username in the org profile */
    username: string
    /** The user's email in the org profile */
    email: string
    /** The user's display name (e.g. full name) in the org profile */
    displayName: string
}

/**
 * Sends a GraphQL mutation to accept an invitation to an org
 *
 * @return An Observable that does not emit items and completes when done
 */
export function acceptUserInvite(options: AcceptUserInviteOptions): Observable<GQL.IOrgMember> {
    return currentUser
        .take(1)
        .mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in')
            }
            return mutateGraphQL(`
                mutation AcceptUserInvite {
                    acceptUserInvite(
                        inviteToken: $inviteToken,
                        username: $username,
                        email: $email,
                        displayName: $displayName,
                        avatarUrl: $avatarURL
                    ) {
                        org {
                            name
                        }
                    }
                }
            `, {
                ...options,
                avatarURL: user.avatarURL
            })
        })
        .map(({ data, errors }) => {
            if (!data || !data.acceptUserInvite) {
                events.AcceptInviteFailed.log()
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            events.InviteAccepted.log({
                organization: {
                    invite: {
                        user_email: data.acceptUserInvite.email,
                        user_id: data.acceptUserInvite.userID
                    },
                    org_id: data.acceptUserInvite.org.id,
                    org_name: data.acceptUserInvite.org.name
                }
            })
            return data.acceptUserInvite
        })
}

/**
 * Sends a GraphQL mutation to remove a user from an org
 *
 * @param orgID The org's ID
 * @param userID The user's ID to remove
 * @return An Observable that does emits `undefined` when done, then completes
 */
export function removeUserFromOrg(orgID: number, userID: string): Observable<never> {
    return mutateGraphQL(`
        mutation removeUserFromOrg {
            removeUserFromOrg(userID: $userID, orgID: $orgID) {
                alwaysNil
            }
        }
    `, {
        userID,
        orgID
    })
        .mergeMap(({ data, errors }) => {
            const eventData = {
                organization: {
                    remove: {
                        user_id: userID
                    },
                    org_id: orgID
                }
            }
            if (errors && errors.length > 0) {
                events.RemoveOrgMemberFailed.log(eventData)
                throw Object.assign(new Error(errors.map(e => e.message).join('\n')), { errors })
            }
            events.OrgMemberRemoved.log(eventData)
            // Reload user data
            return fetchCurrentUser()
        })
}
