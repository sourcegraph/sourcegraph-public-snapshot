import { concat, Observable } from 'rxjs'
import { map, mergeMap, take } from 'rxjs/operators'
import { currentUser, refreshCurrentUser } from '../auth'
import { gql, mutateGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { eventLogger } from '../tracking/eventLogger'
import { settingsFragment } from '../user/settings/backend'
import { createAggregateError } from '../util/errors'

/**
 * Sends a GraphQL mutation to create an organization and returns an Observable that emits the new organization,
 * then completes.
 */
export function createOrganization(args: {
    /** The name of the organization. */
    name: string
    /** The new organization's display name (e.g. full name) in the organization profile. */
    displayName?: string
}): Observable<GQL.IOrg> {
    return currentUser.pipe(
        take(1),
        mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in.')
            }

            return mutateGraphQL(
                gql`
                    mutation createOrganization($name: String!, $displayName: String) {
                        createOrganization(name: $name, displayName: $displayName) {
                            id
                            name
                        }
                    }
                `,
                args
            )
        }),
        mergeMap(({ data, errors }) => {
            if (!data || !data.createOrganization) {
                eventLogger.log('NewOrgFailed')
                throw createAggregateError(errors)
            }
            eventLogger.log('NewOrgCreated', {
                organization: {
                    org_id: data.createOrganization.id,
                    org_name: data.createOrganization.name,
                },
            })
            return concat(refreshCurrentUser(), [data.createOrganization])
        })
    )
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
                    mutation AcceptUserInvite($inviteToken: String!) {
                        acceptUserInvite(inviteToken: $inviteToken) {
                            alwaysNil
                        }
                    }
                `,
                options
            )
        }),
        map(({ data, errors }) => {
            if (!data || !data.acceptUserInvite) {
                eventLogger.log('AcceptInviteFailed')
                throw createAggregateError(errors)
            }
            return
        })
    )
}

/**
 * Sends a GraphQL mutation to remove a user from an organization.
 *
 * @return An Observable that emits `undefined` when done, then completes
 */
export function removeUserFromOrganization(args: {
    /** The ID of the user to remove. */
    user: GQL.ID
    /** The organization's ID. */
    organization: GQL.ID
}): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation removeUserFromOrganization($user: ID!, $organization: ID!) {
                removeUserFromOrganization(user: $user, organization: $organization) {
                    alwaysNil
                }
            }
        `,
        args
    ).pipe(
        mergeMap(({ data, errors }) => {
            const eventData = {
                organization: {
                    remove: {
                        user_id: args.user,
                    },
                    org_id: args.organization,
                },
            }
            if (errors && errors.length > 0) {
                eventLogger.log('RemoveOrgMemberFailed', eventData)
                throw createAggregateError(errors)
            }
            eventLogger.log('OrgMemberRemoved', eventData)
            // Reload user data
            return concat(refreshCurrentUser(), [void 0])
        })
    )
}

/**
 * Sends a GraphQL mutation to update an org
 *
 * @param id The ID of the org
 * @param displayName The display name of the org
 * @return Observable that emits `undefined`, then completes
 */
export function updateOrg(id: string, displayName: string): Observable<void> {
    return currentUser.pipe(
        take(1),
        mergeMap(user => {
            if (!user) {
                throw new Error('User must be signed in.')
            }

            const variables = {
                id,
                displayName,
            }
            return mutateGraphQL(
                gql`
                    mutation updateOrg($id: ID!, $displayName: String) {
                        updateOrg(id: $id, displayName: $displayName) {
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
                    },
                    org_id: id,
                },
            }
            if (!data || (errors && errors.length > 0)) {
                eventLogger.log('UpdateOrgSettingsFailed', eventData)
                throw createAggregateError(errors)
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
                updateOrgSettings(id: $id, lastKnownSettingsID: $lastKnownSettingsID, contents: $contents) {
                    ...SettingsFields
                }
            }
            ${settingsFragment}
        `,
        { id, lastKnownSettingsID, contents }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return
        })
    )
}
