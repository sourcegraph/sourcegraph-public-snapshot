import { Driver, createDriverForTest } from '../../../shared/src/testing/driver'
import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { afterEachSaveScreenshotIfFailed } from '../../../shared/src/testing/screenshotReporter'
import { commonWebGraphQlResults } from './graphQlResults'
import { siteID, siteGQLID } from './jscontext'
import assert from 'assert'
import expect from 'expect'
import { SearchResult } from '../graphql-operations'

const searchResults = (): SearchResult => ({
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
            dynamicFilters: [
                {
                    value: 'archived:yes',
                    label: 'archived:yes',
                    count: 5,
                    limitHit: true,
                    kind: 'repo',
                },
                {
                    value: 'fork:yes',
                    label: 'fork:yes',
                    count: 46,
                    limitHit: true,
                    kind: 'repo',
                },
                {
                    value: 'repo:^github\\.com/Algorilla/manta-ray$',
                    label: 'github.com/Algorilla/manta-ray',
                    count: 1,
                    limitHit: false,
                    kind: 'repo',
                },
            ],
            results: [
                {
                    __typename: 'Repository',
                    id: 'UmVwb3NpdG9yeTozODcxOTM4Nw==',
                    name: 'github.com/Algorilla/manta-ray',
                    label: {
                        html:
                            '\u003Cp\u003E\u003Ca href="/github.com/Algorilla/manta-ray" rel="nofollow"\u003Egithub.com/Algorilla/manta-ray\u003C/a\u003E\u003C/p\u003E\n',
                    },
                    url: '/github.com/Algorilla/manta-ray',
                    icon:
                        'data:image/svg+xml;base64,PHN2ZyB2ZXJzaW9uPSIxLjEiIGlkPSJMYXllcl8xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIKCSB2aWV3Qm94PSIwIDAgNjQgNjQiIHN0eWxlPSJlbmFibGUtYmFja2dyb3VuZDpuZXcgMCAwIDY0IDY0OyIgeG1sOnNwYWNlPSJwcmVzZXJ2ZSI+Cjx0aXRsZT5JY29ucyA0MDA8L3RpdGxlPgo8Zz4KCTxwYXRoIGQ9Ik0yMywyMi40YzEuMywwLDIuNC0xLjEsMi40LTIuNHMtMS4xLTIuNC0yLjQtMi40Yy0xLjMsMC0yLjQsMS4xLTIuNCwyLjRTMjEuNywyMi40LDIzLDIyLjR6Ii8+Cgk8cGF0aCBkPSJNMzUsMjYuNGMxLjMsMCwyLjQtMS4xLDIuNC0yLjRzLTEuMS0yLjQtMi40LTIuNHMtMi40LDEuMS0yLjQsMi40UzMzLjcsMjYuNCwzNSwyNi40eiIvPgoJPHBhdGggZD0iTTIzLDQyLjRjMS4zLDAsMi40LTEuMSwyLjQtMi40cy0xLjEtMi40LTIuNC0yLjRzLTIuNCwxLjEtMi40LDIuNFMyMS43LDQyLjQsMjMsNDIuNHoiLz4KCTxwYXRoIGQ9Ik01MCwxNmgtMS41Yy0wLjMsMC0wLjUsMC4yLTAuNSwwLjV2MzVjMCwwLjMtMC4yLDAuNS0wLjUsMC41aC0yN2MtMC41LDAtMS0wLjItMS40LTAuNmwtMC42LTAuNmMtMC4xLTAuMS0wLjEtMC4yLTAuMS0wLjQKCQljMC0wLjMsMC4yLTAuNSwwLjUtMC41SDQ0YzEuMSwwLDItMC45LDItMlYxMmMwLTEuMS0wLjktMi0yLTJIMTRjLTEuMSwwLTIsMC45LTIsMnYzNi4zYzAsMS4xLDAuNCwyLjEsMS4yLDIuOGwzLjEsMy4xCgkJYzEuMSwxLjEsMi43LDEuOCw0LjIsMS44SDUwYzEuMSwwLDItMC45LDItMlYxOEM1MiwxNi45LDUxLjEsMTYsNTAsMTZ6IE0xOSwyMGMwLTIuMiwxLjgtNCw0LTRjMS40LDAsMi44LDAuOCwzLjUsMgoJCWMxLjEsMS45LDAuNCw0LjMtMS41LDUuNFYzM2MxLTAuNiwyLjMtMC45LDQtMC45YzEsMCwyLTAuNSwyLjgtMS4zQzMyLjUsMzAsMzMsMjkuMSwzMywyOHYtMC42Yy0xLjItMC43LTItMi0yLTMuNQoJCWMwLTIuMiwxLjgtNCw0LTRjMi4yLDAsNCwxLjgsNCw0YzAsMS41LTAuOCwyLjctMiwzLjVoMGMtMC4xLDIuMS0wLjksNC40LTIuNSw2Yy0xLjYsMS42LTMuNCwyLjQtNS41LDIuNWMtMC44LDAtMS40LDAuMS0xLjksMC4zCgkJYy0wLjIsMC4xLTEsMC44LTEuMiwwLjlDMjYuNiwzOCwyNywzOC45LDI3LDQwYzAsMi4yLTEuOCw0LTQsNHMtNC0xLjgtNC00YzAtMS41LDAuOC0yLjcsMi0zLjRWMjMuNEMxOS44LDIyLjcsMTksMjEuNCwxOSwyMHoiLz4KPC9nPgo8L3N2Zz4K',
                    detail: { html: '\u003Cp\u003ERepository name match\u003C/p\u003E\n' },
                    matches: [],
                },
            ],
            alert: null,
            elapsedMilliseconds: 103,
        },
    },
})

describe.only('Search onboarding', () => {
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
                    suggestions: [],
                },
            }),
            Search: searchResults,
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
                                contents: JSON.stringify({ experimentalFeatures: { showOnboardingTour: true } }),
                            },
                        },
                        {
                            __typename: 'Site',
                            id: siteGQLID,
                            siteID,
                            latestSettings: {
                                id: 470,
                                contents: JSON.stringify({ experimentalFeatures: { showOnboardingTour: true } }),
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

    describe('Onboarding', () => {
        it('displays all steps in the language onboarding flow', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.tour-card')
            await driver.page.waitForSelector('.test-tour-language-button')
            await driver.page.click('.test-tour-language-button')
            await driver.page.waitForSelector('#monaco-query-input')
            const inputContents = await driver.page.evaluate(
                () => document.querySelector('#monaco-query-input .view-lines')?.textContent
            )
            assert.strictEqual(inputContents, 'lang:')

            await driver.page.waitForSelector('.test-tour-step-2')
            await driver.page.keyboard.type('typescript')
            await driver.page.keyboard.press('Space')
            await driver.page.waitForSelector('.test-tour-step-3')
            await driver.page.waitForSelector('.test-tour-language-example')
            await driver.page.click('.test-tour-language-example')

            await driver.page.waitForSelector('.test-tour-step-4')
            await driver.page.click('.test-search-button')
            await driver.assertWindowLocation(
                '/search?q=lang:typescript+try%7B:%5Bmy_match%5D%7D&patternType=structural&onboardingTour=true'
            )
            await driver.page.waitForSelector('.test-tour-step-5')
            await driver.page.waitForSelector('.test-tour-structural-next-button')
            await driver.page.click('.test-tour-structural-next-button')
            await driver.page.waitForSelector('.test-tour-step-6')
            await driver.page.click('.test-search-help-dropdown-button-icon')
        })

        it('displays all steps in the repo onboarding flow', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.evaluate(() => {
                localStorage.setItem('has-seen-onboarding-tour', 'false')
                localStorage.setItem('has-cancelled-onboarding-tour', 'false')
                location.reload()
            })
            await driver.page.waitForSelector('.tour-card')
            await driver.page.waitForSelector('.test-tour-repo-button')
            await driver.page.click('.test-tour-repo-button')
            await driver.page.waitForSelector('#monaco-query-input')
            const inputContents = await driver.page.evaluate(
                () => document.querySelector('#monaco-query-input .view-lines')?.textContent
            )
            assert.strictEqual(inputContents, 'repo:')

            await driver.page.waitForSelector('.test-tour-step-2')
            await driver.page.keyboard.type('sourcegraph ')
            await driver.page.waitForSelector('.test-tour-step-3')
            await driver.page.keyboard.type('test')
            await driver.page.waitForSelector('.test-tour-step-4')
            await driver.page.click('.test-search-button')
            await driver.assertWindowLocation('/search?q=repo:sourcegraph+test&patternType=literal&onboardingTour=true')
            await driver.page.waitForSelector('.test-tour-step-5')
            await driver.page.click('.test-search-help-dropdown-button-icon')
        })
        it('advances filter-lang only after the autocomplete is closed and there is whitespace after the filter', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.evaluate(() => {
                localStorage.setItem('has-seen-onboarding-tour', 'false')
                localStorage.setItem('has-cancelled-onboarding-tour', 'false')
                location.reload()
            })
            await driver.page.waitForSelector('.tour-card')
            await driver.page.waitForSelector('.test-tour-language-button')
            await driver.page.click('.test-tour-language-button')
            await driver.page.waitForSelector('#monaco-query-input')
            const inputContents = await driver.page.evaluate(
                () => document.querySelector('#monaco-query-input .view-lines')?.textContent
            )
            assert.strictEqual(inputContents, 'lang:')
            await driver.page.waitForSelector('.test-tour-step-2')
            await driver.page.keyboard.type('java')
            let tourStep2 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-2'))
            let tourStep3 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-3'))
            expect(tourStep2).toBeTruthy()
            expect(tourStep3).toBeNull()
            await driver.page.keyboard.type('script')
            await driver.page.keyboard.press('Tab')
            await driver.page.keyboard.press('Space')
            await driver.page.waitForSelector('.test-tour-step-3')
            tourStep3 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-3'))
            tourStep2 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-2'))
            expect(tourStep3).toBeTruthy()
        })
        it('advances filter-repository only if there is whitespace after the repo filter', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.evaluate(() => {
                localStorage.setItem('has-seen-onboarding-tour', 'false')
                localStorage.setItem('has-cancelled-onboarding-tour', 'false')
                location.reload()
            })
            await driver.page.waitForSelector('.tour-card')
            await driver.page.waitForSelector('.test-tour-repo-button')
            await driver.page.click('.test-tour-repo-button')
            await driver.page.waitForSelector('#monaco-query-input')
            const inputContents = await driver.page.evaluate(
                () => document.querySelector('#monaco-query-input .view-lines')?.textContent
            )
            assert.strictEqual(inputContents, 'repo:')
            await driver.page.waitForSelector('.test-tour-step-2')
            await driver.page.keyboard.type('sourcegraph')
            let tourStep2 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-2'))
            let tourStep3 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-3'))
            expect(tourStep2).toBeTruthy()
            expect(tourStep3).toBeNull()
            await driver.page.keyboard.press('Space')
            await driver.page.waitForSelector('.test-tour-step-3')
            tourStep3 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-3'))
            tourStep2 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-2'))
            expect(tourStep3).toBeTruthy()
        })
    })
})
