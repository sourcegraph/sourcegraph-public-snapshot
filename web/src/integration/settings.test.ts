import { createDriverForTest } from '../../../shared/src/testing/driver'
import MockDate from 'mockdate'
import { getConfig } from '../../../shared/src/testing/config'
import assert from 'assert'
import expect from 'expect'
import { ExternalServiceKind } from '../../../shared/src/graphql/schema'
import { describeIntegration } from './helpers'
i

describeIntegration('Search', ({ initGeneration, describe }) => {
    initGeneration(async () => {
        // Reset date mocking
        MockDate.reset()
        const { gitHubToken, sourcegraphBaseUrl, headless, slowMo, testUserPassword } = getConfig(
            'gitHubToken',
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
        const repoSlugs = ['gorilla/mux', 'sourcegraph/jsonrpc2']
        await driver.ensureLoggedIn({ username: 'test', password: testUserPassword, email: 'test@test.com' })
        await driver.resetUserSettings()
        await driver.ensureHasExternalService({
            kind: ExternalServiceKind.GITHUB,
            displayName: 'e2e-test-github',
            config: JSON.stringify({
                url: 'https://github.com',
                token: gitHubToken,
                repos: repoSlugs,
            }),
            ensureRepos: repoSlugs.map(slug => `github.com/${slug}`),
        })
        return { driver, sourcegraphBaseUrl }
    })

    describe('Settings', ({ test }) => {
        test('2.2.1 User settings are saved and applied', async ({
            driver,
            sourcegraphBaseUrl,
            overrideGql,
            waitForRequest,
        }) => {
            overrideGql({
                FetchSettings: once((request, response) => {
                    const resp: GqlSpecificResposense = {
                        inline: 'this is just an object from whatever',
                        oneMoreTHing: true,
                    }

                    // har
                    // gql.json
                    // [{ req: {gqlReq}, resps}]
                }),
            })

            const getSettings = async () => {
                await driver.page.waitForSelector('.e2e-settings-file .monaco-editor')
                return driver.page.evaluate(
                    () => document.querySelector<HTMLElement>('.e2e-settings-file .monaco-editor')?.textContent
                )
            }

            await driver.page.goto(sourcegraphBaseUrl + '/users/test/settings')

            const newSettings = '{ /* These are new settings */}'

            expect(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.e2e-save-toolbar-save')?.disabled
                )
            ).toBe(true)

            const currentSettings = await getSettings()
            await driver.replaceText({
                selector: '.e2e-settings-file .monaco-editor',
                newText: newSettings,
                selectMethod: 'keyboard',
                enterTextMethod: 'type',
            })
            expect(currentSettings).toEqual(newSettings)

            expect(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLButtonElement>('.e2e-save-toolbar-save')?.disabled
                )
            ).toBe(false)

            await driver.findElementWithText('Save changes', { action: 'click' })

            // Wait/assert request
            await waitForRequest('OverrideSettings', {
                variables: { newValue: newSettings },
                response: {},
            })

            // await driver.page.waitForFunction(
            //     () => document.evaluate("//*[text() = ' Saving...']", document).iterateNext() === null
            // )
            // await driver.page.reload()

            // const currentSettings2 = await getSettings()
            // if (JSON.stringify(currentSettings2) !== JSON.stringify(newSettings)) {
            //     throw new Error(
            //         `Settings ${JSON.stringify(currentSettings2)} did not match (new) saved settings ${JSON.stringify(
            //             newSettings
            //         )}`
            //     )
            // }

            // // When you type (or paste) "{" into the empty user settings editor it adds a "}". That's why
            // // we cannot type all the previous text, because then we would have two "}" at the end.
            // const textToTypeFromPrevious = previousSettings.replace(/}$/, '')
            // // Restore old settings
            // await driver.replaceText({
            //     selector: '.e2e-settings-file .monaco-editor',
            //     newText: textToTypeFromPrevious,
            //     selectMethod: 'keyboard',
            //     enterTextMethod: 'paste',
            // })
            // await driver.findElementWithText('Save changes', { action: 'click' })
            // await driver.page.waitForFunction(
            //     () => document.evaluate("//*[text() = ' Saving...']", document).iterateNext() === null
            // )
            // const previousSettings2 = await getSettings()
            // await driver.page.reload()

            // const currentSettings3 = await getSettings()
            // if (currentSettings3 !== previousSettings2) {
            //     throw new Error(
            //         `Settings ${JSON.stringify(currentSettings3)} did not match (old) saved settings ${JSON.stringify(
            //             previousSettings2
            //         )}`
            //     )
            // }
        })
    })
})
