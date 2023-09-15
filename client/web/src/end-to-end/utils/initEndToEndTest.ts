import MockDate from 'mockdate'

import { getConfig } from '@sourcegraph/shared/src/testing/config'
import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'

const { sourcegraphBaseUrl } = getConfig('gitHubDotComToken', 'sourcegraphBaseUrl')

export async function initEndToEndTest(): Promise<Driver> {
    // Reset date mocking
    MockDate.reset()

    const config = getConfig('headless', 'slowMo', 'testUserPassword')

    // Start browser
    const driver = await createDriverForTest({
        sourcegraphBaseUrl,
        logBrowserConsole: true,
        ...config,
    })

    try {
        await driver.ensureSignedIn({ username: 'test', password: config.testUserPassword, email: 'test@test.com' })
        await driver.resetUserSettings()
    } catch (error) {
        console.error('failed to ensureSignedIn', error)
    }

    return driver
}
