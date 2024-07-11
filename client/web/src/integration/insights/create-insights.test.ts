import assert from 'assert'

import delay from 'delay'
import { afterEach, beforeEach, describe, it } from 'mocha'

import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { TimeIntervalStepUnit } from '../../graphql-operations'
import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../context'
import { createEditorAPI } from '../utils'

import { LANG_STATS_INSIGHT_DATA_FIXTURE, SEARCH_INSIGHT_LIVE_PREVIEW_FIXTURE } from './fixtures/runtime-insights'
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
        await accessibilityAudit(driver.page)
    })

    it('is styled correctly, without welcome popup', async () => {
        overrideInsightsGraphQLApi({ testContext })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/create')

        // Waiting for all important part page be rendered.
        await driver.page.waitForSelector('[data-testid="create-search-insights"]')
        await driver.page.waitForSelector('[data-testid="create-lang-usage-insight"]')
        await accessibilityAudit(driver.page)
    })

    it('should run a proper GQL mutation if code-stats insight has been created', async () => {
        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                /**
                 * Mock for async repositories field validation.
                 */
                CheckRepositoryExists: () => ({
                    repository: {
                        __typename: 'Repository',
                        name: 'github.com/sourcegraph/sourcegraph',
                    },
                }),

                LangStatsInsightContent: () => LANG_STATS_INSIGHT_DATA_FIXTURE,

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

    it.skip('should run a proper GQL mutation if search-based insight has been created', async () => {
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
                CheckRepositoryExists: () => ({
                    repository: {
                        __typename: 'Repository',
                        name: 'github.com/sourcegraph/sourcegraph',
                    },
                }),

                // Mocks live preview chart
                GetInsightPreview: () => SEARCH_INSIGHT_LIVE_PREVIEW_FIXTURE,

                // Mock for repository suggest component
                RepositorySearchSuggestions: () => ({
                    repositories: { nodes: [] },
                }),

                CreateSearchBasedInsight: () => ({
                    __typename: 'Mutation',
                    createLineChartSearchInsight: {
                        view: {
                            id: '001',
                            isFrozen: false,
                            defaultFilters: {
                                includeRepoRegex: null,
                                excludeRepoRegex: null,
                                searchContexts: [],
                                __typename: 'InsightViewFilters',
                            },
                            dashboardReferenceCount: 0,
                            dashboards: { nodes: [] },

                            repositoryDefinition: {
                                repositories: ['github.com/sourcegraph/sourcegraph'],
                                __typename: 'InsightRepositoryScope',
                            },

                            presentation: {
                                __typename: 'LineChartInsightViewPresentation',
                                title: 'Test insight title',
                                seriesPresentation: [
                                    {
                                        seriesId: '1',
                                        label: 'test series #1 title',
                                        color: 'var(--oc-cyan-7)',
                                        __typename: 'LineChartDataSeriesPresentation',
                                    },
                                    {
                                        seriesId: '2',
                                        label: 'test series #2 title',
                                        color: 'var(--oc-grape-7)',
                                        __typename: 'LineChartDataSeriesPresentation',
                                    },
                                ],
                            },
                            dataSeriesDefinitions: [
                                {
                                    seriesId: '1',
                                    query: 'test series #1 query',
                                    timeScope: {
                                        unit: TimeIntervalStepUnit.MONTH,
                                        value: 2,
                                        __typename: 'InsightIntervalTimeScope',
                                    },
                                    isCalculated: false,
                                    generatedFromCaptureGroups: false,
                                    groupBy: null,
                                    __typename: 'SearchInsightDataSeriesDefinition',
                                },
                                {
                                    seriesId: '1',
                                    query: 'test series #2 query',
                                    timeScope: {
                                        unit: TimeIntervalStepUnit.MONTH,
                                        value: 2,
                                        __typename: 'InsightIntervalTimeScope',
                                    },
                                    isCalculated: false,
                                    generatedFromCaptureGroups: false,
                                    groupBy: null,
                                    __typename: 'SearchInsightDataSeriesDefinition',
                                },
                            ],
                            defaultSeriesDisplayOptions: {
                                limit: null,
                                numSamples: null,
                                sortOptions: {
                                    direction: null,
                                    mode: null,
                                },
                            },
                            __typename: 'InsightView',
                        },
                        __typename: 'InsightViewPayload',
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
        {
            const editor = await createEditorAPI(driver, '[data-testid="series-form"]:nth-child(1) .test-query-input')
            await editor.replace('test series #1 query')
        }

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
        {
            const editor = await createEditorAPI(driver, '[data-testid="series-form"]:nth-child(2) .test-query-input')
            await editor.replace('test series #2 query')
        }

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
