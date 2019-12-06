/**
 * @jest-environment node
 */

import * as path from 'path'
import { saveScreenshotsUponFailuresAndClosePage } from '../../../shared/src/e2e/screenshotReporter'
import { createDriverForTest, Driver } from '../../../shared/src/e2e/driver'
import { retry } from '../../../shared/src/e2e/e2e-test-utils'
import { ExternalServiceKind } from '../../../shared/src/graphql/schema'
import { testSingleFilePage } from './shared'
import { getConfig } from '../../../shared/src/e2e/config'

// By default, these tests run against a local Bitbucket instance and a local Sourcegraph instance.
// You can run them against other instances by setting the below env vars in addition to SOURCEGRAPH_BASE_URL.

const BITBUCKET_BASE_URL = process.env.BITBUCKET_BASE_URL || 'http://localhost:7990'
const BITBUCKET_USERNAME = process.env.BITBUCKET_USERNAME || 'test'
const BITBUCKET_PASSWORD = process.env.BITBUCKET_PASSWORD || 'test'
const TEST_NATIVE_INTEGRATION = Boolean(process.env.TEST_NATIVE_INTEGRATION)
const REPO_PATH_PREFIX = new URL(BITBUCKET_BASE_URL).hostname

const BITBUCKET_INTEGRATION_JAR_URL = 'https://storage.googleapis.com/sourcegraph-for-bitbucket-server/latest.jar'

const { sourcegraphBaseUrl } = getConfig('sourcegraphBaseUrl')

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
    if (new URL(page.url()).pathname.endsWith('/login')) {
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
    await driver.page.click('.next-step [name="connect"]')
    await retry(async () => {
        const browsePage = '/projects/SOURCEGRAPH/repos/jsonrpc2/browse'
        await driver.page.goto(BITBUCKET_BASE_URL + browsePage)
        // Retry until not redirected to the "import in progress" page anymore
        expect(new URL(driver.page.url()).pathname).toBe(new URL(browsePage, BITBUCKET_BASE_URL).pathname)
        // Ensure this is not a 404 page
        expect(await driver.page.$('.filebrowser-content')).toBeTruthy()
    })
}

/**
 * Configures the Sourcegraph for Bitbucket Server integration on the Bitbucket instance.
 */
async function configureSourcegraphIntegration(driver: Driver): Promise<void> {
    await driver.ensureHasCORSOrigin({ corsOriginURL: new URL(BITBUCKET_BASE_URL).origin })
    await bitbucketLogin(driver)
    await driver.page.goto(BITBUCKET_BASE_URL + '/plugins/servlet/upm?source=side_nav_manage_addons')
    await driver.page.waitForSelector('#upm-manage-plugins-user-installed')
    const sourcegraphPluginSelector = '.upm-plugin[data-key="com.sourcegraph.plugins.sourcegraph-bitbucket"]'
    if (await driver.page.$(sourcegraphPluginSelector)) {
        // Enable if needed
        if (await driver.page.$(`${sourcegraphPluginSelector}.disabled`)) {
            await driver.page.click(`${sourcegraphPluginSelector} [data-action="ENABLE"]`)
            await driver.page.waitForSelector(`${sourcegraphPluginSelector} [data-action="DISABLE"]`)
        }
    } else {
        // Install
        await driver.page.click('#upm-upload')
        await driver.page.waitForSelector('#upm-upload-url')
        await driver.page.type('#upm-upload-url', BITBUCKET_INTEGRATION_JAR_URL)
        await driver.page.click('#upm-upload-dialog button.confirm')
        await driver.page.waitForSelector(sourcegraphPluginSelector)
    }
    await driver.page.reload()
    await driver.page.waitForSelector('#sourcegraph-admin-link')
    await driver.page.click('#sourcegraph-admin-link')
    await driver.page.waitForSelector('form#admin')
    // The Sourcegraph URL input field is disabled until the Sourcegraph URL has been fetched.
    await retry(async () => {
        expect(
            await driver.page.evaluate(() => document.querySelector<HTMLInputElement>('form#admin input#url')!.disabled)
        ).toBe(false)
    })
    await driver.replaceText({ selector: 'form#admin input#url', newText: sourcegraphBaseUrl })
    await driver.page.click('form#admin input#submit')
    await driver.page.waitForSelector('.aui-message-success')
}

/**
 * Runs initial setup for the Bitbucket instance.
 */
async function init(driver: Driver): Promise<void> {
    await driver.ensureLoggedIn({ username: 'test', password: 'test', email: 'test@test.com' })
    if (TEST_NATIVE_INTEGRATION) {
        await configureSourcegraphIntegration(driver)
    } else {
        await bitbucketLogin(driver)
        await driver.setExtensionSourcegraphUrl()
    }
    await importBitbucketRepo(driver)
    await driver.ensureHasExternalService({
        kind: ExternalServiceKind.BITBUCKETSERVER,
        displayName: `Bitbucket ${BITBUCKET_BASE_URL}`,
        config: JSON.stringify({
            url: BITBUCKET_BASE_URL,
            username: BITBUCKET_USERNAME,
            password: BITBUCKET_PASSWORD,
            repos: ['SOURCEGRAPH/jsonrpc2'],
        }),
        ensureRepos: [REPO_PATH_PREFIX + '/SOURCEGRAPH/jsonrpc2'],
    })
    await driver.ensureHasCORSOrigin({ corsOriginURL: BITBUCKET_BASE_URL })
}

describe('Sourcegraph browser extension on Bitbucket Server', () => {
    let driver: Driver

    beforeAll(async () => {
        try {
            driver = await createDriverForTest({ loadExtension: !TEST_NATIVE_INTEGRATION, sourcegraphBaseUrl })
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
