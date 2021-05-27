import { View } from 'sourcegraph';

import { ExtensionsResult, SharedGraphQlOperations } from '@sourcegraph/shared/out/src/graphql-operations';

import { WebGraphQlOperations } from '../../../graphql-operations';
import { WebIntegrationTestContext } from '../../context';
import { commonWebGraphQlResults } from '../../graphQlResults';
import { siteGQLID, siteID } from '../../jscontext';

import { generateSearchInsightExtensionBundle } from './search-based-insight-extension';

/**
 * Search based fake bundle URL.
 * */
const searchBasedInsightBundleURL = 'https://sourcegraph.com/-/static/extension/search-based-insight.js'

/**
 * Fake manifest of search based insight extension.
 * */
const searchBasedInsightRawManifest = JSON.stringify({
    url: searchBasedInsightBundleURL,
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

const extensionNodes: ExtensionsResult['extensionRegistry']['extensions']['nodes'] = [
    {
        extensionID: 'search-based-insight',
        id: 'test-extension',
        manifest: { raw: searchBasedInsightRawManifest },
        url: '/extensions/search-based-insight',
        viewerCanAdminister: false,
    },
]

interface OverrideGraphQLExtensionsProps {
    testContext: WebIntegrationTestContext,
    overrides: Partial<WebGraphQlOperations & SharedGraphQlOperations>
    insightExtensionsMocks?: Record<string, View>
    userSettings?: Record<any, any>
}

export function overrideGraphQLExtensions(props: OverrideGraphQLExtensionsProps): void {
    const { testContext, overrides, insightExtensionsMocks, userSettings = {} } = props

    testContext.overrideGraphQL({
        ...commonWebGraphQlResults,
        ...overrides,
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
                    },
                    {
                        __typename: 'User',
                        id: 'TestGQLUserID',
                        username: 'testusername',
                        settingsURL: '/user/testusername/settings',
                        displayName: 'test',
                        viewerCanAdminister: true,
                        latestSettings: {
                            id: 310,
                            contents: JSON.stringify({
                                extensions: { 'search-based-insight': true },
                                ...userSettings
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
    })

    // Mock extension bundle
    testContext.server.get(searchBasedInsightBundleURL).intercept((request, response) => {
        response.type('application/javascript; charset=utf-8')
            .send(generateSearchInsightExtensionBundle(insightExtensionsMocks))
    })
}
