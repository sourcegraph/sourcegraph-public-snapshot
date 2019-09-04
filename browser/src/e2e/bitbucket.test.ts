import * as path from 'path'
import { saveScreenshotsUponFailuresAndClosePage } from '../../../shared/src/e2e/screenshotReporter'
import { sourcegraphBaseUrl, createDriverForTest, Driver } from '../../../shared/src/e2e/driver'
import { retry } from '../../../shared/src/e2e/e2e-test-utils'
import { ExternalServiceKind } from '../../../shared/src/graphql/schema'
import { testSingleFilePage } from './shared'

// By default, these tests run against a local Bitbucket instance and a local Sourcegraph instance.
// You can run them against other instances by setting the below env vars in addition to SOURCEGRAPH_BASE_URL.

const BITBUCKET_BASE_URL = process.env.BITBUCKET_BASE_URL || 'http://localhost:7990'
const BITBUCKET_USERNAME = process.env.BITBUCKET_USERNAME || 'test'
const BITBUCKET_PASSWORD = process.env.BITBUCKET_PASSWORD || 'test'
const REPO_PATH_PREFIX = new URL(BITBUCKET_BASE_URL).hostname

// 1 minute test timeout. This must be greater than the default Puppeteer
// command timeout of 30s in order to get the stack trace to point to the
// Puppeteer command that failed instead of a cryptic Jest test timeout
// location.
jest.setTimeout(1000 * 60 * 1000)

/**
 * Logs into Bitbucket.
 */
async function bitbucketLogin({ page }: Driver): Promise<void> {
    await page.goto(BITBUCKET_BASE_URL)
    if (new URL(page.url()).pathname === '/login') {
        await page.type('#j_username', BITBUCKET_USERNAME)
        await page.type('#j_password', BITBUCKET_PASSWORD)
        await Promise.all([page.click('#submit'), page.waitForNavigation()])
    }
}

/**
 * Adds sourcegraph/jsonrpc2 to this Bitbucket instance.
 */
async function importBitbucketRepo(driver: Driver): Promise<void> {
    // Import repo (idempotent)
    await driver.page.goto(BITBUCKET_BASE_URL + '/plugins/servlet/import-repository/SOURCEGRAPH')
    await driver.page.waitForSelector('button[data-source="GIT"]')
    await driver.page.click('button[data-source="GIT"]')
    await driver.page.waitForSelector('input[name="url"]')
    await driver.page.type('.source-form.git-specific input[name="url"]', 'https://github.com/sourcegraph/jsonrpc2')
    // Need to focus the next input field to trigger validation and have the submit button be enabled
    await driver.page.focus('.source-form.git-specific input[name="username"]')
    await retry(async () => {
        const browsePage = '/projects/SOURCEGRAPH/repos/jsonrpc2/browse'
        await driver.page.goto(BITBUCKET_BASE_URL + browsePage)
        // Retry until not redirected to the "import in progress" page anymore
        expect(new URL(driver.page.url()).pathname).toBe(browsePage)
    })
}

/**
 * Runs initial setup for the Bitbucket instance.
 */
async function init(driver: Driver): Promise<void> {
    await driver.ensureLoggedIn()
    await driver.setExtensionSourcegraphUrl()
    await driver.ensureHasExternalService({
        kind: ExternalServiceKind.BITBUCKETSERVER,
        displayName: 'Bitbucket',
        config: JSON.stringify({
            url: BITBUCKET_BASE_URL,
            username: BITBUCKET_USERNAME,
            password: BITBUCKET_PASSWORD,
            repos: ['SOURCEGRAPH/jsonrpc2'],
        }),
        ensureRepos: [REPO_PATH_PREFIX + '/SOURCEGRAPH/jsonrpc2'],
    })
    await driver.ensureHasCORSOrigin({ corsOriginURL: BITBUCKET_BASE_URL })
    await bitbucketLogin(driver)
    await importBitbucketRepo(driver)
}

describe('Sourcegraph browser extension on Bitbucket Server', () => {
    let driver: Driver

    beforeAll(async () => {
        try {
            driver = await createDriverForTest({ loadExtension: true })
            await init(driver)
        } catch (err) {
            console.error(err)
            setTimeout(() => process.exit(1), 100)
        }
    }, 4 * 60 * 1000)

    afterAll(async () => {
        await driver.close()
    })

    // Take a screenshot when a test fails.
    saveScreenshotsUponFailuresAndClosePage(
        path.resolve(__dirname, '..', '..', '..', '..'),
        path.resolve(__dirname, '..', '..', '..', '..', 'puppeteer'),
        () => driver.page
    )

    testSingleFilePage({
        getDriver: () => driver,
        url: `${BITBUCKET_BASE_URL}/projects/SOURCEGRAPH/repos/jsonrpc2/browse/call_opt.go?until=4fb7cd90793ee6ab445f466b900e6bffb9b63d78&untilPath=call_opt.go`,
        repoName: `${REPO_PATH_PREFIX}/SOURCEGRAPH/jsonrpc2`,
        sourcegraphBaseUrl,
        lineSelector: '.line',
    })
})
