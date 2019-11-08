import { getTestTools } from './util/init'
import { getConfig } from '../../../shared/src/e2e/config'
import { editCriticalSiteConfig, ensureLoggedInOrCreateTestUser } from './util/helpers'
import * as jsoncEdit from '@sqs/jsonc-parser/lib/edit'
import { getManagementConsoleState } from './util/api'
import { Driver } from '../../../shared/src/e2e/driver'
import { GraphQLClient } from './util/GraphQLClient'
import { TestResourceManager } from './util/TestResourceManager'
import { retry } from '../../../shared/src/e2e/e2e-test-utils'

/**
 * @jest-environment node
 */

describe('Critical config test suite', () => {
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
    let gqlClient: GraphQLClient
    let resourceManager: TestResourceManager
    beforeAll(async () => {
        ;({ driver, gqlClient, resourceManager } = await getTestTools(config))
    })
    afterAll(async () => {
        await resourceManager.destroyAll()
    })
    test(
        'htmlBodyTop',
        async () => {
            const formattingOptions = { eol: '\n', insertSpaces: true, tabSize: 2 }
            const { plaintextPassword: managementConsolePassword } = await getManagementConsoleState(gqlClient)
            if (!managementConsolePassword) {
                throw new Error('empty management console password')
            }

            resourceManager.add(
                'Configuration',
                'htmlBodyTop',
                await editCriticalSiteConfig(config.managementConsoleUrl, managementConsolePassword, contents =>
                    jsoncEdit.setProperty(
                        contents,
                        ['htmlBodyTop'],
                        '<div id="htmlBodyTopContent">TEST</div>',
                        formattingOptions
                    )
                )
            )

            retry(
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
})
