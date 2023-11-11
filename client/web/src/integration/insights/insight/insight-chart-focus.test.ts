import assert from 'assert'

import { beforeEach, describe, it } from 'mocha'
import { Key } from 'ts-key-enum'

import { hasFocus } from '@sourcegraph/shared/src/testing/dom-test-helpers'
import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from '../../context'
import {
    GET_INSIGHT_VIEW_SEARCH_BASED_INSIGHT,
    LANG_STAT_INSIGHT_CONTENT,
    LANG_STATS_INSIGHT,
    SEARCH_BASED_INSIGHT,
} from '../fixtures/dashboards'
import { mockDashboardWithInsights } from '../utils/mock-helpers'

describe('Code insights [Insight Card] should has a proper focus management ', () => {
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
    afterEachSaveScreenshotIfFailed(() => driver.page)

    it('for search based (line-chart) insight card', async () => {
        mockDashboardWithInsights({
            testContext,
            dashboardId: 'DASHBOARD_WITH_SEARCH',
            insightMock: SEARCH_BASED_INSIGHT,
            insightViewMock: GET_INSIGHT_VIEW_SEARCH_BASED_INSIGHT,
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/insights/dashboards/DASHBOARD_WITH_SEARCH')

        await driver.page.waitForSelector('[aria-label="Search Based insight"]')
        await driver.page.focus('[aria-label="Search Based insight"]')

        await driver.page.keyboard.press(Key.Tab)
        assert.strictEqual(
            await hasFocus(driver, '[aria-label="Search Based insight"] h2 a'),
            true,
            'Insight title should be focused'
        )

        await driver.page.keyboard.press(Key.Tab)
        assert.strictEqual(await hasFocus(driver, '[aria-label="Filters"]'), true, 'Insight filters should be focused')

        await driver.page.keyboard.press(Key.Tab)
        assert.strictEqual(
            await hasFocus(driver, '[aria-label="Insight options"]'),
            true,
            'Insight context menu should be focused'
        )

        const dataSeries = GET_INSIGHT_VIEW_SEARCH_BASED_INSIGHT.insightViews.nodes[0]?.dataSeries

        if (!dataSeries) {
            assert.fail('Insight errored')
        }

        await driver.page.waitForSelector('[aria-label="Chart series"]')

        // Focus the chart container (SVG root element)
        await driver.page.keyboard.press(Key.Tab)

        // Enter the arrow keys keyboard navigation mode
        await driver.page.keyboard.press(Key.Enter)

        for (let lineIndex = 0; lineIndex <= dataSeries.length - 1; lineIndex++) {
            const series = dataSeries[lineIndex]

            for (let pointIndex = 0; pointIndex < series.points.length; pointIndex++) {
                assert.strictEqual(
                    await hasFocus(
                        driver,
                        `[aria-label="Chart series"] > [role="listitem"]:nth-child(${lineIndex + 1}) a:nth-child(${
                            pointIndex + 1
                        })`
                    ),
                    true,
                    'Insight data point should be focused'
                )

                await driver.page.keyboard.press(Key.ArrowRight)
            }
        }
    })

    it('for lang-stats (pie chart) insight card', async () => {
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

        await driver.page.waitForSelector('[aria-label="Lang Stats insight"]')
        await driver.page.focus('[aria-label="Lang Stats insight"]')

        await driver.page.keyboard.press(Key.Tab)
        assert.strictEqual(
            await hasFocus(driver, '[aria-label="Lang Stats insight"] h2 a'),
            true,
            'Insight title should be focused'
        )

        await driver.page.keyboard.press(Key.Tab)
        assert.strictEqual(
            await hasFocus(driver, '[aria-label="Insight options"]'),
            true,
            'Insight context menu should be focused'
        )

        const arcs = LANG_STAT_INSIGHT_CONTENT.search.stats.languages

        await driver.page.waitForSelector('[aria-label="Pie chart"] a')

        for (let arcIndex = 0; arcIndex < Math.min(arcs.length, 6); arcIndex++) {
            await driver.page.keyboard.press(Key.Tab)

            const aElement = await driver.page.evaluate(
                (arcIndex: number) => document.querySelector(`[aria-label="Pie chart"] g:nth-child(${arcIndex + 1}) a`),
                arcIndex
            )

            const activeElement = await driver.page.evaluate(() => document.activeElement)

            assert.strictEqual(aElement === activeElement, true, 'Insight pie arc should be focused')
        }
    })
})
