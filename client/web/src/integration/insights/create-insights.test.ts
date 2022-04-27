import assert from 'assert'

import delay from 'delay'

import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, WebIntegrationTestContext } from '../context'
import { percySnapshotWithVariants } from '../utils'

import {
    MIGRATION_TO_GQL_INSIGHT_COMMITS_FIXTURE,
    MIGRATION_TO_GQL_INSIGHT_MATCHES_DATA_FIXTURE,
    SOURCEGRAPH_LANG_STATS_INSIGHT_DATA_FIXTURE,
} from './fixtures/runtime-insights'
import { overrideInsightsGraphQLApi } from './utils/override-insights-graphql-api'

describe('Code insight create insight page', () => {
    let driver: Driver
    let testContext: WebIntegrationTestContext

    before(async () => {
        driver = await createDriverForTest()
    })

    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
            customContext: {
                // Enforce using a new gql API for code insights pages
                codeInsightsGqlApiEnabled: true,
            },
        })
    })

    after(() => driver?.close())
    afterEach(() => testContext?.dispose())
    afterEachSaveScreenshotIfFailed(() => driver.page)

    it('is styled correctly, with welcome popup', async () => {
        overrideInsightsGraphQLApi({ testContext })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/create')

        // Waiting for all important part page be rendered.
        await driver.page.waitForSelector('[data-testid="create-search-insights"]')
        await driver.page.waitForSelector('[data-testid="create-lang-usage-insight"]')
        await driver.page.waitForSelector('[data-testid="explore-extensions"]')

        await percySnapshotWithVariants(driver.page, 'Create new insight page — Welcome popup')
        await accessibilityAudit(driver.page)
    })

    it('is styled correctly, without welcome popup', async () => {
        overrideInsightsGraphQLApi({ testContext })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/create')

        // Waiting for all important part page be rendered.
        await driver.page.waitForSelector('[data-testid="create-search-insights"]')
        await driver.page.waitForSelector('[data-testid="create-lang-usage-insight"]')
        await driver.page.waitForSelector('[data-testid="explore-extensions"]')

        await percySnapshotWithVariants(driver.page, 'Create new insight page')
        await accessibilityAudit(driver.page)
    })

    it('should run a proper GQL mutation if code-stats insight has been created', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                /**
                 * Mock for async repositories field validation.
                 */
                BulkRepositoriesSearch: () => ({
                    repoSearch0: { name: 'github.com/sourcegraph/sourcegraph' },
                }),

                LangStatsInsightContent: () => SOURCEGRAPH_LANG_STATS_INSIGHT_DATA_FIXTURE,

                /** Mock for repository suggest component. */
                RepositorySearchSuggestions: () => ({
                    repositories: { nodes: [] },
                }),

                CreateLangStatsInsight: () => ({
                    __typename: 'Mutation',
                    createPieChartSearchInsight: {
                        __typename: 'InsightViewPayload',
                        view: { __typename: 'InsightView', id: '001' },
                    },
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
        await accessibilityAudit(driver.page)

        const addToUserConfigRequest = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[data-testid="insight-save-button"]')
        }, 'CreateLangStatsInsight')

        // Check that new org settings config has edited insight
        assert.deepStrictEqual(addToUserConfigRequest.input, {
            query: '',
            repositoryScope: {
                repositories: ['github.com/sourcegraph/sourcegraph'],
            },
            presentationOptions: {
                title: 'Test insight title',
                otherThreshold: 0.03,
            },
        })
    })

    it('should run a proper GQL mutation if search-based insight has been created', async () => {
        // Mock `Date.now` to stabilize timestamps
        await driver.page.evaluateOnNewDocument(() => {
            // Number of ms between Unix epoch and June 31, 2021
            const mockMs = new Date('June 1, 2021 00:00:00 UTC').getTime()
            Date.now = () => mockMs
        })

        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mock for async repositories field validation.
                BulkRepositoriesSearch: () => ({
                    repoSearch0: { name: 'github.com/sourcegraph/sourcegraph' },
                }),

                // Mocks of commits searching and data search itself for live preview chart
                BulkSearchCommits: () => MIGRATION_TO_GQL_INSIGHT_COMMITS_FIXTURE,
                BulkSearch: () => MIGRATION_TO_GQL_INSIGHT_MATCHES_DATA_FIXTURE,

                // Mock for repository suggest component
                RepositorySearchSuggestions: () => ({
                    repositories: { nodes: [] },
                }),

                CreateSearchBasedInsight: () => ({
                    __typename: 'Mutation',
                    createLineChartSearchInsight: {
                        __typename: 'InsightViewPayload',
                        view: { __typename: 'InsightView', id: '001' },
                    },
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
        await driver.page.click('[data-testid="series-form"]:nth-child(1) label[title="Cyan"]')

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
        await driver.page.waitForSelector('[data-testid="code-search-insight-live-preview"] circle')

        const addToUserConfigRequest = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[data-testid="insight-save-button"]')
        }, 'CreateSearchBasedInsight')

        // Check that new org settings config has edited insight
        assert.deepStrictEqual(addToUserConfigRequest.input, {
            dataSeries: [
                {
                    query: 'test series #1 query',
                    options: {
                        label: 'test series #1 title',
                        lineColor: 'var(--oc-cyan-7)',
                    },
                    repositoryScope: {
                        repositories: ['github.com/sourcegraph/sourcegraph'],
                    },
                    timeScope: {
                        stepInterval: {
                            unit: 'MONTH',
                            value: 2,
                        },
                    },
                },
                {
                    query: 'test series #2 query',
                    options: {
                        label: 'test series #2 title',
                        lineColor: 'var(--oc-grape-7)',
                    },
                    repositoryScope: {
                        repositories: ['github.com/sourcegraph/sourcegraph'],
                    },
                    timeScope: {
                        stepInterval: {
                            unit: 'MONTH',
                            value: 2,
                        },
                    },
                },
            ],
            options: {
                title: 'Test insight title',
            },
        })
    })
})
