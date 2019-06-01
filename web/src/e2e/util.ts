import { percySnapshot as realPercySnapshot } from '@percy/puppeteer'
import * as os from 'os'
import puppeteer, { LaunchOptions } from 'puppeteer'
import { Key } from 'ts-key-enum'
import { readEnvBoolean, readEnvString, retry } from '../util/e2e-test-utils'

export const percySnapshot = readEnvBoolean({ variable: 'PERCY_ON', defaultValue: false })
    ? realPercySnapshot
    : async () => undefined

/**
 * Used in the external service configuration.
 */
export const gitHubToken = readEnvString({ variable: 'GITHUB_TOKEN' })

export const baseURL = readEnvString({ variable: 'SOURCEGRAPH_BASE_URL', defaultValue: 'http://localhost:3080' })

/**
 * Specifies how `replaceText` will select the content of the element. No
 * single method works in all cases:
 *
 * - Meta+A doesn't work in input boxes https://github.com/GoogleChrome/puppeteer/issues/1313
 * - selectall doesn't work in the Monaco editor
 */
type ReplaceTextMethod = 'selectall' | 'keyboard'

export async function newDriverForTest(): Promise<Driver> {
    let args: string[] = []
    if (process.getuid() === 0) {
        // TODO don't run as root in CI
        console.warn('Running as root, disabling sandbox')
        args = ['--no-sandbox', '--disable-setuid-sandbox']
    }
    const launchOpt: LaunchOptions = {
        args: [...args, '--window-size=1280,1024'],
        headless: readEnvBoolean({ variable: 'HEADLESS', defaultValue: false }),
    }
    const browser = await puppeteer.launch(launchOpt)
    const page = await browser.newPage()
    page.on('console', message => {
        if (message.text().indexOf('Download the React DevTools') !== -1) {
            return
        }
        const seen = new WeakSet()
        console.log(
            'Browser console message:',
            JSON.stringify(
                {
                    type: message.type(),
                    text: message.text(),
                    location: message.location(),
                    args: message.args().length > 0 ? message.args() : undefined,
                },
                (key, value) => {
                    if (typeof value === 'object' && value !== null) {
                        if (seen.has(value)) {
                            return
                        }
                        seen.add(value)
                    }
                    return value
                },
                2
            )
        )
    })
    return new Driver(browser, page)
}

export class Driver {
    constructor(public browser: puppeteer.Browser, public page: puppeteer.Page) {}

    public async ensureLoggedIn(): Promise<void> {
        await this.page.goto(baseURL)
        await this.page.evaluate(() => {
            localStorage.setItem('has-dismissed-browser-ext-toast', 'true')
            localStorage.setItem('has-dismissed-integrations-toast', 'true')
            localStorage.setItem('has-dismissed-survey-toast', 'true')
        })
        const url = new URL(await this.page.url())
        if (url.pathname === '/site-admin/init') {
            await this.page.type('input[name=email]', 'test@test.com')
            await this.page.type('input[name=username]', 'test')
            await this.page.type('input[name=password]', 'test')
            await this.page.click('button[type=submit]')
            await this.page.waitForNavigation()
        } else if (url.pathname === '/sign-in') {
            await this.page.type('input', 'test')
            await this.page.type('input[name=password]', 'test')
            await this.page.click('button[type=submit]')
            await this.page.waitForNavigation()
        }
    }

    public async close(): Promise<void> {
        await this.browser.close()
    }

    public async replaceText({
        selector,
        newText,
        method = 'selectall',
    }: {
        selector: string
        newText: string
        method?: ReplaceTextMethod
    }): Promise<void> {
        const selectAllByMethod: Record<ReplaceTextMethod, () => Promise<void>> = {
            selectall: async () => {
                await this.page.evaluate(() => document.execCommand('selectall', false))
            },
            keyboard: async () => {
                const modifier = os.platform() === 'darwin' ? Key.Meta : Key.Control
                await this.page.keyboard.down(modifier)
                await this.page.keyboard.press('a')
                await this.page.keyboard.up(modifier)
            },
        }

        // The Monaco editor sometimes detaches nodes from the DOM, causing
        // `click()` to fail unpredictably.
        await retry(async () => {
            await this.page.waitForSelector(selector)
            await this.page.click(selector)
        })
        await selectAllByMethod[method]()
        await this.page.keyboard.press(Key.Backspace)
        await this.page.keyboard.type(newText)
    }

    public async ensureHasExternalService(
        kind: string,
        displayName: string,
        config: string,
        ensureRepos?: string[]
    ): Promise<void> {
        await this.page.goto(baseURL + '/site-admin/external-services')
        await this.page.waitFor('.e2e-filtered-connection')
        await this.page.waitForSelector('.e2e-filtered-connection__loader', { hidden: true })

        // Matches buttons for deleting external services named ${displayName}.
        const deleteButtonSelector = `[data-e2e-external-service-name="${displayName}"] .e2e-delete-external-service-button`
        if (await this.page.$(deleteButtonSelector)) {
            const accept = async (dialog: puppeteer.Dialog) => {
                await dialog.accept()
                this.page.off('dialog', accept)
            }
            this.page.on('dialog', accept)
            await this.page.click(deleteButtonSelector)
        }

        await (await this.page.waitForSelector('.e2e-goto-add-external-service-page', { visible: true })).click()

        await (await this.page.waitForSelector(`.linked-external-service-card--${kind}`, { visible: true })).click()

        await this.replaceText({ selector: '#e2e-external-service-form-display-name', newText: displayName })

        // Type in a new external service configuration.
        await this.replaceText({
            selector: '.view-line',
            newText: config,
            method: 'keyboard',
        })
        await this.page.click('.e2e-add-external-service-button')
        await this.page.waitForNavigation()

        if (ensureRepos) {
            // Wait for repositories to sync.
            await this.page.goto(baseURL + '/site-admin/repositories?query=gorilla%2Fmux')
            await retry(async () => {
                await this.page.reload()
                await this.page.waitForSelector(`.repository-node[data-e2e-repository='github.com/gorilla/mux']`, {
                    timeout: 5000,
                })
            })

            // Clone the repositories
            for (const slug of ensureRepos) {
                await this.page.goto(baseURL + `/site-admin/repositories?query=${encodeURIComponent(slug)}`)
                await this.page.waitForSelector(`.repository-node[data-e2e-repository='github.com/${slug}']`, {
                    visible: true,
                })
            }
        }
    }

    public async assertWindowLocation(location: string, isAbsolute = false): Promise<any> {
        const url = isAbsolute ? location : baseURL + location
        await retry(async () => {
            expect(await this.page.evaluate(() => window.location.href)).toEqual(url)
        })
    }

    public async assertWindowLocationPrefix(locationPrefix: string, isAbsolute = false): Promise<any> {
        const prefix = isAbsolute ? locationPrefix : baseURL + locationPrefix
        await retry(async () => {
            const loc: string = await this.page.evaluate(() => window.location.href)
            expect(loc.startsWith(prefix)).toBeTruthy()
        })
    }

    public async assertStickyHighlightedToken(label: string): Promise<void> {
        await this.page.waitForSelector('.selection-highlight-sticky', { visible: true }) // make sure matched token is highlighted
        await retry(async () =>
            expect(
                await this.page.evaluate(() => document.querySelector('.selection-highlight-sticky')!.textContent)
            ).toEqual(label)
        )
    }

    public async assertAllHighlightedTokens(label: string): Promise<void> {
        const highlightedTokens = await this.page.evaluate(() =>
            Array.from(document.querySelectorAll('.selection-highlight')).map(el => el.textContent || '')
        )
        expect(highlightedTokens.every(txt => txt === label)).toBeTruthy()
    }

    public async assertNonemptyLocalRefs(): Promise<void> {
        // verify active group is references
        await this.page.waitForXPath(
            "//*[contains(@class, 'panel__tabs')]//*[contains(@class, 'tab-bar__tab--active') and contains(text(), 'References')]"
        )
        // verify there are some references
        await this.page.waitForSelector('.panel__tabs-content .file-match-children__item', { visible: true })
    }

    public async assertNonemptyExternalRefs(): Promise<void> {
        // verify active group is references
        await this.page.waitForXPath(
            "//*[contains(@class, 'panel__tabs')]//*[contains(@class, 'tab-bar__tab--active') and contains(text(), 'References')]"
        )
        // verify there are some references
        await this.page.waitForSelector('.panel__tabs-content .hierarchical-locations-view__item', { visible: true })
    }
}
