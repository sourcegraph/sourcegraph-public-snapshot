import { builtinAuthProvider, siteGQLID, siteID } from './jscontext'
import { WebGraphQlOperations } from '../graphql-operations'
import { SharedGraphQlOperations } from '../../../shared/src/graphql-operations'
import { testUserID, sharedGraphQlResults } from '../../../shared/src/testing/integration/graphQlResults'

/**
 * Predefined results for GraphQL requests that are made on almost every page.
 */
export const commonWebGraphQlResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ...sharedGraphQlResults,
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
            tags: [],
            url: '/users/test',
            settingsURL: '/users/test/settings',
            organizations: { nodes: [] },
            session: { canSignOut: true },
            viewerCanAdminister: true,
        },
    }),
    ViewerSettings: () => ({
        viewerSettings: {
            subjects: [
                {
                    __typename: 'DefaultSettings',
                    settingsURL: null,
                    viewerCanAdminister: false,
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
            ],
            final: JSON.stringify({}),
        },
    }),
    SiteFlags: () => ({
        site: {
            needsRepositoryConfiguration: false,
            freeUsersExceeded: false,
            alerts: [],
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
        statusMessages: [],
    }),

    SiteAdminActivationStatus: () => ({
        externalServices: { totalCount: 3 },
        repositoryStats: {
            gitDirBytes: '1825299556',
            indexedLinesCount: '2616264',
        },
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
    EventLogsData: () => ({
        node: {
            eventLogs: {
                nodes: [],
                totalCount: 0,
                pageInfo: {
                    hasNextPage: false,
                    endCursor: null,
                },
            },
        },
    }),
    savedSearches: () => ({
        savedSearches: [],
    }),
    LogEvent: () => ({
        logEvent: {
            alwaysNil: null,
        },
    }),
    LogUserEvent: () => ({
        logUserEvent: {
            alwaysNil: null,
        },
    }),
    SearchContexts: () => ({
        searchContexts: [
            {
                __typename: 'SearchContext',
                id: '1',
                spec: 'global',
                autoDefined: true,
                description: 'All repositories on Sourcegraph',
            },
            {
                __typename: 'SearchContext',
                id: '2',
                spec: '@username',
                autoDefined: true,
                description: 'Your repositories on Sourcegraph',
            },
        ],
    }),
    UserRepositories: () => ({
        node: {
            repositories: {
                totalCount: 0,
                nodes: [],
                pageInfo: { hasNextPage: false },
            },
        },
    }),
}
