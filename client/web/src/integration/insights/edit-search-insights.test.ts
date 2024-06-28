import assert from 'assert'

import delay from 'delay'
import { afterEach, beforeEach, describe, it } from 'mocha'
import { Key } from 'ts-key-enum'

import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../context'
import { createEditorAPI } from '../utils'

import { createJITMigrationToGQLInsightMetadataFixture } from './fixtures/insights-metadata'
import { SEARCH_INSIGHT_LIVE_PREVIEW_FIXTURE } from './fixtures/runtime-insights'
import { overrideInsightsGraphQLApi } from './utils/override-insights-graphql-api'

interface InsightValues {
    series: {
        name?: string | null
        query?: string | null
        stroke?: string | null
    }[]
    repositories?: string
    title?: string
    visibility?: string
    step: Record<string, string | number | undefined>
}

function getInsightFormValues(): InsightValues {
    const repositories = document.querySelector<HTMLInputElement>('[name="repositories"]')?.value
    const title = document.querySelector<HTMLInputElement>('input[name="title"]')?.value
    const visibility = document.querySelector<HTMLInputElement>('input[name="visibility"]:checked')?.value
    const granularityType = document.querySelector<HTMLInputElement>('input[name="step"]:checked')?.value ?? ''
    const granularityValue = document.querySelector<HTMLInputElement>('input[name="stepValue"]')?.value

    const series = [...document.querySelectorAll('[data-testid="form-series"] [data-testid="series-card"]')].map(
        card => ({
            name: card.querySelector('[data-testid="series-name"]')?.textContent,
            query: card.querySelector('[data-testid="series-query"]')?.textContent,
            stroke: card.querySelector<HTMLElement>('[data-testid="series-color-mark"]')?.style.color,
        })
    )

    return {
        series,
        repositories,
        title,
        visibility,
        step: { [granularityType]: granularityValue },
    }
}

async function clearAndType(driver: Driver, selector: string, value: string): Promise<void> {
    await driver.page.click(selector, { clickCount: 3 })
    await driver.page.type(selector, value)
}

describe('Code insight edit insight page', () => {
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

    it.skip('should run a proper GQL mutation if search based insight has been updated', async () => {
        // Mock `Date.now` to stabilize timestamps
        await driver.page.evaluateOnNewDocument(() => {
            // Number of ms between Unix epoch and June 31, 2021
            const mockMs = new Date('June 1, 2021 00:00:00 UTC').getTime()
            Date.now = () => mockMs
        })

        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mock insight config query
                GetInsights: () => ({
                    __typename: 'Query',
                    insightViews: {
                        __typename: 'InsightViewConnection',
                        nodes: [createJITMigrationToGQLInsightMetadataFixture({ type: 'just-in-time' })],
                    },
                }),

                // Mocks live preview chart
                GetInsightPreview: () => SEARCH_INSIGHT_LIVE_PREVIEW_FIXTURE,

                // Mock for repository suggest component
                RepositorySearchSuggestions: () => ({
                    repositories: { nodes: [{ id: '001', name: 'github.com/sourcegraph/about' }] },
                }),

                UpdateLineChartSearchInsight: () => ({
                    __typename: 'Mutation',
                    updateLineChartSearchInsight: {
                        __typename: 'InsightViewPayload',
                        view: createJITMigrationToGQLInsightMetadataFixture({ type: 'just-in-time' }),
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/edit/001')

        // Waiting for all important part of creation form will be rendered.
        await driver.page.waitForSelector('[name="repositories"]')

        // Edit all insight form fields ↓

        // Move user cursor at the end of the input
        await driver.page.click('[name="repositories"]', { clickCount: 3 })
        await driver.page.keyboard.press('ArrowRight')

        // Add new repo to repositories field
        await driver.page.keyboard.type(', github.com/sourcegraph/about')

        // Wait a bit while suggestion debounce will fire fetch event
        await delay(600)

        // Pick first suggestions
        await driver.page.keyboard.press(Key.ArrowDown)
        await driver.page.keyboard.press(Key.Enter)

        // Change insight title
        await clearAndType(driver, 'input[name="title"]', 'Test insight title')

        // Edit first insight series
        await driver.page.click(
            '[data-testid="form-series"] [data-testid="series-card"]:nth-child(1) [data-testid="series-edit-button"]'
        )
        await clearAndType(
            driver,
            '[data-testid="series-form"]:nth-child(1) input[name="seriesName"]',
            'test edited series title'
        )

        {
            const editor = await createEditorAPI(driver, '[data-testid="series-form"]:nth-child(1) .test-query-input')
            await editor.replace('test edited series query')
        }

        await driver.page.click('[data-testid="series-form"]:nth-child(1) label[title="Cyan"]')

        // Remove second insight series
        await driver.page.click(
            '[data-testid="form-series"] [data-testid="series-card"] [data-testid="series-delete-button"]'
        )

        // Add new series
        await driver.page.click('[data-testid="form-series"] [data-testid="add-series-button"]')
        await clearAndType(
            driver,
            '[data-testid="series-form"]:nth-child(2) input[name="seriesName"]',
            'new test series title'
        )

        {
            const editor = await createEditorAPI(driver, '[data-testid="series-form"]:nth-child(2) .test-query-input')
            await editor.replace('new test series query')
        }

        // Change insight Granularity
        await driver.page.type('input[name="stepValue"]', '2')
        await driver.page.click('input[name="step"][value="days"]')

        const editInsightMutationVariables = await testContext.waitForGraphQLRequest(async () => {
            await driver.page.click('[data-testid="insight-save-button"]')
        }, 'UpdateLineChartSearchInsight')

        // FE works in a way that it adds a synthetic FE runtime id to any newly created series
        // this id uses a following naming rule - runtime-series.<UUID>. In order to stable this test check
        // we remove UUID part of the runtime series
        const newlyCreateSeriesId = editInsightMutationVariables.input.dataSeries[1].seriesId

        editInsightMutationVariables.input.dataSeries[1].seriesId = newlyCreateSeriesId?.match('runtime-series.')
            ? 'runtime-series'
            : newlyCreateSeriesId

        // Check that new org settings config has edited insight
        assert.deepStrictEqual(editInsightMutationVariables, {
            input: {
                repositoryScope: {
                    repositories: ['github.com/sourcegraph/sourcegraph', 'github.com/sourcegraph/about'],
                    repositoryCriteria: null,
                },
                dataSeries: [
                    {
                        seriesId: '001',
                        query: 'test edited series query',
                        options: {
                            label: 'test edited series title',
                            lineColor: 'var(--oc-cyan-7)',
                        },
                        timeScope: {
                            stepInterval: {
                                unit: 'DAY',
                                value: 62,
                            },
                        },
                    },
                    {
                        seriesId: 'runtime-series',
                        query: 'new test series query',
                        options: {
                            label: 'new test series title',
                            lineColor: 'var(--oc-grape-7)',
                        },
                        timeScope: {
                            stepInterval: {
                                unit: 'DAY',
                                value: 62,
                            },
                        },
                    },
                ],
                presentationOptions: {
                    title: 'Test insight title',
                },
                viewControls: {
                    filters: {
                        excludeRepoRegex: '',
                        includeRepoRegex: '',
                        searchContexts: [],
                    },
                    seriesDisplayOptions: {
                        limit: null,
                        numSamples: null,
                        sortOptions: {
                            direction: 'DESC',
                            mode: 'RESULT_COUNT',
                        },
                    },
                },
            },
            id: '001',
        })
    })

    it.skip('should open the edit page with pre-filled fields with values from user/org settings', async () => {
        // Mock `Date.now` to stabilize timestamps
        await driver.page.evaluateOnNewDocument(() => {
            const mockDate = new Date('June 1, 2021 00:00:00 UTC')
            const offset = mockDate.getTimezoneOffset() * 60 * 1000
            // Number of ms between Unix epoch and June 31, 2021
            const mockMs = mockDate.getTime() + offset
            Date.now = () => mockMs
        })

        overrideInsightsGraphQLApi({
            testContext,
            overrides: {
                // Mocks live preview chart
                GetInsightPreview: () => SEARCH_INSIGHT_LIVE_PREVIEW_FIXTURE,

                // Mock for repository suggest component
                RepositorySearchSuggestions: () => ({
                    repositories: { nodes: [] },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/all')
        await driver.page.waitForSelector('[data-testid="line-chart__content"] svg circle')

        // Click on edit button of insight context menu (three dots-menu)
        await driver.page.click(
            '[data-testid="insight-card.searchInsights.insight.graphQLTypesMigration"] [data-testid="InsightContextMenuButton"]'
        )
        await driver.page.click(
            '[data-testid="context-menu.searchInsights.insight.graphQLTypesMigration"] [data-testid="InsightContextMenuEditLink"]'
        )

        // Check redirect URL for edit insight page
        assert.strictEqual(
            driver.page.url().endsWith('/insights/edit/searchInsights.insight.graphQLTypesMigration?dashboardId=all'),
            true
        )

        // Waiting for all important part of creation form will be rendered.
        await driver.page.waitForSelector('[data-testid="search-insight-edit-page-content"]')
        await driver.page.waitForSelector(
            '[data-testid="line-chart__content"] [data-line-name="Imports of new graphql-operations types"] circle'
        )
        await accessibilityAudit(driver.page)

        // Gather all filled inputs within a creation UI form.
        const grabbedInsightInfo = await driver.page.evaluate(getInsightFormValues)

        assert.deepStrictEqual(grabbedInsightInfo, {
            title: 'Migration to new GraphQL TS types',
            repositories: 'github.com/sourcegraph/sourcegraph',
            visibility: 'TestUserID',
            series: [
                {
                    name: 'Imports of old GQL.* types',
                    query: 'patternType:regex case:yes \\*\\sas\\sGQL',
                    stroke: 'var(--oc-red-7)',
                },
                {
                    name: 'Imports of new graphql-operations types',
                    query: "patternType:regexp case:yes /graphql-operations'",
                    stroke: 'var(--oc-blue-7)',
                },
            ],
            step: {
                weeks: '6',
            },
        })
    })
})
