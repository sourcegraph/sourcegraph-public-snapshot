import mkdirp from 'mkdirp-promise'
import * as path from 'path'
import puppeteer from 'puppeteer'

const REPO_ROOT = path.resolve(__dirname, '..', '..', '..', '..')
const SCREENSHOT_DIRECTORY = path.resolve(__dirname, '..', '..', 'puppeteer')
const EXTENSION_DIRECTORY = path.resolve(__dirname, '..', '..', 'build/chrome')

describe('Browser extension e2e test suite', () => {
    let authenticate: (page: puppeteer.Page) => Promise<void>
    let baseURL: string

    const overrideAuthSecret = process.env.OVERRIDE_AUTH_SECRET
    if (!overrideAuthSecret) {
        throw new Error('Auth secret not set - unable to execute tests')
    }

    if (process.env.SOURCEGRAPH_BASE_URL && process.env.SOURCEGRAPH_BASE_URL !== 'http://localhost:3080') {
        baseURL = process.env.SOURCEGRAPH_BASE_URL
    } else {
        baseURL = 'http://localhost:3080'
    }

    console.log('Using base URL', baseURL)

    authenticate = page => page.setExtraHTTPHeaders({ 'X-Override-Auth-Secret': overrideAuthSecret })

    const browserWSEndpoint = process.env.BROWSER_WS_ENDPOINT

    const disableDefaultFeatureFlags = async (page: puppeteer.Page) => {
        // Make feature flags mirror production
        await page.goto(baseURL)
        await page.evaluate(() => {
            window.localStorage.clear()
            window.localStorage.setItem('disableDefaultFeatureFlags', 'true')
        })
    }

    let browser: puppeteer.Browser
    let page: puppeteer.Page

    if (browserWSEndpoint) {
        before('Connect to browser', async () => {
            browser = await puppeteer.connect({ browserWSEndpoint })
        })
        after('Disconnect from browser', async () => {
            if (browser) {
                await browser.disconnect()
            }
        })
    } else {
        before('Start browser', async () => {
            let args: string[] | undefined
            if (process.getuid() === 0) {
                // TODO don't run as root in CI
                console.warn('Running as root, disabling sandbox')
                args = [
                    '--no-sandbox',
                    '--disable-setuid-sandbox',
                    `--disable-extensions-except=${EXTENSION_DIRECTORY}`,
                    `--load-extension=${EXTENSION_DIRECTORY}`,
                ]
            }
            browser = await puppeteer.launch({ args, headless: false })
        })
        after('Close browser', async () => {
            if (browser) {
                await browser.close()
            }
        })
    }

    beforeEach('Open page', async () => {
        page = await browser.newPage()
        await authenticate(page)
        await disableDefaultFeatureFlags(page)
    })

    afterEach('Close page', async function(): Promise<void> {
        if (page) {
            if (this.currentTest && this.currentTest.state === 'failed') {
                await mkdirp(SCREENSHOT_DIRECTORY)
                const filePath = path.join(
                    SCREENSHOT_DIRECTORY,
                    this.currentTest.fullTitle().replace(/\W/g, '_') + '.png'
                )
                await page.screenshot({ path: filePath })
                if (process.env.CI) {
                    // Print image with ANSI escape code for Buildkite
                    // https://buildkite.com/docs/builds/images-in-log-output
                    const relativePath = path.relative(REPO_ROOT, filePath)
                    console.log(`\u001B]1338;url="artifact://${relativePath}";alt="Screenshot"\u0007`)
                }
            }
            await page.close()
        }
    })

    const RepoHomepagePage = 'https://github.com/gorilla/mux'

    it('works!', async () => {
        await page.goto(RepoHomepagePage)
        await page.waitForSelector('li#open-on-sourcegraph')
    })
})
