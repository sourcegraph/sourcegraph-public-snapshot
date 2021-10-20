import assert from 'assert'

import expect from 'expect'

import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { siteID, siteGQLID } from './jscontext'

describe('Search onboarding', () => {
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
            AutoDefinedSearchContexts: () => ({
                autoDefinedSearchContexts: [],
            }),
            ViewerSettings: () => ({
                viewerSettings: {
                    __typename: 'SettingsCascade',
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
                            allowSiteSettingsEdits: true,
                        },
                    ],
                    final: JSON.stringify({}),
                },
            }),
            SearchSidebarGitRefs: () => ({
                repository: {
                    __typename: 'Repository',
                    id: 'repo',
                    gitRefs: {
                        __typename: 'GitRefConnection',
                        nodes: [],
                        pageInfo: {
                            hasNextPage: false,
                        },
                        totalCount: 0,
                    },
                },
            }),
            GetTemporarySettings: () => ({
                temporarySettings: {
                    __typename: 'TemporarySettings',
                    contents: JSON.stringify({
                        'user.daysActiveCount': 1,
                        'user.lastDayActive': new Date().toDateString(),
                    }),
                },
            }),
        })
        testContext.overrideSearchStreamEvents([
            // Used for suggestions
            {
                type: 'matches',
                data: [{ type: 'repo', repository: '^github\\.com/sourcegraph/sourcegraph$' }],
            },
            { type: 'done', data: {} },
        ])
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    const waitAndFocusInput = async () => {
        await driver.page.waitForSelector('.monaco-editor .view-lines')
        await driver.page.click('.monaco-editor .view-lines')
    }

    describe('Onboarding', () => {
        it('only diplay tour after input is focused', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            let tourCard = await driver.page.evaluate(() => document.querySelector('.tour-card'))
            expect(tourCard).toBeNull()
            await waitAndFocusInput()
            tourCard = await driver.page.evaluate(() => document.querySelector('.tour-card'))
            expect(tourCard).toBeTruthy()
        })

        it('displays all steps in the language onboarding flow', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await waitAndFocusInput()
            await driver.page.waitForSelector('.tour-card')
            await driver.page.waitForSelector('.tour-language-button')
            await driver.page.click('.tour-language-button')
            await driver.page.waitForSelector('#monaco-query-input')
            const inputContents = await driver.page.evaluate(
                () => document.querySelector('#monaco-query-input .view-lines')?.textContent
            )
            assert.strictEqual(inputContents, 'lang:')

            await driver.page.waitForSelector('.test-tour-step-2')
            await driver.page.keyboard.type('typesc')
            await driver.page.waitForSelector('.monaco-query-input .suggest-widget.visible')
            await driver.page.keyboard.press('Tab')
            await driver.page.waitForSelector('.test-tour-step-3')
            await driver.page.keyboard.press('Space')
            await driver.page.keyboard.type('test')

            await driver.page.waitForSelector('.test-tour-step-4')
            await driver.page.click('.test-search-button')
        })

        it('displays all steps in the repo onboarding flow', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await waitAndFocusInput()
            await driver.page.waitForSelector('.tour-card')
            await driver.page.waitForSelector('.tour-repo-button')
            await driver.page.click('.tour-repo-button')
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
        })

        it('advances filter-lang when an autocomplete suggestion is selected', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await waitAndFocusInput()
            await driver.page.waitForSelector('.tour-card')
            await driver.page.waitForSelector('.tour-language-button')
            await driver.page.click('.tour-language-button')
            await driver.page.waitForSelector('#monaco-query-input')
            const inputContents = await driver.page.evaluate(
                () => document.querySelector('#monaco-query-input .view-lines')?.textContent
            )
            assert.strictEqual(inputContents, 'lang:')
            await driver.page.waitForSelector('.test-tour-step-2')
            await driver.page.keyboard.type('TypeScr')
            await driver.page.waitForSelector('.monaco-query-input .suggest-widget.visible')
            let tourStep2 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-2'))
            let tourStep3 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-3'))
            expect(tourStep2).toBeTruthy()
            expect(tourStep3).toBeNull()
            await driver.page.keyboard.press('Tab')
            await driver.page.waitForSelector('.test-tour-step-3')
            tourStep3 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-3'))
            tourStep2 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-2'))
            expect(tourStep3).toBeTruthy()
        })

        it('advances filter-repository when an autocomplete suggestion is selected', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await waitAndFocusInput()
            await driver.page.waitForSelector('.tour-card')
            await driver.page.waitForSelector('.tour-repo-button')
            await driver.page.click('.tour-repo-button')
            await driver.page.waitForSelector('#monaco-query-input')
            const inputContents = await driver.page.evaluate(
                () => document.querySelector('#monaco-query-input .view-lines')?.textContent
            )
            assert.strictEqual(inputContents, 'repo:')
            await driver.page.waitForSelector('.test-tour-step-2')
            await driver.page.keyboard.type('sourcegraph')
            await driver.page.waitForSelector('.monaco-query-input .suggest-widget.visible')
            let tourStep2 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-2'))
            let tourStep3 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-3'))
            expect(tourStep2).toBeTruthy()
            expect(tourStep3).toBeNull()
            await driver.page.keyboard.press('Tab')
            await driver.page.waitForSelector('.test-tour-step-3')
            tourStep3 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-3'))
            tourStep2 = await driver.page.evaluate(() => document.querySelector('.test-tour-step-2'))
            expect(tourStep3).toBeTruthy()
        })

        it('advances filter-repository when a user types their own repository', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await waitAndFocusInput()
            await driver.page.waitForSelector('.tour-card')
            await driver.page.waitForSelector('.tour-repo-button')
            await driver.page.click('.tour-repo-button')
            await driver.page.waitForSelector('#monaco-query-input')
            const inputContents = await driver.page.evaluate(
                () => document.querySelector('#monaco-query-input .view-lines')?.textContent
            )
            assert.strictEqual(inputContents, 'repo:')
            await driver.page.waitForSelector('.test-tour-step-2')
            await driver.page.keyboard.type('sourcegraph/sourcegraph')
            await driver.page.waitForSelector('.monaco-query-input .suggest-widget.visible')
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
