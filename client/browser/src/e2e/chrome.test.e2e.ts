import puppeteer from 'puppeteer'
import { chromeExtensionPath } from './utils/extension'

describe('Sourcegraph Chrome extension', () => {
    let authenticate: (page: puppeteer.Page) => Promise<void>

    let browser: puppeteer.Browser
    let page: puppeteer.Page

    const overrideAuthSecret = process.env.OVERRIDE_AUTH_SECRET
    if (!overrideAuthSecret) {
        throw new Error('Auth secret not set - unable to execute tests')
    }

    authenticate = page => page.setExtraHTTPHeaders({ 'X-Override-Auth-Secret': overrideAuthSecret })

    before(async function(): Promise<void> {
        this.timeout(90 * 1000)

        browser = await puppeteer.launch({
            headless: false,
            args: [`--disable-extensions-except=${chromeExtensionPath}`, `--load-extension=${chromeExtensionPath}`],
        })
    })

    after(async () => {
        if (browser) {
            await browser.close()
        }
    })

    beforeEach('Open page', async () => {
        page = await browser.newPage()
        await authenticate(page)
    })

    const repoBaseURL = 'https://github.com/gorilla/mux'

    it('injects View on Sourcegraph', async () => {
        await page.goto(repoBaseURL)
        await page.waitForSelector('li#open-on-sourcegraph.use-extensions')
    })
})
