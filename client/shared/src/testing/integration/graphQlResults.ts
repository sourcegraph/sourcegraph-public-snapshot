import type { AuthenticatedUser } from '../../auth'
import type { SharedGraphQlOperations } from '../../graphql-operations'

export const testUserID = 'TestUserID'
export const settingsID = 123

export const currentUserMock = {
    __typename: 'User',
    id: testUserID,
    databaseID: 1,
    username: 'test',
    avatarURL: null,
    displayName: null,
    siteAdmin: true,
    tosAccepted: true,
    url: '/users/test',
    settingsURL: '/users/test/settings',
    organizations: { nodes: [] },
    session: { canSignOut: true },
    viewerCanAdminister: true,
    emails: [{ email: 'felix@sourcegraph.com', isPrimary: true, verified: true }],
    latestSettings: null,
    hasVerifiedEmail: true,
    permissions: {
        __typename: 'PermissionConnection',
        nodes: [
            { __typename: 'Permission', id: 'id1', displayName: 'BATCH_CHANGES#READ' },
            { __typename: 'Permission', id: 'id2', displayName: 'BATCH_CHANGES#WRITE' },
        ],
    },
} satisfies AuthenticatedUser

/**
 * Predefined results for GraphQL requests that are made on almost every page.
 */
export const sharedGraphQlResults: Partial<SharedGraphQlOperations> = {}

export const emptyResponse = {
    alwaysNil: null,
}
