import assert from 'assert'

import { afterEach, beforeEach, describe, it } from 'mocha'

import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../../context'
import {
    CAPTURE_GROUP_INSIGHT,
    COMPUTE_INSIGHT,
    GET_INSIGHT_VIEW_CAPTURE_GROUP_INSIGHT,
    GET_INSIGHT_VIEW_COMPUTE_INSIGHT,
    GET_INSIGHT_VIEW_SEARCH_BASED_INSIGHT,
    LANG_STAT_INSIGHT_CONTENT,
    LANG_STATS_INSIGHT,
    SEARCH_BASED_INSIGHT,
} from '../fixtures/dashboards'
import { mockDashboardWithInsights } from '../utils/mock-helpers'

describe('Code insights [Dashboard card]', () => {
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

    it('renders lang stats insight card with proper options context', async () => {
        mockDashboardWithInsights({
            testContext,
            dashboardId: 'DASHBOARD_WITH_LANG_INSIGHT',
            insightMock: LANG_STATS_INSIGHT,
            overrides: {
                // Mock lang stats insight content
                LangStatsInsightContent: () => LANG_STAT_INSIGHT_CONTENT,
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/DASHBOARD_WITH_LANG_INSIGHT')
        await driver.page.waitForSelector('[aria-label="Pie chart"]')

        const numberOfArcs = await driver.page.$$eval('[aria-label="Pie chart"] path', elements => elements.length)
        const numberHeadings = await driver.page.$$eval('[aria-label="Pie chart"] p', elements => elements.length)

        // Why 12?, because LANG_STAT_INSIGHT_CONTENT mock has 14 entries and only five of them
        // are rendered because all other (7 groups) are too small, and they are grouped and presented
        // by special "Other" category (5 original rendered groups + one special Other group = 6)
        // 12 paths because we render 2 paths per pie part (one for the pie arc itself and one for the annotation line)
        assert.strictEqual(numberOfArcs, 12)

        // Why 6?, because LANG_STAT_INSIGHT_CONTENT mock has 14 entries and only five of them
        // are rendered because all other (7 groups) are too small, and they are grouped and presented
        // by special "Other" category (5 original rendered groups + one special Other group = 6)
        assert.strictEqual(numberHeadings, 6)

        await checkOptionsMenu(driver, { yAxis: false, exportData: false })
        await checkFilterMenu(driver, false)
    })

    it('renders capture group insight card with proper options context', async () => {
        mockDashboardWithInsights({
            testContext,
            dashboardId: 'DASHBOARD_WITH_CAPTURE_GROUP',
            insightMock: CAPTURE_GROUP_INSIGHT,
            insightViewMock: GET_INSIGHT_VIEW_CAPTURE_GROUP_INSIGHT,
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/DASHBOARD_WITH_CAPTURE_GROUP')
        await driver.page.waitForSelector('[aria-label="Line chart"]')

        const numberOfLines = await driver.page.$$eval('[aria-label="Line chart"] path', elements => elements.length)
        const numberOfPointLinks = await driver.page.$$eval('[aria-label="Line chart"] a', elements => elements.length)

        // Why 20 and 189? See GET_INSIGHT_VIEW_CAPTURE_GROUP_INSIGHT dataset mock, it has 20 lines and 189 points
        assert.strictEqual(numberOfLines, 20)
        assert.strictEqual(numberOfPointLinks, 189)

        await checkOptionsMenu(driver, { yAxis: false, exportData: true })
        await checkFilterMenu(driver, true)
    })

    it('renders search insight card with proper options context', async () => {
        mockDashboardWithInsights({
            testContext,
            dashboardId: 'DASHBOARD_WITH_SEARCH',
            insightMock: SEARCH_BASED_INSIGHT,
            insightViewMock: GET_INSIGHT_VIEW_SEARCH_BASED_INSIGHT,
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/DASHBOARD_WITH_SEARCH')
        await driver.page.waitForSelector('[aria-label="Line chart"]')
        await driver.page.waitForSelector('[aria-label="Line chart"] path')

        const numberOfLines = await driver.page.$$eval('[aria-label="Line chart"] path', elements => elements.length)
        const numberOfPointLinks = await driver.page.$$eval('[aria-label="Line chart"] a', elements => elements.length)

        // Why 2 and 27? See GET_INSIGHT_VIEW_SEARCH_BASED_INSIGHT dataset mock, it has 2 lines and 27 points
        assert.strictEqual(numberOfLines, 2)
        assert.strictEqual(numberOfPointLinks, 27)

        await checkOptionsMenu(driver, { yAxis: true, exportData: true })
        await checkFilterMenu(driver, true)
    })

    it('renders compute insight card with proper options context', async () => {
        mockDashboardWithInsights({
            testContext,
            dashboardId: 'DASHBOARD_WITH_COMPUTE',
            insightMock: COMPUTE_INSIGHT,
            insightViewMock: GET_INSIGHT_VIEW_COMPUTE_INSIGHT,
            overrides: {
                ViewerSettings: () => ({
                    viewerSettings: {
                        __typename: 'SettingsCascade',
                        subjects: [
                            {
                                __typename: 'DefaultSettings',
                                id: 'DefaultSettingsTestId',
                                settingsURL: null,
                                viewerCanAdminister: false,
                                latestSettings: {
                                    id: 0,
                                    contents: JSON.stringify({
                                        experimentalFeatures: { codeInsightsCompute: true },
                                    }),
                                },
                            },
                        ],
                        final: JSON.stringify({}),
                    },
                }),
            },
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/DASHBOARD_WITH_COMPUTE')
        await driver.page.waitForSelector('[aria-label="Bar chart"]')
        await driver.page.waitForSelector('[aria-label="Bar chart content"]')

        const numberOfBars = await driver.page.$$eval('[aria-label="Bar chart"] rect', elements => elements.length)

        // Why 2? See GET_INSIGHT_VIEW_COMPUTE_INSIGHT dataset mock, it has 1 series
        // Visx also renders a rectangle for the chart background. 1 series + 1 background = 2 rectangles
        assert.strictEqual(numberOfBars, 2)

        await checkOptionsMenu(driver, { yAxis: false, exportData: true })
        await checkFilterMenu(driver, true)
    })
})

interface MenuOptions {
    yAxis: boolean
    exportData: boolean
}

/**
 * Asserts the options menu renders the correct options
 */
async function checkOptionsMenu(driver: Driver, options: MenuOptions): Promise<void> {
    const { yAxis, exportData } = options
    await driver.page.click('[aria-label="Insight options"]')

    const menuOptions = await driver.page.$$eval('[role="dialog"][aria-modal="true"] [role="menuitem"]', elements =>
        elements.map(element => element.textContent)
    )

    if (exportData) {
        // Check that Line chart menu options
        assert.deepStrictEqual(menuOptions, [
            'Edit',
            'Get shareable link',
            'Export data',
            'Remove from this dashboard',
            'Delete',
        ])
    } else {
        assert.deepStrictEqual(menuOptions, ['Edit', 'Get shareable link', 'Remove from this dashboard', 'Delete'])
    }

    if (yAxis) {
        const startYAxisAt0 = await driver.page.$eval(
            '[role="dialog"][aria-modal="true"] [role="menuitemcheckbox"]',
            element => element.textContent
        )

        assert.strictEqual(startYAxisAt0, 'Start Y Axis at 0')
    }
}

/**
 * Asserts the filter menu renders the correct options
 */
async function checkFilterMenu(driver: Driver, shouldHaveFilterButton = true): Promise<void> {
    if (shouldHaveFilterButton) {
        await driver.page.click('[aria-label="Filters"]')
        const filterPanel = await driver.page.$('[aria-label="Drill-down filters panel"]')

        // Should open filter panel on filter panel icon click
        assert.strictEqual(filterPanel !== null, true)
    } else {
        const filtersButton = await driver.page.$('[aria-label="Active filters"], [aria-label="Filters"]')

        // Lang's stats insight doesn't support filters functionality, so we shouldn't have any filter UI
        assert.strictEqual(filtersButton, null)
    }
}
