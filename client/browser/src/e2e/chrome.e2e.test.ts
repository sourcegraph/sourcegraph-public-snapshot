import * as path from 'path'
import puppeteer from 'puppeteer'
import { ensureLoggedIn, readEnvString } from '../../../../shared/src/util/e2e-test-utils'
import { saveScreenshotsUponFailuresAndClosePage } from '../../../../shared/src/util/screenshotReporter'

const chromeExtensionPath = path.resolve(__dirname, '..', '..', 'build/chrome')

const SOURCEGRAPH_BASE_URL = readEnvString({
    variable: 'SOURCEGRAPH_BASE_URL',
    defaultValue: 'https://sourcegraph.com',
})

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
    let browser: puppeteer.Browser
    let page: puppeteer.Page

    // Open browser.
    beforeAll(
        async (): Promise<void> => {
            jest.setTimeout(90 * 1000)

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
                args,
            })
        }
    )

    // Open page.
    beforeEach(async () => {
        page = await browser.newPage()
    })

    // Take a screenshot when a test fails.
    saveScreenshotsUponFailuresAndClosePage(
        path.resolve(__dirname, '..', '..', '..', '..'),
        path.resolve(__dirname, '..', '..', '..', '..', 'puppeteer'),
        () => page
    )

    // Close browser.
    afterAll(async () => {
        if (browser) {
            if (page && !page.isClosed()) {
                await page.close()
            }
            await browser.close()
        }
    })

    const repoBaseURL = 'https://github.com/gorilla/mux'

    test('connects to sourcegraph.com by default', async () => {
        await page.goto('chrome-extension://bmfbcejdknlknpncfpeloejonjoledha/options.html')
        await page.waitForSelector('.server-url-form__input-container__status__indicator--success')
    })

    if (SOURCEGRAPH_BASE_URL !== 'https://sourcegraph.com') {
        test("sets the Sourcegraph URL and warns the user he's not logged in", async () => {
            await page.goto('chrome-extension://bmfbcejdknlknpncfpeloejonjoledha/options.html')
            await page.waitForSelector('.server-url-form__input-container__input')
            await page.focus('.server-url-form__input-container__input')
            // Erase input value
            await page.evaluate(
                () =>
                    ((document.querySelector('.server-url-form__input-container__input')! as HTMLInputElement).value =
                        '')
            )
            await page.keyboard.type(SOURCEGRAPH_BASE_URL)
            // Clear the focus on the input field.
            await page.evaluate(() => (document.activeElement! as HTMLElement).blur())
            await page.waitForSelector('.server-url-form__input-container__status__indicator--error')
            await getTokenWithSelector(page, 'Sign in to your instance', '.server-url-form__error a')
        })

        test('connects to the instance once the user is logged in', async () => {
            await ensureLoggedIn({ page, baseURL: SOURCEGRAPH_BASE_URL })
            await page.goto('chrome-extension://bmfbcejdknlknpncfpeloejonjoledha/options.html')
            await page.waitForSelector('.server-url-form__input-container__status__indicator--success')
        })
    }

    test('injects View on Sourcegraph', async () => {
        await page.goto(repoBaseURL)
        await page.waitForSelector('li#open-on-sourcegraph')
    })

    test('injects toolbar for code views', async () => {
        await page.goto('https://github.com/gorilla/mux/blob/master/mux.go')
        await page.waitForSelector('.code-view-toolbar')
    })

    test('provides tooltips for single file', async () => {
        await page.goto('https://github.com/gorilla/mux/blob/master/mux.go')

        const element = await getTokenWithSelector(page, 'NewRouter', 'span.pl-en')

        await clickElement(page, element)

        await page.waitForSelector('.e2e-tooltip-go-to-definition')
    })

    const tokens = {
        base: { text: 'matchHost', selector: 'span.pl-v' },
        head: { text: 'regexpType', selector: 'span.pl-v' },
    }

    for (const diffType of ['unified', 'split']) {
        for (const side of ['base', 'head']) {
            test(`provides tooltips for diff files (${diffType}, ${side})`, async () => {
                await page.goto(`https://github.com/gorilla/mux/pull/328/files?diff=${diffType}`)

                const token = tokens[side as 'base' | 'head']
                const element = await getTokenWithSelector(page, token.text, token.selector)

                // Scrolls the element into view so that code view is in view.
                await element.hover()
                await page.waitForSelector('[data-path="regexp.go"] .code-view-toolbar .open-on-sourcegraph')
                await clickElement(page, element)
                await page.waitForSelector('.e2e-tooltip-go-to-definition')
            })
        }
    }
})
