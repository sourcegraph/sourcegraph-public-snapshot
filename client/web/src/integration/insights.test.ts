import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { siteGQLID, siteID } from './jscontext'
import { percySnapshotWithVariants } from './utils'

describe('Code insights page', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })

        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
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
                                contents: JSON.stringify({
                                    experimentalFeatures: { codeInsights: true },
                                }),
                            },
                            settingsURL: '/site-admin/global-settings',
                            viewerCanAdminister: true,
                        },
                    ],
                    final: JSON.stringify({}),
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
