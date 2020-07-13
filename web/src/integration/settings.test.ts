import { createDriverForTest } from '../../../shared/src/testing/driver'
import MockDate from 'mockdate'
import { getConfig } from '../../../shared/src/testing/config'
import assert from 'assert'
import { IOrgConnection, IUserEmail, IOrg } from '../../../shared/src/graphql/schema'
import { describeIntegration } from './helpers'
import { retry } from '../../../shared/src/testing/utils'
import { commonGraphQlResults, testUserID, settingsID } from './graphQlResults'

describeIntegration('Settings', ({ initGeneration, describe }) => {
    initGeneration(async () => {
        // Reset date mocking
        MockDate.reset()
        const { sourcegraphBaseUrl, headless, slowMo, testUserPassword } = getConfig(
            'sourcegraphBaseUrl',
            'headless',
            'slowMo',
            'testUserPassword'
        )

        // Start browser
        const driver = await createDriverForTest({
            sourcegraphBaseUrl,
            logBrowserConsole: true,
            headless,
            slowMo,
        })
        await driver.ensureLoggedIn({ username: 'test', password: testUserPassword, email: 'test@test.com' })
        return { driver, sourcegraphBaseUrl }
    })

    describe('User settings page', ({ it }) => {
        it('updates user settings', async ({ driver, sourcegraphBaseUrl, overrideGraphQL, waitForGraphQLRequest }) => {
            overrideGraphQL({
                ...commonGraphQlResults,
                SettingsCascade: () => ({
                    settingsSubject: {
                        settingsCascade: {
                            subjects: [
                                {
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
                User: () => ({
                    user: {
                        __typename: 'User',
                        id: testUserID,
                        username: 'test',
                        displayName: null,
                        url: '/users/test',
                        settingsURL: '/users/test/settings',
                        avatarURL: null,
                        viewerCanAdminister: true,
                        siteAdmin: true,
                        builtinAuth: true,
                        createdAt: '2020-03-02T11:52:15Z',
                        emails: [{ email: 'test@sourcegraph.test', verified: true } as IUserEmail],
                        organizations: { nodes: [] as IOrg[] } as IOrgConnection,
                        permissionsInfo: null,
                    },
                }),
            })

            const getSettingsEditorContent = async (): Promise<string | null | undefined> => {
                await driver.page.waitForSelector('.e2e-settings-file .monaco-editor .view-lines')
                return driver.page.evaluate(
                    () =>
                        document
                            .querySelector<HTMLElement>('.e2e-settings-file .monaco-editor .view-lines')
                            ?.textContent?.replace(/\u00A0/g, ' ') // Monaco replaces all spaces with &nbsp;
                )
            }

            await driver.page.goto(sourcegraphBaseUrl + '/users/test/settings')

            await driver.page.waitForSelector('.e2e-settings-file .monaco-editor')
            await driver.page.waitForSelector('.e2e-save-toolbar-save')

            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.e2e-save-toolbar-save')?.disabled
                ),
                true,
                'Expected save button to be disabled'
            )

            // Replace with new settings
            const newSettings = '{ /* These are new settings */}'
            await driver.replaceText({
                selector: '.e2e-settings-file .monaco-editor .view-lines',
                newText: newSettings,
                selectMethod: 'keyboard',
                enterTextMethod: 'type',
            })
            await retry(async () => {
                const currentSettings = await getSettingsEditorContent()
                assert.strictEqual(currentSettings, newSettings)
            })

            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.e2e-save-toolbar-save')?.disabled
                ),
                false,
                'Expected save button to not be disabled'
            )

            // Assert mutation is done when save button is clicked
            const overrideSettingsVariables = await waitForGraphQLRequest(async () => {
                await driver.findElementWithText('Save changes', { action: 'click' })
            }, 'OverwriteSettings')

            assert.deepStrictEqual(overrideSettingsVariables, {
                contents: newSettings,
                lastID: settingsID,
                subject: testUserID,
            })
        })
    })
})
