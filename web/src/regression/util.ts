import { createDriverForTest, Driver } from '../../../shared/src/e2e/driver'

export function regressionTestInit(): void {
    // 1 minute test timeout. This must be greater than the default Puppeteer
    // command timeout of 30s in order to get the stack trace to point to the
    // Puppeteer command that failed instead of a cryptic Jest test timeout
    // location.
    jest.setTimeout(1 * 60 * 1000)

    process.on('unhandledRejection', error => {
        console.error('Caught unhandledRejection:', error)
    })

    process.on('rejectionHandled', error => {
        console.error('Caught rejectionHandled:', error)
    })
}

export async function createAndInitializeDriver(): Promise<Driver> {
    const driver = await createDriverForTest()
    await driver.ensureLoggedIn()
    return driver
}
