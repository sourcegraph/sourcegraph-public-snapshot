import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import {
    BulkSearchCommits,
    BulkSearchFields,
    BulkSearchRepositories,
    WebGraphQlOperations,
} from '../../../graphql-operations'
import { WebIntegrationTestContext } from '../../context'
import { commonWebGraphQlResults } from '../../graphQlResults'
import { siteGQLID, siteID } from '../../jscontext'

/**
 * Some of insight creation UI gql api requests do not have
 * generated types due their dynamic nature. Because of that we
 * must write these api call types below manually for testing purposes.
 */
interface CustomInsightsOperations {
    /** API handler used for repositories field async validation. */
    BulkRepositoriesSearch: () => Record<string, BulkSearchRepositories>

    /** Internal API handler for fetching commits data for live preview chart. */
    BulkSearchCommits: () => Record<string, BulkSearchCommits>

    /**
     * Internal API handler for fetching actual data according to search commits
     * for live preview chart.
     */
    BulkSearch: () => Record<string, BulkSearchFields>
}

interface OverrideGraphQLExtensionsProps {
    testContext: WebIntegrationTestContext
    overrides?: Partial<WebGraphQlOperations & SharedGraphQlOperations & CustomInsightsOperations>
    userSettings?: Record<any, any>
    orgSettings?: Record<any, any>
}

/**
 * Test setup handler used for mocking common parts of API, extension insight API and
 * extension js bundle requests.
 *
 * @param props - Custom override for code insight APIs (gql, user setting, extensions)
 */
export function overrideGraphQLExtensions(props: OverrideGraphQLExtensionsProps): void {
    const { testContext, overrides = {}, userSettings = {}, orgSettings = {} } = props

    testContext.overrideGraphQL({
        ...commonWebGraphQlResults,
        // Mock temporary settings cause code insights beta modal UI relies on this handler to show/hide
        // modal UI on all code insights related pages.
        GetTemporarySettings: () => ({
            temporarySettings: {
                __typename: 'TemporarySettings',
                contents: JSON.stringify({ 'insights.freeBetaAccepted': true }),
            },
        }),
        Insights: () => ({ insights: { nodes: [] } }),
        CurrentAuthState: () => ({
            currentUser: {
                __typename: 'User',
                id: testUserID,
                databaseID: 1,
                username: 'test',
                avatarURL: null,
                email: 'vova@sourcegraph.com',
                displayName: null,
                siteAdmin: true,
                tags: [],
                url: '/users/test',
                settingsURL: '/users/test/settings',
                organizations: {
                    nodes: [
                        {
                            name: 'test organization',
                            displayName: 'Test organization',
                            id: 'Org_test_id',
                            settingsURL: '/organizations/test_organization/settings',
                            url: '/organizations/test_organization/settings',
                        },
                    ],
                },
                session: { canSignOut: true },
                viewerCanAdminister: true,
            },
        }),
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
                            contents: JSON.stringify({
                                experimentalFeatures: { codeInsights: true },
                            }),
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
                        allowSiteSettingsEdits: true,
                    },
                    {
                        __typename: 'Org',
                        name: 'test organization',
                        displayName: 'Test organization',
                        id: 'Org_test_id',
                        viewerCanAdminister: true,
                        settingsURL: '/organizations/test_organization/settings',
                        latestSettings: {
                            id: 320,
                            contents: JSON.stringify({
                                ...orgSettings,
                            }),
                        },
                    },
                    {
                        __typename: 'User',
                        id: testUserID,
                        username: 'testusername',
                        settingsURL: '/user/testusername/settings',
                        displayName: 'test',
                        viewerCanAdminister: true,
                        latestSettings: {
                            id: 310,
                            contents: JSON.stringify({
                                ...userSettings,
                            }),
                        },
                    },
                ],
                final: JSON.stringify({}),
            },
        }),
        Extensions: () => ({ extensionRegistry: { __typename: 'ExtensionRegistry', extensions: { nodes: [] } } }),
        ...overrides,
    })
}
