import { createDriverForTest, Driver } from '../../../../shared/src/e2e/driver'
import { saveScreenshotsUponFailuresAndClosePage } from '../../../../shared/src/e2e/screenshotReporter'
import * as path from 'path'
import { Config } from '../../../../shared/src/e2e/config'
import { currentProductVersion } from './api'
import { GraphQLClient, createGraphQLClient } from './GraphQLClient'
import * as semver from 'semver'
import { TestResourceManager } from './TestResourceManager'

/**
 * Semver constraint on the Sourcegraph product version. Uses the syntax specified in
 * https://www.npmjs.com/package/semver. This should be updated when a change is made that breaks
 * compatibility between the regression tests and the Sourcegraph GUI. For example, when a new CSS
 * class is added that the regression tests rely on to identify a particular component.
 *
 * Note(beyang): this may not be up-to-date and might not be a useful mechanism, as supporting patch
 * releases is hard.
 */
const supportedSourcegraphVersionConstraint = '>=3.9'

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
    if (
        version !== 'dev' &&
        !semver.satisfies(version, supportedSourcegraphVersionConstraint, { includePrerelease: true })
    ) {
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
 * Returns tools for regression tests. The GraphQL client is authenticated with the sudo token, so
 * therefore has admin rights. The driver drives the puppeteer browser page instance. The resource
 * manager manages cleanup of resources created during the test.
 */
export async function getTestTools(
    config: Pick<
        Config,
        | 'sourcegraphBaseUrl'
        | 'logBrowserConsole'
        | 'slowMo'
        | 'headless'
        | 'sudoToken'
        | 'sudoUsername'
        | 'keepBrowser'
    >
): Promise<{
    gqlClient: GraphQLClient
    driver: Driver
    resourceManager: TestResourceManager
}> {
    const driver = await createAndInitializeDriver(config)
    const gqlClient = createGraphQLClient({
        baseUrl: config.sourcegraphBaseUrl,
        token: config.sudoToken,
        sudoUsername: config.sudoUsername,
    })
    await setTestDefaults(driver, gqlClient)
    const resourceManager = new TestResourceManager()

    return {
        driver,
        gqlClient,
        resourceManager,
    }
}

/**
 * Returns a Puppeteer driver with a 5s command timeout. It is important that none of the Jest test
 * timeouts is under 5s. Otherwise, the timeout error will be a cryptic Jest timeout error, instead
 * of an error pointing to the timed-out Puppeteer command.
 */
export async function createAndInitializeDriver(
    config: Pick<Config, 'sourcegraphBaseUrl' | 'logBrowserConsole' | 'slowMo' | 'headless' | 'keepBrowser'>
): Promise<Driver> {
    const driver = await createDriverForTest(config)
    driver.page.setDefaultNavigationTimeout(5 * 1000) // 5s navigation timeout
    return driver
}
