import { createDriverForTest, Driver } from '../../../../shared/src/e2e/driver'

export function regressionTestInit(): void {
    // 10s test timeout. This must be greater than the Puppeteer navigation timeout (set to 5s
    // below) in order to get the stack trace to point to the Puppeteer command that failed instead
    // of a cryptic Jest test timeout location.
    jest.setTimeout(10 * 1000)

    process.on('unhandledRejection', error => {
        console.error('Caught unhandledRejection:', error)
    })

    process.on('rejectionHandled', error => {
        console.error('Caught rejectionHandled:', error)
    })
}

/**
 * Returns a Puppeteer driver with a 5s command timeout. It is important that none of the Jest test
 * timeouts is under 5s. Otherwise, the timeout error will be a cryptic Jest timeout error, instead
 * of an error pointing to the timed-out Puppeteer command.
 */
export async function createAndInitializeDriver(sourcegraphBaseUrl: string): Promise<Driver> {
    const driver = await createDriverForTest({ sourcegraphBaseUrl })
    driver.page.setDefaultNavigationTimeout(5 * 1000) // 5s navigation timeout
    return driver
}
