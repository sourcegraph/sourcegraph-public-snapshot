import { SearchGraphQlOperations } from '@sourcegraph/search'
import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { mergeSettings } from '@sourcegraph/shared/src/settings/settings'
import { testUserID, sharedGraphQlResults } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import { WebGraphQlOperations } from '../graphql-operations'
import {
    collaboratorsPayload,
    recentFilesPayload,
    recentSearchesPayload,
    savedSearchesPayload,
} from '../search/panels/utils'

import { builtinAuthProvider, siteGQLID, siteID } from './jscontext'

/**
 * Helper function for creating user and organization/site settings.
 */
export const createViewerSettingsGraphQLOverride = (
    settings: { user?: Settings; site?: Settings } = {}
): Pick<SharedGraphQlOperations, 'ViewerSettings'> => {
    const { user: userSettings = {}, site: siteSettings = {} } = settings
    return {
        ViewerSettings: () => ({
            viewerSettings: {
                __typename: 'SettingsCascade',
                subjects: [
                    {
                        __typename: 'DefaultSettings',
                        settingsURL: null,
                        viewerCanAdminister: false,
                        latestSettings: {
                            id: 0,
                            contents: JSON.stringify(userSettings),
                        },
                    },
                    {
                        __typename: 'Site',
                        id: siteGQLID,
                        siteID,
                        latestSettings: {
                            id: 470,
                            contents: JSON.stringify(siteSettings),
                        },
                        settingsURL: '/site-admin/global-settings',
                        viewerCanAdminister: true,
                        allowSiteSettingsEdits: true,
                    },
                ],
                final: JSON.stringify(mergeSettings([siteSettings, userSettings])),
            },
        }),
    }
}

/**
 * Predefined results for GraphQL requests that are made on almost every page.
 */
export const commonWebGraphQlResults: Partial<
    WebGraphQlOperations & SharedGraphQlOperations & SearchGraphQlOperations
> = {
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
            tosAccepted: true,
            url: '/users/test',
            settingsURL: '/users/test/settings',
            organizations: { nodes: [] },
            session: { canSignOut: true },
            viewerCanAdminister: true,
            searchable: true,
            emails: [],
        },
    }),
    ...createViewerSettingsGraphQLOverride(),
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
            __typename: 'SettingsCascade',
            subjects: [],
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
        //     __typename: 'SettingsCascade',
        //     subjects: [],
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
            __typename: 'User',
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
    LogEvents: () => ({
        logEvents: {
            alwaysNil: null,
        },
    }),
    AutoDefinedSearchContexts: () => ({
        autoDefinedSearchContexts: [
            {
                __typename: 'SearchContext',
                id: '1',
                spec: 'global',
                name: 'global',
                namespace: null,
                autoDefined: true,
                public: true,
                description: 'All repositories on Sourcegraph',
                updatedAt: '2021-03-15T19:39:11Z',
                repositories: [],
                query: '',
                viewerCanManage: false,
            },
            {
                __typename: 'SearchContext',
                id: '2',
                spec: '@test',
                name: 'test',
                namespace: {
                    __typename: 'User',
                    id: 'u1',
                    namespaceName: 'test',
                },
                autoDefined: true,
                public: true,
                description: 'Your repositories on Sourcegraph',
                updatedAt: '2021-03-15T19:39:11Z',
                repositories: [],
                query: '',
                viewerCanManage: false,
            },
        ],
    }),
    ListSearchContexts: () => ({
        searchContexts: {
            nodes: [],
            totalCount: 0,
            pageInfo: { hasNextPage: false, endCursor: null },
        },
    }),
    IsSearchContextAvailable: () => ({
        isSearchContextAvailable: false,
    }),
    ExternalServices: () => ({
        externalServices: {
            totalCount: 0,
            nodes: [],
            pageInfo: { hasNextPage: false, endCursor: null },
        },
    }),
    ExternalServicesScopes: () => ({
        externalServices: {
            nodes: [],
        },
    }),
    EvaluateFeatureFlag: () => ({
        evaluateFeatureFlag: false,
    }),
    OrgFeatureFlagValue: () => ({
        organizationFeatureFlagValue: false,
    }),
    OrgFeatureFlagOverrides: () => ({
        organizationFeatureFlagOverrides: [],
    }),
    GetTemporarySettings: () => ({
        temporarySettings: {
            __typename: 'TemporarySettings',
            contents: JSON.stringify({
                'user.daysActiveCount': 1,
                'user.lastDayActive': new Date().toDateString(),
                'search.usedNonGlobalContext': true,
            }),
        },
    }),
    EditTemporarySettings: () => ({
        editTemporarySettings: {
            alwaysNil: null,
        },
    }),
    HomePanelsQuery: () => ({
        node: {
            __typename: 'User',
            recentlySearchedRepositoriesLogs: recentSearchesPayload(),
            recentSearchesLogs: recentSearchesPayload(),
            recentFilesLogs: recentFilesPayload(),
            collaborators: collaboratorsPayload(),
        },
        savedSearches: savedSearchesPayload(),
    }),
}
