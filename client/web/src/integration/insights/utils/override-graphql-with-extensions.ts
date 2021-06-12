import { View } from 'sourcegraph'

import { ExtensionsResult, SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'
import { ErrorLike } from '@sourcegraph/shared/src/util/errors'

import {
    BulkSearchCommits,
    BulkSearchFields,
    BulkSearchRepositories,
    WebGraphQlOperations,
} from '../../../graphql-operations'
import { WebIntegrationTestContext } from '../../context'
import { commonWebGraphQlResults } from '../../graphQlResults'
import { siteGQLID, siteID } from '../../jscontext'

import { getCodeStatsInsightExtensionBundle, getSearchInsightExtensionBundle } from './insight-extension-bundles'

/**
 * Search based fake bundle URL.
 */
const searchBasedInsightExtensionBundleURL = 'https://sourcegraph.com/-/static/extension/search-based-insight.js'

/**
 * Fake manifest of search based insight extension.
 */
const searchBasedInsightRawManifest = JSON.stringify({
    url: searchBasedInsightExtensionBundleURL,
    activationEvents: ['*'],
    browserslist: [],
    contributes: {},
    description: 'Search based insight extension',
    devDependencies: {},
    extensionID: 'search-based-insight',
    license: 'MIT',
    main: 'dist/search-based-insight.js',
    name: 'search-based-insight',
    publisher: 'mock-author',
    readme: '# Search based insight (Sourcegraph extension))\n',
    scripts: {},
    title: 'Search based insight',
    version: '0.0.0-DEVELOPMENT',
})

/**
 * Code stats insight fake extension bundle URL.
 */
const codeStatsInsightExtensionBundleURL = 'https://sourcegraph.com/-/static/extension/code-stats-insight.js'

/**
 * Fake manifest of code stats insight extension.
 */
const codeStatsInsightRawManifest = JSON.stringify({
    url: codeStatsInsightExtensionBundleURL,
    activationEvents: ['*'],
    browserslist: [],
    contributes: {},
    description: 'Code stats insight extension',
    devDependencies: {},
    extensionID: 'code-stats-insight',
    license: 'MIT',
    main: 'dist/code-stats-insight.js',
    name: 'code-stats-insight',
    publisher: 'mock-author',
    readme: '# Code stats insight (Sourcegraph extension))\n',
    scripts: {},
    title: 'Code stats insight',
    version: '0.0.0-DEVELOPMENT',
})

const extensionNodes: ExtensionsResult['extensionRegistry']['extensions']['nodes'] = [
    {
        extensionID: 'search-based-insight',
        id: 'test-search-extension',
        manifest: { raw: searchBasedInsightRawManifest },
        url: '/extensions/search-based-insight',
        viewerCanAdminister: false,
    },
    {
        extensionID: 'code-stats-insight',
        id: 'test-code-stats-extension',
        manifest: { raw: codeStatsInsightRawManifest },
        url: '/extensions/code-stats-insight',
        viewerCanAdminister: false,
    },
]

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
    /** Page driver context. */
    testContext: WebIntegrationTestContext
    /** Overrides for gql API calls. */
    overrides?: Partial<WebGraphQlOperations & SharedGraphQlOperations & CustomInsightsOperations>
    /**
     * Mock map data for insight extension mocking system.
     * Key is an insight ID and value is mocked insight data
     */
    insightExtensionsMocks?: Record<string, View | undefined | ErrorLike>
    /** User settings. */
    userSettings?: Record<any, any>
    /** Organization setting. */
    orgSettings?: Record<any, any>
}

/**
 * Test setup handler used for mocking common parts of API, extension insight API and
 * extension js bundle requests.
 *
 * @param props - Custom override for code insight APIs (gql, user setting, extensions)
 */
export function overrideGraphQLExtensions(props: OverrideGraphQLExtensionsProps): void {
    const { testContext, overrides = {}, insightExtensionsMocks = {}, userSettings = {}, orgSettings = {} } = props

    testContext.overrideGraphQL({
        ...commonWebGraphQlResults,
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
                                extensions: {},
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
                                extensions: {
                                    'search-based-insight': true,
                                    'code-stats-insight': true,
                                },
                                ...userSettings,
                            }),
                        },
                    },
                ],
                final: JSON.stringify({}),
            },
        }),
        Extensions: () => ({
            extensionRegistry: {
                extensions: {
                    nodes: extensionNodes,
                },
            },
        }),
        ...overrides,
    })

    // Mock extension bundle
    testContext.server.get(searchBasedInsightExtensionBundleURL).intercept((request, response) => {
        response
            .type('application/javascript; charset=utf-8')
            .send(getSearchInsightExtensionBundle(insightExtensionsMocks))
    })

    testContext.server.get(codeStatsInsightExtensionBundleURL).intercept((request, response) => {
        response
            .type('application/javascript; charset=utf-8')
            .send(getCodeStatsInsightExtensionBundle(insightExtensionsMocks))
    })
}
