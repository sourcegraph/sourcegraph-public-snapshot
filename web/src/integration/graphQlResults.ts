import { GraphQLOverrides } from './helpers'
import { StatusMessage, IOrg, IAlert } from '../../../shared/src/graphql/schema'
import { builtinAuthProvider, siteID, siteGQLID } from './jscontext'

export const testUserID = 'TestUserID'
export const settingsID = 123

/**
 * Predefined results for GraphQL requests that are made on almost every page.
 */
export const commonGraphQlResults: GraphQLOverrides = {
    CurrentAuthState: () => ({
        currentUser: {
            __typename: 'User',
            id: testUserID,
            databaseID: 1,
            username: 'test',
            avatarURL: null,
            email: 'felix@sourcegraph.com',
            displayName: null,
            siteAdmin: true,
            tags: [] as string[],
            url: '/users/test',
            settingsURL: '/users/test/settings',
            organizations: { nodes: [] as IOrg[] },
            session: { canSignOut: true },
            viewerCanAdminister: true,
        },
    }),
    ViewerSettings: () => ({
        viewerSettings: {
            subjects: [
                {
                    __typename: 'DefaultSettings' as const,
                    latestSettings: {
                        id: 0,
                        contents: JSON.stringify({}),
                    },
                },
                {
                    __typename: 'Site',
                    id: siteGQLID,
                    siteID,
                    latestSettings: {
                        id: 470,
                        contents: JSON.stringify({}),
                    },
                    settingsURL: '/site-admin/global-settings',
                    viewerCanAdminister: true,
                },
            ] as any, // this is needed because ts-graphql-plugin has a bug in detecting types for unions in fragments
            final: JSON.stringify({}),
        },
    }),
    SiteFlags: () => ({
        site: {
            needsRepositoryConfiguration: false,
            freeUsersExceeded: false,
            alerts: [] as IAlert[],
            authProviders: {
                nodes: [builtinAuthProvider],
            },
            disableBuiltInSearches: false,
            sendsEmailVerificationEmails: true,
            updateCheck: {
                pending: false,
                checkedAt: '2020-07-07T12:31:16+02:00',
                errorMessage: null,
                updateVersionAvailable: null,
            },
            productSubscription: {
                license: { expiresAt: '3021-05-28T16:06:40Z' },
                noLicenseWarningUserCount: null,
            },
            productVersion: '0.0.0+dev',
        },
        productVersion: '0.0.0+dev',
    }),

    StatusMessages: () => ({
        statusMessages: [] as StatusMessage[],
    }),

    SiteAdminActivationStatus: () => ({
        externalServices: { totalCount: 3 },
        repositories: { totalCount: 9 },
        viewerSettings: {
            final: JSON.stringify({}),
        },
        users: { totalCount: 2 },
        currentUser: {
            usageStatistics: {
                searchQueries: 171,
                findReferencesActions: 14,
                codeIntelligenceActions: 670,
            },
        },
    }),
    // Note this is the response not for the admin
    ActivationStatus: () => ({
        // externalServices: { totalCount: 3 },
        // repositories: { totalCount: 9 },
        // viewerSettings: {
        //     final: JSON.stringify({}),
        // },
        // users: { totalCount: 2 },
        currentUser: {
            usageStatistics: {
                searchQueries: 171,
                findReferencesActions: 14,
                codeIntelligenceActions: 670,
            },
        },
    }),
    logEvent: () => ({
        logEvent: {
            alwaysNil: null,
        },
    }),
    logUserEvent: () => ({
        logUserEvent: {
            alwaysNil: null,
        },
    }),
}
