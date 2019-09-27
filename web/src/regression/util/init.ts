import { createDriverForTest, Driver } from '../../../../shared/src/e2e/driver'
import { saveScreenshotsUponFailuresAndClosePage } from '../../../../shared/src/e2e/screenshotReporter'
import * as path from 'path'
import { getConfig } from '../../../../shared/src/e2e/config'
import { currentProductVersion } from './api'
import { GraphQLClient } from './GraphQLClient'
import * as semver from 'semver'

/**
 * Semver constraint on the Sourcegraph product version. Uses the syntax specified in
 * https://www.npmjs.com/package/semver. This should be updated when a change is made that breaks
 * compatibility between the regression tests and the Sourcegraph GUI. For example, when a new CSS
 * class is added that the regression tests rely on to identify a particular component.
 */
const supportedSourcegraphVersionConstraint = '>=3.8'

/**
 * Sets default timeout and error handlers for regression tests. Includes:
 * - Default Jest test timeout
 * - Top-level rejection handlers
 * - Screenshot on failure
 *
 * This should be called in the top-level `beforeAll` function of each regression test suite,
 * after the driver is initailized.
 */
export async function setTestDefaults(driver: Driver, gqlClient: GraphQLClient): Promise<void> {
    const version = await currentProductVersion(gqlClient)
    if (version !== 'dev' && !semver.satisfies(version, supportedSourcegraphVersionConstraint)) {
        throw new Error(
            `Sourcegraph version ${JSON.stringify(
                version
            )} is unsupported. These tests require a version that satisfies the constraint ${JSON.stringify(
                supportedSourcegraphVersionConstraint
            )} or is "dev"`
        )
    }

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

    // Take a screenshot when a test fails.
    saveScreenshotsUponFailuresAndClosePage(
        path.resolve(__dirname, '..', '..', '..'),
        path.resolve(__dirname, '..', '..', '..', 'puppeteer'),
        () => driver.page
    )
}

/**
 * Returns a Puppeteer driver with a 5s command timeout. It is important that none of the Jest test
 * timeouts is under 5s. Otherwise, the timeout error will be a cryptic Jest timeout error, instead
 * of an error pointing to the timed-out Puppeteer command.
 */
export async function createAndInitializeDriver(sourcegraphBaseUrl: string): Promise<Driver> {
    const { logBrowserConsole, slowMo, headless } = getConfig('logBrowserConsole', 'slowMo', 'headless')
    const driver = await createDriverForTest({ sourcegraphBaseUrl, logBrowserConsole, slowMo, headless })
    driver.page.setDefaultNavigationTimeout(5 * 1000) // 5s navigation timeout
    return driver
}
