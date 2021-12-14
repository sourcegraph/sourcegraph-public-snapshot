import assert from 'assert'

import delay from 'delay'

import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { emptyResponse } from '@sourcegraph/shared/src/testing/integration/graphQlResults'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'
import { percySnapshotWithVariants } from '../utils'

import {
    INSIGHT_TYPES_MIGRATION_BULK_SEARCH,
    INSIGHT_TYPES_MIGRATION_COMMITS,
    LangStatsInsightContent,
} from './utils/insight-mock-data'
import { overrideGraphQLExtensions } from './utils/override-insights-graphql'

describe('Code insight create insight page', () => {
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
    it('is styled correctly, with welcome popup', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/create')

        // Waiting for all important part page be rendered.
        await driver.page.waitForSelector('[data-testid="create-search-insights"]')
        await driver.page.waitForSelector('[data-testid="create-lang-usage-insight"]')
        await driver.page.waitForSelector('[data-testid="explore-extensions"]')

        await percySnapshotWithVariants(driver.page, 'Create new insight page — Welcome popup')
    })

    it('is styled correctly, without welcome popup', async () => {
        overrideGraphQLExtensions({ testContext })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/create')

        // Waiting for all important part page be rendered.
        await driver.page.waitForSelector('[data-testid="create-search-insights"]')
        await driver.page.waitForSelector('[data-testid="create-lang-usage-insight"]')
        await driver.page.waitForSelector('[data-testid="explore-extensions"]')

        await percySnapshotWithVariants(driver.page, 'Create new insight page')
    })

    it('should update user/org setting if code stats insight has been created', async () => {
        overrideGraphQLExtensions({
            testContext,
            overrides: {
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
                            contents: JSON.stringify({}),
                        },
                    },
                }),

                /**
                 * Mock for async repositories field validation.
                 */
                BulkRepositoriesSearch: () => ({
                    repoSearch0: { name: 'github.com/sourcegraph/sourcegraph' },
                }),

                LangStatsInsightContent: () => LangStatsInsightContent,

                /** Mock for repository suggest component. */
                RepositorySearchSuggestions: () => ({
                    repositories: { nodes: [] },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/create/lang-stats')

        // Waiting for all important part of creation form will be rendered.
        await driver.page.waitForSelector('[data-testid="code-stats-insight-creation-page-content"]')

        // Add new repo to repositories field
        await driver.page.type('[name="repository"]', 'github.com/sourcegraph/sourcegraph')
        // Wait until async validation on repository field will be finished
        await delay(1000)

        // With repository filled input we have to have code stats insight live preview
        // charts - pie chart
        await driver.page.waitForSelector('[data-testid="pie-chart-arc-element"]')

        // To blur repository input
        await driver.page.click('input[name="title"]')
        // Change insight title
        await driver.page.type('input[name="title"]', 'Test insight title')

        await percySnapshotWithVariants(driver.page, 'Code insights create new language usage insight')

        const addToUserConfigRequest = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[data-testid="insight-save-button"]')
        }, 'OverwriteSettings')

        // Check that new org settings config has edited insight
        assert.deepStrictEqual(JSON.parse(addToUserConfigRequest.contents), {
            'codeStatsInsights.insight.testInsightTitle': {
                title: 'Test insight title',
                repository: 'github.com/sourcegraph/sourcegraph',
                otherThreshold: 0.03,
            },
        })
    })

    it('should update user/org settings if search based insight has been created', async () => {
        // Mock `Date.now` to stabilize timestamps
        await driver.page.evaluateOnNewDocument(() => {
            // Number of ms between Unix epoch and June 31, 2021
            const mockMs = new Date('June 1, 2021 00:00:00 UTC').getTime()
            Date.now = () => mockMs
        })

        overrideGraphQLExtensions({
            testContext,
            overrides: {
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
                            contents: JSON.stringify({}),
                        },
                    },
                }),

                // Mock for async repositories field validation.
                BulkRepositoriesSearch: () => ({
                    repoSearch0: { name: 'github.com/sourcegraph/sourcegraph' },
                }),

                // Mocks of commits searching and data search itself for live preview chart
                BulkSearchCommits: () => INSIGHT_TYPES_MIGRATION_COMMITS,
                BulkSearch: () => INSIGHT_TYPES_MIGRATION_BULK_SEARCH,

                // Mock for repository suggest component
                RepositorySearchSuggestions: () => ({
                    repositories: { nodes: [] },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/create/search')

        // Waiting for all important part of creation form will be rendered.
        await driver.page.waitForSelector('[data-testid="search-insight-create-page-content"]')

        // Add new repo to repositories field
        await driver.page.type('[name="repositories"]', 'github.com/sourcegraph/sourcegraph')

        // Change insight title
        await driver.page.type('input[name="title"]', 'Test insight title')

        // Create chart data series

        // Add first series name
        await driver.page.type(
            '[data-testid="series-form"]:nth-child(1) input[name="seriesName"]',
            'test series #1 title'
        )

        // Add first series query
        await driver.page.waitForSelector('[data-testid="series-form"]:nth-child(1) #monaco-query-input')
        await driver.replaceText({
            selector: '[data-testid="series-form"]:nth-child(1) #monaco-query-input',
            newText: 'test series #1 query',
            enterTextMethod: 'type',
            selectMethod: 'keyboard',
        })

        // Pick first series color
        await driver.page.click('[data-testid="series-form"]:nth-child(1) label span[title="Cyan"]')

        // Add second series
        await driver.page.click('[data-testid="form-series"] [data-testid="add-series-button"]')

        // Add second series name
        await driver.page.type(
            '[data-testid="series-form"]:nth-child(2) input[name="seriesName"]',
            'test series #2 title'
        )

        // Add second series query
        await driver.page.waitForSelector('[data-testid="series-form"]:nth-child(1) #monaco-query-input')
        await driver.replaceText({
            selector: '[data-testid="series-form"]:nth-child(2) #monaco-query-input',
            newText: 'test series #2 query',
            enterTextMethod: 'type',
            selectMethod: 'keyboard',
        })

        // With two filled data series our mock for live preview should work - render line chart with two lines
        await driver.page.waitForSelector('[data-testid="line-chart__content"] svg circle')

        const addToUserConfigRequest = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[data-testid="insight-save-button"]')
        }, 'OverwriteSettings')

        // Check that new org settings config has edited insight
        assert.deepStrictEqual(JSON.parse(addToUserConfigRequest.contents), {
            'searchInsights.insight.testInsightTitle': {
                title: 'Test insight title',
                repositories: ['github.com/sourcegraph/sourcegraph'],
                series: [
                    {
                        name: 'test series #1 title',
                        query: 'test series #1 query',
                        stroke: 'var(--oc-cyan-7)',
                    },
                    {
                        name: 'test series #2 title',
                        query: 'test series #2 query',
                        stroke: 'var(--oc-grape-7)',
                    },
                ],
                step: {
                    months: 2,
                },
            },
        })
    })
})
