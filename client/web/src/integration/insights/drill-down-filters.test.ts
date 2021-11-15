import assert from 'assert'

import delay from 'delay'
import { Key } from 'ts-key-enum'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { emptyResponse } from '@sourcegraph/shared/src/testing/integration/graphQlResults'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'

import { BACKEND_INSIGHTS } from './utils/insight-mock-data'
import { overrideGraphQLExtensions } from './utils/override-insights-graphql'

describe('Backend insight drill down filters', () => {
    let driver: Driver
    let testContext: WebIntegrationTestContext

    before(async () => {
        driver = await createDriverForTest()
    })

    after(() => driver?.close())

    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
    })

    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('should update user settings if drill-down filters have been persisted', async () => {
        const userSubjectSettigns = {
            'insights.allrepos': {
                'searchInsights.insight.backend_ID_001': {},
            },
        }

        overrideGraphQLExtensions({
            testContext,
            userSettings: userSubjectSettigns,
            overrides: {
                // Mock back-end insights with standard gql API handler.
                Insights: () => ({ insights: { nodes: BACKEND_INSIGHTS } }),
                OverwriteSettings: () => ({
                    settingsMutation: {
                        overwriteSettings: {
                            empty: emptyResponse,
                        },
                    },
                }),

                SubjectSettings: () => ({
                    settingsSubject: {
                        latestSettings: {
                            id: 310,
                            contents: JSON.stringify(userSubjectSettigns),
                        },
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await driver.page.waitForSelector('[data-testid="line-chart__content"] svg circle')

        await driver.page.click('button[aria-label="Filters"]')
        await driver.page.waitForSelector('[role="dialog"][aria-label="Drill-down filters panel"]')
        await driver.page.type('[name="excludeRepoRegexp"]', 'github.com/sourcegraph/sourcegraph')

        // Close the drill-down filter panel
        await driver.page.keyboard.press(Key.Escape)
        await driver.page.waitForSelector('[role="dialog"][aria-label="Drill-down filters panel"]', {
            hidden: true,
        })

        // In this time we should see active button state (filter dot should appear if we've got some filters)
        await driver.page.click('button[aria-label="Active filters"]')

        const variables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[role="dialog"][aria-label="Drill-down filters panel"] button[type="submit"]')
        }, 'Insights')

        assert.deepStrictEqual(variables, {
            ids: ['searchInsights.insight.backend_ID_001'],
            includeRepoRegex: '',
            excludeRepoRegex: 'github.com/sourcegraph/sourcegraph',
        })
    })

    it('should create a new insight with predefined filters via drill-down flow insight creation', async () => {
        const userSubjectSettigns = {
            'insights.allrepos': {
                'searchInsights.insight.backend_ID_001': {
                    title: 'Linear backend insight with filters',
                    repositories: [],
                    series: [
                        {
                            name: 'Series #1',
                            query: 'test query string',
                            stroke: 'var(--primary)',
                        },
                    ],
                    filters: {
                        includeRepoRegexp: '',
                        excludeRepoRegexp: 'github.com/sourcegraph/sourcegraph',
                    },
                },
            },
        }

        overrideGraphQLExtensions({
            testContext,
            userSettings: userSubjectSettigns,
            overrides: {
                // Mock back-end insights with standard gql API handler.
                Insights: () => ({ insights: { nodes: BACKEND_INSIGHTS } }),
                OverwriteSettings: () => ({
                    settingsMutation: {
                        overwriteSettings: {
                            empty: emptyResponse,
                        },
                    },
                }),

                SubjectSettings: () => ({
                    settingsSubject: {
                        latestSettings: {
                            id: 310,
                            contents: JSON.stringify(userSubjectSettigns),
                        },
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/all')
        await driver.page.waitForSelector('[data-testid="line-chart__content"] svg circle')

        await driver.page.click('button[aria-label="Active filters"]')
        await driver.page.waitForSelector('[role="dialog"][aria-label="Drill-down filters panel"]')

        await driver.page.click(
            '[role="dialog"][aria-label="Drill-down filters panel"] button[data-testid="save-as-new-view-button"]'
        )

        await driver.page.type('[name="insightName"]', 'Insight with filters')

        // Wait until async validation of the insight name field will pass
        await delay(500)

        const variables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[role="dialog"][aria-label="Drill-down filters panel"] button[type="submit"]')
        }, 'OverwriteSettings')

        assert.deepStrictEqual(JSON.parse(variables.contents), {
            'insights.allrepos': {
                'searchInsights.insight.backend_ID_001': {
                    title: 'Linear backend insight with filters',
                    repositories: [],
                    series: [
                        {
                            name: 'Series #1',
                            query: 'test query string',
                            stroke: 'var(--primary)',
                        },
                    ],
                    filters: {
                        includeRepoRegexp: '',
                        excludeRepoRegexp: 'github.com/sourcegraph/sourcegraph',
                    },
                },
                'searchInsights.insight.insightWithFilters': {
                    title: 'Insight with filters',
                    repositories: [],
                    series: [
                        {
                            name: 'Series #1',
                            query: 'test query string',
                            stroke: 'var(--primary)',
                        },
                    ],
                    filters: {
                        includeRepoRegexp: '',
                        excludeRepoRegexp: 'github.com/sourcegraph/sourcegraph',
                    },
                },
            },
        })
    })
})
