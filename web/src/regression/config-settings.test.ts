import { getTestTools } from './util/init'
import { getConfig } from '../../../shared/src/e2e/config'
import { editCriticalSiteConfig, getCriticalSiteConfig } from './util/helpers'
import * as jsoncEdit from '@sqs/jsonc-parser/lib/edit'
import * as jsonc from '@sqs/jsonc-parser'
import { Driver } from '../../../shared/src/e2e/driver'
import { TestResourceManager } from './util/TestResourceManager'
import { retry } from '../../../shared/src/e2e/e2e-test-utils'
import { CriticalConfiguration, BuiltinAuthProvider } from '../schema/critical.schema'

/**
 * @jest-environment node
 */

describe('Critical config test suite', () => {
    const formattingOptions = { eol: '\n', insertSpaces: true, tabSize: 2 }
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'headless',
        'slowMo',
        'keepBrowser',
        'noCleanup',
        'sourcegraphBaseUrl',
        'managementConsoleUrl',
        'testUserPassword',
        'logBrowserConsole'
    )
    let driver: Driver
    let resourceManager: TestResourceManager
    beforeAll(async () => {
        ;({ driver, resourceManager } = await getTestTools(config))
    })
    beforeEach(() => {
        resourceManager = new TestResourceManager()
    })
    afterEach(async () => {
        await resourceManager.destroyAll()
    })
    test(
        'htmlBodyTop',
        async () => {
            resourceManager.add(
                'Configuration',
                'htmlBodyTop',
                await editCriticalSiteConfig(config.managementConsoleUrl, contents =>
                    jsoncEdit.setProperty(
                        contents,
                        ['htmlBodyTop'],
                        '<div id="htmlBodyTopContent">TEST</div>',
                        formattingOptions
                    )
                )
            )

            await retry(
                async () => {
                    await driver.page.goto(config.sourcegraphBaseUrl)
                    await driver.page.reload()
                    await driver.page.waitForSelector('#htmlBodyTopContent')
                },
                { retries: 10 }
            )
        },
        10 * 1000
    )

    test(
        'builtin auth provider: allowSignup',
        async () => {
            const criticalConfig = await getCriticalSiteConfig(config.managementConsoleUrl)
            const criticalConfigParsed: CriticalConfiguration = jsonc.parse(criticalConfig.Contents)
            const setBuiltinAuthProvider = async (p: BuiltinAuthProvider) => {
                let builtinAuthProviderIndex = -1
                if (criticalConfigParsed['auth.providers']) {
                    for (let i = 0; i < criticalConfigParsed['auth.providers'].length; i++) {
                        if (criticalConfigParsed['auth.providers'][i].type === 'builtin') {
                            builtinAuthProviderIndex = i
                            break
                        }
                    }
                }
                const editFns = []
                if (builtinAuthProviderIndex !== -1) {
                    editFns.push((contents: string) =>
                        jsoncEdit.setProperty(
                            contents,
                            ['auth.providers', builtinAuthProviderIndex],
                            undefined,
                            formattingOptions
                        )
                    )
                }
                editFns.push((contents: string) =>
                    jsoncEdit.setProperty(contents, ['auth.providers', -1], p, formattingOptions)
                )
                return await editCriticalSiteConfig(config.managementConsoleUrl, ...editFns)
            }

            resourceManager.add(
                'Configuration',
                'builtin auth provider: allowSignup',
                await setBuiltinAuthProvider({ type: 'builtin', allowSignup: false })
            )
            await retry(
                async () => {
                    await driver.page.goto(config.sourcegraphBaseUrl)
                    await driver.page.reload()
                    await driver.findElementWithText('Sign in', { wait: { timeout: 2000 } })
                    expect(await driver.page.evaluate(() => document.body.innerText.includes('Sign up'))).toBeFalsy()
                },
                { retries: 5 } // configuration propagation is eventually consistent
            )

            await setBuiltinAuthProvider({ type: 'builtin', allowSignup: true })
            await retry(
                async () => {
                    await driver.page.goto(config.sourcegraphBaseUrl)
                    await driver.page.reload()
                    await driver.findElementWithText('Sign in', { wait: { timeout: 2000 } })
                    expect(await driver.page.evaluate(() => document.body.innerText.includes('Sign up'))).toBeTruthy()
                },
                { retries: 5 } // configuration propagation is eventually consistent
            )
        },
        20 * 1000
    )
})
