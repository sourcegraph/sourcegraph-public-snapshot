import { ExtensionsResult, SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations';
import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebGraphQlOperations } from '../../graphql-operations';
import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'
import { commonWebGraphQlResults } from '../graphQlResults'
import { siteGQLID, siteID } from '../jscontext'
import { percySnapshotWithVariants } from '../utils'

import { generateSearchInsightExtensionBundle } from './search-based-insight-extension';

/** Search based fake bundle URL. */
const searchBasedInsightBundleURL = 'https://sourcegraph.com/-/static/extension/search-based-insight.js'

/** Fake manifest of search insight extension. */
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
}

function overrideGraphQLExtensions(props: OverrideGraphQLExtensionsProps) {
    const { testContext, overrides } = props

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
                                'searchInsights.insight.graphQLTypesMigration': {
                                    'title': 'Migration to new GraphQL TS types',
                                    'repositories': [
                                        'github.com/sourcegraph/sourcegraph'
                                    ],
                                    'series': [
                                        {
                                            'name': 'Imports of old GQL.* types',
                                            'query': 'patternType:regex case:yes \\*\\sas\\sGQL',
                                            'stroke': 'var(--oc-red-7)'
                                        },
                                        {
                                            'name': 'Imports of new graphql-operations types',
                                            'query': "patternType:regexp case:yes /graphql-operations'",
                                            'stroke': 'var(--oc-blue-7)'
                                        }
                                    ],
                                    'step': {
                                        'weeks': 6
                                    }
                                },
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
            .send(generateSearchInsightExtensionBundle())
    })
}

describe('Code insights page', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest({ sourcegraphBaseUrl: 'https://sourcegraph.test:3443', devtools: true })
    })
    after(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })

        overrideGraphQLExtensions(testContext, {
            Insights: () => ({
                insights: {
                    nodes: [
                        {
                            title: 'Testing Insight',
                            description: 'Insight for testing',
                            series: [
                                {
                                    label: 'Insight',
                                    points: [
                                        {
                                            dateTime: '2021-02-11T00:00:00Z',
                                            value: 9,
                                        },
                                        {
                                            dateTime: '2021-01-27T00:00:00Z',
                                            value: 8,
                                        },
                                        {
                                            dateTime: '2021-01-12T00:00:00Z',
                                            value: 7,
                                        },
                                        {
                                            dateTime: '2020-12-28T00:00:00Z',
                                            value: 6,
                                        },
                                        {
                                            dateTime: '2020-12-13T00:00:00Z',
                                            value: 5,
                                        },
                                        {
                                            dateTime: '2020-11-28T00:00:00Z',
                                            value: 4,
                                        },
                                        {
                                            dateTime: '2020-11-13T00:00:00Z',
                                            value: 3,
                                        },
                                        {
                                            dateTime: '2020-10-29T00:00:00Z',
                                            value: 2,
                                        },
                                        {
                                            dateTime: '2020-10-14T00:00:00Z',
                                            value: 1,
                                        },
                                        {
                                            dateTime: '2020-09-29T00:00:00Z',
                                            value: 0,
                                        },
                                    ],
                                },
                            ],
                        },
                    ],
                },
            }),
        })
    })

    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('is styled correctly', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights')
        await driver.page.waitForSelector('[data-testid="line-chart__content"] svg circle')

        await percySnapshotWithVariants(driver.page, 'Code insights page')
    })
})
