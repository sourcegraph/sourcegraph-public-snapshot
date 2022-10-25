import { concat, Observable } from 'rxjs'
import { map, mergeMap } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'

import { refreshAuthenticatedUser } from '../auth'
import { requestGraphQL } from '../backend/graphql'
import {
    CreateOrganizationResult,
    CreateOrganizationVariables,
    RemoveUserFromOrganizationResult,
    RemoveUserFromOrganizationVariables,
    Scalars,
    UpdateOrganizationResult,
    UpdateOrganizationVariables,
} from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

/**
 * Sends a GraphQL mutation to create an organization and returns an Observable that emits the new organization,
 * then completes.
 */
export function createOrganization(args: {
    /** The name of the organization. */
    name: string
    /** The new organization's display name (e.g. full name) in the organization profile. */
    displayName?: string
}): Promise<CreateOrganizationResult['createOrganization']> {
    return requestGraphQL<CreateOrganizationResult, CreateOrganizationVariables>(
        gql`
            mutation CreateOrganization($name: String!, $displayName: String) {
                createOrganization(name: $name, displayName: $displayName) {
                    id
                    name
                    settingsURL
                }
            }
        `,
        { name: args.name, displayName: args.displayName ?? null }
    )
        .pipe(
            mergeMap(({ data, errors }) => {
                if (!data || !data.createOrganization) {
                    eventLogger.log('NewOrgFailed')
                    throw createAggregateError(errors)
                }
                eventLogger.log('NewOrgCreated')
                return concat(refreshAuthenticatedUser(), [data.createOrganization])
            })
        )
        .toPromise()
}

/**
 * Sends a GraphQL mutation to remove a user from an organization.
 *
 * @returns An Observable that emits `undefined` when done, then completes
 */
export function removeUserFromOrganization(args: {
    /** The ID of the user to remove. */
    user: Scalars['ID']
    /** The organization's ID. */
    organization: Scalars['ID']
}): Observable<void> {
    return requestGraphQL<RemoveUserFromOrganizationResult, RemoveUserFromOrganizationVariables>(
        gql`
            mutation RemoveUserFromOrganization($user: ID!, $organization: ID!) {
                removeUserFromOrganization(user: $user, organization: $organization) {
                    alwaysNil
                }
            }
        `,
        args
    ).pipe(
        mergeMap(({ errors }) => {
            if (errors && errors.length > 0) {
                eventLogger.log('RemoveOrgMemberFailed')
                throw createAggregateError(errors)
            }
            eventLogger.log('OrgMemberRemoved')
            // Reload user data
            return concat(refreshAuthenticatedUser(), [undefined])
        })
    )
}

/**
 * Sends a GraphQL mutation to update an organization.
 *
 * @param id The ID of the organization.
 * @param displayName The display name of the organization.
 * @returns Observable that emits `undefined`, then completes
 */
export function updateOrganization(id: Scalars['ID'], displayName: string): Promise<void> {
    return requestGraphQL<UpdateOrganizationResult, UpdateOrganizationVariables>(
        gql`
            mutation UpdateOrganization($id: ID!, $displayName: String) {
                updateOrganization(id: $id, displayName: $displayName) {
                    id
                }
            }
        `,
        {
            id,
            displayName,
        }
    )
        .pipe(
            map(({ data, errors }) => {
                if (!data || (errors && errors.length > 0)) {
                    eventLogger.log('UpdateOrgSettingsFailed')
                    throw createAggregateError(errors)
                }
                eventLogger.log('OrgSettingsUpdated')
                return
            })
        )
        .toPromise()
}

export const GET_ORG_FEATURE_FLAG_VALUE = gql`
    query OrgFeatureFlagValue($orgID: ID!, $flagName: String!) {
        organizationFeatureFlagValue(orgID: $orgID, flagName: $flagName)
    }
`
export const ORG_CODE_FEATURE_FLAG_EMAIL_INVITE = 'org-email-invites'
export const GITHUB_APP_FEATURE_FLAG_NAME = 'github-app-cloud'
export const ORG_DELETION_FEATURE_FLAG_NAME = 'org-deletion'
