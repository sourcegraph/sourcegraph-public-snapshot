import type { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import type { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { mergeSettings } from '@sourcegraph/shared/src/settings/settings'
import { currentUserMock, sharedGraphQlResults } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import type { WebGraphQlOperations } from '../graphql-operations'

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
                        id: 'TestDefaultSettingsID',
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
export const commonWebGraphQlResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ...sharedGraphQlResults,
    CurrentAuthState: () => ({
        currentUser: currentUserMock,
    }),
    ...createViewerSettingsGraphQLOverride(),
    GlobalAlertsSiteFlags: () => ({
        site: {
            __typename: 'Site',
            id: 'TestSiteID',
            needsRepositoryConfiguration: false,
            freeUsersExceeded: false,
            alerts: [],
            authProviders: {
                nodes: [builtinAuthProvider],
            },
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
        codeIntelligenceConfigurationPolicies: {
            totalCount: 1,
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
    SavedSearches: () => ({
        savedSearches: {
            nodes: [],
            totalCount: 0,
            pageInfo: { startCursor: null, endCursor: null, hasNextPage: false, hasPreviousPage: false },
        },
    }),
    LogEvents: () => ({
        logEvents: {
            alwaysNil: null,
        },
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
    EvaluateFeatureFlag: () => ({
        evaluateFeatureFlag: false,
    }),
    OrgFeatureFlagOverrides: () => ({
        organizationFeatureFlagOverrides: [],
    }),
    SearchHistoryEventLogsQuery: () => ({
        currentUser: {
            __typename: 'User',
            recentSearchLogs: {
                __typename: 'EventLogsConnection',
                nodes: [],
            },
        },
    }),
    DefaultSearchContextSpec: () => ({
        defaultSearchContext: {
            __typename: 'SearchContext',
            spec: 'global',
        },
    }),
}
