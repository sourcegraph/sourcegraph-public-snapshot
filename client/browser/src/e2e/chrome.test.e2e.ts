import * as path from 'path'
import puppeteer from 'puppeteer'

const chromeExtensionPath = path.resolve(__dirname, '..', '..', 'build/chrome')

async function getTokenWithSelector(
    page: puppeteer.Page,
    token: string,
    selector: string
): Promise<puppeteer.ElementHandle> {
    const elements = await page.$$(selector)

    let element: puppeteer.ElementHandle<HTMLElement> | undefined
    for (const elem of elements) {
        const text = await page.evaluate(element => element.textContent, elem)
        if (text === token) {
            element = elem
            break
        }
    }

    if (!element) {
        throw new Error(`Unable to find token '${token}' with selector ${selector}`)
    }

    return element
}

async function clickElement(page: puppeteer.Page, element: puppeteer.ElementHandle): Promise<void> {
    // Wait for JS to be evaluated (https://github.com/GoogleChrome/puppeteer/issues/1805#issuecomment-357999249).
    await page.waitFor(500)
    await element.click()
}

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

        let args: string[] = [
            `--disable-extensions-except=${chromeExtensionPath}`,
            `--load-extension=${chromeExtensionPath}`,
        ]

        if (process.getuid() === 0) {
            // TODO don't run as root in CI
            console.warn('Running as root, disabling sandbox')
            args = [...args, '--no-sandbox', '--disable-setuid-sandbox']
        }

        browser = await puppeteer.launch({
            headless: false,
            // args: [`--disable-extensions-except=${chromeExtensionPath}`, `--load-extension=${chromeExtensionPath}`],
            args,
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
        await page.waitForSelector('li#open-on-sourcegraph')
    })

    it('injects toolbar for code views', async () => {
        await page.goto('https://github.com/gorilla/mux/blob/master/mux.go')
        await page.waitForSelector('.code-view-toolbar')
    })

    it('provides tooltips for single file', async () => {
        await page.goto('https://github.com/gorilla/mux/blob/master/mux.go')

        const element = await getTokenWithSelector(page, 'NewRouter', 'span.pl-en')

        await clickElement(page, element)

        await page.waitForSelector('.e2e-tooltip-j2d')
    })

    const tokens = {
        base: { text: 'matchHost', selector: 'span.pl-v' },
        head: { text: 'typ', selector: 'span.pl-v' },
    }

    for (const diffType of ['unified', 'split']) {
        for (const side of ['base', 'head']) {
            it(`provides tooltips for diff files (${diffType}, ${side})`, async () => {
                await page.goto(`https://github.com/gorilla/mux/pull/328/files?diff=${diffType}`)

                const token = tokens[side]
                const element = await getTokenWithSelector(page, token.text, token.selector)

                // Scrolls the element into view so that code view is in view.
                await element.hover()
                await page.waitForSelector(
                    '[data-path="regexp.go"] .code-view-toolbar .open-on-sourcegraph:nth-of-type(2)'
                )
                await clickElement(page, element)
                await page.waitForSelector('.e2e-tooltip-j2d')
            })
        }
    }
})
