import { getTestTools } from './util/init'
import { getConfig } from '../../../shared/src/e2e/config'
import { editSiteConfig } from './util/helpers'
import * as jsoncEdit from '@sqs/jsonc-parser/lib/edit'
import * as jsonc from '@sqs/jsonc-parser'
import { Driver } from '../../../shared/src/e2e/driver'
import { TestResourceManager } from './util/TestResourceManager'
import { retry } from '../../../shared/src/e2e/e2e-test-utils'
import { BuiltinAuthProvider, SiteConfiguration } from '../schema/site.schema'
import { fetchSiteConfiguration } from './util/api'
import { GraphQLClient } from './util/GraphQLClient'

/**
 * @jest-environment node
 */

describe('Site config test suite', () => {
    const formattingOptions = { eol: '\n', insertSpaces: true, tabSize: 2 }
    const config = getConfig(
        'sudoToken',
        'sudoUsername',
        'headless',
        'slowMo',
        'keepBrowser',
        'noCleanup',
        'sourcegraphBaseUrl',
        'testUserPassword',
        'logBrowserConsole'
    )
    let driver: Driver
    let resourceManager: TestResourceManager
    let gqlClient: GraphQLClient
    beforeAll(async () => {
        ;({ driver, resourceManager, gqlClient } = await getTestTools(config))
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
                await editSiteConfig(gqlClient, contents =>
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
                    await driver.page.waitForSelector('#htmlBodyTopContent', { timeout: 1000 })
                },
                { retries: 10 }
            )
        },
        10 * 1000
    )

    test(
        'builtin auth provider: allowSignup',
        async () => {
            const siteConfig = await fetchSiteConfiguration(gqlClient).toPromise()
            const siteConfigParsed: SiteConfiguration = jsonc.parse(siteConfig.configuration.effectiveContents)
            // console.log('# siteConfig', siteConfigParsed)
            const setBuiltinAuthProvider = async (p: BuiltinAuthProvider) => {
                const authProviders = siteConfigParsed['auth.providers']
                const foundIndices =
                    authProviders?.map((p, i) => (p.type === 'builtin' ? i : -1)).filter(i => i !== -1) || []
                const builtinAuthProviderIndex = foundIndices.length > 0 ? foundIndices[0] : -1
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
                return editSiteConfig(gqlClient, ...editFns)
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
