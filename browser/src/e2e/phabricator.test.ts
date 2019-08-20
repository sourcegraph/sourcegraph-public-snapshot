import * as path from 'path'
import { saveScreenshotsUponFailuresAndClosePage } from '../../../shared/src/e2e/screenshotReporter'
import { getTokenWithSelector } from '../../../shared/src/e2e/e2e-test-utils'
import { baseURL, createDriverForTest, Driver, gitHubToken } from '../../../shared/src/e2e/driver'

const PHABRICATOR_BASE_URL = 'http://127.0.0.1'

// 1 minute test timeout. This must be greater than the default Puppeteer
// command timeout of 30s in order to get the stack trace to point to the
// Puppeteer command that failed instead of a cryptic Jest test timeout
// location.
jest.setTimeout(1000 * 60 * 1000)

/**
 * Logs into Phabricator as admin/sourcegraph.
 */
async function phabricatorLogin({ page }: Driver): Promise<void> {
    await page.goto(PHABRICATOR_BASE_URL)
    await page.waitForSelector('.phabricator-wordmark')
    if (await page.$('input[name=username]')) {
        await page.type('input[name=username]', 'admin')
        await page.type('input[name=password]', 'sourcegraph')
        await page.click('button[name="__submit__"]')
    }
    await page.waitForSelector('.phabricator-core-user-menu')
}

/**
 * Waits for the jrpc repository to finish cloning.
 */
async function repositoryCloned(driver: Driver): Promise<void> {
    await driver.page.goto(PHABRICATOR_BASE_URL + '/source/jrpc/manage/status/')
    try {
        await getTokenWithSelector(driver.page, 'Fully Imported', 'td.phui-status-item-target')
    } catch (err) {
        await new Promise<void>(resolve => setTimeout(resolve, 1000))
        await repositoryCloned(driver)
    }
}

/**
 * Adds sourcegraph/jsonrpc2 to this Phabricator instance.
 */
async function addPhabricatorRepo(driver: Driver): Promise<void> {
    // Add new repo to Diffusion
    await driver.page.goto(PHABRICATOR_BASE_URL + '/diffusion/edit/?vcs=git')
    await driver.page.waitForSelector('input[name=shortName]')
    await driver.page.type('input[name=name]', 'sourcegraph/jsonrpc2')
    await driver.page.type('input[name=callsign]', 'JRPC')
    await driver.page.type('input[name=shortName]', 'jrpc')
    await driver.page.click('button[type=submit]')
    // Configure it to clone github.com/sourcegraph/jsonrpc2
    await driver.page.goto(PHABRICATOR_BASE_URL + '/source/jrpc/uri/edit/')
    await driver.page.waitForSelector('input[name=uri]')
    await driver.page.type('input[name=uri]', 'https://github.com/sourcegraph/jsonrpc2.git')
    await driver.page.select('select[name=io]', 'observe')
    await driver.page.select('select[name=display]', 'always')
    const saveRepo = await getTokenWithSelector(driver.page, 'Create Repository URI', 'button')
    await saveRepo.click()
    // Activate the repo and wait for it to clone
    await driver.page.goto(PHABRICATOR_BASE_URL + '/source/jrpc/manage/')
    await driver.page.waitForSelector('a[href="/source/jrpc/edit/activate/"]')
    await driver.page.click('a[href="/source/jrpc/edit/activate/"]')
    await driver.page.waitForSelector('form[action="/source/jrpc/edit/activate/"]')
    if ((await driver.page.$x('//button[text()="Activate Repository"]')).length > 0) {
        await (await driver.page.$x('//button[text()="Activate Repository"]'))[0].click()
        await driver.page.waitForNavigation()
        await repositoryCloned(driver)
    }

    // Configure the Sourcegraph URL
    await driver.page.goto(PHABRICATOR_BASE_URL + '/config/edit/sourcegraph.url/')
    await driver.replaceText({
        selector: 'input[name=value]',
        newText: baseURL,
    })
    await (await getTokenWithSelector(driver.page, 'Save Config Entry', 'button')).click()

    // Configure the repository mappings
    await driver.page.goto(PHABRICATOR_BASE_URL + '/config/edit/sourcegraph.callsignMappings/')
    await driver.replaceText({
        selector: 'textarea[name=value]',
        newText: `[
        {
          "path": "github.com/sourcegraph/jsonrpc2",
          "callsign": "JRPC"
        }
      ]`,
    })
    const setCallsignMappings = await getTokenWithSelector(driver.page, 'Save Config Entry', 'button')
    await setCallsignMappings.click()
    await driver.page.waitForNavigation()
}

/**
 * Runs initial setup for the Phabricator instance.
 */
async function init(driver: Driver): Promise<void> {
    await driver.ensureLoggedIn()
    await driver.ensureHasExternalService({
        kind: 'GITHUB',
        displayName: 'Github (phabricator)',
        config: JSON.stringify({
            url: 'https://github.com',
            token: gitHubToken,
            repos: ['sourcegraph/jsonrpc2'],
            repositoryQuery: ['none'],
        }),
    })
    await driver.ensureHasCORSOrigin({ corsOriginURL: 'http://127.0.0.1' })
    await phabricatorLogin(driver)
    await addPhabricatorRepo(driver)
}

describe('Sourcegraph Phabricator extension', () => {
    let driver: Driver

    beforeAll(async () => {
        driver = await createDriverForTest()
        await init(driver)
    }, 4 * 60 * 1000)

    // Take a screenshot when a test fails.
    saveScreenshotsUponFailuresAndClosePage(
        path.resolve(__dirname, '..', '..', '..', '..'),
        path.resolve(__dirname, '..', '..', '..', '..', 'puppeteer'),
        () => driver.page
    )

    it('Adds "View on Sourcegraph buttons to files" and code intelligence hovers', async () => {
        await driver.newPage()
        await driver.page.goto(PHABRICATOR_BASE_URL + '/source/jrpc/browse/master/call_opt.go')
        await driver.page.waitForSelector('.code-view-toolbar .open-on-sourcegraph')

        // Pause to give codeintellify time to register listeners for
        // tokenization (only necessary in CI, not sure why).
        await driver.page.waitFor(1000)

        // Trigger tokenization of the line.
        const n = 5
        const codeLine = await driver.page.waitForSelector(`.diffusion-source > tbody > tr:nth-child(${n}) > td`)
        await codeLine.click()

        // Once the line is tokenized, we can click on the individual token we want a hover for.
        const codeElement = await driver.page.waitForXPath(`//tbody/tr[${n}]//span[text()="CallOption"]`)
        await codeElement.click()
        await driver.page.waitForSelector('.e2e-tooltip-go-to-definition')
    })

    afterAll(async () => {
        await driver.close()
    })
})
