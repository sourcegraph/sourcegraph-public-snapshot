import assert from 'assert'

import { afterEach, beforeEach, describe, it } from 'mocha'

import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { settingsID, testUserID } from '@sourcegraph/shared/src/testing/integration/graphQlResults'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'
import { retry } from '@sourcegraph/shared/src/testing/utils'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { createEditorAPI, isElementDisabled, percySnapshotWithVariants } from './utils'

describe('Settings', () => {
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
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    describe('User settings page', () => {
        it('updates user settings', async () => {
            testContext.overrideGraphQL({
                ...commonWebGraphQlResults,
                SettingsCascade: () => ({
                    settingsSubject: {
                        settingsCascade: {
                            final: '',
                            subjects: [
                                {
                                    __typename: 'User',
                                    id: '123',
                                    settingsURL: '#',
                                    viewerCanAdminister: true,
                                    username: 'testuser',
                                    displayName: 'Test User',
                                    latestSettings: {
                                        id: settingsID,
                                        contents: JSON.stringify({}),
                                    },
                                },
                            ],
                        },
                    },
                }),
                OverwriteSettings: () => ({
                    settingsMutation: {
                        overwriteSettings: {
                            empty: {
                                alwaysNil: null,
                            },
                        },
                    },
                }),
                UserAreaUserProfile: () => ({
                    user: {
                        __typename: 'User',
                        id: testUserID,
                        username: 'test',
                        displayName: null,
                        url: '/users/test',
                        settingsURL: '/users/test/settings',
                        avatarURL: null,
                        viewerCanAdminister: true,
                        builtinAuth: true,
                        createdAt: '2020-03-02T11:52:15Z',
                        roles: {
                            __typename: 'RoleConnection',
                            nodes: [],
                        },
                    },
                }),
                UserSettingsAreaUserProfile: () => ({
                    node: {
                        __typename: 'User',
                        id: testUserID,
                        username: 'test',
                        displayName: null,
                        url: '/users/test',
                        settingsURL: '/users/test/settings',
                        avatarURL: null,
                        viewerCanAdminister: true,
                        viewerCanChangeUsername: true,
                        siteAdmin: true,
                        builtinAuth: true,
                        createdAt: '2020-03-02T11:52:15Z',
                        emails: [{ email: 'test@sourcegraph.test', verified: true, isPrimary: true }],
                        organizations: { nodes: [] },
                        permissionsInfo: null,
                        scimControlled: false,
                        roles: {
                            __typename: 'RoleConnection',
                            nodes: [],
                        },
                    },
                }),
            })

            await driver.page.goto(driver.sourcegraphBaseUrl + '/users/test/settings')

            await driver.page.waitForSelector('.test-save-toolbar-save')

            assert.strictEqual(
                await isElementDisabled(driver, '.test-save-toolbar-save'),
                true,
                'Expected save button to be disabled'
            )

            // The editor API needs to be created before taking the screenshot
            // (waits for the editor to be ready)
            const editor = await createEditorAPI(driver, '.test-settings-file .test-editor')

            await percySnapshotWithVariants(driver.page, 'Settings page')
            await accessibilityAudit(driver.page)

            // Replace with new settings
            const newSettings = '{ /* These are new settings */}'
            await editor.replace(newSettings, 'paste')
            await retry(async () => {
                const currentSettings = await editor.getValue()
                assert.strictEqual(currentSettings, newSettings)
            })

            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.test-save-toolbar-save')?.disabled
                ),
                false,
                'Expected save button to not be disabled'
            )

            // Assert mutation is done when save button is clicked
            const overrideSettingsVariables = await testContext.waitForGraphQLRequest(async () => {
                await driver.findElementWithText('Save', { action: 'click' })
            }, 'OverwriteSettings')

            assert.deepStrictEqual(overrideSettingsVariables, {
                contents: newSettings,
                lastID: settingsID,
                subject: testUserID,
            })
        })
    })
})
