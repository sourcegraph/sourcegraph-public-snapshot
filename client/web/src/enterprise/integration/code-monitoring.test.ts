import assert from 'assert'
import { commonWebGraphQlResults } from '../../integration/graphQlResults'
import { Driver, createDriverForTest } from '../../../../shared/src/testing/driver'
import { WebIntegrationTestContext, createWebIntegrationTestContext } from '../../integration/context'
import { afterEachSaveScreenshotIfFailed } from '../../../../shared/src/testing/screenshotReporter'
import { siteID, siteGQLID } from '../../integration/jscontext'
import { SearchResult } from '../../graphql-operations'

describe.only('Code monitoring', () => {
    // let driver: Driver
    // before(async () => {
    //     driver = await createDriverForTest()
    // })
    // after(() => driver?.close())
    // let testContext: WebIntegrationTestContext
    // beforeEach(async function () {
    //     testContext = await createWebIntegrationTestContext({
    //         driver,
    //         currentTest: this.currentTest!,
    //         directory: __dirname,
    //     })

    //     testContext.overrideGraphQL({
    //         ...commonWebGraphQlResults,
    //         ViewerSettings: () => ({
    //             viewerSettings: {
    //                 subjects: [
    //                     {
    //                         __typename: 'DefaultSettings',
    //                         settingsURL: null,
    //                         viewerCanAdminister: false,
    //                         latestSettings: {
    //                             id: 0,
    //                             contents: JSON.stringify({ experimentalFeatures: { codeMonitoring: true } }),
    //                         },
    //                     },
    //                     {
    //                         __typename: 'Site',
    //                         id: siteGQLID,
    //                         siteID,
    //                         latestSettings: {
    //                             id: 470,
    //                             contents: JSON.stringify({ experimentalFeatures: { codeMonitoring: true } }),
    //                         },
    //                         settingsURL: '/site-admin/global-settings',
    //                         viewerCanAdminister: true,
    //                     },
    //                 ],
    //                 final: JSON.stringify({}),
    //             },
    //         }),
    //     })
    // })
    // afterEachSaveScreenshotIfFailed(() => driver.page)
    // afterEach(() => testContext?.dispose())
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
            SearchSuggestions: () => ({
                search: {
                    suggestions: [{ __typename: 'Repository', name: '^github\\.com/sourcegraph/sourcegraph$' }],
                },
            }),
            Search: (): SearchResult => ({
                search: {
                    results: {
                        __typename: 'SearchResults',
                        limitHit: true,
                        matchCount: 30,
                        approximateResultCount: '30+',
                        missing: [],
                        cloning: [],
                        repositoriesCount: 372,
                        timedout: [],
                        indexUnavailable: false,
                        dynamicFilters: [],
                        results: [],
                        alert: null,
                        elapsedMilliseconds: 103,
                    },
                },
            }),
            RepoGroups: () => ({
                repoGroups: [],
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
                                    experimentalFeatures: { showOnboardingTour: true, codeMonitoring: true },
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
                                    experimentalFeatures: { showOnboardingTour: true, codeMonitoring: true },
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

    describe('Code monitoring form advances sequentially', () => {
        it('disables the actions area until trigger is complete', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/code-monitoring/new')
            await driver.page.waitForSelector('.test-action-button')

            assert.strictEqual(
                await driver.page.evaluate(() => document.querySelector('.test-action-button')),
                'true',
                'Expected action button to be disabled'
            )

            assert.strictEqual(
                await driver.page.evaluate(() => {
                    document.querySelector('.test-action-button')?.getAttribute('disabled')
                }),
                true,
                'Expected action button to be disabled'
            )

            await driver.page.waitForSelector('.test-trigger-button')
            await driver.page.click('.test-trigger-button')
            await driver.page.waitForSelector('.test-action-input')
            await driver.page.type('.test-trigger-input', 'foobar type:diff patterntype:literal')
            await driver.page.waitForSelector('.test-submit-trigger')
            await driver.page.click('.test-submit-trigger')

            assert.strictEqual(
                await driver.page.evaluate(() => {
                    document.querySelector('.test-action-button')?.getAttribute('disabled')
                }),
                false,
                'Expected action button to be disabled'
            )
        })
    })
})
