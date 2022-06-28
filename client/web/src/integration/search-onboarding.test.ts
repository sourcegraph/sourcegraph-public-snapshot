import assert from 'assert'

import expect from 'expect'

import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults, createViewerSettingsGraphQLOverride } from './graphQlResults'
import { createEditorAPI, EditorAPI, enableEditor, withSearchQueryInput } from './utils'

describe('Search onboarding', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())

    withSearchQueryInput((editorName, editorSelector) => {
        describe(`Onboarding (${editorName})`, () => {
            let editor: EditorAPI
            let testContext: WebIntegrationTestContext

            beforeEach(async function () {
                editor = createEditorAPI(driver, editorName, editorSelector)

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
                    ...createViewerSettingsGraphQLOverride({
                        user: {
                            experimentalFeatures: {
                                showOnboardingTour: true,
                                ...enableEditor(editorName).experimentalFeatures,
                            },
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

            it('only diplay tour after input is focused', async () => {
                await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
                let tourCard = await driver.page.evaluate(() => document.querySelector('.tour-card'))
                expect(tourCard).toBeNull()
                await editor.focus()
                tourCard = await driver.page.evaluate(() => document.querySelector('.tour-card'))
                expect(tourCard).toBeTruthy()
            })

            it('displays all steps in the language onboarding flow', async () => {
                await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
                await editor.focus()
                await driver.page.waitForSelector('.tour-card')
                await driver.page.waitForSelector('[data-testid="tour-language-button"]')
                await driver.page.click('[data-testid="tour-language-button"]')
                const inputContents = await editor.getValue()
                assert.strictEqual(inputContents, 'lang:')

                await driver.page.waitForSelector('.test-tour-step-2')
                await driver.page.keyboard.type('typesc')
                await editor.waitForSuggestion()
                await driver.page.keyboard.press('Tab')
                await driver.page.waitForSelector('.test-tour-step-3')
                await driver.page.keyboard.press('Space')
                await driver.page.keyboard.type('test')

                await driver.page.waitForSelector('.test-tour-step-4')
                await driver.page.click('.test-search-button')
            })

            it('displays all steps in the repo onboarding flow', async () => {
                await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
                await editor.focus()
                await driver.page.waitForSelector('.tour-card')
                await driver.page.waitForSelector('[data-testid="tour-repo-button"]')
                await driver.page.click('[data-testid="tour-repo-button"]')
                const inputContents = await editor.getValue()
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
                await editor.focus()
                await driver.page.waitForSelector('.tour-card')
                await driver.page.waitForSelector('[data-testid="tour-language-button"]')
                await driver.page.click('[data-testid="tour-language-button"]')
                const inputContents = await editor.getValue()
                assert.strictEqual(inputContents, 'lang:')

                await driver.page.waitForSelector('.test-tour-step-2')
                await driver.page.keyboard.type('TypeScr')
                await editor.waitForSuggestion()
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
                await editor.focus()
                await driver.page.waitForSelector('.tour-card')
                await driver.page.waitForSelector('[data-testid="tour-repo-button"]')
                await driver.page.click('[data-testid="tour-repo-button"]')
                const inputContents = await editor.getValue()
                assert.strictEqual(inputContents, 'repo:')

                await driver.page.waitForSelector('.test-tour-step-2')
                await driver.page.keyboard.type('sourcegraph')
                await editor.waitForSuggestion()
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
                await editor.focus()
                await driver.page.waitForSelector('.tour-card')
                await driver.page.waitForSelector('[data-testid="tour-repo-button"]')
                await driver.page.click('[data-testid="tour-repo-button"]')
                const inputContents = await editor.getValue()
                assert.strictEqual(inputContents, 'repo:')

                await driver.page.waitForSelector('.test-tour-step-2')
                await driver.page.keyboard.type('sourcegraph/sourcegraph')
                await editor.waitForSuggestion()
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
})
